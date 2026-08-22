// This file verifies HTTP union records preserve the nilability of every
// branch when rendering their package-owned sum type.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestBuildHTTPUnionTypeDataMarksNilableBranches(t *testing.T) {
	union := unionWithBranchTypes()
	record := &wireUnionRecord{
		name:         "Value",
		kindName:     "ValueKind",
		kindConsts:   []string{"ValueKindArray", "ValueKindBool", "ValueKindBytes", "ValueKindMap", "ValueKindObject", "ValueKindString"},
		constructors: []string{"NewValueArray", "NewValueBool", "NewValueBytes", "NewValueMap", "NewValueObject", "NewValueString"},
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
