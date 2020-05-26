// Package openapiv2 produces OpenAPI Specification 2.0 (https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md)
// for the HTTP endpoints.
package openapiv2

import (
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

type (
	// V2 represents an instance of a Swagger object.
	// See https://github.com/OAI/OpenAPI-Specification
	V2 struct {
		Swagger             string                         `json:"swagger,omitempty" yaml:"swagger,omitempty"`
		Info                *Info                          `json:"info,omitempty" yaml:"info,omitempty"`
		Host                string                         `json:"host,omitempty" yaml:"host,omitempty"`
		BasePath            string                         `json:"basePath,omitempty" yaml:"basePath,omitempty"`
		Schemes             []string                       `json:"schemes,omitempty" yaml:"schemes,omitempty"`
		Consumes            []string                       `json:"consumes,omitempty" yaml:"consumes,omitempty"`
		Produces            []string                       `json:"produces,omitempty" yaml:"produces,omitempty"`
		Paths               map[string]interface{}         `json:"paths" yaml:"paths"`
		Definitions         map[string]*openapi.Schema     `json:"definitions,omitempty" yaml:"definitions,omitempty"`
		Parameters          map[string]*Parameter          `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		Responses           map[string]*Response           `json:"responses,omitempty" yaml:"responses,omitempty"`
		SecurityDefinitions map[string]*SecurityDefinition `json:"securityDefinitions,omitempty" yaml:"securityDefinitions,omitempty"`
		Tags                []*openapi.Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
		ExternalDocs        *openapi.ExternalDocs          `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	}

	// Info provides metadata about the API. The metadata can be used by the clients if needed,
	// and can be presented in the OpenAPI UI for convenience.
	Info struct {
		Title          string                 `json:"title" yaml:"title"`
		Description    string                 `json:"description,omitempty" yaml:"description,omitempty"`
		TermsOfService string                 `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
		Contact        *expr.ContactExpr      `json:"contact,omitempty" yaml:"contact,omitempty"`
		License        *expr.LicenseExpr      `json:"license,omitempty" yaml:"license,omitempty"`
		Version        string                 `json:"version" yaml:"version"`
		Extensions     map[string]interface{} `json:"-" yaml:"-"`
	}

	// Path holds the relative paths to the individual endpoints.
	Path struct {
		// Ref allows for an external definition of this path item.
		Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
		// Get defines a GET operation on this path.
		Get *Operation `json:"get,omitempty" yaml:"get,omitempty"`
		// Put defines a PUT operation on this path.
		Put *Operation `json:"put,omitempty" yaml:"put,omitempty"`
		// Post defines a POST operation on this path.
		Post *Operation `json:"post,omitempty" yaml:"post,omitempty"`
		// Delete defines a DELETE operation on this path.
		Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
		// Options defines a OPTIONS operation on this path.
		Options *Operation `json:"options,omitempty" yaml:"options,omitempty"`
		// Head defines a HEAD operation on this path.
		Head *Operation `json:"head,omitempty" yaml:"head,omitempty"`
		// Patch defines a PATCH operation on this path.
		Patch *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
		// Parameters is the list of parameters that are applicable for all the operations
		// described under this path.
		Parameters []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Operation describes a single API operation on a path.
	Operation struct {
		// Tags is a list of tags for API documentation control. Tags
		// can be used for logical grouping of operations by services or
		// any other qualifier.
		Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
		// Summary is a short summary of what the operation does. For maximum readability
		// in the swagger-ui, this field should be less than 120 characters.
		Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
		// Description is a verbose explanation of the operation behavior.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// ExternalDocs points to additional external documentation for this operation.
		ExternalDocs *openapi.ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
		// OperationID is a unique string used to identify the operation.
		OperationID string `json:"operationId,omitempty" yaml:"operationId,omitempty"`
		// Consumes is a list of MIME types the operation can consume.
		Consumes []string `json:"consumes,omitempty" yaml:"consumes,omitempty"`
		// Produces is a list of MIME types the operation can produce.
		Produces []string `json:"produces,omitempty" yaml:"produces,omitempty"`
		// Parameters is a list of parameters that are applicable for this operation.
		Parameters []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		// Responses is the list of possible responses as they are returned from executing
		// this operation.
		Responses map[string]*Response `json:"responses,omitempty" yaml:"responses,omitempty"`
		// Schemes is the transfer protocol for the operation.
		Schemes []string `json:"schemes,omitempty" yaml:"schemes,omitempty"`
		// Deprecated declares this operation to be deprecated.
		Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
		// Security is a declaration of which security schemes are applied for this operation.
		Security []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Parameter describes a single operation parameter.
	Parameter struct {
		// Name of the parameter. Parameter names are case sensitive.
		Name string `json:"name" yaml:"name"`
		// In is the location of the parameter.
		// Possible values are "query", "header", "path", "formData" or "body".
		In string `json:"in" yaml:"in"`
		// Description is a brief description of the parameter.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// Required determines whether this parameter is mandatory.
		Required bool `json:"required" yaml:"required"`
		// Schema defining the type used for the body parameter, only if "in" is body
		Schema *openapi.Schema `json:"schema,omitempty" yaml:"schema,omitempty"`

		// properties below only apply if "in" is not body

		//  Type of the parameter. Since the parameter is not located at the request body,
		// it is limited to simple types (that is, not an object).
		Type string `json:"type,omitempty" yaml:"type,omitempty"`
		// Format is the extending format for the previously mentioned type.
		Format string `json:"format,omitempty" yaml:"format,omitempty"`
		// AllowEmptyValue sets the ability to pass empty-valued parameters.
		// This is valid only for either query or formData parameters and allows you to
		// send a parameter with a name only or an empty value. Default value is false.
		AllowEmptyValue bool `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
		// Items describes the type of items in the array if type is "array".
		Items *Items `json:"items,omitempty" yaml:"items,omitempty"`
		// CollectionFormat determines the format of the array if type array is used.
		// Possible values are csv, ssv, tsv, pipes and multi.
		CollectionFormat string `json:"collectionFormat,omitempty" yaml:"collectionFormat,omitempty"`
		// Default declares the value of the parameter that the server will use if none is
		// provided, for example a "count" to control the number of results per page might
		// default to 100 if not supplied by the client in the request.
		Default          interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
		Maximum          *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
		Minimum          *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
		MaxLength        *int          `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
		MinLength        *int          `json:"minLength,omitempty" yaml:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
		MaxItems         *int          `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
		MinItems         *int          `json:"minItems,omitempty" yaml:"minItems,omitempty"`
		UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
		Enum             []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
		MultipleOf       float64       `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Response describes an operation response.
	Response struct {
		// Description of the response. GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// Schema is a definition of the response structure. It can be a primitive,
		// an array or an object. If this field does not exist, it means no content is
		// returned as part of the response. As an extension to the Schema Object, its root
		// type value may also be "file".
		Schema *openapi.Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
		// Headers is a list of headers that are sent with the response.
		Headers map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
		// Ref references a global API response.
		// This field is exclusive with the other fields of Response.
		Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Header represents a header parameter.
	Header struct {
		// Description is a brief description of the parameter.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		//  Type of the header. it is limited to simple types (that is, not an object).
		Type string `json:"type,omitempty" yaml:"type,omitempty"`
		// Format is the extending format for the previously mentioned type.
		Format string `json:"format,omitempty" yaml:"format,omitempty"`
		// Items describes the type of items in the array if type is "array".
		Items *Items `json:"items,omitempty" yaml:"items,omitempty"`
		// CollectionFormat determines the format of the array if type array is used.
		// Possible values are csv, ssv, tsv, pipes and multi.
		CollectionFormat string `json:"collectionFormat,omitempty" yaml:"collectionFormat,omitempty"`
		// Default declares the value of the parameter that the server will use if none is
		// provided, for example a "count" to control the number of results per page might
		// default to 100 if not supplied by the client in the request.
		Default          interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
		Maximum          *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
		Minimum          *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
		MaxLength        *int          `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
		MinLength        *int          `json:"minLength,omitempty" yaml:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
		MaxItems         *int          `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
		MinItems         *int          `json:"minItems,omitempty" yaml:"minItems,omitempty"`
		UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
		Enum             []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
		MultipleOf       float64       `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	}

	// SecurityDefinition allows the definition of a security scheme that can be used by the
	// operations. Supported schemes are basic authentication, an API key (either as a header or
	// as a query parameter) and OAuth2's common flows (implicit, password, application and
	// access code).
	SecurityDefinition struct {
		// Type of the security scheme. Valid values are "basic", "apiKey" or "oauth2".
		Type string `json:"type" yaml:"type"`
		// Description for security scheme
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// Name of the header or query parameter to be used when type is "apiKey".
		Name string `json:"name,omitempty" yaml:"name,omitempty"`
		// In is the location of the API key when type is "apiKey".
		// Valid values are "query" or "header".
		In string `json:"in,omitempty" yaml:"in,omitempty"`
		// Flow is the flow used by the OAuth2 security scheme when type is "oauth2"
		// Valid values are "implicit", "password", "application" or "accessCode".
		Flow string `json:"flow,omitempty" yaml:"flow,omitempty"`
		// The oauth2 authorization URL to be used for this flow.
		AuthorizationURL string `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
		// TokenURL  is the token URL to be used for this flow.
		TokenURL string `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
		// Scopes list the  available scopes for the OAuth2 security scheme.
		Scopes map[string]string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Scope corresponds to an available scope for an OAuth2 security scheme.
	Scope struct {
		// Description for scope
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
	}

	// Items is a limited subset of JSON-Schema's items object. It is used by parameter
	// definitions that are not located in "body".
	Items struct {
		//  Type of the items. it is limited to simple types (that is, not an object).
		Type string `json:"type,omitempty" yaml:"type,omitempty"`
		// Format is the extending format for the previously mentioned type.
		Format string `json:"format,omitempty" yaml:"format,omitempty"`
		// Items describes the type of items in the array if type is "array".
		Items *Items `json:"items,omitempty" yaml:"items,omitempty"`
		// CollectionFormat determines the format of the array if type array is used.
		// Possible values are csv, ssv, tsv, pipes and multi.
		CollectionFormat string `json:"collectionFormat,omitempty" yaml:"collectionFormat,omitempty"`
		// Default declares the value of the parameter that the server will use if none is
		// provided, for example a "count" to control the number of results per page might
		// default to 100 if not supplied by the client in the request.
		Default          interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
		Maximum          *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
		Minimum          *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
		MaxLength        *int          `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
		MinLength        *int          `json:"minLength,omitempty" yaml:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
		MaxItems         *int          `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
		MinItems         *int          `json:"minItems,omitempty" yaml:"minItems,omitempty"`
		UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
		Enum             []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
		MultipleOf       float64       `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	}

	// These types are used in openapi.MarshalJSON() to avoid recursive call of json.Marshal().
	_Info               Info
	_Path               Path
	_Operation          Operation
	_Parameter          Parameter
	_Response           Response
	_SecurityDefinition SecurityDefinition
)

// MarshalJSON returns the JSON encoding of i.
func (i Info) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_Info(i), i.Extensions)
}

// MarshalJSON returns the JSON encoding of p.
func (p Path) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_Path(p), p.Extensions)
}

// MarshalJSON returns the JSON encoding of o.
func (o Operation) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_Operation(o), o.Extensions)
}

// MarshalJSON returns the JSON encoding of p.
func (p Parameter) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_Parameter(p), p.Extensions)
}

// MarshalJSON returns the JSON encoding of r.
func (r Response) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_Response(r), r.Extensions)
}

// MarshalJSON returns the JSON encoding of s.
func (s SecurityDefinition) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_SecurityDefinition(s), s.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (i Info) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_Info(i), i.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (p Path) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_Path(p), p.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (o Operation) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_Operation(o), o.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (p Parameter) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_Parameter(p), p.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (r Response) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_Response(r), r.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (s SecurityDefinition) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_SecurityDefinition(s), s.Extensions)
}
