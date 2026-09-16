// This file checks that every package name is chosen before core generators or
// plugins write files.
package generator

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpdata "goa.design/goa/v3/http/codegen/testdata"
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
	union := &expr.AttributeExpr{Type: &expr.Union{TypeName: "Value", TypeKey: "type", ValueKey: "value"}}
	lateUnion := &expr.AttributeExpr{Type: &expr.Union{TypeName: "Late", TypeKey: "kind", ValueKey: "data"}}

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
					types, err := planned.ClaimPackage(typesPath)
					if err != nil {
						return err
					}
					_, err = types.DeclareUnion(union)
					return err
				},
				Generate: func(plan *Plan) ([]*codegen.File, error) {
					events = append(events, "core-render-first")
					generation := plan.Generation()
					if err := assertGeneration(generation); err != nil {
						return nil, err
					}
					declaration, err := generation.Package(typesPath).Union(union)
					if err != nil {
						return nil, err
					}
					if declaration.Name() == "" {
						return nil, fmt.Errorf("union name is empty during render")
					}
					_, lateDeclare = generation.Package(typesPath).DeclareUnion(lateUnion)
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

	err := executeGeneration("generated.local/gen", []eval.Root{root}, command, registry)
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

func TestCommandsPlanOnlyTheirFiles(t *testing.T) {
	root := codegen.RunDSL(t, httpdata.AliasTypeDSL)

	genRun, err := newGenerationRun("gen", newDefaultRegistry())
	require.NoError(t, err)
	genResult, err := genRun.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	require.Nil(t, genResult.plan.example)
	require.NotNil(t, genResult.plan.openapi)

	exampleRun, err := newGenerationRun("example", newDefaultRegistry())
	require.NoError(t, err)
	exampleResult, err := exampleRun.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	require.NotNil(t, exampleResult.plan.example)
	require.Nil(t, exampleResult.plan.openapi)
}

// TestRenderUsesRetainedPlans proves that file rendering does not look up
// services or transports from the prepared design after planning finishes.
func TestRenderUsesRetainedPlans(t *testing.T) {
	root := codegen.RunDSL(t, httpdata.AliasTypeDSL)
	plan := mustTestPlan(
		t,
		"generated.local/gen",
		[]eval.Root{root},
		planServiceData,
		planTransportData,
	)

	plan.preparedRoots = nil
	plan.services = nil
	plan.http = nil
	plan.jsonrpcHTTP = nil
	plan.jsonrpc = nil
	plan.grpc = nil

	serviceFiles, err := serviceFiles(plan)
	require.NoError(t, err)
	require.NotEmpty(t, serviceFiles)
	transportFiles, err := transportFiles(plan)
	require.NoError(t, err)
	require.NotEmpty(t, transportFiles)
}

// TestPreparedRootsRejectFileRenderMutation proves that persistent mutations
// made by templates and file finalizers are rejected after rendering completes.
func TestPreparedRootsRejectFileRenderMutation(t *testing.T) {
	for _, phase := range []string{"template", "finalizer"} {
		t.Run(phase, func(t *testing.T) {
			root := codegen.RunDSL(t, httpdata.AliasTypeDSL)
			dir := t.TempDir()
			writeGeneratedModule(t, filepath.Join(dir, codegen.Gendir), "generated.local/gen")
			mutate := func() {
				root.API.HTTP.Services[0].HTTPEndpoints[0].Routes[0].Path = "/changed"
			}
			first := &codegen.File{
				Path: "first.txt",
				SectionTemplates: []*codegen.SectionTemplate{{
					Name:   "first",
					Source: "first",
				}},
			}
			if phase == "template" {
				first.SectionTemplates[0].Source = "{{ mutate }}"
				first.SectionTemplates[0].FuncMap = map[string]any{"mutate": func() string {
					mutate()
					return "first"
				}}
			} else {
				first.FinalizeFunc = func(_ string) error {
					mutate()
					return nil
				}
			}
			registry := testRegistry("test", func() coreGenerator {
				return coreGenerator{name: "files", Generate: func(_ *Plan) ([]*codegen.File, error) {
					return []*codegen.File{first}, nil
				}}
			})

			_, err := generate(dir, "test", false, registry)
			require.ErrorContains(t, err, "generated file renders mutated prepared design")
		})
	}
}
