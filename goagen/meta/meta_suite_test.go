package meta_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMeta(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Meta Suite")
}
