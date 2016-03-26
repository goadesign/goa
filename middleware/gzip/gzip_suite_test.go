package gzip_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGzip(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gzip Suite")
}
