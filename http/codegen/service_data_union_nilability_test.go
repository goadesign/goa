// This file verifies HTTP union records preserve the nilability of every
// branch when rendering a value that holds one selected branch.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestBuildHTTPUnionTypeDataMarksNilableBranches(t *testing.T) {
	union := unionWithBranchTypes()
	kindNames := []string{"ValueKindArray", "ValueKindBool", "ValueKindBytes", "ValueKindMap", "ValueKindObject", "ValueKindString"}
	constructorNames := []string{"NewValueArray", "NewValueBool", "NewValueBytes", "NewValueMap", "NewValueObject", "NewValueString"}
	kindDeclarations := make([]*codegen.NameDeclaration, len(kindNames))
	constructorDeclarations := make([]*codegen.NameDeclaration, len(constructorNames))
	for index := range kindNames {
		kindDeclarations[index] = codegen.NewExactName(codegen.NameConstant, kindNames[index])
		constructorDeclarations[index] = codegen.NewExactName(codegen.NameFunction, constructorNames[index])
	}
	record := &wireUnionRecord{
		declaration:  codegen.NewExactName(codegen.NameType, "Value"),
		kind:         codegen.NewExactName(codegen.NameType, "ValueKind"),
		kindDecls:    kindDeclarations,
		ctorDecls:    constructorDeclarations,
		name:         "Value",
		kindName:     "ValueKind",
		kindConsts:   kindNames,
		constructors: constructorNames,
	}
	data := buildHTTPUnionTypeData(union, codegen.NewAttributeScope(codegen.NewNameScope()), record)

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
