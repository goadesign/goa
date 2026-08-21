// This file defines the run-private plan shared by core generators and plugins.
// Task-specific retained analyses are added as typed private fields by the
// subsystem tasks that consume this lifecycle foundation.
package generator

import "goa.design/goa/v3/codegen"

type (
	// Plan is the immutable generation context passed from planning to rendering.
	// Its fields are private so it cannot become a generic analysis registry.
	Plan struct {
		generation *codegen.Generation
	}
)

// Generation returns the declaration and import catalog for this run.
func (p *Plan) Generation() *codegen.Generation {
	return p.generation
}
