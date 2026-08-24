// This file protects released generator types that do not expose or run the
// internal generation sequence.
package generator

import (
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
