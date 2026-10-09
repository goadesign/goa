// These tests run the ordinary DSL evaluation pipeline. Invalid JSON mappings
// must fail expression validation before finalization or generation can begin.
package dsl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestOneOfUntaggedEvaluation(t *testing.T) {
	for _, test := range []struct {
		name    string
		declare func()
		failure string
	}{
		{name: "array and string", declare: func() {
			Attribute("manifest", ArrayOf(String))
			Attribute("dynamic", String, func() { Enum("dynamic") })
		}},
		{name: "numeric overlap", declare: func() {
			Attribute("integer", Int64)
			Attribute("number", Float64)
		}, failure: "same JSON kind"},
		{name: "byte string overlap", declare: func() {
			Attribute("text", String)
			Attribute("binary", Bytes)
		}, failure: "same JSON kind"},
		{name: "object map overlap", declare: func() {
			Attribute("record", func() { Attribute("value", String) })
			Attribute("mapping", MapOf(String, String))
		}, failure: "same JSON kind"},
		{name: "open JSON", declare: func() {
			Attribute("anything", Any)
		}, failure: "one known JSON kind"},
		{name: "nested union", declare: func() {
			OneOf("nested", func() { Attribute("text", String) })
		}, failure: "one known JSON kind"},
		{name: "custom representation", declare: func() {
			Attribute("custom", String, func() { Meta("struct:field:type", "Custom") })
		}, failure: "one known JSON kind"},
		{name: "flatten conflict", declare: func() {
			Meta("oneof:json:flatten")
			Attribute("record", func() { Attribute("value", String) })
		}, failure: "cannot select flattening"},
		{name: "discriminator conflict", declare: func() {
			Meta("oneof:type:field", "kind")
			Attribute("text", String)
		}, failure: "cannot select flattening"},
		{name: "value conflict", declare: func() {
			Meta("oneof:value:field", "data")
			Attribute("text", String)
		}, failure: "cannot select flattening"},
		{name: "empty discriminator metadata", declare: func() {
			Meta("oneof:type:field")
			Attribute("text", String)
		}, failure: "cannot select flattening"},
		{name: "union custom representation", declare: func() {
			Meta("struct:field:type", "Custom")
			Attribute("text", String)
		}, failure: "cannot select a custom Go representation"},
		{name: "flag argument", declare: func() {
			Meta("oneof:json:untagged", "true")
			Attribute("text", String)
		}, failure: "takes no values"},
		{name: "valid raw default", declare: func() {
			Attribute("manifest", ArrayOf(String))
			Attribute("dynamic", String, func() { Enum("dynamic") })
			Default("dynamic")
			Example("dynamic")
		}},
		{name: "invalid raw default", declare: func() {
			Attribute("manifest", ArrayOf(String))
			Attribute("dynamic", String, func() { Enum("dynamic") })
			Default("other")
		}, failure: "must be one of"},
		{name: "invalid raw example", declare: func() {
			Attribute("manifest", ArrayOf(String))
			Attribute("dynamic", String, func() { Enum("dynamic") })
			Example("other")
		}, failure: "example value OneOf branch \"dynamic\" must be one of"},
		{name: "invalid integer example", declare: func() {
			Attribute("number", Int64)
			Example(1.5)
		}, failure: "has type float64"},
		{name: "null array example", declare: func() {
			Attribute("files", ArrayOfRequired(String))
			Example([]any{nil})
		}, failure: "must not be nil"},
	} {
		t.Run(test.name, func(t *testing.T) {
			eval.Reset()
			expr.Root = new(expr.RootExpr)
			expr.GeneratedResultTypes = new(expr.ResultTypesRoot)
			require.NoError(t, eval.Register(expr.Root))
			require.NoError(t, eval.Register(expr.GeneratedResultTypes))
			API("union-evaluation", func() {})
			Service("records", func() {
				Method("echo", func() {
					Payload(func() {
						OneOf("resources", func() {
							Meta("oneof:json:untagged")
							test.declare()
						})
					})
					Result(String)
				})
			})
			failure := eval.RunDSL()
			if test.failure != "" {
				require.ErrorContains(t, failure, test.failure)
				return
			}
			require.NoError(t, failure)
			attribute := expr.Root.Service("records").Method("echo").Payload.Find("resources")
			require.True(t, expr.AsUnion(attribute.Type).Untagged)
			require.True(t, expr.AsUnion(expr.DupAtt(attribute).Type).Untagged)
			require.True(t, expr.AsUnion(expr.NewAttributeGraphCopier().Copy(attribute).Type).Untagged)
		})
	}
}
