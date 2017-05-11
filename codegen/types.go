package codegen

import (
	"fmt"
	"sort"
	"strings"

	"goa.design/goa.v2/design"
)

// GoTypeDef returns the Go code that defines a Go type which matches the data
// structure definition (the part that comes after `type foo`). If public is
// true then the generated type is public and does not includes JSON, XML and
// form tags.
func GoTypeDef(att *design.AttributeExpr, public bool) string {
	switch actual := att.Type.(type) {
	case design.Primitive:
		return GoType(actual, public)
	case *design.Array:
		d := GoTypeDef(actual.ElemType, public)
		if design.IsObject(actual.ElemType.Type) {
			d = "*" + d
		}
		return "[]" + d
	case *design.Map:
		keyDef := GoTypeDef(actual.KeyType, public)
		if design.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := GoTypeDef(actual.ElemType, public)
		if design.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case design.Object:
		var ss []string
		ss = append(ss, "struct {")
		WalkAttributes(actual, func(name string, at *design.AttributeExpr) error {
			var (
				fn   string
				tdef string
				desc string
				tags string
			)
			{
				fn = GoifyAtt(at, name, true)
				tdef = GoTypeDef(at, public)
				if (at.Type.Kind() != design.BytesKind) &&
					(design.IsPrimitive(at.Type) && !public || design.IsObject(at.Type) || att.IsPrimitivePointer(name)) {
					tdef = "*" + tdef
				}
				if at.Description != "" {
					desc = fmt.Sprintf("// %s\n\t", at.Description)
				}
				if !public {
					tags = attributeTags(att, at, name)
				}
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s%s", desc, fn, tdef, tags))
			return nil
		})
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	case design.UserType:
		return GoType(actual, public)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// GoTypeRef returns the Go code that refers to the Go type which matches the
// given data type. If public is true then the reference is to a public type.
func GoTypeRef(dt design.DataType, public bool) string {
	tname := GoType(dt, public)
	if _, ok := dt.(design.Object); ok {
		return tname
	}
	if design.IsObject(dt) {
		return "*" + tname
	}
	return tname
}

// GoPackageTypeRef returns the Go code that refers to the given type. If the
// type is a user type then it is assumed to be defined in the given package.
func GoPackageTypeRef(dt design.DataType, pack string) string {
	tdef := GoTypeRef(dt, true)
	if _, ok := dt.(design.UserType); ok {
		if design.IsObject(dt) {
			return "*" + pack + "." + tdef[1:]
		}
		return pack + "." + tdef
	}
	return tdef
}

// GoTypeName produces a valid Go type identifier for the given data type.
func GoTypeName(dt design.DataType) string {
	switch actual := dt.(type) {
	case design.Primitive:
		return GoNativeTypeName(dt)
	case *design.Array:
		return GoTypeName(actual.ElemType.Type) + "Array"
	case *design.Map:
		return GoTypeName(actual.KeyType.Type) + GoTypeName(actual.ElemType.Type) + "Map"
	case design.Object:
		return "Object"
	case design.UserType:
		return Goify(actual.Name(), true)
	case design.CompositeExpr:
		return GoTypeName(actual.Attribute().Type)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// GoNativeTypeName returns a valid Go identifier for the given data type.
// GoNativeTypeName panics if t is not a primitive type.
func GoNativeTypeName(t design.DataType) string {
	switch t.Kind() {
	case design.BooleanKind:
		return "Boolean"
	case design.IntKind, design.Int32Kind, design.Int64Kind:
		return "Integer"
	case design.UIntKind, design.UInt32Kind, design.UInt64Kind:
		return "Unsigned"
	case design.Float32Kind, design.Float64Kind:
		return "Float"
	case design.StringKind:
		return "String"
	case design.BytesKind:
		return "Bytes"
	case design.AnyKind:
		return "Any"
	default:
		panic(fmt.Sprintf("cannot compute type name for non primitive %T", t)) // bug
	}
}

// GoType returns the Go type name of the given data type. It returns the
// public name if public is true.
func GoType(dt design.DataType, public bool) string {
	switch actual := dt.(type) {
	case design.Primitive:
		return GoNativeType(dt)
	case *design.Array:
		return "[]" + GoTypeRef(actual.ElemType.Type, public)
	case *design.Map:
		return fmt.Sprintf("map[%s]%s", GoTypeRef(actual.KeyType.Type, public), GoTypeRef(actual.ElemType.Type, public))
	case design.Object:
		return "map[string]interface{}"
	case design.UserType:
		return Goify(actual.Name(), public)
	case design.CompositeExpr:
		return GoType(actual.Attribute().Type, public)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// GoNativeType returns the Go built-in type corresponding to the given
// primitive type. GoNativeType panics if t is not a primitive type.
func GoNativeType(t design.DataType) string {
	switch t.Kind() {
	case design.BooleanKind:
		return "bool"
	case design.IntKind:
		return "int"
	case design.Int32Kind:
		return "int32"
	case design.Int64Kind:
		return "int64"
	case design.UIntKind:
		return "uint"
	case design.UInt32Kind:
		return "uint32"
	case design.UInt64Kind:
		return "uint64"
	case design.Float32Kind:
		return "float32"
	case design.Float64Kind:
		return "float64"
	case design.StringKind:
		return "string"
	case design.BytesKind:
		return "[]byte"
	case design.AnyKind:
		return "interface{}"
	default:
		panic(fmt.Sprintf("cannot compute native Go type for %T", t)) // bug
	}
}

// attributeTags computes the struct field tags.
func attributeTags(parent, att *design.AttributeExpr, name string) string {
	var elems []string
	keys := make([]string, len(att.Metadata))
	i := 0
	for k := range att.Metadata {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := att.Metadata[key]
		if strings.HasPrefix(key, "struct:tag:") {
			name := key[11:]
			value := strings.Join(val, ",")
			elems = append(elems, fmt.Sprintf("%s:\"%s\"", name, value))
		}
	}
	if len(elems) > 0 {
		return " `" + strings.Join(elems, " ") + "`"
	}
	// Default algorithm
	return fmt.Sprintf(" `form:\"%s,omitempty\" json:\"%s,omitempty\" xml:\"%s,omitempty\"`", name, name, name)
}
