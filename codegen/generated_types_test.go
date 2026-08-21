// This file verifies that one generated package owns the public names of every
// relocated declaration planned into it.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestGeneratedTypesRejectRelocatedNameCollision(t *testing.T) {
	var first, second expr.UserType
	root := RunDSL(t, func() {
		first = dsl.Type("foo-bar", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("first", dsl.String)
		})
		second = dsl.Type("foo_bar", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("second", dsl.String)
		})
	})

	generation := NewGeneration("generated.local/gen", []eval.Root{root})
	types := generation.GeneratedPackage("generated.local/gen/types")
	_, err := types.DeclareUserType(first)
	require.NoError(t, err)
	_, err = types.DeclareUserType(second)
	require.ErrorContains(t, err, "foo-bar")
	require.ErrorContains(t, err, "foo_bar")
	require.ErrorContains(t, err, "FooBar")
}
