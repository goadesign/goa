// This file executes the prepare, plan, freeze, and render phases for explicit
// roots. The public filesystem-facing generator and isolated tests both use
// this path, so lifecycle behavior has one implementation.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// generationRun owns every fresh core and plugin instance for one execution.
	generationRun struct {
		cores   []coreGenerator
		plugins []Plugin
	}
)

// executeGeneration instantiates fresh core and plugin objects, prepares roots,
// and renders files from one retained frozen plan.
func executeGeneration(genpkg string, roots []eval.Root, command string, registry *registry) ([]*codegen.File, error) {
	run, err := newGenerationRun(command, registry)
	if err != nil {
		return nil, err
	}
	return run.execute(genpkg, roots)
}

// newGenerationRun snapshots immutable factories and invokes each exactly once.
func newGenerationRun(command string, registry *registry) (*generationRun, error) {
	coreFactories, pluginDescriptors, err := registry.snapshot(command)
	if err != nil {
		return nil, err
	}
	cores := make([]coreGenerator, len(coreFactories))
	for i, factory := range coreFactories {
		cores[i] = factory()
	}
	plugins := make([]Plugin, len(pluginDescriptors))
	for i, descriptor := range pluginDescriptors {
		plugins[i] = descriptor.factory()
	}
	return &generationRun{cores: cores, plugins: plugins}, nil
}

// execute runs all phases for explicit prepared-root inputs.
func (r *generationRun) execute(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	for _, plugin := range r.plugins {
		if plugin.Prepare != nil {
			if err := plugin.Prepare(genpkg, roots); err != nil {
				return nil, err
			}
		}
	}
	for _, root := range roots {
		if design, ok := root.(*expr.RootExpr); ok {
			codegen.NormalizeRoot(design)
		}
	}

	plan := &Plan{generation: codegen.NewGeneration(genpkg, roots)}
	for _, core := range r.cores {
		if core.Plan != nil {
			if err := core.Plan(plan); err != nil {
				return nil, err
			}
		}
	}
	for _, plugin := range r.plugins {
		if plugin.Plan != nil {
			if err := plugin.Plan(plan); err != nil {
				return nil, err
			}
		}
	}
	if err := plan.Generation().Freeze(); err != nil {
		return nil, err
	}

	var files []*codegen.File
	for _, core := range r.cores {
		if core.Generate == nil {
			continue
		}
		generated, err := core.Generate(plan)
		if err != nil {
			return nil, err
		}
		files = append(files, generated...)
	}
	for _, plugin := range r.plugins {
		if plugin.Generate == nil {
			continue
		}
		generated, err := plugin.Generate(plan, files)
		if err != nil {
			return nil, err
		}
		files = generated
	}
	return files, nil
}
