package loadbalancer

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/router"
	"github.com/asdine/storm/v3"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Properties struct {
	AuthSecret string
	Name       string
	Capacity   int64
	Whitelist  []string
	Fee        float32
	Selection  string
	Port       int32
}

func StartLoadBalancerServer(props Properties) {
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

	// start server
	log.Infof("Starting vedran load balancer on port :%d...", props.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", props.Port), router.CreateNewApiRouter(database))
	if err != nil {
		log.Error(err)
	}

	// close database connection
	err = database.Close()
	log.Error(err)
}
