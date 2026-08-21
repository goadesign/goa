// This file defines the fresh core generator objects selected by each command.
// Factories are immutable; every run receives new callback values and retains
// one Plan from declaration planning through rendering.
package generator

import "goa.design/goa/v3/codegen"

type (
	// coreGenerator plans and renders one core subsystem for a single run.
	coreGenerator struct {
		// name identifies the subsystem in lifecycle diagnostics.
		name string
		// Plan declares package symbols and retains run-specific analysis.
		Plan func(*Plan) error
		// Generate renders files from the same frozen plan.
		Generate func(*Plan) ([]*codegen.File, error)
	}

	// generatorFactory creates one core generator instance for a run.
	generatorFactory func() coreGenerator
)

// genGeneratorFactories returns fresh service, transport, and OpenAPI factories.
func genGeneratorFactories() []generatorFactory {
	return []generatorFactory{
		func() coreGenerator {
			return coreGenerator{
				name: "service",
				Plan: func(plan *Plan) error {
					return planServiceData(plan.Generation())
				},
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					return serviceFiles(plan)
				},
			}
		},
		func() coreGenerator {
			return coreGenerator{
				name: "transport",
				Plan: func(plan *Plan) error {
					return planTransportData(plan.Generation())
				},
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					return transportFiles(plan)
				},
			}
		},
		func() coreGenerator {
			return coreGenerator{
				name: "openapi",
				Plan: func(plan *Plan) error {
					return planServiceData(plan.Generation())
				},
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					return openAPIFiles(plan)
				},
			}
		},
	}
}

// exampleGeneratorFactories returns a fresh example generator factory.
func exampleGeneratorFactories() []generatorFactory {
	return []generatorFactory{
		func() coreGenerator {
			return coreGenerator{
				name: "example",
				Plan: func(plan *Plan) error {
					return planTransportData(plan.Generation())
				},
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					return exampleFiles(plan)
				},
			}
		},
	}
}
