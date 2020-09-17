package router

import (
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(db *storm.DB) *mux.Router {
	router := mux.NewRouter()
	// initialize repos
	nodeRepo := repositories.NewNodeRepo(db)
	// initialize controllers
	apiController := controllers.NewApiController(nodeRepo)
	// map controllers handlers to endpoints
	router.HandleFunc("/api/v1/nodes", apiController.RegisterHandler).Methods("POST").Name("/api/v1/nodes")

	return router
}
