/*
Package client generates the code for a client. This includes:

    - A service package which defines the service interfaces
    - An endpoint package which defines endpoints that wrap the services methods.
    - transport packages for each of the transports defined in the design.
*/
package client

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/writers"
	"goa.design/goa.v2/design"
	rest "goa.design/goa.v2/rest/design"
)

// Writers returns the server writers.
func Writers(d *design.RootExpr, r *rest.RootExpr) []codegen.FileWriter {
	var ws []codegen.FileWriter
	for _, s := range d.Services {
		ws = append(ws, writers.ServiceWriter(d.API, s))
		ws = append(ws, writers.EndpointsWriter(d.API, s))
	}
	if r != nil {
		// ws = append(ws, restcodegen.ClientWriters(r)...)
	}
	return ws
}
