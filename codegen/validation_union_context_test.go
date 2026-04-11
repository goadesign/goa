// These tests regress union validation code generation when sum-type object
// branches are validated in value contexts such as HTTP request bodies.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestUnionValidationPreservesValueContextForRequiredOnlyObjectBranches(t *testing.T) {
	root := RunDSL(t, requiredObjectUnionDSL)
	scope := NewNameScope()
	unionT := root.UserType("UnionUserValidate")
	att := &expr.AttributeExpr{Type: unionT}

	valueCode := ValidationCode(att, nil, NewAttributeContext(false, false, false, "", scope), true, false, false, "target")
	require.NotContains(t, valueCode, "ValidateSomeType(actual)")
	require.NotContains(t, valueCode, "ValidateSomeOtherType(actual)")
	require.NotContains(t, valueCode, "if actual != nil")

	pointerCode := ValidationCode(att, nil, NewAttributeContext(true, false, false, "", scope), true, false, false, "target")
	require.Contains(t, pointerCode, "ValidateSomeType(actual)")
	require.Contains(t, pointerCode, "ValidateSomeOtherType(actual)")
}

// requiredObjectUnionDSL defines a OneOf with required-only object branches so
// validation generation can distinguish pointer and value contexts.
func requiredObjectUnionDSL() {
	var someType = dsl.Type("SomeType", func() {
		dsl.Attribute("a", dsl.String)
		dsl.Required("a")
	})
	var someOtherType = dsl.Type("SomeOtherType", func() {
		dsl.Attribute("b", dsl.String)
		dsl.Required("b")
	})

	_ = dsl.Type("UnionUserValidate", func() {
		dsl.OneOf("values", func() {
			dsl.Attribute("SomeType", someType)
			dsl.Attribute("SomeOtherType", someOtherType)
		})
	})
}
