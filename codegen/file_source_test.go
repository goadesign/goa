// This file checks that in-memory source production shares formatting with
// disk rendering while leaving file existence and finalization to Render.
package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFileSourceMatchesRenderWithoutEffects preserves disk-rendered bytes and
// proves source inspection does not run the file's external finalizer.
func TestFileSourceMatchesRenderWithoutEffects(t *testing.T) {
	finalized := false
	file := &File{
		Path: "nested/example.go",
		SectionTemplates: []*SectionTemplate{{
			Name: "example", Source: "package generated\nimport \"strings\"\nvar Value=1\n",
		}},
		FinalizeFunc: func(string) error {
			finalized = true
			return nil
		},
	}
	source, err := file.Source()
	require.NoError(t, err)
	require.False(t, finalized)
	directory := t.TempDir()
	path, err := file.Render(directory)
	require.NoError(t, err)
	require.True(t, finalized)
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, source, written)
	require.NotContains(t, string(source), "strings")
	source[0] = 'x'
	fresh, err := file.Source()
	require.NoError(t, err)
	require.Equal(t, written, fresh)
	file.SkipExist = true
	require.NoError(t, os.WriteFile(filepath.Join(directory, file.Path), []byte("existing"), 0o600))
	fresh, err = file.Source()
	require.NoError(t, err)
	require.Equal(t, written, fresh)
}
