package genschema_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenSchema(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GenSchema Suite")
}
