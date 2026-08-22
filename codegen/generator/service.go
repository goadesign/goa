// This file assembles service-owned generated files after every participating
// Goa design root has planned and frozen its package declarations.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// serviceFiles returns the service files described by plan's frozen package
// declarations and run-owned example state.
func serviceFiles(plan *Plan) ([]*codegen.File, error) {
	designRoots := serviceRoots(plan.Generation().Roots())
	plans := make([]*service.Plan, len(designRoots))
	for index, root := range designRoots {
		plans[index] = plan.Service(root)
	}
	return service.Files(plans...)
}

// planServiceData declares service-owned generated package types for every Goa
// design root in generation.
func planServiceData(plan *Plan) error {
	if plan.services != nil {
		return nil
	}
	plan.services = make(map[*expr.RootExpr]*service.Plan)
	roots := serviceRoots(plan.Generation().Roots())
	inputs := make([]service.PlanInput, len(roots))
	for index, root := range roots {
		inputs[index] = service.PlanInput{Root: root, Examples: plan.exampleGenerator(root)}
	}
	servicePlans, err := service.NewPlans(plan.Generation(), inputs...)
	if err != nil {
		return err
	}
	for index, root := range roots {
		plan.services[root] = servicePlans[index]
	}
	return nil
}

// serviceRoots returns every Goa design root that emits files into the same
// generated package tree.
func serviceRoots(roots []eval.Root) []*expr.RootExpr {
	var designRoots []*expr.RootExpr
	for _, root := range roots {
		if design, ok := root.(*expr.RootExpr); ok {
			designRoots = append(designRoots, design)
		}
	}
	return designRoots
}
