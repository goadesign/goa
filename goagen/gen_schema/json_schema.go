package genschema

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"

	"github.com/goadesign/goa/design"
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
		DefaultValue interface{}            `json:"default,omitempty"`
		Example      interface{}            `json:"example,omitempty"`

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
		Minimum              *float64      `json:"minimum,omitempty"`
		Maximum              *float64      `json:"maximum,omitempty"`
		MinLength            *int          `json:"minLength,omitempty"`
		MaxLength            *int          `json:"maxLength,omitempty"`
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
	// JSONFile is an extension used by Swagger to represent a file download.
	JSONFile = "file"
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
	scheme := "http"
	if len(api.Schemes) > 0 {
		scheme = api.Schemes[0]
	}
	u := url.URL{Scheme: scheme, Host: api.Host}
	href := u.String()
	links := []*JSONLink{
		{
			Href: href,
			Rel:  "self",
		},
		{
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
		ID:          fmt.Sprintf("%s/schema", href),
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
		for _, v := range mt.Views {
			buildMediaTypeSchema(api, mt, v.Name, s)
		}
	}
	r.IterateActions(func(a *design.ActionDefinition) error {
		var requestSchema *JSONSchema
		if a.Payload != nil {
			requestSchema = TypeSchema(api, a.Payload)
			requestSchema.Description = a.Name + " payload"
		}
		if a.Params != nil {
			params := design.DupAtt(a.Params)
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

// MediaTypeRef produces the JSON reference to the media type definition with the given view.
func MediaTypeRef(api *design.APIDefinition, mt *design.MediaTypeDefinition, view string) string {
	projected, _, err := mt.Project(view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	if _, ok := Definitions[projected.TypeName]; !ok {
		GenerateMediaTypeDefinition(api, projected, "default")
	}
	ref := fmt.Sprintf("#/definitions/%s", projected.TypeName)
	return ref
}

// TypeRef produces the JSON reference to the type definition.
func TypeRef(api *design.APIDefinition, ut *design.UserTypeDefinition) string {
	if _, ok := Definitions[ut.TypeName]; !ok {
		GenerateTypeDefinition(api, ut)
	}
	return fmt.Sprintf("#/definitions/%s", ut.TypeName)
}

// GenerateMediaTypeDefinition produces the JSON schema corresponding to the given media type and
// given view.
func GenerateMediaTypeDefinition(api *design.APIDefinition, mt *design.MediaTypeDefinition, view string) {
	if _, ok := Definitions[mt.TypeName]; ok {
		return
	}
	s := NewJSONSchema()
	s.Title = fmt.Sprintf("Mediatype identifier: %s", mt.Identifier)
	Definitions[mt.TypeName] = s
	buildMediaTypeSchema(api, mt, view, s)
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
		if name := actual.Name(); name != "any" {
			s.Type = JSONType(actual.Name())
		}
		switch actual.Kind() {
		case design.UUIDKind:
			s.Format = "uuid"
		case design.DateTimeKind:
			s.Format = "date-time"
		case design.NumberKind:
			s.Format = "double"
		case design.IntegerKind:
			s.Format = "int64"
		}
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
		// Use "default" view by default
		s.Ref = MediaTypeRef(api, actual, design.DefaultView)
	}
	return s
}

type mergeItems []struct {
	a, b   interface{}
	needed bool
}

func (s *JSONSchema) createMergeItems(other *JSONSchema) mergeItems {
	return mergeItems{
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
		{
			a: s.Minimum, b: other.Minimum,
			needed: (s.Minimum == nil && s.Minimum != nil) ||
				(s.Minimum != nil && other.Minimum != nil && *s.Minimum > *other.Minimum),
		},
		{
			a: s.Maximum, b: other.Maximum,
			needed: (s.Maximum == nil && other.Maximum != nil) ||
				(s.Maximum != nil && other.Maximum != nil && *s.Maximum < *other.Maximum),
		},
		{
			a: s.MinLength, b: other.MinLength,
			needed: (s.MinLength == nil && other.MinLength != nil) ||
				(s.MinLength != nil && other.MinLength != nil && *s.MinLength > *other.MinLength),
		},
		{
			a: s.MaxLength, b: other.MaxLength,
			needed: (s.MaxLength == nil && other.MaxLength != nil) ||
				(s.MaxLength != nil && other.MaxLength != nil && *s.MaxLength > *other.MaxLength),
		},
	}
}

// Merge does a two level deep merge of other into s.
func (s *JSONSchema) Merge(other *JSONSchema) {
	items := s.createMergeItems(other)
	for _, v := range items {
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
	if at.View != "" {
		inner := NewJSONSchema()
		inner.Ref = MediaTypeRef(api, at.Type.(*design.MediaTypeDefinition), at.View)
		s.Merge(inner)
		return s
	}
	s.Merge(TypeSchema(api, at.Type))
	if s.Ref != "" {
		// Ref is exclusive with other fields
		return s
	}
	s.DefaultValue = toStringMap(at.DefaultValue)
	s.Description = at.Description
	s.Example = at.GenerateExample(api.RandomGenerator(), nil)
	val := at.Validation
	if val == nil {
		return s
	}
	s.Enum = val.Values
	s.Format = val.Format
	s.Pattern = val.Pattern
	if val.Minimum != nil {
		s.Minimum = val.Minimum
	}
	if val.Maximum != nil {
		s.Maximum = val.Maximum
	}
	if val.MinLength != nil {
		s.MinLength = val.MinLength
	}
	if val.MaxLength != nil {
		s.MaxLength = val.MaxLength
	}
	s.Required = val.Required
	return s
}

// toStringMap converts map[interface{}]interface{} to a map[string]interface{} when possible.
func toStringMap(val interface{}) interface{} {
	switch actual := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range actual {
			m[toString(k)] = toStringMap(v)
		}
		return m
	case []interface{}:
		mapSlice := make([]interface{}, len(actual))
		for i, e := range actual {
			mapSlice[i] = toStringMap(e)
		}
		return mapSlice
	default:
		return actual
	}
}

// toString returns the string representation of the given type.
func toString(val interface{}) string {
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

// buildMediaTypeSchema initializes s as the JSON schema representing mt for the given view.
func buildMediaTypeSchema(api *design.APIDefinition, mt *design.MediaTypeDefinition, view string, s *JSONSchema) {
	s.Media = &JSONMedia{Type: mt.Identifier}
	projected, linksUT, err := mt.Project(view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	if linksUT != nil {
		links := linksUT.Type.ToObject()
		lnames := make([]string, len(links))
		i := 0
		for n := range links {
			lnames[i] = n
			i++
		}
		sort.Strings(lnames)
		for _, ln := range lnames {
			var (
				att  = links[ln]
				lmt  = att.Type.(*design.MediaTypeDefinition)
				r    = lmt.Resource
				href string
			)
			if r != nil {
				href = toSchemaHref(api, r.CanonicalAction().Routes[0])
			}
			sm := NewJSONSchema()
			sm.Ref = MediaTypeRef(api, lmt, "default")
			s.Links = append(s.Links, &JSONLink{
				Title:        ln,
				Rel:          ln,
				Description:  att.Description,
				Href:         href,
				Method:       "GET",
				TargetSchema: sm,
				MediaType:    lmt.Identifier,
			})
		}
	}
	buildAttributeSchema(api, s, projected.AttributeDefinition)
}
