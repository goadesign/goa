// This file lets transport generators copy authored attributes into the value
// shapes used by HTTP codecs. Copies omit service package placement and retain
// validation and custom field types; the authored expressions stay unchanged.
package codegen

import "goa.design/goa/v3/expr"

// WireAttribute returns a copy of att for HTTP encoding and decoding. It replaces
// named scalar and collection aliases with their underlying types, combines
// their validation, and uses their definition's defaults and examples. Object,
// result, and union types remain structured values. The copy omits service
// package placement metadata and retains custom Go field type metadata.
//
// Transport generators must create and retain this copy during planning, before
// generated names are frozen. Keep the original attribute separately when
// planning conversions between service values and transport values.
func WireAttribute(att *expr.AttributeExpr) *expr.AttributeExpr {
	return copyWireAttribute(expr.DupAtt(att), make(map[expr.UserType]struct{}))
}

// copyWireAttribute prepares a copied attribute and its nested definitions for
// HTTP codecs. Recording visited definitions lets recursive types terminate
// without merging unrelated definitions that happen to have the same name.
func copyWireAttribute(att *expr.AttributeExpr, seen map[expr.UserType]struct{}) *expr.AttributeExpr {
	delete(att.Meta, "struct:pkg:path")
	switch dt := att.Type.(type) {
	case expr.UserType:
		if dt == expr.Empty {
			// An empty result uses a shared definition. Leave it unchanged so
			// copying a transport value does not change other designs.
			return att
		}
		_, resultType := dt.(*expr.ResultTypeExpr)
		alias := !resultType && !expr.IsObject(dt)
		if alias {
			// For a named scalar or collection, use the default and examples
			// from its definition in the copied transport attribute.
			att.DefaultValue = dt.Attribute().DefaultValue
			att.UserExamples = dt.Attribute().UserExamples
		}
		origin := dt.Origin()
		if _, ok := seen[origin]; !ok {
			seen[origin] = struct{}{}
			dt.SetAttribute(copyWireAttribute(dt.Attribute(), seen))
		}
		if alias {
			// Resolve nested aliases before replacing this reference. HTTP
			// decoding then receives the underlying value type and all inherited
			// validation, including when the same definition appears again.
			att.Type = dt.Attribute().Type
			if v := dt.Attribute().Validation; v != nil {
				if att.Validation == nil {
					att.Validation = v
				} else {
					att.Validation.Merge(v)
				}
			}
		}
	case *expr.Array:
		dt.ElemType = copyWireAttribute(dt.ElemType, seen)
	case *expr.Map:
		dt.KeyType = copyWireAttribute(dt.KeyType, seen)
		dt.ElemType = copyWireAttribute(dt.ElemType, seen)
	case *expr.Object:
		obj := make(expr.Object, len(*dt))
		for i, nat := range *dt {
			obj[i] = &expr.NamedAttributeExpr{Name: nat.Name, Attribute: copyWireAttribute(nat.Attribute, seen)}
		}
		att.Type = &obj
	case *expr.Union:
		// Copy each union branch before the transport generator assigns
		// names to its request and response types.
		for _, branch := range dt.Values {
			branch.Attribute = copyWireAttribute(branch.Attribute, seen)
		}
	}
	return att
}
