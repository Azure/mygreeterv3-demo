package server

import (
	"context"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ReadStorageAccount(ctx context.Context, in *pb.ReadStorageAccountRequest) (*pb.ReadStorageAccountResponse, error) {
	logger := ctxlogger.GetLogger(ctx)
	if s.AccountsClient == nil {
		logger.Error("AccountsClient is nil in ReadStorageAccount(), azuresdk feature is likely disabled")
		return nil, status.Errorf(codes.Unimplemented, "AccountsClient is nil in ReadStorageAccount(), azuresdk feature is likely disabled")
	}

	resp, err := s.AccountsClient.GetProperties(context.Background(), in.GetRgName(), in.GetSaName(), &armstorage.AccountsClientGetPropertiesOptions{Expand: nil})
	if err != nil {
		logger.Error("GetProperties() error: " + err.Error())
		return nil, HandleError(err, "GetProperties()")
	}

	storageAccount := resp.Account

	readStorageAccount := &pb.StorageAccount{
		Id:       *storageAccount.ID,
		Name:     *storageAccount.Name,
		Location: *storageAccount.Location,
	}

	logger.Info("Read storage account: " + *storageAccount.Name + " in " + *storageAccount.Location)
	return &pb.ReadStorageAccountResponse{StorageAccount: readStorageAccount}, nil
}
