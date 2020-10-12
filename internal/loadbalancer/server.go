package loadbalancer

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/router"
	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/server"
	"github.com/asdine/storm/v3"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func StartLoadBalancerServer(props configuration.Configuration) {
	configuration.Config = props

	// set auth secret
	err := auth.SetAuthSecret(props.AuthSecret)
	if err != nil {
		// terminate app: no auth secret provided
		log.Fatalf("Unable to start vedran load balancer: %v", err)
	}
	// ---------------------------------------------
	l := log.New()
	l.SetLevel(log.DebugLevel)
	s, err := server.NewServer(&server.ServerConfig{
		Address:        ":5223",
		PortRange:      "10000:50000",
		AuthHandler: func(s string) bool {
			token, err := jwt.ParseWithClaims(s, &auth.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte("authsecret"), nil
			})

			if err == nil {
				if _, ok := token.Claims.(*auth.CustomClaims); ok && token.Valid {
					return true
				}
			}

			return false
		},
		Logger: log.NewEntry(l),
	})
	if err != nil {
		log.Error("SERVER ", err)
	}
	go s.Start()
	// ---------------------------------------------
	// init database
	database, err := storm.Open("vedran-load-balancer.db")
	if err != nil {
		// terminate app: unable to start database connection
		log.Fatalf("Unable to start vedran load balancer: %v", err)
	}
	log.Debug("Successfully connected to database")

	whitelistEnabled := len(props.Whitelist) > 0
	// save whitelisted id-s
	if whitelistEnabled {
		log.Debugf("Whitelisting enabled, whitelisted node IDs: %v", props.Whitelist)
		for _, nodeId := range props.Whitelist {
			err = database.Set(models.WhitelistBucket, nodeId, true)
			if err != nil {
				// terminate app: unable to save whitelist id-s
				log.Fatalf("Unable to start vedran load balancer: %v", err)
			}
		}
	} else {
		log.Debug("Whitelisting disabled")
	}

	// start server
	log.Infof("Starting vedran load balancer on port :%d...", props.Port)
	r := router.CreateNewApiRouter(database, whitelistEnabled)
	err = http.ListenAndServe(fmt.Sprintf(":%d", props.Port), r)
	if err != nil {
		log.Error(err)
	}

	// close database connection
	err = database.Close()
	log.Error(err)
}
