package logattrs

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogAttrs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogAttrs Suite")
}
