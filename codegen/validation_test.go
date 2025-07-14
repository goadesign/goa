package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
)

func TestRecursiveValidationCode(t *testing.T) {
	root := RunDSL(t, testdata.ValidationTypesDSL)
	var (
		scope = NewNameScope()

		integerT = root.UserType("Integer")
		stringT  = root.UserType("String")
		floatT   = root.UserType("Float")
		aliasT   = root.UserType("AliasType")
		userT    = root.UserType("UserType")
		arrayUT  = root.UserType("ArrayUserType")
		arrayT   = root.UserType("Array")
		mapT     = root.UserType("Map")
		unionT   = root.UserType("Union")
		rtT      = root.UserType("Result")
		rtcolT   = root.UserType("Collection")
		colT     = root.UserType("TypeWithCollection")
		deepT    = root.UserType("Deep")
	)
	cases := []struct {
		Name       string
		Type       expr.UserType
		Required   bool
		Pointer    bool
		UseDefault bool
	}{
		{"integer-required", integerT, true, false, false},
		{"integer-pointer", integerT, false, true, false},
		{"integer-use-default", integerT, false, false, true},
		{"float-required", floatT, true, false, false},
		{"float-pointer", floatT, false, true, false},
		{"float-use-default", floatT, false, false, true},
		{"string-required", stringT, true, false, false},
		{"string-pointer", stringT, false, true, false},
		{"string-use-default", stringT, false, false, true},
		{"alias-type", aliasT, true, false, false},
		{"user-type-required", userT, true, false, false},
		{"user-type-pointer", userT, false, true, false},
		{"user-type-default", userT, false, false, true},
		{"user-type-array-required", arrayUT, true, true, false},
		{"array-required", arrayT, true, false, false},
		{"array-pointer", arrayT, false, true, false},
		{"array-use-default", arrayT, false, false, true},
		{"map-required", mapT, true, false, false},
		{"map-pointer", mapT, false, true, false},
		{"map-use-default", mapT, false, false, true},
		{"union", unionT, true, false, false},
		{"result-type-pointer", rtT, false, true, false},
		{"collection-required", rtcolT, true, false, false},
		{"collection-pointer", rtcolT, false, true, false},
		{"type-with-collection-pointer", colT, false, true, false},
		{"type-with-embedded-type", deepT, false, true, false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := NewAttributeContext(c.Pointer, false, c.UseDefault, "", scope)
			code := ValidationCode(&expr.AttributeExpr{Type: c.Type}, nil, ctx, c.Required, expr.IsAlias(c.Type), false, "target")
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
			testutil.AssertGo(t, "testdata/golden/validation_"+c.Name+".go.golden", code)
		})
	}
	// Special case of unions with views
	t.Run("union-with-view", func(t *testing.T) {
		ctx := NewAttributeContext(false, false, false, "", scope)
		code := ValidationCode(&expr.AttributeExpr{Type: unionT}, nil, ctx, true, false, true, "target")
		code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
		testutil.AssertGo(t, "testdata/golden/validation_union-with-view.go.golden", code)
	})

	// Test case for OneOf with format validation in views (Issue #3747)
	t.Run("union-with-format-validation", func(t *testing.T) {
		// Test with pointer context (typical for views) to ensure union values are not treated as pointers
		ctx := NewAttributeContext(true, false, true, "", scope)
		root := RunDSL(t, testdata.UnionWithFormatValidationDSL)
		oneofT := root.UserType("OneOfWithFormat")
		code := ValidationCode(&expr.AttributeExpr{Type: oneofT}, nil, ctx, true, false, true, "target")
		code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
		assert.Equal(t, testdata.UnionWithFormatValidationCode, code)
	})
}
