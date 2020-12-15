package router

import (
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(controller *controllers.ApiController) *mux.Router {
	router := mux.NewRouter()
	createRoutes(controller, router)
	return router
}
