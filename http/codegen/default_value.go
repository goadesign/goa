// This file translates authored defaults into the exact values accepted by
// generated HTTP command-line flags. Service defaults remain in service code;
// this file only owns HTTP field names and JSON encoding choices.
package codegen

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// requestElementDefault returns the default used by one generated request
// field. Object payloads own field defaults; complete primitive and collection
// payloads carry the default on the transport attribute itself.
func requestElementDefault(payload *expr.AttributeExpr, name string, attribute *expr.AttributeExpr) any {
	if expr.IsObject(payload.Type) {
		return payload.GetDefault(name)
	}
	return attribute.DefaultValue
}

// clientBodyDefault translates an authored default into the value that the
// generated HTTP command parser serializes. A top-level byte slice is entered
// as plain flag text. Byte slices inside JSON values remain byte slices so the
// standard JSON encoder writes their base64 representation.
func clientBodyDefault(attribute *expr.AttributeExpr, value any) any {
	if value == nil {
		return nil
	}
	return projectHTTPDefault(attribute, reflect.ValueOf(value), true)
}

// projectHTTPDefault translates one designed value into its HTTP wire shape.
// The complete design is available here, so generated programs only parse the
// already specialized flag default.
func projectHTTPDefault(attribute *expr.AttributeExpr, value reflect.Value, topLevel bool) any {
	value = concreteDefaultValue(value)
	switch {
	case expr.IsPrimitive(attribute.Type):
		if attribute.Type.Kind() != expr.BytesKind {
			return value.Interface()
		}
		bytes := defaultBytes(value)
		if topLevel {
			return string(bytes)
		}
		return bytes
	case expr.IsArray(attribute.Type):
		array := expr.AsArray(attribute.Type)
		items := make([]any, value.Len())
		for index := range items {
			items[index] = projectHTTPDefault(array.ElemType, value.Index(index), false)
		}
		return items
	case expr.IsMap(attribute.Type):
		mapped := expr.AsMap(attribute.Type)
		items := make(map[string]any, value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			key := httpDefaultMapKey(mapped.KeyType, iterator.Key())
			items[key] = projectHTTPDefault(mapped.ElemType, iterator.Value(), false)
		}
		return items
	case expr.IsObject(attribute.Type):
		mapped := expr.NewMappedAttributeExpr(attribute)
		items := make(map[string]any)
		for _, field := range *expr.AsObject(mapped.Type) {
			fieldValue, ok := defaultObjectField(value, field.Name, field.Attribute)
			if !ok {
				continue
			}
			name, include := httpDefaultFieldName(field.Attribute, mapped.ElemName(field.Name))
			if include {
				items[name] = projectHTTPDefault(field.Attribute, fieldValue, false)
			}
		}
		return items
	case expr.IsUnion(attribute.Type):
		return projectHTTPUnionDefault(expr.AsUnion(attribute.Type), value)
	default:
		panic(fmt.Sprintf("cannot translate HTTP default for %s", attribute.Type.Name())) // bug
	}
}

// projectHTTPUnionDefault keeps the union keys defined by the DSL and
// translates the selected branch value recursively.
func projectHTTPUnionDefault(union *expr.Union, value reflect.Value) any {
	typeValue, ok := defaultObjectField(value, union.GetTypeKey(), nil)
	if !ok {
		panic(fmt.Sprintf("union default is missing %q", union.GetTypeKey())) // bug
	}
	branchName := concreteDefaultValue(typeValue).String()
	var branch *expr.AttributeExpr
	for _, candidate := range union.Values {
		if candidate.Name == branchName {
			branch = candidate.Attribute
			break
		}
	}
	if branch == nil {
		panic(fmt.Sprintf("union default selects unknown branch %q", branchName)) // bug
	}
	branchValue, ok := defaultObjectField(value, union.GetValueKey(), branch)
	if !ok {
		panic(fmt.Sprintf("union default is missing %q", union.GetValueKey())) // bug
	}
	return map[string]any{
		union.GetTypeKey():  branchName,
		union.GetValueKey(): projectHTTPDefault(branch, branchValue, false),
	}
}

// httpDefaultMapKey returns the JSON object key written for one designed map
// key. JSON represents byte keys with the same base64 text used for byte
// values and represents every other allowed primitive key as text.
func httpDefaultMapKey(attribute *expr.AttributeExpr, value reflect.Value) string {
	value = concreteDefaultValue(value)
	if attribute.Type.Kind() == expr.BytesKind {
		return base64.StdEncoding.EncodeToString(defaultBytes(value))
	}
	return fmt.Sprint(value.Interface())
}

// httpDefaultFieldName applies the same name precedence as the generated JSON
// struct tag: a complete json tag, then a json name override, then the name
// mapped by the HTTP body design.
func httpDefaultFieldName(attribute *expr.AttributeExpr, mappedName string) (string, bool) {
	for _, key := range []string{"struct:tag:json", "struct:tag:json:name"} {
		values := attribute.Meta[key]
		if len(values) == 0 {
			continue
		}
		name := strings.SplitN(strings.Join(values, ","), ",", 2)[0]
		if name == "-" {
			return "", false
		}
		if name != "" {
			return name, true
		}
		return mappedName, true
	}
	return mappedName, true
}

// defaultBytes accepts both forms allowed by authored defaults and returns the
// bytes that the HTTP flag or JSON encoder must receive.
func defaultBytes(value reflect.Value) []byte {
	switch value.Kind() {
	case reflect.String:
		return []byte(value.String())
	case reflect.Slice:
		return value.Bytes()
	default:
		panic(fmt.Sprintf("byte default has Go type %s", value.Type())) // bug
	}
}

// concreteDefaultValue removes interface and pointer wrappers from an authored
// default before the design determines its HTTP representation.
func concreteDefaultValue(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	return value
}

// defaultObjectField returns one authored object field from either a map or a
// Go struct. Missing map entries remain absent from the projected JSON object.
func defaultObjectField(value reflect.Value, name string, attribute *expr.AttributeExpr) (reflect.Value, bool) {
	value = concreteDefaultValue(value)
	switch value.Kind() {
	case reflect.Map:
		iterator := value.MapRange()
		for iterator.Next() {
			key := concreteDefaultValue(iterator.Key())
			if key.Kind() == reflect.String && key.String() == name {
				return iterator.Value(), true
			}
		}
		return reflect.Value{}, false
	case reflect.Struct:
		fieldName := codegen.Goify(name, true)
		if attribute != nil {
			fieldName = codegen.GoifyAtt(attribute, name, true)
		}
		field := value.FieldByName(fieldName)
		return field, field.IsValid()
	default:
		panic(fmt.Sprintf("object default has Go type %s", value.Type())) // bug
	}
}
