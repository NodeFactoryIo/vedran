package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/controllers"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"strconv"
	"strings"
)

var (
	secret string
	totalReward string
	loadbalancerUrl string

	totalRewardAsFloat64 float64
)

var payoutCmd = &cobra.Command{
	Use: "payout",
	Short: "Starts payout script",
	Run: payoutCommand,
	Args: func(cmd *cobra.Command, args []string) error {
		result, err := strconv.ParseFloat(totalReward, 64)
		if err != nil {
			return errors.New("invalid total reward value")
		}
		totalRewardAsFloat64 = result

		if !(strings.HasPrefix(loadbalancerUrl, "http://") ||
			strings.HasPrefix(loadbalancerUrl, "https://")) {
			loadbalancerUrl = "http://" + loadbalancerUrl
		}
		return nil
	},
}

func init() {
	payoutCmd.Flags().StringVar(
		&secret,
		"secret",
		"",
		"[REQUIRED] loadbalancer wallet secret",
	)
	payoutCmd.Flags().StringVar(
		&totalReward,
		"total-reward",
		"",
		"[REQUIRED] total reward pool in Planck",
	)
	payoutCmd.Flags().StringVar(
		&loadbalancerUrl,
		"load-balancer-url",
		"localhost:80",
		"[OPTIONAL] url on which loadbalancer is listening")
	RootCmd.AddCommand(payoutCmd)
}

func payoutCommand(_ *cobra.Command, _ []string) {
	DisplayBanner()
	fmt.Println("Payout script running...")

	stats, err := fetchStatsFromEndpoint(loadbalancerUrl + "/api/v1/stats")
	if err != nil {
		log.Errorf("Unable to fetch stats from loadbalancer, because of %v", err)
		return
	}

	// calculate distribution
	nodeStatsDetails := make(map[string]models.NodeStatsDetails, len(stats.Stats))
	for nodeId, nodeStats := range stats.Stats {
		nodeStatsDetails[nodeId] = nodeStats.Stats
	}

	distributionByNode := payout.CalculatePayoutDistributionByNode(
		nodeStatsDetails,
		totalRewardAsFloat64,
		float64(stats.Fee),
	)

	transactionDetails, err := payout.ExecuteAllPayoutTransactions(
		distributionByNode,
		stats.Stats,
		secret,
		loadbalancerUrl,
	)
	if err != nil {
		log.Errorf("Unable to execute payout transactions, because of %v", err)
		return
	}

	// todo - prettify displaying transactions status
	log.Info(transactionDetails)
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
