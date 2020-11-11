package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/stats"
	"math"
	"net/http"

	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/models"
	log "github.com/sirupsen/logrus"
)

const pingOffset = 5

func (c ApiController) PingHandler(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(auth.RequestContextKey).(*auth.RequestContext)

	lastPingTime, downtimeDuration, err := c.repositories.PingRepo.CalculateDowntime(request.NodeId, request.Timestamp)
	if err != nil {
		log.Errorf("Unable to calculate node downtime, error: %v", err)
	}

	if math.Abs(downtimeDuration.Seconds()) > (stats.PingIntervalInSeconds + pingOffset) {
		downtime := models.Downtime{
			Start:  lastPingTime,
			End:    request.Timestamp,
			NodeId: request.NodeId,
		}
		err = c.repositories.DowntimeRepo.Save(&downtime)
		if err != nil {
			log.Errorf("Unable to save node downtime, error: %v", err)
		}

		log.Debugf("Saved node %s downtime of: %f", request.NodeId, math.Abs(downtimeDuration.Seconds()))
	}

	// save ping to database
	ping := models.Ping{
		NodeId:    request.NodeId,
		Timestamp: request.Timestamp,
	}
	err = c.repositories.PingRepo.Save(&ping)
	if err != nil {
		log.Errorf("Unable to save ping %v to database, error: %v", ping, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Debugf("Ping from node %s", ping.NodeId)
}
