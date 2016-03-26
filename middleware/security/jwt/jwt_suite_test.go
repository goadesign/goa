package jwt_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestJWTSecurityMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JWT Security Middleware")
}
