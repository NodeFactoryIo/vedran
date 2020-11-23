package router

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/gorilla/mux"
)

func createRoute(route string, method string, handler http.HandlerFunc, router *mux.Router, authorized bool) {
	var r *mux.Route
	if authorized {
		r = router.Handle(route, auth.AuthMiddleware(handler))
	} else {
		r = router.Handle(route, handler)
	}
	r.Methods(method)
	r.Name(route)
	log.Debugf("Created route %s\t%s", method, route)
}

func createRoutes(apiController *controllers.ApiController, router *mux.Router) {
	createRoute("/", "POST", apiController.RPCHandler, router, false)
	createRoute("/ws", "GET", apiController.WSHandler, router, false)

	createRoute("/api/v1/nodes", "POST", apiController.RegisterHandler, router, false)
	createRoute("/api/v1/nodes/pings", "POST", apiController.PingHandler, router, true)
	createRoute("/api/v1/nodes/metrics", "PUT", apiController.SaveMetricsHandler, router, true)
	createRoute("/api/v1/stats", "GET", apiController.StatisticsHandlerAllStats, router, false)
	createRoute("/api/v1/stats/node/{id}", "GET", apiController.StatisticsHandlerStatsForNode, router, false)
}
