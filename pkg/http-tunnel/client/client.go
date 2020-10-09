// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/proto"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

const (
	DefaultBackoffInterval    = 500 * time.Millisecond
	DefaultBackoffMultiplier  = 1.5
	DefaultBackoffMaxInterval = 60 * time.Second
	DefaultBackoffMaxTime     = 15 * time.Minute
)

// Client is responsible for creating connection to the server, handling control
// messages. It uses ProxyFunc for transferring data between server and local
// services.
type Client struct {
	config *clientData

	conn           net.Conn
	connMu         sync.Mutex
	httpServer     *http2.Server
	serverErr      error
	lastDisconnect time.Time
	logger         *log.Entry
}

// ClientConfig is configuration of the Client.
type ClientConfig struct {
	// TlsCrtFilePath specifies path to certificate file for tls connection
	TlsCrtFilePath string
	// TlsKeyFilePath specifies path to key file for tls connection
	TlsKeyFilePath string
	// ServerAddress specifies TCP address of the tunnel server.
	ServerAddress string
	// Tunnels specifies the tunnels client requests to be opened on server.
	Tunnels map[string]*Tunnel
	// Logger is optional logger. If nil logging is disabled.
	Logger *log.Entry
	// AuthToken
	AuthToken string
}

type clientData struct {
	serverAddr      string
	tlsClientConfig *tls.Config
	// dialTLS specifies an optional dial function that creates a tls
	// connection to the server. If dialTLS is nil, tls.Dial is used.
	dialTLS func(network, addr string, config *tls.Config) (net.Conn, error)
	// backoff specifies backoff policy on server connection retry. If nil
	// when dial fails it will not be retried.
	backoff   Backoff
	tunnels   map[string]*proto.Tunnel
	proxy     ProxyFunc
	logger    *log.Entry
	idName    string
	authToken string
}

// NewClient creates a new unconnected Client based on configuration. Caller
// must invoke Start() on returned instance in order to connect server.
func NewClient(config *ClientConfig) (*Client, error) {
	clientData := &clientData{}

	if config.ServerAddress == "" {
		return nil, errors.New("provided empty server address")
	}
	clientData.serverAddr = config.ServerAddress

	if config.AuthToken == "" {
		return nil, errors.New("provided empty auth token")
	}
	clientData.authToken = config.AuthToken

	if config.TlsKeyFilePath == "" {
		return nil, errors.New("provided tls key file path empty")
	}
	if config.TlsCrtFilePath == "" {
		return nil, errors.New("provided tls cert file path empty")
	}
	tlsconf, err := TlsClientConfig(
		config.TlsCrtFilePath,
		config.TlsKeyFilePath,
		"",
		config.ServerAddress)
	if err != nil {
		log.Error("TLS ", err)
	}
	clientData.tlsClientConfig = tlsconf

	logger := config.Logger
	if logger == nil {
		l := log.New()
		l.SetLevel(log.ErrorLevel)
		logger = log.NewEntry(l)
	}
	clientData.logger = logger

	if config.Tunnels == nil {
		return nil, errors.New("tunnels maping is nil")
	}
	clientData.tunnels = MapTunnels(config.Tunnels)
	clientData.proxy = CreateProxy(config.Tunnels, logger)

	clientData.backoff = ExpBackoff(BackoffConfig{
		Interval:    DefaultBackoffInterval,
		Multiplier:  DefaultBackoffMultiplier,
		MaxInterval: DefaultBackoffMaxInterval,
		MaxTime:     DefaultBackoffMaxTime,
	})

	clientData.idName = "" // TODO

	return newClient(clientData)
}

func newClient(config *clientData) (*Client, error) {
	if config.serverAddr == "" {
		return nil, errors.New("missing serverAddr")
	}
	if config.tlsClientConfig == nil {
		return nil, errors.New("missing tlsClientConfig")
	}
	if len(config.tunnels) == 0 {
		return nil, errors.New("missing tunnels")
	}
	if config.proxy == nil {
		return nil, errors.New("missing proxy")
	}

	logger := config.logger
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}

	c := &Client{
		config:     config,
		httpServer: &http2.Server{},
		logger:     logger,
	}

	return c, nil
}

