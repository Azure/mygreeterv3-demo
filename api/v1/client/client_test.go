package client

import (
	log "log/slog"

	"github.com/Azure/aks-middleware/grpc/interceptor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	attrs := []log.Attr{}
	It("should create a new client", func() {
		_, err := NewClient("localhost:50051", interceptor.GetClientInterceptorLogOptions(log.Default(), attrs))
		Expect(err).To((BeNil()))
	})

	It("should return an error for an invalid address", func() {
		_, err := NewClient("", interceptor.GetClientInterceptorLogOptions(log.Default(), attrs))
		Expect(err).To(Not(BeNil()))

	})
})
