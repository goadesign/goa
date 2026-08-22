// This file verifies one gRPC plan keeps the exact design, service plan, and
// generated files selected for a generation run.
package codegen

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestNewPlansKeepsExactInputs checks that each result keeps its input pair.
func TestNewPlansKeepsExactInputs(t *testing.T) {
	roots := grpcPlanRoots(t, "First", "Second")
	generation, services := grpcServicePlans(t, roots)
	plans, err := NewPlans(generation,
		PlanInput{Root: roots[1], Service: services[1]},
		PlanInput{Root: roots[0], Service: services[0]},
	)
	require.NoError(t, err)
	require.Same(t, generation, plans[0].Generation())
	require.Same(t, roots[1], plans[0].Root())
	require.Same(t, services[1], plans[0].Service())
	require.Same(t, roots[0], plans[1].Root())
	require.Same(t, services[0], plans[1].Service())
}

// TestNewPlansRequiresEveryRoot checks that a batch cannot omit a design.
func TestNewPlansRequiresEveryRoot(t *testing.T) {
	roots := grpcPlanRoots(t, "First", "Second")
	generation, services := grpcServicePlans(t, roots)
	_, err := NewPlans(generation, PlanInput{Root: roots[0], Service: services[0]})
	require.EqualError(t, err, "gRPC planning requires all 2 gRPC roots, got 1")
}

// TestNewPlansRejectsDuplicateRoot checks that a batch cannot repeat a design.
func TestNewPlansRejectsDuplicateRoot(t *testing.T) {
	roots := grpcPlanRoots(t, "First", "Second")
	generation, services := grpcServicePlans(t, roots)
	_, err := NewPlans(generation,
		PlanInput{Root: roots[0], Service: services[0]},
		PlanInput{Root: roots[0], Service: services[0]},
	)
	require.EqualError(t, err, fmt.Sprintf("gRPC root %p is planned more than once", roots[0]))
}

// TestNewPlansRejectsMismatchedServicePlan checks that input pairs must match.
func TestNewPlansRejectsMismatchedServicePlan(t *testing.T) {
	roots := grpcPlanRoots(t, "First", "Second")
	generation, services := grpcServicePlans(t, roots)
	_, err := NewPlans(generation,
		PlanInput{Root: roots[0], Service: services[1]},
		PlanInput{Root: roots[1], Service: services[0]},
	)
	require.EqualError(t, err, "gRPC plan input does not pair a design with its service plan")
}

// TestPlanLinksOnceAfterFreeze checks the required link order.
func TestPlanLinksOnceAfterFreeze(t *testing.T) {
	roots := grpcPlanRoots(t, "Calc")
	generation, services := grpcServicePlans(t, roots)
	plans, err := NewPlans(generation, PlanInput{Root: roots[0], Service: services[0]})
	require.NoError(t, err)
	require.EqualError(t, plans[0].Link(), "gRPC plan cannot link before generation freeze")
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())
	require.EqualError(t, plans[0].Link(), "gRPC plan is already linked")
}

// TestPlanReturnsStoredFiles checks that later reads reuse the linked files.
func TestPlanReturnsStoredFiles(t *testing.T) {
	roots := grpcPlanRoots(t, "Calc")
	generation, services := grpcServicePlans(t, roots)
	plans, err := NewPlans(generation, PlanInput{Root: roots[0], Service: services[0]})
	require.NoError(t, err)
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())
	want := grpcPlanFileSignatures(t, plans[0])
	roots[0].API.GRPC.Services = append(roots[0].API.GRPC.Services, &expr.GRPCServiceExpr{})
	require.Equal(t, want, grpcPlanFileSignatures(t, plans[0]))
	require.Equal(t, want, grpcPlanFileSignatures(t, plans[0]))
}

// TestNewPlansIsIndependentOfInputOrder checks that input order does not
// change files written to the same generated packages.
func TestNewPlansIsIndependentOfInputOrder(t *testing.T) {
	forwardNames, forwardErr := collidingGRPCPlanResult(t, false)
	reverseNames, reverseErr := collidingGRPCPlanResult(t, true)
	require.Empty(t, forwardErr)
	require.Empty(t, reverseErr)
	require.Equal(t, forwardNames, reverseNames)
}

