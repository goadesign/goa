package design

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/julienschmidt/httprouter"
)

var (
	// Design is the API definition created via DSL.
	Design *APIDefinition

	// WildcardRegex is the regular expression used to capture path parameters.
	WildcardRegex = regexp.MustCompile(`/(?::|\*)([a-zA-Z0-9_]+)`)
)

type (
	// DSLDefinition is the common interface implemented by all definitions.
	DSLDefinition interface {
		// Context is used to build error messages that refer to the definition.
		Context() string
	}

	// APIDefinition defines the global properties of the API.
	APIDefinition struct {
		// API name
		Name string
		// API Title
		Title string
		// API description
		Description string // API description
		// API hostname
		Host string
		// Common base path to all API actions
		BasePath string
		// Common path parameters to all API actions
		BaseParams *AttributeDefinition
		// TermsOfService describes or links to the API terms of service
		TermsOfService string
		// Contact provides the API users with contact information
		Contact *ContactDefinition
		// License describes the API license
		License *LicenseDefinition
		// Docs points to the API external documentation
		Docs *DocsDefinition
		// Exposed resources indexed by name
		Resources map[string]*ResourceDefinition
		// Traits available to all API resources and actions indexed by name
		Traits map[string]*TraitDefinition
		// Responses available to all API actions indexed by name
		Responses map[string]*ResponseDefinition
		// Response template factories available to all API actions indexed by name
		ResponseTemplates map[string]*ResponseTemplateDefinition
		// Built-in responses
		DefaultResponses map[string]*ResponseDefinition
		// Built-in response templates
		DefaultResponseTemplates map[string]*ResponseTemplateDefinition
		// User types
		Types map[string]*UserTypeDefinition
		// Media types
		MediaTypes map[string]*MediaTypeDefinition
		// dsl contains the DSL used to create this definition if any.
		DSL func()
	}

	// ContactDefinition contains the API contact information.
	ContactDefinition struct {
		// Name of the contact person/organization
		Name string `json:"name,omitempty"`
		// Email address of the contact person/organization
		Email string `json:"email,omitempty"`
		// URL pointing to the contact information
		URL string `json:"url,omitempty"`
	}

	// LicenseDefinition contains the license information for the API.
	LicenseDefinition struct {
		// Name of license used for the API
		Name string `json:"name,omitempty"`
		// URL to the license used for the API
		URL string `json:"url,omitempty"`
	}

	// DocsDefinition points to external documentation.
	DocsDefinition []*struct {
		// Description of documentation.
		Description string `json:"description,omitempty"`
		// URL to documentation.
		URL string `json:"url,omitempty"`
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
		CanonicalActionName string
		// Map of response definitions that apply to all actions indexed by name.
		Responses map[string]*ResponseDefinition
		// Path and query string parameters that apply to all actions.
		Params *AttributeDefinition
		// Request headers that apply to all actions.
		Headers *AttributeDefinition
		// dsl contains the DSL used to create this definition if any.
		DSL func()
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
		// Response header definitions
		Headers *AttributeDefinition
		// Parent action or resource
		Parent DSLDefinition
	}

	// ResponseTemplateDefinition defines a response template.
	// A response template is a function that takes an arbitrary number
	// of strings and returns a response definition.
	ResponseTemplateDefinition struct {
		// Response template name
		Name string
		// Response template function
		Template func(params ...string) *ResponseDefinition
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
		Parent *ResourceDefinition
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
		// Attribute reference type if any
		Reference DataType
		// Optional description
		Description string
		// Optional validation functions
		Validations []ValidationDefinition
		// Optional member default value
		DefaultValue interface{}
		// Optional view used to render Attribute (only applies to media type attributes).
		View string
	}

	// LinkDefinition defines a media type link, it specifies a URL to a related resource.
	LinkDefinition struct {
		// Link name
		Name string
		// View used to render link if not "link"
		View string
		// URITemplate is the RFC6570 URI template of the link Href.
		URITemplate string

		// Parent media Type
		Parent *MediaTypeDefinition
	}

	// ViewDefinition defines which members and links to render when building a response.
	// The view is a JSON object whose property names must match the names of the parent media
	// type members.
	// The members fields are inherited from the parent media type but may be overridden.
	ViewDefinition struct {
		// Set of properties included in view
		*AttributeDefinition
		// Name of view
		Name string
		// Parent media Type
		Parent *MediaTypeDefinition
	}

	// TraitDefinition defines a set of reusable properties.
	TraitDefinition struct {
		// Trait name
		Name string
		// Trait DSL
		DSL func()
	}

	// RouteDefinition represents an action route.
	RouteDefinition struct {
		// Verb is the HTTP method, e.g. "GET", "POST", etc.
		Verb string
		// Path is the URL path e.g. "/tasks/:id"
		Path string
		// Parent is the action this route applies to.
		Parent *ActionDefinition
	}

	// ValidationDefinition is the common interface for all validation data structures.
	// It doesn't expose any method and simply exists to help with documentation.
	ValidationDefinition interface {
		DSLDefinition
	}

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

	// PatternValidationDefinition represents a pattern validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor33
	PatternValidationDefinition struct {
		Pattern string
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

	// UserTypeIterator is the type of functions given to IterateUserTypes.
	UserTypeIterator func(m *UserTypeDefinition) error

	// ActionIterator is the type of functions given to IterateActions.
	ActionIterator func(a *ActionDefinition) error

	// ResponseIterator is the type of functions given to IterateResponses.
	ResponseIterator func(r *ResponseDefinition) error
)

// Context returns the generic definition name used in error messages.
func (a *APIDefinition) Context() string {
	if a.Name != "" {
		return fmt.Sprintf("api %#v", a.Name)
	}
	return "unnamed API"
}

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

// IterateUserTypes calls the given iterator passing in each user type sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateUserTypes returns that
// error.
func (a *APIDefinition) IterateUserTypes(it UserTypeIterator) error {
	names := make([]string, len(a.Types))
	i := 0
	for n := range a.Types {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(a.Types[n]); err != nil {
			return err
		}
	}
	return nil
}

// IterateResponses calls the given iterator passing in each response sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateResponses returns that
// error.
func (a *APIDefinition) IterateResponses(it ResponseIterator) error {
	names := make([]string, len(a.Responses))
	i := 0
	for n := range a.Responses {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(a.Responses[n]); err != nil {
			return err
		}
	}
	return nil
}

// NewResourceDefinition creates a resource definition but does not
// execute the DSL.
func NewResourceDefinition(name string, dsl func()) *ResourceDefinition {
	return &ResourceDefinition{
		Name:      name,
		MediaType: "plain/text",
		DSL:       dsl,
	}
}

// Context returns the generic definition name used in error messages.
func (r *ResourceDefinition) Context() string {
	if r.Name != "" {
		return fmt.Sprintf("resource %#v", r.Name)
	}
	return "unnamed resource"
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

// CanonicalAction returns the canonical action of the resource if any.
// The canonical action is used to compute hrefs to resources.
func (r *ResourceDefinition) CanonicalAction() *ActionDefinition {
	name := r.CanonicalActionName
	if name == "" {
		name = "show"
	}
	ca, _ := r.Actions[name]
	return ca
}

// URITemplate returns a httprouter compliant URI template to this resource.
// The result is the empty string if the resource does not have a "show" action
// and does not define a different canonical action.
func (r *ResourceDefinition) URITemplate() string {
	ca := r.CanonicalAction()
	if ca == nil || len(ca.Routes) == 0 {
		return ""
	}
	return ca.Routes[0].FullPath()
}

// FullPath computes the base path to the resource actions concatenating the API and parent resource
// base paths as needed.
func (r *ResourceDefinition) FullPath() string {
	var basePath string
	if p := r.Parent(); p != nil {
		if ca := p.CanonicalAction(); ca != nil {
			if routes := ca.Routes; len(routes) > 0 {
				// Note: all these tests should be true at code generation time
				// as DSL validation makes sure that parent resources have a
				// canonical path.
				basePath = filepath.Join(routes[0].FullPath())
			}
		}
	} else {
		basePath = Design.BasePath
	}
	return httprouter.CleanPath(filepath.Join(basePath, r.BasePath))
}

// Parent returns the parent resource if any, nil otherwise.
func (r *ResourceDefinition) Parent() *ResourceDefinition {
	if r.ParentName != "" {
		if parent, ok := Design.Resources[r.ParentName]; ok {
			return parent
		}
	}
	return nil
}

// Context returns the generic definition name used in error messages.
func (c *ContactDefinition) Context() string {
	if c.Name != "" {
		return fmt.Sprintf("contact %s", c.Name)
	}
	return "unnamed contact"
}

// Context returns the generic definition name used in error messages.
func (l *LicenseDefinition) Context() string {
	if l.Name != "" {
		return fmt.Sprintf("license %s", l.Name)
	}
	return "unnamed license"
}

// Context returns the generic definition name used in error messages.
func (d DocsDefinition) Context() string {
	return fmt.Sprintf("documentation for %s", Design.Name)
}

// Context returns the generic definition name used in error messages.
func (t *UserTypeDefinition) Context() string {
	if t.TypeName != "" {
		return fmt.Sprintf("type %#v", t.TypeName)
	}
	return "unnamed type"
}

// Context returns the generic definition name used in error messages.
func (r *ResponseDefinition) Context() string {
	var prefix, suffix string
	if r.Name != "" {
		prefix = fmt.Sprintf("response %#v", r.Name)
	} else {
		prefix = "unnamed response"
	}
	if r.Parent != nil {
		suffix = fmt.Sprintf(" of %s", r.Parent.Context())
	}
	return prefix + suffix
}

// Dup returns a copy of the response definition.
func (r *ResponseDefinition) Dup() *ResponseDefinition {
	res := ResponseDefinition{
		Name:        r.Name,
		Status:      r.Status,
		Description: r.Description,
		MediaType:   r.MediaType,
	}
	if r.Headers != nil {
		res.Headers = r.Headers.Dup()
	}
	return &res
}

// Merge merges other into target. Only the fields of target that are not already set are merged.
func (r *ResponseDefinition) Merge(other *ResponseDefinition) {
	if other == nil {
		return
	}
	if r.Name == "" {
		r.Name = other.Name
	}
	if r.Status == 0 {
		r.Status = other.Status
	}
	if r.Description == "" {
		r.Description = other.Description
	}
	if r.MediaType == "" {
		r.MediaType = other.MediaType
	}
	if other.Headers != nil {
		otherHeaders := other.Headers.Type.ToObject()
		if len(otherHeaders) > 0 {
			if r.Headers == nil {
				r.Headers = &AttributeDefinition{Type: Object{}}
			}
			headers := r.Headers.Type.ToObject()
			for n, h := range otherHeaders {
				if _, ok := headers[n]; !ok {
					headers[n] = h
				}
			}
		}
	}
}

// Context returns the generic definition name used in error messages.
func (r *ResponseTemplateDefinition) Context() string {
	if r.Name != "" {
		return fmt.Sprintf("response template %#v", r.Name)
	}
	return "unnamed response template"
}

// Context returns the generic definition name used in error messages.
func (a *ActionDefinition) Context() string {
	var prefix, suffix string
	if a.Name != "" {
		suffix = fmt.Sprintf(" action %#v", a.Name)
	} else {
		suffix = " unnamed action"
	}
	if a.Parent != nil {
		prefix = a.Parent.Context()
	}
	return prefix + suffix
}

// AllParams returns the path and query string parameters of the action across all its routes.
func (a *ActionDefinition) AllParams() *AttributeDefinition {
	var res *AttributeDefinition
	if a.Params != nil {
		res = a.Params.Dup()
	} else {
		res = &AttributeDefinition{Type: Object{}}
	}
	res = res.Merge(a.Parent.BaseParams)
	res = res.Merge(Design.BaseParams)
	if p := a.Parent.Parent(); p != nil {
		res = res.Merge(p.CanonicalAction().AllParams())
	}
	return res
}

// AllParamNames returns the path and query string parameter names of the action across all its
// routes.
func (a *ActionDefinition) AllParamNames() []string {
	var params []string
	for _, r := range a.Routes {
		for _, p := range r.Params() {
			found := false
			for _, pa := range params {
				if pa == p {
					found = true
					break
				}
			}
			if !found {
				params = append(params, p)
			}
		}
	}
	sort.Strings(params)
	return params
}

// Context returns the generic definition name used in error messages.
func (a *AttributeDefinition) Context() string {
	return ""
}

// AllRequired returns the complete list of all required attribute names, nil
// if it doesn't have a RequiredValidationDefinition validation.
func (a *AttributeDefinition) AllRequired() []string {
	for _, v := range a.Validations {
		if r, ok := v.(*RequiredValidationDefinition); ok {
			return r.Names
		}
	}
	return nil
}

// IsRequired returns true if the given string matches the name of a required
// attribute, false otherwise.
func (a *AttributeDefinition) IsRequired(attName string) bool {
	for _, name := range a.AllRequired() {
		if name == attName {
			return true
		}
	}
	return false
}

// Dup returns a copy of the attribute definition.
// Note: the primitive underlying types are not duplicated for simplicity.
func (a *AttributeDefinition) Dup() *AttributeDefinition {
	valDup := make([]ValidationDefinition, len(a.Validations))
	for i, v := range a.Validations {
		valDup[i] = v
	}
	dupType := a.Type
	if dupType != nil {
		dupType = dupType.Dup()
	}
	dup := AttributeDefinition{
		Type:         dupType,
		Description:  a.Description,
		Validations:  valDup,
		DefaultValue: a.DefaultValue,
	}
	return &dup
}

// Merge merges the argument attributes into the target and returns the target overriding existing
// attributes with identical names.
// This only applies to attributes of type Object and Merge panics if the
// argument or the target is not of type Object.
func (a *AttributeDefinition) Merge(other *AttributeDefinition) *AttributeDefinition {
	if other == nil {
		return a
	}
	if a == nil {
		return other
	}
	left := a.Type.(Object)
	right := other.Type.(Object)
	if left == nil || right == nil {
		panic("cannot merge non object attributes") // bug
	}
	for n, v := range right {
		left[n] = v
	}
	return a
}

// Inherit merges the properties of existing target type attributes with the argument's.
// The algorithm is recursive so that child attributes are also merged.
func (a *AttributeDefinition) Inherit(parent *AttributeDefinition) {
	if a == nil || parent == nil {
		return
	}
	o := a.Type.ToObject()
	p := parent.Type.ToObject()
	if o == nil || p == nil {
		return
	}
	for _, v := range parent.Validations {
		found := false
		for _, vc := range a.Validations {
			if v == vc {
				found = true
				break
			}
		}
		if !found {
			a.Validations = append(a.Validations, parent)
		}
	}
	for n, att := range o {
		if patt, ok := p[n]; ok {
			if att.Description == "" {
				att.Description = patt.Description
			}
			for _, v := range patt.Validations {
				found := false
				for _, vc := range att.Validations {
					if v == vc {
						found = true
						break
					}
				}
				if !found {
					att.Validations = append(att.Validations, v)
				}
			}
			if att.DefaultValue == nil {
				att.DefaultValue = patt.DefaultValue
			}
			if att.View == "" {
				att.View = patt.View
			}
			if att.Type == nil {
				att.Type = patt.Type
			} else if co := att.Type.ToObject(); co != nil {
				if po := patt.Type.ToObject(); po != nil {
					for n, att := range co {
						if pcatt, ok := po[n]; ok {
							att.Inherit(pcatt)
						}
					}
				}
			}
		}
	}
}

// Context returns the generic definition name used in error messages.
func (l *LinkDefinition) Context() string {
	var prefix, suffix string
	if l.Name != "" {
		prefix = fmt.Sprintf("link %#v", l.Name)
	} else {
		prefix = "unnamed link"
	}
	if l.Parent != nil {
		suffix = fmt.Sprintf(" of %s", l.Parent.Context())
	}
	return prefix + suffix
}

// Attribute returns the linked attribute.
func (l *LinkDefinition) Attribute() *AttributeDefinition {
	p := l.Parent.ToObject()
	if p == nil {
		return nil
	}
	att, _ := p[l.Name]

	return att
}

// MediaType returns the media type of the linked attribute.
func (l *LinkDefinition) MediaType() *MediaTypeDefinition {
	att := l.Attribute()
	mt, _ := att.Type.(*MediaTypeDefinition)
	return mt
}

// Context returns the generic definition name used in error messages.
func (v *ViewDefinition) Context() string {
	var prefix, suffix string
	if v.Name != "" {
		prefix = fmt.Sprintf("view %#v", v.Name)
	} else {
		prefix = "unnamed view"
	}
	if v.Parent != nil {
		suffix = fmt.Sprintf(" of %s", v.Parent.Context())
	}
	return prefix + suffix
}

// Context returns the generic definition name used in error messages.
func (t *TraitDefinition) Context() string {
	if t.Name != "" {
		return fmt.Sprintf("trait %#v", t.Name)
	}
	return "unnamed trait"
}

// Context returns the generic definition name used in error messages.
func (r *RouteDefinition) Context() string {
	return fmt.Sprintf(`route %s "%s" of %s`, r.Verb, r.Path, r.Parent.Context())
}

// Params returns the route parameters.
// For example for the route "GET /foo/:fooID" Params returns []string{"fooID"}.
func (r *RouteDefinition) Params() []string {
	return ExtractWildcards(r.FullPath())
}

// FullPath returns the action full path computed by concatenating the API and resource base paths
// with the action specific path.
func (r *RouteDefinition) FullPath() string {
	if strings.HasPrefix(r.Path, "//") {
		return httprouter.CleanPath(r.Path[1:])
	}
	var base string
	if r.Parent != nil && r.Parent.Parent != nil {
		base = r.Parent.Parent.FullPath()
	}
	return httprouter.CleanPath(filepath.Join(base, r.Path))
}

// Context returns the generic definition name used in error messages.
func (v *EnumValidationDefinition) Context() string {
	return "enum validation"
}

// Context returns the generic definition name used in error messages.
func (f *FormatValidationDefinition) Context() string {
	return "format validation"
}

// Context returns the generic definition name used in error messages.
func (f *PatternValidationDefinition) Context() string {
	return "pattern validation"
}

// Context returns the generic definition name used in error messages.
func (m *MinimumValidationDefinition) Context() string {
	return "min value validation"
}

// Context returns the generic definition name used in error messages.
func (m *MaximumValidationDefinition) Context() string {
	return "max value validation"
}

// Context returns the generic definition name used in error messages.
func (m *MinLengthValidationDefinition) Context() string {
	return "min length validation"
}

// Context returns the generic definition name used in error messages.
func (m *MaxLengthValidationDefinition) Context() string {
	return "max length validation"
}

// Context returns the generic definition name used in error messages.
func (r *RequiredValidationDefinition) Context() string {
	return "required field validation"
}

// ExtractWildcards returns the names of the wildcards that appear in path.
func ExtractWildcards(path string) []string {
	matches := WildcardRegex.FindAllStringSubmatch(path, -1)
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}
