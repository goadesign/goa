// Named HTTP scalars keep their service types while wire shaping resolves the
// primitive value and inherited validation on a detached expression graph.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestMakeHTTPTypeResolvesNamedScalarChains(t *testing.T) {
	for _, depth := range []int{0, 1, 2, 3} {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			minimum, maximum := 4, 8
			base := &expr.AttributeExpr{
				Type:       expr.String,
				Validation: &expr.ValidationExpr{Pattern: "^[a-z]+$"},
				Meta:       expr.MetaExpr{"struct:pkg:path": {"types"}},
			}
			field := base
			var definitions []*expr.UserTypeExpr
			for level := 1; level <= depth; level++ {
				definition := &expr.UserTypeExpr{
					TypeName:      fmt.Sprintf("Level%d", level),
					UID:           fmt.Sprintf("level-%d", level),
					AttributeExpr: field,
				}
				definitions = append(definitions, definition)
				field = &expr.AttributeExpr{Type: definition}
			}
			if depth == 0 {
				base.Validation.MinLength = &minimum
			} else {
				outer := definitions[len(definitions)-1].Attribute()
				if outer.Validation == nil {
					outer.Validation = &expr.ValidationExpr{}
				}
				outer.Validation.MinLength = &minimum
			}
			field.Validation = &expr.ValidationExpr{MaxLength: &maximum}
			if depth == 0 {
				field.Validation.Pattern = "^[a-z]+$"
				field.Validation.MinLength = &minimum
			}
			body := &expr.AttributeExpr{Type: &expr.Object{
				{Name: "first", Attribute: field},
				{Name: "second", Attribute: &expr.AttributeExpr{
					Type:       field.Type,
					Validation: field.Validation.Dup(),
				}},
			}}

			for attempt := range 2 {
				wire := makeHTTPType(body)
				for _, name := range []string{"first", "second"} {
					value := expr.AsObject(wire.Type).Attribute(name)
					require.Equal(t, expr.String, value.Type, "attempt %d field %s", attempt, name)
					require.Equal(t, "^[a-z]+$", value.Validation.Pattern)
					require.Equal(t, minimum, *value.Validation.MinLength)
					require.Equal(t, maximum, *value.Validation.MaxLength)
					require.NotContains(t, value.Meta, "struct:pkg:path")
				}
			}

			require.Equal(t, expr.String, base.Type)
			require.Equal(t, "^[a-z]+$", base.Validation.Pattern)
			require.Contains(t, base.Meta, "struct:pkg:path", "wire shaping must not mutate authored metadata")
			for level := 1; level < len(definitions); level++ {
				require.Same(t, definitions[level-1], definitions[level].Attribute().Type)
			}
			if depth > 1 {
				require.Empty(t, definitions[len(definitions)-1].Attribute().Validation.Pattern, "inherited validation stays on the source base")
			}
		})
	}
}

func TestMakeHTTPTypePreservesNamedScalarDefaultAndExamples(t *testing.T) {
	baseExample := &expr.ExampleExpr{Value: "base-example"}
	outerExample := &expr.ExampleExpr{Value: "outer-example"}
	base := &expr.UserTypeExpr{
		TypeName: "Base", UID: "base",
		AttributeExpr: &expr.AttributeExpr{
			Type: expr.String, DefaultValue: "base-default",
			UserExamples: []*expr.ExampleExpr{baseExample},
		},
	}
	outer := &expr.UserTypeExpr{
		TypeName: "Outer", UID: "outer",
		AttributeExpr: &expr.AttributeExpr{
			Type: base, DefaultValue: "outer-default",
			UserExamples: []*expr.ExampleExpr{outerExample},
		},
	}
	field := &expr.AttributeExpr{Type: outer, DefaultValue: "field-default"}

	wire := makeHTTPType(field)

	require.Equal(t, expr.String, wire.Type)
	require.Equal(t, "outer-default", wire.DefaultValue, "retain the existing definition-level selection")
	require.Len(t, wire.UserExamples, 1)
	require.Equal(t, "outer-example", wire.UserExamples[0].Value)
	require.Equal(t, "field-default", field.DefaultValue)
	require.Equal(t, "outer-default", outer.Attribute().DefaultValue)
	require.Equal(t, "base-default", base.Attribute().DefaultValue)
	require.Same(t, base, outer.Attribute().Type)
}
