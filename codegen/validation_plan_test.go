// This file verifies that validation planning preserves service and view
// output without reading expressions after package names are fixed.
package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestValidationPlanPreservesRulesPathsAndRequiredness compares copied rule
// rendering with the existing service/view validation generator.
func TestValidationPlanPreservesRulesPathsAndRequiredness(t *testing.T) {
	minimum := 2.0
	minLength := 3
	attribute := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "name", Attribute: &expr.AttributeExpr{
				Type: expr.String,
				Validation: &expr.ValidationExpr{
					Pattern:   "^[a-z]+$",
					MinLength: &minLength,
				},
			}},
			{Name: "count", Attribute: &expr.AttributeExpr{
				Type:       expr.Int,
				Validation: &expr.ValidationExpr{Minimum: &minimum},
			}},
			{Name: "nested", Attribute: &expr.AttributeExpr{Type: &expr.Object{}}},
		},
		Validation: &expr.ValidationExpr{Required: []string{"name", "nested"}},
	}
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}
	legacyContext := NewAttributeContext(
		policy.Pointer,
		policy.IgnoreRequired,
		policy.UseDefault,
		"",
		NewNameScope(),
	)
	legacyContext.UnionPointer = policy.UnionPointer
	want := ValidationCode(attribute, nil, legacyContext, true, false, true, "target")

	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: policy,
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)

	attribute.Validation = nil
	attribute.Type = expr.String
	minimum = 99
	minLength = 99

	linked, err := plan.Link(layout.Link("generated.local/gen/service", validationPlanTestQualifier))
	require.NoError(t, err)
	require.Equal(t, want, linked.Render("target", "target"))
	require.Equal(t, []GoTypeImport{
		{Name: "utf8", Path: "unicode/utf8"},
		{Name: "goa", Path: "goa.design/goa/v3/pkg"},
	}, linked.Imports())
}

// TestValidationPlanImportsOnlyUsedRuntimePackages checks that standalone
// validation users receive the packages named directly by rendered checks.
func TestValidationPlanImportsOnlyUsedRuntimePackages(t *testing.T) {
	for _, test := range []struct {
		name            string
		validation      *expr.ValidationExpr
		wantPreferences []GoTypeImport
		wantImports     []GoTypeImport
	}{
		{
			name:       "no checks",
			validation: nil,
		},
		{
			name:       "pattern",
			validation: &expr.ValidationExpr{Pattern: "^[a-z]+$"},
			wantPreferences: []GoTypeImport{
				{Name: "goa", Path: "goa.design/goa/v3/pkg"},
			},
			wantImports: []GoTypeImport{
				{Name: "goa", Path: "goa.design/goa/v3/pkg"},
			},
		},
		{
			name: "string length",
			validation: func() *expr.ValidationExpr {
				minimum := 2
				return &expr.ValidationExpr{MinLength: &minimum}
			}(),
			wantPreferences: []GoTypeImport{
				{Path: "unicode/utf8"},
				{Name: "goa", Path: "goa.design/goa/v3/pkg"},
			},
			wantImports: []GoTypeImport{
				{Name: "utf8", Path: "unicode/utf8"},
				{Name: "goa", Path: "goa.design/goa/v3/pkg"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			attribute := &expr.AttributeExpr{Type: expr.String, Validation: test.validation}
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner:  "generated.local/gen/service",
				Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
			})
			require.NoError(t, err)
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
			require.NoError(t, err)
			require.Equal(t, test.wantPreferences, plan.ImportPreferences())
			linked, err := plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
			require.NoError(t, err)
			require.Equal(t, test.wantImports, linked.Imports())
		})
	}
}

