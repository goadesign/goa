// This file builds the error definition that a method inherits from its
// service or API. HTTP and gRPC settings may change how the error is sent, but
// they do not change the service error value.
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

	// effectiveErrorCopier copies an inherited error without changing the
	// evaluated design. The maps reconnect recursive types to their copies.
	effectiveErrorCopier struct {
		attributes map[*AttributeExpr]*AttributeExpr
		userTypes  map[UserType]UserType
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
	first = effectiveErrorAttribute(first)
	second = effectiveErrorAttribute(second)
	return equivalentErrorAttributeNodes(first, second, make(map[attributePair]struct{}))
}

// differingErrorQualifierSettings lists error settings that would change the
// generated service error returned to callers.
func differingErrorQualifierSettings(first, second *AttributeExpr) []string {
	first = effectiveErrorAttribute(first)
	second = effectiveErrorAttribute(second)
	qualifiers := []struct {
		name string
		key  string
	}{
		{name: "temporary", key: "goa:error:temporary"},
		{name: "timeout", key: "goa:error:timeout"},
		{name: "fault", key: "goa:error:fault"},
	}
	var different []string
	for _, qualifier := range qualifiers {
		_, firstSet := first.Meta[qualifier.key]
		_, secondSet := second.Meta[qualifier.key]
		if firstSet != secondSet {
			different = append(different, qualifier.name)
		}
	}
	return different
}

// effectiveErrorAttribute returns a detached copy with References and Bases
// applied by AttributeExpr.Finalize. Validation can therefore compare the
// value contracts code generation will see without mutating evaluated design.
func effectiveErrorAttribute(source *AttributeExpr) *AttributeExpr {
	copier := &effectiveErrorCopier{
		attributes: make(map[*AttributeExpr]*AttributeExpr),
		userTypes:  make(map[UserType]UserType),
	}
	result := copier.attribute(source)
	result.Finalize()
	return result
}

// attribute copies one attribute shell before following its type and
// inheritance edges so self-recursive graphs terminate on the copied shell.
func (c *effectiveErrorCopier) attribute(source *AttributeExpr) *AttributeExpr {
	if source == nil {
		return nil
	}
	if copied, ok := c.attributes[source]; ok {
		return copied
	}
	copied := &AttributeExpr{
		Description:  source.Description,
		DefaultValue: cloneErrorContractValue(source.DefaultValue),
		DSLFunc:      source.DSLFunc,
	}
	c.attributes[source] = copied
	if source.Docs != nil {
		docs := *source.Docs
		copied.Docs = &docs
	}
	if source.Validation != nil {
		copied.Validation = cloneErrorValidation(source.Validation)
	}
	if source.Meta != nil {
		copied.Meta = source.Meta.Dup()
	}
	if len(source.UserExamples) > 0 {
		copied.UserExamples = make([]*ExampleExpr, len(source.UserExamples))
		for index, example := range source.UserExamples {
			copy := *example
			copy.Value = cloneErrorContractValue(example.Value)
			copied.UserExamples[index] = &copy
		}
	}
	copied.Type = c.dataType(source.Type)
	copied.Bases = c.dataTypes(source.Bases)
	copied.References = c.dataTypes(source.References)
	return copied
}

// cloneErrorValidation detaches slices and scalar pointers that
// ValidationExpr.Dup deliberately shares with its source.
func cloneErrorValidation(source *ValidationExpr) *ValidationExpr {
	copied := source.Dup()
	copied.Values = make([]any, len(source.Values))
	for index, value := range source.Values {
		copied.Values[index] = cloneErrorContractValue(value)
	}
	copied.ExclusiveMinimum = dupFloat(source.ExclusiveMinimum)
	copied.Minimum = dupFloat(source.Minimum)
	copied.Maximum = dupFloat(source.Maximum)
	copied.ExclusiveMaximum = dupFloat(source.ExclusiveMaximum)
	copied.MinLength = dupInt(source.MinLength)
	copied.MaxLength = dupInt(source.MaxLength)
	return copied
}

