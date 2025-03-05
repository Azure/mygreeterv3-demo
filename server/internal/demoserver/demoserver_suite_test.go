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
