package script

import (
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

var statsEndpoint, _ = url.Parse("/api/v1/stats")

func ExecutePayout(secret string, totalReward float64, loadbalancerUrl *url.URL) error {
	response, err := fetchStatsFromEndpoint(loadbalancerUrl.ResolveReference(statsEndpoint))
	if err != nil {
		return fmt.Errorf("unable to fetch stats from loadbalancer, %v", err)
	}

	distributionByNode := payout.CalculatePayoutDistributionByNode(
		response.Stats,
		totalReward,
		float64(response.Fee),
	)

	// todo - call sending payout transactions
	log.Info(distributionByNode)
	return nil
}

func fetchStatsFromEndpoint(endpoint *url.URL) (*controllers.StatsResponse, error) {
	resp, err := http.Get(endpoint.String())
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
