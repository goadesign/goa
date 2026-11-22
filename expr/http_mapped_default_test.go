// These tests evaluate complete designs to check that defaults written on HTTP
// parameters, headers and cookies agree with the object payload attribute
// default, including defaults the payload inherits through Extend, Reference
// and named types before Finalize resolves them.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestHTTPMappedDefaults(t *testing.T) {
	stringPayload := func(def ...any) func() any {
		return func() any {
			return func() {
				dsl.Attribute("mode", dsl.String, func() {
					if len(def) > 0 {
						dsl.Default(def[0])
					}
				})
			}
		}
	}
	typedPayload := func(dataType expr.DataType, def any) func() any {
		return func() any {
			return func() {
				dsl.Attribute("mode", dataType, func() {
					dsl.Default(def)
				})
			}
		}
	}
	baseType := func(def ...any) expr.UserType {
		return dsl.Type("Base", func() {
			dsl.Attribute("mode", dsl.String, func() {
				if len(def) > 0 {
					dsl.Default(def[0])
				}
			})
		})
	}
	query := func(def any) func() {
		return func() {
			dsl.GET("/")
			dsl.Param("mode", func() {
				dsl.Default(def)
			})
		}
	}
	cases := []struct {
		name    string
		payload func() any
		mapping func()
		want    string
	}{
		{name: "query same default", payload: stringPayload("a"), mapping: query("a")},
		{name: "query default on payload only", payload: stringPayload("a"), mapping: func() {
			dsl.GET("/")
			dsl.Param("mode")
		}},
		{name: "query default only", payload: stringPayload(), mapping: query("a"), want: `query parameter "mode" defines a default value but payload attribute "mode" does not`},
		{name: "query zero default only", payload: stringPayload(), mapping: query(""), want: `query parameter "mode" defines a default value but payload attribute "mode" does not`},
		{name: "query different default", payload: stringPayload("b"), mapping: query("a"), want: `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"; object payloads own field defaults, so remove the Default from the query parameter or make it match.`},
		{name: "query renamed", payload: stringPayload(), mapping: func() {
			dsl.GET("/")
			dsl.Param("mode:m", func() {
				dsl.Default("a")
			})
		}, want: `query parameter "m" defines a default value but payload attribute "mode" does not`},
		{name: "query zero string", payload: stringPayload(""), mapping: query("")},
		{name: "query zero int", payload: typedPayload(dsl.Int, 0), mapping: query(0)},
		{name: "query false", payload: typedPayload(dsl.Boolean, false), mapping: query(false)},
		{name: "query int64 representations", payload: typedPayload(dsl.Int64, 10), mapping: query(int64(10))},
		{name: "query float64 representations", payload: typedPayload(dsl.Float64, 1), mapping: query(1.0)},
		{name: "query bytes representations", payload: typedPayload(dsl.Bytes, "a"), mapping: query([]byte("a"))},
		{name: "query array representations", payload: typedPayload(dsl.ArrayOf(dsl.Int64), []int{1, 2}), mapping: query([]int64{1, 2})},
		{name: "query different array", payload: typedPayload(dsl.ArrayOf(dsl.Int64), []int{1, 2}), mapping: query([]int64{2, 1}), want: `query parameter "mode" default value []int64{2, 1} differs from payload attribute "mode" default value []int{1, 2}`},
		{name: "path same default", payload: stringPayload("a"), mapping: func() {
			dsl.GET("/{mode}")
			dsl.Param("mode", func() {
				dsl.Default("a")
			})
		}},
		{name: "path different default", payload: stringPayload("b"), mapping: func() {
			dsl.GET("/{mode}")
			dsl.Param("mode", func() {
				dsl.Default("a")
			})
		}, want: `path parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`},
		{name: "header same default", payload: stringPayload("a"), mapping: func() {
			dsl.GET("/")
			dsl.Header("mode:X-Mode", func() {
				dsl.Default("a")
			})
		}},
		{name: "header default only", payload: stringPayload(), mapping: func() {
			dsl.GET("/")
			dsl.Header("mode:X-Mode", func() {
				dsl.Default("a")
			})
		}, want: `header "X-Mode" defines a default value but payload attribute "mode" does not`},
		{name: "cookie same default", payload: stringPayload("a"), mapping: func() {
			dsl.GET("/")
			dsl.Cookie("mode", func() {
				dsl.Default("a")
			})
		}},
		{name: "cookie different default", payload: stringPayload("b"), mapping: func() {
			dsl.GET("/")
			dsl.Cookie("mode", func() {
				dsl.Default("a")
			})
		}, want: `cookie "mode" default value "a" differs from payload attribute "mode" default value "b"`},
		{name: "named type extend same default", payload: func() any {
			base := baseType("a")
			return dsl.Type("Payload", func() {
				dsl.Extend(base)
			})
		}, mapping: query("a")},
		{name: "inline extend same default", payload: func() any {
			base := baseType("a")
			return func() {
				dsl.Extend(base)
			}
		}, mapping: query("a")},
		{name: "named type extend different default", payload: func() any {
			base := baseType("b")
			return dsl.Type("Payload", func() {
				dsl.Extend(base)
			})
		}, mapping: query("a"), want: `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`},
		{name: "inline extend different default", payload: func() any {
			base := baseType("b")
			return func() {
				dsl.Extend(base)
			}
		}, mapping: query("a"), want: `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`},
		{name: "extend without default", payload: func() any {
			base := baseType()
			return func() {
				dsl.Extend(base)
			}
		}, mapping: query("a"), want: `query parameter "mode" defines a default value but payload attribute "mode" does not`},
		{name: "extend replaces payload attribute", payload: func() any {
			base := baseType("b")
			return func() {
				dsl.Attribute("mode", dsl.String, func() {
					dsl.Default("a")
				})
				dsl.Extend(base)
			}
		}, mapping: query("a"), want: `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`},
		{name: "reference after attribute same default", payload: func() any {
			base := baseType("a")
			return dsl.Type("Payload", func() {
				dsl.Attribute("mode")
				dsl.Reference(base)
			})
		}, mapping: query("a")},
		{name: "reference before attribute same default", payload: func() any {
			base := baseType("a")
			return dsl.Type("Payload", func() {
				dsl.Reference(base)
				dsl.Attribute("mode")
			})
		}, mapping: query("a")},
		{name: "reference different default", payload: func() any {
			base := baseType("b")
			return dsl.Type("Payload", func() {
				dsl.Attribute("mode")
				dsl.Reference(base)
			})
		}, mapping: query("a"), want: `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`},
		{name: "named scalar type default", payload: func() any {
			mode := dsl.Type("Mode", dsl.String, func() {
				dsl.Default("a")
			})
			return func() {
				dsl.Attribute("mode", mode)
			}
		}, mapping: query("a")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			design := func() {
				payload := test.payload()
				dsl.Service("s", func() {
					dsl.Method("m", func() {
						dsl.Payload(payload)
						dsl.HTTP(test.mapping)
					})
				})
			}
			if test.want == "" {
				expr.RunDSL(t, design)
				return
			}
			require.ErrorContains(t, expr.RunInvalidDSL(t, design), test.want)
		})
	}
}
