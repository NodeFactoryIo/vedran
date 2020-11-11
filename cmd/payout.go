package cmd

import (
	"errors"
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/script"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	err := script.ExecutePayout(secret, totalRewardAsFloat64, loadbalancerUrl)
	if err != nil {
		log.Errorf("Unable to execute payout, because of %v", err)
		return
	}
	log.Info("Payout execution finished")
}

