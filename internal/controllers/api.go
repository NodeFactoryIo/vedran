package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
)

type ApiController struct {
	nodeRepo    models.NodeRepository
	pingRepo    models.PingRepository
	metricsRepo models.MetricsRepository
}

func NewApiController(
	nodeRepo models.NodeRepository,
	pingRepo models.PingRepository,
	metricsRepo models.MetricsRepository,
) *ApiController {
	return &ApiController{
		nodeRepo: nodeRepo,
		pingRepo: pingRepo,
		metricsRepo: metricsRepo,
	}
}
