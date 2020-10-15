package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/repositories"
)

type ApiController struct {
	whitelistEnabled  bool
	nodeRepository    repositories.NodeRepository
	pingRepository    repositories.PingRepository
	metricsRepository repositories.MetricsRepository
	recordRepository  repositories.RecordRepository
}

func NewApiController(
	whitelistEnabled bool,
	nodeRepository repositories.NodeRepository,
	pingRepository repositories.PingRepository,
	metricsRepository repositories.MetricsRepository,
	recordRepository repositories.RecordRepository,
) *ApiController {
	return &ApiController{
		whitelistEnabled:  whitelistEnabled,
		nodeRepository:    nodeRepository,
		pingRepository:    pingRepository,
		metricsRepository: metricsRepository,
		recordRepository:  recordRepository,
	}
}
