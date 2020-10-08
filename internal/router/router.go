package router

import (
	"net/http"

	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(db *storm.DB, whitelistEnabled bool) *mux.Router {
	router := mux.NewRouter()
	// initialize repos
	nodeRepo, err := repositories.NewNodeRepo(db)
	if err != nil {
		panic(err)
	}

	pingRepo := repositories.NewPingRepo(db)
	metricsRepo := repositories.NewMetricsRepo(db)
	// initialize controllers
	apiController := controllers.NewApiController(whitelistEnabled, nodeRepo, pingRepo, metricsRepo)
	// map controllers handlers to endpoints
	createRoute("/", "POST", apiController.RPCHandler, router, false)

	createRoute("/api/v1/nodes", "POST", apiController.RegisterHandler, router, false)
	createRoute("/api/v1/nodes/pings", "POST", apiController.PingHandler, router, true)
	createRoute("/api/v1/nodes/metrics", "PUT", apiController.SaveMetricsHandler, router, true)
	return router
}

func createRoute(route string, method string, handler http.HandlerFunc, router *mux.Router, authorized bool) {
	var r *mux.Route
	if authorized {
		r = router.Handle(route, auth.AuthMiddleware(handler))
	} else {
		r = router.Handle(route, handler)
	}
	r.Methods(method)
	r.Name(route)
}
