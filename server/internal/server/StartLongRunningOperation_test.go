package server

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"time"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/async/operations"

	oc "github.com/Azure/OperationContainer/api/v1"
	ocMock "github.com/Azure/OperationContainer/api/v1/mock"
	opbus "github.com/Azure/aks-async/operationsbus"
	"github.com/Azure/aks-async/servicebus"
	"github.com/DATA-DOG/go-sqlmock"

	asyncMocks "github.com/Azure/aks-async/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Match satisfies sqlmock.Argument interface.
// Required for checking that the operationId matches a string format.
type AnyString struct{}

func (a AnyString) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok
}

var _ = Describe("Mock Testing", func() {
	var (
		ctrl                     *gomock.Controller
		s                        *Server
		mockSender               *asyncMocks.MockSenderInterface
		db                       *sql.DB
		mockDb                   sqlmock.Sqlmock
		entityTableName          string
		operationContainerClient *ocMock.MockOperationContainerClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockSender = asyncMocks.NewMockSenderInterface(ctrl)
		db, mockDb, _ = sqlmock.New()
		entityTableName = "hcp"
		operationContainerClient = ocMock.NewMockOperationContainerClient(ctrl)
		s = &Server{
			serviceBusSender:         mockSender,
			dbClient:                 db,
			entityTableName:          entityTableName,
			operationContainerClient: operationContainerClient,
		}
	})

	AfterEach(func() {
		db.Close()
		ctrl.Finish()
	})

	Context("async operations", func() {
		It("should return operationId and insert new operation into database", func() {
			protoExpirationTime := timestamppb.New(time.Now().Add(1 * time.Hour))
			in := &pb.StartLongRunningOperationRequest{
				EntityId:            "test",
				EntityType:          "test",
				ExpirationTimestamp: protoExpirationTime,
			}
			initialOperationStatus := oc.Status_PENDING.String()

			query := fmt.Sprintf(`
        MERGE INTO %s AS target
        USING \(SELECT @p1 AS entity_id_value, @p2 AS entity_type_value, @p3 AS last_operation_id_value, @p4 AS operation_name_value, @p5 AS operation_status_value\) AS source
        ON target.entity_id = source.entity_id_value and target.entity_type = source.entity_type_value
        WHEN MATCHED AND \(target.operation_status = 'COMPLETED' OR target.operation_status = 'FAILED' OR target.operation_status = 'CANCELLED' OR target.operation_status = 'UNKNOWN'\) THEN
          UPDATE SET
            target.last_operation_id = source.last_operation_id_value,
            target.operation_name = source.operation_name_value,
            target.operation_status = source.operation_status_value
        WHEN NOT MATCHED THEN
          INSERT \(entity_id, entity_type, last_operation_id, operation_name, operation_status\)
          VALUES \(source.entity_id_value, source.entity_type_value, source.last_operation_id_value, source.operation_name_value, source.operation_status_value\);
       `, s.entityTableName)
			// Need to use AnyString{} since we only care that it's a string, not really the values of it.
			mockDb.ExpectExec(query).WithArgs(in.GetEntityId(), in.GetEntityType(), AnyString{}, operations.LroName, initialOperationStatus).WillReturnResult(sqlmock.NewResult(1, 1))

			operationContainerClient.EXPECT().CreateOperationStatus(gomock.Any(), gomock.Any()).Return(nil, nil)
			mockSender.EXPECT().SendMessage(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			out, err := s.StartLongRunningOperation(context.Background(), in)
			Expect(err).To(BeNil())
			Expect(out.OperationId).NotTo(BeNil())

			err = mockDb.ExpectationsWereMet()
			Expect(err).To(BeNil())
		})
		It("should fail on sender failure", func() {
			protoExpirationTime := timestamppb.New(time.Now().Add(1 * time.Hour))
			in := &pb.StartLongRunningOperationRequest{
				EntityId:            "test",
				EntityType:          "test",
				ExpirationTimestamp: protoExpirationTime,
			}

			errorMessage := "ServiceBus Sender error"
			err := errors.New(errorMessage)
			mockSender.EXPECT().SendMessage(gomock.Any(), gomock.Any()).Return(err).Times(1)
			out, err := s.StartLongRunningOperation(context.Background(), in)
			Expect(err).ToNot(BeNil())
			Expect(out).To(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorMessage))

			err = mockDb.ExpectationsWereMet()
			Expect(err).To(BeNil())
		})
		It("should fail on database query failure", func() {
			protoExpirationTime := timestamppb.New(time.Now().Add(1 * time.Hour))
			in := &pb.StartLongRunningOperationRequest{
				EntityId:            "test",
				EntityType:          "test",
				ExpirationTimestamp: protoExpirationTime,
			}
			initialOperationStatus := oc.Status_PENDING.String()

			query := fmt.Sprintf(`
        MERGE INTO %s AS target
        USING \(SELECT @p1 AS entity_id_value, @p2 AS entity_type_value, @p3 AS last_operation_id_value, @p4 AS operation_name_value, @p5 AS operation_status_value\) AS source
        ON target.entity_id = source.entity_id_value and target.entity_type = source.entity_type_value
        WHEN MATCHED AND \(target.operation_status = 'COMPLETED' OR target.operation_status = 'FAILED' OR target.operation_status = 'CANCELLED' OR target.operation_status = 'UNKNOWN'\) THEN
          UPDATE SET
            target.last_operation_id = source.last_operation_id_value,
            target.operation_name = source.operation_name_value,
            target.operation_status = source.operation_status_value
        WHEN NOT MATCHED THEN
          INSERT \(entity_id, entity_type, last_operation_id, operation_name, operation_status\)
          VALUES \(source.entity_id_value, source.entity_type_value, source.last_operation_id_value, source.operation_name_value, source.operation_status_value\);
       `, s.entityTableName)

			// Need to use AnyString{} since we only care that it's a string, not really the values of it.
			errorMessage := "Database failure error."
			err := errors.New(errorMessage)
			mockDb.ExpectExec(query).WithArgs(in.GetEntityId(), in.GetEntityType(), AnyString{}, operations.LroName, initialOperationStatus).WillReturnResult(sqlmock.NewErrorResult(err))

			mockSender.EXPECT().SendMessage(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			_, err = s.StartLongRunningOperation(context.Background(), in)
			Expect(err).ToNot(BeNil())

			err = mockDb.ExpectationsWereMet()
			Expect(err).To(BeNil())
		})
		It("should fail when operationcontainer fails", func() {
			protoExpirationTime := timestamppb.New(time.Now().Add(1 * time.Hour))
			in := &pb.StartLongRunningOperationRequest{
				EntityId:            "20",
				EntityType:          "Cluster",
				ExpirationTimestamp: protoExpirationTime,
			}
			initialOperationStatus := oc.Status_PENDING.String()

			query := fmt.Sprintf(`
        MERGE INTO %s AS target
        USING \(SELECT @p1 AS entity_id_value, @p2 AS entity_type_value, @p3 AS last_operation_id_value, @p4 AS operation_name_value, @p5 AS operation_status_value\) AS source
        ON target.entity_id = source.entity_id_value and target.entity_type = source.entity_type_value
        WHEN MATCHED AND \(target.operation_status = 'COMPLETED' OR target.operation_status = 'FAILED' OR target.operation_status = 'CANCELLED' OR target.operation_status = 'UNKNOWN'\) THEN
          UPDATE SET
            target.last_operation_id = source.last_operation_id_value,
            target.operation_name = source.operation_name_value,
            target.operation_status = source.operation_status_value
        WHEN NOT MATCHED THEN
          INSERT \(entity_id, entity_type, last_operation_id, operation_name, operation_status\)
          VALUES \(source.entity_id_value, source.entity_type_value, source.last_operation_id_value, source.operation_name_value, source.operation_status_value\);
       `, s.entityTableName)

			// Need to use AnyString{} since we only care that it's a string, not really the values of it.
			mockDb.ExpectExec(query).WithArgs(in.GetEntityId(), in.GetEntityType(), AnyString{}, operations.LroName, initialOperationStatus).WillReturnResult(sqlmock.NewResult(1, 1))

			mockSender.EXPECT().SendMessage(gomock.Any(), gomock.Any()).Return(nil).Times(1)

			errMessage := "Something went wrong with OperationContainer!"
			err := errors.New(errMessage)
			operationContainerClient.EXPECT().CreateOperationStatus(gomock.Any(), gomock.Any()).Return(nil, err)
			out, err := s.StartLongRunningOperation(context.Background(), in)
			Expect(err).NotTo(BeNil())
			Expect(out).To(BeNil())
		})
	})
})

