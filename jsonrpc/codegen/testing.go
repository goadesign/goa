package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// CreateJSONRPCServices creates a new ServicesData instance for JSON-RPC
// testing. The root is normalized first like the production Generate flow
// does before the generators read the design.
func CreateJSONRPCServices(root *expr.RootExpr) *httpcodegen.ServicesData {
	codegen.NormalizeRoot(root)
	services := service.NewServicesData(root)
	return httpcodegen.NewJSONRPCServicesData(services, &root.API.JSONRPC.HTTPExpr)
}
