// Package openapi produces OpenAPI Specification 2.0 (https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md)
// for the HTTP endpoints.
package openapi

import (
	"encoding/json"

	"goa.design/goa/v3/expr"
	yaml "gopkg.in/yaml.v3"
)

type (
	// V3 represents an instance of a Swagger object.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.2.md
	V3 struct {
		// OpenAPI must contain the semantic version of the OpenAPI Specification Version
		OpenAPI      string                 `json:"openapi" yaml:"openapi"`
		Info         *Info                  `json:"info" yaml:"info"`
		Servers      []*Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
		Paths        map[string]interface{} `json:"paths" yaml:"paths"`
		Components   *Components            `json:"components,omitempty" yaml:"components,omitempty"`
		Security     []*Security            `json:"security,omitempty" yaml:"componsecurityents,omitempty"`
		Tags         []*Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
		ExternalDocs *ExternalDocs          `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
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

	// Server is an object representing a Server.
	Server struct {
		URL         string                     `json:"url" yaml:"url"`
		Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
		Variables   map[string]*ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
		Extensions  map[string]interface{}     `json:"-" yaml:"-"`
	}

	// ServerVariable is an object for server URL template substitution.
	ServerVariable struct {
		Enum        []string               `json:"enum,omitempty" yaml:"enum,omitempty"`
		Default     string                 `json:"default" yaml:"default"`
		Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Extensions  map[string]interface{} `json:"-" yaml:"-"`
	}

	// Path holds the relative paths to the individual endpoints.
	Path struct {
		// Ref allows for an external definition of this path item.
		Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
		// Summary is intended to apply to all operations in this path.
		Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
		// Description is intended to apply to all operations in this path.
		// CommonMark syntax MAY be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
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
		// Trace defines a TRACE operation on tis path
		Trace *Operation `json:"trace,omitempty" yaml:"trace,omitempty"`
		// Servers is an alternative server array to service all operations in this path.
		Servers []*Server `json:"servers,omitempty" yaml:"servers,omitempty"`
		// Parameters is the list of parameters that are applicable for all the operations
		// described under this path.
		Parameters []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
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
		// Deprecated specifies that a parameter is deprecated and should be transitioned
		// out of usage
		Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
		// AllowEmptyValue sets the ability to pass empty valued parameters. This is valid only
		// for query paramters and allows sending a parameter with an empty value.
		AllowEmptyValue bool `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
		// Style describes how the parameter value will be serialized depending on the type
		// of the parameter value. Default values of `style` based on the values of `in`:
		/// for query - form; for path - simple; for header - simple; for cookie - form.
		// In order to support common ways of serializing simple parameters, a set of style values
		// are defined. https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#style-values
		Style string `json:"style,omitempty" yaml:"style,omitempty"`
		// Explode generates separate parameters if parameter value is of type
		// array or object
		Explode bool `json:"explode,omitempty" yaml:"explode,omitempty"`
		// AllowReserved determines whether the parameter value should allow
		// reserved characters as defined by RFC3986. This property only applies to parameters
		// with an in value of query. The default value is false.
		AllowReserved bool `json:"allowreserved,omitempty" yaml:"allowreserved,omitempty"`
		// Schema defining the type used for the body parameter, only if "in" is body
		Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
		// Example of the media type. The example should match the specified schema and encoding
		// properties if present. The example object is mutually exclusive of the examples object.
		Example interface{} `json:"example,omitempty" yaml"example,omitempty"`
		// Examples of the media type. Each example should contain a value in the correct format
		// as specified in the parameter encoding. The examples object is mutually exclusive of the example object.
		Examples map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`
		// Content is a map containing the representations for the parameter. The key is the media type and
		// the value describes it. The map must only contain one entry.
		// // A parameter must contain either a schema property, or a content property, but not both.
		Content map[string]*MediaType `json:"content,omitempty" yaml:"content,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Example describes an example object
	Example struct {
		// Summary contains short description of the example
		Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
		// Description contains long description of the example
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// Value field and externalValue field are mutually exclusive
		// Value contains embedded literal example. To represent examples of media types that
		// cannot naturally represented in JSON or YAML, use a string value to contain the example,
		// escaping where necessary.
		Value interface{} `json:"value,omitempty" yaml:"value,omitempty"`
		// ExternalValue contains URL that points to the literal example. This provides the capability
		// to reference examples that cannot easily be included in JSON or YAML documents.
		ExternalValue string `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// MediaType struct provides schema and examples for the media type.
	MediaType struct {
		// Schema defines the type used for request body
		Schema *Schema
		// Example of the media type. The example object should be in the correct
		// format as specified by the media type. The example object is mutually exclusive of the examples object.
		Example interface{}
		// Examples of the media type. Each example object should match
		// the media type and specified schema if present.
		Examples map[string]*Example
		// Encoding is a map between a property name and its encoding information.
		// The key, being the property name, must exist in the schema as a property.
		Encoding map[string]*Encoding
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Components hols set of reusable objects. All objects defined within the components object will have no
	// effect on the API unless they are explicitly referenced from properties outside the components object
	Components struct {
		// Schemas is an object to hold reusable schema objects
		Schemas map[string]*Schema `json:"schemas,omitempty" yaml:"schemas,omitempty"`
		// Responses is an object to hold reusable response objects
		Responses map[string]*Response `json:"responses,omitempty" yaml:"responses,omitempty"`
		// Parameters is an object to hold reusable parameter objects
		Parameters map[string]*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		// Examples is an object to hold reusable example objects
		Examples map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`
		// RequestBodies is an object to hold reusable request bodies
		RequestBodies map[string]*RequestBody `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
		// Headers is an object to hold reusable header objects
		Headers map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
		// SecuritySchemes is an object to hold reusable securityScheme objects
		SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
		// links is an object to hold reusable link objects
		Links map[string]*Link `json:"links,omitempty" yaml:"links,omitempty"`
		// CallBacks is an object to hold reusable callback objects
		CallBacks map[string]*Callback `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	}

	// Response describes an operation response.
	Response struct {
		// Description of the response. GFM syntax can be used for rich text representation.
		Description string `json:"description" yaml:"description"`
		// Headers is a list of headers that are sent with the response.
		Headers map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
		// Content is a map containing descriptions of potential response payloads.
		Content map[string]*MediaType `json:"content,omitempty" yaml:"content,omitempty"`
		// Links contains map of operation links that can be followed from the response.
		// Key of the map is a short name for the link following namingg constrainsts of the names
		// for component objects
		Links map[string]*Link `json:"links,omitempty" yaml:"links,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}
	// Link represents a possible design-time link for a response.
	Link struct {
		OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
		OperationID  string                 `json:operationId,omitempty" yaml:"operationId,omitempty`
		Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
		Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Server       *Server                `json:"server,omitempty" yaml:"server,omitempty"`
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
		ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
		// OperationID is a unique string used to identify the operation.
		OperationID string `json:"operationId,omitempty" yaml:"operationId,omitempty"`
		// Parameters is a list of parameters that are applicable for this operation.
		Parameters []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		// RequestBody describes a single request body. requestBody is only supported in HTTP methods
		// where the HTTP 1.1 specification RFC7231 has explicitly defined semantics for request bodies.
		RequestBody *RequestBody `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
		// Responses is the list of possible responses as they are returned from executing
		// this operation.
		Responses map[string]*Response `json:"responses,omitempty" yaml:"responses,omitempty"`
		// Schemes is the transfer protocol for the operation.
		Schemes []string `json:"schemes,omitempty" yaml:"schemes,omitempty"`
		// Deprecated declares this operation to be deprecated.
		Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
		// Security is a declaration of which security schemes are applied for this operation.
		Security []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// ExternalDocs allows referencing an external document for extended
	// documentation.
	ExternalDocs struct {
		// Description is a short description of the target documentation.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// URL for the target documentation.
		URL string `json:"url" yaml:"url"`
	}

	RequestBody struct {
	}

	SecurityRequirement struct {
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

	// SecurityScheme allows the definition of a security scheme that can be used by the
	// operations. Supported schemes are basic authentication, an API key (either as a header or
	// as a query parameter) and OAuth2's common flows (implicit, password, application and
	// access code).
	SecurityScheme struct {
		// Type of the security scheme. Valid values are "basic", "apiKey" or "oauth2"
		// or "openIdConnect".
		Type string `json:"type" yaml:"type"`
		// Description for security scheme. CommonMark Syntax(https://spec.commonmark.org/)
		// may be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// Name of the header query or cookie parameter to be used when type is "apiKey".
		Name string `json:"name" yaml:"name"`
		// In is the location of the API key when type is "apiKey".
		// Valid values are "query", "header" or "cookie"
		In string `json:"in" yaml:"in"`
		// Scheme is the name of the HTTP Authorization scheme to be used in the Authorization header
		// as defined in RFC7235(https://tools.ietf.org/html/rfc7235#section-5.1)
		Scheme string `json:"scheme" yaml:"scheme"`
		// A hint to the client to identify how the bearer token is formatted.
		// Bearer tokens are usually generated by an authorization server,
		// so this information is primarily for documentation purposes.
		BearerFormat string `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
		// Flow is the flow used by the OAuth2 security scheme when type is "oauth2"
		// Valid values are "implicit", "password", "application" or "accessCode".
		Flow OAuthFlows `json:"flow,omitempty" yaml:"flow,omitempty"`
		// OpenId Connect URL to discover OAuth2 configuration values.
		// This MUST be in the form of a URL.
		OpenIDConnectURL string `json:"openIdConnect" yaml:"openIdConnect"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// OAuthFlows Allows configuration of the supported OAuth Flows.
	OAuthFlows struct {
		// Implicit for OAuth Implicit flow configuration
		Implicit OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
		// Password for OAuth Resource Owner Password flow configuration
		Password OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
		// ClientCredentials for OAuth Client Credentials flow
		ClientCredentials OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
		// AuthorizationCode for OAuth AuthorizationCode flow
		AuthorizationCode OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
	}

	// OAuthFlow contians configuration details for a supported OAuth Flow
	OAuthFlow struct {
		// AuthorizationURL for the OAuth flow. This MUST be in the form of a URL
		// Applies to "implicit" and "authorizationCode" flows.
		AuthorizationURL string `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
		// TokenURL for the OAuth flow. This MUST be in the form of a URL.
		// Applies to "password", "clientCredentials" and "authorizationCode" flows
		TokenURL string `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
		// RefreshURL for the URL for obtaining refresh tokens. This must be in the
		// form of URL. Applies to OAuth2 flow.
		RefreshURL string `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
		// Scopes list the  available scopes for OAuth2 security scheme. A map between the scope name and
		// a short description. Applies to OAuth2 flow.
		Scopes map[string]string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	}
	// Scope corresponds to an available scope for an OAuth2 security scheme.
	Scope struct {
		// Description for scope
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
	}

	// ExternalDocs allows referencing an external document for extended
	// documentation.
	ExternalDocs struct {
		// Description is a short description of the target documentation.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// URL for the target documentation.
		URL string `json:"url" yaml:"url"`
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

	// Tag allows adding meta data to a single tag that is used by the Operation Object. It is
	// not mandatory to have a Tag Object per tag used there.
	Tag struct {
		// Name of the tag.
		Name string `json:"name,omitempty" yaml:"name,omitempty"`
		// Description is a short description of the tag.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		// ExternalDocs is additional external documentation for this tag.
		ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// These types are used in marshalJSON() to avoid recursive call of json.Marshal().
	_Info               Info
	_Path               Path
	_Operation          Operation
	_Parameter          Parameter
	_Response           Response
	_SecurityDefinition SecurityDefinition
	_Tag                Tag
)

func marshalJSON(v interface{}, extensions map[string]interface{}) ([]byte, error) {
	marshaled, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(extensions) == 0 {
		return marshaled, nil
	}
	var unmarshaled interface{}
	if err := json.Unmarshal(marshaled, &unmarshaled); err != nil {
		return nil, err
	}
	asserted := unmarshaled.(map[string]interface{})
	for k, v := range extensions {
		asserted[k] = v
	}
	merged, err := json.Marshal(asserted)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// MarshalJSON returns the JSON encoding of i.
func (i Info) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Info(i), i.Extensions)
}

// MarshalJSON returns the JSON encoding of p.
func (p Path) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Path(p), p.Extensions)
}

// MarshalJSON returns the JSON encoding of o.
func (o Operation) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Operation(o), o.Extensions)
}

// MarshalJSON returns the JSON encoding of p.
func (p Parameter) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Parameter(p), p.Extensions)
}

// MarshalJSON returns the JSON encoding of r.
func (r Response) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Response(r), r.Extensions)
}

// MarshalJSON returns the JSON encoding of s.
func (s SecurityDefinition) MarshalJSON() ([]byte, error) {
	return marshalJSON(_SecurityDefinition(s), s.Extensions)
}

// MarshalJSON returns the JSON encoding of t.
func (t Tag) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Tag(t), t.Extensions)
}

func marshalYAML(v interface{}, extensions map[string]interface{}) (interface{}, error) {
	if len(extensions) == 0 {
		return v, nil
	}
	marshaled, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var unmarshaled map[string]interface{}
	if err := yaml.Unmarshal(marshaled, &unmarshaled); err != nil {
		return nil, err
	}
	for k, v := range extensions {
		unmarshaled[k] = v
	}
	return unmarshaled, nil
}

// MarshalYAML returns value which marshaled in place of the original value
func (i Info) MarshalYAML() (interface{}, error) {
	return marshalYAML(_Info(i), i.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (p Path) MarshalYAML() (interface{}, error) {
	return marshalYAML(_Path(p), p.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (o Operation) MarshalYAML() (interface{}, error) {
	return marshalYAML(_Operation(o), o.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (p Parameter) MarshalYAML() (interface{}, error) {
	return marshalYAML(_Parameter(p), p.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (r Response) MarshalYAML() (interface{}, error) {
	return marshalYAML(_Response(r), r.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (s SecurityDefinition) MarshalYAML() (interface{}, error) {
	return marshalYAML(_SecurityDefinition(s), s.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (t Tag) MarshalYAML() (interface{}, error) {
	return marshalYAML(_Tag(t), t.Extensions)
}
