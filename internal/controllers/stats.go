package controllers

import (
	"encoding/json"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/stats"
	muxhelpper "github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type StatsResponse struct {
	Stats map[string]models.NodeStatsDetails `json:"stats"`
}

func (c *ApiController) StatisticsHandlerAllStats(w http.ResponseWriter, r *http.Request) {
	statisticsForPayout, err := stats.CalculateStatisticsFromLastPayout(c.repositories)
	if err != nil {
		log.Errorf("Failed to calculate statistics, because %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(StatsResponse{
		Stats: statisticsForPayout,
	})
}

func (c *ApiController) StatisticsHandlerStatsForNode(w http.ResponseWriter, r *http.Request) {
	vars := muxhelpper.Vars(r)
	nodeId, ok := vars["id"]
	if !ok || len(nodeId) < 1 {
		log.Error("Missing URL parameter node id")
		http.NotFound(w, r)
		return
	}

	nodeStatisticsFromLastPayout, err := stats.CalculateNodeStatisticsFromLastPayout(c.repositories, nodeId)
	if err != nil {
		log.Errorf("Failed to calculate statistics for node %s, because %v", nodeId, err)
		if err.Error() == "not found" {
			http.NotFound(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(nodeStatisticsFromLastPayout)
}
