package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
)

// RunGRPCDSL returns the GRPC DSL root resulting from running the given DSL.
func RunGRPCDSL(t *testing.T, dsl func()) *expr.RootExpr {
	// reset all roots and codegen data structures
	service.Services = make(service.ServicesData)
	GRPCServices = make(ServicesData)
	return expr.RunGRPCDSL(t, dsl)
}

func sectionCode(t *testing.T, section ...*codegen.SectionTemplate) string {
	var code bytes.Buffer
	for _, s := range section {
		if err := s.Write(&code); err != nil {
			t.Fatal(err)
		}
	}
	return code.String()
}
