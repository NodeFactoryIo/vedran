package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	verbose bool
)

var RootCmd = &cobra.Command{
	Use:   "vedran",
	Short: "Vedran cli tool",
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
