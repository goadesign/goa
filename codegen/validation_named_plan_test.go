// Named definitions retain their declaration identity while validation reads
// their underlying fields and elements. These tests exercise both transport and
// service layouts and ensure rendering never rereads changed expressions.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestValidationPlanRetainsNamedDefinitionStructure(t *testing.T) {
	const owner = "generated.local/gen/types"
	for _, shape := range []string{"object", "array", "map", "union"} {
		for _, pointer := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/pointer=%t", shape, pointer), func(t *testing.T) {
				minimum, length := 1.0, 1
				text := &expr.AttributeExpr{
					Type: expr.String, Validation: &expr.ValidationExpr{MinLength: &length},
				}
				number := &expr.AttributeExpr{
					Type: expr.Int, Validation: &expr.ValidationExpr{Minimum: &minimum},
				}
				direct := &expr.AttributeExpr{}
				switch shape {
				case "object":
					direct.Type = &expr.Object{
						{Name: "label", Attribute: text},
						{Name: "count", Attribute: number},
					}
					direct.Validation = &expr.ValidationExpr{Required: []string{"label"}}
				case "array":
					direct.Type = &expr.Array{ElemType: text, NonNullableElems: true}
					direct.Validation = &expr.ValidationExpr{MinLength: &length}
				case "map":
					direct.Type = &expr.Map{KeyType: text, ElemType: number}
					direct.Validation = &expr.ValidationExpr{MinLength: &length}
				case "union":
					direct.Type = &expr.Union{
						TypeName: "Choice",
						Values: []*expr.NamedAttributeExpr{
							{Name: "text", Attribute: text},
							{Name: "number", Attribute: number},
						},
					}
				}
				base := &expr.UserTypeExpr{TypeName: "Base", AttributeExpr: direct}
				middle := goTypeTestUserType("Middle", base)
				attribute := &expr.AttributeExpr{Type: middle}
				control := *direct
				if shape == "array" || shape == "map" {
					outerLength := 2
					attribute.Validation = &expr.ValidationExpr{MaxLength: &outerLength}
					control.Validation = &expr.ValidationExpr{
						MinLength: &length, MaxLength: &outerLength,
					}
				}
				generation, err := NewGeneration("generated.local/gen", nil)
				require.NoError(t, err)
				baseDeclaration := declareGoTypeTestUserType(t, generation, owner, base)
				middleDeclaration := declareGoTypeTestUserType(t, generation, owner, middle)
				bindings := map[expr.DataType]GoTypeBinding{
					base:   {Owner: owner, Type: baseDeclaration},
					middle: {Owner: owner, Type: middleDeclaration},
				}
				if union, ok := direct.Type.(*expr.Union); ok {
					bindings[union] = GoTypeBinding{
						Owner: owner, Union: declareGoTypeTestUnion(t, generation, owner, union),
					}
				}
				options := GoTypePlanOptions{
					Owner: owner,
					Policy: GoLayoutPolicy{
						Pointer: pointer, UnionPointer: pointer, ArrayElementPointer: pointer,
						UseDefault: true, SumType: true,
					},
					Bind: goTypeTestBinder(bindings), RetainNamedValue: true,
				}
				directLayout, err := PlanGoType(&control, options)
				require.NoError(t, err)
				directValidation, err := NewValidationPlan(&control, directLayout, ValidationPlanOptions{Required: true})
				require.NoError(t, err)
				namedLayout, err := PlanGoType(attribute, options)
				require.NoError(t, err)
				namedValidation, err := NewValidationPlan(attribute, namedLayout, ValidationPlanOptions{Required: true})
				require.NoError(t, err)
				require.Same(t, middleDeclaration, namedLayout.TypeDeclaration())
				require.NoError(t, generation.Freeze())
				directLinked, err := directValidation.Link(directLayout.Link(owner, validationPlanTestQualifier))
				require.NoError(t, err)
				target := "value"
				wantImports := directValidation.ImportPreferences()
				if shape == "union" {
					receiver := directLayout.Link(owner, validationPlanTestQualifier).RefWithPointer(namedLayout.ReferenceIsPointer())
					target = fmt.Sprintf("(%s)(value)", receiver)
					wantImports = append(wantImports, GoTypeImport{Path: owner})
				}
				want := directLinked.Render(target, "body")
				require.NotEmpty(t, want)
				namedType := namedLayout.Link(owner, validationPlanTestQualifier)
				namedReference := namedType.Ref()
				require.Equal(t, middleDeclaration.Ref(middle), namedReference)

				attribute.Type = expr.Boolean
				attribute.Validation = nil
				base.AttributeExpr.Type = expr.String
				base.AttributeExpr.Validation = nil
				middle.Attribute().Type = expr.Int
				text.Validation, number.Validation = nil, nil
				minimum, length = 99, 99

				linked, err := namedValidation.Link(namedType)
				require.NoError(t, err)
				require.Equal(t, want, linked.Render("value", "body"))
				require.Equal(t, namedReference, namedType.Ref())
				require.Equal(t, wantImports, namedValidation.ImportPreferences())
			})
		}
	}
}

