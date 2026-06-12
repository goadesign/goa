package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
)

func TestRecursiveValidationCode(t *testing.T) {
	root := RunDSL(t, testdata.ValidationTypesDSL)
	var (
		scope = NewNameScope()

		integerT  = root.UserType("Integer")
		stringT   = root.UserType("String")
		floatT    = root.UserType("Float")
		aliasT    = root.UserType("AliasType")
		userT     = root.UserType("UserType")
		arrayUT   = root.UserType("ArrayUserType")
		arrayT    = root.UserType("Array")
		arrayReqT = root.UserType("ArrayRequired")
		mapT      = root.UserType("Map")
		unionT    = root.UserType("Union")
		rtT       = root.UserType("Result")
		rtcolT    = root.UserType("Collection")
		colT      = root.UserType("TypeWithCollection")
		deepT     = root.UserType("Deep")
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
		{"array-required-non-nullable-elems", arrayReqT, true, false, false},
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
		testutil.AssertGo(t, "testdata/golden/validation_union-with-format-validation.go.golden", code)
	})
}

// TestRecursiveValidationWithCycleGuard tests that recursive types are
// properly handled without infinite loops. The recursion guard should prevent
// cycles while still allowing validation of the same type in different contexts.
func TestRecursiveValidationWithCycleGuard(t *testing.T) {
	root := RunDSL(t, testdata.RecursiveValidationDSL)
	scope := NewNameScope()

	cases := []struct {
		Name     string
		TypeName string
		Pointer  bool
	}{
		{"recursive-type-pointer", "RecursiveType", true},
		{"recursive-type-required", "RecursiveType", false},
		{"container-with-recursive-array", "ContainerWithRecursiveArray", true},
		{"container-with-recursive-map", "ContainerWithRecursiveMap", true},
		{"nested-recursive-pointer", "NestedRecursive", true},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := NewAttributeContext(c.Pointer, false, false, "", scope)
			ut := root.UserType(c.TypeName)
			code := ValidationCode(&expr.AttributeExpr{Type: ut}, nil, ctx, !c.Pointer, false, false, "target")
			// Verify code is generated (not empty) and doesn't cause infinite recursion
			if code == "" {
				t.Error("Expected validation code to be generated")
			}
			// Verify the code can be formatted (indicates valid Go code)
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
			if code == "" {
				t.Error("Expected formatted code to be generated")
			}
		})
	}
}

// TestMultipleAliasTypesInSameStruct tests that multiple fields with the same
// alias type can be validated independently. Previously, the recursion guard
// would incorrectly block validation of the second field.
func TestMultipleAliasTypesInSameStruct(t *testing.T) {
	root := RunDSL(t, testdata.ValidationTypesDSL)
	scope := NewNameScope()

	aliasT := root.UserType("AliasType")
	ctx := NewAttributeContext(false, false, false, "", scope)

	code := ValidationCode(&expr.AttributeExpr{Type: aliasT}, nil, ctx, true, false, false, "target")
	code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")

	// Verify both alias fields are validated (required_alias and alias)
	// The code should contain validation for both fields
	if !strings.Contains(code, "required_alias") {
		t.Error("Expected validation code for 'required_alias' field")
	}
	if !strings.Contains(code, "target.alias") {
		t.Error("Expected validation code for 'alias' field")
	}
	// Verify both get pattern validation
	if strings.Count(code, "ValidatePattern") < 2 {
		t.Errorf("Expected at least 2 pattern validations (one per alias field), got %d", strings.Count(code, "ValidatePattern"))
	}
}

// TestChainedAliasValidationCode verifies that validation code generated for
// user type alias chains two and three levels deep merges the validations of
// every chain level and that generating the code does not mutate the design
// expression tree (no flattening write-backs, no pointer aliasing).
func TestChainedAliasValidationCode(t *testing.T) {
	root := RunDSL(t, testdata.ChainedAliasValidationDSL)
	scope := NewNameScope()

	cases := []struct {
		Name     string
		TypeName string
		Required bool
		Pointer  bool
	}{
		{"chain-holder-required", "ChainHolder", true, false},
		{"chain-holder-pointer", "ChainHolder", false, true},
		{"chain-mid", "ChainMid", true, false},
		{"chain-pass", "ChainPass", true, false},
		{"chain-obj-leaf", "ChainObjLeaf", true, false},
		{"chain-obj-holder", "ChainObjHolder", true, false},
		{"chain-parent-pointer", "ChainParent", false, true},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ut := root.UserType(c.TypeName)
			before := snapshotValidations(t, ut)

			ctx := NewAttributeContext(c.Pointer, false, false, "", scope)
			code := ValidationCode(&expr.AttributeExpr{Type: ut}, nil, ctx, c.Required, expr.IsAlias(ut), false, "target")
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")
			testutil.AssertGo(t, "testdata/golden/validation_"+c.Name+".go.golden", code)

			assertValidationsUnchanged(t, before)
		})
	}
}

