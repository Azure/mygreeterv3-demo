package longRunningOperation

import (
	"context"
	"errors"
	"time"

	opbus "github.com/Azure/aks-async/operationsbus"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc/codes"
)

// Setting the variable to ensure all functions of the ApiOperation interface are implemented.
var _ opbus.ApiOperation = &LongRunningOperation{}

type LongRunningOperation struct {
	Name                string
	Operation           opbus.OperationRequest
	LroEntity           *LongRunningEntity
	OperationId         string
	ExpirationTimestamp *timestamppb.Timestamp
}

var CreateLroEntityFunc opbus.EntityFactoryFunc = func(id string) opbus.Entity {
	return NewLongRunningEntity(id)
}

func (lro *LongRunningOperation) InitOperation(ctx context.Context, opRequest opbus.OperationRequest) (opbus.ApiOperation, error) {
	logger := ctxlogger.GetLogger(ctx)

	logger.Info("Initializing LongRunningOperation")
	lro.Operation = opRequest
	lro.Name = opRequest.OperationName
	lro.OperationId = opRequest.OperationId

	return nil, nil
}

func (lro *LongRunningOperation) Run(ctx context.Context) error {
	logger := ctxlogger.GetLogger(ctx)
	logger.Info("Running the long running operation!")

	// Logic for running the operation
	time.Sleep(20 * time.Second)
	logger.Info("Finished running the long running operation.")

	return nil
}

func (lro *LongRunningOperation) GuardConcurrency(ctx context.Context, entity opbus.Entity) *opbus.CategorizedError {
	logger := ctxlogger.GetLogger(ctx)
	logger.Info("Guarding concurrency for operation.")

	if latestOperationId := entity.GetLatestOperationID(); lro.OperationId != latestOperationId {
		err := errors.New("OperaionId and LastOperationId don't match!")
		ce := opbus.NewCategorizedError(err.Error(), "", int(codes.Canceled), err)
		return ce
	}

	return nil
}

func (lro *LongRunningOperation) GetOperationRequest() *opbus.OperationRequest {
	return &lro.Operation
}
