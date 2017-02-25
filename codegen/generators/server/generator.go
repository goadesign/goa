/*
Package server generates the code for a server. This includes:

    - A service package which defines the service interfaces
    - An endpoint package which defines endpoints that wrap the services methods.
    - transport packages for each of the transports defined in the design.
*/
package server

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/endpoint"
	"goa.design/goa.v2/codegen/service"
	"goa.design/goa.v2/design"
	restcodegen "goa.design/goa.v2/rest/codegen"
	rest "goa.design/goa.v2/rest/design"
)

// Writers returns the server writers.
func Writers(api *design.APIExpr) (writers []codegen.FileWriter) {
	for _, s := range design.Root.Services {
		writers = append(writers, service.Writer(api, s))
		writers = append(writers, endpoint.Writer(api, s))
	}
	if rest.RootExpr != nil {
		writers = append(writers, restcodegen.Writers(rest.RootExpr))
	}
	return
}
