package cmd

import (
	"github.com/NodeFactoryIo/vedran/internal/router"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts vedran load balancer",
	Run:   startCommand,
}

func init() {
	RootCmd.AddCommand(startCmd)
}

func startCommand(_ *cobra.Command, _ []string) {
	err := http.ListenAndServe(":4000", router.CreateNewApiRouter())
	log.Println("Started vedran load balancer on port :4000...")
	log.Fatal(err)
}
