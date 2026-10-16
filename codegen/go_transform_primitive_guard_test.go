// These tests execute generated primitive field conversions so custom guards
// preserve explicit pointer values and apply defaults when a pointer is absent.
package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTransformPrimitiveFieldGuard(t *testing.T) {
	cases := []struct {
		name          string
		sourcePointer bool
		targetPointer bool
		sourceAlias   bool
		targetAlias   bool
		required      bool
		sourceDefault any
		targetDefault any
		guard         string
		handled       bool
		assertions    string
	}{
		{
			name: "value condition", targetPointer: true, sourceDefault: false,
			guard: "if source.Value {\n", handled: true,
			assertions: `if convert(source{Value: false}).Value != nil { t.Fatal("false must remain absent") }
			if got := convert(source{Value: true}).Value; got == nil || !*got { t.Fatal("true must be copied") }`,
		},
		{
			name: "source alias condition", sourceAlias: true, targetPointer: true, sourceDefault: false,
			guard: "if source.Value {\n", handled: true,
			assertions: `if convert(source{Value: false}).Value != nil { t.Fatal("false alias must remain absent") }
			if got := convert(source{Value: true}).Value; got == nil || !*got { t.Fatal("true alias must be copied") }`,
		},
		{
			name: "target alias condition", targetAlias: true, targetPointer: true, sourceDefault: false,
			guard: "if source.Value {\n", handled: true,
			assertions: `if convert(source{Value: false}).Value != nil { t.Fatal("false must remain absent") }
			if got := convert(source{Value: true}).Value; got == nil || !*got { t.Fatal("true must become an alias") }`,
		},
		{
			name: "pointer default", sourcePointer: true, targetDefault: true,
			guard: "if source.Value != nil {\n", handled: true,
			assertions: `if !convert(source{}).Value { t.Fatal("absent value must take the default") }
			zero := false
			if convert(source{Value: &zero}).Value { t.Fatal("explicit false must survive") }`,
		},
		{
			name: "pointer alias default", sourcePointer: true, sourceAlias: true, targetDefault: true,
			guard: "if source.Value != nil {\n", handled: true,
			assertions: `if !convert(source{}).Value { t.Fatal("absent alias must take the default") }
			zero := Flag(false)
			if convert(source{Value: &zero}).Value { t.Fatal("explicit false alias must survive") }`,
		},
		{
			name: "pointer copy", sourcePointer: true, targetPointer: true,
			guard: "if source.Value != nil {\n", handled: true,
			assertions: `if convert(source{}).Value != nil { t.Fatal("absent pointer must remain absent") }
			zero := false
			if got := convert(source{Value: &zero}).Value; got == nil || *got { t.Fatal("explicit false pointer must survive") }`,
		},
		{
			name: "required pointer", sourcePointer: true, required: true,
			guard: "if source.Value != nil {\n", handled: true,
			assertions: `if convert(source{}).Value { t.Fatal("skipped required pointer must leave zero") }
			value := true
			if !convert(source{Value: &value}).Value { t.Fatal("present required pointer must be copied") }`,
		},
		{
			name: "empty handled guard", sourcePointer: true,
			handled: true,
			assertions: `zero := false
			if convert(source{Value: &zero}).Value { t.Fatal("explicit false must survive") }`,
		},
		{
			name: "unhandled guard", sourcePointer: true, targetDefault: true,
			guard: "if false {\n",
			assertions: `if !convert(source{}).Value { t.Fatal("normal default must survive") }
			zero := false
			if convert(source{Value: &zero}).Value { t.Fatal("normal pointer copy must survive") }`,
		},
	}
	var generated strings.Builder
	generated.WriteString("package generated\nimport \"testing\"\ntype Flag bool\n")
	for index, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			sourceField := &expr.AttributeExpr{Type: expr.Boolean, DefaultValue: test.sourceDefault}
			targetField := &expr.AttributeExpr{Type: expr.Boolean, DefaultValue: test.targetDefault}
			alias := &expr.UserTypeExpr{TypeName: "Flag", AttributeExpr: &expr.AttributeExpr{Type: expr.Boolean}}
			if test.sourceAlias {
				sourceField.Type = alias
			}
			if test.targetAlias {
				targetField.Type = alias
			}
			source := &expr.AttributeExpr{Type: &expr.Object{{Name: "value", Attribute: sourceField}}}
			target := &expr.AttributeExpr{Type: &expr.Object{{Name: "value", Attribute: targetField}}}
			if test.required {
				source.Validation = &expr.ValidationExpr{Required: []string{"value"}}
				target.Validation = &expr.ValidationExpr{Required: []string{"value"}}
			}
			// A value source without a default needs a required field to use
			// a Go value; pointer contexts still retain presence for decoding.
			if !test.sourcePointer && test.sourceDefault == nil {
				source.Validation = &expr.ValidationExpr{Required: []string{"value"}}
			}
			if !test.targetPointer && test.targetDefault == nil {
				target.Validation = &expr.ValidationExpr{Required: []string{"value"}}
			}
			scope := NewNameScope()
			sourceContext := NewAttributeContext(test.sourcePointer, false, true, "", scope)
			targetContext := NewAttributeContext(test.targetPointer, false, true, "", scope)
			plan, err := NewTransformPlan(source, target, "", &TransformHooks{
				GuardCondition: func(attribute *expr.AttributeExpr, variable string, required, pointer bool) (string, bool) {
					require.True(t, expr.IsPrimitive(attribute.Type))
					require.Equal(t, "source.Value", variable)
					require.Equal(t, source.IsRequired("value"), required)
					require.Equal(t, test.sourcePointer, pointer)
					return test.guard, test.handled
				},
			})
			require.NoError(t, err)
			require.NoError(t, plan.BindContexts(sourceContext, targetContext))
			code, helpers, err := plan.Render("source", "target", true)
			require.NoError(t, err)
			require.Empty(t, helpers)
			if test.handled && test.guard == "" {
				require.NotContains(t, code, "if source.Value != nil")
			}
			sourceType := sourceContext.Scope.Name(source, "", test.sourcePointer, true)
			targetType := targetContext.Scope.Name(target, "", test.targetPointer, true)
			fmt.Fprintf(&generated, "func TestCase%d(t *testing.T) {\ntype source = %s\nconvert := func(source source) *%s {\n%s\nreturn target\n}\n%s\n}\n",
				index, sourceType, targetType, code, test.assertions)
		})
	}
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte("module example.com/primitiveguard\n\ngo 1.26.0\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "guard_test.go"), []byte(generated.String()), 0o600))
	command := exec.Command("go", "test", "-count=1", "./...")
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoError(t, err, "generated field guards failed:\n%s", output)
}
