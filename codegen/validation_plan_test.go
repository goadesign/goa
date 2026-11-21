// This file verifies that validation planning preserves service and view
// output without reading expressions after package names are fixed.
package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
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
			policy := GoLayoutPolicy{UseDefault: true, SumType: true}
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner:  "generated.local/gen/service",
				Policy: policy,
			})
			require.NoError(t, err)
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
			require.NoError(t, err)
			require.Equal(t, test.wantPreferences, plan.ImportPreferences())
			require.Equal(t, test.wantPreferences, ValidationRuntimeImports(attribute, policy))
			linked, err := plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
			require.NoError(t, err)
			require.Equal(t, test.wantImports, linked.Imports())
		})
	}
}

// TestValidationRuntimeImportsStopAtNestedValidators verifies that a parent
// file does not reserve packages used only inside a named child validator.
func TestValidationRuntimeImportsStopAtNestedValidators(t *testing.T) {
	minimum := 1
	child := &expr.UserTypeExpr{
		TypeName: "Child",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "name", Attribute: &expr.AttributeExpr{
				Type:       expr.String,
				Validation: &expr.ValidationExpr{MinLength: &minimum},
			}},
		}},
	}
	parent := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
	}}
	policy := GoLayoutPolicy{UseDefault: true, SumType: true}

	require.Equal(t, []GoTypeImport{{Name: "goa", Path: GoaImport("").Path}}, ValidationRuntimeImports(parent, policy))
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
		Validation: &expr.ValidationExpr{MinLength: &minimum, Format: expr.FormatUUID},
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
	require.Contains(t, code, "goa2.ValidateFormat")
	require.Contains(t, code, "goa2.FormatUUID")
	require.NotContains(t, code, "goa.FormatUUID")
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

// TestNeedsValidationRecursiveCollections checks that walking a recursive array
// or map terminates and still finds a rule after the recursive field.
func TestNeedsValidationRecursiveCollections(t *testing.T) {
	for _, container := range []string{"array", "map"} {
		for _, constrained := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/constrained=%t", container, constrained), func(t *testing.T) {
				node := goTypeTestUserType("Node", &expr.Object{})
				element := &expr.AttributeExpr{Type: node}
				var children expr.DataType = &expr.Array{ElemType: element}
				if container == "map" {
					children = &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: element}
				}
				label := &expr.AttributeExpr{Type: expr.String}
				if constrained {
					label.Validation = &expr.ValidationExpr{Pattern: "^valid$"}
				}
				node.Attribute().Type = &expr.Object{
					{Name: "children", Attribute: &expr.AttributeExpr{Type: children}},
					{Name: "label", Attribute: label},
				}
				require.Equal(t, constrained, NeedsValidation(&expr.AttributeExpr{Type: node}, GoLayoutPolicy{SumType: true}))
			})
		}
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

