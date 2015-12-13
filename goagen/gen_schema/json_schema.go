package genschema

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/raphael/goa/design"
)

type (
	// JSONSchema represents an instance of a JSON schema.
	// See http://json-schema.org/documentation.html
	JSONSchema struct {
		Schema string `json:"$schema,omitempty"`
		// Core schema
		ID           string                 `json:"id,omitempty"`
		Title        string                 `json:"title,omitempty"`
		Type         JSONType               `json:"type,omitempty"`
		Items        *JSONSchema            `json:"items,omitempty"`
		Properties   map[string]*JSONSchema `json:"properties,omitempty"`
		Definitions  map[string]*JSONSchema `json:"definitions,omitempty"`
		Description  string                 `json:"description,omitempty"`
		DefaultValue interface{}            `json:"defaultValue,omitempty"`

		// Hyper schema
		Media     *JSONMedia  `json:"media,omitempty"`
		ReadOnly  bool        `json:"readOnly,omitempty"`
		PathStart string      `json:"pathStart,omitempty"`
		Links     []*JSONLink `json:"links,omitempty"`
		Ref       string      `json:"$ref,omitempty"`

		// Validation
		Enum                 []interface{} `json:"enum,omitempty"`
		Format               string        `json:"format,omitempty"`
		Pattern              string        `json:"pattern,omitempty"`
		Minimum              float64       `json:"minimum,omitempty"`
		Maximum              float64       `json:"maximum,omitempty"`
		MinLength            int           `json:"minLength,omitempty"`
		MaxLength            int           `json:"maxLength,omitempty"`
		Required             []string      `json:"required,omitempty"`
		AdditionalProperties bool          `json:"additionalProperties,omitempty"`

		// Union
		AnyOf []*JSONSchema `json:"anyOf,omitempty"`
	}

	// JSONType is the JSON type enum.
	JSONType string

	// JSONMedia represents a "media" field in a JSON hyper schema.
	JSONMedia struct {
		BinaryEncoding string `json:"binaryEncoding,omitempty"`
		Type           string `json:"type,omitempty"`
	}

	// JSONLink represents a "link" field in a JSON hyper schema.
	JSONLink struct {
		Title        string      `json:"title,omitempty"`
		Description  string      `json:"description,omitempty"`
		Rel          string      `json:"rel,omitempty"`
		Href         string      `json:"href,omitempty"`
		Method       string      `json:"method,omitempty"`
		Schema       *JSONSchema `json:"schema,omitempty"`
		TargetSchema *JSONSchema `json:"targetSchema,omitempty"`
		MediaType    string      `json:"mediaType,omitempty"`
		EncType      string      `json:"encType,omitempty"`
	}
)

const (
	// JSONArray represents a JSON array.
	JSONArray JSONType = "array"
	// JSONBoolean represents a JSON boolean.
	JSONBoolean = "boolean"
	// JSONInteger represents a JSON number without a fraction or exponent part.
	JSONInteger = "integer"
	// JSONNumber represents any JSON number. Number includes integer.
	JSONNumber = "number"
	// JSONNull represents the JSON null value.
	JSONNull = "null"
	// JSONObject represents a JSON object.
	JSONObject = "object"
	// JSONString represents a JSON string.
	JSONString = "string"
)

// SchemaRef is the JSON Hyper-schema standard href.
const SchemaRef = "http://json-schema.org/draft-04/hyper-schema"

var (
	// Definitions contains the generated JSON schema definitions
	Definitions map[string]*JSONSchema
)

// Initialize the global variables
func init() {
	Definitions = make(map[string]*JSONSchema)
}

// NewJSONSchema instantiates a new JSON schema.
func NewJSONSchema() *JSONSchema {
	js := JSONSchema{
		Properties:  make(map[string]*JSONSchema),
		Definitions: make(map[string]*JSONSchema),
	}
	return &js
}

// JSON serializes the schema into JSON.
// It makes sure the "$schema" standard field is set if needed prior to delegating to the standard
// JSON marshaler.
func (s *JSONSchema) JSON() ([]byte, error) {
	if s.Ref == "" {
		s.Schema = SchemaRef
	}
	return json.Marshal(s)
}

// APISchema produces the API JSON hyper schema.
func APISchema(api *design.APIDefinition) *JSONSchema {
	api.IterateResources(func(r *design.ResourceDefinition) error {
		GenerateResourceDefinition(api, r)
		return nil
	})
	links := []*JSONLink{
		&JSONLink{
			Href: ServiceURL,
			Rel:  "self",
		},
		&JSONLink{
			Href:   "/schema",
			Method: "GET",
			Rel:    "self",
			TargetSchema: &JSONSchema{
				Schema:               SchemaRef,
				AdditionalProperties: true,
			},
		},
	}
	s := JSONSchema{
		ID:          fmt.Sprintf("%s/schema", ServiceURL),
		Title:       api.Title,
		Description: api.Description,
		Type:        JSONObject,
		Definitions: Definitions,
		Properties:  propertiesFromDefs(Definitions, "#/definitions/"),
		Links:       links,
	}
	return &s
}

