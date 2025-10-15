package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// RunHTTPDSL returns the HTTP DSL root resulting from running the given DSL.
func RunHTTPDSL(t *testing.T, dsl func()) *expr.RootExpr {
	// reset all roots and codegen data structures
	root := expr.RunDSL(t, dsl)
	return root
}

// CreateHTTPServices creates a new ServicesData instance for testing.
func CreateHTTPServices(root *expr.RootExpr) *ServicesData {
	return NewServicesData(service.NewServicesData(root), root.API.HTTP)
}
