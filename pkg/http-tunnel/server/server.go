// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/proto"
	"github.com/inconshreveable/go-vhost"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Server is responsible for proxying public connections to the client over a
// tunnel connection.
type Server struct {
	*registry
	config *serverData

	listener    net.Listener
	connPool    *connPool
	httpClient  *http.Client
	logger      *log.Entry
	vhostMuxer  *vhost.TLSMuxer
	PortPool    *AddrPool
	authHandler func(string) bool
}

// ServerConfig defines all data needed for running the Server.
type ServerConfig struct {
	// Address is TCP address to listen for client connections. If empty ":0" is used.
	Address string
	// Address Pool enables Port AutoAssignation. If empty "10000:50000" is used.
	PortRange string `default:"10000:50000"`
	// AuthHandler is function validates provided auth token
	AuthHandler func(string) bool
	// Logger is optional logger. If nil logging is disabled.
	Logger *log.Entry
}

type serverData struct {
	addr        string
	portRange   string
	listener    net.Listener
	logger      *log.Entry
	authHandler func(string) bool
}

// NewServer creates a new Server based on configuration.
// Caller must invoke Start() on returned instance in order to start server
func NewServer(config *ServerConfig) (*Server, error) {
	serverData := &serverData{}

	if config.Address == "" {
		return nil, errors.New("provided empty address")
	}
	serverData.addr = config.Address

	if config.PortRange == "" {
		config.PortRange = "10000:50000"
	}
	serverData.portRange = config.PortRange


	logger := config.Logger
	if logger == nil {
		l := log.New()
		l.SetLevel(log.ErrorLevel)
		logger = log.NewEntry(l)
	}
	serverData.logger = logger

	if config.AuthHandler == nil {
		return nil, errors.New("provided auth handler is nil")
	}
	serverData.authHandler = config.AuthHandler

	return newServer(serverData)
}

func newServer(serverData *serverData) (*Server, error) {
	pPool := &AddrPool{}
	err := pPool.Init(serverData.portRange)
	if err != nil {
		return nil, fmt.Errorf("failed to create port range pool: %s", err)
	}

	listener, err := listener(serverData)
	if err != nil {
		return nil, fmt.Errorf("listener failed: %s", err)
	}

	s := &Server{
		registry: newRegistry(serverData.logger),
		PortPool: pPool,
		config:   serverData,
		listener: listener,
		logger:   serverData.logger,
	}

	s.authHandler = serverData.authHandler

	t := &http2.Transport{}
	pool := newConnPool(t, s.disconnected)
	t.ConnPool = pool
	s.connPool = pool
	s.httpClient = &http.Client{
		Transport: t,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return s, nil
}

func listener(config *serverData) (net.Listener, error) {
	if config.listener != nil {
		return config.listener, nil
	}

	if config.addr == "" {
		return nil, errors.New("missing addr")
	}

	return net.Listen("tcp", config.addr)
}

// disconnected clears resources used by client, it's invoked by connection pool
// when client goes away.
func (s *Server) disconnected(identifier string) {
	ilogger := s.logger.WithFields(log.Fields{"identifier": identifier})
	ilogger.Debug("disconnected")

	i := s.registry.clear(identifier)
	if i == nil {
		ilogger.Errorf("ERROR ON DISCONNECT (registry not found)")
		return
	}

	iclogger := ilogger.WithFields(log.Fields{"client-name": i.ClientName})
	iclogger.Debug("DISCONNECT")

	for _, l := range i.Listeners {
		iclogger.Debugf("close listener for %v", l.Addr())
		_ = l.Close()
		_ = s.PortPool.Release(i.ClientName)
	}
}

// Start starts accepting connections form clients. For accepting http traffic
// from end users server must be run as handler on http server.
func (s *Server) Start() {
	addr := s.listener.Addr().String()

	alogger := s.logger.WithFields(log.Fields{"address": addr})
	alogger.Info("start http-tunnel server")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				alogger.Debug("control connection listener closed")
				return
			}

			alogger.Error("accept of control connection failed", err)
			continue
		}

		if err := tunnel.KeepAlive(conn); err != nil {
			alogger.Error("TCP keepalive for control connection failed", err)
		}

		go s.handleClient(conn)
	}
}

