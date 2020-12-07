package script

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/payout"

	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	log "github.com/sirupsen/logrus"
)

func ExecutePayout(privateKey string, totalReward float64, loadbalancerUrl *url.URL) ([]*payout.TransactionDetails, error) {
	log.Infof("New payout started with total reward: %s", strconv.FormatFloat(totalReward, 'f', 0, 64))
	response, err := fetchStatsFromEndpoint(statsEndpoint(loadbalancerUrl), privateKey)
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
		privateKey,
		wsEndpoint(loadbalancerUrl).String(),
	)
}

func fetchStatsFromEndpoint(endpoint *url.URL, secret string) (*controllers.LoadbalancerStatsResponse, error) {
	sig, err := signature.Sign([]byte(controllers.GetStatsSignedData()), secret)
	if err != nil {
		return nil, err
	}

	request, _ := http.NewRequest("POST", endpoint.String(), nil)
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
