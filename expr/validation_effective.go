package expr

import "slices"

// EffectiveValidation computes the validation that effectively applies to the
// given attribute, combining the validation defined directly on the attribute
// with the validations defined by the user type alias chain of its type. It is
// pure: it never mutates att or any expression reachable from it, and the
// returned expression (including its slices) shares no memory with the design
// expression tree.
//
// The merge order is part of the contract as ValidationExpr.Merge gives
// precedence to the receiver (first writer wins for Values, Format and
// Pattern). Writing UT0 for att.Type when it is a user type, UT1 for the user
// type of UT0.Attribute().Type when it is one, and so on down the alias
// chain, validations are merged in this exact order:
//
//  1. att.Validation
//  2. UT1.Attribute().Validation, UT2.Attribute().Validation, ... down the
//     alias chain (skipping the immediate user type UT0)
//  3. UT0.Attribute().Validation last
//
// This order preserves the output that codegen historically produced by
// flattening alias chain validations into the attribute before merging the
// immediate user type validation.
//
// EffectiveValidation returns nil when no validation applies.
func EffectiveValidation(att *AttributeExpr) *ValidationExpr {
	var res *ValidationExpr
	merge := func(val *ValidationExpr) {
		if val == nil {
			return
		}
		if res == nil {
			res = &ValidationExpr{}
		}
		res.Merge(val)
	}
	merge(att.Validation)
	if ut, ok := att.Type.(UserType); ok {
		inner, isUT := ut.Attribute().Type.(UserType)
		for isUT {
			merge(inner.Attribute().Validation)
			inner, isUT = inner.Attribute().Type.(UserType)
		}
		merge(ut.Attribute().Validation)
	}
	if res == nil {
		return nil
	}
	// Detach the result: Merge assigns slices and scalar pointers from its
	// argument, so clone them to guarantee the caller cannot mutate the
	// design expressions through the returned value.
	res.Values = slices.Clone(res.Values)
	res.Required = slices.Clone(res.Required)
	res.ExclusiveMinimum = dupFloat(res.ExclusiveMinimum)
	res.Minimum = dupFloat(res.Minimum)
	res.Maximum = dupFloat(res.Maximum)
	res.ExclusiveMaximum = dupFloat(res.ExclusiveMaximum)
	res.MinLength = dupInt(res.MinLength)
	res.MaxLength = dupInt(res.MaxLength)
	return res
}

// dupFloat returns a fresh copy of the given float pointer, nil for nil.
func dupFloat(p *float64) *float64 {
	if p == nil {
		return nil
	}
	f := *p
	return &f
}

// dupInt returns a fresh copy of the given int pointer, nil for nil.
func dupInt(p *int) *int {
	if p == nil {
		return nil
	}
	i := *p
	return &i
}
