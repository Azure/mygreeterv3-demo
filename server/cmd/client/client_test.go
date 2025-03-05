package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/server/internal/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Cobra Cmd test", func() {
	var testServer *server.Server
	var serverPort int
	var cmd *cobra.Command

	BeforeEach(func() {
		options := server.Options{
			Port:                0,
			HTTPPort:            0,
			JsonLog:             true,
			EnableAzureSDKCalls: false,
			SubscriptionID:      "test",
			RemoteAddr:          "",
		}
		testServer = server.NewServer()
		testServer.Init(options)
		testServer.Serve(options)

		serverPort = testServer.GrpcPort

		Eventually(func() bool {
			return testServer.IsRunning()
		}, 10*time.Second).Should(BeTrue())

		cmd = &cobra.Command{
			Use:   "hello",
			Short: "Call SayHello",
			Run:   hello,
		}
	})

	AfterEach(func() {
		if testServer != nil && testServer.IsRunning() {
			testServer.Cleanup()
		}
		testServer = nil
		serverPort = 0
		SetOutput(os.Stdout)
	})

	It("should call Execute() and log the response message", func() {
		var buf bytes.Buffer
		SetOutput(&buf)

		host := fmt.Sprintf("localhost:%d", serverPort)
		options.RemoteAddr = host
		options.JsonLog = true

		hello(cmd, nil)
		Expect(buf.String()).To(ContainSubstring("Echo back what you sent me (SayHello)"))
	})

	It("should call Execute() and log error", func() {
		var buf bytes.Buffer
		SetOutput(&buf)

		options.RemoteAddr = "localhost:0"
		options.JsonLog = true

		hello(cmd, nil)
		Expect(buf.String()).To(ContainSubstring("connect: connection refused"))
	})
})
