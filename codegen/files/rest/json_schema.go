package rest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

type (
	// Schema represents an instance of a JSON schema.
	// See http://json-schema.org/documentation.html
	Schema struct {
		Schema string `json:"$schema,omitempty"`
		// Core schema
		ID           string             `json:"id,omitempty"`
		Title        string             `json:"title,omitempty"`
		Type         Type               `json:"type,omitempty"`
		Items        *Schema            `json:"items,omitempty"`
		Properties   map[string]*Schema `json:"properties,omitempty"`
		Definitions  map[string]*Schema `json:"definitions,omitempty"`
		Description  string             `json:"description,omitempty"`
		DefaultValue interface{}        `json:"default,omitempty"`
		Example      interface{}        `json:"example,omitempty"`

		// Hyper schema
		Media     *Media  `json:"media,omitempty"`
		ReadOnly  bool    `json:"readOnly,omitempty"`
		PathStart string  `json:"pathStart,omitempty"`
		Links     []*Link `json:"links,omitempty"`
		Ref       string  `json:"$ref,omitempty"`

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
		AnyOf []*Schema `json:"anyOf,omitempty"`
	}

	// Type is the JSON type enum.
	Type string

	// Media represents a "media" field in a JSON hyper schema.
	Media struct {
		BinaryEncoding string `json:"binaryEncoding,omitempty"`
		Type           string `json:"type,omitempty"`
	}

	// Link represents a "link" field in a JSON hyper schema.
	Link struct {
		Title        string  `json:"title,omitempty"`
		Description  string  `json:"description,omitempty"`
		Rel          string  `json:"rel,omitempty"`
		Href         string  `json:"href,omitempty"`
		Method       string  `json:"method,omitempty"`
		Schema       *Schema `json:"schema,omitempty"`
		TargetSchema *Schema `json:"targetSchema,omitempty"`
		MediaType    string  `json:"mediaType,omitempty"`
		EncType      string  `json:"encType,omitempty"`
	}
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
	// File is an extension used by Swagger to represent a file download.
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

// JSON serializes the schema into JSON.
// It makes sure the "$schema" standard field is set if needed prior to
// delegating to the standard // JSON marshaler.
func (s *Schema) JSON() ([]byte, error) {
	if s.Ref == "" {
		s.Schema = SchemaRef
	}
	return json.Marshal(s)
}

// APISchema produces the API JSON hyper schema.
func APISchema(api *design.APIExpr, r *rest.RootExpr) *Schema {
	for _, res := range r.Resources {
		GenerateResourceDefinition(api, res)
	}
	href := api.Servers[0].URL
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

// GenerateResourceDefinition produces the JSON schema corresponding to the given
// API resource. It stores the results in cachedSchema.
func GenerateResourceDefinition(api *design.APIExpr, res *rest.ResourceExpr) {
	s := NewSchema()
	s.Description = res.Description()
	s.Type = Object
	s.Title = res.Name()
	Definitions[res.Name()] = s
	for _, a := range res.Actions {
		var requestSchema *Schema
		if a.MethodExpr.Payload != nil {
			requestSchema = TypeSchema(api, a.MethodExpr.Payload.Type)
			requestSchema.Description = a.Name() + " payload"
		}
		if a.Params() != nil {
			params := a.MappedParams()
			// We don't want to keep the path params, these are
			// defined inline in the href
			for _, r := range a.Routes {
				for _, p := range r.Params() {
					delete(design.AsObject(params.Type), p)
				}
			}
		}
		var targetSchema *Schema
		var identifier string
		for _, resp := range a.Responses {
			if mt := resp.MediaType(); mt != nil {
				if identifier == "" {
					identifier = mt.Identifier
				} else {
					identifier = ""
				}
				if targetSchema == nil {
					targetSchema = TypeSchema(api, mt)
				} else if targetSchema.AnyOf == nil {
					firstSchema := targetSchema
					targetSchema = NewSchema()
					targetSchema.AnyOf = []*Schema{firstSchema, TypeSchema(api, mt)}
				} else {
					targetSchema.AnyOf = append(targetSchema.AnyOf, TypeSchema(api, mt))
				}
			}
		}
		for i, r := range a.Routes {
			link := Link{
				Title:        a.Name(),
				Rel:          a.Name(),
				Href:         toSchemaHref(r),
				Method:       r.Method,
				Schema:       requestSchema,
				TargetSchema: targetSchema,
				MediaType:    identifier,
			}
			if i == 0 {
				if ca := a.Resource.CanonicalAction(); ca != nil {
					if ca.Name() == a.Name() {
						link.Rel = "self"
					}
				}
			}
			s.Links = append(s.Links, &link)
		}
	}
}

// MediaTypeRef produces the JSON reference to the media type definition with
// the given view.
func MediaTypeRef(api *design.APIExpr, mt *design.MediaTypeExpr, view string) string {
	projected, err := new(design.Projector).Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	if _, ok := Definitions[projected.MediaType.TypeName]; !ok {
		GenerateMediaTypeDefinition(api, projected.MediaType, "default")
	}
	ref := fmt.Sprintf("#/definitions/%s", projected.MediaType.TypeName)
	return ref
}

// TypeRef produces the JSON reference to the type definition.
func TypeRef(api *design.APIExpr, ut *design.UserTypeExpr) string {
	if _, ok := Definitions[ut.TypeName]; !ok {
		GenerateTypeDefinition(api, ut)
	}
	return fmt.Sprintf("#/definitions/%s", ut.TypeName)
}

// GenerateMediaTypeDefinition produces the JSON schema corresponding to the
// given media type and given view.
func GenerateMediaTypeDefinition(api *design.APIExpr, mt *design.MediaTypeExpr, view string) {
	if _, ok := Definitions[mt.TypeName]; ok {
		return
	}
	s := NewSchema()
	s.Title = fmt.Sprintf("Mediatype identifier: %s", mt.Identifier)
	Definitions[mt.TypeName] = s
	buildMediaTypeSchema(api, mt, view, s)
}

// GenerateTypeDefinition produces the JSON schema corresponding to the given
// type.
func GenerateTypeDefinition(api *design.APIExpr, ut *design.UserTypeExpr) {
	if _, ok := Definitions[ut.TypeName]; ok {
		return
	}
	s := NewSchema()
	s.Title = ut.TypeName
	Definitions[ut.TypeName] = s
	buildAttributeSchema(api, s, ut.AttributeExpr)
}

// TypeSchema produces the JSON schema corresponding to the given data type.
func TypeSchema(api *design.APIExpr, t design.DataType) *Schema {
	s := NewSchema()
	switch actual := t.(type) {
	case design.Primitive:
		if name := actual.Name(); name != "any" {
			s.Type = Type(actual.Name())
		}
		switch actual.Kind() {
		case design.IntKind, design.Int64Kind,
			design.UIntKind, design.UInt64Kind:
			s.Format = "int64"
		case design.Int32Kind, design.UInt32Kind:
			s.Format = "int32"
		case design.Float32Kind:
			s.Format = "float"
		case design.Float64Kind:
			s.Format = "double"
		}
	case *design.Array:
		s.Type = Array
		s.Items = NewSchema()
		buildAttributeSchema(api, s.Items, actual.ElemType)
	case design.Object:
		s.Type = Object
		for n, at := range actual {
			prop := NewSchema()
			buildAttributeSchema(api, prop, at)
			s.Properties[n] = prop
		}
	case *design.Map:
		s.Type = Object
		s.AdditionalProperties = true
	case *design.UserTypeExpr:
		s.Ref = TypeRef(api, actual)
	case *design.MediaTypeExpr:
		// Use "default" view by default
		s.Ref = MediaTypeRef(api, actual, design.DefaultView)
	}
	return s
}

type mergeItems []struct {
	a, b   interface{}
	needed bool
}

func (s *Schema) createMergeItems(other *Schema) mergeItems {
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
func (s *Schema) Merge(other *Schema) {
	items := s.createMergeItems(other)
	for _, v := range items {
		if v.needed && v.b != nil {
			reflect.Indirect(reflect.ValueOf(v.a)).Set(reflect.ValueOf(v.b))
		}
	}

	for n, p := range other.Properties {
		if _, ok := s.Properties[n]; !ok {
			if s.Properties == nil {
				s.Properties = make(map[string]*Schema)
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
func buildAttributeSchema(api *design.APIExpr, s *Schema, at *design.AttributeExpr) *Schema {
	s.Merge(TypeSchema(api, at.Type))
	if s.Ref != "" {
		// Ref is exclusive with other fields
		return s
	}
	s.DefaultValue = toStringMap(at.DefaultValue)
	s.Description = at.Description
	s.Example = at.Example(api.Random())
	val := at.Validation
	if val == nil {
		return s
	}
	s.Enum = val.Values
	s.Format = string(val.Format)
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

// toStringMap converts map[interface{}]interface{} to a map[string]interface{}
// when possible.
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

// toSchemaHref produces a href that replaces the path wildcards with JSON
// schema references when appropriate.
func toSchemaHref(r *rest.RouteExpr) string {
	params := r.Params()
	args := make([]interface{}, len(params))
	for i, p := range params {
		args[i] = fmt.Sprintf("/{%s}", p)
	}
	tmpl := rest.WildcardRegex.ReplaceAllLiteralString(r.FullPath(), "%s")
	return fmt.Sprintf(tmpl, args...)
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

// buildMediaTypeSchema initializes s as the JSON schema representing mt for the
// given view.
func buildMediaTypeSchema(api *design.APIExpr, mt *design.MediaTypeExpr, view string, s *Schema) {
	s.Media = &Media{Type: mt.Identifier}
	projected, err := new(design.Projector).Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	if projected.Links != nil {
		links := design.AsObject(projected.Links)
		lnames := make([]string, len(links))
		i := 0
		for n := range links {
			lnames[i] = n
			i++
		}
		sort.Strings(lnames)
		for _, ln := range lnames {
			// TBD: compute the href of the mediatype by looking at
			// all the API actions and finding one that returns it.
			var (
				att = links[ln]
				lmt = att.Type.(*design.MediaTypeExpr)
			)
			sm := NewSchema()
			sm.Ref = MediaTypeRef(api, lmt, "default")
			s.Links = append(s.Links, &Link{
				Title:        ln,
				Rel:          ln,
				Description:  att.Description,
				Method:       "GET",
				TargetSchema: sm,
				MediaType:    lmt.Identifier,
			})
		}
	}
	buildAttributeSchema(api, s, projected.MediaType.AttributeExpr)
}
