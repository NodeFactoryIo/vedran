package cmd

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/script"
	"github.com/NodeFactoryIo/vedran/internal/ui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
)

var (
	privateKey         string
	totalReward        string
	rawLoadbalancerUrl string
	feeAddress         string

	loadbalancerURL      *url.URL
	totalRewardAsFloat64 float64
)

var payoutCmd = &cobra.Command{
	Use:   "payout",
	Short: "Starts payout script",
	Run:   payoutCommand,
	Args: func(cmd *cobra.Command, args []string) error {
		var err error
		totalRewardAsFloat64, err = ValidatePayoutFlags(totalReward, feeAddress, true)
		if err != nil {
			return err
		}

		loadbalancerURL, err = url.Parse(rawLoadbalancerUrl)
		if err != nil {
			return fmt.Errorf("invalid loadbalancer URL: %v", err)
		}

		return nil
	},
}

func init() {
	payoutCmd.Flags().StringVar(
		&privateKey,
		"private-key",
		"",
		"[REQUIRED] loadbalancer wallet private key",
	)
	payoutCmd.Flags().StringVar(
		&totalReward,
		"payout-reward",
		"-1",
		"[REQUIRED] total reward pool in Planck",
	)
	payoutCmd.Flags().StringVar(
		&rawLoadbalancerUrl,
		"load-balancer-url",
		"http://localhost:80",
		"[OPTIONAL] url on which loadbalancer is listening",
	)
	startCmd.Flags().StringVar(
		&feeAddress,
		"lb-payout-fee-address",
		"",
		"[OPTIONAL] Address on which load balancer fee will be sent. If omitted, load balancer fee will be left on load balancer wallet after payout",
	)

	_ = startCmd.MarkFlagRequired("private-key")

	RootCmd.AddCommand(payoutCmd)
}

func payoutCommand(_ *cobra.Command, _ []string) {
	DisplayBanner()
	fmt.Println("Payout script running...")
	transactions, err := script.ExecutePayout(privateKey, totalRewardAsFloat64, feeAddress, loadbalancerURL)
	if transactions != nil {
		// display even if only part of transactions executed
		ui.DisplayTransactionsStatus(transactions)
	}
	if err != nil {
		log.Errorf("Unable to execute payout, because of: %v", err)
		return
	} else {
		log.Info("Payout execution finished")
	}
}
