// These tests keep copied defaults, mapped names, original child identities,
// and hook mutation errors stable while an object conversion reuses its maps.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTransformObjectMatchesCopiedDefaults(t *testing.T) {
	for _, secondDefault := range []any{"second", nil} {
		t.Run(map[bool]string{true: "absent", false: "different"}[secondDefault == nil], func(t *testing.T) {
			first := &expr.UserTypeExpr{
				TypeName: "Label",
				AttributeExpr: &expr.AttributeExpr{
					Type: expr.String, DefaultValue: "first",
				},
			}
			second := first.Dup(&expr.AttributeExpr{Type: expr.String, DefaultValue: secondDefault})
			firstField := &expr.AttributeExpr{Type: first}
			secondField := &expr.AttributeExpr{Type: second}
			object := &expr.AttributeExpr{Type: &expr.Object{
				{Name: "a:left", Attribute: firstField},
				{Name: "b:right", Attribute: secondField},
			}}
			matches := newTransformObjectMatches(object, object)
			for range 2 {
				var names []string
				matches.walk(func(source, target *expr.MappedAttributeExpr, sourceChild, targetChild *expr.AttributeExpr, name string) {
					names = append(names, name)
					require.Equal(t, "first", source.GetDefault(name))
					require.Equal(t, "first", target.GetDefault(name))
					require.False(t, source.IsPrimitivePointer(name, true))
					require.Same(t, sourceChild, targetChild)
					if name == "a" {
						require.Same(t, firstField, sourceChild)
						require.Equal(t, "left", source.ElemName(name))
					} else {
						require.Same(t, secondField, sourceChild)
						require.Equal(t, "right", source.ElemName(name))
					}
				})
				require.Equal(t, []string{"a", "b"}, names)
			}
			require.Equal(t, secondDefault, second.Attribute().DefaultValue)
		})
	}
}

func TestTransformObjectRejectsMutationBetweenPasses(t *testing.T) {
	for _, phase := range []string{"primitive", "object"} {
		t.Run(phase, func(t *testing.T) {
			source := &expr.AttributeExpr{Type: &expr.Object{
				{Name: "label", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}}
			target := expr.DupAtt(source)
			hooks := &TransformHooks{}
			if phase == "primitive" {
				hooks.ConvertPrimitive = func(source, _ *expr.AttributeExpr, _ string, _, _ bool, _ *TransformAttrs) (string, bool) {
					source.DefaultValue = "changed"
					return "", false
				}
			} else {
				hooks.ObjectDeref = func(target *expr.AttributeExpr) (string, bool) {
					target.Validation = &expr.ValidationExpr{Required: []string{"label"}}
					return "", false
				}
			}
			plan, err := NewTransformPlan(source, target, "", hooks)
			require.NoError(t, err)
			require.NoError(t, plan.BindContexts(
				NewAttributeContext(false, false, true, "", NewNameScope()),
				NewAttributeContext(false, false, true, "", NewNameScope()),
			))
			code, helpers, err := plan.Render("source", "target", true)
			require.EqualError(t, err, "transform render hook changed the retained plan")
			require.Empty(t, code)
			require.Empty(t, helpers)
		})
	}
}
