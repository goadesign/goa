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
//   - It defines marshaler tags on each fields using the HTTP element names.
//
//   - It produced fields with pointers even if the corresponding attribute is
//     required when ptr is true so that the generated code may validate
//     explicitly.
//
// useDefault directs whether fields holding primitive types with default values
// should hold pointers when ptr is false. If it is true then the fields are
// values even when not required (to account for the fact that they have a
// default value so cannot be nil) otherwise the fields are values only when
// required.
func goTypeDef(scope *codegen.NameScope, att *expr.AttributeExpr, ptr, useDefault bool) string {
	ctx := codegen.NewAttributeContext(ptr, false, useDefault, "", scope)
	ctx.UnionPointer = true
	return goTypeDefForContext(att, ctx)
}

// goTypeDefForContext recursively renders an HTTP body type using the same
// field representation consulted by transport conversion and validation.
func goTypeDefForContext(att *expr.AttributeExpr, ctx *codegen.AttributeContext) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		if t, _ := codegen.GetMetaType(att); t != "" {
			return t
		}
		return codegen.GoNativeTypeName(actual)
	case *expr.Array:
		d := goTypeDefForContext(actual.ElemType, ctx)
		if expr.IsObject(actual.ElemType.Type) || ctx.IsArrayElementPointer(actual) {
			d = "*" + d
		}
		return "[]" + d
	case *expr.Map:
		keyDef := goTypeDefForContext(actual.KeyType, ctx)
		if expr.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := goTypeDefForContext(actual.ElemType, ctx)
		if expr.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case *expr.Object:
		var ss []string
		ss = append(ss, "struct {")
		ma := expr.NewMappedAttributeExpr(att)
		codegen.WalkMappedAttr(ma, func(name, elem string, _ bool, at *expr.AttributeExpr) error { // nolint: errcheck
			var (
				fn   string
				tdef string
				desc string
				tags string
			)
			{
				fn = codegen.GoifyAtt(at, name, true)
				tdef = goTypeDefForContext(at, ctx)
				if ctx.IsFieldPointer(name, att) {
					tdef = "*" + tdef
				}
				if at.Description != "" {
					desc = codegen.Comment(at.Description) + "\n\t"
				}
				var optional bool
				{
					switch {
					case ctx.Pointer:
						optional = true
					case ctx.UseDefault:
						optional = !ma.IsRequired(name) && !ma.HasDefaultValue(name)
					default:
						optional = !ma.IsRequired(name)
					}
				}
				tags = attributeTags(at, elem, optional)
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s%s", desc, fn, tdef, tags))
			return nil
		})
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	case expr.UserType, *expr.Union:
		return ctx.Scope.Name(att, ctx.Pkg(att), ctx.Pointer, ctx.UseDefault)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// attributeTags computes the struct field tags.
func attributeTags(att *expr.AttributeExpr, t string, optional bool) string {
	if tags := codegen.AttributeTags(att); tags != "" {
		return tags
	}
	var o string
	// Always use omitempty for JSON-RPC ID attributes, even when required
	// since it is part of a different top-level field in the transport
	if optional || isJSONRPCID(att) {
		o = ",omitempty"
	}
	jsonName := t
	if att != nil && att.Meta != nil {
		if v := att.Meta["struct:tag:json:name"]; len(v) > 0 && v[0] != "" {
			jsonName = strings.Join(v, ",")
		}
	}
	return fmt.Sprintf(" `form:\"%s%s\" json:\"%s%s\" xml:\"%s%s\"`", t, o, jsonName, o, t, o)
}

// isJSONRPCID checks if the attribute is marked as a JSON-RPC ID attribute
func isJSONRPCID(att *expr.AttributeExpr) bool {
	if att.Meta == nil {
		return false
	}
	_, ok := att.Meta["jsonrpc:id"]
	return ok
}
