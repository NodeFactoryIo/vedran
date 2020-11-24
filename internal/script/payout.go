package script

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"net/http"
	"net/url"

	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/payout"
)

var statsEndpoint, _ = url.Parse("/api/v1/stats")

func ExecutePayout(secret string, totalReward float64, loadbalancerUrl *url.URL) ([]*payout.TransactionDetails, error) {
	stats, err := fetchStatsFromEndpoint(loadbalancerUrl.ResolveReference(statsEndpoint), secret)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch stats from loadbalancer, %v", err)
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

	return payout.ExecuteAllPayoutTransactions(
		distributionByNode,
		stats.Stats,
		secret,
		loadbalancerUrl.String(),
	)
}

func fetchStatsFromEndpoint(endpoint *url.URL, secret string) (*controllers.LoadbalancerStatsResponse, error) {
	startPayout := controllers.LoadbalancerStatsRequest{StartPayout: true}
	payload, err := json.Marshal(startPayout)
	if err != nil {
		return nil, err
	}
	sig, err := signature.Sign([]byte("loadbalancer-request"), secret)
	if err != nil {
		return nil, err
	}

	request, _ := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(payload))
	request.Header.Set("X-Signature", string(sig))

	c := &http.Client{}
	resp, err := c.Do(request)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	stats := controllers.LoadbalancerStatsResponse{}
	err = dec.Decode(&stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
