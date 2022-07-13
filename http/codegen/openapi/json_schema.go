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
		Type         Type               `json:"type,omitempty" yaml:"type,omitempty"`
		Items        *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
		Properties   map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
		Definitions  map[string]*Schema `json:"definitions,omitempty" yaml:"definitions,omitempty"`
		Description  string             `json:"description,omitempty" yaml:"description,omitempty"`
		DefaultValue interface{}        `json:"default,omitempty" yaml:"default,omitempty"`
		Example      interface{}        `json:"example,omitempty" yaml:"example,omitempty"`

		// Hyper schema
		Media     *Media  `json:"media,omitempty" yaml:"media,omitempty"`
		ReadOnly  bool    `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
		PathStart string  `json:"pathStart,omitempty" yaml:"pathStart,omitempty"`
		Links     []*Link `json:"links,omitempty" yaml:"links,omitempty"`
		Ref       string  `json:"$ref,omitempty" yaml:"$ref,omitempty"`

		// Validation
		Enum                 []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
		Format               string        `json:"format,omitempty" yaml:"format,omitempty"`
		Pattern              string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
		ExclusiveMinimum     *float64      `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
		Minimum              *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
		ExclusiveMaximum     *float64      `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
		Maximum              *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
		MinLength            *int          `json:"minLength,omitempty" yaml:"minLength,omitempty"`
		MaxLength            *int          `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
		MinItems             *int          `json:"minItems,omitempty" yaml:"minItems,omitempty"`
		MaxItems             *int          `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
		Required             []string      `json:"required,omitempty" yaml:"required,omitempty"`
		AdditionalProperties interface{}   `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

		// Union
		AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`

		// Extensions defines the OpenAPI extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
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

// SchemaRef is the JSON Hyper-schema standard href.
const SchemaRef = "http://json-schema.org/draft-04/hyper-schema"

var (
	// Definitions contains the generated JSON schema definitions
	Definitions map[string]*Schema
)

// Initialize the global variables
func init() {
	Definitions = make(map[string]*Schema)
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
		Type:        Object,
		Definitions: Definitions,
		Properties:  propertiesFromDefs(Definitions, "#/definitions/"),
		Links:       links,
	}
	return &s
}

// GenerateServiceDefinition produces the JSON schema corresponding to the given
// service. It stores the results in cachedSchema.
func GenerateServiceDefinition(api *expr.APIExpr, res *expr.HTTPServiceExpr) {
	s := NewSchema()
	s.Description = res.Description()
	s.Type = Object
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
				if targetSchema == nil {
					targetSchema = TypeSchemaWithPrefix(api, mt, a.Name())
				} else if targetSchema.AnyOf == nil {
					firstSchema := targetSchema
					targetSchema = NewSchema()
					targetSchema.AnyOf = []*Schema{firstSchema, TypeSchemaWithPrefix(api, mt, a.Name())}
				} else {
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
func ResultTypeRefWithPrefix(api *expr.APIExpr, mt *expr.ResultTypeExpr, view string, prefix string) string {
	projected, err := expr.Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	if _, ok := Definitions[projected.TypeName]; !ok {
		projected.TypeName = codegen.Goify(prefix, true) + codegen.Goify(projected.TypeName, true)
		GenerateResultTypeDefinition(api, projected, "default")
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
		s.Type = Type(actual.Name())
		switch actual.Kind() {
		case expr.AnyKind:
			s.Type = Type("string")
			s.Format = "binary"
		case expr.IntKind, expr.Int64Kind,
			expr.UIntKind, expr.UInt64Kind:
			s.Type = Type("integer")
			s.Format = "int64"
		case expr.Int32Kind, expr.UInt32Kind:
			s.Type = Type("integer")
			s.Format = "int32"
		case expr.Float32Kind:
			s.Type = Type("number")
			s.Format = "float"
		case expr.Float64Kind:
			s.Type = Type("number")
			s.Format = "double"
		case expr.BytesKind:
			s.Type = Type("string")
			s.Format = "byte"
		}
	case *expr.Array:
		s.Type = Array
		s.Items = NewSchema()
		buildAttributeSchema(api, s.Items, actual.ElemType)
	case *expr.Object:
		s.Type = Object
		for _, nat := range *actual {
			prop := NewSchema()
			buildAttributeSchema(api, prop, nat.Attribute)
			s.Properties[nat.Name] = prop
		}
	case *expr.Map:
		s.Type = Object
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
func ToString(val interface{}) string {
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

// ToStringMap converts map[interface{}]interface{} to a map[string]interface{}
// when possible.
func ToStringMap(val interface{}) interface{} {
	switch actual := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range actual {
			m[ToString(k)] = ToStringMap(v)
		}
		return m
	case []interface{}:
		mapSlice := make([]interface{}, len(actual))
		for i, e := range actual {
			mapSlice[i] = ToStringMap(e)
		}
		return mapSlice
	default:
		return actual
	}
}

// MarshalJSON returns the JSON encoding of s.
func (s *Schema) MarshalJSON() ([]byte, error) {
	return MarshalJSON((*_Schema)(s), s.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (s *Schema) MarshalYAML() (interface{}, error) {
	return MarshalYAML((*_Schema)(s), s.Extensions)
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
	s.Example = at.Example(api.Random())
	s.Extensions = ExtensionsFromExpr(at.Meta)
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
	s.Format = string(val.Format)
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
	s.Required = val.Required
}

// toSchemaHrefs produces hrefs that replace the path wildcards with JSON
// schema references when appropriate.
func toSchemaHrefs(r *expr.RouteExpr) []string {
	paths := r.FullPaths()
	res := make([]string, len(paths))
	for i, path := range paths {
		params := expr.ExtractHTTPWildcards(path)
		args := make([]interface{}, len(params))
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
