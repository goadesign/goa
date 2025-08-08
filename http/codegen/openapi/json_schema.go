package openapi

import (
	"encoding/json"
	"fmt"
	"strconv"

	"goa.design/goa/v3/codegen"
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
		Type         any                `json:"type,omitempty" yaml:"type,omitempty"` // Can be string or []string for OpenAPI 3.1
		Items        *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
		Properties   map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
		Definitions  map[string]*Schema `json:"definitions,omitempty" yaml:"definitions,omitempty"`
		Description  string             `json:"description,omitempty" yaml:"description,omitempty"`
		DefaultValue any                `json:"default,omitempty" yaml:"default,omitempty"`
		Examples     []any              `json:"examples,omitempty" yaml:"examples,omitempty"` // OpenAPI 3.1 uses examples array
		Example      any                `json:"example,omitempty" yaml:"example,omitempty"`   // Keep for backward compatibility

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

		// Union
		AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`

		// Content encoding and media type (OpenAPI 3.1)
		ContentEncoding  string `json:"contentEncoding,omitempty" yaml:"contentEncoding,omitempty"`
		ContentMediaType string `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`

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
)

// _Schema is a type alias used in marshalJSON() to avoid recursive call of json.Marshal().
type _Schema Schema

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

// SchemaRef is the JSON Hyper-schema standard href.
// Updated to JSON Schema 2020-12 for OpenAPI 3.1 compatibility
const SchemaRef = "https://json-schema.org/draft/2020-12/schema"

var (
	// Definitions contains the generated JSON schema definitions
	Definitions map[string]*Schema
	
	// GenerateForOpenAPIv2 indicates whether we're generating for OpenAPI v2 (Swagger)
	// When true, schemas will use v2 compatible format (single example, string type)
	// When false, schemas will use v3.1 format (examples array, type arrays for nullable)
	GenerateForOpenAPIv2 bool
)

// Initialize the global variables
func init() {
	Definitions = make(map[string]*Schema)
	GenerateForOpenAPIv2 = false // Default to v3 behavior
}

// SetType sets the type of the schema. It handles both string and array types
// for OpenAPI 3.1 compatibility.
func (s *Schema) SetType(t Type) {
	s.Type = string(t)
}

// SetNullableType sets the type to an array including "null" for OpenAPI 3.1 nullable support.
// For example, SetNullableType(String) results in ["string", "null"].
func (s *Schema) SetNullableType(t Type) {
	s.Type = []any{string(t), "null"}
}

// AddType adds a type to the schema's type array. If Type is currently a string,
// it converts it to an array first.
func (s *Schema) AddType(t Type) {
	switch current := s.Type.(type) {
	case string:
		s.Type = []any{current, string(t)}
	case []any:
		s.Type = append(current, string(t))
	case nil:
		s.Type = string(t)
	default:
		s.Type = []any{string(t)}
	}
}

// IsNullable returns true if the schema allows null values.
func (s *Schema) IsNullable() bool {
	switch t := s.Type.(type) {
	case []any:
		for _, v := range t {
			if v == "null" {
				return true
			}
		}
	}
	return false
}

// NewSchema instantiates a new JSON schema.
func NewSchema() *Schema {
	js := Schema{
		Properties:  make(map[string]*Schema),
		Definitions: make(map[string]*Schema),
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

// MarshalJSON implements json.Marshaler to handle v2 vs v3 differences
func (s *Schema) MarshalJSON() ([]byte, error) {
	// For OpenAPI v2, we need to adjust the output
	if GenerateForOpenAPIv2 {
		// Make a copy to avoid modifying the original
		copy := *s
		
		// Convert Type array back to string for v2
		if typeArray, ok := copy.Type.([]any); ok && len(typeArray) > 0 {
			// Get the first non-null type
			for _, t := range typeArray {
				if str, ok := t.(string); ok && str != "null" {
					copy.Type = str
					break
				}
			}
		}
		
		// Convert Examples array to single Example for v2
		if len(copy.Examples) > 0 && copy.Example == nil {
			copy.Example = copy.Examples[0]
			copy.Examples = nil
		}
		
		return MarshalJSON((*_Schema)(&copy), copy.Extensions)
	}
	
	// For v3, clear Example field if Examples is set
	if len(s.Examples) > 0 {
		copy := *s
		copy.Example = nil // Don't output both example and examples in v3
		return MarshalJSON((*_Schema)(&copy), copy.Extensions)
	}
	
	return MarshalJSON((*_Schema)(s), s.Extensions)
}

// MarshalYAML implements yaml.Marshaler to handle v2 vs v3 differences
func (s *Schema) MarshalYAML() (interface{}, error) {
	// For OpenAPI v2, we need to adjust the output
	if GenerateForOpenAPIv2 {
		// Make a copy to avoid modifying the original
		copy := *s
		
		// Convert Type array back to string for v2
		if typeArray, ok := copy.Type.([]any); ok && len(typeArray) > 0 {
			// Get the first non-null type
			for _, t := range typeArray {
				if str, ok := t.(string); ok && str != "null" {
					copy.Type = str
					break
				}
			}
		}
		
		// Convert Examples array to single Example for v2
		if len(copy.Examples) > 0 && copy.Example == nil {
			copy.Example = copy.Examples[0]
			copy.Examples = nil
		}
		
		return MarshalYAML((*_Schema)(&copy), copy.Extensions)
	}
	
	// For v3, clear Example field if Examples is set
	if len(s.Examples) > 0 {
		copy := *s
		copy.Example = nil // Don't output both example and examples in v3
		return MarshalYAML((*_Schema)(&copy), copy.Extensions)
	}
	
	return MarshalYAML((*_Schema)(s), s.Extensions)
}

// APISchema produces the API JSON hyper schema.
func APISchema(api *expr.APIExpr, r *expr.RootExpr) *Schema {
	for _, res := range r.API.HTTP.Services {
		GenerateServiceDefinition(api, res)
	}
	href := string(api.Servers[0].Hosts[0].URIs[0])
	links := []*Link{
		{
			Href: href,
			Rel:  "self",
		},
		{
			Href:   "/schema",
			Method: "GET",
			Rel:    "self",
			TargetSchema: &Schema{
				Schema:               SchemaRef,
				AdditionalProperties: true,
			},
		},
	}
	s := Schema{
		ID:          fmt.Sprintf("%s/schema", href),
		Title:       api.Title,
		Description: api.Description,
		Definitions: Definitions,
		Properties:  propertiesFromDefs(Definitions, "#/definitions/"),
		Links:       links,
	}
	s.SetType(Object)
	return &s
}

// GenerateServiceDefinition produces the JSON schema corresponding to the given
// service. It stores the results in Definitions.
func GenerateServiceDefinition(api *expr.APIExpr, res *expr.HTTPServiceExpr) {
	s := NewSchema()
	s.Description = res.Description()
	s.SetType(Object)
	s.Title = res.Name()
	Definitions[res.Name()] = s
	for _, a := range res.HTTPEndpoints {
		var requestSchema *Schema
		if a.MethodExpr.Payload.Type != expr.Empty {
			requestSchema = AttributeTypeSchema(api, a.MethodExpr.Payload)
			requestSchema.Description = a.Name() + " payload"
		}
		var targetSchema *Schema
		var identifier string
		for _, resp := range a.Responses {
			dt := resp.Body.Type
			if mt := dt.(*expr.ResultTypeExpr); mt != nil {
				if identifier == "" {
					identifier = mt.Identifier
				} else {
					identifier = ""
				}
				switch {
				case targetSchema == nil:
					targetSchema = TypeSchemaWithPrefix(api, mt, a.Name())
				case targetSchema.AnyOf == nil:
					firstSchema := targetSchema
					targetSchema = NewSchema()
					targetSchema.AnyOf = []*Schema{firstSchema, TypeSchemaWithPrefix(api, mt, a.Name())}
				default:
					targetSchema.AnyOf = append(targetSchema.AnyOf, TypeSchemaWithPrefix(api, mt, a.Name()))
				}
			}
		}
		for i, r := range a.Routes {
			for j, href := range toSchemaHrefs(r) {
				link := Link{
					Title:        a.Name(),
					Rel:          a.Name(),
					Href:         href,
					Method:       r.Method,
					Schema:       requestSchema,
					TargetSchema: targetSchema,
					ResultType:   identifier,
				}
				if i == 0 && j == 0 {
					if ca := a.Service.CanonicalEndpoint(); ca != nil {
						if ca.Name() == a.Name() {
							link.Rel = "self"
						}
					}
				}
				s.Links = append(s.Links, &link)
			}
		}
	}
}

// ResultTypeRef produces the JSON reference to the media type definition with
// the given view.
func ResultTypeRef(api *expr.APIExpr, mt *expr.ResultTypeExpr, view string) string {
	return ResultTypeRefWithPrefix(api, mt, view, "")
}

// ResultTypeRefWithPrefix produces the JSON reference to the media type definition with
// the given view and adds the provided prefix to the type name
func ResultTypeRefWithPrefix(api *expr.APIExpr, mt *expr.ResultTypeExpr, view, prefix string) string {
	projected, err := expr.Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	var metaName string
	if n, ok := mt.Meta["openapi:typename"]; ok {
		metaName = codegen.Goify(n[0], true)
	}
	if metaName != "" {
		projected.TypeName = metaName
	}
	if _, ok := Definitions[projected.TypeName]; !ok {
		projected.TypeName = codegen.Goify(prefix, true) + codegen.Goify(projected.TypeName, true)
		if metaName != "" {
			projected.TypeName = metaName
		}
		GenerateResultTypeDefinition(api, projected, expr.DefaultView)
	}
	return fmt.Sprintf("#/definitions/%s", projected.TypeName)
}

// TypeRef produces the JSON reference to the type definition.
func TypeRef(api *expr.APIExpr, ut *expr.UserTypeExpr) string {
	return TypeRefWithPrefix(api, ut, "")
}

// TypeRefWithPrefix produces the JSON reference to the type definition and adds the provided prefix
// to the type name
func TypeRefWithPrefix(api *expr.APIExpr, ut *expr.UserTypeExpr, prefix string) string {
	typeName := ut.TypeName
	if prefix != "" {
		typeName = codegen.Goify(prefix, true) + codegen.Goify(ut.TypeName, true)
	}
	if n, ok := ut.Meta["openapi:typename"]; ok {
		typeName = codegen.Goify(n[0], true)
	}
	if _, ok := Definitions[typeName]; !ok {
		GenerateTypeDefinitionWithName(api, ut, typeName)
	}
	return fmt.Sprintf("#/definitions/%s", typeName)
}

// GenerateResultTypeDefinition produces the JSON schema corresponding to the
// given media type and given view.
func GenerateResultTypeDefinition(api *expr.APIExpr, mt *expr.ResultTypeExpr, view string) {
	if _, ok := Definitions[mt.TypeName]; ok {
		return
	}
	s := NewSchema()
	s.Title = fmt.Sprintf("Mediatype identifier: %s", mt.Identifier)
	Definitions[mt.TypeName] = s
	buildResultTypeSchema(api, mt, view, s)
}

// GenerateTypeDefinition produces the JSON schema corresponding to the given
// type.
func GenerateTypeDefinition(api *expr.APIExpr, ut *expr.UserTypeExpr) {
	GenerateTypeDefinitionWithName(api, ut, ut.TypeName)
}

// GenerateTypeDefinitionWithName produces the JSON schema corresponding to the given
// type with provided type name.
func GenerateTypeDefinitionWithName(api *expr.APIExpr, ut *expr.UserTypeExpr, typeName string) {
	if _, ok := Definitions[typeName]; ok {
		return
	}
	s := NewSchema()

	s.Title = typeName
	Definitions[typeName] = s
	buildAttributeSchema(api, s, ut.AttributeExpr)
}

// TypeSchema produces the JSON schema corresponding to the given data type.
func TypeSchema(api *expr.APIExpr, t expr.DataType) *Schema {
	return TypeSchemaWithPrefix(api, t, "")
}

// TypeSchemaWithPrefix produces the JSON schema corresponding to the given data type
// and adds the provided prefix to the type name
func TypeSchemaWithPrefix(api *expr.APIExpr, t expr.DataType, prefix string) *Schema {
	s := NewSchema()
	switch actual := t.(type) {
	case expr.Primitive:
		s.SetType(Type(actual.Name()))
		switch actual.Kind() {
		case expr.AnyKind:
			// A schema without a type matches any data type.
			// See https://swagger.io/docs/specification/data-models/data-types/#any.
			s.SetType(Type(""))
		case expr.IntKind, expr.Int64Kind,
			expr.UIntKind, expr.UInt64Kind:
			// Use int64 format for IntKind and UIntKind because the OpenAPI
			// generator produced int32 by default.
			s.SetType(Type("integer"))
			s.Format = "int64"
		case expr.Int32Kind, expr.UInt32Kind:
			s.SetType(Type("integer"))
			s.Format = "int32"
		case expr.Float32Kind:
			s.SetType(Type("number"))
			s.Format = "float"
		case expr.Float64Kind:
			s.SetType(Type("number"))
			s.Format = "double"
		case expr.BytesKind:
			s.SetType(Type("string"))
			s.Format = "byte"
		}
	case *expr.Array:
		s.SetType(Array)
		s.Items = NewSchema()
		buildAttributeSchema(api, s.Items, actual.ElemType)
	case *expr.Object:
		s.SetType(Object)
		for _, nat := range *actual {
			if !MustGenerate(nat.Attribute.Meta) {
				continue
			}
			prop := NewSchema()
			buildAttributeSchema(api, prop, nat.Attribute)
			s.Properties[nat.Name] = prop
		}
	case *expr.Map:
		s.SetType(Object)
		if actual.KeyType.Type == expr.String && actual.ElemType.Type != expr.Any {
			// Use free-form objects when elements are of type "Any"
			additionalProperties := NewSchema()
			s.AdditionalProperties = buildAttributeSchema(api, additionalProperties, actual.ElemType)
		} else {
			s.AdditionalProperties = true
		}
	case *expr.Union:
		for _, val := range actual.Values {
			s.AnyOf = append(s.AnyOf, AttributeTypeSchemaWithPrefix(api, val.Attribute, prefix))
		}
	case *expr.UserTypeExpr:
		s.Ref = TypeRefWithPrefix(api, actual, prefix)
	case *expr.ResultTypeExpr:
		// Use "default" view by default
		s.Ref = ResultTypeRefWithPrefix(api, actual, expr.DefaultView, prefix)
	}
	return s
}

// AttributeTypeSchema produces the JSON schema corresponding to the given attribute.
func AttributeTypeSchema(api *expr.APIExpr, at *expr.AttributeExpr) *Schema {
	return AttributeTypeSchemaWithPrefix(api, at, "")
}

// AttributeTypeSchemaWithPrefix produces the JSON schema corresponding to the given attribute
// and adds the provided prefix to the type name
func AttributeTypeSchemaWithPrefix(api *expr.APIExpr, at *expr.AttributeExpr, prefix string) *Schema {
	s := TypeSchemaWithPrefix(api, at.Type, prefix)
	initAttributeValidation(s, at)
	return s
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


// Dup creates a shallow clone of the given schema.
func (s *Schema) Dup() *Schema {
	js := Schema{
		ID:                   s.ID,
		Description:          s.Description,
		Schema:               s.Schema,
		Type:                 s.Type,
		DefaultValue:         s.DefaultValue,
		Title:                s.Title,
		Media:                s.Media,
		ReadOnly:             s.ReadOnly,
		PathStart:            s.PathStart,
		Links:                s.Links,
		Ref:                  s.Ref,
		Enum:                 s.Enum,
		Format:               s.Format,
		Pattern:              s.Pattern,
		Minimum:              s.Minimum,
		Maximum:              s.Maximum,
		MinLength:            s.MinLength,
		MaxLength:            s.MaxLength,
		MinItems:             s.MinItems,
		MaxItems:             s.MaxItems,
		Required:             s.Required,
		AdditionalProperties: s.AdditionalProperties,
	}
	for n, p := range s.Properties {
		js.Properties[n] = p.Dup()
	}
	if s.Items != nil {
		js.Items = s.Items.Dup()
	}
	for n, d := range s.Definitions {
		js.Definitions[n] = d.Dup()
	}
	return &js
}

// buildAttributeSchema initializes the given JSON schema that corresponds to
// the given attribute.
func buildAttributeSchema(api *expr.APIExpr, s *Schema, at *expr.AttributeExpr) *Schema {
	s.Merge(TypeSchema(api, at.Type))
	if s.Ref != "" {
		// Ref is exclusive with other fields
		return s
	}
	s.DefaultValue = ToStringMap(at.DefaultValue)
	s.Description = at.Description
	// Handle examples based on OpenAPI version
	if example := at.Example(api.ExampleGenerator); example != nil {
		if GenerateForOpenAPIv2 {
			// For v2, use single Example field
			s.Example = example
		} else {
			// For v3.1, use Examples array
			s.Examples = []any{example}
		}
	}
	s.Extensions = ExtensionsFromExpr(at.Meta)
	if ap := AdditionalPropertiesFromExpr(at.Meta); ap != nil {
		s.AdditionalProperties = ap
	}
	initAttributeValidation(s, at)

	return s
}

// initAttributeValidation initializes validation rules for an attribute.
func initAttributeValidation(s *Schema, at *expr.AttributeExpr) {
	val := at.Validation
	if val == nil {
		return
	}
	s.Enum = val.Values
	if val.Format != "" {
		s.Format = string(val.Format)
	}
	s.Pattern = val.Pattern
	if val.ExclusiveMinimum != nil {
		s.ExclusiveMinimum = val.ExclusiveMinimum
	}
	if val.Minimum != nil {
		s.Minimum = val.Minimum
	}
	if val.ExclusiveMaximum != nil {
		s.ExclusiveMaximum = val.ExclusiveMaximum
	}
	if val.Maximum != nil {
		s.Maximum = val.Maximum
	}
	if val.MinLength != nil {
		if _, ok := at.Type.(*expr.Array); ok {
			s.MinItems = val.MinLength
		} else {
			s.MinLength = val.MinLength
		}
	}
	if val.MaxLength != nil {
		if _, ok := at.Type.(*expr.Array); ok {
			s.MaxItems = val.MaxLength
		} else {
			s.MaxLength = val.MaxLength
		}
	}
	for _, v := range val.Required {
		if a := at.Find(v); a != nil {
			if !MustGenerate(a.Meta) {
				continue
			}
		}
		s.Required = append(s.Required, v)
	}
}

// toSchemaHrefs produces hrefs that replace the path wildcards with JSON
// schema references when appropriate.
func toSchemaHrefs(r *expr.RouteExpr) []string {
	paths := r.FullPaths()
	res := make([]string, len(paths))
	for i, path := range paths {
		params := expr.ExtractHTTPWildcards(path)
		args := make([]any, len(params))
		for j, p := range params {
			args[j] = fmt.Sprintf("/{%s}", p)
		}
		tmpl := expr.HTTPWildcardRegex.ReplaceAllLiteralString(path, "%s")
		res[i] = fmt.Sprintf(tmpl, args...)
	}
	return res
}

// propertiesFromDefs creates a Properties map referencing the given definitions
// under the given path.
func propertiesFromDefs(definitions map[string]*Schema, path string) map[string]*Schema {
	res := make(map[string]*Schema, len(definitions))
	for n := range definitions {
		if n == "identity" {
			continue
		}
		s := NewSchema()
		s.Ref = path + n
		res[n] = s
	}
	return res
}

// buildResultTypeSchema initializes s as the JSON schema representing mt for the
// given view.
func buildResultTypeSchema(api *expr.APIExpr, mt *expr.ResultTypeExpr, view string, s *Schema) {
	s.Media = &Media{Type: mt.Identifier}
	projected, err := expr.Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	buildAttributeSchema(api, s, projected.AttributeExpr)
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
