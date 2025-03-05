package server

import (
	"context"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateResourceGroup(ctx context.Context, in *pb.UpdateResourceGroupRequest) (*pb.UpdateResourceGroupResponse, error) {
	logger := ctxlogger.GetLogger(ctx)
	if s.ResourceGroupClient == nil {
		logger.Error("ResourceGroupClient is nil in UpdateResourceGroup(), azuresdk feature is likely disabled")
		return nil, status.Errorf(codes.Unimplemented, "ResourceGroupClient is nil in UpdateResourceGroup(), azuresdk feature is likely disabled")
	}

	tags := make(map[string]*string)
	for k, v := range in.GetTags() {
		tags[k] = to.Ptr(v)
	}

	update := armresources.ResourceGroupPatchable{
		Tags: tags,
	}

	rg, err := s.ResourceGroupClient.Update(
		ctx,
		in.GetName(),
		update,
		nil)

	if err != nil {
		logger.Error("Update() error: " + err.Error())
		return nil, HandleError(err, "Update()")
	}

	updatedResourceGroup := &pb.ResourceGroup{
		Id:       *rg.ID,
		Name:     *rg.Name,
		Location: *rg.Location,
	}
	logger.Info("Updated resource group: " + *rg.Name)
	return &pb.UpdateResourceGroupResponse{ResourceGroup: updatedResourceGroup}, nil
}