// TestValidationPlanImportPreferencesIncludeExternalValidators checks that
// planning includes only packages containing validation functions that the
// generated checks call.
func TestValidationPlanImportPreferencesIncludeExternalValidators(t *testing.T) {
	const (
		owner      = "generated.local/gen/service"
		childOwner = "generated.local/gen/shared"
	)
	minimum := 1.0
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "count", Attribute: &expr.AttributeExpr{
			Type:       expr.Int,
			Validation: &expr.ValidationExpr{Minimum: &minimum},
		}},
	})
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	childDeclaration := declareGoTypeTestUserType(t, generation, childOwner, child)
	validator := NewExactName(NameFunction, "ValidateChild")
	require.NoError(t, generation.Package(childOwner).DeclareName(validator))
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: child}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: child}},
	}}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  owner,
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			child: {Owner: childOwner, Type: childDeclaration},
		}),
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
		Required: true,
		Bind: func(ValidatorBindingRequest) (*NameDeclaration, error) {
			return validator, nil
		},
	})
	require.NoError(t, err)

	require.Equal(t, []GoTypeImport{
		{Name: "goa", Path: "goa.design/goa/v3/pkg"},
		{Name: "shared", Path: childOwner},
	}, plan.ImportPreferences())
}

// TestValidationPlanUsesFinalRuntimeImportNames proves rendered checks and
// reported imports use the same collision-safe package names.
func TestValidationPlanUsesFinalRuntimeImportNames(t *testing.T) {
	minimum := 2
	attribute := &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimum},
	}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)
	linked, err := plan.Link(layout.Link(layout.Owner(), func(importPath string) string {
		switch importPath {
		case "goa.design/goa/v3/pkg":
			return "goa2"
		case "unicode/utf8":
			return "utf82"
		default:
			t.Fatalf("unexpected validation import %q", importPath)
			return ""
		}
	}))
	require.NoError(t, err)

	require.Equal(t, []GoTypeImport{
		{Name: "utf82", Path: "unicode/utf8"},
		{Name: "goa2", Path: "goa.design/goa/v3/pkg"},
	}, linked.Imports())
	code := linked.Render("target", "target")
	require.Contains(t, code, "utf82.RuneCountInString")
	require.Contains(t, code, "goa2.MergeErrors")
}

// TestValidationPlanSharesOptionalFieldGuard verifies that copied validation
// rules use the one nil check selected for their containing field.
func TestValidationPlanSharesOptionalFieldGuard(t *testing.T) {
	minLength := 2
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{
			Type: expr.String,
			Validation: &expr.ValidationExpr{
				Pattern:   "^[a-z]+$",
				MinLength: &minLength,
			},
		}},
	}}
	policy := GoLayoutPolicy{UseDefault: true, SumType: true}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: policy,
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)
	linked, err := plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
	require.NoError(t, err)

	code := linked.Render("target", "target")
	require.Equal(t, 1, strings.Count(code, "if target.Name != nil"))
	require.Contains(t, code, "goa.ValidatePattern")
	require.Contains(t, code, "goa.InvalidLengthError")
}

// TestValidationPlanCopiesEnumValues verifies that accepted mutable enum
// values cannot change a validation program after planning.
func TestValidationPlanCopiesEnumValues(t *testing.T) {
	bytesValue := []byte{1, 2}
	arrayValue := []any{
		bytesValue,
		map[string]any{"nested": []any{"kept"}},
	}
	mapValue := map[string]any{"array": arrayValue}
	attribute := &expr.AttributeExpr{
		Type: expr.Any,
		Validation: &expr.ValidationExpr{Values: []any{
			bytesValue,
			arrayValue,
			mapValue,
		}},
	}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)

	bytesValue[0] = 9
	arrayValue[1].(map[string]any)["nested"].([]any)[0] = "changed"
	mapValue["added"] = true
	attribute.Validation.Values[0] = "replaced"

	require.Equal(t, []any{
		[]byte{1, 2},
		[]any{
			[]byte{1, 2},
			map[string]any{"nested": []any{"kept"}},
		},
		map[string]any{"array": []any{
			[]byte{1, 2},
			map[string]any{"nested": []any{"kept"}},
		}},
	}, plan.root.rules.values)
}

