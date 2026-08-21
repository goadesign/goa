// This file verifies that the generator plans every declaration before any
// core generator or plugin renders files from the frozen generation catalog.
package generator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestGeneratePhasesShareOneGeneration(t *testing.T) {
	command := fmt.Sprintf("test-generation-phases-%p", t)
	root := codegen.RunDSL(t, func() {})

	var (
		events        []string
		planned       *codegen.Generation
		lateDeclare   error
		preparedRoots []eval.Root
	)
	typesPath := "generated.local/gen/types"
	union := &expr.Union{TypeName: "Value", TypeKey: "type", ValueKey: "value"}
	lateUnion := &expr.Union{TypeName: "Late", TypeKey: "kind", ValueKey: "data"}

	assertGeneration := func(generation *codegen.Generation) error {
		if planned != generation {
			return fmt.Errorf("generation changed between plan and render")
		}
		if len(generation.Roots()) != len(preparedRoots) {
			return fmt.Errorf("generation roots changed after plugin preparation")
		}
		return nil
	}
	registry := newRegistry()
	registry.addCommand(command,
		func() coreGenerator {
			return coreGenerator{
				Plan: func(plan *Plan) error {
					events = append(events, "core-plan-first")
					planned = plan.Generation()
					typesPath = planned.GenPkg() + "/types"
					_, err := planned.GeneratedPackage(typesPath).DeclareUnion(union)
					return err
				},
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					events = append(events, "core-render-first")
					generation := plan.Generation()
					if err := assertGeneration(generation); err != nil {
						return nil, err
					}
					declaration, err := generation.GeneratedPackage(typesPath).Union(union)
					if err != nil {
						return nil, err
					}
					if declaration.Name() == "" {
						return nil, fmt.Errorf("union name is empty during render")
					}
					_, lateDeclare = generation.GeneratedPackage(typesPath).DeclareUnion(lateUnion)
					if lateDeclare == nil {
						return nil, fmt.Errorf("render declared a new union after freeze")
					}
					return nil, nil
				},
			}
		},
		func() coreGenerator {
			return coreGenerator{
				Plan: func(plan *Plan) error {
					events = append(events, "core-plan-second")
					return assertGeneration(plan.Generation())
				},
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					events = append(events, "core-render-second")
					return nil, assertGeneration(plan.Generation())
				},
			}
		},
	)
	registry.registerPlugin(
		"lifecycle",
		command,
		pluginNormal,
		func() Plugin {
			return Plugin{
				Prepare: func(_ string, roots []eval.Root) error {
					events = append(events, "plugin-prepare")
					preparedRoots = roots
					return nil
				},
				Plan: func(plan *Plan) error {
					events = append(events, "plugin-plan")
					return assertGeneration(plan.Generation())
				},
				Generate: func(plan *Plan, files []*codegen.File) ([]*codegen.File, error) {
					events = append(events, "plugin-render")
					return files, assertGeneration(plan.Generation())
				},
			}
		},
	)

	_, err := executeGeneration("generated.local/gen", []eval.Root{root}, command, registry)
	require.NoError(t, err)
	require.ErrorContains(t, lateDeclare, "frozen")
	require.Equal(t, []string{
		"plugin-prepare",
		"core-plan-first",
		"core-plan-second",
		"plugin-plan",
		"core-render-first",
		"core-render-second",
		"plugin-render",
	}, events)
}
