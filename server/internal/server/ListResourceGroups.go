package server

import (
	"context"
	"strconv"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) ListResourceGroups(ctx context.Context, in *emptypb.Empty) (*pb.ListResourceGroupResponse, error) {
	logger := ctxlogger.GetLogger(ctx)
	if s.ResourceGroupClient == nil {
		logger.Error("ResourceGroupClient is nil in ListResourceGroups(), azuresdk feature is likely disabled")
		return nil, status.Errorf(codes.Unimplemented, "ResourceGroupClient is nil in ListResourceGroups(), azuresdk feature is likely disabled")
	}
	// creating a pager that is used to iterate over collection of resourceGroups'
	// Pass in nil to options parameter of NewListPager to get default pager
	pager := s.ResourceGroupClient.NewListPager(nil)
	var resourceGroups []*armresources.ResourceGroup
	var resourceGroupList []*pb.ResourceGroup
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			logger.Error("NextPage() error: " + err.Error())
			return nil, HandleError(err, "NextPage()")
		}
		if resp.Value != nil {
			resourceGroups = append(resourceGroups, resp.Value...)
		}
	}

	logger.Info("Resource groups found: " + strconv.Itoa(len(resourceGroups)))

	for _, rg := range resourceGroups {
		resourceGroup := &pb.ResourceGroup{
			Id:       *rg.ID,
			Name:     *rg.Name,
			Location: *rg.Location,
		}
		resourceGroupList = append(resourceGroupList, resourceGroup)
	}
	return &pb.ListResourceGroupResponse{RgList: resourceGroupList}, nil
}
