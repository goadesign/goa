package openapiv2

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func TestDefinitionNameRecursiveCollision(t *testing.T) {
	for _, fields := range [][]string{{"first", "second"}, {"second", "first"}} {
		t.Run(fields[0], func(t *testing.T) {
			builder := newSchemaBuilder(openapi.Values{})
			generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test"))
			makeNode := func(field string) *expr.UserTypeExpr {
				node := &expr.UserTypeExpr{TypeName: "Node", AttributeExpr: &expr.AttributeExpr{Meta: expr.MetaExpr{"openapi:typename": {"Node"}}}}
				node.Type = &expr.Object{
					&expr.NamedAttributeExpr{Name: field, Attribute: &expr.AttributeExpr{Type: expr.String}},
					&expr.NamedAttributeExpr{Name: "children", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: node}}}},
				}
				node.Validation = &expr.ValidationExpr{Required: []string{field, "children"}}
				return node
			}
			first, second := makeNode(fields[0]), makeNode(fields[1])
			require.Equal(t, "#/$defs/Node", builder.typeRefWithPrefix(nil, first, "", generator))
			require.Equal(t, "#/$defs/Node2", builder.typeRefWithPrefix(nil, second, "", generator))
			require.Equal(t, "#/$defs/Node2", builder.typeRefWithPrefix(nil, makeNode(fields[1]), "", generator))
			for i, name := range []string{"Node", "Node2"} {
				schema := builder.definitions[name]
				require.Contains(t, schema.Properties, fields[i])
				require.Equal(t, []string{fields[i], "children"}, schema.Required)
				require.Equal(t, "#/$defs/"+name, schema.Properties["children"].Items.Ref)
			}
			other := makeNode(fields[0])
			other.Meta["openapi:typename"] = []string{"Other"}
			require.Equal(t, "#/$defs/Other", builder.typeRefWithPrefix(nil, other, "", generator))
			delete(other.Meta, "openapi:typename")
			require.Equal(t, "#/$defs/PrefixNode", builder.typeRefWithPrefix(nil, other, "Prefix", generator))
		})
	}
}

func TestDefinitionNameFormatterValues(t *testing.T) {
	cases := []struct {
		name   string
		change func(*expr.AttributeExpr)
	}{
		{"default", func(at *expr.AttributeExpr) { at.Find("value").DefaultValue = "second" }},
		{"example", func(at *expr.AttributeExpr) { at.Find("value").UserExamples = []*expr.ExampleExpr{{Value: "second"}} }},
		{"constraint", func(at *expr.AttributeExpr) { at.Validation = &expr.ValidationExpr{Required: []string{"value"}} }},
		{"visibility", func(at *expr.AttributeExpr) { at.Find("value").Meta = expr.MetaExpr{"openapi:generate": {"false"}} }},
		{"extension", func(at *expr.AttributeExpr) { at.Meta = expr.MetaExpr{"openapi:extension:x-test": {"second"}} }},
		{"additional properties", func(at *expr.AttributeExpr) { at.Meta = expr.MetaExpr{"openapi:additionalProperties": {"false"}} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first := &expr.UserTypeExpr{TypeName: "Item", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}}}
			second := &expr.UserTypeExpr{TypeName: "Item", AttributeExpr: expr.DupAtt(first.AttributeExpr)}
			tc.change(second.AttributeExpr)
			builder := newSchemaBuilder(openapi.Values{})
			require.Equal(t, "Item", builder.assignDefinitionName("Item", first))
			require.Equal(t, "Item2", builder.assignDefinitionName("Item", second))
		})
	}
}

func TestDefinitionNameDocumentValues(t *testing.T) {
	for _, example := range []bool{false, true} {
		first := &expr.UserTypeExpr{TypeName: "Item", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			&expr.NamedAttributeExpr{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
		}}}
		second := &expr.UserTypeExpr{TypeName: "Item", AttributeExpr: &expr.AttributeExpr{Type: first.Type}}
		values := openapi.Values{}
		if example {
			values = values.WithExamples(second.AttributeExpr, []*expr.ExampleExpr{{Value: map[string]any{"value": "second"}}})
		} else {
			values = values.WithDescription(second.AttributeExpr, "Second description")
		}
		builder := newSchemaBuilder(values)
		require.Equal(t, "Item", builder.assignDefinitionName("Item", first))
		require.Equal(t, "Item2", builder.assignDefinitionName("Item", second))
	}
}

func TestDefinitionNamePrimitiveAliases(t *testing.T) {
	alias := func(name string, required bool) *expr.UserTypeExpr {
		at := &expr.AttributeExpr{Type: expr.Int}
		if required {
			minimum := 1.0
			at.Validation = &expr.ValidationExpr{Minimum: &minimum}
		}
		return &expr.UserTypeExpr{TypeName: name, AttributeExpr: at}
	}
	item := func(typ expr.DataType) *expr.UserTypeExpr {
		return &expr.UserTypeExpr{TypeName: "Item", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			&expr.NamedAttributeExpr{Name: "value", Attribute: &expr.AttributeExpr{Type: typ}},
		}}}
	}
	builder := newSchemaBuilder(openapi.Values{})
	require.Equal(t, "Item", builder.assignDefinitionName("Item", item(alias("Age", false))))
	require.Equal(t, "Item", builder.assignDefinitionName("Item", item(alias("Years", false))))
	require.Equal(t, "Item", builder.assignDefinitionName("Item", item(expr.Int)))
	require.Equal(t, "Item2", builder.assignDefinitionName("Item", item(alias("Positive", true))))
}

func TestDefinitionNamePrimitiveAliasFieldOverride(t *testing.T) {
	item := func(aliasName string, aliasMinimum float64) *expr.UserTypeExpr {
		alias := &expr.UserTypeExpr{TypeName: aliasName, AttributeExpr: &expr.AttributeExpr{
			Type: expr.Int, Validation: &expr.ValidationExpr{Minimum: &aliasMinimum},
		}}
		fieldMinimum := 10.0
		return &expr.UserTypeExpr{TypeName: "Item", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			&expr.NamedAttributeExpr{Name: "value", Attribute: &expr.AttributeExpr{
				Type: alias, Validation: &expr.ValidationExpr{Minimum: &fieldMinimum},
			}},
		}}}
	}
	first, second := item("Age", 1), item("Years", 2)
	builder := newSchemaBuilder(openapi.Values{})
	generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test"))
	require.Equal(t, "#/$defs/Item", builder.typeRefWithPrefix(nil, first, "", generator))
	require.Equal(t, "#/$defs/Item", builder.typeRefWithPrefix(nil, second, "", generator))
	require.Equal(t, 10.0, *builder.definitions["Item"].Properties["value"].Minimum)
	require.Len(t, builder.definitions, 1)
}
