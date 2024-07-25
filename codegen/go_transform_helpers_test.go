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
			for _, f := range funcs {
				actual = append(actual, f.Name)
			}
			assert.Equal(t, c.HelperNames, actual, "invalid helper names")
		})
	}
}
