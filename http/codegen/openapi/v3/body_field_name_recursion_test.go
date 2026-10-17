// This file checks recursive named bodies before schema expansion. Both mapped
// and unmapped children must find the reference recorded for their own type.
package openapiv3

import (
	"fmt"
	"hash/fnv"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func TestBodyFieldNameRecursiveHash(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		node := dsl.Type("Node", func() {
			dsl.Field(1, "node_name", dsl.String)
			dsl.Field(2, "children", dsl.ArrayOf("Node"))
			dsl.Required("node_name", "children")
		})
		document := dsl.Type("Document", func() { dsl.Field(1, "tree", node) })
		for _, name := range []string{"mapped", "plain"} {
			dsl.Service(name, func() {
				if name == "mapped" {
					dsl.HTTP(func() { dsl.BodyFieldName(node, "node_name:nodeName") })
				}
				dsl.Method("exchange", func() {
					dsl.Payload(document)
					dsl.Result(document)
					dsl.HTTP(func() { dsl.POST("/" + name) })
				})
			})
		}
	})
	for _, name := range []string{"mapped", "plain"} {
		t.Run(name, func(t *testing.T) {
			endpoint := root.API.HTTP.Service(name).Endpoint("exchange")
			for _, body := range []*expr.AttributeExpr{endpoint.Body, endpoint.Responses[0].Body} {
				attribute := body.Find("tree")
				typ := attribute.Type.(expr.UserType)
				child := expr.AsArray(typ.Attribute().Find("children").Type).ElemType
				sf := newSchemafier(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")), openapi.Values{})
				first := sf.hashAttribute(child, fnv.New64())
				sf.hashes[first] = []schemaReference{{name: "Node", ref: toRef("Node")}}
				mapped := expr.NewMappedAttributeExpr(typ.Attribute())
				require.NotNil(t, mapped)
				context := sf.at(expr.UserTypeExampleIdentity(typ)).member("children").arrayElement(0)
				second := context.hashAttribute(child, fnv.New64())
				require.Equal(t, first, second)
				require.Equal(t, []schemaReference{{name: "Node", ref: toRef("Node")}}, context.hashes[second])
				require.Same(t, typ, child.Type)
			}
		})
	}
	for _, order := range [][]string{{"mapped", "plain"}, {"plain", "mapped"}} {
		t.Run(order[0]+" first", func(t *testing.T) {
			sf := newSchemafier(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")), openapi.Values{})
			for _, name := range order {
				endpoint := root.API.HTTP.Service(name).Endpoint("exchange")
				request := sf.schemafy(endpoint.Body.Find("tree"))
				response := sf.schemafy(endpoint.Responses[0].Body.Find("tree"))
				require.Equal(t, request.Ref, response.Ref)
				schema := sf.schemas[nameFromRef(request.Ref)]
				require.Equal(t, request.Ref, schema.Properties["children"].Items.Ref)
				if name == "mapped" {
					require.Contains(t, schema.Properties, "nodeName")
					require.NotContains(t, schema.Properties, "node_name")
				} else {
					require.Contains(t, schema.Properties, "node_name")
					require.NotContains(t, schema.Properties, "nodeName")
				}
			}
			require.Len(t, sf.schemas, 2)
		})
	}
}

func TestSchemafyRecursiveNameCollision(t *testing.T) {
	makeNode := func(typeName, componentName, fieldName string) *expr.AttributeExpr {
		node := &expr.UserTypeExpr{TypeName: typeName}
		node.AttributeExpr = &expr.AttributeExpr{
			Meta: expr.MetaExpr{"openapi:typename": {componentName}},
			Type: &expr.Object{
				{Name: fieldName, Attribute: &expr.AttributeExpr{Type: expr.String}},
				{Name: "children", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: node}}}},
			},
			Validation: &expr.ValidationExpr{Required: []string{fieldName, "children"}},
		}
		return &expr.AttributeExpr{Type: node}
	}
	sf := newSchemafier(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")), openapi.Values{})
	first := sf.schemafy(makeNode("First", "Node", "first"))
	second := sf.schemafy(makeNode("Second", "Node", "second"))
	separatelyNamed := sf.schemafy(makeNode("Third", "Other", "first"))
	require.Equal(t, toRef("Node"), first.Ref)
	require.Equal(t, toRef("Node2"), second.Ref)
	require.Equal(t, toRef("Other"), separatelyNamed.Ref)
	for _, reference := range []*openapi.Schema{first, second, separatelyNamed} {
		schema := sf.schemas[nameFromRef(reference.Ref)]
		require.Equal(t, reference.Ref, schema.Properties["children"].Items.Ref)
	}
	require.Contains(t, sf.schemas["Node"].Properties, "first")
	require.Contains(t, sf.schemas["Node2"].Properties, "second")
}

func TestSchemafyExplicitNameReusesAssignedComponent(t *testing.T) {
	for _, preferred := range []bool{false, true} {
		t.Run(fmt.Sprintf("preferred %t", preferred), func(t *testing.T) {
			object := &expr.Object{{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}}}
			implicit := &expr.AttributeExpr{Type: &expr.UserTypeExpr{
				TypeName:      "Node",
				AttributeExpr: &expr.AttributeExpr{Type: object},
			}}
			sf := newSchemafier(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")), openapi.Values{})
			if preferred {
				implicit.Type.(*expr.UserTypeExpr).TypeName = "Original"
				sf.preferredNames[sf.hashAttribute(implicit, fnv.New64())] = []string{"Node"}
			}
			first := sf.schemafy(implicit)
			explicit := expr.DupAtt(implicit)
			explicit.Type.(expr.UserType).Attribute().Meta = expr.MetaExpr{"openapi:typename": {"Node"}}
			second := sf.schemafy(explicit)
			require.Equal(t, toRef("Node"), first.Ref)
			require.Equal(t, first.Ref, second.Ref)
			require.Len(t, sf.schemas, 1)
			explicitOther := expr.DupAtt(implicit)
			explicitOther.Type.(expr.UserType).Attribute().Meta = expr.MetaExpr{"openapi:typename": {"Other"}}
			other := sf.schemafy(explicitOther)
			require.Equal(t, toRef("Other"), other.Ref)
			require.Len(t, sf.schemas, 2)
		})
	}
}
