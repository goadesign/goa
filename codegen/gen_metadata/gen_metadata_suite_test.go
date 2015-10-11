package genmetadata_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GenMetadata Suite")
}
