package router

import (
	"github.com/NodeFactoryIo/vedran/internal/controlers"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(db *storm.DB) *mux.Router {
	router := mux.NewRouter()
	// initialize repos
	nodeRepo := repositories.NewNodeRepo(db)
	// initialize controllers
	baseController := controlers.NewBaseController(nodeRepo)
	// map controllers handlers to endpoints
	router.HandleFunc("/api/v1/nodes", baseController.RegisterHandler).Methods("POST").Name("/api/v1/nodes")

	return router
}
