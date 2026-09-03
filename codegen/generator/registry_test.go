// This file supplies isolated command registries to generator integration
// tests. Test factories adapt the transitional Generation-based core renderers
// without restoring mutable production hooks.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

type (
	// testGenfunc describes one retained planner and renderer used by an
	// isolated generator command.
	testGenfunc struct {
		// Plan declares package symbols and retains the analysis used by Generate.
		Plan func(*Plan) error
		// Generate renders fixture files from the linked plan.
		Generate func(*Plan) ([]*codegen.File, error)
	}
)

// testRegistry returns an isolated registry for one command.
func testRegistry(command string, factories ...generatorFactory) *registry {
	registry := newRegistry()
	registry.addCommand(command, factories...)
	return registry
}

// testRegistryFromGenfuncs creates one isolated command from fixture callbacks.
func testRegistryFromGenfuncs(callbacks []testGenfunc) *registry {
	factories := make([]generatorFactory, len(callbacks))
	for i, callback := range callbacks {
		factories[i] = testGenerator(callback.Plan, callback.Generate)
	}
	return testRegistry("gen", factories...)
}

// testRenderOnly adapts a root-based rendering fixture into a test callback.
func testRenderOnly(generate func(string, []eval.Root) ([]*codegen.File, error)) testGenfunc {
	return testGenfunc{Generate: func(plan *Plan) ([]*codegen.File, error) {
		generation := plan.Generation()
		return generate(generation.GenPkg(), generation.Roots())
	}}
}

// testGenerator returns a fresh core generator that receives one retained plan
// from declaration collection through rendering.
func testGenerator(plan func(*Plan) error, generate func(*Plan) ([]*codegen.File, error)) generatorFactory {
	return func() coreGenerator {
		return coreGenerator{
			name:     "test",
			Plan:     plan,
			Generate: generate,
		}
	}
}