// TestValidationPredicatesPure verifies that hasValidations and repeated
// ValidationCode invocations have no write side effects on the design tree
// and that repeated calls yield identical output (the legacy implementation
// flattened alias chain validations in place, so a first call could change
// the result of subsequent consumers).
func TestValidationPredicatesPure(t *testing.T) {
	root := RunDSL(t, testdata.ChainedAliasValidationDSL)
	scope := NewNameScope()
	ctx := NewAttributeContext(false, false, false, "", scope)

	child := root.UserType("ChainChild")
	before := snapshotValidations(t, child)
	require.True(t, hasValidations(ctx, child), "ChainChild contains alias chain validations")
	assertValidationsUnchanged(t, before)

	holder := root.UserType("ChainHolder")
	before = snapshotValidations(t, holder)
	att := &expr.AttributeExpr{Type: holder}
	first := ValidationCode(att, nil, ctx, true, false, false, "target")
	second := ValidationCode(att, nil, ctx, true, false, false, "target")
	require.Equal(t, first, second, "ValidationCode is not idempotent")
	assertValidationsUnchanged(t, before)
}

// validationSnapshot captures the identity and deep value of an attribute
// validation so mutations can be detected after running codegen.
type validationSnapshot struct {
	att *expr.AttributeExpr
	ptr *expr.ValidationExpr
	dup *expr.ValidationExpr
}

// snapshotValidations records the validation of every attribute reachable
// from the given user type, including down alias chains.
func snapshotValidations(t *testing.T, ut expr.UserType) []validationSnapshot {
	t.Helper()
	var snaps []validationSnapshot
	require.NoError(t, Walk(&expr.AttributeExpr{Type: ut}, func(a *expr.AttributeExpr) error {
		snaps = append(snaps, validationSnapshot{att: a, ptr: a.Validation, dup: dupValidationForTest(a.Validation)})
		return nil
	}))
	return snaps
}

// assertValidationsUnchanged verifies that every snapshotted attribute still
// holds the exact same *ValidationExpr pointer with unchanged contents.
func assertValidationsUnchanged(t *testing.T, snaps []validationSnapshot) {
	t.Helper()
	for _, s := range snaps {
		if s.ptr == nil {
			require.Nil(t, s.att.Validation, "validation was assigned to an attribute that had none")
			continue
		}
		require.Same(t, s.ptr, s.att.Validation, "attribute validation pointer was replaced")
		require.Equal(t, s.dup, dupValidationForTest(s.att.Validation), "attribute validation contents were mutated")
	}
}

// dupValidationForTest deep copies a validation, including the Values slice
// and scalar pointers, so in-place mutations of the original are detected.
func dupValidationForTest(v *expr.ValidationExpr) *expr.ValidationExpr {
	if v == nil {
		return nil
	}
	dupFloat := func(p *float64) *float64 {
		if p == nil {
			return nil
		}
		f := *p
		return &f
	}
	dupInt := func(p *int) *int {
		if p == nil {
			return nil
		}
		i := *p
		return &i
	}
	d := v.Dup()
	d.Values = append([]any(nil), v.Values...)
	d.ExclusiveMinimum = dupFloat(v.ExclusiveMinimum)
	d.Minimum = dupFloat(v.Minimum)
	d.Maximum = dupFloat(v.Maximum)
	d.ExclusiveMaximum = dupFloat(v.ExclusiveMaximum)
	d.MinLength = dupInt(v.MinLength)
	d.MaxLength = dupInt(v.MaxLength)
	return d
}

// TestAliasTypeInArrayAndMap tests that alias types work correctly when
// nested in arrays and maps, ensuring the recursion guard doesn't interfere.
func TestAliasTypeInArrayAndMap(t *testing.T) {
	root := RunDSL(t, testdata.ValidationTypesDSL)
	scope := NewNameScope()

	var (
		alias = root.UserType("Alias")
	)

	// Create a type with alias in array
	arrayWithAlias := &expr.AttributeExpr{
		Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: alias},
		},
	}

	// Create a type with alias in map
	mapWithAlias := &expr.AttributeExpr{
		Type: &expr.Map{
			KeyType:  &expr.AttributeExpr{Type: expr.String},
			ElemType: &expr.AttributeExpr{Type: alias},
		},
	}

	cases := []struct {
		Name   string
		Att    *expr.AttributeExpr
		Target string
	}{
		{"alias-in-array", arrayWithAlias, "target"},
		{"alias-in-map", mapWithAlias, "target"},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := NewAttributeContext(false, false, false, "", scope)
			code := ValidationCode(c.Att, nil, ctx, true, false, false, c.Target)
			code = FormatTestCode(t, "package foo\nfunc Validate() (err error){\n"+code+"}")

			// Verify validation code is generated
			if code == "" {
				t.Error("Expected validation code to be generated")
			}
			// Verify it contains validation for the alias type
			if !strings.Contains(code, "ValidatePattern") && !strings.Contains(code, "InvalidLengthError") {
				t.Error("Expected validation code to contain pattern or length validation for alias type")
			}
		})
	}
}