// GenerateResourceDefinition produces the JSON schema corresponding to the given API resource.
// It stores the results in cachedSchema.
func GenerateResourceDefinition(api *design.APIDefinition, r *design.ResourceDefinition) {
	s := NewJSONSchema()
	s.Description = r.Description
	s.Type = JSONObject
	s.Title = r.Name
	Definitions[r.Name] = s
	if mt, ok := api.MediaTypes[r.MediaType]; ok {
		buildMediaTypeSchema(api, mt, s)
	}
	r.IterateActions(func(a *design.ActionDefinition) error {
		var requestSchema *JSONSchema
		if a.Payload != nil {
			requestSchema = TypeSchema(api, a.Payload)
			requestSchema.Description = a.Name + " payload"
		}
		if a.Params != nil {
			params := a.Params.Dup()
			// We don't want to keep the path params, these are defined inline in the href
			for _, r := range a.Routes {
				for _, p := range r.Params() {
					delete(params.Type.ToObject(), p)
				}
			}
		}
		var targetSchema *JSONSchema
		var identifier string
		for _, resp := range a.Responses {
			if mt, ok := api.MediaTypes[resp.MediaType]; ok {
				if identifier == "" {
					identifier = mt.Identifier
				} else {
					identifier = ""
				}
				if targetSchema == nil {
					targetSchema = TypeSchema(api, mt)
				} else if targetSchema.AnyOf == nil {
					firstSchema := targetSchema
					targetSchema = NewJSONSchema()
					targetSchema.AnyOf = []*JSONSchema{firstSchema, TypeSchema(api, mt)}
				} else {
					targetSchema.AnyOf = append(targetSchema.AnyOf, TypeSchema(api, mt))
				}
			}
		}
		for i, r := range a.Routes {
			link := JSONLink{
				Title:        a.Name,
				Rel:          a.Name,
				Href:         toSchemaHref(api, r),
				Method:       r.Verb,
				Schema:       requestSchema,
				TargetSchema: targetSchema,
				MediaType:    identifier,
			}
			if i == 0 {
				if ca := a.Parent.CanonicalAction(); ca != nil {
					if ca.Name == a.Name {
						link.Rel = "self"
					}
				}
			}
			s.Links = append(s.Links, &link)
		}
		return nil
	})
}

// MediaTypeRef produces the JSON reference to the media type definition.
func MediaTypeRef(api *design.APIDefinition, mt *design.MediaTypeDefinition) string {
	if _, ok := Definitions[mt.TypeName]; !ok {
		GenerateMediaTypeDefinition(api, mt)
	}
	return fmt.Sprintf("#/definitions/%s", mt.TypeName)
}

// TypeRef produces the JSON reference to the type definition.
func TypeRef(api *design.APIDefinition, ut *design.UserTypeDefinition) string {
	if _, ok := Definitions[ut.TypeName]; !ok {
		GenerateTypeDefinition(api, ut)
	}
	return fmt.Sprintf("#/definitions/%s", ut.TypeName)
}

// GenerateMediaTypeDefinition produces the JSON schema corresponding to the given media type.
func GenerateMediaTypeDefinition(api *design.APIDefinition, mt *design.MediaTypeDefinition) {
	if _, ok := Definitions[mt.TypeName]; ok {
		return
	}
	s := NewJSONSchema()
	s.Title = fmt.Sprintf("Mediatype identifier: %s", mt.Identifier)
	Definitions[mt.TypeName] = s
	buildMediaTypeSchema(api, mt, s)
}

// GenerateTypeDefinition produces the JSON schema corresponding to the given type.
func GenerateTypeDefinition(api *design.APIDefinition, ut *design.UserTypeDefinition) {
	if _, ok := Definitions[ut.TypeName]; ok {
		return
	}
	s := NewJSONSchema()
	s.Title = ut.TypeName
	Definitions[ut.TypeName] = s
	buildAttributeSchema(api, s, ut.AttributeDefinition)
}

// TypeSchema produces the JSON schema corresponding to the given data type.
func TypeSchema(api *design.APIDefinition, t design.DataType) *JSONSchema {
	s := NewJSONSchema()
	switch actual := t.(type) {
	case design.Primitive:
		s.Type = JSONType(actual.Name())
	case *design.Array:
		s.Type = JSONArray
		s.Items = NewJSONSchema()
		buildAttributeSchema(api, s.Items, actual.ElemType)
	case design.Object:
		s.Type = JSONObject
		for n, at := range actual {
			prop := NewJSONSchema()
			buildAttributeSchema(api, prop, at)
			s.Properties[n] = prop
		}
	case *design.Hash:
		s.Type = JSONObject
		s.AdditionalProperties = true
	case *design.UserTypeDefinition:
		s.Ref = TypeRef(api, actual)
	case *design.MediaTypeDefinition:
		s.Ref = MediaTypeRef(api, actual)
	}
	return s
}

