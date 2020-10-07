// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package tunnel

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunel/proto"
	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func transfer(dst io.Writer, src io.Reader, logger *log.Entry) {
	n, err := io.Copy(dst, src)
	if err != nil {
		if !strings.Contains(err.Error(), "context canceled") && !strings.Contains(err.Error(), "CANCEL") {
			logger.Error("copy error", err)
		}
	}
	log.Debugf("transferred %d bytes", n)
}

func setXForwardedFor(h http.Header, remoteAddr string) {
	clientIP, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := h["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		h.Set("X-Forwarded-For", clientIP)
	}
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

func copyHeader(dst, src http.Header) {
	for k, v := range src {
		vv := make([]string, len(v))
		copy(vv, v)
		dst[k] = vv
	}
}

type countWriter struct {
	w     io.Writer
	count int64
}

func (cw *countWriter) Write(p []byte) (n int, err error) {
	n, err = cw.w.Write(p)
	cw.count += int64(n)
	return
}

type flushWriter struct {
	w io.Writer
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

func TlsServerConfig(tlsCrt string, tlsKey string, rootCA string) (*tls.Config, error) {
	// load certs
	cert, err := tls.LoadX509KeyPair(tlsCrt, tlsKey)
	if err != nil {
		return nil, err
	}

	// load root CA for client authentication
	clientAuth := tls.RequireAnyClientCert
	var roots *x509.CertPool
	if rootCA != "" {
		roots = x509.NewCertPool()
		rootPEM, err := ioutil.ReadFile(rootCA)
		if err != nil {
			return nil, err
		}
		if ok := roots.AppendCertsFromPEM(rootPEM); !ok {
			return nil, err
		}
		clientAuth = tls.RequireAndVerifyClientCert
	}

	return &tls.Config{
		Certificates:           []tls.Certificate{cert},
		ClientAuth:             clientAuth,
		ClientCAs:              roots,
		SessionTicketsDisabled: true,
		MinVersion:             tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		PreferServerCipherSuites: true,
		NextProtos:               []string{"h2"},
	}, nil
}

func TlsClientConfig(TLSCrt string, TLSKey string, RootCA string, ServerAddr string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(TLSCrt, TLSKey)
	if err != nil {
		return nil, err
	}

	var roots *x509.CertPool
	if RootCA != "" {
		roots = x509.NewCertPool()
		rootPEM, err := ioutil.ReadFile(RootCA)
		if err != nil {
			return nil, err
		}
		if ok := roots.AppendCertsFromPEM(rootPEM); !ok {
			return nil, err
		}
	}

	host, _, err := net.SplitHostPort(ServerAddr)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		ServerName:         host,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: roots == nil,
		RootCAs:            roots,
	}, nil
}

// BackoffConfig defines behavior of staggering reconnection retries.
type BackoffConfig struct {
	Interval    time.Duration `yaml:"interval"`
	Multiplier  float64       `yaml:"multiplier"`
	MaxInterval time.Duration `yaml:"max_interval"`
	MaxTime     time.Duration `yaml:"max_time"`
}

func ExpBackoff(c BackoffConfig) *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = c.Interval
	b.Multiplier = c.Multiplier
	b.MaxInterval = c.MaxInterval
	b.MaxElapsedTime = c.MaxTime

	return b
}

type Tunnel struct {
	Protocol   string `yaml:"proto,omitempty"`
	Addr       string `yaml:"addr,omitempty"`
	Auth       string `yaml:"auth,omitempty"`
	Host       string `yaml:"host,omitempty"`
	RemoteAddr string `yaml:"remote_addr,omitempty"`
}

func MapTunnels(m map[string]*Tunnel) map[string]*proto.Tunnel {
	p := make(map[string]*proto.Tunnel)

	for name, t := range m {
		p[name] = &proto.Tunnel{
			Protocol: t.Protocol,
			Host:     t.Host,
			Auth:     t.Auth,
			Addr:     t.RemoteAddr,
		}
	}

	return p
}

func CreateProxy(m map[string]*Tunnel, logger *log.Entry) ProxyFunc {
	httpURL := make(map[string]*url.URL)
	tcpAddr := make(map[string]string)

	for v, t := range m {
		fmt.Printf("Creating proxy for %#+v/%#+v\n", v, t)
		switch t.Protocol {
		case proto.HTTP:
			u, err := url.Parse(t.Addr)
			if err != nil {
				logger.Fatal("invalid tunnel address: %s", err)
			}
			httpURL[t.Host] = u
		case proto.TCP, proto.TCP4, proto.TCP6:
			fmt.Printf("Setting config for %s | REMOTE: %s | LOCAL: %s\n", v, t.RemoteAddr, t.Addr)
			tcpAddr[v] = t.Addr
		case proto.SNI:
			tcpAddr[t.Host] = t.Addr
		}
	}

	return Proxy(ProxyFuncs{
		HTTP: NewMultiHTTPProxy(httpURL, logger.WithField("proxy", "HTTP")).Proxy,
		TCP:  NewMultiTCPProxy(tcpAddr, logger.WithField("proxy", "TCP")).Proxy,
	})
}



