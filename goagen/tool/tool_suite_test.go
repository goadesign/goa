package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tool Suite")
}
