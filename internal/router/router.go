package router

import (
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/gorilla/mux"
)



func CreateNewApiRouter(repos repositories.Repos, whitelistEnabled bool) *mux.Router {
	router := mux.NewRouter()

	// initialize controllers
	apiController := controllers.NewApiController(
		whitelistEnabled,
		repos.NodeRepo,
		repos.PingRepo,
		repos.MetricsRepo,
		repos.RecordRepo,
	)

	createRoutes(apiController, router)

	return router
}
