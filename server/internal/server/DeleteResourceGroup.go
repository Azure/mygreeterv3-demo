package server

import (
	"context"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) DeleteResourceGroup(ctx context.Context, in *pb.DeleteResourceGroupRequest) (*emptypb.Empty, error) {
	logger := ctxlogger.GetLogger(ctx)
	if s.ResourceGroupClient == nil {
		logger.Error("ResourceGroupClient is nil in DeleteResourceGroup(), azuresdk feature is likely disabled")
		return &emptypb.Empty{}, status.Errorf(codes.Unimplemented, "ResourceGroupClient is nil in DeleteResourceGroup(), azuresdk feature is likely disabled")
	}
	poller, err := s.ResourceGroupClient.BeginDelete(ctx, in.GetName(), nil)
	if err != nil {
		logger.Error("BeginDelete() error: " + err.Error())
		return &emptypb.Empty{}, HandleError(err, "BeginDelete()")
	}
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		logger.Error("PollUntilDone() error: " + err.Error())
		return &emptypb.Empty{}, HandleError(err, "PollUntilDone()")
	}

	logger.Info("Deleted resource group: " + in.GetName())
	return &emptypb.Empty{}, nil
}
