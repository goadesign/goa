package expr

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// chainUT builds a user type whose attribute has the given type and
// validation. Chaining calls produces alias chains.
func chainUT(name string, typ DataType, val *ValidationExpr) *UserTypeExpr {
	return &UserTypeExpr{
		TypeName:      name,
		UID:           "test#" + name,
		AttributeExpr: &AttributeExpr{Type: typ, Validation: val},
	}
}

func TestEffectiveValidation(t *testing.T) {
	fp := func(f float64) *float64 { return &f }
	ip := func(i int) *int { return &i }

	cases := []struct {
		Name string
		Att  func() *AttributeExpr
		Want *ValidationExpr
	}{
		{
			Name: "no validation",
			Att: func() *AttributeExpr {
				return &AttributeExpr{Type: String}
			},
			Want: nil,
		},
		{
			Name: "attribute only",
			Att: func() *AttributeExpr {
				return &AttributeExpr{
					Type:       String,
					Validation: &ValidationExpr{Pattern: "^a", Required: []string{"x"}},
				}
			},
			Want: &ValidationExpr{Pattern: "^a", Required: []string{"x"}},
		},
		{
			Name: "user type only",
			Att: func() *AttributeExpr {
				ut := chainUT("UT0", String, &ValidationExpr{Format: FormatEmail})
				return &AttributeExpr{Type: ut}
			},
			Want: &ValidationExpr{Format: FormatEmail},
		},
		{
			Name: "attribute wins over user type",
			Att: func() *AttributeExpr {
				ut := chainUT("UT0", String, &ValidationExpr{Format: FormatHostname, MinLength: ip(2)})
				return &AttributeExpr{
					Type:       ut,
					Validation: &ValidationExpr{Format: FormatEmail},
				}
			},
			Want: &ValidationExpr{Format: FormatEmail, MinLength: ip(2)},
		},
		{
			Name: "two level chain merges deeper level before immediate",
			Att: func() *AttributeExpr {
				// UT0 -> UT1: UT1's validation merges before UT0's so its
				// Format wins (legacy flatten-then-merge order).
				ut1 := chainUT("UT1", String, &ValidationExpr{Format: FormatDateTime, Pattern: "^p1"})
				ut0 := chainUT("UT0", ut1, &ValidationExpr{Format: FormatHostname, MinLength: ip(3)})
				return &AttributeExpr{Type: ut0}
			},
			Want: &ValidationExpr{Format: FormatDateTime, Pattern: "^p1", MinLength: ip(3)},
		},
		{
			Name: "three level chain with attribute validation",
			Att: func() *AttributeExpr {
				ut2 := chainUT("UT2", String, &ValidationExpr{Pattern: "^base", MinLength: ip(2), Required: []string{"c"}})
				ut1 := chainUT("UT1", ut2, &ValidationExpr{MaxLength: ip(20), Required: []string{"b"}})
				ut0 := chainUT("UT0", ut1, &ValidationExpr{Format: FormatHostname})
				return &AttributeExpr{
					Type:       ut0,
					Validation: &ValidationExpr{Values: []any{"a", "b"}, Required: []string{"a"}},
				}
			},
			Want: &ValidationExpr{
				Values:    []any{"a", "b"},
				Pattern:   "^base",
				Format:    FormatHostname,
				MinLength: ip(2),
				MaxLength: ip(20),
				Required:  []string{"a", "b", "c"},
			},
		},
		{
			Name: "chain with empty levels",
			Att: func() *AttributeExpr {
				ut2 := chainUT("UT2", String, &ValidationExpr{Minimum: fp(1), Maximum: fp(10)})
				ut1 := chainUT("UT1", ut2, nil)
				ut0 := chainUT("UT0", ut1, nil)
				return &AttributeExpr{Type: ut0}
			},
			Want: &ValidationExpr{Minimum: fp(1), Maximum: fp(10)},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			att := c.Att()
			before := dupAttForTest(att)

			got := EffectiveValidation(att)
			require.Equal(t, c.Want, got)

			// Purity: the input expression tree is unchanged.
			require.True(t, reflect.DeepEqual(before, dupAttForTest(att)), "EffectiveValidation mutated the input expressions")

			if got == nil {
				return
			}

			// Freshness: the result is not a pointer owned by any input
			// expression and its slices share no memory with the inputs.
			for a := att; a != nil; {
				require.NotSame(t, a.Validation, got)
				if a.Validation != nil {
					if len(got.Values) > 0 && len(a.Validation.Values) > 0 {
						require.NotSame(t, &got.Values[0], &a.Validation.Values[0])
					}
					if len(got.Required) > 0 && len(a.Validation.Required) > 0 {
						require.NotSame(t, &got.Required[0], &a.Validation.Required[0])
					}
				}
				ut, ok := a.Type.(UserType)
				if !ok {
					break
				}
				a = ut.Attribute()
			}

			// Mutating the result must not leak into the inputs.
			got.Pattern = "mutated"
			got.Format = "mutated"
			got.Required = append(got.Required, "mutated")
			if len(got.Values) > 0 {
				got.Values[0] = "mutated"
			}
			if got.Minimum != nil {
				*got.Minimum = -12345
			}
			if got.MinLength != nil {
				*got.MinLength = -12345
			}
			require.True(t, reflect.DeepEqual(before, dupAttForTest(att)), "mutating the result of EffectiveValidation leaked into the input expressions")
		})
	}
}

// dupAttForTest deep copies the validation-relevant state of the attribute
// and its user type alias chain for before/after comparisons.
func dupAttForTest(att *AttributeExpr) []*ValidationExpr {
	var res []*ValidationExpr
	for a := att; a != nil; {
		res = append(res, dupValForTest(a.Validation))
		ut, ok := a.Type.(UserType)
		if !ok {
			break
		}
		a = ut.Attribute()
	}
	return res
}

// dupValForTest deep copies a validation including its Values slice and
// scalar pointers so that in-place mutations of the original are detected.
func dupValForTest(v *ValidationExpr) *ValidationExpr {
	if v == nil {
		return nil
	}
	d := v.Dup()
	d.Values = append([]any(nil), v.Values...)
	d.ExclusiveMinimum = dupFloat(v.ExclusiveMinimum)
	d.Minimum = dupFloat(v.Minimum)
	d.Maximum = dupFloat(v.Maximum)
	d.ExclusiveMaximum = dupFloat(v.ExclusiveMaximum)
	d.MinLength = dupInt(v.MinLength)
	d.MaxLength = dupInt(v.MaxLength)
	return d
}
