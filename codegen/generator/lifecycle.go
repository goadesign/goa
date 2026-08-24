// This file runs plugin preparation, name selection, file planning, and file
// writing in one order shared by production generation and tests.
package generator

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

type (
	// generationRun stores the new core generators and plugins used by one run.
	generationRun struct {
		cores   []coreGenerator
		plugins []runPlugin
	}

	// runPlugin stores one plugin's registered name and its Prepare, Plan, and
	// Generate functions for this run.
	runPlugin struct {
		name string
		Plugin
	}

	// generationResult stores the generation state and files produced by one
	// run.
	generationResult struct {
		plan  *Plan
		files []*codegen.File
	}
)

// executeGeneration creates new core generators and plugins, prepares the
// designs, chooses all names, and reports whether generation succeeded.
func executeGeneration(genpkg string, roots []eval.Root, command string, registry *registry) error {
	run, err := newGenerationRun(command, registry)
	if err != nil {
		return err
	}
	_, err = run.execute(genpkg, roots)
	return err
}

// newGenerationRun copies the registered factories and calls each one once.
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

// execute prepares the supplied designs, chooses names, and builds files.
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
