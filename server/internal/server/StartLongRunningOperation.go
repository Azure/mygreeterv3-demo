package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	database "github.com/Azure/aks-async/database"
	opbus "github.com/Azure/aks-async/operationsbus"
	"github.com/Azure/aks-middleware/grpc/server/ctxlogger"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/async/operations"
	oc "github.com/Azure/OperationContainer/api/v1"
)

func (s *Server) StartLongRunningOperation(ctx context.Context, in *pb.StartLongRunningOperationRequest) (*pb.StartLongRunningOperationResponse, error) {
	logger := ctxlogger.GetLogger(ctx)
	logger.Info("Starting async operation.")

	operationId, err := uuid.NewV4()
	if err != nil {
		logger.Error("Failed to generate UUID: " + err.Error())
		return nil, err
	}

	operation := &opbus.OperationRequest{
		OperationName:       operations.LroName,
		APIVersion:          "",
		OperationId:         operationId.String(),
		Body:                nil,
		HttpMethod:          "",
		RetryCount:          0,
		EntityId:            in.GetEntityId(),
		EntityType:          in.GetEntityType(),
		ExpirationTimestamp: in.GetExpirationTimestamp(),
	}

	// We add operations in the following order:
	// 1. Put the operationRequest in the service bus.
	// 2. Update the entity store with the desired state introduced by this operation.
	// 3. Add the operation to the operationContainer.
	// We follow this process because if the service bus step fails, then we can simply return an error. However,
	// if we first update the entity store with the desired state (which succeed), and then sending the operationRequest
	// to the service bus fails, the operation would never be updated to a terminated state since there's
	// no processor to process it. The ordering avoids this issue.
	marshalledOperation, err := json.Marshal(operation)
	if err != nil {
		logger.Error("Error marshalling operation: " + err.Error())
		return nil, err
	}

	logger.Info("Sending message to Service Bus")
	err = s.serviceBusSender.SendMessage(ctx, marshalledOperation)
	if err != nil {
		logger.Error("Error sending message to service bus: " + err.Error())
		return nil, err
	}

	// This query checks that the operationId doesn't already exists and inserts it into the table if successful all in a single query to avoid race conditions.
	// If this step fails, the processor will simply retry to process the operation until the "Max Delivery Count" is reached, at which point the message
	// will be sent to the DeadLetterQueue, to be set as Failed.
	query := fmt.Sprintf(`
    MERGE INTO %s AS target
    USING (SELECT @p1 AS entity_id_value, @p2 AS entity_type_value, @p3 AS last_operation_id_value, @p4 AS operation_name_value, @p5 AS operation_status_value) AS source
    ON target.entity_id = source.entity_id_value and target.entity_type = source.entity_type_value
    WHEN MATCHED AND (target.operation_status = '%s' OR target.operation_status = '%s' OR target.operation_status = '%s' OR target.operation_status = '%s') THEN
      UPDATE SET
        target.last_operation_id = source.last_operation_id_value,
        target.operation_name = source.operation_name_value,
        target.operation_status = source.operation_status_value
    WHEN NOT MATCHED THEN
      INSERT (entity_id, entity_type, last_operation_id, operation_name, operation_status)
      VALUES (source.entity_id_value, source.entity_type_value, source.last_operation_id_value, source.operation_name_value, source.operation_status_value);
  `, s.entityTableName, oc.Status_COMPLETED.String(), oc.Status_FAILED.String(), oc.Status_CANCELLED.String(), oc.Status_UNKNOWN.String())
	initialOperationStatus := oc.Status_PENDING.String()
	_, err = database.ExecDb(ctx, s.dbClient, query, in.GetEntityId(), in.GetEntityType(), operationId.String(), operations.LroName, initialOperationStatus)
	// If no rows were affected, the ExecDb function will throw an error saying `No rows were affected!`. With this error we could
	// conclude that the query run successfully but didn't insert the record due to another record already existing with the same
	// operation_id.
	if err != nil {
		logger.Error("Error in operations query: " + err.Error())
		//TODO(mheberling): Change this to return a known type of error in aks-async/database, instead of errors.New(...)
		if strings.Index(err.Error(), "No rows were affected!") == -1 {
			logger.Error("The combination of entityId " + in.GetEntityId() + " and entityType " + in.GetEntityType() + "was found in a non finalized state. Entity was not updated.")
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	createOperationStatusRequest := &oc.CreateOperationStatusRequest{
		OperationName:       operations.LroName,
		EntityId:            in.GetEntityId(),
		ExpirationTimestamp: in.GetExpirationTimestamp(),
		OperationId:         operationId.String(),
	}

	// Add the operation to the db.
	logger.Info("Adding operation to db.")
	_, err = s.operationContainerClient.CreateOperationStatus(ctx, createOperationStatusRequest)
	if err != nil {
		logger.Error("Error creating operation status: " + err.Error())
		return nil, err
	}

	startOperationResponse := &pb.StartLongRunningOperationResponse{OperationId: operationId.String()}
	return startOperationResponse, nil
}
