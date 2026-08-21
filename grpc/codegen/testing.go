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
	return root
}

// CreateGRPCServices creates a new ServicesData instance for testing. The
// root is normalized first like the production Generate flow does before the
// generators read the design.
func CreateGRPCServices(root *expr.RootExpr) *ServicesData {
	codegen.NormalizeRoot(root)
	return NewServicesData(createServiceServices(root))
}

// createServiceServices performs the complete package declaration lifecycle
// required by transport test helpers.
func createServiceServices(root *expr.RootExpr) *service.ServicesData {
	generation := codegen.NewGeneration("goa.design/goa/example", nil)
	if err := service.Plan(root, generation); err != nil {
		panic(err)
	}
	if err := generation.Freeze(); err != nil {
		panic(err)
	}
	services, err := service.NewServicesData(root, generation)
	if err != nil {
		panic(err)
	}
	return services
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
