// These tests preserve exact snapshot comparisons and measure the work of
// checking each planning hook on a large object.
package codegen

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTransformValidationComparisonMatchesReflection(t *testing.T) {
	zero, one, nan := 0.0, 1.0, math.NaN()
	zeroLength, oneLength := 0, 1
	cases := []*expr.ValidationExpr{
		nil, {},
		{Format: expr.FormatDate}, {Pattern: "[a-z]+"},
		{Values: []any{}}, {Values: []any{"value"}},
		{Values: []any{map[string]any{"nested": []int{1, 2}}}},
		{Values: []any{map[string]any{"nested": []int{1, 3}}}},
		{Values: []any{nan}},
		{Required: []string{}}, {Required: []string{"first", "second"}},
		{Required: []string{"second", "first"}},
		{Minimum: &zero}, {Minimum: &one}, {Minimum: &nan},
		{Maximum: &zero}, {Maximum: &one},
		{ExclusiveMinimum: &zero}, {ExclusiveMinimum: &one},
		{ExclusiveMaximum: &zero}, {ExclusiveMaximum: &one},
		{MinLength: &zeroLength}, {MinLength: &oneLength},
		{MaxLength: &zeroLength}, {MaxLength: &oneLength},
	}
	for leftIndex, left := range cases {
		for rightIndex, right := range cases {
			assert.Equal(t, reflect.DeepEqual(left, right), transformValidationsEqual(left, right), "cases %d and %d", leftIndex, rightIndex)
			if right != nil {
				copied := expr.NewAttributeGraphCopier().Copy(&expr.AttributeExpr{Type: expr.String, Validation: right}).Validation
				assert.Equal(t, reflect.DeepEqual(left, copied), transformValidationsEqual(left, copied), "case %d and copy of %d", leftIndex, rightIndex)
			}
		}
	}
}

func TestTransformMetadataComparisonMatchesReflection(t *testing.T) {
	cases := []expr.MetaExpr{
		nil, {}, {"key": nil}, {"key": {}}, {"key": {"first"}},
		{"other": {"first"}}, {"key": {"first", "second"}},
		{"key": {"second", "first"}}, {"key": {"first"}, "other": {}},
	}
	for leftIndex, left := range cases {
		for rightIndex, right := range cases {
			assert.Equal(t, reflect.DeepEqual(left, right), transformMetadataEqual(left, right), "cases %d and %d", leftIndex, rightIndex)
		}
	}
}

func TestTransformPlanChecksPreviouslyVisitedSibling(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}}
	var retained *expr.AttributeExpr
	var fieldCalls int
	_, err := NewTransformPlan(attribute, attribute, "", &TransformHooks{
		UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
			if retained == nil {
				retained = source
			}
			return source, target, nil
		},
		FieldPairAttrs: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			fieldCalls++
			if fieldCalls == 2 {
				expr.AsObject(retained.Type).Attribute("first").Meta = expr.MetaExpr{"changed": {"yes"}}
			}
			return source, target
		},
	})
	assert.Equal(t, 2, fieldCalls)
	assert.EqualError(t, err, "transform planning hook FieldPairAttrs changed the retained plan")
}

func TestTransformSnapshotRechecksRecursiveGraph(t *testing.T) {
	node := &expr.UserTypeExpr{TypeName: "Node"}
	field := &expr.AttributeExpr{Type: expr.String, DefaultValue: map[string]any{"values": []string{"before"}}}
	node.AttributeExpr = &expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: field},
		{Name: "next", Attribute: &expr.AttributeExpr{Type: node}},
	}}
	attribute := &expr.AttributeExpr{Type: node}
	plan, err := NewTransformPlan(attribute, attribute, "", nil)
	require.NoError(t, err)
	assert.False(t, plan.changed())
	assert.False(t, plan.changed())
	retained := expr.AsObject(plan.source.Type).Attribute("value")
	retained.DefaultValue.(map[string]any)["values"].([]string)[0] = "after"
	assert.True(t, plan.changed())
	retained.DefaultValue.(map[string]any)["values"].([]string)[0] = "before"
	assert.False(t, plan.changed())
	retainedTarget := expr.AsObject(plan.target.Type).Attribute("value")
	retainedTarget.Validation = &expr.ValidationExpr{Required: []string{}}
	assert.True(t, plan.changed())
}

func BenchmarkTransformPlanningMutationChecks(b *testing.B) {
	for _, size := range []int{64, 256} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			object := make(expr.Object, 0, size)
			for index := range size {
				object = append(object, &expr.NamedAttributeExpr{
					Name: fmt.Sprintf("field%d", index),
					Attribute: &expr.AttributeExpr{
						Type:       expr.String,
						Validation: &expr.ValidationExpr{Pattern: "[a-z]+"},
						Meta:       expr.MetaExpr{"field:tag": {"value"}},
					},
				})
			}
			attribute := &expr.AttributeExpr{Type: &object}
			hooks := &TransformHooks{
				UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
					return source, target, nil
				},
				FieldPairAttrs: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
					return source, target
				},
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_, err := NewTransformPlan(attribute, attribute, "", hooks)
				require.NoError(b, err)
			}
		})
	}
}
