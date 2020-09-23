package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (c ApiController) PingHandler(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(auth.RequestContextKey).(*auth.RequestContext)

	// save ping to database
	ping := models.Ping{
		NodeId:    request.NodeId,
		Timestamp: request.Timestamp,
	}
	err := c.pingRepo.Save(&ping)
	if err != nil {
		log.Errorf("Unable to save ping %v to database, error: %v", ping, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Debugf("Ping from node %s at %s", ping.NodeId, ping.Timestamp.String())
}