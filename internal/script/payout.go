package script

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/api"
	"github.com/NodeFactoryIo/vedran/internal/constants"
	"github.com/centrifuge/go-substrate-rpc-client/v2/signature"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/payout"

	log "github.com/sirupsen/logrus"
)

func ExecutePayout(
	privateKey string,
	totalReward float64,
	lbFeeAddress string,
	loadbalancerUrl *url.URL,
) ([]*payout.TransactionDetails, error) {
	log.Info("New payout started.")

	substrateAPI, err := api.InitializeSubstrateAPI(wsEndpoint(loadbalancerUrl).String())
	if err != nil {
		return nil, fmt.Errorf("unable to initialize substrate API, because of %v", err)
	}

	metadataLatest, err := substrateAPI.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch latest metadata, because of %v", err)
	}

	// Use wildcard 42 - Generic Substrate wildcard
	// https://github.com/paritytech/substrate/wiki/External-Address-Format-(SS58)#address-type
	keyringPair, err := signature.KeyringPairFromSecret(privateKey, 42)
	if err != nil {
		return nil, fmt.Errorf("invalid private key, %v", err)
	}

	// distribute entire balance on address if total reward not set
	if totalReward == -1 {
		balance, err := payout.GetBalance(metadataLatest, keyringPair, substrateAPI)
		if err != nil {
			return nil, err
		}
		totalReward = float64(balance.Int64())
	}

	log.Infof("Total reward: %s", strconv.FormatFloat(totalReward, 'f', 0, 64))

	response, err := fetchStatsFromEndpoint(
		statsEndpoint(loadbalancerUrl), privateKey, fmt.Sprintf("%f", totalReward),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch stats from loadbalancer, %v", err)
	}

	distributionByNode := payout.CalculatePayoutDistributionByNode(
		response.Stats,
		totalReward,
		payout.LoadBalancerDistributionConfiguration{
			FeePercentage:       float64(response.Fee),
			PayoutAddress:       lbFeeAddress,
			DifferentFeeAddress: lbFeeAddress != "",
		},
	)

	return payout.ExecuteAllPayoutTransactions(
		distributionByNode,
		substrateAPI,
		keyringPair,
	)
}

func fetchStatsFromEndpoint(endpoint *url.URL, secret string, totalReward string) (*controllers.LoadbalancerStatsResponse, error) {
	sig, err := signature.Sign([]byte(constants.StatsSignedData), secret)
	if err != nil {
		return nil, err
	}

	payloadBuf := new(bytes.Buffer)
	_ = json.NewEncoder(payloadBuf).Encode(controllers.LoadbalancerStatsRequest{TotalReward: totalReward})

	request, _ := http.NewRequest("POST", endpoint.String(), payloadBuf)
	request.Header.Set("X-Signature", hexutil.Encode(sig))

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
