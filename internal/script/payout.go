package script

import (
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	"net/http"
	"net/url"
)

var statsEndpoint, _ = url.Parse("/api/v1/stats")

func ExecutePayout(secret string, totalReward float64, loadbalancerUrl *url.URL) ([]*payout.TransactionDetails, error) {
	response, err := fetchStatsFromEndpoint(loadbalancerUrl.ResolveReference(statsEndpoint))
	if err != nil {
		return nil, fmt.Errorf("unable to fetch stats from loadbalancer, %v", err)
	}

	distributionByNode := payout.CalculatePayoutDistributionByNode(
		response.Stats,
		totalReward,
		float64(response.Fee),
	)

	return payout.ExecuteAllPayoutTransactions(
		distributionByNode,
		response.Stats,
		secret,
		"ws://localhost:4444/ws",
	)
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
