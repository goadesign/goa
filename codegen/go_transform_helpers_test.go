package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/expr"
)

func TestGoTransformHelpers(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	var (
		scope = NewNameScope()
		// types to test
		simple        = root.UserType("Simple")
		recursive     = root.UserType("Recursive")
		composite     = root.UserType("Composite")
		deep          = root.UserType("Deep")
		deepArray     = root.UserType("DeepArray")
		simpleAlias   = root.UserType("SimpleAlias")
		mapAlias      = root.UserType("NestedMapAlias")
		arrayMapAlias = root.UserType("ArrayMapAlias")
		collection    = root.UserType("ResultTypeCollection")
		// attribute contexts used in test cases
		defaultCtx = NewAttributeContext(false, false, true, "", scope)
	)
	tc := []struct {
		Name        string
		Type        expr.DataType
		HelperNames []string
	}{
		{"simple", simple, nil},
		{"recursive", recursive, []string{"transformRecursiveToRecursive"}},
		{"composite", composite, []string{"transformSimpleToSimple"}},
		{"deep", deep, []string{"transformCompositeToComposite", "transformSimpleToSimple"}},
		{"deep-array", deepArray, []string{"transformCompositeToComposite", "transformSimpleToSimple"}},
		{"simple-alias", simpleAlias, nil},
		{"nested-map-alias", mapAlias, nil},
		{"array-map-alias", arrayMapAlias, nil},
		{"result-type-collection", collection, []string{"transformResultTypeToResultType"}},
	}
	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			require.NotNil(t, c.Type, "source type not found in testdata")
			_, funcs, err := GoTransform(&expr.AttributeExpr{Type: c.Type}, &expr.AttributeExpr{Type: c.Type}, "source", "target", defaultCtx, defaultCtx, "", true)
			require.NoError(t, err)
			assert.Equal(t, len(c.HelperNames), len(funcs), "invalid helpers count")
			var actual []string
			if len(funcs) > 0 {
				actual = make([]string, len(funcs))
				for i, f := range funcs {
					actual[i] = f.Name
				}
			}
			assert.Equal(t, c.HelperNames, actual, "invalid helper names")
		})
	}
}

func TestGoTransformHelperNilGuardForPointerUserTypes(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	scope := NewNameScope()

	composite := root.UserType("Composite")
	require.NotNil(t, composite, "source type not found in testdata")

	pointerCtx := NewAttributeContext(true, false, false, "", scope)
	defaultCtx := NewAttributeContext(false, false, true, "", scope)

	_, funcs, err := GoTransform(&expr.AttributeExpr{Type: composite}, &expr.AttributeExpr{Type: composite}, "source", "target", pointerCtx, defaultCtx, "", true)
	require.NoError(t, err)
	require.NotEmpty(t, funcs)

	found := false
	for _, fn := range funcs {
		if fn.Name != "transformSimpleToSimple" {
			continue
		}
		found = true
		assert.Contains(t, fn.Code, "if v == nil {\n\treturn nil\n}\n")
	}
	require.True(t, found, "expected nested user type helper to be generated")
}
