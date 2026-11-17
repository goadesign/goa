package dsl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestOneOfAnonymousBranchesKeepNamedReferences(t *testing.T) {
	for _, test := range []struct {
		name      string
		branch    func(string)
		reference func(expr.DataType) expr.DataType
	}{
		{"array", func(name string) { Attribute("children", ArrayOf(name)) }, func(dt expr.DataType) expr.DataType { return expr.AsArray(dt).ElemType.Type }},
		{"map", func(name string) { Attribute("children", MapOf(String, name)) }, func(dt expr.DataType) expr.DataType { return expr.AsMap(dt).ElemType.Type }},
		{"object", func(name string) { Attribute("children", func() { Attribute("child", name) }) }, func(dt expr.DataType) expr.DataType { return expr.AsObject(dt).Attribute("child").Type }},
	} {
		for _, mutual := range []bool{false, true} {
			t.Run(test.name+map[bool]string{false: "/self", true: "/mutual"}[mutual], func(t *testing.T) {
				var node, target expr.UserType
				codegen.RunDSL(t, func() {
					node = Type("Node", func() {
						name := "Node"
						if mutual {
							name = "Other"
						}
						OneOf("value", func() { test.branch(name) })
						Attribute("label", String)
					})
					target = node
					if mutual {
						target = Type("Other", func() {
							Attribute("node", node)
							Attribute("label", String)
						})
					}
				})
				union := expr.AsUnion(node.Attribute().Find("value").Type)
				child := test.reference(union.Values[0].Attribute.Type)
				require.Same(t, target, child)
				require.NotNil(t, child.(expr.UserType).Attribute().Find("label"))
				require.Equal(t, codegen.NewUnionTypeID(union), codegen.NewUnionTypeID(expr.AsUnion(expr.Dup(node).(expr.UserType).Attribute().Find("value").Type)))
			})
		}
	}
}

func TestOneOfAnonymousBranchAttributeProperties(t *testing.T) {
	var container, detail expr.UserType
	codegen.RunDSL(t, func() {
		detail = ResultType("application/vnd.detail", func() {
			Attributes(func() { Attribute("label", String) })
			View("tiny", func() { Attribute("label") })
		})
		container = Type("Container", func() {
			OneOf("choice", func() {
				Attribute("label", String, func() {
					Default("start")
					MinLength(1)
					Meta("example", "value")
				})
				Attribute("collections", ArrayOf(CollectionOf(detail)))
				Attribute("object", func() {
					Attribute("detail", detail, func() { View("tiny") })
					Required("detail")
				})
			})
		})
	})
	union := expr.AsUnion(container.Attribute().Find("choice").Type)
	branch := union.Values[0].Attribute
	definition := branch.Type.(expr.UserType).Attribute()
	require.Same(t, branch.AuthoredAttribute(), definition.AuthoredAttribute())
	require.Equal(t, "start", definition.DefaultValue)
	require.Equal(t, 1, *definition.Validation.MinLength)
	require.Equal(t, []string{"value"}, definition.Meta["example"])
	definition.Meta["example"] = []string{"changed"}
	minimum := 2
	definition.Validation.MinLength = &minimum
	require.Equal(t, []string{"value"}, branch.Meta["example"])
	require.Equal(t, 1, *branch.Validation.MinLength)
	collections := expr.AsArray(union.Values[1].Attribute.Type).ElemType.Type.(*expr.ResultTypeExpr)
	require.Equal(t, "DetailCollection", collections.Name())
	require.Equal(t, "tiny", collections.Views[0].Name)
	object := union.Values[2].Attribute.Type.(expr.UserType).Attribute()
	require.Equal(t, []string{"detail"}, object.Validation.Required)
	field := object.Find("detail")
	require.Same(t, detail, field.Type)
	require.Equal(t, []string{"tiny"}, field.Meta[expr.ViewMetaKey])
}