type TunnelExt struct {
	IdName  string
	Tunnels map[string]*proto.Tunnel
}

func (s *Server) handleClient(conn net.Conn) {
	alogger := s.logger.WithFields(log.Fields{"address": conn.RemoteAddr()})
	alogger.Info("try connect")

	var (
		conid   string
		req     *http.Request
		resp    *http.Response
		tunnels TunnelExt
		err     error

		inConnPool bool
		token      string
	)

	conid = conn.RemoteAddr().String()

	s.PreSubscribe(conid)

	if err = conn.SetDeadline(time.Time{}); err != nil {
		alogger.Error("setting infinite deadline failed", err)
		goto reject
	}

	if err := s.connPool.AddConn(conn, conid); err != nil {
		alogger.Error("adding connection failed", err)
		goto reject
	}
	inConnPool = true

	req, err = http.NewRequest(http.MethodConnect, s.connPool.URL(conid), nil)
	if err != nil {
		alogger.Error("handshake request creation failed", err)
		goto reject
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), tunnel.DefaultTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	resp, err = s.httpClient.Do(req)
	if err != nil {
		alogger.Error("handshake failed 1", err)
		goto reject
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Status %s", resp.Status)
		alogger.Error("handshake failed 2", err)
		goto reject
	}

	// needs additional auth
	if s.authHandler != nil {
		token = resp.Header.Get("X-Auth-Header")
		if token == "" {
			err = errors.New("Auth header missing")
			alogger.Error("handshake failed", err)
			goto reject
		}

		authorized := s.authHandler(token)
		if !authorized {
			err = errors.New("Unauthorized request")
			alogger.Error("handshake failed", err)
			goto reject
		}
	}

	if resp.ContentLength == 0 {
		err = fmt.Errorf("tunnels Content-Legth: 0")
		alogger.Error("handshake failed 3", err)
		goto reject
	}

	if err = json.NewDecoder(&io.LimitedReader{R: resp.Body, N: 126976}).Decode(&tunnels); err != nil {
		alogger.Error("handshake failed 4", err)
		goto reject
	}

	alogger.Infof("client name has been set to %s and id %s", tunnels.IdName, conid)

	s.Subscribe(tunnels.IdName, conid)

	if len(tunnels.Tunnels) == 0 {
		err = fmt.Errorf("no tunnels")
		alogger.Error("handshake failed 5", err)
		goto reject
	}

	if err = s.addTunnels(tunnels.IdName, tunnels.Tunnels); err != nil {
		alogger.Error("handshake failed 6", err)
		goto reject
	}

	alogger.Infof("%s connected", tunnels.IdName)

	return

reject:
	log.Info("rejected")

	if inConnPool {
		s.notifyError(err, conid)
		s.connPool.DeleteConn(tunnels.IdName)
	}

	conn.Close()
}

// notifyError tries to send error to client.
func (s *Server) notifyError(serverError error, conid string) {
	if serverError == nil {
		return
	}

	req, err := http.NewRequest(http.MethodConnect, s.connPool.URL(conid), nil)
	if err != nil {
		s.logger.Errorf("client error notification failed for %s with %v", conid, err)
		return
	}

	req.Header.Set(proto.HeaderError, serverError.Error())

	ctx, cancel := context.WithTimeout(context.Background(), tunnel.DefaultTimeout)
	defer cancel()

	_, _ = s.httpClient.Do(req.WithContext(ctx))
}

func (s *Server) adrListenRegister(in string, cid string, portname string) (string, error) {
	inarr := strings.Split(in, ":")
	host := inarr[0]
	port := inarr[1]
	if port == "AUTO" {
		port, err := s.PortPool.Acquire(cid, portname)
		if err != nil {
			return "", fmt.Errorf("Error on acquire port from port pool:%s", err)
		}
		addr := host + ":" + strconv.Itoa(port)

		s.logger.WithFields(log.Fields{
			"client-id": cid,
			"portname":  portname,
			"address":   addr,
		}).Info("address auto assign")

		return addr, nil
	}
	return in, nil
}

