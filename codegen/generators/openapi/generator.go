/*
Package openapi generates the OpenAPI specification of the service.
*/
package openapi

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/writers"
	rest "goa.design/goa.v2/rest/design"
)

// RestWriters returns the HTTP server writers.
func RestWriters(r *rest.RootExpr) ([]codegen.FileWriter, error) {
	w, err := writers.OpenAPI(r)
	if err != nil {
		return nil, err
	}
	return []codegen.FileWriter{w}, nil
}
