package cmd

import (
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/script"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"strconv"
)

var (
	secret             string
	totalReward        string
	rawLoadbalancerUrl string

	loadbalancerURL      *url.URL
	totalRewardAsFloat64 float64
)

var payoutCmd = &cobra.Command{
	Use:   "payout",
	Short: "Starts payout script",
	Run:   payoutCommand,
	Args: func(cmd *cobra.Command, args []string) error {
		result, err := strconv.ParseFloat(totalReward, 64)
		if err != nil {
			return errors.New("invalid total reward value")
		}
		totalRewardAsFloat64 = result

		loadbalancerURL, err = url.Parse(rawLoadbalancerUrl)
		if err != nil {
			return fmt.Errorf("invalid loadbalancer URL: %v", err)
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
	_ = payoutCmd.MarkFlagRequired("secret")
	payoutCmd.Flags().StringVar(
		&totalReward,
		"total-reward",
		"",
		"[REQUIRED] total reward pool in Planck",
	)
	_ = payoutCmd.MarkFlagRequired("total-reward")
	payoutCmd.Flags().StringVar(
		&rawLoadbalancerUrl,
		"load-balancer-url",
		"http://localhost:80",
		"[OPTIONAL] url on which loadbalancer is listening")
	RootCmd.AddCommand(payoutCmd)
}

func payoutCommand(_ *cobra.Command, _ []string) {
	DisplayBanner()
	fmt.Println("Payout script running...")
	transactions, err := script.ExecutePayout(secret, totalRewardAsFloat64, loadbalancerURL)
	if transactions != nil {
		// display even if only part of transactions executed
		DisplayTransactionsStatus(transactions)
	}
	if err != nil {
		log.Errorf("Unable to execute payout, because of: %v", err)
		return
	} else {
		log.Info("Payout execution finished")
	}
}
