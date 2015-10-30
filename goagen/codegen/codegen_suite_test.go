package codegen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCodegen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Codegen Suite")
}
