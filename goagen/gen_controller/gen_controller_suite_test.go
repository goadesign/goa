package gencontroller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GenController Suite")
}
