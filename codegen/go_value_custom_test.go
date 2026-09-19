// This file checks the public renderer with values independent of stored
// defaults, then compiles and evaluates its custom primitive expressions.
package codegen

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

type (
	renderTestInt          int64
	renderTestUint         uint64
	renderTestFloat        float64
	renderTestText         string
	renderTestBool         bool
	renderTestBytes[T any] []byte
	renderTestObject       map[string]any
)

func TestRenderGoValueCustomBytesUseLinkedImport(t *testing.T) {
	attribute := &expr.AttributeExpr{
		Type: expr.String,
		Meta: expr.MetaExpr{"struct:field:type": {"json.RawMessage", "encoding/json", "json"}},
	}
	plan, err := PlanGoType(attribute, GoTypePlanOptions{Owner: "generated.local/gen/service"})
	require.NoError(t, err)
	layout := plan.Link("generated.local/gen/service", func(importPath string) string {
		require.Equal(t, "encoding/json", importPath)
		return "encoded"
	})
	code, err := RenderGoValue(attribute, json.RawMessage("true"), layout, false, nil, "")
	require.NoError(t, err)
	require.Equal(t, "encoded.RawMessage([]byte{0x74, 0x72, 0x75, 0x65})", code.Expression)
	compileTransformSource(t, "package transformtest\nimport encoded \"encoding/json\"\nvar value = "+code.Expression+"\n")
}

func TestRenderGoValueChecksSeparateCustomValue(t *testing.T) {
	for _, stored := range []any{nil, true} {
		attribute := &expr.AttributeExpr{
			Type: expr.Boolean, DefaultValue: stored,
			Meta: expr.MetaExpr{"struct:field:type": {"Enabled"}},
		}
		layout := goValueTestLayout(t, attribute, GoLayoutPolicy{}, nil)
		code, err := RenderGoValue(attribute, renderTestText("x"), layout, false, nil, "")
		require.ErrorContains(t, err, "boolean default has Go type")
		require.Empty(t, code)
		require.Equal(t, stored, attribute.DefaultValue)
		require.Equal(t, "Enabled", layout.Def())

		code, err = RenderGoValue(attribute, renderTestBool(true), layout, false, nil, "")
		require.NoError(t, err)
		require.Equal(t, "Enabled(true)", code.Expression)
	}
}

func TestRenderGoValueChecksNestedCustomValue(t *testing.T) {
	leaf := &expr.AttributeExpr{
		Type: expr.Boolean,
		Meta: expr.MetaExpr{"struct:field:type": {"Enabled"}},
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "values", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: leaf}}},
	}}
	layout := goValueTestLayout(t, attribute, GoLayoutPolicy{}, nil)
	code, err := RenderGoValue(attribute, map[string]any{
		"values": []renderTestText{"x"},
	}, layout, false, nil, "")
	require.ErrorContains(t, err, "array element 0")
	require.ErrorContains(t, err, "boolean default has Go type")
	require.Empty(t, code)
}

func TestRenderGoValueCustomLiteralErrors(t *testing.T) {
	for _, test := range []struct {
		name      string
		primitive expr.Primitive
		value     any
		custom    string
		want      string
	}{
		{"fraction", expr.Int32, renderTestFloat(0.5), "Count", "integer default"},
		{"whole float", expr.Int32, renderTestFloat(1), "Count", "integer default"},
		{"negative unsigned", expr.UInt64, renderTestInt(-1), "Count", "must not be negative"},
		{"nan", expr.Float64, renderTestFloat(math.NaN()), "Number", "must be finite"},
		{"infinity", expr.Float32, renderTestFloat(math.Inf(1)), "Number", "must be finite"},
		{"string coercion", expr.String, renderTestInt(65), "Text", "string default"},
		{"nonbyte slice", expr.String, []int{1}, "Text", "string default"},
		{"ordinary bytes string", expr.String, []byte("true"), "", "string default"},
		{"custom Any container", expr.Any, renderTestObject{"value": true}, "Object", "unsupported underlying type map"},
	} {
		t.Run(test.name, func(t *testing.T) {
			attribute := &expr.AttributeExpr{Type: test.primitive}
			if test.custom != "" {
				attribute.Meta = expr.MetaExpr{"struct:field:type": {test.custom}}
			}
			layout := goValueTestLayout(t, attribute, GoLayoutPolicy{}, nil)
			code, err := RenderGoValue(attribute, test.value, layout, false, nil, "")
			require.ErrorContains(t, err, test.want)
			require.Empty(t, code)
		})
	}
}

