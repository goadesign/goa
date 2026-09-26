// This file checks that small maximum lengths produce valid example sizes and
// values, using fixed random inputs to cover every possible subtraction.
package expr_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

type (
	lengthRandomizerFactory int

	lengthRandomizer struct {
		expr.DeterministicRandomizer
		value int
	}
)

func TestNewLengthSmallMaxLength(t *testing.T) {
	cases := []struct {
		maximum int
		lengths [3]int
	}{
		{0, [3]int{0, 0, 0}},
		{1, [3]int{1, 0, 1}},
		{2, [3]int{2, 1, 0}},
		{3, [3]int{3, 2, 1}},
		{4, [3]int{3, 3, 2}},
	}
	for _, tc := range cases {
		for value, length := range tc.lengths {
			t.Run(fmt.Sprintf("maximum_%d/random_%d", tc.maximum, value), func(t *testing.T) {
				att := &expr.AttributeExpr{
					Type:       expr.String,
					Validation: &expr.ValidationExpr{MaxLength: &tc.maximum},
				}
				r := expr.NewExampleGenerator(lengthRandomizerFactory(value)).At(
					expr.MethodPayloadExampleIdentity(exampleMethod("lengths", "bounded")),
				)

				require.Equal(t, length, expr.NewLength(att, r))
			})
		}
	}
}

func TestExampleSmallMaxLength(t *testing.T) {
	cases := []struct {
		name  string
		typ   expr.DataType
		empty any
	}{
		{"string", expr.String, ""},
		{"bytes", expr.Bytes, []byte{}},
		{"array", &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}, []string{}},
		{"map", &expr.Map{
			KeyType:  &expr.AttributeExpr{Type: expr.String},
			ElemType: &expr.AttributeExpr{Type: expr.String},
		}, map[string]string{}},
	}
	for _, tc := range cases {
		for maximum := range 3 {
			for value := range 3 {
				t.Run(fmt.Sprintf("%s/maximum_%d/random_%d", tc.name, maximum, value), func(t *testing.T) {
					att := &expr.AttributeExpr{
						Type:       tc.typ,
						Validation: &expr.ValidationExpr{MaxLength: &maximum},
					}
					r := expr.NewExampleGenerator(lengthRandomizerFactory(value)).At(
						expr.MethodPayloadExampleIdentity(exampleMethod("lengths", tc.name)),
					)

					var example any
					require.NotPanics(t, func() {
						example = att.Example(r)
					})
					require.IsType(t, tc.empty, example)
					require.LessOrEqual(t, reflect.ValueOf(example).Len(), maximum)
				})
			}
		}
	}
}

// NewRandomizer gives each example an independent generator with the selected
// integer input and fixed values for all other methods.
func (f lengthRandomizerFactory) NewRandomizer(expr.ExampleIdentity) expr.Randomizer {
	return &lengthRandomizer{value: int(f)}
}

// Int returns the selected test input so length checks do not depend on a seed.
func (r *lengthRandomizer) Int() int {
	return r.value
}