// addTunnels invokes addHost or addListener based on data from proto.Tunnel. If
// a tunnel cannot be added whole batch is reverted.
func (s *Server) addTunnels(cname string, tunnels map[string]*proto.Tunnel) error {
	i := &RegistryItem{
		Hosts:      []*HostAuth{},
		Listeners:  []net.Listener{},
		ClientName: cname,
	}

	var err error
	var portnames []string

	for name, t := range tunnels {
		portnames = append(portnames, name)
		cplogger := s.logger.WithFields(log.Fields{"client-id": cname, "port-name": name})
		switch t.Protocol {
		case proto.HTTP:
			i.Hosts = append(i.Hosts, &HostAuth{t.Host, NewAuth(t.Auth)})
		case proto.TCP, proto.TCP4, proto.TCP6, proto.UNIX:
			var l net.Listener
			addr, err := s.adrListenRegister(t.Addr, cname, name)
			if err != nil {
				goto rollback
			}
			l, err = net.Listen(t.Protocol, addr)
			if err != nil {
				goto rollback
			}

			cplogger.Infof("open listener for address %v", l.Addr())

			i.Listeners = append(i.Listeners, l)
		case proto.SNI:
			if s.vhostMuxer == nil {
				err = fmt.Errorf("unable to configure SNI for tunnel %s: %s", name, t.Protocol)
				goto rollback
			}
			var l net.Listener
			l, err = s.vhostMuxer.Listen(t.Host)
			if err != nil {
				goto rollback
			}

			cplogger.Infof("add SNI vhost for host %s", t.Host)

			i.Listeners = append(i.Listeners, l)
		default:
			err = fmt.Errorf("unsupported protocol for tunnel %s: %s", name, t.Protocol)
			goto rollback
		}
	}
	i.ListenerNames = portnames

	err = s.set(i, cname)
	if err != nil {
		goto rollback
	}

	for k, l := range i.Listeners {
		go s.listen(l, i.ClientName, i.ListenerNames[k])
	}

	return nil

rollback:
	for _, l := range i.Listeners {
		l.Close()
	}

	return err
}

// Unsubscribe removes client from registry, disconnects client if already
// connected and returns it's RegistryItem.
func (s *Server) Unsubscribe(identifier string, idname string) *RegistryItem {
	s.connPool.DeleteConn(identifier)
	return s.registry.Unsubscribe(identifier, idname)
}

// Ping measures the RTT response time.
func (s *Server) Ping(identifier string) (time.Duration, error) {
	return s.connPool.Ping(identifier)
}

func (s *Server) listen(l net.Listener, cname string, pname string) {
	addr := l.Addr().String()
	cplogger := s.logger.WithFields(log.Fields{"client-name": cname, "port-name": pname})

	for {
		conn, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") ||
				strings.Contains(err.Error(), "listener closed") {
				cplogger.Errorf("listener closed for address %s", addr)
				return
			}

			cplogger.Errorf("listener closed for address %s with error %v", addr, err)
			continue
		}

		msg := &proto.ControlMessage{
			Action:         proto.ActionProxy,
			ForwardedProto: l.Addr().Network(),
		}

		msg.ForwardedId = pname
		msg.ForwardedHost = l.Addr().String()
		err = tunnel.KeepAlive(conn)

		cpclogger := cplogger.WithFields(log.Fields{"ctrl-msg": msg})
		if err != nil {
			cpclogger.Error("TCP keepalive for tunneled connection failed", err)
		}

		go func() {
			if err := s.proxyConn(cname, conn, msg); err != nil {
				cpclogger.Error("proxy error", err)
			}
		}()
	}
}

// ServeHTTP proxies http connection to the client.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := s.RoundTrip(r)
	if err == errUnauthorised {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"User Visible Realm\"")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		s.logger.WithFields(log.Fields{
			"addr": r.RemoteAddr,
			"host": r.Host,
			"url":  r.URL,
		}).Error("round trip failed", err)

		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	transfer(w, resp.Body, s.logger.WithFields(log.Fields{
		"dir": "client to user",
		"dst": r.RemoteAddr,
		"src": r.Host,
	}))
}

