// This file owns immutable plugin factories and creates fresh callback objects
// for every generation run. The registry seals when its first run snapshots
// factories, preventing process history from changing later runs.
package generator

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

type (
	// PrepareFunc may amend evaluated roots before normalization and planning.
	// Preparation is the only plugin phase allowed to mutate design expressions.
	PrepareFunc func(genpkg string, roots []eval.Root) error

	// Plugin contains the optional callbacks run on one fresh plugin instance.
	// Plan and Generate receive the same retained Plan pointer.
	Plugin struct {
		// Prepare may amend roots before the Generation snapshot is created.
		Prepare PrepareFunc
		// Plan declares plugin-owned package symbols before generation freeze.
		Plan func(*Plan) error
		// Generate appends or transforms files using the frozen retained plan.
		Generate func(*Plan, []*codegen.File) ([]*codegen.File, error)
	}

	// PluginFactory creates one independent plugin instance for each run.
	PluginFactory func() Plugin

	// registry owns core and plugin factories used by one command namespace.
	registry struct {
		mu       sync.Mutex
		commands map[string][]generatorFactory
		plugins  []pluginDescriptor
		sealed   bool
	}

	// pluginDescriptor is immutable registration metadata retained globally.
	pluginDescriptor struct {
		name     string
		command  string
		position pluginPosition
		factory  PluginFactory
	}

	// pluginPosition defines the three stable registration groups.
	pluginPosition uint8
)

const (
	pluginFirst pluginPosition = iota
	pluginNormal
	pluginLast
)

var defaultRegistry = newDefaultRegistry()

// RegisterPlugin registers a factory in the normal alphabetically ordered
// group. It panics when name is empty, command is unknown, factory is nil, the
// command already has a plugin with name, or generation has already started.
func RegisterPlugin(name, command string, factory PluginFactory) {
	defaultRegistry.registerPlugin(name, command, pluginNormal, factory)
}

// RegisterPluginFirst registers a factory before normal and Last plugins. It
// enforces the same registration contract as RegisterPlugin.
func RegisterPluginFirst(name, command string, factory PluginFactory) {
	defaultRegistry.registerPlugin(name, command, pluginFirst, factory)
}

// RegisterPluginLast registers a factory after First and normal plugins. It
// enforces the same registration contract as RegisterPlugin.
func RegisterPluginLast(name, command string, factory PluginFactory) {
	defaultRegistry.registerPlugin(name, command, pluginLast, factory)
}

// newRegistry creates an empty mutable registry for init-time setup or tests.
func newRegistry() *registry {
	return &registry{commands: make(map[string][]generatorFactory)}
}

// newDefaultRegistry creates the production command registry before external
// package initialization registers plugins.
func newDefaultRegistry() *registry {
	registry := newRegistry()
	registry.commands["gen"] = genGeneratorFactories()
	registry.commands["example"] = exampleGeneratorFactories()
	return registry
}

// addCommand installs private core factories in an isolated test registry.
func (r *registry) addCommand(command string, factories ...generatorFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		panic("generator registry is sealed")
	}
	r.commands[command] = slices.Clone(factories)
}

// registerPlugin records one named factory for a known command before the
// first snapshot. Plugin names uniquely identify their owner within a command.
func (r *registry) registerPlugin(name, command string, position pluginPosition, factory PluginFactory) {
	if factory == nil {
		panic("plugin factory is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		panic("generator plugin registry is sealed")
	}
	if name == "" {
		panic("plugin name is empty")
	}
	if _, ok := r.commands[command]; !ok {
		panic(fmt.Sprintf("unknown generator command %q", command))
	}
	for _, plugin := range r.plugins {
		if plugin.command == command && plugin.name == name {
			panic(fmt.Sprintf("plugin %q is already registered for command %q", name, command))
		}
	}
	r.plugins = append(r.plugins, pluginDescriptor{
		name:     name,
		command:  command,
		position: position,
		factory:  factory,
	})
}

// snapshot seals the registry and returns copied factories in stable order.
func (r *registry) snapshot(command string) ([]generatorFactory, []pluginDescriptor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	factories, ok := r.commands[command]
	if !ok {
		return nil, nil, fmt.Errorf("unknown command %q", command)
	}
	r.sealed = true
	plugins := make([]pluginDescriptor, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		if plugin.command == command {
			plugins = append(plugins, plugin)
		}
	}
	slices.SortFunc(plugins, func(left, right pluginDescriptor) int {
		if left.position != right.position {
			return int(left.position) - int(right.position)
		}
		return strings.Compare(left.name, right.name)
	})
	return slices.Clone(factories), plugins, nil
}
