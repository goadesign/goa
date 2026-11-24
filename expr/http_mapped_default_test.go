// These tests evaluate complete designs to check that defaults written on HTTP
// parameters, headers and cookies agree with the object payload attribute
// default, including defaults the payload inherits through Extend, Reference
// and named types before Finalize resolves them.
package expr_test

import (
	"encoding/json"
	"flag"
	"math"
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
		name        string
		payload     func() any
		payloadArgs func() []any
		mapping     func()
		want        string
		forbid      string
	}{
		{
			name:    "query same default",
			payload: stringPayload("a"),
			mapping: query("a"),
		},
		{
			name:    "query default on payload only",
			payload: stringPayload("a"),
			mapping: func() {
				dsl.GET("/")
				dsl.Param("mode")
			},
		},
		{
			name:    "query default only",
			payload: stringPayload(),
			mapping: query("a"),
			want:    `query parameter "mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name:    "query zero default only",
			payload: stringPayload(),
			mapping: query(""),
			want:    `query parameter "mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name:    "query different default",
			payload: stringPayload("b"),
			mapping: query("a"),
			want:    `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"; object payloads own field defaults, so remove the Default from the query parameter or make it match.`,
		},
		{
			name:    "query renamed",
			payload: stringPayload(),
			mapping: func() {
				dsl.GET("/")
				dsl.Param("mode:m", func() {
					dsl.Default("a")
				})
			},
			want: `query parameter "m" defines a default value but payload attribute "mode" does not`,
		},
		{
			name:    "query zero string",
			payload: stringPayload(""),
			mapping: query(""),
		},
		{
			name:    "query zero int",
			payload: typedPayload(dsl.Int, 0),
			mapping: query(0),
		},
		{
			name:    "query false",
			payload: typedPayload(dsl.Boolean, false),
			mapping: query(false),
		},
		{
			name:    "query int64 representations",
			payload: typedPayload(dsl.Int64, 10),
			mapping: query(int64(10)),
		},
		{
			name:    "query float64 representations",
			payload: typedPayload(dsl.Float64, 1),
			mapping: query(1.0),
		},
		{
			name:    "query bytes representations",
			payload: typedPayload(dsl.Bytes, "a"),
			mapping: query([]byte("a")),
		},
		{
			name:    "query array representations",
			payload: typedPayload(dsl.ArrayOf(dsl.Int64), []int{1, 2}),
			mapping: query([]int64{1, 2}),
		},
		{
			name:    "query different array",
			payload: typedPayload(dsl.ArrayOf(dsl.Int64), []int{1, 2}),
			mapping: query([]int64{2, 1}),
			want:    `query parameter "mode" default value []int64{2, 1} differs from payload attribute "mode" default value []int{1, 2}`,
		},
		{
			name:    "query map representations",
			payload: typedPayload(dsl.MapOf(dsl.String, dsl.Int64), expr.MapVal{"a": 1, "b": 2}),
			mapping: query(map[string]int64{"a": 1, "b": 2}),
		},
		{
			name:    "query different map",
			payload: typedPayload(dsl.MapOf(dsl.String, dsl.Int64), map[string]int64{"a": 1}),
			mapping: query(map[string]int64{"a": 2}),
			want:    `query parameter "mode" default value map[string]int64{"a":2} differs from payload attribute "mode" default value map[string]int64{"a":1}`,
		},
		{
			name:    "query colliding map keys",
			payload: typedPayload(dsl.MapOf(dsl.Int64, dsl.String), expr.MapVal{1: "a", 2: "b"}),
			mapping: query(expr.MapVal{1: "a", int64(1): "a"}),
			want:    `differs from payload attribute "mode" default value`,
		},
		{
			name: "query custom numeric field type same constant",
			payload: func() any {
				return func() {
					dsl.Attribute("mode", dsl.Int64, func() {
						dsl.Meta("struct:field:type", "flag.ErrorHandling", "flag")
						dsl.Default(flag.ContinueOnError)
					})
				}
			},
			mapping: query(flag.ContinueOnError),
		},
		{
			name: "query custom numeric field type primitive constant",
			payload: func() any {
				return func() {
					dsl.Attribute("mode", dsl.Int64, func() {
						dsl.Meta("struct:field:type", "flag.ErrorHandling", "flag")
						dsl.Default(flag.ContinueOnError)
					})
				}
			},
			mapping: query(0),
		},
		{
			name: "query custom numeric field type different constant",
			payload: func() any {
				return func() {
					dsl.Attribute("mode", dsl.Int64, func() {
						dsl.Meta("struct:field:type", "flag.ErrorHandling", "flag")
						dsl.Default(flag.ExitOnError)
					})
				}
			},
			mapping: query(flag.ContinueOnError),
			want:    `query parameter "mode" default value 0 differs from payload attribute "mode" default value 1`,
		},
		{
			name: "query custom string named type default",
			payload: func() any {
				raw := dsl.Type("RawMessage", dsl.String, func() {
					dsl.Meta("struct:field:type", "json.RawMessage", "encoding/json")
					dsl.Default(json.RawMessage("true"))
				})
				return func() {
					dsl.Attribute("mode", raw)
				}
			},
			mapping: query("true"),
		},
		{
			name: "query custom string named type different default",
			payload: func() any {
				raw := dsl.Type("RawMessage", dsl.String, func() {
					dsl.Meta("struct:field:type", "json.RawMessage", "encoding/json")
					dsl.Default(json.RawMessage("true"))
				})
				return func() {
					dsl.Attribute("mode", raw)
				}
			},
			mapping: query("false"),
			want:    `differs from payload attribute "mode" default value`,
		},
		{
			name:    "query invalid out-of-range default reports range error only",
			payload: typedPayload(dsl.Int32, int(math.MaxInt32)+1),
			mapping: query(int(math.MaxInt32) + 1),
			want:    `outside the range of int32`,
			forbid:  `differs from payload attribute`,
		},
		{
			name:    "path same default",
			payload: stringPayload("a"),
			mapping: func() {
				dsl.GET("/{mode}")
				dsl.Param("mode", func() {
					dsl.Default("a")
				})
			},
		},
		{
			name:    "path different default",
			payload: stringPayload("b"),
			mapping: func() {
				dsl.GET("/{mode}")
				dsl.Param("mode", func() {
					dsl.Default("a")
				})
			},
			want: `path parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`,
		},
		{
			name:    "header same default",
			payload: stringPayload("a"),
			mapping: func() {
				dsl.GET("/")
				dsl.Header("mode:X-Mode", func() {
					dsl.Default("a")
				})
			},
		},
		{
			name:    "header default only",
			payload: stringPayload(),
			mapping: func() {
				dsl.GET("/")
				dsl.Header("mode:X-Mode", func() {
					dsl.Default("a")
				})
			},
			want: `header "X-Mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name:    "cookie same default",
			payload: stringPayload("a"),
			mapping: func() {
				dsl.GET("/")
				dsl.Cookie("mode", func() {
					dsl.Default("a")
				})
			},
		},
		{
			name:    "cookie different default",
			payload: stringPayload("b"),
			mapping: func() {
				dsl.GET("/")
				dsl.Cookie("mode", func() {
					dsl.Default("a")
				})
			},
			want: `cookie "mode" default value "a" differs from payload attribute "mode" default value "b"`,
		},
		{
			name: "named type extend same default",
			payload: func() any {
				base := baseType("a")
				return dsl.Type("Payload", func() {
					dsl.Extend(base)
				})
			},
			mapping: query("a"),
		},
		{
			name: "inline extend same default",
			payload: func() any {
				base := baseType("a")
				return func() {
					dsl.Extend(base)
				}
			},
			mapping: query("a"),
		},
		{
			name: "named type extend different default",
			payload: func() any {
				base := baseType("b")
				return dsl.Type("Payload", func() {
					dsl.Extend(base)
				})
			},
			mapping: query("a"),
			want:    `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`,
		},
		{
			name: "inline extend different default",
			payload: func() any {
				base := baseType("b")
				return func() {
					dsl.Extend(base)
				}
			},
			mapping: query("a"),
			want:    `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`,
		},
		{
			name: "extend without default",
			payload: func() any {
				base := baseType()
				return func() {
					dsl.Extend(base)
				}
			},
			mapping: query("a"),
			want:    `query parameter "mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name: "extend replaces payload attribute",
			payload: func() any {
				base := baseType("b")
				return func() {
					dsl.Attribute("mode", dsl.String, func() {
						dsl.Default("a")
					})
					dsl.Extend(base)
				}
			},
			mapping: query("a"),
			want:    `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`,
		},
		{
			name: "reference after attribute same default",
			payload: func() any {
				base := baseType("a")
				return dsl.Type("Payload", func() {
					dsl.Attribute("mode")
					dsl.Reference(base)
				})
			},
			mapping: query("a"),
		},
		{
			name: "reference before attribute same default",
			payload: func() any {
				base := baseType("a")
				return dsl.Type("Payload", func() {
					dsl.Reference(base)
					dsl.Attribute("mode")
				})
			},
			mapping: query("a"),
		},
		{
			name: "reference different default",
			payload: func() any {
				base := baseType("b")
				return dsl.Type("Payload", func() {
					dsl.Attribute("mode")
					dsl.Reference(base)
				})
			},
			mapping: query("a"),
			want:    `query parameter "mode" default value "a" differs from payload attribute "mode" default value "b"`,
		},
		{
			name: "reference to earlier extended type inherits default",
			payload: func() any {
				ancestor := baseType("a")
				shared := dsl.Type("SharedBase", func() {
					dsl.Extend(ancestor)
				})
				return dsl.Type("Payload", func() {
					dsl.Attribute("mode", dsl.String)
					dsl.Reference(shared)
				})
			},
			mapping: query("a"),
		},
		{
			name: "customized payload dup does not resolve unfinalized reference base",
			payload: func() any {
				ancestor := baseType("a")
				shared := dsl.Type("SharedBase", func() {
					dsl.Extend(ancestor)
				})
				return dsl.Type("Payload", func() {
					dsl.Attribute("mode", dsl.String)
					dsl.Reference(shared)
				})
			},
			payloadArgs: func() []any {
				return []any{"custom description"}
			},
			mapping: query("a"),
			want:    `query parameter "mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name: "type referencing extended result type does not inherit unfinalized default",
			payload: func() any {
				ancestor := baseType("a")
				shared := dsl.ResultType("application/vnd.goa.shared", func() {
					dsl.Extend(ancestor)
				})
				return dsl.Type("Payload", func() {
					dsl.Attribute("mode", dsl.String)
					dsl.Reference(shared)
				})
			},
			mapping: query("a"),
			want:    `query parameter "mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name: "forward-declared type referencing later extended type does not inherit unfinalized default",
			payload: func() any {
				var shared expr.UserType
				payload := dsl.Type("Payload", func() {
					dsl.Attribute("mode", dsl.String)
					dsl.Reference(shared)
				})
				ancestor := baseType("a")
				shared = dsl.Type("SharedBase", func() {
					dsl.Extend(ancestor)
				})
				return payload
			},
			mapping: query("a"),
			want:    `query parameter "mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name: "named scalar type default",
			payload: func() any {
				mode := dsl.Type("Mode", dsl.String, func() {
					dsl.Default("a")
				})
				return func() {
					dsl.Attribute("mode", mode)
				}
			},
			mapping: query("a"),
		},
		{
			name: "reference after untyped attribute ignores reference scalar type default",
			payload: func() any {
				mode := dsl.Type("Mode", dsl.String, func() {
					dsl.Default("a")
				})
				ref := dsl.Type("Ref", func() {
					dsl.Attribute("mode", mode)
				})
				return dsl.Type("Payload", func() {
					dsl.Attribute("mode")
					dsl.Reference(ref)
				})
			},
			mapping: query("a"),
			want:    `query parameter "mode" defines a default value but payload attribute "mode" does not`,
		},
		{
			name: "payload reference overrides field scalar type default",
			payload: func() any {
				mode := dsl.Type("Mode", dsl.String, func() {
					dsl.Default("from-type")
				})
				return dsl.Type("Sub", func() {
					dsl.Attribute("mode", mode)
				})
			},
			payloadArgs: func() []any {
				ref := baseType("from-ref")
				return []any{func() {
					dsl.Reference(ref)
				}}
			},
			mapping: query("from-ref"),
		},
		{
			name: "payload reference overrides field scalar type default mismatch",
			payload: func() any {
				mode := dsl.Type("Mode", dsl.String, func() {
					dsl.Default("from-type")
				})
				return dsl.Type("Sub", func() {
					dsl.Attribute("mode", mode)
				})
			},
			payloadArgs: func() []any {
				ref := baseType("from-ref")
				return []any{func() {
					dsl.Reference(ref)
				}}
			},
			mapping: query("from-type"),
			want:    `query parameter "mode" default value "from-type" differs from payload attribute "mode" default value "from-ref"`,
		},
		{
			name: "reference after attribute with custom field type default",
			payload: func() any {
				ref := dsl.Type("Ref", func() {
					dsl.Attribute("mode", dsl.Int64, func() {
						dsl.Meta("struct:field:type", "flag.ErrorHandling", "flag")
						dsl.Default(flag.ContinueOnError)
					})
				})
				return dsl.Type("Payload", func() {
					dsl.Attribute("mode", dsl.Int64)
					dsl.Reference(ref)
				})
			},
			mapping: query(0),
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			design := func() {
				payload := test.payload()
				var args []any
				if test.payloadArgs != nil {
					args = test.payloadArgs()
				}
				dsl.Service("s", func() {
					dsl.Method("m", func() {
						dsl.Payload(payload, args...)
						dsl.HTTP(test.mapping)
					})
				})
			}
			if test.want == "" {
				root := expr.RunDSL(t, design)
				require.True(t, root.Service("s").Method("m").Payload.HasDefaultValue("mode"))
				return
			}
			err := expr.RunInvalidDSL(t, design)
			require.ErrorContains(t, err, test.want)
			if test.forbid != "" {
				require.NotContains(t, err.Error(), test.forbid)
			}
		})
	}
}

func TestHTTPMappedDefaultsChildBeforeParentCanonicalReference(t *testing.T) {
	root := expr.RunDSL(t, func() {
		parentRef := dsl.Type("ParentRef", func() {
			dsl.Attribute("parent_id", dsl.String, func() {
				dsl.Default("acme")
			})
		})
		dsl.Service("Child", func() {
			dsl.HTTP(func() {
				dsl.Path("/children")
				dsl.Parent("Parent")
			})
			dsl.Method("list", func() {
				dsl.Payload(func() {
					dsl.Attribute("limit", dsl.Int)
				})
				dsl.HTTP(func() {
					dsl.GET("")
				})
			})
		})
		dsl.Service("Parent", func() {
			dsl.Method("show", func() {
				dsl.Payload(func() {
					dsl.Attribute("parent_id", dsl.String)
					dsl.Reference(parentRef)
				})
				dsl.HTTP(func() {
					dsl.GET("/parents/{parent_id}")
					dsl.Param("parent_id", func() {
						dsl.Default("acme")
					})
				})
			})
		})
	})
	require.Equal(t, "acme", root.Service("Child").Method("list").Payload.GetDefault("parent_id"))
	require.Equal(t, "acme", root.Service("Parent").Method("show").Payload.GetDefault("parent_id"))
}
