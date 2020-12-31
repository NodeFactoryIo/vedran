package router

import (
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(controller *controllers.ApiController, privateKey string) *mux.Router {
	router := mux.NewRouter()
	createRoutes(controller, router, privateKey)
	return router
}
