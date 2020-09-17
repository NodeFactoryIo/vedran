package cmd

import (
	"github.com/NodeFactoryIo/vedran/internal/loadbalancer"
	"github.com/spf13/cobra"
)

var (
	authSecretFlag string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts vedran load balancer",
	Run:   startCommand,
}

func init() {
	startCmd.Flags().StringVar(
		&authSecretFlag,
		"auth-secret",
		"",
		"example flag")
	RootCmd.AddCommand(startCmd)
}

func startCommand(_ *cobra.Command, _ []string) {
	loadbalancer.StartLoadBalancerServer(authSecretFlag, "4000")
}
