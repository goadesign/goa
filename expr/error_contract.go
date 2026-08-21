// This file defines the transport-independent error contract used when an
// endpoint inherits HTTP or gRPC response policy from its service or API.
// Transport policy may select status codes and wire fields, but the method's
// effective error remains the concrete service value encoded on every path.
package expr

import (
	"reflect"
	"slices"
)

type (
	// attributePair identifies two nodes already compared while traversing
	// recursive error types.
	attributePair struct {
		first  *AttributeExpr
		second *AttributeExpr
	}
)

// equivalentErrorAttributes reports whether two error attributes generate the
// same service value contract. Descriptions and examples are documentation;
// types, validations, defaults, and metadata affect generated code or runtime
// behavior and must match.
func equivalentErrorAttributes(first, second *AttributeExpr) bool {
	if first == second {
		return true
	}
	if first == nil || second == nil {
		return false
	}
	return equivalentErrorAttributeNodes(first, second, make(map[attributePair]struct{}))
}

// equivalentErrorAttributeNodes compares every contract-bearing node while
// stopping when recursive user types revisit the same declaration pair.
func equivalentErrorAttributeNodes(first, second *AttributeExpr, seen map[attributePair]struct{}) bool {
	if first == second {
		return true
	}
	pair := attributePair{first: first, second: second}
	if _, ok := seen[pair]; ok {
		return true
	}
	seen[pair] = struct{}{}
	if !equivalentErrorValidation(first.Validation, second.Validation) ||
		!reflect.DeepEqual(first.DefaultValue, second.DefaultValue) ||
		!equivalentErrorMetadata(first.Meta, second.Meta) {
		return false
	}

	switch firstType := first.Type.(type) {
	case Primitive:
		secondType, ok := second.Type.(Primitive)
		return ok && firstType == secondType
	case UserType:
		secondType, ok := second.Type.(UserType)
		return ok &&
			firstType.Name() == secondType.Name() &&
			equivalentErrorAttributeNodes(firstType.Attribute(), secondType.Attribute(), seen)
	case *Object:
		secondType, ok := second.Type.(*Object)
		if !ok || len(*firstType) != len(*secondType) {
			return false
		}
		for _, field := range *firstType {
			other := secondType.Attribute(field.Name)
			if other == nil || !equivalentErrorAttributeNodes(field.Attribute, other, seen) {
				return false
			}
		}
	case *Array:
		secondType, ok := second.Type.(*Array)
		return ok &&
			firstType.NonNullableElems == secondType.NonNullableElems &&
			equivalentErrorAttributeNodes(firstType.ElemType, secondType.ElemType, seen)
	case *Map:
		secondType, ok := second.Type.(*Map)
		return ok &&
			equivalentErrorAttributeNodes(firstType.KeyType, secondType.KeyType, seen) &&
			equivalentErrorAttributeNodes(firstType.ElemType, secondType.ElemType, seen)
	case *Union:
		secondType, ok := second.Type.(*Union)
		if !ok ||
			firstType.TypeName != secondType.TypeName ||
			firstType.GetTypeKey() != secondType.GetTypeKey() ||
			firstType.GetValueKey() != secondType.GetValueKey() ||
			len(firstType.Values) != len(secondType.Values) {
			return false
		}
		for _, branch := range firstType.Values {
			var other *AttributeExpr
			for _, candidate := range secondType.Values {
				if candidate.Name == branch.Name {
					other = candidate.Attribute
					break
				}
			}
			if other == nil || !equivalentErrorAttributeNodes(branch.Attribute, other, seen) {
				return false
			}
		}
	default:
		return false
	}
	return true
}

// equivalentErrorValidation compares validation behavior independently of the
// authored order of required fields and enum values.
func equivalentErrorValidation(first, second *ValidationExpr) bool {
	if first == nil {
		first = new(ValidationExpr)
	}
	if second == nil {
		second = new(ValidationExpr)
	}
	firstScalars, secondScalars := *first, *second
	firstScalars.Required, secondScalars.Required = nil, nil
	firstScalars.Values, secondScalars.Values = nil, nil
	return reflect.DeepEqual(firstScalars, secondScalars) &&
		equivalentStringSet(first.Required, second.Required) &&
		equivalentValueSet(first.Values, second.Values)
}

// equivalentStringSet reports whether both slices contain the same distinct
// strings; validation order does not affect runtime behavior.
func equivalentStringSet(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for _, value := range first {
		if !slices.Contains(second, value) {
			return false
		}
	}
	return true
}

// equivalentValueSet reports whether both enum lists contain the same values
// regardless of declaration order.
func equivalentValueSet(first, second []any) bool {
	if len(first) != len(second) {
		return false
	}
	matched := make([]bool, len(second))
	for _, value := range first {
		found := false
		for index, candidate := range second {
			if !matched[index] && reflect.DeepEqual(value, candidate) {
				matched[index] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// equivalentErrorMetadata compares metadata keys and ordered values while
// treating nil and empty maps or value slices as the same absent content.
func equivalentErrorMetadata(first, second MetaExpr) bool {
	if len(first) != len(second) {
		return false
	}
	for key, values := range first {
		other, ok := second[key]
		if !ok || !slices.Equal(values, other) {
			return false
		}
	}
	return true
}
