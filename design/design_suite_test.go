package design_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDesign(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Design Suite")
}
