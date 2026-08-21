// This file verifies that plugin registration rejects ambiguous ownership
// before a generation run seals and snapshots the command registry.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPluginRegistrationRejectsInvalidIdentity verifies that malformed plugin
// ownership is rejected before any generation run snapshots the registry.
func TestPluginRegistrationRejectsInvalidIdentity(t *testing.T) {
	tests := []struct {
		name    string
		plugin  string
		command string
	}{
		{name: "empty plugin name", command: "test"},
		{name: "unknown command", plugin: "plugin", command: "missing"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := newRegistry()
			registry.addCommand("test")

			require.Panics(t, func() {
				registry.registerPlugin(test.plugin, test.command, pluginNormal, func() Plugin {
					return Plugin{}
				})
			})
		})
	}
}

// TestPluginRegistrationRejectsDuplicateCommandName verifies that a plugin
// cannot acquire two ordering positions for the same command and owner name.
func TestPluginRegistrationRejectsDuplicateCommandName(t *testing.T) {
	positions := []struct {
		name     string
		position pluginPosition
	}{
		{name: "first", position: pluginFirst},
		{name: "normal", position: pluginNormal},
		{name: "last", position: pluginLast},
	}

	for _, initial := range positions {
		for _, duplicate := range positions {
			t.Run(initial.name+" then "+duplicate.name, func(t *testing.T) {
				registry := newRegistry()
				registry.addCommand("test")
				registry.registerPlugin("plugin", "test", initial.position, func() Plugin {
					return Plugin{}
				})

				require.Panics(t, func() {
					registry.registerPlugin("plugin", "test", duplicate.position, func() Plugin {
						return Plugin{}
					})
				})
			})
		}
	}
}

// TestPluginRegistrationScopesNamesToCommand verifies that two commands may
// use the same owner name because each command snapshots its own plugin list.
func TestPluginRegistrationScopesNamesToCommand(t *testing.T) {
	registry := newRegistry()
	registry.addCommand("first")
	registry.addCommand("second")
	registry.registerPlugin("plugin", "first", pluginNormal, func() Plugin {
		return Plugin{}
	})
	registry.registerPlugin("plugin", "second", pluginNormal, func() Plugin {
		return Plugin{}
	})

	_, first, err := registry.snapshot("first")
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, "first", first[0].command)

	_, second, err := registry.snapshot("second")
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.Equal(t, "second", second[0].command)
}
