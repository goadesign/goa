// This file changes explicit HTTP property names in authored defaults. Scalar
// representations retain the conversion already used by OpenAPI defaults;
// strings containing JSON and map keys are never interpreted as field names.
package openapi

import (
	"reflect"
	"strings"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/internal/codegenname"
)

// ProjectDefault applies explicit HTTP property mappings to an authored default.
// Defaults without mapped properties retain the existing map-key conversion.
func ProjectDefault(attribute *expr.AttributeExpr, value any) any {
	if value == nil || !hasMappedProperties(attribute.Type, make(map[expr.UserType]bool)) {
		return ToStringMap(value)
	}
	return ToStringMap(mapDefaultProperties(attribute.Type, value))
}

// hasMappedProperties follows typed values and stops at recursive declarations.
// It reports whether a value needs property renaming before JSON serialization.
func hasMappedProperties(typ expr.DataType, seen map[expr.UserType]bool) bool {
	switch actual := typ.(type) {
	case expr.UserType:
		if seen[actual] {
			return false
		}
		seen[actual] = true
		return hasMappedProperties(actual.Attribute().Type, seen)
	case *expr.Object:
		for _, field := range *actual {
			if strings.Contains(field.Name, ":") || hasMappedProperties(field.Attribute.Type, seen) {
				return true
			}
		}
	case *expr.Array:
		return hasMappedProperties(actual.ElemType.Type, seen)
	case *expr.Map:
		return hasMappedProperties(actual.ElemType.Type, seen)
	case *expr.Union:
		for _, branch := range actual.Values {
			if hasMappedProperties(branch.Attribute.Type, seen) {
				return true
			}
		}
	}
	return false
}

// mapDefaultProperties copies containers that contain mapped properties and
// keeps scalar defaults in their authored representation, including pointers.
func mapDefaultProperties(typ expr.DataType, value any) any {
	if value == nil || !hasMappedProperties(typ, make(map[expr.UserType]bool)) {
		return value
	}
	if named, ok := typ.(expr.UserType); ok {
		return mapDefaultProperties(named.Attribute().Type, value)
	}
	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		reflected = reflected.Elem()
	}
	switch actual := typ.(type) {
	case *expr.Object:
		var values map[string]any
		if reflected.Kind() == reflect.Map {
			values, _ = exampleMap(reflected.Interface())
		}
		result := make(map[string]any, len(*actual))
		for _, field := range *actual {
			name, wire, mapped := strings.Cut(field.Name, ":")
			if !mapped {
				wire = name
			}
			member, exists := values[name]
			if reflected.Kind() == reflect.Struct {
				// Authored struct defaults use generated Go field names, just
				// as the default validator and service literal generator do.
				goName := codegenname.AttributeName(name, field.Attribute.Meta["struct:field:name"])
				fieldValue := reflected.FieldByName(codegenname.Goify(goName, true))
				exists = fieldValue.IsValid()
				if exists {
					member = fieldValue.Interface()
				}
			}
			if exists {
				result[wire] = mapDefaultProperties(field.Attribute.Type, member)
			}
		}
		return result
	case *expr.Array:
		result := make([]any, reflected.Len())
		for index := range result {
			result[index] = mapDefaultProperties(actual.ElemType.Type, reflected.Index(index).Interface())
		}
		return result
	case *expr.Map:
		result := make(map[any]any, reflected.Len())
		for _, key := range reflected.MapKeys() {
			result[key.Interface()] = mapDefaultProperties(actual.ElemType.Type, reflected.MapIndex(key).Interface())
		}
		return result
	case *expr.Union:
		if actual.Untagged {
			branch := actual.UntaggedBranch(reflected.Interface())
			return mapDefaultProperties(branch.Attribute.Type, reflected.Interface())
		}
		fields, _ := exampleMap(reflected.Interface())
		discriminator := reflect.ValueOf(fields[actual.GetTypeKey()])
		for discriminator.Kind() == reflect.Pointer || discriminator.Kind() == reflect.Interface {
			discriminator = discriminator.Elem()
		}
		for _, branch := range actual.Values {
			if branch.Name != discriminator.String() {
				continue
			}
			if actual.Flatten {
				result := mapDefaultProperties(branch.Attribute.Type, fields).(map[string]any)
				result[actual.GetTypeKey()] = fields[actual.GetTypeKey()]
				return result
			}
			return map[string]any{
				actual.GetTypeKey():  fields[actual.GetTypeKey()],
				actual.GetValueKey(): mapDefaultProperties(branch.Attribute.Type, fields[actual.GetValueKey()]),
			}
		}
		panic("validated union default has no matching branch")
	}
	return value
}
