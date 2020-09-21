package router

import (
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
	"net/http"
)

func CreateNewApiRouter(db *storm.DB) *mux.Router {
	router := mux.NewRouter()
	// initialize repos
	nodeRepo := repositories.NewNodeRepo(db)
	pingRepo := repositories.NewPingRepo(db)
	// initialize controllers
	apiController := controllers.NewApiController(nodeRepo, pingRepo)
	// map controllers handlers to endpoints
	createRoute("/api/v1/nodes", "POST", apiController.RegisterHandler, router, false)
	createRoute("/api/v1/nodes/pings", "POST", apiController.PingHandler, router, true)
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
