// This file verifies the released plugin registration calls without running a
// second generation pipeline. The generator package consumes the copied
// registrations tested here.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/internal/pluginregistry"
	"goa.design/goa/v3/eval"
)

// TestReleasedPluginRegistrationSignatures verifies that existing plugin
// packages still compile against the four-argument Goa v3 API.
func TestReleasedPluginRegistrationSignatures(t *testing.T) {
	var register func(string, string, PrepareFunc, GenerateFunc)

	register = RegisterPlugin
	require.NotNil(t, register)
	register = RegisterPluginFirst
	require.NotNil(t, register)
	register = RegisterPluginLast
	require.NotNil(t, register)
}

// TestPluginRegistryRejectsInvalidRegistration verifies that invalid plugin
// definitions fail when they are registered, before generation can start.
func TestPluginRegistryRejectsInvalidRegistration(t *testing.T) {
	tests := []struct {
		name     string
		plugin   string
		command  string
		generate GenerateFunc
		error    string
	}{
		{name: "empty name", command: "gen", generate: unchangedFiles, error: "plugin name is empty"},
		{name: "unknown command", plugin: "plugin", command: "other", generate: unchangedFiles, error: `unknown generator command "other"`},
		{name: "missing generate", plugin: "plugin", command: "gen", error: "plugin generate function is nil"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := pluginregistry.New()
			require.PanicsWithValue(t, test.error, func() {
				registerPluginIn(registry, test.plugin, test.command, pluginNormal, test.generate)
			})
		})
	}
}

// TestPluginRegistryKeepsDuplicateRegistrationOrder verifies that the released
// API keeps every callback when packages reuse the same command and name.
func TestPluginRegistryKeepsDuplicateRegistrationOrder(t *testing.T) {
	registry := pluginregistry.New()
	registerPluginIn(registry, "plugin", "gen", pluginFirst, unchangedFiles)
	registerPluginIn(registry, "plugin", "gen", pluginLast, changedFiles)

	registrations := pluginregistry.SnapshotFrom[PrepareFunc, GenerateFunc](registry)
	require.Len(t, registrations, 2)
	require.Equal(t, pluginFirst, registrations[0].Position)
	require.Equal(t, pluginLast, registrations[1].Position)
	files, err := registrations[1].Generate("generated.local/gen", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "changed", files[0].Path)
	require.PanicsWithValue(t, "plugin registry is sealed", func() {
		registerPluginIn(registry, "late", "gen", pluginNormal, unchangedFiles)
	})
}

// TestPluginRegistrySnapshotIsCopied verifies that a caller cannot change the
// registrations retained for later generation runs.
func TestPluginRegistrySnapshotIsCopied(t *testing.T) {
	registry := pluginregistry.New()
	registerPluginIn(registry, "plugin", "gen", pluginNormal, unchangedFiles)

	first := pluginregistry.SnapshotFrom[PrepareFunc, GenerateFunc](registry)
	first[0].Name = "changed"
	second := pluginregistry.SnapshotFrom[PrepareFunc, GenerateFunc](registry)

	require.Equal(t, "plugin", second[0].Name)
	require.Equal(t, "gen", second[0].Command)
	require.Equal(t, pluginNormal, second[0].Position)
}

// registerPluginIn applies the public registration checks to an isolated
// registry so the test does not stop later process-wide registrations.
func registerPluginIn(registry *pluginregistry.Registry, name, command string, position pluginPosition, generate GenerateFunc) {
	validatePlugin(name, command, generate)
	pluginregistry.RegisterIn[PrepareFunc, GenerateFunc](registry, name, command, position, nil, generate)
}

// unchangedFiles provides a valid generation callback for registration tests.
func unchangedFiles(_ string, _ []eval.Root, files []*File) ([]*File, error) {
	return files, nil
}

// changedFiles gives duplicate registration tests a distinct callback.
func changedFiles(_ string, _ []eval.Root, files []*File) ([]*File, error) {
	return append(files, &File{Path: "changed"}), nil
}
