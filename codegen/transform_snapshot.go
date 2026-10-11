// This file checks the private expression snapshots retained by conversion
// plans. Every planning hook and render attempt must leave those values
// unchanged; callers receive an error when a comparison finds a mutation.
package codegen

import (
	"fmt"
	"reflect"
	"slices"

	"goa.design/goa/v3/expr"
)

// transformAttributesEqual compares the expression facts that conversion
// planning and rendering consume. It follows paired recursive attributes once.
func transformAttributesEqual(left, right *expr.AttributeExpr, seen map[transformAttributePair]struct{}) bool {
	if left == nil || right == nil {
		return left == right
	}
	pair := transformAttributePair{source: left, target: right}
	if _, compared := seen[pair]; compared {
		return true
	}
	seen[pair] = struct{}{}
	if left.Description != right.Description || !reflect.DeepEqual(left.Docs, right.Docs) ||
		!transformValidationsEqual(left.Validation, right.Validation) || !transformMetadataEqual(left.Meta, right.Meta) ||
		!reflect.DeepEqual(left.DefaultValue, right.DefaultValue) || len(left.Bases) != len(right.Bases) ||
		len(left.References) != len(right.References) || len(left.UserExamples) != len(right.UserExamples) {
		return false
	}
	for index := range left.Bases {
		if !transformDataTypesEqual(left.Bases[index], right.Bases[index], seen) {
			return false
		}
	}
	for index := range left.References {
		if !transformDataTypesEqual(left.References[index], right.References[index], seen) {
			return false
		}
	}
	for index := range left.UserExamples {
		if left.UserExamples[index] == nil || right.UserExamples[index] == nil {
			if left.UserExamples[index] != right.UserExamples[index] {
				return false
			}
			continue
		}
		if left.UserExamples[index].Summary != right.UserExamples[index].Summary ||
			left.UserExamples[index].Description != right.UserExamples[index].Description ||
			!reflect.DeepEqual(left.UserExamples[index].Value, right.UserExamples[index].Value) {
			return false
		}
	}
	return transformDataTypesEqual(left.Type, right.Type, seen)
}

// transformDataTypesEqual compares the expression type graph below two
// attributes without comparing DSL functions or expression pointer identities.
func transformDataTypesEqual(left, right expr.DataType, seen map[transformAttributePair]struct{}) bool {
	if left == nil || right == nil {
		return left == right
	}
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false
	}
	switch left := left.(type) {
	case expr.Primitive:
		return left == right.(expr.Primitive)
	case *expr.Array:
		right := right.(*expr.Array)
		return left.NonNullableElems == right.NonNullableElems && transformAttributesEqual(left.ElemType, right.ElemType, seen)
	case *expr.Object:
		right := right.(*expr.Object)
		if len(*left) != len(*right) {
			return false
		}
		for index, attribute := range *left {
			other := (*right)[index]
			if attribute.Name != other.Name || !transformAttributesEqual(attribute.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case *expr.Map:
		right := right.(*expr.Map)
		return transformAttributesEqual(left.KeyType, right.KeyType, seen) &&
			transformAttributesEqual(left.ElemType, right.ElemType, seen)
	case *expr.Union:
		right := right.(*expr.Union)
		if left.TypeName != right.TypeName || left.TypeKey != right.TypeKey || left.ValueKey != right.ValueKey || len(left.Values) != len(right.Values) {
			return false
		}
		for index, attribute := range left.Values {
			other := right.Values[index]
			if attribute.Name != other.Name || !transformAttributesEqual(attribute.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case expr.UserType:
		right := right.(expr.UserType)
		return left.Origin() == right.Origin() && left.Name() == right.Name() &&
			transformAttributesEqual(left.Attribute(), right.Attribute(), seen)
	default:
		panic(fmt.Sprintf("cannot compare transform type %T", left)) // bug
	}
}

// transformValidationsEqual compares each validation value without asking
// reflection to walk the fixed validation structure. Enum values keep their
// arbitrary Go types, so reflection still compares those nested values.
func transformValidationsEqual(left, right *expr.ValidationExpr) bool {
	if left == right {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return left.Format == right.Format && left.Pattern == right.Pattern &&
		transformValidationPointersEqual(left.ExclusiveMinimum, right.ExclusiveMinimum) &&
		transformValidationPointersEqual(left.Minimum, right.Minimum) &&
		transformValidationPointersEqual(left.Maximum, right.Maximum) &&
		transformValidationPointersEqual(left.ExclusiveMaximum, right.ExclusiveMaximum) &&
		transformValidationPointersEqual(left.MinLength, right.MinLength) &&
		transformValidationPointersEqual(left.MaxLength, right.MaxLength) &&
		(left.Required == nil) == (right.Required == nil) &&
		slices.Equal(left.Required, right.Required) && reflect.DeepEqual(left.Values, right.Values)
}

// transformValidationPointersEqual preserves reflection's pointer comparison:
// an identical pointer is equal, including a pointer to a floating-point NaN;
// different non-nil pointers are equal only when their values compare equal.
func transformValidationPointersEqual[T comparable](left, right *T) bool {
	if left == right {
		return true
	}
	return left != nil && right != nil && *left == *right
}

// transformMetadataEqual compares all metadata keys and string values. Nil
// maps and slices stay distinct from empty ones, as in the retained snapshot.
func transformMetadataEqual(left, right expr.MetaExpr) bool {
	if (left == nil) != (right == nil) || len(left) != len(right) {
		return false
	}
	for name, values := range left {
		other, ok := right[name]
		if !ok || (values == nil) != (other == nil) || !slices.Equal(values, other) {
			return false
		}
	}
	return true
}

// changed compares both expression graphs with the copies made before
// planning. Hooks may inspect these graphs; changed reports any value edits
// so planning or rendering can reject the hook.
func (p *TransformPlan) changed() bool {
	if p.comparisonPairs == nil {
		p.comparisonPairs = make(map[transformAttributePair]struct{})
	}
	// Clear the visited pairs before each walk so a previous successful check
	// cannot hide changes made by the next hook.
	clear(p.comparisonPairs)
	if !transformAttributesEqual(p.source, p.sourceBaseline, p.comparisonPairs) {
		return true
	}
	clear(p.comparisonPairs)
	return !transformAttributesEqual(p.target, p.targetBaseline, p.comparisonPairs)
}
