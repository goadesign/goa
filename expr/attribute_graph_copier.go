// This file copies complete attribute graphs for generators that must keep a
// private expression snapshot and still resolve every copied node back to the
// exact input node used during planning.
package expr

import (
	"fmt"
	"reflect"
	"slices"
)

type (
	// AttributeGraphCopier copies one connected attribute graph without merging
	// distinct user types or breaking recursive links. Copies retain the authored
	// declaration and finalized state of every attribute. Reuse one copier for
	// every root that must share copied nodes.
	AttributeGraphCopier struct {
		attributes map[*AttributeExpr]*AttributeExpr
		originals  map[*AttributeExpr]*AttributeExpr
		types      map[DataType]DataType
	}

	// attributeValueReference identifies mutable data on the active copy path.
	// Slice length and capacity distinguish separate views of one array.
	attributeValueReference struct {
		typeOf   reflect.Type
		pointer  uintptr
		length   int
		capacity int
	}
)

// NewAttributeGraphCopier creates a copier for one expression graph. Call Copy
// for each root that belongs to that graph, then use Original when a generator
// must resolve a copied node through declarations planned from the input graph.
func NewAttributeGraphCopier() *AttributeGraphCopier {
	return &AttributeGraphCopier{
		attributes: make(map[*AttributeExpr]*AttributeExpr),
		originals:  make(map[*AttributeExpr]*AttributeExpr),
		types:      make(map[DataType]DataType),
	}
}

// Copy returns a deep copy of attribute. Shared and recursive nodes remain
// shared in the result. Calling Copy with a result made by this copier returns
// that result unchanged.
func (c *AttributeGraphCopier) Copy(attribute *AttributeExpr) *AttributeExpr {
	if attribute == nil {
		return nil
	}
	if _, copied := c.originals[attribute]; copied {
		return attribute
	}
	if copied, ok := c.attributes[attribute]; ok {
		return copied
	}
	copied := &AttributeExpr{
		Description: attribute.Description,
		DSLFunc:     attribute.DSLFunc,
		finalized:   attribute.finalized,
		authored:    attribute.AuthoredAttribute(),
	}
	c.attributes[attribute] = copied
	c.originals[copied] = attribute
	copied.Type = c.dataType(attribute.Type)
	copied.Bases = c.dataTypes(attribute.Bases)
	copied.References = c.dataTypes(attribute.References)
	if attribute.Docs != nil {
		docs := *attribute.Docs
		copied.Docs = &docs
	}
	if attribute.Validation != nil {
		copied.Validation = copyAttributeValidation(attribute.Validation)
	}
	copied.Meta = copyAttributeMeta(attribute.Meta)
	copied.DefaultValue = copyAttributeValue(attribute.DefaultValue)
	if len(attribute.UserExamples) > 0 {
		copied.UserExamples = make([]*ExampleExpr, len(attribute.UserExamples))
		for index, example := range attribute.UserExamples {
			if example == nil {
				continue
			}
			copiedExample := *example
			copiedExample.Value = copyAttributeValue(example.Value)
			copied.UserExamples[index] = &copiedExample
		}
	}
	return copied
}

// Original returns the input node copied to create attribute. It returns
// attribute unchanged when this copier did not create it.
func (c *AttributeGraphCopier) Original(attribute *AttributeExpr) *AttributeExpr {
	if original := c.originals[attribute]; original != nil {
		return original
	}
	return attribute
}

// dataTypes copies a list through the same graph so repeated and recursive
// declarations remain shared.
func (c *AttributeGraphCopier) dataTypes(dataTypes []DataType) []DataType {
	if dataTypes == nil {
		return nil
	}
	copied := make([]DataType, len(dataTypes))
	for index, dataType := range dataTypes {
		copied[index] = c.dataType(dataType)
	}
	return copied
}

// dataType installs each new type before copying its children so a recursive
// child points back to the same copied type.
func (c *AttributeGraphCopier) dataType(dataType DataType) DataType {
	if dataType == nil || dataType == Empty {
		return dataType
	}
	if copied, ok := c.types[dataType]; ok {
		return copied
	}
	switch actual := dataType.(type) {
	case Primitive:
		return actual
	case *Array:
		copied := &Array{NonNullableElems: actual.NonNullableElems}
		c.types[dataType] = copied
		copied.ElemType = c.Copy(actual.ElemType)
		return copied
	case *Object:
		copied := &Object{}
		c.types[dataType] = copied
		for _, named := range *actual {
			copied.Set(named.Name, c.Copy(named.Attribute))
		}
		return copied
	case *Map:
		copied := &Map{}
		c.types[dataType] = copied
		copied.KeyType = c.Copy(actual.KeyType)
		copied.ElemType = c.Copy(actual.ElemType)
		return copied
	case *Union:
		copied := &Union{
			TypeName: actual.TypeName,
			TypeKey:  actual.TypeKey,
			ValueKey: actual.ValueKey,
			Values:   make([]*NamedAttributeExpr, len(actual.Values)),
		}
		c.types[dataType] = copied
		for index, named := range actual.Values {
			copied.Values[index] = &NamedAttributeExpr{
				Name:      named.Name,
				Attribute: c.Copy(named.Attribute),
			}
		}
		return copied
	case *ResultTypeExpr:
		copied := actual.Dup(nil).(*ResultTypeExpr)
		copied.ContentType = actual.ContentType
		copied.Views = make([]*ViewExpr, len(actual.Views))
		c.types[dataType] = copied
		copied.SetAttribute(c.Copy(actual.Attribute()))
		for index, view := range actual.Views {
			copied.Views[index] = &ViewExpr{
				AttributeExpr: c.Copy(view.AttributeExpr),
				Name:          view.Name,
				Parent:        copied,
			}
		}
		return copied
	case UserType:
		copied := actual.Dup(nil)
		c.types[dataType] = copied
		copied.SetAttribute(c.Copy(actual.Attribute()))
		return copied
	default:
		panic(fmt.Sprintf("cannot copy attribute type %T", dataType)) // bug
	}
}