var _ = Describe("Fakes Testing", func() {
	var (
		s                        *Server
		db                       *sql.DB
		mockDb                   sqlmock.Sqlmock
		entityTableName          string
		ctrl                     *gomock.Controller
		operationContainerClient *ocMock.MockOperationContainerClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		sbClient := servicebus.NewFakeServiceBusClient()
		sbSender, _ := sbClient.NewServiceBusSender(nil, "", nil)
		db, mockDb, _ = sqlmock.New()
		entityTableName = "hcp"
		operationContainerClient = ocMock.NewMockOperationContainerClient(ctrl)
		s = &Server{
			ResourceGroupClient:      nil,
			serviceBusClient:         sbClient,
			serviceBusSender:         sbSender,
			dbClient:                 db,
			entityTableName:          entityTableName,
			operationContainerClient: operationContainerClient,
		}
	})

	Context("Message should exist in the service bus", func() {
		It("should send the message successfully", func() {
			protoExpirationTime := timestamppb.New(time.Now().Add(1 * time.Hour))
			initialOperationStatus := oc.Status_PENDING.String()
			in := &pb.StartLongRunningOperationRequest{
				EntityId:            "test",
				EntityType:          "test",
				ExpirationTimestamp: protoExpirationTime,
			}

			// Can mostly ignore this since we tested it above. We care more about the service bus mock in this test.
			// Need to add it because otherwise the server call will complain froma null pointer to try and access the db
			// if it doesn't exist in the test.
			query := fmt.Sprintf(`
        MERGE INTO %s AS target
        USING \(SELECT @p1 AS entity_id_value, @p2 AS entity_type_value, @p3 AS last_operation_id_value, @p4 AS operation_name_value, @p5 AS operation_status_value\) AS source
        ON target.entity_id = source.entity_id_value and target.entity_type = source.entity_type_value
        WHEN MATCHED AND \(target.operation_status = 'COMPLETED' OR target.operation_status = 'FAILED' OR target.operation_status = 'CANCELLED' OR target.operation_status = 'UNKNOWN'\) THEN
          UPDATE SET
            target.last_operation_id = source.last_operation_id_value,
            target.operation_name = source.operation_name_value,
            target.operation_status = source.operation_status_value
        WHEN NOT MATCHED THEN
          INSERT \(entity_id, entity_type, last_operation_id, operation_name, operation_status\)
          VALUES \(source.entity_id_value, source.entity_type_value, source.last_operation_id_value, source.operation_name_value, source.operation_status_value\);
       `, s.entityTableName)
			mockDb.ExpectExec(query).WithArgs(in.GetEntityId(), in.GetEntityType(), AnyString{}, operations.LroName, initialOperationStatus).WillReturnResult(sqlmock.NewResult(1, 1))

			operationContainerClient.EXPECT().CreateOperationStatus(gomock.Any(), gomock.Any()).Return(nil, nil)

			_, err := s.StartLongRunningOperation(context.Background(), in)
			Expect(err).ToNot(HaveOccurred())

			sbReceiver, _ := s.serviceBusClient.NewServiceBusReceiver(nil, "", nil)
			msg, err := sbReceiver.ReceiveMessage(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg).NotTo(BeNil())

			opRequestExpected := opbus.NewOperationRequest("LongRunningOperation", "", "", "test", "test", 0, protoExpirationTime, nil, "", nil)

			var opRequestReceived opbus.OperationRequest
			err = json.Unmarshal(msg, &opRequestReceived)
			Expect(err).ToNot(HaveOccurred())

			Expect(opRequestReceived.OperationName).To(Equal(opRequestExpected.OperationName))
			Expect(opRequestReceived.OperationId).NotTo(BeNil())
			Expect(opRequestReceived.RetryCount).To(Equal(opRequestExpected.RetryCount))
			Expect(opRequestReceived.EntityType).To(Equal(opRequestExpected.EntityType))
			Expect(opRequestReceived.EntityId).To(Equal(opRequestExpected.EntityId))
			Expect(opRequestReceived.ExpirationTimestamp).To(Equal(opRequestExpected.ExpirationTimestamp))
			Expect(opRequestReceived.APIVersion).To(Equal(opRequestExpected.APIVersion))

			err = mockDb.ExpectationsWereMet()
			Expect(err).To(BeNil())
		})
	})
})
