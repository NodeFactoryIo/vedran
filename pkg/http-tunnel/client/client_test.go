// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package client

// TODO

//func TestClient_Dial(t *testing.T) {
//	t.Parallel()
//
//	c, err := newClient(&clientData{
//		serverAddr: "8.8.8.8:5223",
//		tunnels: map[string]*proto.Tunnel{"test": {}},
//		proxy:   Proxy(ProxyFuncs{}),
//		logger:  log.NewEntry(log.StandardLogger()),
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	conn, err := c.dial()
//	if err != nil {
//		t.Fatal("Dial error", err)
//	}
//	if conn == nil {
//		t.Fatal("Expected connection", err)
//	}
//	conn.Close()
//}

//func TestClient_DialBackoff(t *testing.T) {
//	t.Parallel()
//
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	b := tunnelmock.NewMockBackoff(ctrl)
//	gomock.InOrder(
//		b.EXPECT().NextBackOff().Return(50*time.Millisecond).Times(2),
//		b.EXPECT().NextBackOff().Return(-time.Millisecond),
//	)
//
//	c, err := newClient(&clientData{
//		serverAddr:      "8.8.8.8:10000",
//		backoff:         b,
//		tunnels:         map[string]*proto.Tunnel{"test": {}},
//		proxy:           Proxy(ProxyFuncs{}),
//		logger:          log.NewEntry(log.StandardLogger()),
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	start := time.Now()
//	_, err = c.dial()
//
//	if time.Since(start) < 100*time.Millisecond {
//		t.Fatal("Wait mismatch", err)
//	}
//
//	if err.Error() != "backoff limit exeded: foobar" {
//		t.Fatal("Error mismatch", err)
//	}
//}
