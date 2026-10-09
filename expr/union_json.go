// This file checks JSON mappings for Goa OneOf declarations. Untagged branches
// need distinct JSON kinds, and flattened objects must leave the discriminator
// name to the union. Invalid designs stop during expression validation.
package expr

import (
	"reflect"
	"strings"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/internal/codegenname"
)

// JSONKind returns the JSON token kind written by a Goa type: '"' for strings,
// '0' for numbers, 't' for booleans, '[' for arrays, and '{' for objects. It returns
// zero when one fixed kind cannot be derived, including Any and nested unions.
// Named types with custom Go representations are also unknown because their
// JSON methods may write a different kind from the authored primitive.
func JSONKind(dataType DataType) byte {
	seen := make(map[UserType]struct{})
	for {
		if dataType == nil {
			return 0
		}
		named, ok := dataType.(UserType)
		if !ok {
			break
		}
		if _, visited := seen[named]; visited {
			return 0
		}
		seen[named] = struct{}{}
		if _, custom := named.Attribute().Meta["struct:field:type"]; custom {
			return 0
		}
		dataType = named.Attribute().Type
	}
	switch dataType.Kind() {
	case StringKind, BytesKind:
		return '"'
	case IntKind, Int32Kind, Int64Kind, UIntKind, UInt32Kind, UInt64Kind, Float32Kind, Float64Kind:
		return '0'
	case BooleanKind:
		return 't'
	case ArrayKind:
		return '['
	case ObjectKind, MapKind:
		return '{'
	default:
		return 0
	}
}

// UntaggedBranch selects the branch whose JSON kind matches an authored Go
// default or example. It leaves nested fields and validation rules to their
// existing validators. Unknown or ambiguous kinds return nil. Byte slices
// select the string kind because JSON encodes them as base64 strings.
func (u *Union) UntaggedBranch(value any) *NamedAttributeExpr {
	concrete, ok := concreteDefaultValue(reflect.ValueOf(value))
	if !ok {
		return nil
	}
	var kind byte
	switch concrete.Kind() {
	case reflect.String:
		kind = '"'
	case reflect.Bool:
		kind = 't'
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		kind = '0'
	case reflect.Array:
		kind = '['
	case reflect.Slice:
		kind = '['
		if concrete.Type().Elem().Kind() == reflect.Uint8 {
			kind = '"'
		}
	case reflect.Map, reflect.Struct:
		kind = '{'
	default:
		return nil
	}
	var selected *NamedAttributeExpr
	for _, branch := range u.Values {
		if JSONKind(branch.Attribute.Type) != kind {
			continue
		}
		if selected != nil {
			return nil
		}
		selected = branch
	}
	return selected
}

// validateUnionJSON checks mapping choices and branches on the evaluated
// expression. Generators receive a mapping whose branch selection is unambiguous.
func validateUnionJSON(attribute *AttributeExpr, parent eval.Expression) *eval.ValidationErrors {
	union := AsUnion(attribute.Type)
	if union.Untagged {
		errors := new(eval.ValidationErrors)
		_, flatten := attribute.Meta["oneof:json:flatten"]
		_, discriminator := attribute.Meta["oneof:type:field"]
		_, valueField := attribute.Meta["oneof:value:field"]
		if union.Flatten || union.TypeKey != "" || union.ValueKey != "" || flatten || discriminator || valueField {
			errors.Add(parent, "untagged OneOf %q cannot select flattening, a discriminator, or a value field", union.TypeName)
		}
		if _, custom := attribute.Meta["struct:field:type"]; custom {
			errors.Add(parent, "untagged OneOf %q cannot select a custom Go representation", union.TypeName)
		}
		kinds := make(map[byte]string)
		for _, branch := range union.Values {
			kind := JSONKind(branch.Attribute.Type)
			if _, custom := branch.Attribute.Meta["struct:field:type"]; custom {
				kind = 0
			}
			if kind == 0 {
				errors.Add(parent, "untagged OneOf %q branch %q must have one known JSON kind", union.TypeName, branch.Name)
				continue
			}
			if prior, exists := kinds[kind]; exists {
				errors.Add(parent, "untagged OneOf %q branches %q and %q have the same JSON kind", union.TypeName, prior, branch.Name)
			}
			kinds[kind] = branch.Name
		}
		return errors
	}
	if !union.Flatten {
		return nil
	}
	errors := new(eval.ValidationErrors)
	if union.ValueKey != "" {
		errors.Add(parent, "flattened OneOf %q cannot select a value field", union.TypeName)
	}
	for _, branch := range union.Values {
		object := AsObject(branch.Attribute.Type)
		if object == nil {
			errors.Add(parent, "flattened OneOf %q branch %q must be an object", union.TypeName, branch.Name)
			continue
		}
		mapped := NewMappedAttributeExpr(branch.Attribute)
		for _, field := range *AsObject(mapped.Type) {
			// HTTP mappings and service JSON tags can use different names. Neither
			// representation may take the field that the union writes itself.
			goName := codegenname.Goify(codegenname.AttributeName(field.Name, field.Attribute.Meta["struct:field:name"]), true)
			names := []string{mapped.ElemName(field.Name), goName}
			if tag, ok := field.Attribute.Meta["struct:tag:json"]; ok {
				name := strings.SplitN(strings.Join(tag, ","), ",", 2)[0]
				if name == "-" {
					continue
				}
				if name != "" {
					names = []string{name}
				}
			} else if tag, ok := field.Attribute.Meta["struct:tag:json:name"]; ok && len(tag) > 0 {
				names = []string{mapped.ElemName(field.Name), strings.Join(tag, ",")}
			}
			for _, name := range names {
				if name == union.GetTypeKey() {
					errors.Add(parent, "flattened OneOf %q branch %q field %q uses discriminator JSON name %q", union.TypeName, branch.Name, field.Name, name)
					break
				}
			}
		}
	}
	return errors
}