// TestNeedsValidation reports whether the validation renderer can write code
// for local rules, nested rules, and values with no rules.
func TestNeedsValidation(t *testing.T) {
	minimum := 1.0
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "count", Attribute: &expr.AttributeExpr{
			Type:       expr.Int,
			Validation: &expr.ValidationExpr{Minimum: &minimum},
		}},
	})
	tests := []struct {
		name      string
		attribute *expr.AttributeExpr
		want      bool
	}{
		{
			name: "local rule",
			attribute: &expr.AttributeExpr{
				Type:       expr.String,
				Validation: &expr.ValidationExpr{Pattern: ".+"},
			},
			want: true,
		},
		{
			name: "nested rule",
			attribute: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
			}},
			want: true,
		},
		{
			name:      "no rules",
			attribute: &expr.AttributeExpr{Type: &expr.Object{{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}}}},
		},
	}
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, NeedsValidation(test.attribute, policy))
		})
	}
}

// TestValidationPlanChecksOnlyRepresentableNullElements verifies that a null
// check follows the generated element type instead of the raw DSL flag.
func TestValidationPlanChecksOnlyRepresentableNullElements(t *testing.T) {
	array := &expr.Array{
		ElemType: &expr.AttributeExpr{
			Type:       expr.String,
			Validation: &expr.ValidationExpr{Pattern: "[a-z]+"},
		},
		NonNullableElems: true,
	}
	attribute := &expr.AttributeExpr{Type: array}
	tests := []struct {
		name      string
		jsonBody  bool
		wantCheck bool
	}{
		{name: "service values"},
		{name: "JSON input pointers", jsonBody: true, wantCheck: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := GoLayoutPolicy{
				UseDefault:          true,
				SumType:             true,
				ArrayElementPointer: test.jsonBody,
			}
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner:  "generated.local/gen/service",
				Policy: policy,
			})
			require.NoError(t, err)
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
			require.NoError(t, err)
			linked, err := plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
			require.NoError(t, err)
			code := linked.Render("target", "target")
			require.Equal(t, test.wantCheck, strings.Contains(code, "e == nil"))
			if test.jsonBody {
				require.Contains(t, code, "goa.ValidatePattern(\"target[*]\", *e, \"[a-z]+\")")
			} else {
				require.Contains(t, code, "goa.ValidatePattern(\"target[*]\", e, \"[a-z]+\")")
			}
			require.True(t, NeedsValidation(attribute, policy))
		})
	}

	objectArray := &expr.AttributeExpr{Type: &expr.Array{
		ElemType:         &expr.AttributeExpr{Type: &expr.Object{}},
		NonNullableElems: true,
	}}
	policy := GoLayoutPolicy{UseDefault: true, SumType: true}
	layout, err := PlanGoType(objectArray, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: policy,
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(objectArray, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)
	linked, err := plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
	require.NoError(t, err)
	require.Contains(t, linked.Render("target", "target"), "e == nil")
	require.True(t, NeedsValidation(objectArray, policy))
}

// TestNeedsValidationChecksEverySiblingCopy verifies that one unconstrained
// copy of a type does not hide rules on another copy of the same type.
func TestNeedsValidationChecksEverySiblingCopy(t *testing.T) {
	minLength := 2
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
	})
	unvalidated := expr.DupAtt(&expr.AttributeExpr{Type: child})
	validated := expr.DupAtt(&expr.AttributeExpr{Type: child})
	expr.AsObject(validated.Type.(expr.UserType).Attribute().Type).Attribute("value").Validation =
		&expr.ValidationExpr{MinLength: &minLength}

	tests := []struct {
		name   string
		fields *expr.Object
	}{
		{
			name: "unvalidated copy first",
			fields: &expr.Object{
				{Name: "first", Attribute: unvalidated},
				{Name: "second", Attribute: validated},
			},
		},
		{
			name: "validated copy first",
			fields: &expr.Object{
				{Name: "first", Attribute: validated},
				{Name: "second", Attribute: unvalidated},
			},
		},
	}
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attribute := &expr.AttributeExpr{Type: test.fields}
			require.True(t, NeedsValidation(attribute, policy))
		})
	}
}

