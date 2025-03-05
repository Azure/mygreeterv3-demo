// Auto generated. Can be modified.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/server"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Run:   start,
}

var options = server.Options{}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVar(&options.Port, "port", 50051, "the port to serve the api on")
	startCmd.Flags().BoolVar(&options.JsonLog, "json-log", false, "The format of the log is json or user friendly key-value pairs")
	startCmd.Flags().StringVar(&options.SubscriptionID, "subscription-id", "", "The subscription ID used to access and manage Azure resources")
	startCmd.Flags().BoolVar(&options.EnableAzureSDKCalls, "enable-azureSDK-calls", false, "Toggle to run azureSDK CRUDL calls if cluster is enabled with workload-id")
	startCmd.Flags().IntVar(&options.HTTPPort, "http port", 50061, "the port to serve the gRPC-Gateway on")
	startCmd.Flags().StringVar(&options.RemoteAddr, "remote-addr", "", "the demo server's address for this server to connect to")
	startCmd.Flags().Int64Var(&options.IntervalMilliSec, "interval-milli-sec", options.IntervalMilliSec,
		"The interval between two requests. Negative numbers mean sending one request.")
	startCmd.Flags().StringVar(&options.IdentityResourceID, "identity-resource-id", "", "the MSI used to authenticate to Azure from E2E env")
	startCmd.Flags().StringVar(&options.OperationContainerAddr, "opcon-addr", "localhost:50041", "the remote server's addr for this client to connect to")
	startCmd.Flags().StringVar(&options.ServiceBusHostName, "service-bus-hostname", "servicehubval-resourceName-location-sb-ns.servicebus.windows.net", "The host name used to connect to the service bus.")
	startCmd.Flags().StringVar(&options.ServiceBusQueueName, "service-bus-queue-name", "servicehubval-resourceName-queue", "The name of the queue to which we will send messages.")
	startCmd.Flags().StringVar(&options.DatabaseConnectionString, "database-connection-string", "", "Connection string used to connect to the database")
	startCmd.Flags().StringVar(&options.DatabaseServerUrl, "database-server-url", "", "The server of the database to connect to.")
	startCmd.Flags().StringVar(&options.DatabaseName, "database-name", "", "The name of the database to connect to.")
	startCmd.Flags().IntVar(&options.DatabasePort, "database-port", 1433, "The port to connect to the database")
	startCmd.Flags().StringVar(&options.EntityTableName, "entity-table-name", "hcp", "The name of the table that holds entity metadata and last operation affecting that entity.")
}

func start(cmd *cobra.Command, args []string) {
	newServer := server.NewServer()
	newServer.Init(options)
	newServer.Serve(options)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait indefinitely until a signal is received
	<-stop

	newServer.Cleanup()
}
