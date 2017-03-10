/*
Package openapi generates the OpenAPI specification of the service.
*/
package openapi

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/writers"
	"goa.design/goa.v2/design"
	rest "goa.design/goa.v2/rest/design"
)

// Writers returns the server writers.
func Writers(d *design.RootExpr, r *rest.RootExpr) []codegen.FileWriter {
	var ws []codegen.FileWriter
	// TBD actual OpenAPI writers
	for _, s := range d.Services {
		ws = append(ws, writers.ServiceWriter(d.API, s))
		ws = append(ws, writers.EndpointsWriter(d.API, s))
	}
	return ws
}
