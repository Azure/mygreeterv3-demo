package server

import (
	"context"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Server", func() {
	var (
		mockCtrl        *gomock.Controller
		mockClient      *mock.MockMyGreeterClient
		mockExternalClient *mock.MockMyGreeterClient
		s               *Server
		ctx             context.Context
		in              *pb.HelloRequest
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockMyGreeterClient(mockCtrl)
		mockExternalClient = mock.NewMockMyGreeterClient(mockCtrl)
		s = &Server{client: mockClient, externalClient: mockExternalClient}
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("SayHello", func() {
		Context("when client is not nil and returns a successful response", func() {
			BeforeEach(func() {
				in = &pb.HelloRequest{Name: "Alice", Age: 30, Email: "alice@example.com"}
				expectedReply := &pb.HelloReply{Message: "Hello Alice"}
				mockClient.EXPECT().SayHello(ctx, in).Return(expectedReply, nil)
			})

			It("should return the correct message", func() {
				out, err := s.SayHello(ctx, in)
				Expect(err).To(BeNil())
				Expect(out.Message).To(Equal("Hello Alice| appended by server"))
			})
		})

		Context("when client is nil and externalClient is not nil", func() {
			BeforeEach(func() {
				s.client = nil
				in = &pb.HelloRequest{Name: "Bob", Age: 25, Email: "bob@example.com"}
				expectedReply := &pb.HelloReply{Message: "Hello from external server"}
				mockExternalClient.EXPECT().SayHello(ctx, in).Return(expectedReply, nil)
			})

			It("should forward the request to the external server", func() {
				out, err := s.SayHello(ctx, in)
				Expect(err).To(BeNil())
				expectedMessage := "Hello from external server"
				Expect(out.Message).To(Equal(expectedMessage))
			})
		})

		Context("when both client and externalClient are nil", func() {
			BeforeEach(func() {
				s.client = nil
				s.externalClient = nil
				in = &pb.HelloRequest{Name: "Charlie", Age: 35, Email: "charlie@example.com"}
			})

			It("should return the echo message", func() {
				out, err := s.SayHello(ctx, in)
				Expect(err).To(BeNil())
				expectedMessage := "Echo back what you sent me (SayHello): Charlie 35 charlie@example.com"
				Expect(out.Message).To(Equal(expectedMessage))
			})
		})

		Context("when input name is 'TestPanic'", func() {
			BeforeEach(func() {
				in = &pb.HelloRequest{Name: "TestPanic", Age: 40, Email: "testpanic@example.com"}
			})

			It("should panic", func() {
				Expect(func() { s.SayHello(ctx, in) }).To(Panic())
			})
		})
	})
})