// TestNeedsValidationSelectedResultViews checks that NeedsValidation inspects
// only the fields included in a nested result type's chosen view.
func TestNeedsValidationSelectedResultViews(t *testing.T) {
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}

	for _, tc := range []struct {
		name string
		dsl  func() *expr.AttributeExpr
		want bool
	}{
		{
			name: "default view unvalidated while extended view validated",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("extended", func() {
							dsl.Attribute("id")
							dsl.Attribute("name")
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("labels", dsl.ArrayOf(label, func() { dsl.View("default") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: false,
		},
		{
			name: "extended view validated inside array",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("extended", func() {
							dsl.Attribute("id")
							dsl.Attribute("name")
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("labels", dsl.ArrayOf(label, func() { dsl.View("extended") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: true,
		},
		{
			name: "default view validated while extended view omits validated field",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String, func() { dsl.MinLength(2) })
							dsl.Attribute("summary", dsl.String)
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("summary", func() { dsl.Attribute("summary") })
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("by_key", dsl.MapOf(dsl.String, label, func() {
							dsl.Elem(func() { dsl.View("summary") })
						}))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: false,
		},
		{
			name: "three-level cascaded view reaches leaf validation",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("extended", func() {
							dsl.Attribute("id")
							dsl.Attribute("name")
						})
					})
					item := dsl.ResultType("application/vnd.test.item", func() {
						dsl.TypeName("Item")
						dsl.Attributes(func() {
							dsl.Attribute("labels", dsl.ArrayOf(label))
						})
						dsl.View("default", func() { dsl.Attribute("labels") })
						dsl.View("extended", func() {
							dsl.Attribute("labels", func() { dsl.View("extended") })
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("items", dsl.ArrayOf(item, func() { dsl.View("extended") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: true,
		},
		{
			name: "mutually recursive result views terminate and find validation",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					node := dsl.ResultType("application/vnd.test.node", func() {
						dsl.TypeName("Node")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("title", dsl.String, func() { dsl.MinLength(3) })
							dsl.Attribute("children", dsl.ArrayOf("Node", func() {
								dsl.View("tiny")
							}))
						})
						dsl.View("tiny", func() {
							dsl.Attribute("id")
							dsl.Attribute("children", func() { dsl.View("default") })
						})
						dsl.View("default", func() {
							dsl.Attribute("id")
							dsl.Attribute("title")
							dsl.Attribute("children")
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("nodes", dsl.ArrayOf(node, func() { dsl.View("tiny") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			attribute := tc.dsl()
			assert.Equal(t, tc.want, NeedsValidation(attribute, policy))
		})
	}
}

// TestValidationPlanSelectedResultViews verifies that NewValidationPlan binds a
// nested result validator only when the chosen view contains checks, passing
// the non-default view name to the binder.
func TestValidationPlanSelectedResultViews(t *testing.T) {
	const owner = "generated.local/gen/service/views"
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}

	for _, tc := range []struct {
		name         string
		selectedView string
		wantBound    bool
		wantView     string
	}{
		{
			name:         "default view without checks skips validator binding",
			selectedView: expr.DefaultView,
			wantBound:    false,
		},
		{
			name:         "extended view with checks binds extended validator",
			selectedView: "extended",
			wantBound:    true,
			wantView:     "extended",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var label *expr.ResultTypeExpr
			var holder expr.UserType
			RunDSL(t, func() {
				label = dsl.ResultType("application/vnd.test.label", func() {
					dsl.TypeName("Label")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
					})
					dsl.View("default", func() { dsl.Attribute("id") })
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("name")
					})
				})
				holder = dsl.Type("Holder", func() {
					dsl.Attribute("labels", dsl.ArrayOf(label, func() { dsl.View(tc.selectedView) }))
				})
				dsl.Service("Test", func() {
					dsl.Method("Get", func() { dsl.Result(label); dsl.Payload(holder) })
				})
			})

			generation, err := NewGeneration("generated.local/gen", nil)
			require.NoError(t, err)
			labelDecl := declareGoTypeTestUserType(t, generation, owner, label)
			validator := NewExactName(NameFunction, "ValidateLabelViewExtended")
			require.NoError(t, generation.Package(owner).DeclareName(validator))

			layout, err := PlanGoType(holder.Attribute(), GoTypePlanOptions{
				Owner:  owner,
				Policy: policy,
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
					label: {Owner: owner, Type: labelDecl},
				}),
			})
			require.NoError(t, err)

			var bound int
			var gotView string
			plan, err := NewValidationPlan(holder.Attribute(), layout, ValidationPlanOptions{
				Required: true,
				Bind: func(request ValidatorBindingRequest) (*NameDeclaration, error) {
					bound++
					gotView = request.View
					return validator, nil
				},
			})
			require.NoError(t, err)
			if !tc.wantBound {
				assert.Zero(t, bound)
				assert.Empty(t, plan.ValidatorDeclarations())
				return
			}
			assert.Equal(t, 1, bound)
			assert.Equal(t, tc.wantView, gotView)
			assert.Equal(t, []*NameDeclaration{validator}, plan.ValidatorDeclarations())
		})
	}
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