// grpcPlanRoots creates independent designs with one unary gRPC method.
func grpcPlanRoots(t *testing.T, serviceNames ...string) []*expr.RootExpr {
	t.Helper()
	roots := make([]*expr.RootExpr, len(serviceNames))
	for index, serviceName := range serviceNames {
		roots[index] = expr.RunDSL(t, func() {
			dsl.Service(serviceName, func() {
				dsl.Method("Read", func() {
					dsl.Payload(func() { dsl.Field(1, "value", dsl.String) })
					dsl.Result(func() { dsl.Field(1, "value", dsl.String) })
					dsl.GRPC(func() {})
				})
			})
		})
	}
	return roots
}

// grpcServicePlans creates one generation and the service plan for each root.
func grpcServicePlans(t *testing.T, roots []*expr.RootExpr) (*codegen.Generation, []*service.Plan) {
	t.Helper()
	evaluated := make([]eval.Root, len(roots))
	inputs := make([]service.PlanInput, len(roots))
	for index, root := range roots {
		evaluated[index] = root
		inputs[index] = service.PlanInput{
			Root:     root,
			Examples: expr.NewExampleGenerator(root.API.RandomizerFactory),
		}
	}
	generation, err := codegen.NewGeneration("generated.local/gen", evaluated)
	require.NoError(t, err)
	plans, err := service.NewPlans(generation, inputs...)
	require.NoError(t, err)
	return generation, plans
}

// grpcPlanFileSignatures renders every stored file for a stable comparison.
func grpcPlanFileSignatures(t *testing.T, plan *Plan) []string {
	t.Helper()
	files := plan.ProtoFiles()
	files = append(files, plan.ServerFiles()...)
	files = append(files, plan.ClientFiles()...)
	files = append(files, plan.ServerTypeFiles()...)
	files = append(files, plan.ClientTypeFiles()...)
	files = append(files, plan.ClientCLIFiles()...)
	files = append(files, plan.ExampleServerFiles()...)
	files = append(files, plan.ExampleCLIFiles()...)
	signatures := make([]string, len(files))
	for index, file := range files {
		signatures[index] = file.Path + "\n" + sectionCode(t, file.SectionTemplates...)
	}
	sort.Strings(signatures)
	return signatures
}

// collidingGRPCPlanResult returns stored files or the exact planning error for
// two designs that write the same gRPC and protobuf packages.
func collidingGRPCPlanResult(t *testing.T, reverse bool) ([]string, string) {
	t.Helper()
	makeRoot := func(serviceName, typeName string) *expr.RootExpr {
		return expr.RunDSL(t, func() {
			choice := dsl.Type(typeName, func() {
				dsl.Meta("struct:name:proto", "API2_Choice")
				dsl.OneOf("state", func() {
					dsl.Field(1, "api2URL", dsl.String)
					dsl.Field(2, "reset", dsl.String)
				})
			})
			dsl.Service(serviceName, func() {
				dsl.Method("Sync2URL", func() {
					dsl.Payload(choice)
					dsl.Result(choice)
					dsl.StreamingPayload(choice)
					dsl.StreamingResult(choice)
					dsl.GRPC(func() {})
				})
			})
		})
	}
	roots := []*expr.RootExpr{makeRoot("Foo Bar", "FirstChoice"), makeRoot("Foo-Bar", "SecondChoice")}
	generation, services := grpcServicePlans(t, roots)
	inputs := []PlanInput{{Root: roots[0], Service: services[0]}, {Root: roots[1], Service: services[1]}}
	if reverse {
		inputs[0], inputs[1] = inputs[1], inputs[0]
	}
	plans, err := NewPlans(generation, inputs...)
	if err != nil {
		return nil, err.Error()
	}
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	for _, servicePlan := range services {
		require.NoError(t, servicePlan.Link())
	}
	var names []string
	for _, plan := range plans {
		if err := plan.Link(); err != nil {
			return nil, err.Error()
		}
		names = append(names, grpcPlanFileSignatures(t, plan)...)
	}
	sort.Strings(names)
	return names, ""
}
