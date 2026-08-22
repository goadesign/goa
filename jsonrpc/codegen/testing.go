// This file builds JSON-RPC code-generation analysis in tests using the same
// generation construction, planning, freezing, and rendering as production.
package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// CreateJSONRPCPlan builds and links the same service, HTTP, and JSON-RPC plans
// that production generation uses.
func CreateJSONRPCPlan(root *expr.RootExpr) *Plan {
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	if err != nil {
		panic(err)
	}
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	if err != nil {
		panic(err)
	}
	var applicationHTTP *httpcodegen.Plan
	if len(root.API.HTTP.Services) > 0 {
		applicationPlans, err := httpcodegen.NewPlans(generation, httpcodegen.PlanInput{
			Root:    root,
			Service: servicePlan,
		})
		if err != nil {
			panic(err)
		}
		applicationHTTP = applicationPlans[0]
	}
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{
		Root:    root,
		Service: servicePlan,
	})
	if err != nil {
		panic(err)
	}
	plans, err := NewPlans(generation, PlanInput{
		Root:            root,
		Service:         servicePlan,
		HTTP:            httpPlans[0],
		ApplicationHTTP: applicationHTTP,
	})
	if err != nil {
		panic(err)
	}
	if err := generation.Freeze(); err != nil {
		panic(err)
	}
	if err := servicePlan.Link(); err != nil {
		panic(err)
	}
	if applicationHTTP != nil {
		if err := applicationHTTP.Link(); err != nil {
			panic(err)
		}
	}
	if err := httpPlans[0].Link(); err != nil {
		panic(err)
	}
	if err := plans[0].Link(); err != nil {
		panic(err)
	}
	return plans[0]
}
