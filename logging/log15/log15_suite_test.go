package goalog15_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLog15(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goalog15 Suite")
}
