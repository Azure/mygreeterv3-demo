package server

import (
	"context"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ReadResourceGroup(ctx context.Context, in *pb.ReadResourceGroupRequest) (*pb.ReadResourceGroupResponse, error) {
	logger := ctxlogger.GetLogger(ctx)
	if s.ResourceGroupClient == nil {
		logger.Error("ResourceGroupClient is nil in ReadResourceGroup(), azuresdk feature is likely disabled")
		return nil, status.Errorf(codes.Unimplemented, "ResourceGroupClient is nil in ReadResourceGroup(), azuresdk feature is likely disabled")
	}
	resourceGroupResponse, err := s.ResourceGroupClient.Get(
		ctx,
		in.GetName(),
		nil)

	if err != nil {
		logger.Error("Get() error: " + err.Error())
		return nil, HandleError(err, "Get()")
	}

	resourceGroup := resourceGroupResponse.ResourceGroup

	readResourceGroup := &pb.ResourceGroup{
		Id:       *resourceGroup.ID,
		Name:     *resourceGroup.Name,
		Location: *resourceGroup.Location,
	}

	logger.Info("Read resource group: " + *resourceGroup.Name + " in " + *resourceGroup.Location)
	return &pb.ReadResourceGroupResponse{ResourceGroup: readResourceGroup}, nil
}
