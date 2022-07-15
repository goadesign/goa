package openapiv3

import "goa.design/goa/v3/http/codegen/openapi"

type (
	// OpenAPI is a data structure that encodes the information needed to
	// generate an OpenAPI specification as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md
	OpenAPI struct {
		OpenAPI      string                 `json:"openapi" yaml:"openapi"` // Required
		Info         *Info                  `json:"info" yaml:"info"`       // Required
		Servers      []*Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
		Paths        map[string]*PathItem   `json:"paths" yaml:"paths"` // Required
		Components   *Components            `json:"components,omitempty" yaml:"components,omitempty"`
		Tags         []*openapi.Tag         `json:"tags,omitempty" yaml:"tags,omitempty"`
		Security     []map[string][]string  `json:"security,omitempty" yaml:"security,omitempty"`
		ExternalDocs *openapi.ExternalDocs  `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
		Extensions   map[string]interface{} `json:"-" yaml:"-"`
	}

	// Info represents an OpenAPI Info object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#infoObject
	Info struct {
		Title          string                 `json:"title" yaml:"title"` // Required
		Description    string                 `json:"description,omitempty" yaml:"description,omitempty"`
		TermsOfService string                 `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
		Contact        *Contact               `json:"contact,omitempty" yaml:"contact,omitempty"`
		License        *License               `json:"license,omitempty" yaml:"license,omitempty"`
		Version        string                 `json:"version" yaml:"version"` // Required
		Extensions     map[string]interface{} `json:"-" yaml:"-"`
	}

	// Server represents an OpenAPI Server object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#serverObject
	Server struct {
		URL         string                     `json:"url" yaml:"url"`
		Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
		Variables   map[string]*ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
	}

	// PathItem represents an OpenAPI Path Item object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#pathItemObject
	PathItem struct {
		Ref         string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
		Summary     string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
		Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Connect     *Operation             `json:"connect,omitempty" yaml:"connect,omitempty"`
		Delete      *Operation             `json:"delete,omitempty" yaml:"delete,omitempty"`
		Get         *Operation             `json:"get,omitempty" yaml:"get,omitempty"`
		Head        *Operation             `json:"head,omitempty" yaml:"head,omitempty"`
		Options     *Operation             `json:"options,omitempty" yaml:"options,omitempty"`
		Patch       *Operation             `json:"patch,omitempty" yaml:"patch,omitempty"`
		Post        *Operation             `json:"post,omitempty" yaml:"post,omitempty"`
		Put         *Operation             `json:"put,omitempty" yaml:"put,omitempty"`
		Trace       *Operation             `json:"trace,omitempty" yaml:"trace,omitempty"`
		Servers     []*Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
		Parameters  []*ParameterRef        `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		Extensions  map[string]interface{} `json:"-" yaml:"-"`
	}

	// Components represents an OpenAPI Components object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#componentsObject
	Components struct {
		Schemas         map[string]*openapi.Schema    `json:"schemas,omitempty" yaml:"schemas,omitempty"`
		Parameters      map[string]*ParameterRef      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		Headers         map[string]*HeaderRef         `json:"headers,omitempty" yaml:"headers,omitempty"`
		RequestBodies   map[string]*RequestBodyRef    `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
		Responses       map[string]*ResponseRef       `json:"responses,omitempty" yaml:"responses,omitempty"`
		SecuritySchemes map[string]*SecuritySchemeRef `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
		Examples        map[string]*ExampleRef        `json:"examples,omitempty" yaml:"examples,omitempty"`
		Links           map[string]*LinkRef           `json:"links,omitempty" yaml:"links,omitempty"`
		Callbacks       map[string]*CallbackRef       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
		Extensions      map[string]interface{}        `json:"-" yaml:"-"`
	}

	// Contact represents an OpenAPI Contact object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#contactObject
	Contact struct {
		Name       string                 `json:"name,omitempty" yaml:"name,omitempty"`
		URL        string                 `json:"url,omitempty" yaml:"url,omitempty"`
		Email      string                 `json:"email,omitempty" yaml:"email,omitempty"`
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// License represents an OpenAPI License object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#licenseObject
	License struct {
		Name       string                 `json:"name" yaml:"name"` // Required
		URL        string                 `json:"url,omitempty" yaml:"url,omitempty"`
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// ServerVariable represents an OpenAPI Server Variable object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#serverVariableObject
	ServerVariable struct {
		Enum        []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
		Default     interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
		Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	}

	// Operation represents an OpenAPI Operation object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#operationObject
	Operation struct {
		Tags         []string                `json:"tags,omitempty" yaml:"tags,omitempty"`
		Summary      string                  `json:"summary,omitempty" yaml:"summary,omitempty"`
		Description  string                  `json:"description,omitempty" yaml:"description,omitempty"`
		OperationID  string                  `json:"operationId,omitempty" yaml:"operationId,omitempty"`
		Parameters   []*ParameterRef         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		RequestBody  *RequestBodyRef         `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
		Responses    map[string]*ResponseRef `json:"responses" yaml:"responses"` // Required
		Callbacks    map[string]*CallbackRef `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
		Deprecated   bool                    `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
		Security     []map[string][]string   `json:"security,omitempty" yaml:"security,omitempty"`
		Servers      []*Server               `json:"servers,omitempty" yaml:"servers,omitempty"`
		ExternalDocs *openapi.ExternalDocs   `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
		Extensions   map[string]interface{}  `json:"-" yaml:"-"`
	}

	// Parameter represents an OpenAPI Parameter object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#parameterObject
	Parameter struct {
		Name            string                 `json:"name,omitempty" yaml:"name,omitempty"`
		In              string                 `json:"in,omitempty" yaml:"in,omitempty"`
		Description     string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Style           string                 `json:"style,omitempty" yaml:"style,omitempty"`
		Explode         *bool                  `json:"explode,omitempty" yaml:"explode,omitempty"`
		AllowEmptyValue bool                   `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
		AllowReserved   bool                   `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
		Deprecated      bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
		Required        bool                   `json:"required,omitempty" yaml:"required,omitempty"`
		Schema          *openapi.Schema        `json:"schema,omitempty" yaml:"schema,omitempty"`
		Example         interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
		Examples        map[string]*ExampleRef `json:"examples,omitempty" yaml:"examples,omitempty"`
		Content         map[string]*MediaType  `json:"content,omitempty" yaml:"content,omitempty"`
		Extensions      map[string]interface{} `json:"-" yaml:"-"`
	}

	// Response represents an OpenAPI Response object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#responseObject
	Response struct {
		Description *string                `json:"description,omitempty" yaml:"description,omitempty"`
		Headers     map[string]*HeaderRef  `json:"headers,omitempty" yaml:"headers,omitempty"`
		Content     map[string]*MediaType  `json:"content,omitempty" yaml:"content,omitempty"`
		Links       map[string]*LinkRef    `json:"links,omitempty" yaml:"links,omitempty"`
		Extensions  map[string]interface{} `json:"-" yaml:"-"`
	}

	// MediaType represents an OpenAPI Media Type object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#mediaTypeObject
	MediaType struct {
		Schema     *openapi.Schema        `json:"schema,omitempty" yaml:"schema,omitempty"`
		Example    interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
		Examples   map[string]*ExampleRef `json:"examples,omitempty" yaml:"examples,omitempty"`
		Encoding   map[string]*Encoding   `json:"encoding,omitempty" yaml:"encoding,omitempty"`
		Extensions map[string]interface{} `json:"-" yaml:"-"`
	}

	// Encoding represents an OpenAPI Encoding object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#encodingObject
	Encoding struct {
		ContentType   string                 `json:"contentType,omitempty" yaml:"contentType,omitempty"`
		Headers       map[string]*HeaderRef  `json:"headers,omitempty" yaml:"headers,omitempty"`
		Style         string                 `json:"style,omitempty" yaml:"style,omitempty"`
		Explode       *bool                  `json:"explode,omitempty" yaml:"explode,omitempty"`
		AllowReserved bool                   `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
		Extensions    map[string]interface{} `json:"-" yaml:"-"`
	}

	// Header represents an OpenAPI Header object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#headerObject
	Header struct {
		Description     string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Style           string                 `json:"style,omitempty" yaml:"style,omitempty"`
		Explode         *bool                  `json:"explode,omitempty" yaml:"explode,omitempty"`
		AllowEmptyValue bool                   `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
		AllowReserved   bool                   `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
		Deprecated      bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
		Required        bool                   `json:"required,omitempty" yaml:"required,omitempty"`
		Schema          *openapi.Schema        `json:"schema,omitempty" yaml:"schema,omitempty"`
		Example         interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
		Examples        map[string]*ExampleRef `json:"examples,omitempty" yaml:"examples,omitempty"`
		Content         map[string]*MediaType  `json:"content,omitempty" yaml:"content,omitempty"`
		Extensions      map[string]interface{} `json:"-" yaml:"-"`
	}

	// Link represents an OpenAPI Link object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#linkObject
	Link struct {
		OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
		OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
		Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		Server       *Server                `json:"server,omitempty" yaml:"server,omitempty"`
		RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
		Extensions   map[string]interface{} `json:"-" yaml:"-"`
	}

	// Example represents an OpenAPI Example object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#exampleObject
	Example struct {
		Summary       string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
		Description   string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Value         interface{}            `json:"value,omitempty" yaml:"value,omitempty"`
		ExternalValue string                 `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
		Extensions    map[string]interface{} `json:"-" yaml:"-"`
	}

	// RequestBody represents an OpenAPI RequestBody object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#requestBodyObject
	RequestBody struct {
		Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Required    bool                   `json:"required,omitempty" yaml:"required,omitempty"`
		Content     map[string]*MediaType  `json:"content,omitempty" yaml:"content,omitempty"`
		Extensions  map[string]interface{} `json:"-" yaml:"-"`
	}

	// SecurityScheme represents an OpenAPI SecurityScheme object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#securitySchemeObject
	SecurityScheme struct {
		Type         string                 `json:"type,omitempty" yaml:"type,omitempty"`
		Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
		Name         string                 `json:"name,omitempty" yaml:"name,omitempty"`
		In           string                 `json:"in,omitempty" yaml:"in,omitempty"`
		Scheme       string                 `json:"scheme,omitempty" yaml:"scheme,omitempty"`
		BearerFormat string                 `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
		Flows        *OAuthFlows            `json:"flows,omitempty" yaml:"flows,omitempty"`
		Extensions   map[string]interface{} `json:"-" yaml:"-"`
	}

	// OAuthFlows represents an OpenAPI OAuthFlows object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#oauthFlowsObject
	OAuthFlows struct {
		Implicit          *OAuthFlow             `json:"implicit,omitempty" yaml:"implicit,omitempty"`
		Password          *OAuthFlow             `json:"password,omitempty" yaml:"password,omitempty"`
		ClientCredentials *OAuthFlow             `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
		AuthorizationCode *OAuthFlow             `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
		Extensions        map[string]interface{} `json:"-" yaml:"-"`
	}

	// OAuthFlow represents an OpenAPI OAuthFlow object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#oauthFlowObject
	OAuthFlow struct {
		AuthorizationURL string                 `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
		TokenURL         string                 `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
		RefreshURL       string                 `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
		Scopes           map[string]string      `json:"scopes" yaml:"scopes"`
		Extensions       map[string]interface{} `json:"-" yaml:"-"`
	}

	// These types are used in openapi.MarshalJSON() to avoid recursive call of json.Marshal().
	_Info           Info
	_PathItem       PathItem
	_Operation      Operation
	_Parameter      Parameter
	_Response       Response
	_SecurityScheme SecurityScheme
)

// MediaType implements exampler
func (m *MediaType) setExample(val interface{})             { m.Example = val }
func (m *MediaType) setExamples(val map[string]*ExampleRef) { m.Examples = val }

// Header implements exampler
func (h *Header) setExample(val interface{})             { h.Example = val }
func (h *Header) setExamples(val map[string]*ExampleRef) { h.Examples = val }

// Parameter implements exampler
func (p *Parameter) setExample(val interface{})             { p.Example = val }
func (p *Parameter) setExamples(val map[string]*ExampleRef) { p.Examples = val }

// MarshalJSON returns the JSON encoding of i.
func (i Info) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_Info(i), i.Extensions)
}

// MarshalJSON returns the JSON encoding of p.
func (p PathItem) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_PathItem(p), p.Extensions)
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
func (s SecurityScheme) MarshalJSON() ([]byte, error) {
	return openapi.MarshalJSON(_SecurityScheme(s), s.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (i Info) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_Info(i), i.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (p PathItem) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_PathItem(p), p.Extensions)
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
func (s SecurityScheme) MarshalYAML() (interface{}, error) {
	return openapi.MarshalYAML(_SecurityScheme(s), s.Extensions)
}
