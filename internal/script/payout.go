package script

import (
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func ExecutePayout(secret string, totalReward float64, loadbalancerUrl string) error {
	stats, err := fetchStatsFromEndpoint(loadbalancerUrl + "/api/v1/stats")
	if err != nil {
		return fmt.Errorf("unable to fetch stats from loadbalancer, %v", err)
	}

	// calculate distribution
	nodeStatsDetails := make(map[string]models.NodeStatsDetails, len(stats.Stats))
	for nodeId, nodeStats := range stats.Stats {
		nodeStatsDetails[nodeId] = nodeStats.Stats
	}
	distributionByNode := payout.CalculatePayoutDistributionByNode(
		nodeStatsDetails,
		totalReward,
		float64(stats.Fee),
	)

	// todo - call sending payout transactions
	log.Info(distributionByNode)
	return nil
}

func fetchStatsFromEndpoint(endpoint string) (*controllers.StatsResponse, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	stats := controllers.StatsResponse{}
	err = dec.Decode(&stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
