package http

import (
	"testing"

	"goa.design/goa.v2/codegen/service"
	"goa.design/goa.v2/design/http"
)

// RunHTTPDSL returns the HTTP DSL root resulting from running the given DSL.
func RunHTTPDSL(t *testing.T, dsl func()) *http.RootExpr {
	// reset all roots and codegen data structures
	service.Services = make(service.ServicesData)
	HTTPServices = make(ServicesData)
	return http.RunHTTPDSL(t, dsl)
}
