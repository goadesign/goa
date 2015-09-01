package design

import "regexp"

var (
	// Design is the API definition created via DSL.
	Design *APIDefinition

	// ParamsRegex is the regular expression used to capture path parameters.
	ParamsRegex = regexp.MustCompile("/:([^/]*)/")
)

type (
	// APIDefinition defines the global properties of the API.
	APIDefinition struct {
		// API name
		Name string
		// API Title
		Title string
		// API description
		Description string // API description
		// Common base path to all API actions
		BasePath string
		// Common path parameters to all API actions
		BaseParams *AttributeDefinition
		// Exposed resources indexed by name
		Resources map[string]*ResourceDefinition
		// Traits available to all API resources and actions indexed by name
		Traits map[string]*TraitDefinition
		// Response templates available to all API actions indexed by name
		ResponseTemplates map[string]*ResponseTemplateDefinition
		// Response template factories available to all API actions indexed by name
		ResponseTemplateFuncs map[string]func(params ...string)
		// User types
		UserTypes []*UserTypeDefinition
		// Media types
		MediaTypes []*MediaTypeDefinition
	}

	// ResourceDefinition describes a REST resource.
	// It defines both a media type and a set of actions that can be executed through HTTP
	// requests.
	// A resource is versioned so that multiple versions of the same resource may be exposed
	// by the API.
	ResourceDefinition struct {
		// Resource name
		Name string
		// Common URL prefix to all resource action HTTP requests
		BasePath string
		// Object describing each parameter that appears in BasePath if any
		BaseParams *AttributeDefinition
		// Name of parent resource if any
		ParentName string
		// Optional description
		Description string
		// Optional version
		Version string
		// Default media type, describes the resource attributes
		MediaType string
		// Exposed resource actions indexed by name
		Actions map[string]*ActionDefinition
		// Action with canonical resource path
		CanonicalAction string
	}

	// TypeDefinition describes a named data structure to be used e.g. for request payloads.
	TypeDefinition struct {
		// A media type definition is a JSON object
		Object
		// Name used in generated code
		Name string
		// Optional description
		Description string
	}

	// ResponseTemplateDefinition describes a response template that can be used by API actions
	// to define their responses.
	ResponseTemplateDefinition struct {
		// Name used in generated code
		Name string
		// Optional description
		Description string
		// HTTP status
		Status int
		// Media type used to render link
		MediaType string
	}

	// ActionDefinition defines a resource action.
	// It defines both an HTTP endpoint and the shape of HTTP requests and responses made to
	// that endpoint.
	// The shape of requests is defined via "parameters", there are path parameters
	// (i.e. portions of the URL that define parameter values), query string
	// parameters and a payload parameter (request body).
	ActionDefinition struct {
		// Action name, e.g. "create"
		Name string
		// Action description, e.g. "Creates a task"
		Description string
		// Parent resource
		Resource *ResourceDefinition
		// Action routes
		Routes []*RouteDefinition
		// Set of possible response definitions
		Responses []*ResponseDefinition
		// Path and query string parameters
		Params *AttributeDefinition
		// Payload blueprint (request body) if any
		Payload *UserTypeDefinition
		// Request headers that need to be made available to action
		Headers *AttributeDefinition
	}

	// AttributeDefinition defines a JSON object member with optional description, default
	// value and validations.
	AttributeDefinition struct {
		// Attribute type
		Type DataType
		// Optional description
		Description string
		// Optional validation functions
		Validations []ValidationDefinition
		// Optional member default value
		DefaultValue interface{}
	}

	// LinkDefinition defines a media type link, it specifies a URL to a related resource.
	LinkDefinition struct {
		// Link name
		Name string
		// Optional description
		Description string
		// Member used to render link
		Member *AttributeDefinition
		// Media type used to render link
		MediaType *MediaTypeDefinition
		// View used to render link if not "link"
		View string
	}

	// ViewDefinition defines which members and links to render when building a response.
	// The view is a JSON object whose property names must match the names of the parent media
	// type members.
	// The members fields are inherited from the parent media type but may be overridden.
	ViewDefinition struct {
		// Set of properties included in view
		Object
		// Name of view
		Name string
		// Links to render
		Links []string
		// Parent media type definition
		MediaType *MediaTypeDefinition
	}

	// ResponseDefinition defines a HTTP response status and optional validation rules.
	ResponseDefinition struct {
		// Response name
		Name string
		// HTTP status
		Status int
		// Response description
		Description string
		// Response body media type if any
		MediaType string
		// Response header validations
		Headers []*HeaderDefinition
	}

	// TraitDefinition defines a set of reusable properties.
	TraitDefinition struct {
		// Trait name
		Name string
		// Trait DSL
		Dsl func()
	}

	// RouteDefinition represents an action route.
	RouteDefinition struct {
		// Verb is the HTTP method, e.g. "GET", "POST", etc.
		Verb string
		// Path is the URL path e.g. "/tasks/:id"
		Path string
	}

	// HeaderDefinition define headers that need to be made available to the action.
	HeaderDefinition struct {
		// Header key, e.g. "X-Request-Id"
		Name string
		// Member describes headers including validations.
		Member *AttributeDefinition
	}

	// ValidationDefinition is the common interface for all validation data structures.
	// It doesn't expose any method and simply exists to help with documentation.
	ValidationDefinition interface{}

	// EnumValidationDefinition represents an enum validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor76.
	EnumValidationDefinition struct {
		Values []interface{}
	}

	// FormatValidationDefinition represents a format validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor104.
	FormatValidationDefinition struct {
		Format string
	}

	// MinimumValidationDefinition represents an minimum value validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor21.
	MinimumValidationDefinition struct {
		Min int
	}

	// MaximumValidationDefinition represents a maximum value validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor17.
	MaximumValidationDefinition struct {
		Max int
	}

	// MinLengthValidationDefinition represents an minimum length validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor29.
	MinLengthValidationDefinition struct {
		MinLength int
	}

	// MaxLengthValidationDefinition represents an maximum length validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor26.
	MaxLengthValidationDefinition struct {
		MaxLength int
	}

	// RequiredValidationDefinition represents a required validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor61.
	RequiredValidationDefinition struct {
		Names []string
	}
)

// MediaType returns the media type with the given name/identifier - nil if none exist.
func (a *APIDefinition) MediaType(name string) *MediaTypeDefinition {
	var mt *MediaTypeDefinition
	for _, m := range Design.MediaTypes {
		if m.Name == name {
			mt = m
			break
		}
	}
	return mt
}

// Parent returns the parent resource if any.
func (r *ResourceDefinition) Parent() *ResourceDefinition {
	if r.ParentName == "" {
		return nil
	}
	for _, res := range Design.Resources {
		if res.Name == r.ParentName {
			return res
		}
	}
	return nil
}

// IsRequired returns true if the given string matches the name of a required attribute, false
// otherwise.
// IsRequired panics if the type of a is not Object.
func (a *AttributeDefinition) IsRequired(attName string) bool {
	_, ok := a.Type.(Object)
	if !ok {
		panic("iterated attribute not an object")
	}
	for _, v := range a.Validations {
		if r, ok := v.(*RequiredValidationDefinition); ok {
			for _, name := range r.Names {
				if name == attName {
					return true
				}
			}
		}
	}
	return false
}