// cloneErrorContractValue copies the collection values accepted by defaults,
// enum validations, and examples. Primitive values are immutable and can be
// shared safely.
func cloneErrorContractValue(source any) any {
	switch actual := source.(type) {
	case Val:
		copied := make(Val, len(actual))
		for name, value := range actual {
			copied[name] = cloneErrorContractValue(value)
		}
		return copied
	case ArrayVal:
		copied := make(ArrayVal, len(actual))
		for index, value := range actual {
			copied[index] = cloneErrorContractValue(value)
		}
		return copied
	case MapVal:
		copied := make(MapVal, len(actual))
		for key, value := range actual {
			copied[cloneErrorContractValue(key)] = cloneErrorContractValue(value)
		}
		return copied
	case []any:
		copied := make([]any, len(actual))
		for index, value := range actual {
			copied[index] = cloneErrorContractValue(value)
		}
		return copied
	case []byte:
		return append([]byte(nil), actual...)
	case map[string]any:
		copied := make(map[string]any, len(actual))
		for name, value := range actual {
			copied[name] = cloneErrorContractValue(value)
		}
		return copied
	case map[any]any:
		copied := make(map[any]any, len(actual))
		for key, value := range actual {
			copied[cloneErrorContractValue(key)] = cloneErrorContractValue(value)
		}
		return copied
	default:
		return actual
	}
}

// dataTypes reconnects inheritance declarations to the same copied graph used
// by attribute types.
func (c *effectiveErrorCopier) dataTypes(source []DataType) []DataType {
	if len(source) == 0 {
		return nil
	}
	copied := make([]DataType, len(source))
	for index, dataType := range source {
		copied[index] = c.dataType(dataType)
	}
	return copied
}

// dataType copies each concrete type without registering generated result
// types. User-type shells are installed before their attributes are followed.
func (c *effectiveErrorCopier) dataType(source DataType) DataType {
	switch actual := source.(type) {
	case nil:
		return nil
	case Primitive:
		return actual
	case *Object:
		copied := make(Object, 0, len(*actual))
		for _, field := range *actual {
			copied = append(copied, &NamedAttributeExpr{
				Name:      field.Name,
				Attribute: c.attribute(field.Attribute),
			})
		}
		return &copied
	case *Array:
		return &Array{
			ElemType:         c.attribute(actual.ElemType),
			NonNullableElems: actual.NonNullableElems,
		}
	case *Map:
		return &Map{
			KeyType:  c.attribute(actual.KeyType),
			ElemType: c.attribute(actual.ElemType),
		}
	case *Union:
		copied := &Union{
			TypeName: actual.TypeName,
			TypeKey:  actual.TypeKey,
			ValueKey: actual.ValueKey,
			Values:   make([]*NamedAttributeExpr, len(actual.Values)),
		}
		for index, branch := range actual.Values {
			copied.Values[index] = &NamedAttributeExpr{
				Name:      branch.Name,
				Attribute: c.attribute(branch.Attribute),
			}
		}
		return copied
	case *ResultTypeExpr:
		origin := actual.Origin()
		if copied, ok := c.userTypes[origin]; ok {
			return copied
		}
		copied := &ResultTypeExpr{
			UserTypeExpr: &UserTypeExpr{
				TypeName: actual.TypeName,
				UID:      actual.UID,
			},
			Identifier:  actual.Identifier,
			ContentType: actual.ContentType,
		}
		c.userTypes[origin] = copied
		copied.AttributeExpr = c.attribute(actual.AttributeExpr)
		copied.Views = make([]*ViewExpr, len(actual.Views))
		for index, view := range actual.Views {
			copied.Views[index] = &ViewExpr{
				AttributeExpr: c.attribute(view.AttributeExpr),
				Name:          view.Name,
				Parent:        copied,
			}
		}
		return copied
	case *UserTypeExpr:
		origin := actual.Origin()
		if copied, ok := c.userTypes[origin]; ok {
			return copied
		}
		copied := &UserTypeExpr{
			TypeName: actual.TypeName,
			UID:      actual.UID,
		}
		c.userTypes[origin] = copied
		copied.AttributeExpr = c.attribute(actual.AttributeExpr)
		return copied
	case UserType:
		origin := actual.Origin()
		if copied, ok := c.userTypes[origin]; ok {
			return copied
		}
		copied := actual.Dup(nil)
		c.userTypes[origin] = copied
		copied.SetAttribute(c.attribute(actual.Attribute()))
		return copied
	default:
		panic("unknown error attribute type")
	}
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
		for index, branch := range firstType.Values {
			other := secondType.Values[index]
			if branch.Name != other.Name || !equivalentErrorAttributeNodes(branch.Attribute, other.Attribute, seen) {
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
