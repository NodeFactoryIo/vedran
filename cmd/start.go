package cmd

import (
	"github.com/NodeFactoryIo/vedran/internal/router"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var (
	shellCmd string
)

var exampleCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts vedran load balancer",
	Run:   startCommand,
}

func init() {
	exampleCmd.Flags().StringVar(&shellCmd, "command", "flag", "example flag")
	RootCmd.AddCommand(exampleCmd)
}

func startCommand(_ *cobra.Command, _ []string) {
	log.Println("Starting server on :4000...")
	err := http.ListenAndServe(":4000", router.CreateNewApiRouter())
	log.Fatal(err)
}
