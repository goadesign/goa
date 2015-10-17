package genmetadata

import (
	"fmt"
	"net/url"

	"github.com/raphael/goa/design"
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

type (
	// JSONSchema represents an instance of a JSON schema.
	// See http://json-schema.org/documentation.html
	JSONSchema struct {
		// Core schema
		ID           string
		Type         JSONType
		Properties   map[string]*JSONSchema
		Item         *JSONSchema
		Definitions  map[string]*JSONSchema
		DefaultValue interface{}

		// Hyper schema
		Title     string
		Media     *JSONMedia
		ReadOnly  bool
		PathStart string
		Links     []*JSONLink

		// Validation
		Enum      []interface{}
		Format    string
		Pattern   string
		Minimum   int
		Maximum   int
		MinLength int
		MaxLength int
		Required  []string
	}

	// JSONType is the JSON type enum.
	JSONType string

	// JSONMedia represents a "media" field in a JSON hyper schema.
	JSONMedia struct {
		BinaryEncoding string
		Type           string
	}

	// JSONLink represents a "link" field in a JSON hyper schema.
	JSONLink struct {
		Title        string
		Rel          string
		Href         string
		Method       string
		Schema       *JSONSchema
		TargetSchema *JSONSchema
		MediaType    string
		EncType      string
	}
)

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
}

// TypeSchema produces the JSON schema corresponding to the given data type.
func TypeSchema(t design.DataType) *JSONSchema {
	s := &JSONSchema{}
	if t == nil {
		return s
	}
	switch actual := t.(type) {
	case design.Primitive:
		s.Type = JSONType(actual.Name())
	case *design.Array:
		s.Type = JSONArray
		s.Item = AttributeSchema(actual.ElemType)
	case design.Object:
		s.Type = JSONObject
		s.Properties = make(map[string]*JSONSchema)
		for n, at := range actual {
			s.Properties[n] = AttributeSchema(at)
		}
	case *design.UserTypeDefinition:
		if actual == nil {
			return s
		}
		s.ID = fmt.Sprintf("#%s", url.QueryEscape(actual.Name()))
		s.Merge(AttributeSchema(actual.AttributeDefinition))
	case *design.MediaTypeDefinition:
		s = MediaTypeSchema(actual)
	}
	return s
}

// AttributeSchema produces the JSON schema that corresponds to the given attribute.
func AttributeSchema(at *design.AttributeDefinition) *JSONSchema {
	if at == nil {
		return new(JSONSchema)
	}
	s := TypeSchema(at.Type)
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

// MediaTypeSchema produces the JSON schema corresponding to the given media type.
func MediaTypeSchema(mt *design.MediaTypeDefinition) *JSONSchema {
	s := JSONSchema{
		ID: fmt.Sprintf("#%s", url.QueryEscape(mt.Name())),
	}
	s.Media = &JSONMedia{
		Type: mt.Identifier,
	}
	for _, l := range mt.Links {
		att := l.Attribute() // cannot be nil if DSL validated
		r := l.MediaType().Resource
		var href string
		if r != nil {
			href = r.URITemplate()
		}
		s.Links = append(s.Links, &JSONLink{
			Title:        att.Description,
			Rel:          l.Name,
			Href:         href,
			Method:       "GET",
			TargetSchema: MediaTypeSchema(l.MediaType()),
			MediaType:    l.MediaType().Identifier,
		})
	}
	s.Merge(TypeSchema(mt.UserTypeDefinition))
	return &s
}

// ResourceSchema produces the JSON schema corresponding to the given API resource.
func ResourceSchema(api *design.APIDefinition, r *design.ResourceDefinition) *JSONSchema {
	if mt, ok := api.MediaTypes[r.MediaType]; ok {
		s := MediaTypeSchema(mt)
		s.ID = fmt.Sprintf("/%s", url.QueryEscape(r.Name))
		for _, a := range r.Actions {
			requestSchema := AttributeSchema(a.Params)
			requestSchema.Merge(TypeSchema(a.Payload))
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
					Href:   r.Path,
					Method: r.Verb,
					Schema: requestSchema,
				}
				if respMT != nil {
					link.MediaType = respMT.Identifier
					link.TargetSchema = MediaTypeSchema(respMT)
				}
				s.Links = append(s.Links, &link)
			}
		}
		return s
	}
	return nil
}

// APISchema produces the JSON schema corresponding to the given API definition.
// TBD use definitions for top level media types / types
func APISchema(api *design.APIDefinition) *JSONSchema {
	resourceMap := make(map[string]*JSONSchema)
	api.IterateResources(func(r *design.ResourceDefinition) error {
		resourceMap[r.Name] = ResourceSchema(api, r)
		return nil
	})
	resources := JSONSchema{
		ID:         "#resources",
		Type:       JSONObject,
		Properties: resourceMap,
	}
	s := JSONSchema{
		ID:   fmt.Sprintf("%s/api.json", HostName),
		Type: JSONObject,
		Properties: map[string]*JSONSchema{
			"Resources": &resources,
		},
	}
	return &s
}
