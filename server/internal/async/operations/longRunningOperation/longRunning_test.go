package longRunningOperation

import (
	"context"
	"time"

	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/async/operations"
	opbus "github.com/Azure/aks-async/operationsbus"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operation", func() {
	var (
		ctx                      context.Context
		lastOperationId          string
		differentLastOperationId string
		operation                *opbus.OperationRequest
	)

	BeforeEach(func() {
		ctx = context.Background()
		differentLastOperationId = "3"
		lastOperationId = "2"
		operation = opbus.NewOperationRequest(
			operations.LroName, "v0.0.1", lastOperationId, "1", "Cluster",
			0, nil, nil, "", nil,
		)
	})

	Context("Initializing Operation", func() {
		It("should initialize the operation successfully", func() {
			apiOperation := &LongRunningOperation{
				OperationId: lastOperationId,
			}
			Expect(apiOperation.Name).To(Equal(""))
			_, err := apiOperation.InitOperation(ctx, *operation)
			Expect(err).NotTo(HaveOccurred(), "Failed to initialize the operation")
			Expect(apiOperation.Name).To(Equal(operations.LroName))
		})
	})

	Context("Concurrency Guard", func() {
		It("should allow operations with same lastOperationdId as entity to run", func() {
			apiOperation := &LongRunningOperation{
				OperationId: lastOperationId,
			}
			entity := LongRunningEntity{
				LastOperationId: lastOperationId,
			}
			categorizedError := apiOperation.GuardConcurrency(ctx, entity)
			Expect(categorizedError).To(BeNil(), "Operation wasn't able to guard against concurrency")
		})
		It("should fail operations with different lastOperationId", func() {
			apiOperation := &LongRunningOperation{
				OperationId: lastOperationId,
			}
			entity := LongRunningEntity{
				LastOperationId: differentLastOperationId,
			}
			categorizedError := apiOperation.GuardConcurrency(ctx, entity)
			Expect(categorizedError).NotTo(BeNil(), "Operation didn't guard against concurrency when it should")
		})
	})

	Context("Running Operation and Checking Sleep Duration", func() {
		It("should run the operation and sleep for 20 seconds", func() {
			apiOperation := &LongRunningOperation{}
			start := time.Now()
			err := apiOperation.Run(ctx)
			elapsed := time.Since(start)

			Expect(err).NotTo(HaveOccurred(), "Operation did not run successfully")
			Expect(elapsed).To(BeNumerically(">=", 20*time.Second), "Run did not sleep for 20 seconds")
		})
	})
})
