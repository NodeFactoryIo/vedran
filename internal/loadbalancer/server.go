package loadbalancer

import (
	"fmt"
	"net/http"

	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/router"
	"github.com/asdine/storm/v3"
	log "github.com/sirupsen/logrus"
)

func StartLoadBalancerServer(props configuration.Configuration) {
	configuration.Config = props

	// set auth secret
	err := auth.SetAuthSecret(props.AuthSecret)
	if err != nil {
		// terminate app: no auth secret provided
		log.Fatalf("Unable to start vedran load balancer: %v", err)
	}

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