// RoundTrip is http.RoundTriper implementation.
func (s *Server) RoundTrip(r *http.Request) (*http.Response, error) {
	identifier, auth, ok := s.Subscriber(r.Host)
	if !ok {
		return nil, errClientNotSubscribed
	}

	outr := r.WithContext(r.Context())
	if r.ContentLength == 0 {
		outr.Body = nil // Issue 16036: nil Body for http.Transport retries
	}
	outr.Header = cloneHeader(r.Header)

	if auth != nil {
		token := r.Header.Get("X-Auth-Header")
		if auth.Token != token {
			return nil, errUnauthorised
		}
		outr.Header.Del("X-Auth-Header")
	}

	setXForwardedFor(outr.Header, r.RemoteAddr)

	scheme := r.URL.Scheme
	if scheme == "" {
		if r.TLS != nil {
			scheme = proto.HTTPS
		} else {
			scheme = proto.HTTP
		}
	}
	if r.Header.Get("X-Forwarded-Host") == "" {
		outr.Header.Set("X-Forwarded-Host", r.Host)
		outr.Header.Set("X-Forwarded-Proto", scheme)
	}

	msg := &proto.ControlMessage{
		Action:         proto.ActionProxy,
		ForwardedHost:  r.Host,
		ForwardedProto: scheme,
	}

	return s.proxyHTTP(identifier, outr, msg)
}

func (s *Server) proxyConn(identifier string, conn net.Conn, msg *proto.ControlMessage) error {
	s.logger.WithFields(log.Fields{
		"identifier": identifier,
		"ctrlMsg":    msg,
	}).Debug("proxy connection")

	defer conn.Close()

	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	req, err := s.connectRequest(identifier, msg, pr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		transfer(pw, conn, log.WithContext(s.logger.Context).WithFields(log.Fields{
			"dir": "user to client",
			"dst": identifier,
			"src": conn.RemoteAddr(),
		}))
		cancel()
		close(done)
	}()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("io error: %s", err)
	}
	defer resp.Body.Close()

	transfer(conn, resp.Body, log.WithContext(s.logger.Context).WithFields(log.Fields{
		"dir": "client to user",
		"dst": conn.RemoteAddr(),
		"src": identifier,
	}))

	select {
	case <-done:
	case <-time.After(tunnel.DefaultTimeout):
	}

	s.logger.WithFields(log.Fields{
		"identifier": identifier,
		"ctrlMsg":    msg,
	}).Debug("proxy connection done")

	return nil
}

func (s *Server) proxyHTTP(identifier string, r *http.Request, msg *proto.ControlMessage) (*http.Response, error) {
	s.logger.WithFields(log.Fields{
		"identifier": identifier,
		"ctrlMsg":    msg,
	}).Info("proxy HTTP request")

	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	req, err := s.connectRequest(identifier, msg, pr)
	if err != nil {
		return nil, fmt.Errorf("proxy request error: %s", err)
	}

	go func() {
		cw := &countWriter{pw, 0}
		err := r.Write(cw)
		if err != nil {
			s.logger.WithFields(log.Fields{
				"identifier": identifier,
				"ctrlMsg":    msg,
			}).Error("proxy error", err)
		}

		s.logger.WithFields(log.Fields{
			"identifier": identifier,
			"bytes":      cw.count,
			"dir":        "user to client",
			"dst":        r.Host,
			"src":        r.RemoteAddr,
		}).Info("transferred")

		if r.Body != nil {
			r.Body.Close()
		}
	}()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("io error: %s", err)
	}

	s.logger.WithFields(log.Fields{
		"identifier":  identifier,
		"ctrlMsg":     msg,
		"status code": resp.StatusCode,
	}).Info("proxy HTTP done")

	return resp, nil
}

// connectRequest creates HTTP request to client with a given identifier having
// control message and data input stream, output data stream results from
// response the created request.
func (s *Server) connectRequest(cname string, msg *proto.ControlMessage, r io.Reader) (*http.Request, error) {
	conid := s.registry.GetID(cname)
	if conid == "" {
		return nil, errors.New("could not create request: ID not found")
	}
	req, err := http.NewRequest(http.MethodPut, s.connPool.URL(conid), r)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %s", err)
	}
	msg.WriteToHeader(req.Header)

	return req, nil
}

// Addr returns network address clients connect to.
func (s *Server) Addr() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// Stop closes the server.
func (s *Server) Stop() {
	s.logger.Info("stop http-tunnel server")

	if s.listener != nil {
		s.listener.Close()
	}
}
