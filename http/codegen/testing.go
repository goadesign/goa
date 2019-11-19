package codegen

import (
	"os"
	"testing"

	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// RunHTTPDSL returns the HTTP DSL root resulting from running the given DSL.
func RunHTTPDSL(t *testing.T, dsl func()) *expr.RootExpr {
	// reset all roots and codegen data structures
	service.Services = make(service.ServicesData)
	HTTPServices = make(ServicesData)
	return expr.RunDSL(t, dsl)
}

// makeGolden returns a file object used to write test expectations. If
// makeGolden returns nil then the test should not generate test
// expectations.
func makeGolden(t *testing.T, p string) *os.File {
	t.Helper()
	if os.Getenv("GOLDEN") == "" {
		return nil
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
