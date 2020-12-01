package router

import (
	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/prometheus"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(controller *controllers.ApiController) *mux.Router {
	router := mux.NewRouter()
	createRoutes(controller, router)

	// initialize controllers
	apiController := controllers.NewApiController(
		whitelistEnabled,
		repos,
		actions.NewActions(),
	)

	prometheus.RecordMetrics(repos)
	createRoutes(apiController, router)

	return router
}
