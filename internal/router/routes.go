package router

import (
	customMiddleware "github.com/NodeFactoryIo/vedran/internal/middleware"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"

	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/gorilla/mux"
)

func createRoutes(apiController *controllers.ApiController, router *mux.Router, privateKey string) {
	// Create a custom registry for prometheus.
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	createTrackedRoute("/", "POST", std.Handler("/", mdlw, http.HandlerFunc(apiController.RPCHandler)), router)
	createTrackedRoute("/ws", "GET", std.Handler("/ws", mdlw, http.HandlerFunc(apiController.WSHandler)), router)

	createSignatureVerificationRoute("/api/v1/stats", "POST", apiController.StatisticsHandlerAllStatsForLoadbalancer, router, privateKey)

	// authorized
	createRoute("/api/v1/nodes/pings", "POST", apiController.PingHandler, router, true)
	createRoute("/api/v1/nodes/metrics", "PUT", apiController.SaveMetricsHandler, router, true)
	// unauthorized
	createRoute("/api/v1/nodes", "POST", apiController.RegisterHandler, router, false)
	createRoute("/api/v1/stats", "GET", apiController.StatisticsHandlerAllStats, router, false)
	createRoute("/api/v1/stats/node/{id}", "GET", apiController.StatisticsHandlerStatsForNode, router, false)
	createRoute("/api/v1/stats/lb", "GET", apiController.StatisticsHandlerStatsForLoadBalancer, router, false)
	createRoute("/metrics", "GET", promhttp.Handler().ServeHTTP, router, false)
}

func createRoute(
	route string, method string, handler http.HandlerFunc, router *mux.Router, authorized bool,
) {
	var r *mux.Route
	if authorized {
		r = router.Handle(route, auth.AuthMiddleware(handler))
	} else {
		r = router.Handle(route, handler)
	}
	setUpRoute(route, method, r)
}

func createTrackedRoute(
	route string, method string, handler http.Handler, router *mux.Router,
) {
	r := router.Handle(route, handler)
	setUpRoute(route, method, r)
}

func createSignatureVerificationRoute(
	route string, method string, handler http.HandlerFunc, router *mux.Router, privateKey string,
) {
	r := router.Handle(route, customMiddleware.VerifySignatureMiddleware(handler, privateKey))
	setUpRoute(route, method, r)
}

func setUpRoute(route string, method string, r *mux.Route) {
	r.Methods(method)
	r.Name(route)
	log.Debugf("Created route %s\t%s", method, route)
}
