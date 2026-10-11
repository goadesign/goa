// These tests exercise exact conversion equality without requiring a rendered
// service and measure comparisons of deeply nested named object types.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func BenchmarkGRPCHelperDefinitionNestedTypes(b *testing.B) {
	for _, depth := range []int{16, 64} {
		b.Run(fmt.Sprint(depth), func(b *testing.B) {
			attribute := &expr.AttributeExpr{Type: expr.String}
			for index := range depth {
				attribute = &expr.AttributeExpr{Type: &expr.UserTypeExpr{
					TypeName: fmt.Sprintf("Level%d", index),
					AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
						{Name: "child:wire", Attribute: attribute},
					}, Validation: &expr.ValidationExpr{Required: []string{"child"}}},
				}}
			}
			other := expr.DupAtt(attribute)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				require.True(b, sameGRPCHelperDefinition(attribute, attribute, other, other))
			}
		})
	}
}

func TestGRPCHelperDefinitionKeepsMappedRequiredness(t *testing.T) {
	object := &expr.Object{{Name: "value:wire", Attribute: &expr.AttributeExpr{Type: expr.String}}}
	for _, named := range []bool{false, true} {
		for _, required := range []bool{false, true} {
			for _, override := range []bool{false, true} {
				attribute := &expr.AttributeExpr{Type: object}
				if required {
					attribute.Validation = &expr.ValidationExpr{Required: []string{"value"}}
				}
				if named {
					attribute = &expr.AttributeExpr{Type: &expr.UserTypeExpr{TypeName: "Value", AttributeExpr: attribute}}
				}
				other := expr.DupAtt(attribute)
				if override {
					other.Validation = &expr.ValidationExpr{}
					if !required {
						other.Validation.Required = []string{"value"}
					}
				}
				want := expr.NewMappedAttributeExpr(attribute).IsRequired("value") == expr.NewMappedAttributeExpr(other).IsRequired("value")
				require.Equal(t, want, sameGRPCHelperDefinition(attribute, attribute, other, other), "named=%v required=%v override=%v", named, required, override)
			}
		}
	}
}

func TestGRPCHelperDefinitionKeepsRecursiveConversionFacts(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*expr.AttributeExpr)
	}{
		{name: "default", change: func(a *expr.AttributeExpr) {
			a.DefaultValue = "changed"
		}},
		{name: "custom type", change: func(a *expr.AttributeExpr) {
			a.Meta = expr.MetaExpr{"struct:field:type": {"Value", "example.com/custom"}}
		}},
		{name: "wrapper", change: func(a *expr.AttributeExpr) {
			a.Meta = expr.MetaExpr{wrappedAttrMeta: {"true"}}
		}},
		{name: "protobuf type", change: func(a *expr.AttributeExpr) {
			a.Meta = expr.MetaExpr{"struct:field:proto": {"string"}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			node := &expr.UserTypeExpr{TypeName: "Node", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}}
			*expr.AsObject(node.Type) = expr.Object{
				{Name: "next", Attribute: &expr.AttributeExpr{Type: node}},
				{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}
			first := &expr.AttributeExpr{Type: node}
			next := expr.DupAtt(first)
			require.True(t, sameGRPCHelperDefinition(first, first, next, next))
			test.change(expr.AsObject(next.Type).Attribute("value"))
			require.False(t, sameGRPCHelperDefinition(first, first, next, next))
		})
	}
}
