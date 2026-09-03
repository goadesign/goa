// This file creates a separate example generator for each evaluated Goa root
// in one command. Repeated or concurrent commands therefore do not share the
// random value sequence or the record of types currently being visited.
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
