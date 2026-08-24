// This file lists the generators used by each command. Every command run gets
// new functions, and both functions receive the same Plan.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

type (
	// Genfunc is the released signature of a standalone generator function.
	// The current generator uses the run-wide Plan instead.
	//
	// Deprecated: Register a PluginFactory to add generated files.
	Genfunc func(genpkg string, roots []eval.Root) ([]*codegen.File, error)

	// coreGenerator chooses names and then builds one group of generated files.
	coreGenerator struct {
		// name identifies the file group in error messages.
		name string
		// Plan chooses generated names and saves the data needed to build files.
		Plan func(*Plan) error
		// Generate builds files from that same Plan after all names are final.
		Generate func(*Plan) ([]*codegen.File, error)
	}

	// generatorFactory returns a new pair of generator functions when called.
	generatorFactory func() coreGenerator
)

// genGeneratorFactories returns the service, transport, and OpenAPI generators
// used by the gen command.
func genGeneratorFactories() []generatorFactory {
	return []generatorFactory{
		func() coreGenerator {
			return coreGenerator{
				name:     "service",
				Plan:     planServiceData,
				Generate: serviceFiles,
			}
		},
		func() coreGenerator {
			return coreGenerator{
				name:     "transport",
				Plan:     planTransportData,
				Generate: transportFiles,
			}
		},
		func() coreGenerator {
			return coreGenerator{
				name:     "openapi",
				Plan:     planOpenAPIData,
				Generate: openAPIFiles,
			}
		},
	}
}

// exampleGeneratorFactories returns the generator used by the example command.
func exampleGeneratorFactories() []generatorFactory {
	return []generatorFactory{
		func() coreGenerator {
			return coreGenerator{
				name:     "example",
				Plan:     planExampleData,
				Generate: exampleFiles,
			}
		},
	}
}
