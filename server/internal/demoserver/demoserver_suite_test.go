package demoserver

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDemoserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Demoserver Suite")
}

var _ = Describe("ExternalServer", func() {
	Context("when forwarding SayHello request to externalServer", func() {
		It("should return the response from externalServer", func() {
			// Add test logic here
		})
	})
})
