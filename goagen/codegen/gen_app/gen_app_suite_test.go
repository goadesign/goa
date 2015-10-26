package genapp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GenApp Suite")
}
