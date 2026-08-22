// This file builds gRPC code-generation analysis in tests using the same
// generation construction, planning, freezing, and rendering as production.
package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// RunGRPCDSL returns the GRPC DSL root resulting from running the given DSL.
// It is used only in tests.
func RunGRPCDSL(t *testing.T, dsl func()) *expr.RootExpr {
	// reset all roots and codegen data structures
	root := expr.RunDSL(t, dsl)
	return root
}

// CreateGRPCServices creates a new ServicesData instance for testing.
// Generation construction normalizes the root before any planner reads it.
func CreateGRPCServices(root *expr.RootExpr) *ServicesData {
	return createServiceServices(root)
}

// createServiceServices performs the complete package declaration lifecycle
// required by transport test helpers.
func createServiceServices(root *expr.RootExpr) *ServicesData {
	return createServiceServicesForPackage(root, "generated.local/gen")
}

// createServiceServicesForPackage builds test service analysis for the exact
// generated module path whose imports the test renders.
func createServiceServicesForPackage(root *expr.RootExpr, genpkg string) *ServicesData {
	generation, err := codegen.NewGeneration(genpkg, []eval.Root{root})
	if err != nil {
		panic(err)
	}
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	if err != nil {
		panic(err)
	}
	grpcPlan, err := Plan(generation, PlanInput{Root: root, Service: servicePlan})
	if err != nil {
		panic(err)
	}
	if err := example.Plan(generation); err != nil {
		panic(err)
	}
	if err := generation.Freeze(); err != nil {
		panic(err)
	}
	if err := servicePlan.Link(); err != nil {
		panic(err)
	}
	return NewServicesData(servicePlan.Services(), grpcPlan)
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
