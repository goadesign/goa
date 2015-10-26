package goa_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goa Suite")
}
