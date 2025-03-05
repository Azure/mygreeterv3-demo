package main

import (
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/externalserver"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Run:   start,
}

var options = externalserver.Options{}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVar(&options.Port, "port", 50072, "the addr to serve the api on")
	startCmd.Flags().BoolVar(&options.JsonLog, "json-log", false, "The format of the log is json or user friendly key-value pairs")
}

func start(cmd *cobra.Command, args []string) {
	externalserver.Serve(options)
}
