// These tests preserve conversion-helper comparison and measure planning when
// many different declarations are present.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func BenchmarkTransformHelperRegistryDistinctDeclarations(b *testing.B) {
	for _, count := range []int{64, 256} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			attribute := &expr.AttributeExpr{Type: expr.String}
			for index := range 16 {
				attribute = &expr.AttributeExpr{Type: &expr.UserTypeExpr{
					TypeName: fmt.Sprintf("Level%d", index),
					AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
						{Name: "child", Attribute: attribute},
					}},
				}}
			}
			program, err := NewTransformProgram(&TransformHooks{
				SameHelperDefinition: func(first, _, next, _ *expr.AttributeExpr) bool {
					return transformSemanticAttributesEqual(first, next, make(map[transformSemanticAttributePair]struct{}))
				},
			})
			require.NoError(b, err)
			candidates := make([]*transformHelperCandidate, count)
			for index := range candidates {
				plan := &TransformPlan{program: program, operations: []*transformOperation{{}, {}}}
				candidates[index] = &transformHelperCandidate{
					index:        index,
					plan:         plan,
					helper:       TransformHelper{ID: TransformHelperID{plan: plan}, Source: attribute, Target: attribute},
					definition:   TransformHelperDefinition{Source: attribute, Target: attribute},
					order:        testNameOrder{value: fmt.Sprint(index)},
					sourceLayout: &GoTypePlan{declaration: NewExactName(NameType, fmt.Sprintf("Value%d", index))},
					targetLayout: &GoTypePlan{},
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				registry := &TransformHelperRegistry{candidates: candidates}
				groups, err := registry.Finalize()
				require.NoError(b, err)
				require.Len(b, groups, count)
			}
		})
	}
}

func TestTransformSemanticAttributesKeepMappedRequiredness(t *testing.T) {
	object := &expr.Object{{Name: "value:wire", Attribute: &expr.AttributeExpr{Type: expr.String}}}
	for _, named := range []bool{false, true} {
		for _, required := range []bool{false, true} {
			for _, override := range []bool{false, true} {
				attribute := &expr.AttributeExpr{Type: object}
				if required {
					attribute.Validation = &expr.ValidationExpr{Required: []string{"value"}}
				}
				if named {
					attribute = &expr.AttributeExpr{Type: &expr.UserTypeExpr{TypeName: "Value", AttributeExpr: attribute}}
				}
				other := expr.DupAtt(attribute)
				if override {
					other.Validation = &expr.ValidationExpr{}
					if !required {
						other.Validation.Required = []string{"value"}
					}
				}
				want := expr.NewMappedAttributeExpr(attribute).IsRequired("value") == expr.NewMappedAttributeExpr(other).IsRequired("value")
				require.Equal(t, want, transformSemanticAttributesEqual(attribute, other, make(map[transformSemanticAttributePair]struct{})), "named=%v required=%v override=%v", named, required, override)
			}
		}
	}
}

func TestTransformHelperRegistryPreservesConservativeComparison(t *testing.T) {
	program, err := NewTransformProgram(&TransformHooks{
		SameHelperDefinition: func(first, _, next, _ *expr.AttributeExpr) bool {
			// A hook may decline to share a body even when the emitted code
			// would match. Here it permits A/B and B/C but declines A/C.
			return !(first.Description == "A" && next.Description == "C" ||
				first.Description == "C" && next.Description == "A")
		},
	})
	require.NoError(t, err)
	candidates := make([]*transformHelperCandidate, 5)
	for index, name := range []string{"A", "B", "C", "X", "Y"} {
		value := 0
		if index >= 3 {
			value = index
		}
		attribute := &expr.AttributeExpr{Type: expr.Int, Description: name, DefaultValue: value}
		plan := &TransformPlan{program: program, operations: []*transformOperation{{}, {}}}
		candidates[index] = &transformHelperCandidate{
			index:        index,
			plan:         plan,
			helper:       TransformHelper{ID: TransformHelperID{plan: plan}, Source: attribute, Target: attribute},
			definition:   TransformHelperDefinition{Source: attribute, Target: attribute},
			order:        testNameOrder{value: name},
			sourceLayout: &GoTypePlan{},
			targetLayout: &GoTypePlan{},
		}
	}
	// A calls X while B and C call Y. Once child calls separate A from B,
	// B and C can share one declaration despite their initial separate groups.
	for parent, child := range []int{3, 4, 4} {
		candidates[parent].plan.operations[1].calls = []transformCall{{helper: candidates[child].helper.ID}}
	}
	registry := &TransformHelperRegistry{candidates: candidates}
	groups, err := registry.Finalize()
	require.NoError(t, err)
	classes := make([]int, len(candidates))
	for class, group := range groups {
		for _, candidate := range group.candidates {
			classes[candidate.index] = class
		}
	}
	require.Equal(t, []int{0, 1, 1, 2, 3}, classes)
}
