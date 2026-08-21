// This file defines the emitted Go and JSON identity used to name and emit Goa
// unions consistently. It is separate from expression-type compatibility.
package codegen

import (
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"
)

// UnionTypeHash returns a stable identity for the Go and JSON definition
// generated for u. Unlike expr.Union.Hash, which describes design-type
// compatibility, UnionTypeHash includes the effective JSON envelope keys and
// details that change generated Go branch types, such as package locations,
// field type metadata, and nilability.
func UnionTypeHash(u *expr.Union) string {
	var key strings.Builder
	writeUnionTypeHash(&key, u, make(map[*expr.Object]int), make(map[*expr.Union]int))
	return key.String()
}

// writeUnionTypeHash appends one union definition using length-prefixed values
// so different inputs cannot produce an ambiguous concatenation.
func writeUnionTypeHash(key *strings.Builder, union *expr.Union, objects map[*expr.Object]int, unions map[*expr.Union]int) {
	if index, ok := unions[union]; ok {
		writeUnionHashPart(key, "union-ref")
		writeUnionHashPart(key, strconv.Itoa(index))
		return
	}
	unions[union] = len(unions)
	defer delete(unions, union)
	writeUnionHashPart(key, "union")
	writeUnionHashPart(key, union.TypeName)
	writeUnionHashPart(key, union.GetTypeKey())
	writeUnionHashPart(key, union.GetValueKey())
	for _, value := range union.Values {
		writeUnionHashPart(key, value.Name)
		writeUnionAttributeHash(key, value.Attribute, objects, unions)
	}
}

// writeUnionAttributeHash appends the generated Go identity of an attribute.
func writeUnionAttributeHash(key *strings.Builder, att *expr.AttributeExpr, objects map[*expr.Object]int, unions map[*expr.Union]int) {
	writeUnionHashPart(key, strconv.FormatBool(IsNilable(att.Type)))
	if metaType, ok := att.Meta["struct:field:type"]; ok {
		writeUnionHashPart(key, "meta-type")
		for _, value := range metaType {
			writeUnionHashPart(key, value)
		}
	}
	switch actual := att.Type.(type) {
	case expr.Primitive:
		writeUnionHashPart(key, "primitive")
		writeUnionHashPart(key, GoNativeTypeName(actual))
	case expr.UserType:
		writeUnionHashPart(key, "user")
		writeUnionHashPart(key, Goify(actual.Name(), true))
		writeUnionHashPart(key, actual.Hash())
		if loc := UserTypeLocation(actual); loc != nil {
			writeUnionHashPart(key, loc.RelImportPath)
		} else {
			writeUnionHashPart(key, "")
		}
	case *expr.Array:
		writeUnionHashPart(key, "array")
		writeUnionAttributeHash(key, actual.ElemType, objects, unions)
	case *expr.Map:
		writeUnionHashPart(key, "map")
		writeUnionAttributeHash(key, actual.KeyType, objects, unions)
		writeUnionAttributeHash(key, actual.ElemType, objects, unions)
	case *expr.Object:
		writeUnionObjectHash(key, att, actual, objects, unions)
	case *expr.Union:
		writeUnionTypeHash(key, actual, objects, unions)
	case expr.CompositeExpr:
		writeUnionAttributeHash(key, actual.Attribute(), objects, unions)
	default:
		panic("unknown union branch data type")
	}
}

// writeUnionObjectHash appends the inline Go struct emitted for an object.
func writeUnionObjectHash(key *strings.Builder, parent *expr.AttributeExpr, object *expr.Object, objects map[*expr.Object]int, unions map[*expr.Union]int) {
	if index, ok := objects[object]; ok {
		writeUnionHashPart(key, "object-ref")
		writeUnionHashPart(key, strconv.Itoa(index))
		return
	}
	objects[object] = len(objects)
	defer delete(objects, object)
	writeUnionHashPart(key, "object")
	for _, field := range *object {
		writeUnionHashPart(key, GoifyAtt(field.Attribute, field.Name, true))
		writeUnionHashPart(key, AttributeTagsWithName(parent, field.Name, field.Attribute))
		writeUnionHashPart(key, strconv.FormatBool(goFieldIsPointer(parent, field.Name, false, false)))
		writeUnionAttributeHash(key, field.Attribute, objects, unions)
	}
}

// writeUnionHashPart appends one unambiguous string component to key.
func writeUnionHashPart(key *strings.Builder, value string) {
	key.WriteString(strconv.Itoa(len(value)))
	key.WriteByte(':')
	key.WriteString(value)
}
