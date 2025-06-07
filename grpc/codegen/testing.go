package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// RunGRPCDSL returns the GRPC DSL root resulting from running the given DSL.
// It is used only in tests.
func RunGRPCDSL(t *testing.T, dsl func()) *expr.RootExpr {
	// reset all roots and codegen data structures
	root := expr.RunDSL(t, dsl)
	service.Services = service.ServicesData{Services: make(map[string]*service.Data), Root: root}
	GRPCServices = ServicesData{Services: make(map[string]*ServiceData), Root: root}
	return root
}

func sectionCode(t *testing.T, section ...*codegen.SectionTemplate) string {
	t.Helper()
	var code bytes.Buffer
	for _, s := range section {
		if err := s.Write(&code); err != nil {
			t.Fatal(err)
		}
	}
	return code.String()
}
