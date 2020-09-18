package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"net/http"
	"time"
)

func (c ApiController) PingHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("node-id").(string)
	timestamp := r.Context().Value("timestamp").(time.Time)
	err := c.pingRepo.Save(&models.Ping{
		NodeId:    id,
		Timestamp: timestamp,
	})
	if err != nil {
		// todo handle
	}
}