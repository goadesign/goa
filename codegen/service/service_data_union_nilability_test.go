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
	unionAttribute := &expr.AttributeExpr{Type: union}
	generation := mustTestGeneration(t, "gen", nil)
	pkg := mustClaimTestPackage(t, generation, "gen/service")
	declaration, err := pkg.DeclareUnion(unionAttribute)
	require.NoError(t, err)
	facts := &unionFacts{
		attribute:   unionAttribute,
		union:       union,
		identity:    codegen.NewUnionDeclarationID(unionAttribute),
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

	storageNames := make(map[string]string, len(data.Fields))
	for _, field := range data.Fields {
		storageNames[field.Name] = field.StorageName
	}
	assert.Equal(t, map[string]string{
		"array":  "array",
		"bool":   "bool_",
		"bytes":  "bytes",
		"map":    "map_",
		"object": "object",
		"string": "string_",
	}, storageNames)
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
