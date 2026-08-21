// Generate asks this file for the core callbacks selected by the gen or example
// command. It receives Genfunc records whose Plan callbacks all run before the
// same frozen Generation is passed to their file-producing Generate callbacks.
package generator

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

type (
	// Genfunc plans declarations and renders files for one generation run.
	Genfunc struct {
		// Plan declares generated package types before any generator renders files.
		Plan codegen.PlanFunc
		// Generate renders files from the frozen generation catalog.
		Generate func(*codegen.Generation) ([]*codegen.File, error)
	}
)

// Generators returns the generation lifecycle callbacks for the given command,
// or an error if the command is not supported. Generators is a public variable
// so external code may replace the default generators.
var Generators = generators

// generators returns the generator functions exposed by the generator package
// for the given command.
func generators(cmd string) ([]Genfunc, error) {
	switch cmd {
	case "gen":
		return []Genfunc{
			{Plan: planServiceData, Generate: Service},
			{Plan: planServiceData, Generate: Transport},
			{Plan: planServiceData, Generate: OpenAPI},
		}, nil
	case "example":
		return []Genfunc{{Plan: planServiceData, Generate: Example}}, nil
	default:
		return nil, fmt.Errorf("unknown command %q", cmd)
	}
}

// renderOnly adapts a generator that does not yet plan package declarations to
// render with the generation context selected by the top-level lifecycle.
func renderOnly(generate func(string, []eval.Root) ([]*codegen.File, error)) Genfunc {
	return Genfunc{
		Generate: func(generation *codegen.Generation) ([]*codegen.File, error) {
			return generate(generation.GenPkg, generation.Roots)
		},
	}
}
