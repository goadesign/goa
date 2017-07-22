package rest

import (
	"testing"

	"goa.design/goa.v2/codegen/service"
	"goa.design/goa.v2/design/rest"
)

// RunRestDSL returns the rest DSL root resulting from running the given DSL.
func RunRestDSL(t *testing.T, dsl func()) *rest.RootExpr {
	// reset all roots and codegen data structures
	service.Services = make(service.ServicesData)
	HTTPServices = make(ServicesData)
	return rest.RunRestDSL(t, dsl)
}
