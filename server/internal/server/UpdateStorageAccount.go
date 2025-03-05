package server

import (
	"context"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateStorageAccount(ctx context.Context, in *pb.UpdateStorageAccountRequest) (*pb.UpdateStorageAccountResponse, error) {
	logger := ctxlogger.GetLogger(ctx)
	if s.AccountsClient == nil {
		logger.Error("AccountsClient is nil in UpdateStorageAccount(), azuresdk feature is likely disabled")
		return nil, status.Errorf(codes.Unimplemented, "AccountsClient is nil in UpdateStorageAccount(), azuresdk feature is likely disabled")
	}

	tags := make(map[string]*string)
	for k, v := range in.GetTags() {
		tags[k] = to.Ptr(v)
	}

	params := armstorage.AccountUpdateParameters{
		Tags: tags,
	}

	storageAccount, err := s.AccountsClient.Update(context.Background(), in.GetRgName(), in.GetSaName(), params, nil)
	if err != nil {
		logger.Error("Update() error: " + err.Error())
		return nil, HandleError(err, "Update()")
	}

	updatedStorageAccount := &pb.StorageAccount{
		Id:       *storageAccount.ID,
		Name:     *storageAccount.Name,
		Location: *storageAccount.Location,
	}
	logger.Info("Updated storage account: " + *storageAccount.Name)
	return &pb.UpdateStorageAccountResponse{StorageAccount: updatedStorageAccount}, nil
}
