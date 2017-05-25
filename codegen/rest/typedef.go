package restgen

import (
	"fmt"
	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// GoTypeDef returns the Go code that defines the struct corresponding to ma.
// It defers from the function defined in the codegen package in the following
// ways:
//
//    - It defines marshaler tags on each fields taking using the HTTP element
//      names.
//
//    - It produced fields with pointers even if the corresponding attribute is
//      required so that the generated code may validate explicitly if ptr is
//      true.
//
func GoTypeDef(scope *codegen.NameScope, ma *rest.MappedAttributeExpr, ptr bool) string {
	switch actual := ma.Type.(type) {
	case design.Primitive:
		return codegen.GoNativeTypeName(actual)
	case *design.Array:
		d := GoTypeDef(scope, rest.NewMappedAttributeExpr(actual.ElemType), ptr)
		if design.IsObject(actual.ElemType.Type) {
			d = "*" + d
		}
		return "[]" + d
	case *design.Map:
		keyDef := GoTypeDef(scope, rest.NewMappedAttributeExpr(actual.KeyType), ptr)
		if design.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := GoTypeDef(scope, rest.NewMappedAttributeExpr(actual.ElemType), ptr)
		if design.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case design.Object:
		var ss []string
		ss = append(ss, "struct {")
		mat := ma.Attribute()
		WalkMappedAttr(ma, func(name, elem string, required bool, at *design.AttributeExpr) error {
			var (
				fn   string
				tdef string
				desc string
				tags string
			)
			{
				fn = codegen.GoifyAtt(at, name, true)
				tdef = GoTypeDef(scope, rest.NewMappedAttributeExpr(at), ptr)
				if design.IsPrimitive(at.Type) {
					if ptr || mat.IsPrimitivePointer(name) {
						tdef = "*" + tdef
					}
				} else if design.IsObject(at.Type) {
					tdef = "*" + tdef
				}
				if at.Description != "" {
					desc = codegen.Comment(at.Description) + "\n\t"
				}
				tags = attributeTags(mat, at, elem, ptr || !ma.IsRequired(name))
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s%s", desc, fn, tdef, tags))
			return nil
		})
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	case design.UserType:
		return scope.GoTypeName(actual)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// attributeTags computes the struct field tags.
func attributeTags(parent, att *design.AttributeExpr, t string, optional bool) string {
	if tags := codegen.AttributeTags(parent, att); tags != "" {
		return tags
	}
	var o string
	if optional {
		o = ",omitempty"
	}
	return fmt.Sprintf(" `form:\"%s%s\" json:\"%s%s\" xml:\"%s%s\"`", t, o, t, o, t, o)
}