// TestValidationPlanPreservesContainersUnionsAndValidatorBindings verifies all
// recursive service/view shapes retain exact nested validator declarations.
func TestValidationPlanPreservesContainersUnionsAndValidatorBindings(t *testing.T) {
	const owner = "generated.local/gen/service"
	minLength := 1
	minimum := 4.0
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{
			Type:       expr.String,
			Validation: &expr.ValidationExpr{MinLength: &minLength},
		}},
	})
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "label", Attribute: &expr.AttributeExpr{
				Type:       expr.String,
				Validation: &expr.ValidationExpr{Pattern: ".+"},
			}},
			{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
		},
	}
	mapKey := &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minLength},
	}
	mapValue := &expr.AttributeExpr{
		Type:       expr.Int,
		Validation: &expr.ValidationExpr{Minimum: &minimum},
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "children", Attribute: &expr.AttributeExpr{Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: child},
		}}},
		{Name: "labels", Attribute: &expr.AttributeExpr{Type: &expr.Map{
			KeyType:  mapKey,
			ElemType: mapValue,
		}}},
		{Name: "choice", Attribute: &expr.AttributeExpr{Type: union}},
	}}
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}
	legacyContext := NewAttributeContext(
		policy.Pointer,
		policy.IgnoreRequired,
		policy.UseDefault,
		"",
		NewNameScope(),
	)
	legacyContext.UnionPointer = policy.UnionPointer
	want := ValidationCode(attribute, nil, legacyContext, true, false, true, "target")

	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	childDeclaration := declareGoTypeTestUserType(t, generation, owner, child)
	unionDeclaration := declareGoTypeTestUnion(t, generation, owner, union)
	generatedPackage := generation.Package(owner)
	validator := NewExactName(NameFunction, "ValidateChild")
	require.NoError(t, generatedPackage.DeclareName(validator))

	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  owner,
		Policy: policy,
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			child: {Owner: owner, Type: childDeclaration},
			union: {Owner: owner, Union: unionDeclaration},
		}),
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
		Required: true,
		Bind: func(request ValidatorBindingRequest) (*NameDeclaration, error) {
			require.Same(t, child, request.Attribute.Type)
			require.Equal(t, owner, request.Layout.Owner())
			require.Empty(t, request.View)
			return validator, nil
		},
	})
	require.NoError(t, err)
	require.Equal(t, []*NameDeclaration{validator, validator}, plan.ValidatorDeclarations())

	mapKey.Validation = nil
	mapValue.Validation = nil
	union.Values = nil
	child.SetAttribute(&expr.AttributeExpr{Type: expr.String})
	attribute.Type = expr.String

	require.NoError(t, generation.Freeze())
	linked, err := plan.Link(layout.Link(owner, validationPlanTestQualifier))
	require.NoError(t, err)
	require.Equal(t, want, linked.Render("target", "target"))
	require.Equal(t, []GoTypeImport{
		{Name: "utf8", Path: "unicode/utf8"},
		{Name: "goa", Path: "goa.design/goa/v3/pkg"},
	}, linked.Imports())
}

// TestValidationPlanRejectsUnboundNestedValidator verifies planning never
// falls back to reconstructing a validator name from a user type.
func TestValidationPlanRejectsUnboundNestedValidator(t *testing.T) {
	minLength := 1
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{
			Type:       expr.String,
			Validation: &expr.ValidationExpr{MinLength: &minLength},
		}},
	})
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
	}}
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	declaration := declareGoTypeTestUserType(t, generation, "generated.local/gen/service", child)
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true},
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			child: {Owner: "generated.local/gen/service", Type: declaration},
		}),
	})
	require.NoError(t, err)

	_, err = NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.EqualError(t, err, "plan validation for field \"child\": validator binder must not be nil")
}

// validationPlanTestQualifier resolves the focused generated package aliases.
func validationPlanTestQualifier(importPath string) string {
	switch importPath {
	case "generated.local/gen/service":
		return "service"
	case "goa.design/goa/v3/pkg":
		return "goa"
	case "unicode/utf8":
		return "utf8"
	default:
		panic(fmt.Sprintf("unexpected validation import %q", importPath))
	}
}
