package cmd

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/loadbalancer"
	"github.com/NodeFactoryIo/vedran/pkg/util/random"
	"github.com/spf13/cobra"
)

var (
	authSecret string
	name       string
	capacity string
	whitelist []string
	fee string
	selection string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts vedran load balancer",
	Run:   startCommand,
}

func init() {

	// initialize flags
	startCmd.Flags().StringVar(
		&authSecret, "auth-secret", "", "example flag")
	startCmd.Flags().StringVar(
		&name, "name", fmt.Sprintf("load-balancer-%s", random.String(12, random.Alphabetic)), "")
	startCmd.Flags().StringVar(
		&capacity, "name", "", "")
	startCmd.Flags().StringArrayVar(
		&whitelist, "name", nil, "")
	startCmd.Flags().StringVar(
		&fee, "name", "", "")
	startCmd.Flags().StringVar(
		&selection, "name", "", "")

	// mark required flags
	_ = startCmd.MarkFlagRequired("auth-secret")
	_ = startCmd.MarkFlagRequired("fee")
	_ = startCmd.MarkFlagRequired("selection")

	RootCmd.AddCommand(startCmd)
}

func startCommand(_ *cobra.Command, _ []string) {

	loadbalancer.StartLoadBalancerServer(authSecret, "4000")
}
