package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/proto"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/server"
	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/url"
	"time"
)

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

func CreateProxy(m map[string]*Tunnel, logger *logrus.Entry) ProxyFunc {
	httpURL := make(map[string]*url.URL)
	tcpAddr := make(map[string]string)

	for v, t := range m {
		fmt.Printf("Creating proxy for %#+v/%#+v\n", v, t)
		switch t.Protocol {
		case proto.HTTP:
			u, err := url.Parse(t.Addr)
			if err != nil {
				logger.Fatalf("invalid tunnel address: %v", err)
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
		HTTP: server.NewMultiHTTPProxy(httpURL, logger.WithField("proxy", "HTTP")).Proxy,
		TCP:  server.NewMultiTCPProxy(tcpAddr, logger.WithField("proxy", "TCP")).Proxy,
	})
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
