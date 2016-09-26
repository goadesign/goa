package genswagger_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenSwagger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GenSwagger Suite")
}
