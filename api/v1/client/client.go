package client

import (
	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/interceptor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	log "log/slog"
)

// NewClient returns a client that has all the interceptors registered.
func NewClient(remoteAddr string, options interceptor.ClientInterceptorLogOptions) (pb.MyGreeterClient, error) {
	conn, err := grpc.Dial(
		remoteAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			interceptor.DefaultClientInterceptors(options)...,
		),
	)
	if err != nil {
		// logging for transparency, error handled by retry interceptor
		log.Error("did not connect: " + err.Error())
	}

	return pb.NewMyGreeterClient(conn), err
	// TODO: Figure out how to close the connection when the program exits.
	// defer conn.Close()
}
