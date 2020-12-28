package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/stats"

	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"github.com/ethereum/go-ethereum/common/hexutil"
	muxhelpper "github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var getNow = time.Now

type StatsResponse struct {
	Stats map[string]models.NodeStatsDetails `json:"stats"`
}

// handler for `GET /api/v1/stats`
func (c *ApiController) StatisticsHandlerAllStats(w http.ResponseWriter, r *http.Request) {
	timestamp := getNow()
	statistics, err := stats.CalculateStatisticsFromLastPayout(c.repositories, timestamp)
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

type LoadbalancerStatsRequest struct {
	TotalReward string `json:"total_reward"`
}

// handler for `POST /api/v1/stats`
func (c *ApiController) StatisticsHandlerAllStatsForLoadbalancer(w http.ResponseWriter, r *http.Request) {
	verified, httpStatusCode, err := verifySignatureInHeader(r, c.privateKey)
	if err != nil {
		http.Error(w, http.StatusText(httpStatusCode), httpStatusCode)
		return
	}
	if !verified {
		log.Errorf("Invalid request signature")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	totalRewardAsFloat, err := getTotalRewardFromRequest(r)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	timestamp := getNow()
	statistics, err := stats.CalculateStatisticsFromLastPayout(c.repositories, timestamp)
	if err != nil {
		log.Errorf("Failed to calculate statistics, because %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = c.repositories.PayoutRepo.Save(&models.Payout{
		Timestamp:      timestamp,
		PaymentDetails: statistics,
		LbFee:          totalRewardAsFloat * float64(configuration.Config.Fee),
	})
	if err != nil {
		log.Errorf("Failed to save payout, because %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	feesToNodes := payout.CalculatePayoutDistributionByNode(
		statistics,
		totalRewardAsFloat,
		payout.LoadBalancerDistributionConfiguration{
			FeePercentage:       float64(configuration.Config.Fee),
			DifferentFeeAddress: false,
		},
	)
	for nodeId, amount := range feesToNodes {
		err := c.repositories.FeeRepo.RecordNewFee(nodeId, amount.Int64())
		if err != nil {
			log.Errorf("Failed to save fee for node %s, because %v", nodeId, err)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(LoadbalancerStatsResponse{
		Stats: statistics,
		Fee:   configuration.Config.Fee,
	})
}

func verifySignatureInHeader(r *http.Request, privateKey string) (bool, int, error) {
	sig := r.Header.Get("X-Signature")
	if sig == "" {
		log.Error("Missing signature header")
		return false, http.StatusBadRequest, nil
	}
	sigInBytes, err := hexutil.Decode(sig)
	if err != nil {
		log.Errorf("Unable to decode signature, because of: %v", err)
		return false, http.StatusBadRequest, err
	}
	verified, err := signature.Verify([]byte(StatsSignedData), sigInBytes, privateKey)
	if err != nil {
		log.Errorf("Failed to verify signature, because %v", err)
		return false, http.StatusInternalServerError, err
	}
	return verified, 0, err
}

func getTotalRewardFromRequest(r *http.Request) (float64, error) {
	var statsRequest LoadbalancerStatsRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(reqBody, &statsRequest)
	if err != nil {
		return 0, fmt.Errorf("invalid request body: %v", err)
	}
	totalRewardAsFloat, err := strconv.ParseFloat(statsRequest.TotalReward, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid total reward value: %v", err)
	}
	return totalRewardAsFloat, nil
}

// handler for `GET /api/v1/stats/node/{id}`
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

type LbStatsResponse struct {
	LbFee   string `json:"lb_fee"`
	NodeFee string `json:"nodes_fee"`
}

// handler for `GET /api/v1/stats/lb`
func (c *ApiController) StatisticsHandlerStatsForLoadBalancer(w http.ResponseWriter, r *http.Request) {
	statsResponse := LbStatsResponse{
		LbFee:   strconv.FormatFloat(float64(configuration.Config.Fee), 'f', -1, 32),
		NodeFee: strconv.FormatFloat(float64(1 - configuration.Config.Fee), 'f', -1, 32),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(statsResponse)
}
