package server

import (
	"context"
	"errors"
	"net/http"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1/mock"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	azfake "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ = Describe("Mock Testing", func() {
	var (
		ctrl       *gomock.Controller
		mockClient *mock.MockMyGreeterClient
		s          *Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockMyGreeterClient(ctrl)
		s = &Server{
			client: mockClient,
		}
	})

	Context("when the server is available", func() {
		It("should not retry the request", func() {
			addr := &pb.Address{
				Street:  "123 Main St",
				City:    "Seattle",
				State:   "WA",
				Zipcode: 98012,
			}
			in := &pb.HelloRequest{Name: "Bob", Age: 53, Email: "test@test.com", Address: addr}
			mockClient.EXPECT().SayHello(gomock.Any(), gomock.Any()).Return(&pb.HelloReply{Message: "Hello, this is a successful mock response"}, nil).Times(1)
			out, err := s.SayHello(context.Background(), in)
			Expect(err).ToNot(HaveOccurred())
			Expect(out.Message).To(Equal("Hello, this is a successful mock response| appended by server"))
		})
	})

	Context("when the server is unavailable", func() {
		It("should retry the request", func() {
			addr := &pb.Address{
				Street:  "123 Main St",
				City:    "Seattle",
				State:   "WA",
				Zipcode: 98012,
			}
			in := &pb.HelloRequest{Name: "Bob", Age: 53, Email: "test@test.com", Address: addr}

			mockClient.EXPECT().SayHello(gomock.Any(), gomock.Any()).Return(nil, errors.New("server unavailable")).Times(2)
			_, err := s.SayHello(context.Background(), in)
			Expect(err).To(HaveOccurred())
			_, err = s.SayHello(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Fakes Unit Testing", func() {
	var (
		ctrl       *gomock.Controller
		mockClient *mock.MockMyGreeterClient
		s          *Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockMyGreeterClient(ctrl)

		fakeServer := fake.ResourceGroupsServer{
			BeginDelete: func(ctx context.Context, resourceGroupName string,
				options *armresources.ResourceGroupsClientBeginDeleteOptions) (resp azfake.PollerResponder[armresources.ResourceGroupsClientDeleteResponse],
				errResp azfake.ErrorResponder) {

				if resourceGroupName == "errorGroup" {
					errResp.SetError(errors.New("Cannot find resource group: errorGroup"))
					resp.SetTerminalResponse(http.StatusInternalServerError, armresources.ResourceGroupsClientDeleteResponse{}, nil)
					return
				}
				// Add non terminal response to represent long running operation
				// link to documentation: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azcore/fake#PollerResponder.AddNonTerminalResponse
				resp.AddNonTerminalResponse(http.StatusOK, nil)
				resp.SetTerminalResponse(http.StatusOK, armresources.ResourceGroupsClientDeleteResponse{}, nil)
				return
			},
			CreateOrUpdate: func(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroup,
				options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (resp azfake.Responder[armresources.ResourceGroupsClientCreateOrUpdateResponse], errResp azfake.ErrorResponder) {
				if resourceGroupName == "errorGroup" {
					errResp.SetError(errors.New("Cannot find resource group: errorGroup"))
					resp.SetResponse(http.StatusInternalServerError, armresources.ResourceGroupsClientCreateOrUpdateResponse{}, nil)
					return
				}
				resp.SetResponse(http.StatusOK, armresources.ResourceGroupsClientCreateOrUpdateResponse{ResourceGroup: armresources.ResourceGroup{ID: to.Ptr(resourceGroupName), Name: to.Ptr(resourceGroupName), Location: to.Ptr("westus")}}, nil)
				return
			},

			Get: func(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientGetOptions) (resp azfake.Responder[armresources.ResourceGroupsClientGetResponse], errResp azfake.ErrorResponder) {
				if resourceGroupName == "errorGroup" {
					errResp.SetError(errors.New("Cannot find resource group: errorGroup"))
					resp.SetResponse(http.StatusInternalServerError, armresources.ResourceGroupsClientGetResponse{}, nil)
					return
				}
				resp.SetResponse(http.StatusOK, armresources.ResourceGroupsClientGetResponse{ResourceGroup: armresources.ResourceGroup{ID: to.Ptr(resourceGroupName), Name: to.Ptr(resourceGroupName), Location: to.Ptr("westus")}}, nil)
				return
			},

			Update: func(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroupPatchable, options *armresources.ResourceGroupsClientUpdateOptions) (resp azfake.Responder[armresources.ResourceGroupsClientUpdateResponse], errResp azfake.ErrorResponder) {
				if resourceGroupName == "errorGroup" {
					errResp.SetError(errors.New("Cannot find resource group: errorGroup"))
					resp.SetResponse(http.StatusInternalServerError, armresources.ResourceGroupsClientUpdateResponse{}, nil)
					return
				}
				resp.SetResponse(http.StatusOK, armresources.ResourceGroupsClientUpdateResponse{ResourceGroup: armresources.ResourceGroup{ID: to.Ptr(resourceGroupName), Name: to.Ptr(resourceGroupName), Location: to.Ptr("westus")}}, nil)
				return
			},

			NewListPager: func(options *armresources.ResourceGroupsClientListOptions) (resp azfake.PagerResponder[armresources.ResourceGroupsClientListResponse]) {
				resourceGroupList := armresources.ResourceGroupListResult{
					Value: []*armresources.ResourceGroup{
						{ID: to.Ptr("TestGroup1"), Name: to.Ptr("TestGroup"), Location: to.Ptr("westus")},
						{ID: to.Ptr("TestGroup2"), Name: to.Ptr("TestGroup"), Location: to.Ptr("westus")},
					},
				}
				resp.AddPage(http.StatusOK, armresources.ResourceGroupsClientListResponse{ResourceGroupListResult: resourceGroupList}, nil)
				return
			},
		}

		client, err := armresources.NewResourceGroupsClient("subscriptionID", &azfake.TokenCredential{}, &arm.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				Transport: fake.NewResourceGroupsServerTransport(&fakeServer),
			},
		})
		Expect(err).NotTo(HaveOccurred())
		s = &Server{ResourceGroupClient: client, client: mockClient}
	})

	Context("CRUDL Operations", func() {
		It("CreateResourceGroup() should be successful", func() {
			in := &pb.CreateResourceGroupRequest{Name: "TestGroup", Region: "westus"}
			_, err := s.CreateResourceGroup(context.Background(), in)
			Expect(err).ToNot(HaveOccurred())
		})

		It("ReadResourceGroup() should be successful", func() {
			in := &pb.ReadResourceGroupRequest{Name: "TestGroup"}
			_, err := s.ReadResourceGroup(context.Background(), in)
			Expect(err).ToNot(HaveOccurred())
		})

		It("UpdateResourceGroup() should be successful", func() {
			in := &pb.UpdateResourceGroupRequest{Name: "TestGroup", Tags: map[string]string{"key": "value"}}
			_, err := s.UpdateResourceGroup(context.Background(), in)
			Expect(err).ToNot(HaveOccurred())
		})

		It("DeleteResourceGroup() should be successful", func() {
			in := &pb.DeleteResourceGroupRequest{Name: "TestGroup"}
			_, err := s.DeleteResourceGroup(context.Background(), in)
			Expect(err).ToNot(HaveOccurred())
		})

		It("ListResourceGroup() should be successful", func() {
			_, err := s.ListResourceGroups(context.Background(), &emptypb.Empty{})
			Expect(err).ToNot(HaveOccurred())
		})

	})

	Context("When ResourceGroupClient is nil", func() {
		BeforeEach(func() {
			s.ResourceGroupClient = nil
		})

		It("CreateResourceGroup() should return an error", func() {
			in := &pb.CreateResourceGroupRequest{Name: "TestGroup", Region: "westus"}
			_, err := s.CreateResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})

		It("ReadResourceGroup() should return an error", func() {
			in := &pb.ReadResourceGroupRequest{Name: "TestGroup"}
			_, err := s.ReadResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})

		It("UpdateResourceGroup() should return an error", func() {
			in := &pb.UpdateResourceGroupRequest{Name: "TestGroup", Tags: map[string]string{"key": "value"}}
			_, err := s.UpdateResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})

		It("DeleteResourceGroup() should return an error", func() {
			in := &pb.DeleteResourceGroupRequest{Name: "TestGroup"}
			_, err := s.DeleteResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})

		It("ListResourceGroup() should return an error", func() {
			_, err := s.ListResourceGroups(context.Background(), &emptypb.Empty{})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When ResourceGroupClient encounters an error", func() {
		It("CreateResourceGroup() should return an error for 'errorGroup'", func() {
			in := &pb.CreateResourceGroupRequest{Name: "errorGroup", Region: "westus"}
			_, err := s.CreateResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})

		It("ReadResourceGroup() should return an error for 'errorGroup'", func() {
			in := &pb.ReadResourceGroupRequest{Name: "errorGroup"}
			_, err := s.ReadResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})

		It("UpdateResourceGroup() should return an error for 'errorGroup'", func() {
			in := &pb.UpdateResourceGroupRequest{Name: "errorGroup"}
			_, err := s.UpdateResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})

		It("DeleteResourceGroup() should return an error for 'errorGroup'", func() {
			in := &pb.DeleteResourceGroupRequest{Name: "errorGroup"}
			_, err := s.DeleteResourceGroup(context.Background(), in)
			Expect(err).To(HaveOccurred())
		})
	})
})
