package apidsl_test

import (
	"github.com/goadesign/goa/design/apidsl"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDsl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dsl Suite")
}

var _ = BeforeSuite(func() {
	apidsl.InitDesign()
})
