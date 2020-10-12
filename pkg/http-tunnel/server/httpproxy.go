// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package server

import (
	"bufio"
	"context"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/proto"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
)

// HTTPProxy forwards HTTP traffic.
type HTTPProxy struct {
	httputil.ReverseProxy
	// localURL specifies default base URL of local service.
	localURL *url.URL
	// localURLMap specifies mapping from ControlMessage.ForwardedHost to
	// local service URL, keys may contain host and port, only host or
	// only port. The order of precedence is the following
	// * host and port
	// * port
	// * host
	localURLMap map[string]*url.URL
	// logger is the proxy logger.
	logger *log.Entry
}

// NewHTTPProxy creates a new direct HTTPProxy, everything will be proxied to
// localURL.
func NewHTTPProxy(localURL *url.URL, logger *log.Entry) *HTTPProxy {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}

	p := &HTTPProxy{
		localURL: localURL,
		logger:   logger,
	}
	p.ReverseProxy.Director = p.Director

	return p
}

// NewMultiHTTPProxy creates a new dispatching HTTPProxy, requests may go to
// different backends based on localURLMap.
func NewMultiHTTPProxy(localURLMap map[string]*url.URL, logger *log.Entry) *HTTPProxy {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}

	p := &HTTPProxy{
		localURLMap: localURLMap,
		logger:      logger,
	}
	p.ReverseProxy.Director = p.Director

	return p
}

// Proxy is a ProxyFunc.
func (p *HTTPProxy) Proxy(w io.Writer, r io.ReadCloser, msg *proto.ControlMessage) {
	clogger := p.logger.WithFields(log.Fields{"ctrlMsg": msg})
	switch msg.ForwardedProto {
	case proto.HTTP, proto.HTTPS:
		// ok
	default:
		clogger.Error("unsupported protocol")
		return
	}

	rw, ok := w.(http.ResponseWriter)
	if !ok {
		clogger.Error("expected http.ResponseWriter")
	}

	req, err := http.ReadRequest(bufio.NewReader(r))
	if err != nil {
		clogger.Error("failed to read request", err)
		return
	}

	setXForwardedFor(req.Header, msg.RemoteAddr)
	req.URL.Host = msg.ForwardedHost

	p.ServeHTTP(rw, req)
}

// Director is ReverseProxy Director it changes request URL so that the request
// is correctly routed based on localURL and localURLMap. If no URL can be found
// the request is canceled.
func (p *HTTPProxy) Director(req *http.Request) {
	orig := *req.URL

	target := p.localURLFor(req.URL)
	if target == nil {
		p.logger.Debugf("no target for %v", req.URL)

		_, cancel := context.WithCancel(req.Context())
		cancel()

		return
	}

	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)

	targetQuery := target.RawQuery
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}

	req.Host = req.URL.Host

	p.logger.WithFields(log.Fields{
		"from": &orig,
		"to":   req.URL,
	}).Debug("url rewrite")
}

func singleJoiningSlash(a, b string) string {
	if a == "" || a == "/" {
		return b
	}
	if b == "" || b == "/" {
		return a
	}

	return path.Join(a, b)
}

func (p *HTTPProxy) localURLFor(u *url.URL) *url.URL {
	if len(p.localURLMap) == 0 {
		return p.localURL
	}

	// try host and port
	hostPort := u.Host
	if addr := p.localURLMap[hostPort]; addr != nil {
		return addr
	}

	// try port
	host, port, _ := net.SplitHostPort(hostPort)
	if addr := p.localURLMap[port]; addr != nil {
		return addr
	}

	// try host
	if addr := p.localURLMap[host]; addr != nil {
		return addr
	}

	return p.localURL
}
