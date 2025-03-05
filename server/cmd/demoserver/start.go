package main

import (
	"context"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/demoserver"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/externalserver"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Run:   start,
}

var options = demoserver.Options{}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVar(&options.Port, "port", 50071, "the addr to serve the api on")
	startCmd.Flags().BoolVar(&options.JsonLog, "json-log", false, "The format of the log is json or user friendly key-value pairs")
}

func start(cmd *cobra.Command, args []string) {
	demoserver.Serve(options)
}

func forwardSayHelloToExternalServer(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	conn, err := grpc.Dial("externalserver:50072", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewMyGreeterClient(conn)
	return client.SayHello(ctx, req)
}
