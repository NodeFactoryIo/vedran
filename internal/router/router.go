package router

import (
	"log"

	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
)

func CreateNewApiRouter(db *storm.DB, whitelistEnabled bool) *mux.Router {
	router := mux.NewRouter()

	// initialize repos
	nodeRepo := repositories.NewNodeRepo(db)
	err := nodeRepo.InitNodeRepo()
	if err != nil {
		log.Fatalf("Failed initializing node repo because of: %v", err)
	}
	pingRepo := repositories.NewPingRepo(db)
	metricsRepo := repositories.NewMetricsRepo(db)
	recordRepo := repositories.NewRecordRepo(db)

	// initialize controllers
	apiController := controllers.NewApiController(
		whitelistEnabled,
		nodeRepo,
		pingRepo,
		metricsRepo,
		recordRepo,
	)

	createRoutes(apiController, router)

	return router
}
