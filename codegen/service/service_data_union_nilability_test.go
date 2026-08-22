// This file verifies pointer/value semantics recorded for generated union
// branch fields.
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestBuildUnionTypeDataMarksNilableBranches(t *testing.T) {
	union := unionWithBranchTypes()
	generation := mustTestGeneration(t, "gen", nil)
	pkg := mustClaimTestPackage(t, generation, "gen/service")
	declaration, err := pkg.DeclareUnion(union)
	require.NoError(t, err)
	facts := &unionFacts{
		union:       union,
		identity:    codegen.NewUnionTypeID(union),
		typeKey:     union.GetTypeKey(),
		valueKey:    union.GetValueKey(),
		location:    &codegen.Location{RelImportPath: "gen/service"},
		declaration: declaration,
	}
	require.NoError(t, planUnionRenderFacts(facts, nil, pkg))
	require.NoError(t, generation.Freeze())
	data := buildRetainedUnionTypeData(facts, &importAliases{generation: generation})

	nilable := make(map[string]bool, len(data.Fields))
	for _, field := range data.Fields {
		nilable[field.Name] = field.Nilable
	}
	assert.Equal(t, map[string]bool{
		"array":  true,
		"bool":   false,
		"bytes":  true,
		"map":    true,
		"object": true,
		"string": false,
	}, nilable)
}

func unionWithBranchTypes() *expr.Union {
	return &expr.Union{
		TypeName: "value",
		Values: []*expr.NamedAttributeExpr{
			{Name: "array", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}}},
			{Name: "bool", Attribute: &expr.AttributeExpr{Type: expr.Boolean}},
			{Name: "bytes", Attribute: &expr.AttributeExpr{Type: expr.Bytes}},
			{Name: "map", Attribute: &expr.AttributeExpr{Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: expr.String},
			}}},
			{Name: "object", Attribute: &expr.AttributeExpr{Type: &expr.Object{}}},
			{Name: "string", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
	}
}
