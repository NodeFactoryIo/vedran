package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	"github.com/NodeFactoryIo/vedran/internal/stats"

	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	muxhelpper "github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var getNow = time.Now

type StatsResponse struct {
	Stats map[string]models.NodeStatsDetails `json:"stats"`
}

func (c *ApiController) StatisticsHandlerAllStats(w http.ResponseWriter, r *http.Request) {
	statistics, err := payout.GetStatsForPayout(c.repositories, getNow(), false)
	if err != nil {
		log.Errorf("Failed to calculate statistics, because %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(StatsResponse{
		Stats: statistics,
	})
}

type LoadbalancerStatsResponse struct {
	Stats map[string]models.NodeStatsDetails `json:"stats"`
	Fee   float32                            `json:"fee"`
}

func (c *ApiController) StatisticsHandlerAllStatsForLoadbalancer(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("X-Signature")
	if sig == "" {
		log.Errorf("Missing signature header")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	verified, err := signature.Verify([]byte(GetStatsSignedData()), []byte(sig), configuration.Config.PrivateKey)
	if err != nil {
		log.Errorf("Failed to verify signature, because %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !verified {
		log.Errorf("Invalid request signature")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	statistics, err := payout.GetStatsForPayout(c.repositories, getNow(), true)
	if err != nil {
		log.Errorf("Failed to calculate statistics, because %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(LoadbalancerStatsResponse{
		Stats: statistics,
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
