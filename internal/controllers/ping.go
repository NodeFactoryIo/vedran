package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"net/http"
	"time"
)

func (c ApiController) PingHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("node-id").(string)
	err := c.pingRepo.Save(&models.Ping{
		NodeId:    id,
		Timestamp: time.Now(),
	})
	if err != nil {
		// todo handle
	}
}