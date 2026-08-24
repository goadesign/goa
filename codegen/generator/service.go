// This file assembles generated service files after every participating Goa
// design root has submitted its declarations and all package names are final.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// serviceFiles returns the service files described by plan's completed package
// declarations and the example generator created for this run.
func serviceFiles(plan *Plan) ([]*codegen.File, error) {
	return service.Files(plan.serviceOrder...)
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
	plan.serviceOrder = servicePlans
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
