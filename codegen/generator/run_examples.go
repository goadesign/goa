// This file creates the mutable example state owned by one generation plan.
// Evaluated API roots retain only immutable factories, so repeated and
// concurrent runs never share consumed streams or recursion caches.
package generator

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// newExampleGenerators creates one fresh mutable generator for every Goa
// design root participating in a run.
func newExampleGenerators(roots []eval.Root) map[*expr.RootExpr]*expr.ExampleGenerator {
	generators := make(map[*expr.RootExpr]*expr.ExampleGenerator)
	for _, root := range roots {
		if design, ok := root.(*expr.RootExpr); ok {
			generators[design] = expr.NewExampleGenerator(design.API.RandomizerFactory)
		}
	}
	return generators
}
