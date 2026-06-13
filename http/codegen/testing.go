package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// CreateHTTPServices creates a new ServicesData instance for testing. The
// root is normalized first like the production Generate flow does before the
// generators read the design.
func CreateHTTPServices(root *expr.RootExpr) *ServicesData {
	codegen.NormalizeRoot(root)
	return NewServicesData(service.NewServicesData(root), root.API.HTTP)
}
