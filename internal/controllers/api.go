package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
)

type ApiController struct {
	nodeRepo models.NodeRepository
}

func NewApiController(nodeRepo models.NodeRepository) *ApiController {
	return &ApiController{
		nodeRepo: nodeRepo,
	}
}
