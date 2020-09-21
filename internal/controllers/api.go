package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
)

type ApiController struct {
	nodeRepo models.NodeRepository
	pingRepo models.PingRepository
}

func NewApiController(nodeRepo models.NodeRepository, pingRepo models.PingRepository) *ApiController {
	return &ApiController{
		nodeRepo: nodeRepo,
		pingRepo: pingRepo,
	}
}