// Start connects client to the server, it returns error if there is a
// connection error, or server cannot open requested tunnels. On connection
// error a backoff policy is used to reestablish the connection. When connected
// HTTP/2 server is started to handle ControlMessages.
func (c *Client) Start() error {
	c.logger.Info("start http-tunnel client")

	for {
		conn, err := c.connect()
		if err != nil {
			return err
		}

		c.httpServer.ServeConn(conn, &http2.ServeConnOpts{
			Handler: http.HandlerFunc(c.serveHTTP),
		})

		c.logger.Info("disconnected")

		c.connMu.Lock()
		now := time.Now()
		err = c.serverErr

		// detect disconnect hiccup
		if err == nil && now.Sub(c.lastDisconnect).Seconds() < 5 {
			err = fmt.Errorf("connection is being cut")
		}

		c.conn = nil
		c.serverErr = nil
		c.lastDisconnect = now
		c.connMu.Unlock()

		if err != nil {
			return err
		}
	}
}

func (c *Client) connect() (net.Conn, error) {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		return nil, fmt.Errorf("already connected")
	}

	conn, err := c.dial()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %s", err)
	}
	c.conn = conn

	return conn, nil
}

func (c *Client) dial() (net.Conn, error) {
	var (
		network   = "tcp"
		addr      = c.config.serverAddr
		tlsConfig = c.config.tlsClientConfig
	)

	doDial := func() (conn net.Conn, err error) {
		c.logger.WithFields(log.Fields{
			"network": network,
			"addr":    addr,
		}).Info("dial")

		if c.config.dialTLS != nil {
			conn, err = c.config.dialTLS(network, addr, tlsConfig)
		} else {
			d := &net.Dialer{
				Timeout: tunnel.DefaultTimeout,
			}
			conn, err = d.Dial(network, addr)

			if err == nil {
				err = tunnel.KeepAlive(conn)
			}
			if err == nil {
				conn = tls.Client(conn, tlsConfig)
			}
			if err == nil {
				err = conn.(*tls.Conn).Handshake()
			}
		}

		if err != nil {
			if conn != nil {
				conn.Close()
				conn = nil
			}

			c.logger.WithFields(log.Fields{
				"network": network,
				"addr":    addr,
			}).Error("dial failed", err)
		}

		return
	}

	b := c.config.backoff
	if b == nil {
		return doDial()
	}

	for {
		conn, err := doDial()

		// success
		if err == nil {
			b.Reset()
			return conn, err
		}

		// failure
		d := b.NextBackOff()
		if d < 0 {
			return conn, fmt.Errorf("backoff limit exeded: %s", err)
		}

		// backoff
		c.logger.Debugf("backoff for %v", d)
		time.Sleep(d)
	}
}

func (c *Client) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		if r.Header.Get(proto.HeaderError) != "" {
			c.handleHandshakeError(w, r)
		} else {
			c.handleHandshake(w, r)
		}
		return
	}

	msg, err := proto.ReadControlMessage(r)
	if err != nil {
		c.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clogger := c.logger.WithFields(log.Fields{
		"ctrlMsg": msg,
	})
	clogger.Debug("handle")

	switch msg.Action {
	case proto.ActionProxy:
		c.config.proxy(w, r.Body, msg)
	default:
		clogger.Error("unknown action")
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	clogger.Debug("done")
}

func (c *Client) handleHandshakeError(w http.ResponseWriter, r *http.Request) {
	err := fmt.Errorf(r.Header.Get(proto.HeaderError))

	c.logger.Errorf("handshake error %v for address %s", err, r.RemoteAddr)

	c.connMu.Lock()
	c.serverErr = fmt.Errorf("server error: %s", err)
	c.connMu.Unlock()
}

type TunnelExt struct {
	IdName  string
	Tunnels map[string]*proto.Tunnel
}

func (c *Client) handleHandshake(w http.ResponseWriter, r *http.Request) {
	c.logger.Infof("handshake for address %s", r.RemoteAddr)

	w.Header().Add("X-Auth-Header", c.config.authToken)
	w.WriteHeader(http.StatusOK)

	// te := TunnelExt{
	// 	idName:  "pepito",
	// 	tunnels: c.config.tunnels,
	// }

	te := TunnelExt{
		IdName:  c.config.idName,
		Tunnels: c.config.tunnels,
	}

	// b, err := json.Marshal(c.config.tunnels)
	b, err := json.Marshal(te)
	if err != nil {
		c.logger.Error("handshake failed", err)
		return
	}
	// Datadope function
	_, err = w.Write(b)
	if err != nil {
		c.logger.Error("handshake failed", err)
	}
}

// Stop disconnects client from server.
func (c *Client) Stop() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	c.logger.Info("stop http-tunnel client")

	if c.conn != nil {
		c.conn.Close()
	}
	c.conn = nil
}
