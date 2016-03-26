package goakit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goakit Suite")
}
