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
		mockCtrl   *gomock.Controller
		mockClient *mock.MockMyGreeterClient
		s          *Server
		ctx        context.Context
		in         *pb.HelloRequest
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockMyGreeterClient(mockCtrl)
		s = &Server{client: mockClient}
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

		Context("when client is nil", func() {
			BeforeEach(func() {
				s.client = nil
				in = &pb.HelloRequest{Name: "Bob", Age: 25, Email: "bob@example.com"}
			})

			It("should return the echo message", func() {
				out, err := s.SayHello(ctx, in)
				Expect(err).To(BeNil())
				expectedMessage := "Echo back what you sent me (SayHello): Bob 25 bob@example.com"
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
