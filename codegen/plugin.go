// This file stores plugins registered through the released Goa v3 API. At the
// start of each generation command, the generator copies the registered
// functions. It calls First plugins before normal plugins and Last plugins
// afterward, orders names within each group, and keeps registration order when
// plugins in one group have the same name. This package stores those functions
// but does not call them.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/codegen/internal/pluginregistry"
	"goa.design/goa/v3/eval"
)

type (
	// GenerateFunc may add, remove, or change the files produced by Goa and by
	// plugins that ran earlier. It returns the complete file list for the next
	// plugin.
	GenerateFunc func(genpkg string, roots []eval.Root, files []*File) ([]*File, error)

	// PrepareFunc may change evaluated designs before Goa chooses generated Go
	// names. A nil PrepareFunc means that the plugin does not prepare designs.
	PrepareFunc func(genpkg string, roots []eval.Root) error

	// pluginPosition identifies the three ordering groups supported by the
	// released registration API.
	pluginPosition = pluginregistry.Position
)

const (
	pluginFirst  = pluginregistry.First
	pluginNormal = pluginregistry.Normal
	pluginLast   = pluginregistry.Last
)

// RegisterPlugin adds a plugin to the normal alphabetically ordered group. It
// panics for an empty name, an unknown command, a nil generation function, or
// registration after generation has started. Repeated names remain allowed for
// compatibility with released Goa plugins.
func RegisterPlugin(name, command string, prepare PrepareFunc, generate GenerateFunc) {
	registerPlugin(name, command, pluginNormal, prepare, generate)
}

// RegisterPluginFirst adds a plugin before normal and last plugins. Plugins in
// this group run by name. Plugins with the same name run in registration order.
func RegisterPluginFirst(name, command string, prepare PrepareFunc, generate GenerateFunc) {
	registerPlugin(name, command, pluginFirst, prepare, generate)
}

// RegisterPluginLast adds a plugin after first and normal plugins. Plugins in
// this group run by name. Plugins with the same name run in registration order.
func RegisterPluginLast(name, command string, prepare PrepareFunc, generate GenerateFunc) {
	registerPlugin(name, command, pluginLast, prepare, generate)
}

// register validates and records one plugin definition before generation.
func registerPlugin(name, command string, position pluginPosition, prepare PrepareFunc, generate GenerateFunc) {
	validatePlugin(name, command, generate)
	pluginregistry.Register(name, command, position, prepare, generate)
}

// validatePlugin rejects definitions that the generator cannot execute.
func validatePlugin(name, command string, generate GenerateFunc) {
	if name == "" {
		panic("plugin name is empty")
	}
	if command != "gen" && command != "example" {
		panic(fmt.Sprintf("unknown generator command %q", command))
	}
	if generate == nil {
		panic("plugin generate function is nil")
	}
}
