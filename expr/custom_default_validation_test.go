// Custom field defaults are schema values even when their Go names differ.
// These tests check original numeric values and preserve authored byte values.
package expr_test

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type (
	customDefaultInt          int64
	customDefaultUint         uint64
	customDefaultFloat        float64
	customDefaultString       string
	customDefaultBool         bool
	customDefaultBytes[T any] []byte
)

func TestCustomDefaultSchemaValues(t *testing.T) {
	for _, test := range []struct {
		name      string
		primitive expr.Primitive
		value     any
		rules     func()
		want      string
	}{
		{"bool", Boolean, customDefaultBool(true), nil, ""},
		{"wrong bool kind", Boolean, customDefaultString("true"), nil, "expected boolean"},
		{"int32 maximum", Int32, customDefaultInt(math.MaxInt32), nil, ""},
		{"int32 minimum", Int32, customDefaultInt(math.MinInt32), nil, ""},
		{"int32 above", Int32, customDefaultInt(math.MaxInt32 + 1), nil, "outside the range"},
		{"int32 below", Int32, customDefaultInt(math.MinInt32 - 1), nil, "outside the range"},
		{"int32 truncation", Int32, customDefaultInt(4294967296), nil, "outside the range"},
		{"int64 maximum", Int64, customDefaultInt(math.MaxInt64), nil, ""},
		{"int64 minimum", Int64, customDefaultInt(math.MinInt64), nil, ""},
		{"unsigned to signed overflow", Int64, customDefaultUint(math.MaxUint64), nil, "outside the range"},
		{"uint32 maximum", UInt32, customDefaultUint(math.MaxUint32), nil, ""},
		{"uint32 zero", UInt32, customDefaultInt(0), nil, ""},
		{"uint32 above", UInt32, customDefaultUint(math.MaxUint32 + 1), nil, "outside the range"},
		{"negative unsigned", UInt32, customDefaultInt(-1), nil, "outside the range"},
		{"uint64 maximum", UInt64, customDefaultUint(math.MaxUint64), nil, ""},
		{"integer fraction", Int32, customDefaultFloat(0.5), nil, "expected int32"},
		{"integer whole float", Int32, customDefaultFloat(1), nil, "expected int32"},
		{"float32 maximum", Float32, customDefaultFloat(math.MaxFloat32), nil, ""},
		{"float32 overflow", Float32, customDefaultFloat(math.Nextafter(math.MaxFloat32, math.Inf(1))), nil, "outside the range"},
		{"float32 rounding", Float32, customDefaultFloat(1.2), func() { Enum(1.2) }, ""},
		{"float32 underflow", Float32, customDefaultFloat(math.SmallestNonzeroFloat64), nil, ""},
		{"range before rounding", Float32, customDefaultFloat(1.00000001), func() { Maximum(1) }, "must be at most 1"},
		{"nan", Float64, customDefaultFloat(math.NaN()), nil, "outside the range"},
		{"positive infinity", Float64, customDefaultFloat(math.Inf(1)), nil, "outside the range"},
		{"negative infinity", Float64, customDefaultFloat(math.Inf(-1)), nil, "outside the range"},
		{"integer to float", Float64, customDefaultInt(7), func() { Enum(7) }, ""},
		{"numeric enum", Int32, customDefaultInt(7), func() { Enum(7) }, ""},
		{"wrong numeric enum", Int32, customDefaultInt(8), func() { Enum(7) }, "must be one of"},
		{"enum cannot wrap", UInt32, customDefaultUint(0), func() { Enum(int(4294967296)) }, "must be one of"},
		{"named string", String, customDefaultString("ready"), func() { Enum("ready") }, ""},
		{"integer cannot become rune", String, customDefaultInt(65), nil, "expected string"},
		{"raw string", String, json.RawMessage("true"), func() { Pattern("^true$") }, ""},
		{"plain bytes string", String, []byte("true"), func() { Enum("true") }, ""},
		{"raw bytes", Bytes, json.RawMessage("true"), func() { MinLength(4) }, ""},
		{"generic bytes", Bytes, customDefaultBytes[int]("true"), nil, ""},
		{"named text bytes", Bytes, customDefaultString("é"), func() { MinLength(2) }, ""},
		{"wrong slice", Bytes, []int{1}, nil, "expected bytes"},
		{"raw string pattern", String, json.RawMessage("false"), func() { Pattern("^true$") }, "does not match pattern"},
		{"raw string format", String, json.RawMessage("bad-email"), func() { Format(FormatEmail) }, "does not match format"},
		{"raw string length", String, json.RawMessage("é"), func() { MinLength(2) }, "length must be at least 2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var attribute expr.UserType
			design := func() {
				attribute = Type("CustomDefault", test.primitive, func() {
					Meta("struct:field:type", "Target")
					if test.rules != nil {
						test.rules()
					}
					Default(test.value)
				})
			}
			if test.want == "" {
				expr.RunDSL(t, design)
				require.Equal(t, test.value, attribute.Attribute().DefaultValue)
				require.Equal(t, reflect.TypeOf(test.value), reflect.TypeOf(attribute.Attribute().DefaultValue))
			} else {
				err := expr.RunInvalidDSL(t, design)
				require.ErrorContains(t, err, test.want)
			}
		})
	}
}

func TestCustomDefaultDoesNotExpandOrdinaryEligibility(t *testing.T) {
	for _, test := range []struct {
		primitive expr.Primitive
		value     any
	}{
		{Int32, int64(1)},
		{String, []byte("true")},
		{Int64, customDefaultInt(1)},
	} {
		err := expr.RunInvalidDSL(t, func() {
			Type("Ordinary", test.primitive, func() { Default(test.value) })
		})
		require.ErrorContains(t, err, "default value has type")
	}
}

func TestCustomDefaultNestedValidation(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		Type("NestedDefaults", func() {
			Attribute("items", ArrayOf(Int32, func() {
				Meta("struct:field:type", "Counter")
				Minimum(1)
			}))
			Attribute("labels", MapOf(String, String, func() {
				Key(func() {
					Meta("struct:field:type", "Label")
					Pattern("^valid$")
				})
				Elem(func() {
					Meta("struct:field:type", "Text")
					MinLength(2)
				})
			}))
			OneOf("choice", func() {
				Attribute("text", String, func() {
					Meta("struct:field:type", "Text")
					Enum("ready")
				})
			})
			Attribute("required", String)
			Required("required")
			Default(map[string]any{
				"items":   []customDefaultInt{0, 4294967296},
				"labels":  map[customDefaultString]customDefaultString{"bad": "x"},
				"choice":  map[string]any{"type": "text", "value": customDefaultString("bad")},
				"unknown": true,
			})
		})
	})
	for _, want := range []string{
		`element 0 must be at least 1`, `element 1 value 4294967296 is outside the range`,
		`map key does not match pattern`, `map value for "bad" length must be at least 2`,
		`OneOf branch "text" must be one of`, `is missing required field "required"`,
		`contains unknown field "unknown"`,
	} {
		require.ErrorContains(t, err, want)
	}
}
