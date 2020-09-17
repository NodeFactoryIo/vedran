package loadbalancer

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/router"
	"github.com/asdine/storm/v3"
	"log"
	"net/http"
)

type Properties struct {
	AuthSecret string
	Name string
	Capacity int64
	Whitelist []string
	Fee float32
	Selection string
}

func StartLoadBalancerServer(props Properties, port string) {
	// set auth secret
	err := auth.SetAuthSecret(props.AuthSecret)
	if err != nil {
		// terminate app: no auth secret provided
		log.Fatal(fmt.Sprintf("Unable to start vedran load balancer: %v", err))
	}

	// init database
	database, err := storm.Open("vedran-load-balancer.db")
	if err != nil {
		// terminate app: unable to start database connection
		log.Fatal(fmt.Sprintf("Unable to start vedran load balancer: %v", err))
	}

	// start server
	log.Println("Starting vedran load balancer on port :4000...")
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), router.CreateNewApiRouter(database))
	if err != nil {
		log.Print(err)
	}

	// close database connection
	err = database.Close()
	log.Fatal(err)
}