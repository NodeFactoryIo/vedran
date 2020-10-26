package router

import (
	"github.com/NodeFactoryIo/vedran/internal/actions"

	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(repos repositories.Repos, whitelistEnabled bool) *mux.Router {
	router := mux.NewRouter()

	// initialize controllers
	apiController := controllers.NewApiController(
		whitelistEnabled,
		repos,
		actions.NewActions(),
	)

	createRoutes(apiController, router)

	return router
}
