package codegen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoagen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goagen Suite")
}
