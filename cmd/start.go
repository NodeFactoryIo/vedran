package cmd

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/auth"
	"github.com/NodeFactoryIo/vedran/internal/router"
	"github.com/asdine/storm/v3"
	"github.com/spf13/cobra"
	"log"
	"net/http"
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
		"authentication secret used for generating tokens")
	RootCmd.AddCommand(startCmd)
}

func startCommand(_ *cobra.Command, _ []string) {
	// set auth secret
	err := auth.SetAuthSecret(authSecretFlag)
	if err != nil {
		// terminate app: no auth secret provided
		log.Fatal(fmt.Sprintf("Unable to start vedran load balancer: %v", err))
	}

	// init database
	database, err := storm.Open("vedran-load-balancer.db")
	if err != nil {
		// terminate app: unable to start database connection
		log.Fatal(fmt.Sprintf("Unable to start vedran load balancer: %v", err))
	}

	log.Println("Starting vedran load balancer on port :4000...")

	err = http.ListenAndServe(":4000", router.CreateNewApiRouter(database))
	if err != nil {
		log.Print(err)
	}
	err = database.Close()
	log.Fatal(err)
}
