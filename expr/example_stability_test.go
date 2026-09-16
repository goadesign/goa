// This file verifies that typed example owners make values independent of
// unrelated draws while preserving member-local values in composite examples.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestExampleOrderIndependence verifies that examples are a pure function of
// the design: the value computed for a user type does not depend on how many
// values were drawn from the generator before it nor on which other examples
// were computed first.
func TestExampleOrderIndependence(t *testing.T) {
	newUT := func(name string) *expr.UserTypeExpr {
		return &expr.UserTypeExpr{
			TypeName: name,
			AttributeExpr: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}},
					{Name: "count", Attribute: &expr.AttributeExpr{Type: expr.Int}},
					{Name: "tags", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}}},
				},
			},
		}
	}

	// Reference value computed on a fresh generator.
	stable := newUT("Stable")
	ref := stable.Example(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.UserTypeExampleIdentity(stable),
	))
	require.NotNil(t, ref)

	// Same value after unrelated draws were consumed from the generator.
	r := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.MethodPayloadExampleIdentity(exampleMethod("noise", "draw")),
	)
	noise := make([]any, 0, 14)
	for range 7 {
		noise = append(noise, r.Int(), r.String())
	}
	require.Len(t, noise, 14)
	require.Equal(t, ref, newUT("Stable").Example(r))

	// Same value after another type's example was computed first.
	r = expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.MethodPayloadExampleIdentity(exampleMethod("noise", "other")),
	)
	other := newUT("Other")
	other.Example(r.At(expr.UserTypeExampleIdentity(other)))
	require.Equal(t, ref, newUT("Stable").Example(r))
}

// TestExampleFieldLocality verifies that adding a field to an object does not
// change the examples of its existing fields.
func TestExampleFieldLocality(t *testing.T) {
	small := &expr.Object{
		{Name: "a", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "b", Attribute: &expr.AttributeExpr{Type: expr.Int}},
	}
	large := &expr.Object{
		{Name: "a", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "b", Attribute: &expr.AttributeExpr{Type: expr.Int}},
		{Name: "c", Attribute: &expr.AttributeExpr{Type: expr.Boolean}},
	}
	owner := expr.MethodPayloadExampleIdentity(exampleMethod("locality", "object"))
	exSmall := small.Example(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(owner)).(map[string]any)
	exLarge := large.Example(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(owner)).(map[string]any)
	require.Equal(t, exSmall["a"], exLarge["a"])
	require.Equal(t, exSmall["b"], exLarge["b"])
}

// TestExampleFieldAnchor verifies that the example computed for one element
// extracted from a user type matches the corresponding field value in the
// type's composite example.
func TestExampleFieldAnchor(t *testing.T) {
	field := &expr.AttributeExpr{Type: expr.String}
	ut := &expr.UserTypeExpr{
		TypeName: "Payload",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "id", Attribute: field},
				{Name: "label", Attribute: &expr.AttributeExpr{Type: expr.String}},
			},
		},
	}
	identity := expr.UserTypeExampleIdentity(ut)
	composite := ut.Example(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(identity)).(map[string]any)
	standalone := field.Example(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(identity).Member("id"))
	require.Equal(t, composite["id"], standalone)
}
