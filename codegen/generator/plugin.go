// This file stores the generator functions registered for each command and
// creates a separate plugin value for each run. Registration closes when
// generation first starts.
package generator

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/internal/pluginregistry"
)

type (
	// Plugin contains the optional functions run for one new plugin instance.
	// Plan and Generate receive the same Plan pointer.
	Plugin struct {
		// Prepare may change designs before Goa records their prepared values.
		Prepare codegen.PrepareFunc
		// Plan adds the plugin's package-level names before all names are final.
		Plan func(*Plan) error
		// Generate adds or changes files after all names are final.
		Generate func(*Plan, []*codegen.File) ([]*codegen.File, error)
	}

	// PluginFactory creates one independent plugin instance for each run.
	PluginFactory func() Plugin

	// registry stores the core and plugin factories used by each command.
	registry struct {
		mu                sync.Mutex
		commands          map[string][]generatorFactory
		commandGenerators func(string) ([]generatorFactory, error)
		plugins           []pluginDescriptor
		registeredPlugins func() []registeredPluginDescriptor
		sealed            bool
	}

	// pluginDescriptor stores one plugin registration.
	pluginDescriptor struct {
		name     string
		command  string
		position pluginPosition
		factory  PluginFactory
	}

	// registeredPluginDescriptor copies one plugin registered through the
	// released Goa v3 API before adapting it to a per-run plugin.
	registeredPluginDescriptor struct {
		name     string
		command  string
		position pluginPosition
		prepare  codegen.PrepareFunc
		generate codegen.GenerateFunc
	}

	// pluginPosition defines the three plugin ordering groups.
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

// newRegistry creates an empty list of commands and plugins for setup or tests.
func newRegistry() *registry {
	return &registry{commands: make(map[string][]generatorFactory)}
}

// newDefaultRegistry creates the production command registry before external
// package initialization registers plugins.
func newDefaultRegistry() *registry {
	registry := newRegistry()
	registry.commands["gen"] = genGeneratorFactories()
	registry.commands["example"] = exampleGeneratorFactories()
	registry.commandGenerators = generatorFactories
	registry.registeredPlugins = snapshotRegisteredPlugins
	return registry
}

// addCommand adds core generators to a command used by a test.
func (r *registry) addCommand(command string, factories ...generatorFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		panic("generator registry is sealed")
	}
	r.commands[command] = slices.Clone(factories)
}

// registerPlugin adds one named factory to a known command before generation
// starts. A command cannot contain two plugins with the same name.
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

// snapshot closes registration and returns copied factories in a repeatable order.
func (r *registry) snapshot(command string) ([]generatorFactory, []pluginDescriptor, error) {
	var (
		selectedFactories []generatorFactory
		hasSelection      bool
	)
	if r.commandGenerators != nil {
		var err error
		selectedFactories, err = r.commandGenerators(command)
		if err != nil {
			return nil, nil, err
		}
		hasSelection = true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	factories, ok := r.commands[command]
	if hasSelection {
		factories = selectedFactories
		ok = true
	}
	if !ok {
		return nil, nil, fmt.Errorf("unknown command %q", command)
	}
	r.sealed = true
	selected := make([]pluginDescriptor, 0, len(r.plugins))
	factoryNames := make(map[string]struct{}, len(r.plugins))
	for _, plugin := range r.plugins {
		if plugin.command != command {
			continue
		}
		selected = append(selected, plugin)
		factoryNames[plugin.name] = struct{}{}
	}
	if r.registeredPlugins != nil {
		for _, registered := range r.registeredPlugins() {
			if registered.command != command {
				continue
			}
			if _, ok := factoryNames[registered.name]; ok {
				return nil, nil, fmt.Errorf("plugin %q is already registered for command %q", registered.name, command)
			}
			selected = append(selected, registered.pluginDescriptor())
		}
	}
	slices.SortStableFunc(selected, func(left, right pluginDescriptor) int {
		if left.position != right.position {
			return int(left.position) - int(right.position)
		}
		return strings.Compare(left.name, right.name)
	})
	return slices.Clone(factories), selected, nil
}

// snapshotRegisteredPlugins stops further callback registration and copies
// each registered callback into the list used by this generation run.
func snapshotRegisteredPlugins() []registeredPluginDescriptor {
	plugins := pluginregistry.Snapshot[codegen.PrepareFunc, codegen.GenerateFunc]()
	descriptors := make([]registeredPluginDescriptor, len(plugins))
	for index, plugin := range plugins {
		position := pluginNormal
		if plugin.Position == pluginregistry.First {
			position = pluginFirst
		} else if plugin.Position == pluginregistry.Last {
			position = pluginLast
		}
		descriptors[index] = registeredPluginDescriptor{
			name:     plugin.Name,
			command:  plugin.Command,
			position: position,
			prepare:  plugin.Prepare,
			generate: plugin.Generate,
		}
	}
	return descriptors
}

// pluginDescriptor adapts a released callback pair to the same factory and
// per-run Plan used by newer plugins.
func (p registeredPluginDescriptor) pluginDescriptor() pluginDescriptor {
	return pluginDescriptor{
		name:     p.name,
		command:  p.command,
		position: p.position,
		factory: func() Plugin {
			return Plugin{
				Prepare: p.prepare,
				Generate: func(plan *Plan, files []*codegen.File) ([]*codegen.File, error) {
					return p.generate(plan.Generation().GenPkg(), plan.preparedRoots, files)
				},
			}
		},
	}
}
