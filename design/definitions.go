package design

import (
	"path/filepath"
	"regexp"
	"sort"
)

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
		ResponseTemplateFuncs map[string]func(params ...string) *ResponseTemplateDefinition
		// User types
		UserTypes []*UserTypeDefinition
		// Media types
		MediaTypes map[string]*MediaTypeDefinition
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
		// Media type used to render response
		MediaType string
		// Header definitions, values may be a mix of strings and regexps
		Headers map[string]interface{}
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
		// Map of possible response definitions indexed by name
		Responses map[string]*ResponseDefinition
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

	// ResourceIterator is the type of functions given to IterateResources.
	ResourceIterator func(r *ResourceDefinition) error

	// MediaTypeIterator is the type of functions given to IterateMediaTypes.
	MediaTypeIterator func(m *MediaTypeDefinition) error

	// ActionIterator is the type of functions given to IterateActions.
	ActionIterator func(a *ActionDefinition) error
)

// IterateResources calls the given iterator passing in each resource sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateResources returns that
// error.
func (a *APIDefinition) IterateResources(it ResourceIterator) error {
	names := make([]string, len(a.Resources))
	i := 0
	for n := range a.Resources {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(a.Resources[n]); err != nil {
			return err
		}
	}
	return nil
}

// IterateMediaTypes calls the given iterator passing in each media type sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateMediaTypes returns that
// error.
func (a *APIDefinition) IterateMediaTypes(it MediaTypeIterator) error {
	names := make([]string, len(a.MediaTypes))
	i := 0
	for n := range a.MediaTypes {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(a.MediaTypes[n]); err != nil {
			return err
		}
	}
	return nil
}

// IterateActions calls the given iterator passing in each resource action sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateActions returns that
// error.
func (r *ResourceDefinition) IterateActions(it ActionIterator) error {
	names := make([]string, len(r.Actions))
	i := 0
	for n := range r.Actions {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(r.Actions[n]); err != nil {
			return err
		}
	}
	return nil
}

// CanonicalPathAndParams computes the canonical path and parameters from the canonical action and
// the parents.
// It returns the empty string and nil if the resource or any of its parents has no canonical
// action.
func (r *ResourceDefinition) CanonicalPathAndParams() (path string, params []string) {
	if r.CanonicalAction == "" {
		return "", nil
	}
	ca, ok := r.Actions[r.CanonicalAction]
	if !ok {
		return
	}
	if len(ca.Routes) == 0 {
		return
	}
	var parentPath string
	var parentParams []string
	if r.ParentName != "" {
		parent, ok := Design.Resources[r.ParentName]
		if !ok {
			return
		}
		parentPath, parentParams = parent.CanonicalPathAndParams()
		if parentPath == "" {
			return
		}
	}
	path = filepath.Join(parentPath, ca.Routes[0].Path)
	params = append(parentParams, ca.Routes[0].Params()...)
	return
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

// Params returns the route parameters.
// For example for the route "GET /foo/:fooID" Params returns []string{"fooID"}.
func (r *RouteDefinition) Params() []string {
	matches := ParamsRegex.FindAllStringSubmatch(r.Path, -1)
	params := make([]string, len(matches))
	for i, m := range matches {
		params[i] = m[1]
	}
	return params
}
