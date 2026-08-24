// This file builds a repeatable key from every detail that changes a generated
// union's Go or JSON definition.
package codegen

import (
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// UnionTypeID identifies the Go and JSON definition emitted for a union.
	// It is distinct from expr.Union.Hash, which describes design compatibility.
	UnionTypeID string
)

// NewUnionTypeID returns a repeatable key for union's generated Go and JSON
// definitions. The key includes the effective JSON envelope keys and every
// detail that changes a generated Go branch type: package location, field type
// metadata, and whether the value may be nil.
func NewUnionTypeID(union *expr.Union) UnionTypeID {
	var key strings.Builder
	writeUnionTypeID(
		&key,
		union,
		make(map[*expr.Object]int),
		make(map[*expr.Union]int),
		make(map[expr.UserType]int),
	)
	return UnionTypeID(key.String())
}

// Hash returns the repeatable key used to look up this union's Go name in a
// generated package.
func (id UnionTypeID) Hash() string {
	return string(id)
}

// writeUnionTypeID appends one union definition using length-prefixed values
// so different inputs cannot produce an ambiguous concatenation.
func writeUnionTypeID(key *strings.Builder, union *expr.Union, objects map[*expr.Object]int, unions map[*expr.Union]int, userTypes map[expr.UserType]int) {
	if index, ok := unions[union]; ok {
		writeUnionIDPart(key, "union-ref")
		writeUnionIDPart(key, strconv.Itoa(index))
		return
	}
	unions[union] = len(unions)
	defer delete(unions, union)
	writeUnionIDPart(key, "union")
	writeUnionIDPart(key, union.TypeName)
	writeUnionIDPart(key, union.GetTypeKey())
	writeUnionIDPart(key, union.GetValueKey())
	for _, value := range union.Values {
		writeUnionIDPart(key, value.Name)
		writeUnionAttributeID(key, value.Attribute, objects, unions, userTypes)
	}
}

// writeUnionAttributeID appends every attribute detail that changes generated
// Go code.
func writeUnionAttributeID(key *strings.Builder, att *expr.AttributeExpr, objects map[*expr.Object]int, unions map[*expr.Union]int, userTypes map[expr.UserType]int) {
	writeUnionIDPart(key, strconv.FormatBool(IsNilable(att.Type)))
	if metaType, ok := att.Meta["struct:field:type"]; ok {
		writeUnionIDPart(key, "meta-type")
		for _, value := range metaType {
			writeUnionIDPart(key, value)
		}
	}
	switch actual := att.Type.(type) {
	case expr.Primitive:
		writeUnionIDPart(key, "primitive")
		writeUnionIDPart(key, GoNativeTypeName(actual))
	case expr.UserType:
		writeUnionIDPart(key, "user")
		writeUnionIDPart(key, Goify(actual.Name(), true))
		if loc := UserTypeLocation(actual); loc != nil {
			writeUnionIDPart(key, loc.RelImportPath)
		} else {
			writeUnionIDPart(key, "")
		}
		origin := actual.Origin()
		if index, ok := userTypes[origin]; ok {
			writeUnionIDPart(key, "user-ref")
			writeUnionIDPart(key, strconv.Itoa(index))
			return
		}
		userTypes[origin] = len(userTypes)
		defer delete(userTypes, origin)
		writeUnionAttributeID(key, actual.Attribute(), objects, unions, userTypes)
	case *expr.Array:
		writeUnionIDPart(key, "array")
		writeUnionAttributeID(key, actual.ElemType, objects, unions, userTypes)
	case *expr.Map:
		writeUnionIDPart(key, "map")
		writeUnionAttributeID(key, actual.KeyType, objects, unions, userTypes)
		writeUnionAttributeID(key, actual.ElemType, objects, unions, userTypes)
	case *expr.Object:
		writeUnionObjectID(key, att, actual, objects, unions, userTypes)
	case *expr.Union:
		writeUnionTypeID(key, actual, objects, unions, userTypes)
	case expr.CompositeExpr:
		writeUnionAttributeID(key, actual.Attribute(), objects, unions, userTypes)
	default:
		panic("unknown union branch data type")
	}
}

// writeUnionObjectID appends the inline Go struct emitted for an object.
func writeUnionObjectID(key *strings.Builder, parent *expr.AttributeExpr, object *expr.Object, objects map[*expr.Object]int, unions map[*expr.Union]int, userTypes map[expr.UserType]int) {
	if index, ok := objects[object]; ok {
		writeUnionIDPart(key, "object-ref")
		writeUnionIDPart(key, strconv.Itoa(index))
		return
	}
	objects[object] = len(objects)
	defer delete(objects, object)
	writeUnionIDPart(key, "object")
	for _, field := range *object {
		writeUnionIDPart(key, GoifyAtt(field.Attribute, field.Name, true))
		writeUnionIDPart(key, AttributeTagsWithName(parent, field.Name, field.Attribute))
		writeUnionIDPart(key, strconv.FormatBool(goFieldIsPointer(parent, field.Name, false, false)))
		writeUnionAttributeID(key, field.Attribute, objects, unions, userTypes)
	}
}

// writeUnionIDPart appends one unambiguous string component to key.
func writeUnionIDPart(key *strings.Builder, value string) {
	key.WriteString(strconv.Itoa(len(value)))
	key.WriteByte(':')
	key.WriteString(value)
}
