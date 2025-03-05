package longRunningOperation

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test entry point for Ginkgo
func TestLongRunningOperation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LongRunningOperation Suite")
}
