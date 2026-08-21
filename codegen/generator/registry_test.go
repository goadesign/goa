// This file supplies isolated command registries to generator integration
// tests. Test factories adapt the transitional Generation-based core renderers
// without restoring mutable production hooks.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

type (
	// testGenfunc retains the previous fixture shape while adapting callbacks
	// into fresh core generator objects.
	testGenfunc struct {
		// Plan declares package symbols through the fixture's Generation seam.
		Plan func(*codegen.Generation) error
		// Generate renders fixture files through the same Generation seam.
		Generate func(*codegen.Generation) ([]*codegen.File, error)
	}
)

// testRegistry returns an isolated registry for one command.
func testRegistry(command string, factories ...generatorFactory) *registry {
	registry := newRegistry()
	registry.addCommand(command, factories...)
	return registry
}

// testRegistryFromGenfuncs creates one isolated command from fixture callbacks.
func testRegistryFromGenfuncs(command string, callbacks []testGenfunc) *registry {
	factories := make([]generatorFactory, len(callbacks))
	for i, callback := range callbacks {
		factories[i] = testGenerator(callback.Plan, callback.Generate)
	}
	return testRegistry(command, factories...)
}

// testRenderOnly adapts a root-based rendering fixture into a test callback.
func testRenderOnly(generate func(string, []eval.Root) ([]*codegen.File, error)) testGenfunc {
	return testGenfunc{Generate: func(generation *codegen.Generation) ([]*codegen.File, error) {
		return generate(generation.GenPkg(), generation.Roots())
	}}
}

// testGenerator adapts the current Generation-based core callback functions to
// a fresh run factory. Retained subsystem plans replace this adapter in Tasks 7–10.
func testGenerator(plan func(*codegen.Generation) error, generate func(*codegen.Generation) ([]*codegen.File, error)) generatorFactory {
	return func() coreGenerator {
		generator := coreGenerator{}
		if plan != nil {
			generator.Plan = func(retained *Plan) error {
				return plan(retained.Generation())
			}
		}
		if generate != nil {
			generator.Generate = func(retained *Plan) ([]*codegen.File, error) {
				return generate(retained.Generation())
			}
		}
		return generator
	}
}

// testRenderGenerator adapts a legacy render-only test callback without adding
// a production lifecycle path.
func testRenderGenerator(generate func(string, []eval.Root) ([]*codegen.File, error)) generatorFactory {
	return testGenerator(nil, func(generation *codegen.Generation) ([]*codegen.File, error) {
		return generate(generation.GenPkg(), generation.Roots())
	})
}
