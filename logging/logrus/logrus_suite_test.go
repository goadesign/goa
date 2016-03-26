package goalogrus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLogrus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logrus Suite")
}
