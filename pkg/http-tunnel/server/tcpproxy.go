// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"io"
	"net"

	tunnel "github.com/NodeFactoryIo/vedran/pkg/http-tunnel"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/proto"
	log "github.com/sirupsen/logrus"
)

// TCPProxy forwards TCP streams.
type TCPProxy struct {
	// localAddr specifies default TCP address of the local server.
	localAddr string
	// localAddrMap specifies mapping from ControlMessage.ForwardedHost to
	// local server address, keys may contain host and port, only host or
	// only port. The order of precedence is the following
	// * host and port
	// * port
	// * host
	localAddrMap map[string]string
	// logger is the proxy logger.
	logger *log.Entry
}

// NewTCPProxy creates new direct TCPProxy, everything will be proxied to
// localAddr.
func NewTCPProxy(localAddr string, logger *log.Entry) *TCPProxy {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}
	return &TCPProxy{
		localAddr: localAddr,
		logger:    logger,
	}
}

// NewMultiTCPProxy creates a new dispatching TCPProxy, connections may go to
// different backends based on localAddrMap.
func NewMultiTCPProxy(localAddrMap map[string]string, logger *log.Entry) *TCPProxy {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}
	logger.Debugf("Creating New TCPProxy %+v\n", localAddrMap)
	return &TCPProxy{
		localAddrMap: localAddrMap,
		logger:       logger,
	}
}

// Proxy is a ProxyFunc.
func (p *TCPProxy) Proxy(w io.Writer, r io.ReadCloser, msg *proto.ControlMessage) {
	clogger := p.logger.WithFields(log.Fields{"ctrlMsg": msg})
	switch msg.ForwardedProto {
	case proto.TCP, proto.TCP4, proto.TCP6, proto.UNIX, proto.SNI:
		// ok
	default:
		clogger.Error("unsupported protocol")
		return
	}

	target := p.localAddrFor(msg.ForwardedHost)
	if target == "" {
		target = p.localAddrFor(msg.ForwardedId)
		if target == "" {
			clogger.Error("no target")
			return
		}
	}

	local, err := net.DialTimeout("tcp", target, tunnel.DefaultTimeout)
	if err != nil {
		clogger.WithFields(log.Fields{
			"target": target,
		}).Error("dial failed", err)
		return
	}

	if err := tunnel.KeepAlive(local); err != nil {
		clogger.WithFields(log.Fields{
			"target": target,
		}).Error("TCP keepalive for tunneled connection failed", err)
	}

	go func() {
		loggerWithContext := log.WithContext(p.logger.Context).WithFields(log.Fields{
			"dst": msg.ForwardedHost,
			"src": target,
		})
		transfer(flushWriter{w}, local, loggerWithContext)
	}()

	loggerWithContext := log.WithContext(p.logger.Context).WithFields(log.Fields{
		"dst": msg.ForwardedHost,
		"src": target,
	})
	transfer(local, r, loggerWithContext)

	err = local.Close()
	if err != nil {
		clogger.Errorf("Transfer close failed because of %v")
	}
}

func (p *TCPProxy) localAddrFor(hostPort string) string {

	if len(p.localAddrMap) == 0 {
		return p.localAddr
	}

	// try hostPort
	if addr := p.localAddrMap[hostPort]; addr != "" {
		return addr
	}

	// try port
	host, port, _ := net.SplitHostPort(hostPort)
	if addr := p.localAddrMap[port]; addr != "" {
		return addr
	}

	// try 0.0.0.0:port
	if addr := p.localAddrMap[fmt.Sprintf("0.0.0.0:%s", port)]; addr != "" {
		return addr
	}

	// try host
	if addr := p.localAddrMap[host]; addr != "" {
		return addr
	}

	return p.localAddr
}
