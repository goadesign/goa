// This file verifies that authored defaults satisfy the complete design
// contract before generators receive them.
package expr_test

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type (
	structDefault struct {
		Profile structDefaultProfile
	}

	structDefaultProfile struct {
		Name            string
		Limit           int
		CustomProfileID string
	}

	invalidStructDefault struct {
		Profile invalidStructDefaultProfile
	}

	invalidStructDefaultProfile struct {
		Name  string
		Limit int
		Count string
	}
)

func TestValidDefaultValues(t *testing.T) {
	for _, test := range []struct {
		name string
		dsl  func()
	}{
		{
			name: "nested object array map and union",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("profile", func() {
						Attribute("name", String, func() {
							Pattern("^[a-z]+$")
						})
						Attribute("scores", ArrayOf(Int, func() {
							Minimum(1)
						}))
						Attribute("limits", MapOf(String, Int, func() {
							Key(func() {
								MinLength(2)
							})
							Elem(func() {
								Maximum(10)
							})
						}))
						OneOf("state", func() {
							Attribute("active", func() {
								Attribute("label", String, func() {
									MinLength(2)
								})
								Required("label")
							})
							Attribute("inactive", Empty)
						})
						Required("name", "scores", "limits", "state")
					})
					Default(map[string]any{
						"profile": map[string]any{
							"name":   "valid",
							"scores": []int{1, 2},
							"limits": map[string]int{"ok": 10},
							"state": map[string]any{
								"type":  "inactive",
								"value": map[string]any{},
							},
						},
					})
					Required("profile")
				})
			},
		},
		{
			name: "nested Go struct",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("profile", func() {
						Attribute("name", String, func() {
							Pattern("^[a-z]+$")
						})
						Attribute("limit", Int, func() {
							Minimum(1)
						})
						Attribute("profile_id", String, func() {
							Meta("struct:field:name", "custom_profile_id")
						})
						Required("name", "limit", "profile_id")
					})
					Default(structDefault{
						Profile: structDefaultProfile{
							Name:            "valid",
							Limit:           2,
							CustomProfileID: "profile-1",
						},
					})
					Required("profile")
				})
			},
		},
		{
			name: "native constant for custom primitive field",
			dsl: func() {
				Type("ErrorHandling", Int64, func() {
					Meta("struct:field:type", "flag.ErrorHandling", "flag")
					Default(1)
				})
			},
		},
		{
			name: "numeric enum uses primitive values",
			dsl: func() {
				Type("Score", Float32, func() {
					Enum(1.2, 5, 10, 100.8)
					Default(5.0)
				})
			},
		},
		{
			name: "typed custom primitive value",
			dsl: func() {
				Type("RawMessage", String, func() {
					Meta("struct:field:type", "json.RawMessage", "encoding/json", "json")
					Default(json.RawMessage("true"))
				})
			},
		},
		{
			name: "canonical Any value",
			dsl: func() {
				Type("Metadata", Any, func() {
					Default(map[string]any{
						"enabled": true,
						"count":   2,
						"ratio":   1.5,
						"name":    "valid",
						"items":   []any{"first", nil},
					})
				})
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			expr.RunDSL(t, test.dsl)
		})
	}
}

func TestInvalidNestedStructDefault(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		Type("Defaults", func() {
			Attribute("profile", func() {
				Attribute("name", String, func() {
					Pattern("^[a-z]+$")
				})
				Attribute("limit", Int, func() {
					Minimum(1)
				})
				Attribute("count", Int)
				Required("name", "limit", "count")
			})
			Default(invalidStructDefault{
				Profile: invalidStructDefaultProfile{
					Name:  "INVALID",
					Limit: 0,
					Count: "many",
				},
			})
			Required("profile")
		})
	})
	require.ErrorContains(t, err, `default value field "profile" field "name" does not match pattern`)
	require.ErrorContains(t, err, `default value field "profile" field "limit" must be at least 1`)
	require.ErrorContains(t, err, `default value field "profile" field "count" has type string, expected int`)
}

func TestDeclaredDefaultReportsOnceWhenTypeIsReferenced(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		defaults := Type("Defaults", func() {
			Attribute("name", String, func() {
				Pattern("^valid$")
				Default("invalid")
			})
		})
		Service("Service", func() {
			Method("First", func() {
				Payload(defaults)
			})
			Method("Second", func() {
				Result(defaults)
			})
		})
	})
	require.Equal(t, 1, strings.Count(err.Error(), "default value does not match pattern"))
}

func TestErrorDefaultReportsOnceWhenInherited(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		Service("Service", func() {
			Error("bad_request", String, func() {
				Pattern("^valid$")
				Default("invalid")
			})
			Method("Method", func() {
				Error("bad_request")
			})
		})
	})
	require.Equal(t, 1, strings.Count(err.Error(), "default value does not match pattern"))
}