// Merge does a two level deep merge of other into s.
func (s *JSONSchema) Merge(other *JSONSchema) {
	for _, v := range []struct {
		a, b   interface{}
		needed bool
	}{
		{&s.ID, other.ID, s.ID == ""},
		{&s.Type, other.Type, s.Type == ""},
		{&s.Ref, other.Ref, s.Ref == ""},
		{&s.Items, other.Items, s.Items == nil},
		{&s.DefaultValue, other.DefaultValue, s.DefaultValue == nil},
		{&s.Title, other.Title, s.Title == ""},
		{&s.Media, other.Media, s.Media == nil},
		{&s.ReadOnly, other.ReadOnly, s.ReadOnly == false},
		{&s.PathStart, other.PathStart, s.PathStart == ""},
		{&s.Enum, other.Enum, s.Enum == nil},
		{&s.Format, other.Format, s.Format == ""},
		{&s.Pattern, other.Pattern, s.Pattern == ""},
		{&s.AdditionalProperties, other.AdditionalProperties, s.AdditionalProperties == false},
		{&s.Minimum, other.Minimum, s.Minimum > other.Minimum},
		{&s.Maximum, other.Maximum, s.Maximum < other.Maximum},
		{&s.MinLength, other.MinLength, s.MinLength > other.MinLength},
		{&s.MaxLength, other.MaxLength, s.MaxLength < other.MaxLength},
	} {
		if v.needed && v.b != nil {
			reflect.Indirect(reflect.ValueOf(v.a)).Set(reflect.ValueOf(v.b))
		}
	}

	for n, p := range other.Properties {
		if _, ok := s.Properties[n]; !ok {
			if s.Properties == nil {
				s.Properties = make(map[string]*JSONSchema)
			}
			s.Properties[n] = p
		}
	}

	for n, d := range other.Definitions {
		if _, ok := s.Definitions[n]; !ok {
			s.Definitions[n] = d
		}
	}

	for _, l := range other.Links {
		s.Links = append(s.Links, l)
	}

	for _, r := range other.Required {
		s.Required = append(s.Required, r)
	}
}

// Dup creates a shallow clone of the given schema.
func (s *JSONSchema) Dup() *JSONSchema {
	js := JSONSchema{
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

// buildAttributeSchema initializes the given JSON schema that corresponds to the given attribute.
func buildAttributeSchema(api *design.APIDefinition, s *JSONSchema, at *design.AttributeDefinition) *JSONSchema {
	s.Merge(TypeSchema(api, at.Type))
	s.DefaultValue = at.DefaultValue
	s.Description = at.Description
	for _, val := range at.Validations {
		switch actual := val.(type) {
		case *design.EnumValidationDefinition:
			s.Enum = actual.Values
		case *design.FormatValidationDefinition:
			s.Format = actual.Format
		case *design.PatternValidationDefinition:
			s.Pattern = actual.Pattern
		case *design.MinimumValidationDefinition:
			s.Minimum = actual.Min
		case *design.MaximumValidationDefinition:
			s.Maximum = actual.Max
		case *design.MinLengthValidationDefinition:
			s.MinLength = actual.MinLength
		case *design.MaxLengthValidationDefinition:
			s.MaxLength = actual.MaxLength
		case *design.RequiredValidationDefinition:
			s.Required = actual.Names
		}
	}
	return s
}

// toSchemaHref produces a href that replaces the path wildcards with JSON schema references when
// appropriate.
func toSchemaHref(api *design.APIDefinition, r *design.RouteDefinition) string {
	params := r.Params()
	args := make([]interface{}, len(params))
	for i, p := range params {
		args[i] = fmt.Sprintf("/{%s}", p)
	}
	tmpl := design.WildcardRegex.ReplaceAllLiteralString(r.FullPath(), "%s")
	return fmt.Sprintf(tmpl, args...)
}

// propertiesFromDefs creates a Properties map referencing the given definitions under the given
// path.
func propertiesFromDefs(definitions map[string]*JSONSchema, path string) map[string]*JSONSchema {
	res := make(map[string]*JSONSchema, len(definitions))
	for n := range definitions {
		if n == "identity" {
			continue
		}
		s := NewJSONSchema()
		s.Ref = path + n
		res[n] = s
	}
	return res
}

// buildMediaTypeSchema initializes s as the JSON schema representing mt.
func buildMediaTypeSchema(api *design.APIDefinition, mt *design.MediaTypeDefinition, s *JSONSchema) {
	s.Media = &JSONMedia{Type: mt.Identifier}
	lnames := make([]string, len(mt.Links))
	i := 0
	for n := range mt.Links {
		lnames[i] = n
		i++
	}
	for _, ln := range lnames {
		l := mt.Links[ln]
		att := l.Attribute() // cannot be nil if DSL validated
		r := l.MediaType().Resource
		var href string
		if r != nil {
			href = toSchemaHref(api, r.CanonicalAction().Routes[0])
		}
		s.Links = append(s.Links, &JSONLink{
			Title:        l.Name,
			Rel:          l.Name,
			Description:  att.Description,
			Href:         href,
			Method:       "GET",
			TargetSchema: TypeSchema(api, l.MediaType()),
			MediaType:    l.MediaType().Identifier,
		})
	}
	buildAttributeSchema(api, s, mt.AttributeDefinition)
}
