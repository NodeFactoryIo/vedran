package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
)

type ApiController struct {
	whitelistEnabled bool
	nodeRepo         models.NodeRepository
	pingRepo         models.PingRepository
	metricsRepo      models.MetricsRepository
	downtimeRepo     models.DowntimeRepository
}

func NewApiController(
	whitelistEnabled bool,
	nodeRepo models.NodeRepository,
	pingRepo models.PingRepository,
	metricsRepo models.MetricsRepository,
	downtimeRepo models.DowntimeRepository,
) *ApiController {
	return &ApiController{
		whitelistEnabled: whitelistEnabled,
		nodeRepo:         nodeRepo,
		pingRepo:         pingRepo,
		metricsRepo:      metricsRepo,
		downtimeRepo:     downtimeRepo,
	}
}
