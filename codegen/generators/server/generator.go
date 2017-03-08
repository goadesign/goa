/*
Package server generates the code for a server. This includes:

    - A service package which defines the service interfaces
    - An endpoint package which defines endpoints that wrap the services methods.
    - transport packages for each of the transports defined in the design.
*/
package server

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/writers"
	"goa.design/goa.v2/design"
	restcodegen "goa.design/goa.v2/rest/codegen"
	rest "goa.design/goa.v2/rest/design"
)

// Writers returns the server writers.
func Writers(d *design.RootExpr, r *rest.RootExpr) (ws []codegen.FileWriter) {
	for _, s := range d.Services {
		ws = append(ws, writers.ServiceWriter(api, s))
		ws = append(ws, writers.EndpointsWriter(api, s))
	}
	if r != nil {
		ws = append(ws, restcodegen.ServerWriters(r)...)
	}
	return
}
