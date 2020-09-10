package cmd

import (
	"fmt"
	"log"

	"github.com/eugenmayer/go-exec/exec"
	utils "github.com/eugenmayer/go-exec/utils/dialog"
	"github.com/spf13/cobra"
)

var (
	shellCmd string
)

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Example command",
	Run:   exampleCommand,
}

func init() {
	exampleCmd.Flags().StringVar(&shellCmd, "command", "flag", "example flag")
	RootCmd.AddCommand(exampleCmd)
}

func exampleCommand(_ *cobra.Command, _ []string) {
	// question based execution, get confirmaion
	if utils.ConfirmQuestion("Test confirmation dialog?") {
		stdout, stderr, err := exec.Run("date", verbose)
		if err != nil {
			log.Print(stdout)
			log.Print(stderr)
			log.Fatal(err)
		}
		log.Print(fmt.Sprintf("Confirm:\n %s", stdout))
	}
}
