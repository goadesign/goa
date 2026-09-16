// Package pluginregistry stores callbacks registered through Goa's released
// plugin API until generation starts. Plugin authors may register callbacks,
// but only the generator may stop registration or read the stored list.
package pluginregistry

import (
	"fmt"
	"slices"
	"sync"
)

type (
	// Position identifies when a plugin runs relative to normal plugins.
	Position uint8

	// Registry stores callbacks until the first generation run copies them.
	// Callback types stay paired with the package that defines them.
	Registry struct {
		mu            sync.Mutex
		registrations []storedRegistration
		sealed        bool
	}

	// Registration is one plugin definition with its original callback types.
	Registration[Prepare, Generate any] struct {
		Name     string
		Command  string
		Position Position
		Prepare  Prepare
		Generate Generate
	}

	// storedRegistration keeps callbacks without importing their defining
	// package. Snapshot restores the types supplied by that same package.
	storedRegistration struct {
		name     string
		command  string
		position Position
		prepare  any
		generate any
	}
)

const (
	// First places a plugin before normally ordered plugins.
	First Position = iota
	// Normal places a plugin between first and last plugins.
	Normal
	// Last places a plugin after normally ordered plugins.
	Last
)

var defaultRegistry = New()

// New creates an open plugin registry for Goa or a focused test.
func New() *Registry {
	return &Registry{}
}

// Register records one plugin in the process-wide registry used by Goa.
func Register[Prepare, Generate any](name, command string, position Position, prepare Prepare, generate Generate) {
	RegisterIn(defaultRegistry, name, command, position, prepare, generate)
}

// RegisterIn records one plugin in registry before generation starts.
func RegisterIn[Prepare, Generate any](registry *Registry, name, command string, position Position, prepare Prepare, generate Generate) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.sealed {
		panic("plugin registry is sealed")
	}
	registry.registrations = append(registry.registrations, storedRegistration{
		name:     name,
		command:  command,
		position: position,
		prepare:  prepare,
		generate: generate,
	})
}

// Snapshot stops further process-wide registrations and returns a copy of the
// registered plugins with their original callback types.
func Snapshot[Prepare, Generate any]() []Registration[Prepare, Generate] {
	return SnapshotFrom[Prepare, Generate](defaultRegistry)
}

// SnapshotFrom stops further registrations in registry and returns a copy that
// callers may sort without changing the stored order.
func SnapshotFrom[Prepare, Generate any](registry *Registry) []Registration[Prepare, Generate] {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.sealed = true
	stored := slices.Clone(registry.registrations)
	registrations := make([]Registration[Prepare, Generate], len(stored))
	for index, plugin := range stored {
		prepare, ok := plugin.prepare.(Prepare)
		if !ok {
			panic(fmt.Sprintf("plugin %q has an unexpected prepare callback type", plugin.name))
		}
		generate, ok := plugin.generate.(Generate)
		if !ok {
			panic(fmt.Sprintf("plugin %q has an unexpected generate callback type", plugin.name))
		}
		registrations[index] = Registration[Prepare, Generate]{
			Name:     plugin.name,
			Command:  plugin.command,
			Position: plugin.position,
			Prepare:  prepare,
			Generate: generate,
		}
	}
	return registrations
}
