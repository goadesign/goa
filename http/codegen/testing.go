package codegen

import (
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// CreateHTTPServices creates a new ServicesData instance for testing.
func CreateHTTPServices(root *expr.RootExpr) *ServicesData {
	return NewServicesData(service.NewServicesData(root), root.API.HTTP)
}