func TestInvalidDefaultValues(t *testing.T) {
	for _, test := range []struct {
		name string
		dsl  func()
		want string
	}{
		{
			name: "missing required nested field",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("profile", func() {
						Attribute("name", String)
						Required("name")
					})
					Default(map[string]any{"profile": map[string]any{}})
				})
			},
			want: `default value field "profile" is missing required field "name"`,
		},
		{
			name: "nested pattern",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("profile", func() {
						Attribute("name", String, func() {
							Pattern("^[a-z]+$")
						})
					})
					Default(map[string]any{"profile": map[string]any{"name": "INVALID"}})
				})
			},
			want: `default value field "profile" field "name" does not match pattern`,
		},
		{
			name: "nested enum",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("profile", func() {
						Attribute("mode", String, func() {
							Enum("safe", "fast")
						})
					})
					Default(map[string]any{"profile": map[string]any{"mode": "unknown"}})
				})
			},
			want: `default value field "profile" field "mode" must be one of`,
		},
		{
			name: "nested format",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("profile", func() {
						Attribute("email", String, func() {
							Format(FormatEmail)
						})
					})
					Default(map[string]any{"profile": map[string]any{"email": "not-an-email"}})
				})
			},
			want: `default value field "profile" field "email" does not match format "email"`,
		},
		{
			name: "wrong nested type",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("profile", func() {
						Attribute("count", Int)
					})
					Default(map[string]any{"profile": map[string]any{"count": "many"}})
				})
			},
			want: `default value field "profile" field "count" has type string, expected int`,
		},
		{
			name: "negative unsigned value",
			dsl: func() {
				Type("Defaults", UInt, func() {
					Default(-1)
				})
			},
			want: "default value value -1 is outside the range of uint",
		},
		{
			name: "integer overflow",
			dsl: func() {
				Type("Defaults", Int32, func() {
					Default(int(math.MaxInt32) + 1)
				})
			},
			want: "default value value 2147483648 is outside the range of int32",
		},
		{
			name: "object key type",
			dsl: func() {
				Type("Defaults", func() {
					Attribute("count", Int)
					Default(map[int]int{1: 2})
				})
			},
			want: "default value object key has type int, expected string",
		},
		{
			name: "array element validation",
			dsl: func() {
				Type("Defaults", ArrayOf(String, func() {
					MinLength(3)
				}), func() {
					Default([]string{"ok"})
				})
			},
			want: "default value element 0 length must be at least 3",
		},
		{
			name: "array length validation",
			dsl: func() {
				Type("Defaults", ArrayOf(String), func() {
					MaxLength(1)
					Default([]string{"first", "second"})
				})
			},
			want: "default value length must be at most 1",
		},
		{
			name: "map key validation",
			dsl: func() {
				Type("Defaults", MapOf(String, Int, func() {
					Key(func() {
						Pattern("^[a-z]+$")
					})
				}), func() {
					Default(map[string]int{"INVALID": 1})
				})
			},
			want: "default value map key does not match pattern",
		},
		{
			name: "map element validation",
			dsl: func() {
				Type("Defaults", MapOf(String, Int, func() {
					Elem(func() {
						Minimum(1)
					})
				}), func() {
					Default(map[string]int{"valid": 0})
				})
			},
			want: `default value map value for "valid" must be at least 1`,
		},
		{
			name: "unknown union branch",
			dsl: func() {
				Type("Defaults", func() {
					OneOf("state", func() {
						Attribute("active", String)
						Attribute("inactive", Empty)
						Default(map[string]any{"type": "missing", "value": "value"})
					})
				})
			},
			want: `default value selects unknown OneOf branch "missing"`,
		},
		{
			name: "wrong union branch value",
			dsl: func() {
				Type("Defaults", func() {
					OneOf("state", func() {
						Attribute("active", Int)
						Attribute("inactive", Empty)
						Default(map[string]any{"type": "active", "value": "wrong"})
					})
				})
			},
			want: `default value OneOf branch "active" has type string, expected stateActive`,
		},
		{
			name: "Any map key",
			dsl: func() {
				Type("Defaults", Any, func() {
					Default(map[int]string{1: "invalid"})
				})
			},
			want: "default value object key must be a string",
		},
		{
			name: "Any non-finite number",
			dsl: func() {
				Type("Defaults", Any, func() {
					Default(math.Inf(1))
				})
			},
			want: "default value number must be finite",
		},
		{
			name: "Any unsupported value",
			dsl: func() {
				Type("Defaults", Any, func() {
					Default(struct{ Name string }{Name: "invalid"})
				})
			},
			want: "default value has unsupported Go type",
		},
		{
			name: "untyped nil",
			dsl: func() {
				Type("Defaults", String, func() {
					Default(nil)
				})
			},
			want: "default value must not be nil",
		},
		{
			name: "service error default",
			dsl: func() {
				Service("Service", func() {
					Error("bad_request", String, func() {
						Pattern("^valid$")
						Default("invalid")
					})
					Method("Method", func() {})
				})
			},
			want: `error "bad_request" - default value does not match pattern`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, test.dsl)
			require.ErrorContains(t, err, test.want)
		})
	}
}
