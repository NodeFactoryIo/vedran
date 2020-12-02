package controllers

import (
	"encoding/json"
	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	"github.com/NodeFactoryIo/vedran/internal/stats"
	muxhelpper "github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type StatsResponse struct {
	Stats map[string]models.NodeStatsDetails `json:"stats"`
	Fee   float32                            `json:"fee"`
}

var getNow = time.Now

func (c *ApiController) StatisticsHandlerAllStats(w http.ResponseWriter, r *http.Request) {
	// should check for signature in body and only then record payout
	payoutStatistics, err := payout.GetStatsForPayout(c.repositories, getNow(), false)
	if err != nil {
		log.Errorf("Failed to calculate statistics, because %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(StatsResponse{
		Stats: payoutStatistics,
		Fee:   configuration.Config.Fee,
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

	nodeStatisticsFromLastPayout, err := stats.CalculateNodeStatisticsFromLastPayout(c.repositories, nodeId, getNow())
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
