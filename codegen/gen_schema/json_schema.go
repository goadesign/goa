package genschema

import (
	"encoding/json"
	"fmt"

	"github.com/raphael/goa/design"
)

type (
	// JSONSchema represents an instance of a JSON schema.
	// See http://json-schema.org/documentation.html
	JSONSchema struct {
		// Core schema
		ID           string                 `json:"id,omitempty"`
		Description  string                 `json:"description,omitempty"`
		Schema       string                 `json:"$schema"`
		Type         JSONType               `json:"type,omitempty"`
		Properties   map[string]*JSONSchema `json:"properties,omitempty"`
		Item         *JSONSchema            `json:"item,omitempty"`
		Definitions  map[string]*JSONSchema `json:"definitions,omitempty"`
		DefaultValue interface{}            `json:"defaultValue,omitempty"`

		// Hyper schema
		Title     string      `json:"title,omitempty"`
		Media     *JSONMedia  `json:"media,omitempty"`
		ReadOnly  bool        `json:"readOnly,omitempty"`
		PathStart string      `json:"pathStart,omitempty"`
		Links     []*JSONLink `json:"links,omitempty"`
		Ref       string      `json:"$ref,omitempty"`

		// Validation
		Enum                 []interface{} `json:"enum,omitempty"`
		Format               string        `json:"format,omitempty"`
		Pattern              string        `json:"pattern,omitempty"`
		Minimum              int           `json:"minimum,omitempty"`
		Maximum              int           `json:"maximum,omitempty"`
		MinLength            int           `json:"minLength,omitempty"`
		MaxLength            int           `json:"maxLength,omitempty"`
		Required             []string      `json:"required,omitempty"`
		AdditionalProperties bool          `json:"additionalProperties,omitempty"`
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
	// Resource JSON Schemas that have been generated.
	resourceDefinitions map[string]*JSONSchema

	// Type JSON Schemas that have been generated.
	typeDefinitions map[string]*JSONSchema

	// Media type JSON Schemas that have been generated.
	mediaTypeDefinitions map[string]*JSONSchema
)

// Initialize the global variables
func init() {
	resourceDefinitions = make(map[string]*JSONSchema)
	typeDefinitions = make(map[string]*JSONSchema)
	mediaTypeDefinitions = make(map[string]*JSONSchema)
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
	properties := make(map[string]*JSONSchema, len(api.Resources))
	api.IterateResources(func(r *design.ResourceDefinition) error {
		properties[r.Name] = &JSONSchema{Ref: ResourceRef(api, r)}
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
	definitions := map[string]*JSONSchema{
		"resources": &JSONSchema{
			Description: fmt.Sprintf("%s resources", api.Name),
			Schema:      SchemaRef,
			Type:        JSONObject,
			Properties:  resourceDefinitions,
		},
		"media": &JSONSchema{
			Description: fmt.Sprintf("%s media types", api.Name),
			Schema:      SchemaRef,
			Type:        JSONObject,
			Properties:  mediaTypeDefinitions,
		},
		"types": &JSONSchema{
			Description: fmt.Sprintf("%s user types", api.Name),
			Schema:      SchemaRef,
			Type:        JSONObject,
			Properties:  typeDefinitions,
		},
	}
	s := JSONSchema{
		ID:          fmt.Sprintf("%s/schema", ServiceURL),
		Schema:      SchemaRef,
		Title:       api.Title,
		Description: api.Description,
		Type:        JSONObject,
		Definitions: definitions,
		Properties:  properties,
		Links:       links,
	}
	return &s
}

// ResourceRef produces the JSON reference to the resource definition.
func ResourceRef(api *design.APIDefinition, r *design.ResourceDefinition) string {
	if _, ok := resourceDefinitions[r.Name]; !ok {
		GenerateResourceDefinition(api, r)
	}
	return fmt.Sprintf("#/definitions/resources/%s", r.FormatName(true, false))
}

// MediaTypeRef produces the JSON reference to the media type definition.
func MediaTypeRef(api *design.APIDefinition, mt *design.MediaTypeDefinition) string {
	if _, ok := mediaTypeDefinitions[mt.TypeName]; !ok {
		GenerateMediaTypeDefinition(api, mt)
	}
	return fmt.Sprintf("#/definitions/media/%s", mt.FormatName(true, false))
}

// TypeRef produces the JSON reference to the type definition.
func TypeRef(api *design.APIDefinition, ut *design.UserTypeDefinition) string {
	if _, ok := resourceDefinitions[ut.TypeName]; !ok {
		GenerateTypeDefinition(api, ut)
	}
	return fmt.Sprintf("#/definitions/types/%s", ut.FormatName(true, false))
}

// GenerateResourceDefinition produces the JSON schema corresponding to the given API resource.
// It stores the results in cachedSchema.
func GenerateResourceDefinition(api *design.APIDefinition, r *design.ResourceDefinition) {
	if _, ok := resourceDefinitions[r.Name]; ok {
		return
	}
	s := NewJSONSchema()
	s.Description = r.Description
	s.Type = JSONObject
	resourceDefinitions[r.Name] = s
	if mt, ok := api.MediaTypes[r.MediaType]; ok {
		GenerateMediaTypeDefinition(api, mt)
		mtd, _ := mediaTypeDefinitions[mt.TypeName]
		for n, p := range mtd.Properties {
			s.Properties[n] = p.Dup()
		}
	}
	for _, a := range r.Actions {
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
		var successResponse *design.ResponseDefinition
		for _, resp := range a.Responses {
			if resp.Status > 199 && resp.Status < 300 {
				successResponse = resp
				break
			}
		}
		var respMT *design.MediaTypeDefinition
		if successResponse != nil {
			respMT, _ = api.MediaTypes[successResponse.MediaType]
		}
		for _, r := range a.Routes {
			link := JSONLink{
				Title:  a.Description,
				Rel:    a.Name,
				Href:   toSchemaHref(api, r),
				Method: r.Verb,
				Schema: requestSchema,
			}
			if respMT != nil {
				link.MediaType = respMT.Identifier
				link.TargetSchema = TypeSchema(api, respMT)
			}
			s.Links = append(s.Links, &link)
		}
	}
}

// GenerateMediaTypeDefinition produces the JSON schema corresponding to the given media type.
func GenerateMediaTypeDefinition(api *design.APIDefinition, mt *design.MediaTypeDefinition) {
	if _, ok := mediaTypeDefinitions[mt.TypeName]; ok {
		return
	}
	s := NewJSONSchema()
	s.Media = &JSONMedia{Type: mt.Identifier}
	mediaTypeDefinitions[mt.TypeName] = s
	for _, l := range mt.Links {
		att := l.Attribute() // cannot be nil if DSL validated
		r := l.MediaType().Resource
		var href string
		if r != nil {
			href = toSchemaHref(api, r.CanonicalAction().Routes[0])
		}
		s.Links = append(s.Links, &JSONLink{
			Title:        att.Description,
			Rel:          l.Name,
			Href:         href,
			Method:       "GET",
			TargetSchema: TypeSchema(api, l.MediaType()),
			MediaType:    l.MediaType().Identifier,
		})
	}
	s.Merge(TypeSchema(api, mt.UserTypeDefinition))
}

// GenerateTypeDefinition produces the JSON schema corresponding to the given type.
func GenerateTypeDefinition(api *design.APIDefinition, ut *design.UserTypeDefinition) {
	if _, ok := typeDefinitions[ut.TypeName]; ok {
		return
	}
	s := NewJSONSchema()
	typeDefinitions[ut.TypeName] = s
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
		s.Item = NewJSONSchema()
		buildAttributeSchema(api, s.Item, actual.ElemType)
	case design.Object:
		s.Type = JSONObject
		for n, at := range actual {
			def := NewJSONSchema()
			buildAttributeSchema(api, def, at)
			s.Definitions[n] = def
			prop := NewJSONSchema()
			prop.Ref = n
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
	if s.ID == "" {
		s.ID = other.ID
	}
	if s.Type == "" {
		s.Type = other.Type
	}
	for n, p := range other.Properties {
		if _, ok := s.Properties[n]; !ok {
			if s.Properties == nil {
				s.Properties = make(map[string]*JSONSchema)
			}
			s.Properties[n] = p
		}
	}
	if s.Item == nil {
		s.Item = other.Item
	}
	for n, d := range other.Definitions {
		if _, ok := s.Definitions[n]; !ok {
			s.Definitions[n] = d
		}
	}
	if s.DefaultValue == nil {
		s.DefaultValue = other.DefaultValue
	}
	if s.Title == "" {
		s.Title = other.Title
	}
	if s.Media == nil {
		s.Media = other.Media
	}
	if s.ReadOnly == false {
		s.ReadOnly = other.ReadOnly
	}
	if s.PathStart == "" {
		s.PathStart = other.PathStart
	}
	for _, l := range other.Links {
		s.Links = append(s.Links, l)
	}
	if s.Enum == nil {
		s.Enum = other.Enum
	}
	if s.Format == "" {
		s.Format = other.Format
	}
	if s.Pattern == "" {
		s.Pattern = other.Pattern
	}
	if s.Minimum > other.Minimum {
		s.Minimum = other.Minimum
	}
	if s.Maximum < other.Maximum {
		s.Maximum = other.Maximum
	}
	if s.MinLength > other.MinLength {
		s.MinLength = other.MinLength
	}
	if s.MaxLength < other.MaxLength {
		s.MaxLength = other.MaxLength
	}
	for _, r := range other.Required {
		s.Required = append(s.Required, r)
	}
	if s.AdditionalProperties == false {
		s.AdditionalProperties = other.AdditionalProperties
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
	if s.Item != nil {
		js.Item = s.Item.Dup()
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
		args[i] = fmt.Sprintf("{%s}", p)
	}
	tmpl := design.WildcardRegex.ReplaceAllLiteralString(r.FullPath(), "%s")
	return fmt.Sprintf(tmpl, args...)
}
