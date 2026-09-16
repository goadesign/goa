// This file defines the JSON schema values shared by the OpenAPI generators.
// It also converts Goa examples into the fields visible in those schemas.
package openapi

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strconv"

	"goa.design/goa/v3/expr"
)

type (
	// Schema represents an instance of a JSON schema.
	// See http://json-schema.org/documentation.html
	Schema struct {
		Schema string `json:"$schema,omitempty" yaml:"$schema,omitempty"`
		// Core schema
		ID           string             `json:"id,omitempty" yaml:"id,omitempty"`
		Title        string             `json:"title,omitempty" yaml:"title,omitempty"`
		Type         Type               `json:"type,omitempty" yaml:"type,omitempty"`
		Items        *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
		Properties   map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
		Defs         map[string]*Schema `json:"$defs,omitempty" yaml:"$defs,omitempty"`
		Description  string             `json:"description,omitempty" yaml:"description,omitempty"`
		DefaultValue any                `json:"default,omitempty" yaml:"default,omitempty"`
		Example      any                `json:"example,omitempty" yaml:"example,omitempty"`

		// Hyper schema
		Media     *Media  `json:"media,omitempty" yaml:"media,omitempty"`
		ReadOnly  bool    `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
		PathStart string  `json:"pathStart,omitempty" yaml:"pathStart,omitempty"`
		Links     []*Link `json:"links,omitempty" yaml:"links,omitempty"`
		Ref       string  `json:"$ref,omitempty" yaml:"$ref,omitempty"`

		// Validation
		Enum                 []any    `json:"enum,omitempty" yaml:"enum,omitempty"`
		Format               string   `json:"format,omitempty" yaml:"format,omitempty"`
		Pattern              string   `json:"pattern,omitempty" yaml:"pattern,omitempty"`
		ExclusiveMinimum     *float64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
		Minimum              *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
		ExclusiveMaximum     *float64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
		Maximum              *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
		MinLength            *int     `json:"minLength,omitempty" yaml:"minLength,omitempty"`
		MaxLength            *int     `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
		MinItems             *int     `json:"minItems,omitempty" yaml:"minItems,omitempty"`
		MaxItems             *int     `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
		Required             []string `json:"required,omitempty" yaml:"required,omitempty"`
		AdditionalProperties any      `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

		// Content (JSON Schema 2020-12), used by OpenAPI 3.2 documents to
		// describe string-encoded data such as JSON-encoded SSE event fields.
		ContentMediaType string  `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`
		ContentSchema    *Schema `json:"contentSchema,omitempty" yaml:"contentSchema,omitempty"`

		// Union
		AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`

		// Extensions defines the OpenAPI extensions.
		Extensions map[string]any `json:"-" yaml:"-"`
	}

	// Type is the JSON type enum.
	Type string

	// Media represents a "media" field in a JSON hyper schema.
	Media struct {
		BinaryEncoding string `json:"binaryEncoding,omitempty" yaml:"binaryEncoding,omitempty"`
		Type           string `json:"type,omitempty" yaml:"type,omitempty"`
	}

	// Link represents a "link" field in a JSON hyper schema.
	Link struct {
		Title        string  `json:"title,omitempty" yaml:"title,omitempty"`
		Description  string  `json:"description,omitempty" yaml:"description,omitempty"`
		Rel          string  `json:"rel,omitempty" yaml:"rel,omitempty"`
		Href         string  `json:"href,omitempty" yaml:"href,omitempty"`
		Method       string  `json:"method,omitempty" yaml:"method,omitempty"`
		Schema       *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
		TargetSchema *Schema `json:"targetSchema,omitempty" yaml:"targetSchema,omitempty"`
		ResultType   string  `json:"mediaType,omitempty" yaml:"mediaType,omitempty"`
		EncType      string  `json:"encType,omitempty" yaml:"encType,omitempty"`
	}

	// These types are used in marshalJSON() to avoid recursive call of json.Marshal().
	_Schema Schema
)

const (
	// Array represents a JSON array.
	Array Type = "array"
	// Boolean represents a JSON boolean.
	Boolean = "boolean"
	// Integer represents a JSON number without a fraction or exponent part.
	Integer = "integer"
	// Number represents any JSON number. Number includes integer.
	Number = "number"
	// Null represents the JSON null value.
	Null = "null"
	// Object represents a JSON object.
	Object = "object"
	// String represents a JSON string.
	String = "string"
	// File is an extension used by OpenAPI to represent a file download.
	File = "file"
)

// SchemaRef is the JSON Schema draft 2020-12 meta-schema identifier.
const SchemaRef = "https://json-schema.org/draft/2020-12/schema"

// NewSchema instantiates a new JSON schema.
func NewSchema() *Schema {
	js := Schema{
		Properties: make(map[string]*Schema),
		Defs:       make(map[string]*Schema),
	}
	return &js
}

// JSON serializes the schema into JSON. It makes sure the "$schema" standard
// field is set if needed prior to delegating to the standard JSON marshaler.
func (s *Schema) JSON() ([]byte, error) {
	if s.Ref == "" {
		s.Schema = SchemaRef
	}
	return json.Marshal(s)
}

// ToString returns the string representation of the given type.
func ToString(val any) string {
	switch actual := val.(type) {
	case string:
		return actual
	case int:
		return strconv.Itoa(actual)
	case float64:
		return strconv.FormatFloat(actual, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(actual)
	default:
		panic("unexpected key type")
	}
}

// ToStringMap converts map[any]any to a map[string]any
// when possible.
func ToStringMap(val any) any {
	switch actual := val.(type) {
	case map[any]any:
		m := make(map[string]any)
		for k, v := range actual {
			m[ToString(k)] = ToStringMap(v)
		}
		return m
	case []any:
		mapSlice := make([]any, len(actual))
		for i, e := range actual {
			mapSlice[i] = ToStringMap(e)
		}
		return mapSlice
	default:
		return actual
	}
}

// Example returns an example value projected to the OpenAPI-visible shape of at.
func Example(at *expr.AttributeExpr, r *expr.ExampleGenerator) any {
	return ProjectExample(at, at.Example(r))
}

// ProjectExample removes values for fields that are hidden from the OpenAPI
// schema by metadata while keeping examples usable by service codegen.
func ProjectExample(at *expr.AttributeExpr, val any) any {
	if val == nil {
		return nil
	}
	return projectExample(at.Type, val)
}

// MarshalJSON returns the JSON encoding of s.
func (s *Schema) MarshalJSON() ([]byte, error) {
	return MarshalJSON((*_Schema)(s), s.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (s *Schema) MarshalYAML() (any, error) {
	return MarshalYAML((*_Schema)(s), s.Extensions)
}

// Dup returns an independent copy of the schema. Callers may change any nested
// schema, collection, example, default, or extension without changing s.
func (s *Schema) Dup() *Schema {
	js := Schema{
		Schema:           s.Schema,
		ID:               s.ID,
		Title:            s.Title,
		Type:             s.Type,
		Description:      s.Description,
		DefaultValue:     duplicateJSONValue(s.DefaultValue),
		Example:          duplicateJSONValue(s.Example),
		ReadOnly:         s.ReadOnly,
		PathStart:        s.PathStart,
		Ref:              s.Ref,
		Format:           s.Format,
		Pattern:          s.Pattern,
		ExclusiveMinimum: duplicatePointer(s.ExclusiveMinimum),
		Minimum:          duplicatePointer(s.Minimum),
		ExclusiveMaximum: duplicatePointer(s.ExclusiveMaximum),
		Maximum:          duplicatePointer(s.Maximum),
		MinLength:        duplicatePointer(s.MinLength),
		MaxLength:        duplicatePointer(s.MaxLength),
		MinItems:         duplicatePointer(s.MinItems),
		MaxItems:         duplicatePointer(s.MaxItems),
		Required:         append([]string(nil), s.Required...),
		ContentMediaType: s.ContentMediaType,
	}
	if s.Media != nil {
		media := *s.Media
		js.Media = &media
	}
	if s.Links != nil {
		js.Links = make([]*Link, len(s.Links))
		for index, link := range s.Links {
			copy := *link
			if link.Schema != nil {
				copy.Schema = link.Schema.Dup()
			}
			if link.TargetSchema != nil {
				copy.TargetSchema = link.TargetSchema.Dup()
			}
			js.Links[index] = &copy
		}
	}
	if s.Enum != nil {
		js.Enum = make([]any, len(s.Enum))
		for index, value := range s.Enum {
			js.Enum[index] = duplicateJSONValue(value)
		}
	}
	if additional, ok := s.AdditionalProperties.(*Schema); ok {
		js.AdditionalProperties = additional.Dup()
	} else {
		js.AdditionalProperties = duplicateJSONValue(s.AdditionalProperties)
	}
	if s.Extensions != nil {
		js.Extensions = make(map[string]any, len(s.Extensions))
		for name, value := range s.Extensions {
			js.Extensions[name] = duplicateJSONValue(value)
		}
	}
	if s.ContentSchema != nil {
		js.ContentSchema = s.ContentSchema.Dup()
	}
	if s.Properties != nil {
		js.Properties = make(map[string]*Schema, len(s.Properties))
		for name, property := range s.Properties {
			js.Properties[name] = property.Dup()
		}
	}
	if s.Items != nil {
		js.Items = s.Items.Dup()
	}
	if len(s.AnyOf) > 0 {
		js.AnyOf = make([]*Schema, len(s.AnyOf))
		for i, branch := range s.AnyOf {
			js.AnyOf[i] = branch.Dup()
		}
	}
	if s.Defs != nil {
		js.Defs = make(map[string]*Schema, len(s.Defs))
		for name, definition := range s.Defs {
			js.Defs[name] = definition.Dup()
		}
	}
	return &js
}

// MustGenerate returns true if the meta indicates that a OpenAPI specification should be
// generated, false otherwise.
func MustGenerate(meta expr.MetaExpr) bool {
	m, ok := meta.Last("openapi:generate")
	if !ok {
		m, ok = meta.Last("swagger:generate")
	}
	if ok && m == "false" {
		return false
	}
	return true
}

// AdditionalPropertiesFromExpr extracts the OpenAPI additionalProperties.
func AdditionalPropertiesFromExpr(meta expr.MetaExpr) any {
	m, ok := meta.Last("openapi:additionalProperties")
	if ok && m == "false" {
		return false
	}
	return nil
}

func projectExample(t expr.DataType, val any) any {
	switch actual := t.(type) {
	case expr.Primitive:
		if actual.Kind() == expr.BytesKind {
			return base64.StdEncoding.EncodeToString(reflect.ValueOf(val).Bytes())
		}
		return ToStringMap(val)
	case *expr.UserTypeExpr:
		return ProjectExample(actual.Attribute(), val)
	case *expr.ResultTypeExpr:
		return ProjectExample(actual.Attribute(), val)
	case *expr.Object:
		return projectObjectExample(actual, val)
	case *expr.Array:
		return projectArrayExample(actual, val)
	case *expr.Map:
		return projectMapExample(actual, val)
	default:
		return ToStringMap(val)
	}
}

// duplicateJSONValue copies the maps, slices, arrays, pointers, and interface
// values accepted by JSON fields while preserving their concrete Go types.
func duplicateJSONValue(value any) any {
	if value == nil {
		return nil
	}
	return duplicateJSONReflectValue(reflect.ValueOf(value)).Interface()
}

// duplicateJSONReflectValue recursively copies one reflected JSON value.
func duplicateJSONReflectValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := duplicateJSONReflectValue(value.Elem())
		result := reflect.New(value.Type()).Elem()
		result.Set(copy)
		return result
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.New(value.Type().Elem())
		copy.Elem().Set(duplicateJSONReflectValue(value.Elem()))
		return copy
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			copy.SetMapIndex(
				duplicateJSONReflectValue(iterator.Key()),
				duplicateJSONReflectValue(iterator.Value()),
			)
		}
		return copy
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.MakeSlice(value.Type(), value.Len(), value.Cap())
		for index := range value.Len() {
			copy.Index(index).Set(duplicateJSONReflectValue(value.Index(index)))
		}
		return copy
	case reflect.Array:
		copy := reflect.New(value.Type()).Elem()
		for index := range value.Len() {
			copy.Index(index).Set(duplicateJSONReflectValue(value.Index(index)))
		}
		return copy
	default:
		return value
	}
}

// duplicatePointer copies one scalar schema limit while preserving nil.
func duplicatePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func projectObjectExample(obj *expr.Object, val any) any {
	values, ok := exampleMap(val)
	if !ok {
		return ToStringMap(val)
	}
	projected := make(map[string]any)
	for _, nat := range *obj {
		if !MustGenerate(nat.Attribute.Meta) {
			continue
		}
		if v, ok := values[nat.Name]; ok {
			if projectedValue := ProjectExample(nat.Attribute, v); projectedValue != nil {
				projected[nat.Name] = projectedValue
			}
		}
	}
	return projected
}

func projectArrayExample(array *expr.Array, val any) any {
	values, ok := exampleSlice(val)
	if !ok {
		return ToStringMap(val)
	}
	projected := make([]any, len(values))
	for i, v := range values {
		projected[i] = ProjectExample(array.ElemType, v)
	}
	return projected
}

func projectMapExample(m *expr.Map, val any) any {
	values, ok := exampleMap(val)
	if !ok {
		return ToStringMap(val)
	}
	projected := make(map[string]any, len(values))
	for k, v := range values {
		if projectedValue := ProjectExample(m.ElemType, v); projectedValue != nil {
			projected[k] = projectedValue
		}
	}
	return projected
}

func exampleMap(val any) (map[string]any, bool) {
	switch actual := val.(type) {
	case map[string]any:
		return actual, true
	case map[any]any:
		m := make(map[string]any, len(actual))
		for k, v := range actual {
			m[ToString(k)] = v
		}
		return m, true
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	m := make(map[string]any, rv.Len())
	for _, k := range rv.MapKeys() {
		m[k.String()] = rv.MapIndex(k).Interface()
	}
	return m, true
}

func exampleSlice(val any) ([]any, bool) {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return nil, false
	}
	values := make([]any, rv.Len())
	for i := range values {
		values[i] = rv.Index(i).Interface()
	}
	return values, true
}
