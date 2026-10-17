// These tests check the declaration identities used by generated HTTP types.
// Repeated methods share mapped nested declarations, while separate services,
// derived types and different root body selections retain distinct identities.
package dsl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestBodyFieldNameDeclarationIdentity(t *testing.T) {
	for _, methods := range [][]string{{"whole", "headers"}, {"headers", "whole"}} {
		t.Run(methods[0]+" first", func(t *testing.T) {
			var item, derived expr.UserType
			root, err := runDSL(t, func() {
				item = Type("Item", func() {
					Attribute("display_name", String)
					Attribute("children", ArrayOf("Item"))
				})
				derived = Type("Derived", func() { Extend(item) })
				document := Type("Document", func() {
					Attribute("trace_id", String)
					Attribute("item", item)
					Attribute("derived", derived)
				})
				for _, service := range []string{"first", "second", "plain"} {
					Service(service, func() {
						HTTP(func() {
							if service != "plain" {
								BodyFieldName(item, "display_name:displayName")
								BodyFieldName(derived, "display_name:derivedName")
							}
						})
						for _, method := range methods {
							Method(method, func() {
								Payload(document)
								Result(document)
								HTTP(func() {
									POST("/" + service + "/" + method)
									if method == "headers" {
										Header("trace_id:X-Trace")
									}
									Response(200, func() {
										if method == "headers" {
											Header("trace_id:X-Trace")
										}
									})
								})
							})
						}
					})
				}
			})
			require.NoError(t, err)
			origins := make(map[string]expr.UserType)
			for _, service := range []string{"first", "second", "plain"} {
				svc := root.API.HTTP.Service(service)
				whole, headers := svc.Endpoint("whole"), svc.Endpoint("headers")
				for _, bodies := range [][2]*expr.AttributeExpr{
					{whole.Body, headers.Body},
					{whole.Responses[0].Body, headers.Responses[0].Body},
				} {
					require.NotNil(t, bodies[0].Find("trace_id"))
					require.Nil(t, bodies[1].Find("trace_id"))
					require.NotSame(t, bodies[0].Type.(expr.UserType).Origin(), bodies[1].Type.(expr.UserType).Origin())
					left := bodies[0].Find("item").Type.(expr.UserType)
					right := bodies[1].Find("item").Type.(expr.UserType)
					require.Same(t, left.Origin(), right.Origin())
					require.NotSame(t, left.Attribute(), right.Attribute())
					require.Same(t, left.Origin(), expr.AsArray(left.Attribute().Find("children").Type).ElemType.Type.(expr.UserType).Origin())
					require.NotSame(t, left.Origin(), bodies[0].Find("derived").Type.(expr.UserType).Origin())
					if prior, ok := origins[service]; ok {
						require.Same(t, prior, left.Origin())
					}
					origins[service] = left.Origin()
				}
			}
			require.Same(t, item, origins["plain"])
			require.NotSame(t, item, origins["first"])
			require.NotSame(t, origins["first"], origins["second"])
			require.NotNil(t, item.Attribute().Find("display_name"))
			require.NotNil(t, derived.Attribute().Find("display_name"))
		})
	}
}