func TestRenderedCustomPrimitiveValuesCompileAndPreserveValues(t *testing.T) {
	// The generated test executes each expression. Merely comparing source
	// text would miss a target conversion that changes or cannot accept a value.
	var assertions strings.Builder
	for _, test := range []struct {
		name      string
		primitive expr.Primitive
		value     any
		custom    string
		want      string
	}{
		{"boolean", expr.Boolean, renderTestBool(true), "Enabled", "Enabled(true)"},
		{"wide direct value", expr.Int32, renderTestInt(4294967296), "Wide", "Wide(4294967296)"},
		{"integer digits", expr.Float64, renderTestInt(9007199254740993), "Wide", "Wide(9007199254740993)"},
		{"uint digits", expr.UInt64, renderTestUint(math.MaxUint64), "Unsigned", "Unsigned(18446744073709551615)"},
		{"source float precision", expr.Float32, renderTestFloat(1.0000000000000002), "Number", "Number(1.0000000000000002)"},
		{"native float precision", expr.Float32, 1.0000000000000002, "Number", "1"},
		{"native custom int", expr.Int64, 1, "Wide", "1"},
		{"named text", expr.String, renderTestText("ready"), "Text", `Text("ready")`},
		{"plain bytes", expr.String, []byte("true"), "json.RawMessage", "json.RawMessage([]byte{0x74, 0x72, 0x75, 0x65})"},
		{"named raw string", expr.String, json.RawMessage("true"), "json.RawMessage", "json.RawMessage([]byte{0x74, 0x72, 0x75, 0x65})"},
		{"named raw bytes", expr.Bytes, json.RawMessage("true"), "json.RawMessage", "json.RawMessage([]byte{0x74, 0x72, 0x75, 0x65})"},
		{"native bytes custom", expr.Bytes, "ok", "json.RawMessage", `[]byte("ok")`},
		{"native string meta", expr.String, []byte("é"), "string", "string([]byte{0xc3, 0xa9})"},
		{"byte to text", expr.String, json.RawMessage("é"), "Text", "Text([]byte{0xc3, 0xa9})"},
		{"binary to text", expr.String, []byte{0, 255}, "Text", "Text([]byte{0x0, 0xff})"},
		{"empty byte to text", expr.String, []byte{}, "Text", "Text([]byte{})"},
		{"empty raw", expr.Bytes, json.RawMessage{}, "json.RawMessage", "json.RawMessage([]byte{})"},
		{"generic bytes", expr.Bytes, renderTestBytes[int]("ok"), "Data[int]", "Data[int]([]byte{0x6f, 0x6b})"},
		{"generic string", expr.String, renderTestBytes[int]("ok"), "Label[int]", "Label[int]([]byte{0x6f, 0x6b})"},
		{"custom Any bytes", expr.Any, renderTestBytes[int]("ok"), "Data[int]", "Data[int]([]byte{0x6f, 0x6b})"},
		{"custom Any bool", expr.Any, renderTestBool(true), "Enabled", "Enabled(true)"},
		{"custom Any float", expr.Any, renderTestFloat(1.0000000000000002), "Number", "Number(1.0000000000000002)"},
	} {
		t.Run(test.name, func(t *testing.T) {
			metadata := []string{test.custom}
			if test.custom == "json.RawMessage" {
				metadata = []string{test.custom, "encoding/json", "json"}
			}
			attribute := &expr.AttributeExpr{
				Type: test.primitive,
				Meta: expr.MetaExpr{"struct:field:type": metadata},
			}
			layout := goValueTestLayout(t, attribute, GoLayoutPolicy{}, nil)
			code, err := RenderGoValue(attribute, test.value, layout, false, nil, "")
			require.NoError(t, err)
			require.Empty(t, code.Declarations)
			require.Equal(t, test.want, code.Expression)
			fmt.Fprintf(&assertions, "{ var got %s = %s; var want %s = %s; if !reflect.DeepEqual(got, want) { t.Errorf(%q, got, want) } }\n",
				test.custom, code.Expression, test.custom, test.want, test.name+": got %v, want %v")
		})
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module rendered.local\n\ngo 1.26.0\n"), 0o600))
	source := `package rendered
import ("encoding/json"; "reflect"; "testing")
type (
	Enabled bool
	Wide int64
	Unsigned uint64
	Number float64
	Text string
	Data[T any] []byte
	Label[T any] string
)
func TestValues(t *testing.T) {
` + assertions.String() + "\n}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values_test.go"), []byte(source), 0o600))
	command := exec.Command("go", "test", "./...")
	command.Dir = dir
	output, err := command.CombinedOutput()
	require.NoError(t, err, "%s", output)
}
