package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1/client"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1/restsdk"
	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/logattrs"
	"github.com/Azure/aks-middleware/grpc/interceptor"
	"github.com/Azure/aks-middleware/http/common"

	log "log/slog"

	"github.com/Azure/aks-middleware/http/client/direct/restlogger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func sayHello(buf *bytes.Buffer, port int, req *pb.HelloRequest) {
	logger := log.New(log.NewJSONHandler(buf, nil))
	host := fmt.Sprintf("localhost:%d", port)
	options := interceptor.GetClientInterceptorLogOptions(logger, logattrs.GetAttrs())
	options.APIOutput = buf
	client, err := client.NewClient(host, options)
	// logging the error for transparency, but not failing the test since retry interceptor will handle it
	if err != nil {
		log.Error("did not connect: " + err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client.SayHello(ctx, req)
}

// TODO: Allow user to filter out serial test cases
// These tests cannot be run in parallel
// - TestRetryWhenUnavailable() may send req to port server is running on
var _ = Describe("Interceptor test", func() {

	var server *Server
	var serverPort int
	var in *pb.HelloRequest

	BeforeEach(func() {
		addr := &pb.Address{
			Street:  "123 Main St",
			City:    "Seattle",
			State:   "WA",
			Zipcode: 98012,
		}
		in = &pb.HelloRequest{Name: "Bob", Age: 53, Email: "test@test.com", Address: addr}
	})

	AfterEach(func() {
		if server != nil && server.IsRunning() {
			server.Cleanup()
		}
		server = nil
		serverPort = 0
	})

	Context("when initializing the server", func() {
		It("should correctly initialize the server based on the provided options", func() {
			s := NewServer()

			options := Options{
				EnableAzureSDKCalls: false,
				SubscriptionID:      "test-subscription-id",
				JsonLog:             true,
				RemoteAddr:          "localhost:50051",
				Port:                0,
				HTTPPort:            0,
			}

			s.Init(options)
			Expect(s.ResourceGroupClient).To(BeNil())
			Expect(s.client).ToNot(BeNil())
		})
	})

	Context("when the server is available", func() {
		BeforeEach(func() {
			options := Options{
				Port:                0, // Use 0 to let the system assign an available port
				HTTPPort:            0, // Same for HTTP port
				JsonLog:             true,
				EnableAzureSDKCalls: true,
				SubscriptionID:      "test",
				RemoteAddr:          "",
			}
			server = NewServer()
			server.Init(options)
			server.Serve(options)

			serverPort = server.GrpcPort

			// Explicitly testing state of server
			// Continue with tests once server and grpc-gateway are up and running
			Eventually(func() bool {
				return server.IsRunning()
			}, 10*time.Second).Should(BeTrue())
		})

		It("should not retry the request", func() {
			var buf bytes.Buffer
			sayHello(&buf, serverPort, in)
			Expect(strings.Count(buf.String(), "headers")).To(Equal(1))
			Expect(strings.Count(buf.String(), "OK")).To(Equal(1)) // must not retry
		})

		It("should validate the name length", func() {
			var buf bytes.Buffer
			in.Name = "Z"
			sayHello(&buf, serverPort, in)
			Expect(strings.Count(buf.String(), "headers")).To(Equal(1))
			Expect(buf.String()).To(ContainSubstring("value length must be at least 2 characters")) // must return error b/c name < 2 letters
		})

		It("should validate the age range", func() {
			var buf bytes.Buffer
			in.Age = 353
			sayHello(&buf, serverPort, in)
			Expect(strings.Count(buf.String(), "headers")).To(Equal(1))
			Expect(buf.String()).To(ContainSubstring("value must be greater than or equal to 1 and less than 150")) // must return error b/c age > 150
		})

		It("should validate the email format", func() {
			var buf bytes.Buffer
			in.Email = "test"
			sayHello(&buf, serverPort, in)
			Expect(strings.Count(buf.String(), "headers")).To(Equal(1))
			Expect(buf.String()).To(ContainSubstring("value does not match regex pattern")) // must return error b/c invalid email
		})

		It("should recover from panic", func() {
			os.Setenv("AKS_BIN_VERSION_GITBRANCH", "tomabraham/service")
			var buf bytes.Buffer
			in.Name = "TestPanic"
			sayHello(&buf, serverPort, in) // "TestPanic" is a special name that triggers panic
			Expect(buf.String()).To(ContainSubstring("SayHello.go, line:"))
			Expect(strings.Count(buf.String(), "code = Internal")).To(Equal(1)) // must handle panic by returning gRPC code unknown and output filename/line # of panic
		})
	})

	Context("when the server is unavailable", func() {
		It("should retry the request", func() {
			var buf bytes.Buffer
			sayHello(&buf, serverPort, in)
			Expect(strings.Count(buf.String(), "headers")).To(BeNumerically(">", 1)) // must reStry
		})
	})
})

var _ = Describe("REST call test", func() {

	var server *Server
	var httpPort int

	var logger *log.Logger
	var cfg *restsdk.Configuration
	var apiClient *restsdk.APIClient
	var service *restsdk.MyGreeterApiService
	var helloRequestBody restsdk.HelloRequest

	BeforeEach(func() {
		options := Options{
			Port:                0,
			HTTPPort:            0,
			JsonLog:             true,
			EnableAzureSDKCalls: true,
			SubscriptionID:      "test",
			RemoteAddr:          "",
		}
		server = NewServer()
		server.Init(options)
		server.Serve(options)

		httpPort = server.HttpPort

		Eventually(func() bool {
			return server.IsRunning()
		}, 10*time.Second).Should(BeTrue())

		// Common setup
		logger = log.New(log.NewJSONHandler(os.Stdout, nil))
		cfg = &restsdk.Configuration{
			BasePath:      fmt.Sprintf("http://0.0.0.0:%d", httpPort),
			DefaultHeader: make(map[string]string),
			UserAgent:     "Swagger-Codegen/1.0.0/go",
			HTTPClient:    restlogger.NewLoggingClient(logger),
		}
		apiClient = restsdk.NewAPIClient(cfg)
		service = apiClient.MyGreeterApi
		helloRequestBody = restsdk.HelloRequest{
			Name:  "MyName",
			Age:   53,
			Email: "test@test.com",
		}
	})

	AfterEach(func() {
		if server != nil && server.IsRunning() {
			server.Cleanup()
		}
		server = nil
		httpPort = 0
	})

	Context("when sending a REST request", func() {
		It("should return successfully when making SayHello call", func() {
			resp, _, err := service.MyGreeterSayHello(context.Background(), helloRequestBody)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Message).To(ContainSubstring("Echo back what you sent me (SayHello): MyName 53 test@test.com"))
		})

		It("should only log out headers converted from metadata that align with outgoing header matcher", func() {
			var buf bytes.Buffer
			logger = log.New(log.NewJSONHandler(&buf, nil))
			cfg.HTTPClient = restlogger.NewLoggingClient(logger)

			// Set specific headers
			cfg.DefaultHeader = map[string]string{
				common.RequestCorrelationIDHeader:      "test-correlation-id",
				common.RequestAcsOperationIDHeader:     "test-operation-id",
				common.RequestARMClientRequestIDHeader: "test-client-request-id",
			}

			resp, _, err := service.MyGreeterSayHello(context.Background(), helloRequestBody)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Message).To(ContainSubstring("Echo back what you sent me (SayHello): MyName 53 test@test.com"))

			logs := buf.String()
			Expect(logs).ToNot(ContainSubstring("test-correlation-id"))
			Expect(logs).To(ContainSubstring("test-operation-id"))
			Expect(logs).To(ContainSubstring("test-client-request-id"))
		})
	})
})
