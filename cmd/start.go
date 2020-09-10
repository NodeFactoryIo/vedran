package cmd

import (
	"github.com/NodeFactoryIo/vedran/internal/handlers"
	"github.com/gorilla/mux"
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
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/nodes", handlers.RegisterHandler).Methods("POST")

	log.Println("Starting server on :4000...")
	err := http.ListenAndServe(":4000", router)
	log.Fatal(err)
}
