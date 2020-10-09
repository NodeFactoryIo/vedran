// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package client

import (
	"crypto/tls"
	"errors"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/proto"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/tunnelmock"
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestClient_Dial(t *testing.T) {
	t.Parallel()

	s := httptest.NewTLSServer(nil)
	defer s.Close()

	c, err := newClient(&clientData{
		serverAddr: s.Listener.Addr().String(),
		tlsClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		tunnels: map[string]*proto.Tunnel{"test": {}},
		proxy:   Proxy(ProxyFuncs{}),
	})
	if err != nil {
		t.Fatal(err)
	}

	conn, err := c.dial()
	if err != nil {
		t.Fatal("Dial error", err)
	}
	if conn == nil {
		t.Fatal("Expected connection", err)
	}
	conn.Close()
}

func TestClient_DialBackoff(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := tunnelmock.NewMockBackoff(ctrl)
	gomock.InOrder(
		b.EXPECT().NextBackOff().Return(50*time.Millisecond).Times(2),
		b.EXPECT().NextBackOff().Return(-time.Millisecond),
	)

	d := func(network, addr string, config *tls.Config) (net.Conn, error) {
		return nil, errors.New("foobar")
	}

	c, err := newClient(&clientData{
		serverAddr:      "8.8.8.8",
		tlsClientConfig: &tls.Config{},
		dialTLS:         d,
		backoff:         b,
		tunnels:         map[string]*proto.Tunnel{"test": {}},
		proxy:           Proxy(ProxyFuncs{}),
	})
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	_, err = c.dial()

	if time.Since(start) < 100*time.Millisecond {
		t.Fatal("Wait mismatch", err)
	}

	if err.Error() != "backoff limit exeded: foobar" {
		t.Fatal("Error mismatch", err)
	}
}
