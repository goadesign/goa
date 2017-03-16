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
	rest "goa.design/goa.v2/rest/design"
)

// RestWriters returns the HTTP server writers.
func RestWriters(r *rest.RootExpr) ([]codegen.FileWriter, error) {
	var ws []codegen.FileWriter
	for _, s := range r.Design.Services {
		ws = append(ws, writers.Service(r.Design.API, s))
		ws = append(ws, writers.Endpoint(r.Design.API, s))
	}
	if r != nil {
		// ws = append(ws, restcodegen.ServerWriters(r)...)
	}
	return ws, nil
}
