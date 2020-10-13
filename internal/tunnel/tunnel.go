package tunnel

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/server"
	log "github.com/sirupsen/logrus"
)

func StartTunnelServer(serverPort string, portRange string) {
	logger := log.WithField("context", "tunnel-server")
	s, err := server.NewServer(&server.ServerConfig{
		Address:   fmt.Sprintf(":%s", serverPort),
		PortRange: portRange,
		AuthHandler: func(rawToken string) bool {
			token, err := auth.ParseJwtTokenWithCustomClaims(rawToken)
			if err == nil {
				if _, ok := token.Claims.(*auth.CustomClaims); ok && token.Valid {
					return true
				}
			}
			return false
		},
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("failed to create http tunnel server: %s", err)
	}
	// start server in new goroutine
	go s.Start()
}
