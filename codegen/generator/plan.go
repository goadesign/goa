// This file defines the run-private plan shared by core generators and plugins.
// Task-specific retained analyses are added as typed private fields by the
// subsystem tasks that consume this lifecycle foundation.
package generator

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// Plan retains the typed state shared by planning and rendering in one run.
	// Planning may add declarations to Generation; rendering receives the same
	// plan only after those declarations are frozen.
	Plan struct {
		generation    *codegen.Generation
		preparedRoots []eval.Root
		examples      map[*expr.RootExpr]*expr.ExampleGenerator
		design        *designSnapshot
	}
)

// Generation returns the declaration and import catalog for this run.
func (p *Plan) Generation() *codegen.Generation {
	return p.generation
}

// exampleGenerator returns the mutable example state created for root in this
// run. A root outside the prepared plan is an orchestration bug.
func (p *Plan) exampleGenerator(root *expr.RootExpr) *expr.ExampleGenerator {
	generator, ok := p.examples[root]
	if !ok {
		panic(fmt.Sprintf("example generator requested for unplanned design root %q", root.API.Name))
	}
	return generator
}

// verifyPreparedDesign rejects the first expression change made after
// preparation and identifies the callback or render operation that made it.
func (p *Plan) verifyPreparedDesign(operation string) error {
	path, err := p.design.changedPath(p.preparedRoots)
	if err != nil {
		return fmt.Errorf("%s left prepared design unverifiable: %w", operation, err)
	}
	if path != "" {
		return fmt.Errorf("%s mutated prepared design at %s", operation, path)
	}
	return nil
}
