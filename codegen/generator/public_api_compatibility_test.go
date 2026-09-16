// This file protects released generator types that do not expose or run the
// internal generation sequence.
package generator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

// TestReleasedGeneratorFunctionType checks the released function signature.
func TestReleasedGeneratorFunctionType(t *testing.T) {
	var generate Genfunc = func(string, []eval.Root) ([]*codegen.File, error) {
		return []*codegen.File{{Path: "generated.go"}}, nil
	}
	files, err := generate("generated.local/gen", nil)
	require.NoError(t, err)
	require.Equal(t, "generated.go", files[0].Path)
}

// TestReleasedGeneratorEntryPoints checks the generator functions that plugins
// can call directly or return from Generators.
func TestReleasedGeneratorEntryPoints(t *testing.T) {
	var (
		service   Genfunc = Service
		transport Genfunc = Transport
		openAPI   Genfunc = OpenAPI
		example   Genfunc = Example
	)
	require.NotNil(t, service)
	require.NotNil(t, transport)
	require.NotNil(t, openAPI)
	require.NotNil(t, example)

	original := Generators
	t.Cleanup(func() {
		Generators = original
	})
	Generators = func(command string) ([]Genfunc, error) {
		if command != "custom" {
			return nil, fmt.Errorf("unknown command %q", command)
		}
		return []Genfunc{func(string, []eval.Root) ([]*codegen.File, error) {
			return []*codegen.File{{Path: "custom.go"}}, nil
		}}, nil
	}
	generators, err := Generators("custom")
	require.NoError(t, err)
	require.Len(t, generators, 1)
	require.NotNil(t, generators[0])

	run, err := newGenerationRun("custom", newDefaultRegistry())
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", nil)
	require.NoError(t, err)
	require.Len(t, result.files, 1)
	require.Equal(t, "custom.go", result.files[0].Path)
}
