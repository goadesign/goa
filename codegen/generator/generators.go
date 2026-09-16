// This file lists the generators used by each command. Every command run gets
// new functions, and both functions receive the same Plan.
package generator

import (
	"fmt"
	"reflect"

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

// Generators returns the generator functions for command. Plugins may replace
// it to add, remove, or reorder generators.
var Generators = generators

// generators returns Goa's built-in generator functions for command.
func generators(command string) ([]Genfunc, error) {
	switch command {
	case "gen":
		return []Genfunc{Service, Transport, OpenAPI}, nil
	case "example":
		return []Genfunc{Example}, nil
	default:
		return nil, fmt.Errorf("unknown command %q", command)
	}
}

// generatorFactories returns fresh shared-plan adapters for the released
// generator list. Goa's built-in functions plan together. Additional plugin
// functions receive the prepared roots after every generated name is fixed.
func generatorFactories(command string) ([]generatorFactory, error) {
	generate, err := Generators(command)
	if err != nil {
		return nil, err
	}
	factories := make([]generatorFactory, len(generate))
	for index, generator := range generate {
		if generator == nil {
			return nil, fmt.Errorf("generator %d for command %q is nil", index, command)
		}
		factories[index] = generatorFactoryFor(generator)
	}
	return factories, nil
}

// generatorFactoryFor keeps built-in generators in the shared planning pass
// and adapts other released generator functions to the final rendering pass.
func generatorFactoryFor(generate Genfunc) generatorFactory {
	pointer := reflect.ValueOf(generate).Pointer()
	switch pointer {
	case reflect.ValueOf(Service).Pointer():
		return serviceGeneratorFactory
	case reflect.ValueOf(Transport).Pointer():
		return transportGeneratorFactory
	case reflect.ValueOf(OpenAPI).Pointer():
		return openAPIGeneratorFactory
	case reflect.ValueOf(Example).Pointer():
		return exampleGeneratorFactory
	default:
		return func() coreGenerator {
			return coreGenerator{
				name: "external generator",
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					generation := plan.Generation()
					return generate(generation.GenPkg(), generation.Roots())
				},
			}
		}
	}
}

// runStandaloneGenerator runs one released built-in function with a complete
// plan. It does not run registered plugins because callers invoked the core
// generator directly.
func runStandaloneGenerator(genpkg string, roots []eval.Root, factory generatorFactory) ([]*codegen.File, error) {
	run := generationRun{cores: []coreGenerator{factory()}}
	result, err := run.execute(genpkg, roots)
	if err != nil {
		return nil, err
	}
	return result.files, nil
}

// genGeneratorFactories returns the service, transport, and OpenAPI generators
// used by the gen command.
func genGeneratorFactories() []generatorFactory {
	return []generatorFactory{
		serviceGeneratorFactory,
		transportGeneratorFactory,
		openAPIGeneratorFactory,
	}
}

// exampleGeneratorFactories returns the generator used by the example command.
func exampleGeneratorFactories() []generatorFactory {
	return []generatorFactory{exampleGeneratorFactory}
}

// serviceGeneratorFactory returns a fresh service generator for one run.
func serviceGeneratorFactory() coreGenerator {
	return coreGenerator{
		name:     "service",
		Plan:     planServiceData,
		Generate: serviceFiles,
	}
}

// transportGeneratorFactory returns a fresh transport generator for one run.
func transportGeneratorFactory() coreGenerator {
	return coreGenerator{
		name:     "transport",
		Plan:     planTransportData,
		Generate: transportFiles,
	}
}

// openAPIGeneratorFactory returns a fresh OpenAPI generator for one run.
func openAPIGeneratorFactory() coreGenerator {
	return coreGenerator{
		name:     "openapi",
		Plan:     planOpenAPIData,
		Generate: openAPIFiles,
	}
}

// exampleGeneratorFactory returns a fresh example generator for one run.
func exampleGeneratorFactory() coreGenerator {
	return coreGenerator{
		name:     "example",
		Plan:     planExampleData,
		Generate: exampleFiles,
	}
}
