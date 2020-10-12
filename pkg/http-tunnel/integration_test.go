package tunnel_test

import (
	"bytes"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/client"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/server"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func Test_IntegrationTest(t *testing.T) {
	l := log.New()
	var str bytes.Buffer
	l.SetOutput(&str)
	l.SetLevel(log.DebugLevel)
	s, err := server.NewServer(&server.ServerConfig{
		Address:        ":5223",
		PortRange:      "10000:50000",
		AuthHandler: func(s string) bool {
			return s == "test-token"
		},
		Logger: log.NewEntry(l),
	})
	assert.Nil(t, err)

	go func() {
		s.Start()
	}()

	c, err := client.NewClient(&client.ClientConfig{
		ServerAddress:  "127.0.0.1:5223",
		Tunnels: map[string]*client.Tunnel{
			"": {
				Protocol:   "tcp",
				Addr:       "localhost:3000",
				Auth:       "",
				Host:       "",
				RemoteAddr: "0.0.0.0:AUTO",
			},
		},
		Logger:    log.NewEntry(l),
		AuthToken: "test-token",
		IdName: "test-id",
	})

	assert.Nil(t, err)

	go func() {
		err := c.Start()
		assert.Nil(t, err)
	}()

	time.Sleep(2 * time.Second)

	logStr := str.String()
	// asserting that handshake was successful
	assert.True(t, strings.Contains(logStr, "msg=dial"))
	assert.True(t, strings.Contains(logStr, "msg=\"try connect\""))
	assert.True(t, strings.Contains(logStr, "msg=\"handshake for address 127.0.0.1:5223\""))
	assert.True(t, strings.Contains(logStr, "msg=\"REGISTRY SUBSCRIBE\""))
	assert.True(t, strings.Contains(logStr, "msg=\"REGISTRY SET (OLD FOUND)\""))
	assert.True(t, strings.Contains(logStr, "msg=\"REGISTRY SET (NEW SET)\""))
	assert.True(t, strings.Contains(logStr, "msg=\"test-id connected\""))

	c.Stop()
	s.Stop()
}
