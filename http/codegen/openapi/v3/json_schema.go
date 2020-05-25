package openapiv3

import (
	"encoding/json"
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
		Minimum              *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
		Maximum              *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
		MinLength            *int          `json:"minLength,omitempty" yaml:"minLength,omitempty"`
		MaxLength            *int          `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
		MinItems             *int          `json:"minItems,omitempty" yaml:"minItems,omitempty"`
		MaxItems             *int          `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
		Required             []string      `json:"required,omitempty" yaml:"required,omitempty"`
		AdditionalProperties bool          `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

		// Union
		AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`

		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Type is the JSON type enum.
	Type string

	// Media represents a "media" field in a JSON hyper schema.
	Media struct {
		BinaryEncoding string `json:"binaryEncoding,omitempty" yaml:"binaryEncoding,omitempty"`
		Type           string `json:"type,omitempty" yaml:"type,omitempty"`
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
const SchemaRef = "https://json-schema.org/draft/2019-09/hyper-schema#"

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
