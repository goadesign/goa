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
	generation := codegen.NewGeneration("gen", nil)
	pkg := generation.GeneratedPackage("gen/service")
	_, err := pkg.DeclareUnion(union)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	declaration, err := pkg.Union(union)
	require.NoError(t, err)
	data, err := buildUnionTypeData(
		union,
		declaration,
		newServiceResolver(generation, &expr.ServiceExpr{Name: "service"}, "gen/service"),
		&codegen.Location{RelImportPath: "gen/service"},
		false,
		func(branch *expr.NamedAttributeExpr) (*codegen.UnionBranchDeclaration, error) {
			return pkg.UnionBranch(union, branch.Name)
		},
	)
	assert.NoError(t, err)

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