// copyAttributeValidation copies every slice, pointer, and value held by one
// validation so changing the copy cannot change the input graph.
func copyAttributeValidation(validation *ValidationExpr) *ValidationExpr {
	copied := validation.Dup()
	copied.Values = copyAttributeValue(validation.Values).([]any)
	copied.ExclusiveMinimum = dupFloat(validation.ExclusiveMinimum)
	copied.Minimum = dupFloat(validation.Minimum)
	copied.Maximum = dupFloat(validation.Maximum)
	copied.ExclusiveMaximum = dupFloat(validation.ExclusiveMaximum)
	copied.MinLength = dupInt(validation.MinLength)
	copied.MaxLength = dupInt(validation.MaxLength)
	return copied
}

// copyAttributeMeta copies both the metadata map and each value slice.
func copyAttributeMeta(meta MetaExpr) MetaExpr {
	if meta == nil {
		return nil
	}
	copied := meta.Dup()
	for name, values := range copied {
		copied[name] = slices.Clone(values)
	}
	return copied
}

// copyAttributeValue copies values used by defaults, validations, and examples
// without changing their concrete Go type.
func copyAttributeValue(value any) any {
	if value == nil {
		return nil
	}
	return copyAttributeReflectValue(
		reflect.ValueOf(value),
		make(map[attributeValueReference]struct{}),
	).Interface()
}

// copyAttributeReflectValue copies every mutable field reachable from value.
// Unsupported mutable values are rejected instead of being shared.
func copyAttributeReflectValue(value reflect.Value, active map[attributeValueReference]struct{}) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copied := reflect.New(value.Type()).Elem()
		copied.Set(copyAttributeReflectValue(value.Elem(), active))
		return copied
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		reference := enterAttributeValue(value, active)
		defer delete(active, reference)
		copied := reflect.New(value.Type().Elem())
		copied.Elem().Set(copyAttributeReflectValue(value.Elem(), active))
		return copied
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		reference := enterAttributeValue(value, active)
		defer delete(active, reference)
		copied := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := range value.Len() {
			copied.Index(index).Set(copyAttributeReflectValue(value.Index(index), active))
		}
		return copied
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		reference := enterAttributeValue(value, active)
		defer delete(active, reference)
		copied := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			copied.SetMapIndex(
				copyAttributeReflectValue(iterator.Key(), active),
				copyAttributeReflectValue(iterator.Value(), active),
			)
		}
		return copied
	case reflect.Array:
		copied := reflect.New(value.Type()).Elem()
		for index := range value.Len() {
			copied.Index(index).Set(copyAttributeReflectValue(value.Index(index), active))
		}
		return copied
	case reflect.Struct:
		copied := reflect.New(value.Type()).Elem()
		copied.Set(value)
		for index := range value.NumField() {
			field := value.Type().Field(index)
			if !field.IsExported() {
				if attributeTypeContainsReference(field.Type) {
					panic(fmt.Sprintf(
						"cannot copy attribute value of type %s: unexported field %s contains mutable data",
						value.Type(),
						field.Name,
					))
				}
				continue
			}
			copied.Field(index).Set(copyAttributeReflectValue(value.Field(index), active))
		}
		return copied
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.String:
		return value
	default:
		panic(fmt.Sprintf("cannot copy attribute value of type %s", value.Type()))
	}
}

// enterAttributeValue rejects a value that points back to one already being
// copied. Repeated values outside the active path are copied separately.
func enterAttributeValue(value reflect.Value, active map[attributeValueReference]struct{}) attributeValueReference {
	reference := attributeValueReference{
		typeOf:  value.Type(),
		pointer: value.Pointer(),
	}
	if value.Kind() == reflect.Slice {
		reference.length = value.Len()
		reference.capacity = value.Cap()
	}
	if _, exists := active[reference]; exists {
		panic(fmt.Sprintf("cannot copy cyclic attribute value of type %s", value.Type()))
	}
	active[reference] = struct{}{}
	return reference
}

// attributeTypeContainsReference reports whether a private field could share
// mutable state with the input value.
func attributeTypeContainsReference(valueType reflect.Type) bool {
	switch valueType.Kind() {
	case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface,
		reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return true
	case reflect.Array:
		return attributeTypeContainsReference(valueType.Elem())
	case reflect.Struct:
		for index := range valueType.NumField() {
			if attributeTypeContainsReference(valueType.Field(index).Type) {
				return true
			}
		}
	}
	return false
}