func TestValidationPlanRejectsMissingNamedDefinitionStructure(t *testing.T) {
	const owner = "generated.local/gen/types"
	text := &expr.AttributeExpr{Type: expr.String}
	for _, test := range []struct {
		name  string
		shape expr.DataType
		error string
	}{
		{"object", &expr.Object{{Name: "label", Attribute: text}}, "object layout has 0 fields, expected 1"},
		{"array", &expr.Array{ElemType: text}, "array layout has no element"},
		{"map", &expr.Map{KeyType: text, ElemType: text}, "map layout is incomplete"},
		{"union", &expr.Union{TypeName: "Choice", Values: []*expr.NamedAttributeExpr{{Name: "text", Attribute: text}}}, "union layout has 0 branches, expected 1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			base := goTypeTestUserType("Base", test.shape)
			attribute := &expr.AttributeExpr{Type: base}
			generation, err := NewGeneration("generated.local/gen", nil)
			require.NoError(t, err)
			declaration := declareGoTypeTestUserType(t, generation, owner, base)
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner: owner, Policy: GoLayoutPolicy{Pointer: true, SumType: true},
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
					base: {Owner: owner, Type: declaration},
				}),
			})
			require.NoError(t, err)
			require.NotPanics(t, func() {
				plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
				require.EqualError(t, err, "plan validation for root: "+test.error)
				require.Nil(t, plan)
			})
		})
	}
}

func TestValidationPlanRequiresInheritedPointerField(t *testing.T) {
	const owner = "generated.local/gen/types"
	base := goTypeTestUserType("Base", &expr.Object{
		{Name: "note", Attribute: &expr.AttributeExpr{Type: expr.String}},
	})
	attribute := &expr.AttributeExpr{
		Type: base, Validation: &expr.ValidationExpr{Required: []string{"note"}},
	}
	policy := GoLayoutPolicy{UseDefault: true, SumType: true}
	require.True(t, NeedsValidation(attribute, policy))
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	declaration := declareGoTypeTestUserType(t, generation, owner, base)
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner: owner, Policy: policy, RetainNamedValue: true,
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			base: {Owner: owner, Type: declaration},
		}),
	})
	require.NoError(t, err)
	require.True(t, layout.value.Fields()[0].IsPointer())
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	linked, err := plan.Link(layout.Link(owner, validationPlanTestQualifier))
	require.NoError(t, err)
	require.Contains(t, linked.Render("value", "body"), "if value.Note == nil")
}

func TestValidationPlanEvaluatedNamedRequiredField(t *testing.T) {
	const owner = "generated.local/gen/types"
	for _, longerChain := range []bool{false, true} {
		t.Run(fmt.Sprintf("longer-chain=%t", longerChain), func(t *testing.T) {
			root := RunDSL(t, func() {
				base := d.Type("Base", func() {
					d.Attribute("note", d.String)
				})
				parent := base
				if longerChain {
					parent = d.Type("Middle", base)
				}
				d.Type("Derived", parent, func() {
					d.Required("note")
				})
			})
			attribute := &expr.AttributeExpr{Type: root.UserType("Derived")}
			policy := GoLayoutPolicy{UseDefault: true, SumType: true}
			require.Equal(t, longerChain, attribute.IsPrimitivePointer("note", true))
			require.Equal(t, longerChain, NeedsValidation(attribute, policy))
			require.Contains(t, expr.EffectiveValidation(attribute).Required, "note")
			generation, err := NewGeneration("generated.local/gen", nil)
			require.NoError(t, err)
			bindings := make(map[expr.DataType]GoTypeBinding)
			for _, userType := range root.Types {
				bindings[userType] = GoTypeBinding{
					Owner: owner, Type: declareGoTypeTestUserType(t, generation, owner, userType),
				}
			}
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner: owner, Policy: policy, RetainNamedValue: true, Bind: goTypeTestBinder(bindings),
			})
			require.NoError(t, err)
			structure := layout
			for structure.value != nil {
				structure = structure.value
			}
			require.Equal(t, longerChain, structure.Fields()[0].IsPointer())
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			linked, err := plan.Link(layout.Link(owner, validationPlanTestQualifier))
			require.NoError(t, err)
			if longerChain {
				require.Contains(t, linked.Render("value", "body"), "if value.Note == nil")
			} else {
				require.Empty(t, linked.Render("value", "body"))
			}
		})
	}
}
