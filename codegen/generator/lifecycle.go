// This file executes the prepare, plan, freeze, and render phases for explicit
// roots. The public filesystem-facing generator and isolated tests both use
// this path, so lifecycle behavior has one implementation.
package generator

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

type (
	// generationRun owns every fresh core and plugin instance for one execution.
	generationRun struct {
		cores   []coreGenerator
		plugins []runPlugin
	}

	// runPlugin retains one plugin's registered owner name with its fresh callbacks.
	runPlugin struct {
		name string
		Plugin
	}

	// generationResult retains the exact plan needed to verify later file renders.
	generationResult struct {
		plan  *Plan
		files []*codegen.File
	}
)

// executeGeneration instantiates fresh core and plugin objects, prepares roots,
// and produces file descriptions from one retained frozen plan.
func executeGeneration(genpkg string, roots []eval.Root, command string, registry *registry) ([]*codegen.File, error) {
	run, err := newGenerationRun(command, registry)
	if err != nil {
		return nil, err
	}
	result, err := run.execute(genpkg, roots)
	if err != nil {
		return nil, err
	}
	return result.files, nil
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
	plugins := make([]runPlugin, len(pluginDescriptors))
	for i, descriptor := range pluginDescriptors {
		plugins[i] = runPlugin{name: descriptor.name, Plugin: descriptor.factory()}
	}
	return &generationRun{cores: cores, plugins: plugins}, nil
}

// execute runs all phases for explicit prepared-root inputs.
func (r *generationRun) execute(genpkg string, roots []eval.Root) (*generationResult, error) {
	for _, plugin := range r.plugins {
		if plugin.Prepare != nil {
			if err := plugin.Prepare(genpkg, roots); err != nil {
				return nil, err
			}
		}
	}
	generation, err := codegen.NewGeneration(genpkg, roots)
	if err != nil {
		return nil, err
	}
	design, err := snapshotPreparedDesign(roots)
	if err != nil {
		return nil, err
	}
	plan := &Plan{
		generation:    generation,
		preparedRoots: roots,
		examples:      newExampleGenerators(roots),
		design:        design,
	}
	if err := plan.verifyPreparedDesign("example generator creation"); err != nil {
		return nil, err
	}
	for _, core := range r.cores {
		if core.Plan != nil {
			callbackErr := core.Plan(plan)
			if err := plan.verifyPreparedDesign(fmt.Sprintf("core %q plan", core.name)); err != nil {
				return nil, err
			}
			if callbackErr != nil {
				return nil, callbackErr
			}
		}
	}
	for _, plugin := range r.plugins {
		if plugin.Plan != nil {
			callbackErr := plugin.Plan(plan)
			if err := plan.verifyPreparedDesign(fmt.Sprintf("plugin %q plan", plugin.name)); err != nil {
				return nil, err
			}
			if callbackErr != nil {
				return nil, callbackErr
			}
		}
	}
	freezeErr := plan.Generation().Freeze()
	if err := plan.verifyPreparedDesign("generation freeze"); err != nil {
		return nil, err
	}
	if freezeErr != nil {
		return nil, freezeErr
	}
	linkErr := plan.link()
	if err := plan.verifyPreparedDesign("plan linking"); err != nil {
		return nil, err
	}
	if linkErr != nil {
		return nil, linkErr
	}

	var files []*codegen.File
	for _, core := range r.cores {
		if core.Generate == nil {
			continue
		}
		generated, callbackErr := core.Generate(plan)
		if err := plan.verifyPreparedDesign(fmt.Sprintf("core %q generate", core.name)); err != nil {
			return nil, err
		}
		if callbackErr != nil {
			return nil, callbackErr
		}
		files = append(files, generated...)
	}
	for _, plugin := range r.plugins {
		if plugin.Generate == nil {
			continue
		}
		generated, callbackErr := plugin.Generate(plan, files)
		if err := plan.verifyPreparedDesign(fmt.Sprintf("plugin %q generate", plugin.name)); err != nil {
			return nil, err
		}
		if callbackErr != nil {
			return nil, callbackErr
		}
		files = generated
	}
	return &generationResult{plan: plan, files: files}, nil
}
