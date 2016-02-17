package dslengine_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDslengine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dslengine Suite")
}
