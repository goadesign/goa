/*
Package openapi generates the OpenAPI specification of the service.
*/
package openapi

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
	rest "goa.design/goa.v2/rest/design"
)

// Writers returns the server writers.
func Writers(d *design.RootExpr, r *rest.RootExpr) (ws []codegen.FileWriter) {
	return
}
