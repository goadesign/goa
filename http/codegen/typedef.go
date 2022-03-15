package codegen

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// goTypeDef returns the Go code that defines the struct corresponding to ma.
// It differs from the function defined in the codegen package in the following
// ways:
//
//    - It defines marshaler tags on each fields using the HTTP element names.
//
//    - It produced fields with pointers even if the corresponding attribute is
//      required when ptr is true so that the generated code may validate
//      explicitly.
//
// useDefault directs whether fields holding primitive types with default values
// should hold pointers when ptr is false. If it is true then the fields are
// values even when not required (to account for the fact that they have a
// default value so cannot be nil) otherwise the fields are values only when
// required.
func goTypeDef(scope *codegen.NameScope, att *expr.AttributeExpr, ptr, useDefault bool) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		if t, _ := codegen.GetMetaType(att); t != "" {
			return t
		}
		return codegen.GoNativeTypeName(actual)
	case *expr.Array:
		d := goTypeDef(scope, actual.ElemType, ptr, useDefault)
		if expr.IsObject(actual.ElemType.Type) {
			d = "*" + d
		}
		return "[]" + d
	case *expr.Map:
		keyDef := goTypeDef(scope, actual.KeyType, ptr, useDefault)
		if expr.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := goTypeDef(scope, actual.ElemType, ptr, useDefault)
		if expr.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case *expr.Object:
		var ss []string
		ss = append(ss, "struct {")
		ma := expr.NewMappedAttributeExpr(att)
		mat := ma.Attribute()
		codegen.WalkMappedAttr(ma, func(name, elem string, required bool, at *expr.AttributeExpr) error {
			var (
				fn   string
				tdef string
				desc string
				tags string
			)
			{
				fn = codegen.GoifyAtt(at, name, true)
				tdef = goTypeDef(scope, at, ptr, useDefault)
				if expr.IsPrimitive(at.Type) {
					if (ptr || mat.IsPrimitivePointer(name, useDefault)) && at.Type != expr.Bytes && at.Type != expr.Any {
						tdef = "*" + tdef
					}
				} else if expr.IsObject(at.Type) {
					tdef = "*" + tdef
				}
				if at.Description != "" {
					desc = codegen.Comment(at.Description) + "\n\t"
				}
				var optional bool
				{
					switch {
					case ptr:
						optional = true
					case useDefault:
						optional = !ma.IsRequired(name) && !ma.HasDefaultValue(name)
					default:
						optional = !ma.IsRequired(name)
					}
				}
				tags = attributeTags(mat, at, elem, optional)
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s%s", desc, fn, tdef, tags))
			return nil
		})
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	case expr.UserType, *expr.Union:
		return scope.GoTypeName(att)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// attributeTags computes the struct field tags.
func attributeTags(parent, att *expr.AttributeExpr, t string, optional bool) string {
	if tags := codegen.AttributeTags(parent, att); tags != "" {
		return tags
	}
	var o string
	if optional {
		o = ",omitempty"
	}
	return fmt.Sprintf(" `form:\"%s%s\" json:\"%s%s\" xml:\"%s%s\"`", t, o, t, o, t, o)
}
