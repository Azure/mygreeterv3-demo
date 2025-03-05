package server

import (
	"context"
	"strconv"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListStorageAccounts(ctx context.Context, in *pb.ListStorageAccountRequest) (*pb.ListStorageAccountResponse, error) {
	logger := ctxlogger.GetLogger(ctx)
	if s.AccountsClient == nil {
		logger.Error("AccountsClient is nil in LisgStorageAccounts(), azuresdk feature is likely disabled")
		return nil, status.Errorf(codes.Unimplemented, "AccountsClient is nil in ListStorageAccounts(), azuresdk feature is likely disabled")
	}

	var storageAccounts []*armstorage.Account
	var storageAccountList []*pb.StorageAccount
	pager := s.AccountsClient.NewListByResourceGroupPager(in.GetRgName(), nil)
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			logger.Error("NextPage() error: " + err.Error())
			return nil, HandleError(err, "NextPage()")
		}
		if resp.Value != nil {
			storageAccounts = append(storageAccounts, resp.Value...)
		}
	}

	logger.Info("Storage accounts found: " + strconv.Itoa(len(storageAccounts)))

	for _, sa := range storageAccounts {
		storageAccount := &pb.StorageAccount{
			Id:       *sa.ID,
			Name:     *sa.Name,
			Location: *sa.Location,
		}
		storageAccountList = append(storageAccountList, storageAccount)
	}

	return &pb.ListStorageAccountResponse{SaList: storageAccountList}, nil
}
