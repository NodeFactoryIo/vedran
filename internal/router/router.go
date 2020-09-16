package router

import (
	"github.com/NodeFactoryIo/vedran/internal/handlers"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter() *mux.Router {
	router := mux.NewRouter()

	// define api
	router.HandleFunc("/api/v1/nodes", handlers.RegisterHandler).Methods("POST").Name("/api/v1/nodes")

	return router
}
