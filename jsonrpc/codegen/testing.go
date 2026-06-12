package codegen

import (
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// CreateJSONRPCServices creates a new ServicesData instance for JSON-RPC testing.
func CreateJSONRPCServices(root *expr.RootExpr) *httpcodegen.ServicesData {
	services := service.NewServicesData(root)
	return httpcodegen.NewJSONRPCServicesData(services, &root.API.JSONRPC.HTTPExpr)
}
