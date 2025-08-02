package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// RunJSONRPCDSL returns the DSL root resulting from running the given DSL.
// Used only in tests.
func RunJSONRPCDSL(t *testing.T, dsl func()) *expr.RootExpr {
	// Use the existing expr.RunDSL function
	root := expr.RunDSL(t, dsl)
	return root
}

// CreateJSONRPCServices creates a new ServicesData instance for JSON-RPC testing.
func CreateJSONRPCServices(root *expr.RootExpr) *httpcodegen.ServicesData {
	services := service.NewServicesData(root)
	return httpcodegen.NewServicesData(services, &root.API.JSONRPC.HTTPExpr)
}