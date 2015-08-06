package design

import (
	"bytes"
	"fmt"
	"mime"
	"regexp"
	"sort"
	"strings"

	"bitbucket.org/pkg/inflect"
)

var (
	// Design is the API definition created via DSL.
	Design *APIDefinition

	// ParamsRegex is the regular expression used to capture path parameters.
	ParamsRegex = regexp.MustCompile("/:(.*)/")
)

type (
	// Definition is the interface implemnented by all the design definitions.
	Definition interface {
		// Validate returns nil if the definition is properly initialized (no required
		// field is missing, field formats are all correct etc.), an error otherwise.
		Validate() error
	}

	// APIDefinition defines the global properties of the API.
	APIDefinition struct {
		Name              string                         // API name
		Title             string                         // API Title
		Description       string                         // API description
		BasePath          string                         // Common base path to all API actions
		BaseParams        []*AttributeDefinition         // Common path parameters to all API actions
		Resources         map[string]*ResourceDefinition // Exposed resources indexed by name
		Traits            map[string]*TraitDefinition    // Traits available to all API resources and actions indexed by name
		ResponseTemplates map[string]*ResponseDefinition // Response templates available to all API actions indexed by name
	}

	// ResourceDefinition describes a REST resource.
	// It defines both a media type and a set of actions that can be executed through HTTP
	// requests.
	// A resource is versioned so that multiple versions of the same resource may be exposed
	// by the API.
	ResourceDefinition struct {
		Name            string                       // Resource name
		BasePath        string                       // Common URL prefix to all resource action HTTP requests
		BaseParams      *AttributeDefinition         // Object describing each parameter that appears in BasePath if any
		ParentName      string                       // Name of parent resource if any
		Description     string                       // Optional description
		Version         string                       // Optional version
		MediaType       *MediaTypeDefinition         // Default media type, describes the resource attributes
		Actions         map[string]*ActionDefinition // Exposed resource actions indexed by name
		CanonicalAction string                       // Action with canonical resource path
		computedParent  *ResourceDefinition          // cached computed parent resource
	}

	// MediaTypeDefinition describes the rendering of a resource using property and link
	// definitions. A property corresponds to a single member of the media type,
	// it has a name and a type as well as optional validation rules. A link has a
	// name and a URL that points to a related resource.
	// Media types also define views which describe which members and links to render when
	// building the response body for the corresponding view.
	MediaTypeDefinition struct {
		Object                                  // A media type definition is a JSON object
		Name         string                     // Name used in generated code
		Identifier   string                     // RFC 6838 Media type identifier
		Description  string                     // Optional description
		Resource     *ResourceDefinition        // Corresponding resource if any
		Links        map[string]*LinkDefinition // List of rendered links indexed by name (named hrefs to related resources)
		Views        map[string]*ViewDefinition // List of supported views indexed by name
		isCollection bool                       // Whether media type is for a collection of resources (true) or not (false)
	}

	// ActionDefinition defines a resource action.
	// It defines both an HTTP endpoint and the shape of HTTP requests and responses made to
	// that endpoint.
	// The shape of requests is defined via "parameters", there are path parameters
	// (i.e. portions of the URL that define parameter values), query string
	// parameters and a payload parameter (request body).
	ActionDefinition struct {
		Name        string                // Action name, e.g. "create"
		Description string                // Action description, e.g. "Creates a task"
		Resource    *ResourceDefinition   // Parent resource
		Routes      []*RouteDefinition    // Action routes
		Responses   []*ResponseDefinition // Set of possible response definitions
		Params      *AttributeDefinition  // Path and query string parameters
		Payload     *AttributeDefinition  // Payload blueprint (request body) if any
		Headers     *AttributeDefinition  // Request headers that need to be made available to action
	}

	// AttributeDefinition defines a JSON object member with optional description, default
	// value and validations.
	AttributeDefinition struct {
		Type         DataType               // Attribute type
		Description  string                 // Optional description
		Validations  []ValidationDefinition // Optional validation functions
		DefaultValue interface{}            // Optional member default value
	}

	// LinkDefinition defines a media type link, it specifies a URL to a related resource.
	LinkDefinition struct {
		Name        string               // Link name
		Description string               // Optional description
		Member      *AttributeDefinition // Member used to render link
		MediaType   *MediaTypeDefinition // Media type used to render link
		View        string               // View used to render link if not "link"
	}

	// ViewDefinition defines which members and links to render when building a response.
	// The view is a JSON object whose property names must match the names of the parent media
	// type members.
	// The members fields are inherited from the parent media type but may be overridden.
	ViewDefinition struct {
		Object                         // Set of properties included in view
		Name      string               // Name of view
		Links     []string             // Links to render
		MediaType *MediaTypeDefinition // Parent media type definition
	}

	// ResponseDefinition defines a HTTP response status and optional validation rules.
	ResponseDefinition struct {
		Name        string               // Response name
		Status      int                  // HTTP status
		Description string               // Response description
		MediaType   *MediaTypeDefinition // Response body media type if any
		Headers     []*HeaderDefinition  // Response header validations
	}

	// TraitDefinition defines a set of reusable properties.
	TraitDefinition struct {
		Name string // Trait name
		Dsl  func() // Trait DSL
	}

	// RouteDefinition represents an action route.
	RouteDefinition struct {
		Verb string // HTTP method, e.g. "GET", "POST", etc.
		Path string // URL path e.g. "/tasks/:id"
	}

	// HeaderDefinition define headers that need to be made available to the action.
	HeaderDefinition struct {
		Name   string               // Header key, e.g. "X-Request-Id"
		Member *AttributeDefinition // Header definition including validations
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

// Validate tests whether the API definition is consistent: all resource parent names resolve to
// an actual resource.
func (a *APIDefinition) Validate() error {
	for _, r := range a.Resources {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("Resource %s: %s", r.Name, err)
		}
		if r.ParentName == "" {
			continue
		}
		if r.Parent() == nil {
			return fmt.Errorf("Resource %s: Unknown parent resource %s", r.Name, r.ParentName)
		}
	}
	return nil
}

// Validate tests whether the resource definition is consistent: action names are valid and each action is
// valid.
func (r *ResourceDefinition) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("Resource name cannot be empty")
	}
	found := false
	for _, a := range r.Actions {
		if a.Name == r.CanonicalAction {
			found = true
		}
		if err := a.Validate(); err != nil {
			return fmt.Errorf("Action %s: %s", a.Name, err)
		}
	}
	if r.CanonicalAction != "" && !found {
		return fmt.Errorf("Unknown canonical action '%s'", r.CanonicalAction)
	}
	if r.BaseParams != nil {
		baseParams, ok := r.BaseParams.Type.(Object)
		if !ok {
			return fmt.Errorf("Invalid type for BaseParams, must be an Object")
		}
		vs := ParamsRegex.FindAllStringSubmatch(r.BasePath, -1)
		if len(vs) > 1 {
			vars := vs[1]
			if len(vars) != len(baseParams) {
				return fmt.Errorf("BasePath defines parameters %s but BaseParams has %d elements",
					strings.Join([]string{
						strings.Join(vars[:len(vars)-1], ", "),
						vars[len(vars)-1],
					}, " and "),
					len(baseParams),
				)
			}
			for _, v := range vars {
				found := false
				for n := range baseParams {
					if v == n {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("Variable %s from base path %s does not match any parameter from BaseParams",
						v, r.BasePath)
				}
			}
		} else {
			if len(baseParams) > 0 {
				return fmt.Errorf("BasePath does not use variables defines in BaseParams")
			}
		}
	}
	if r.MediaType != nil {
		if err := r.MediaType.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Parent returns the parent resource if any.
func (r *ResourceDefinition) Parent() *ResourceDefinition {
	if r.computedParent != nil {
		return r.computedParent
	}
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

// Validate tests whether the action definition is consistent: parameters have unique names and it has at least
// one response.
func (a *ActionDefinition) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("Action name cannot be empty")
	}
	for i, r := range a.Responses {
		for j, r2 := range a.Responses {
			if i != j && r.Status == r2.Status {
				return fmt.Errorf("Multiple response definitions with status code %d", r.Status)
			}
		}
		if err := r.Validate(); err != nil {
			return fmt.Errorf("invalid %d response definition: %s", r.Status, err)
		}
	}
	if err := a.ValidateParams(); err != nil {
		return err
	}
	if err := a.Payload.Validate(); err != nil {
		return fmt.Errorf(`invalid payload definition for action "%s": %s`,
			a.Name, err)
	}
	return nil
}

// ContextName computes the name of the context data structure that corresponds to this action.
func (a *ActionDefinition) ContextName() string {
	return inflect.Camelize(a.Name) + inflect.Camelize(a.Resource.Name) + "Context"
}

// PayloadTypeName computes the name of the payload data structure that corresponds to this action.
func (a *ActionDefinition) PayloadTypeName() string {
	return inflect.Camelize(a.Name) + inflect.Camelize(a.Resource.Name) + "Payload"
}

// ValidateParams checks the action parameters (make sure they have names, members and types).
func (a *ActionDefinition) ValidateParams() error {
	if a.Params.Type == nil {
		return nil
	}
	params, ok := a.Params.Type.(Object)
	if !ok {
		return fmt.Errorf("invalid params type %s for action %s", a.Params.Type.GoType(), a.Name)
	}
	for n, p := range params {
		if n == "" {
			return fmt.Errorf("%s has parameter with no name", a.Name)
		} else if p == nil {
			return fmt.Errorf("definition of parameter %s of action %s cannot be nil",
				n, a.Name)
		} else if p.Type == nil {
			return fmt.Errorf("type of parameter %s of action %s cannot be nil",
				n, a.Name)
		}
		if p.Type.Kind() == ObjectType {
			return fmt.Errorf(`parameter %s of action %s cannot be an object, only action payloads may be of type object`,
				n, a.Name)
		}
		if err := p.Validate(); err != nil {
			return fmt.Errorf(`invalid definition for parameter %s of action %s: %s`,
				n, a.Name, err)
		}
	}
	return nil
}

// AttributeIterator is the type of iterator functions given to Iterate to iterate over each
// member of an object attribute.
type AttributeIterator func(name string, att *AttributeDefinition) error

// Validate tests whether the attribute definition is consistent: required fields exist.
func (a *AttributeDefinition) Validate() error {
	o, isObject := a.Type.(Object)
	for _, v := range a.Validations {
		if r, ok := v.(*RequiredValidationDefinition); ok {
			if !isObject {
				return fmt.Errorf("required fields validation defined on attribute of type %s", a.Type.GoType())
			}
			for _, n := range r.Names {
				var found bool
				for an := range o {
					if n == an {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf(`required field "%s" does not exist`, n)
				}
			}
		}
	}
	if isObject {
		for _, att := range o {
			if err := att.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Iterate runs the given function on each member of the attribute.
// If the iterator function returns an error then iterator stops and returns the error.
// Iterate panics if the type of a is not Object.
func (a *AttributeDefinition) Iterate(it AttributeIterator) error {
	o, ok := a.Type.(Object)
	if !ok {
		panic("iterated attribute not an object")
	}
	for n, at := range o {
		if err := it(n, at); err != nil {
			return err
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

// Struct returns the go code for a data structure that contains field which match the attribute
// type fields.
// It panics if the attribute type is not Object.
func (a *AttributeDefinition) Struct() string {
	var buffer bytes.Buffer
	buffer.WriteString("struct {\n")
	o := a.Type.(Object)
	keys := make([]string, len(o))
	i := 0
	for n := range o {
		keys[i] = n
		i++
	}
	sort.Strings(keys)
	for _, n := range keys {
		att := o[n]
		code := att.Type.GoType()
		switch t := att.Type.(type) {
		case *Array:
			code = t.Struct()
		case Object:
			code = att.Struct()
		}
		var omit string
		if !a.IsRequired(n) {
			omit = ",omitempty"
		}
		buffer.WriteString(fmt.Sprintf("\t%s %s `json:\"%s%s\"`\n",
			Goify(n, true), code, n, omit))
	}
	buffer.WriteString("}")
	return buffer.String()
}

// Validate checks that the media type definition is consistent: its identifier is a valid media
// type identifier.
func (m *MediaTypeDefinition) Validate() error {
	if m.Resource == nil && m.Name == "" {
		return fmt.Errorf("Media type must have a name")
	}
	if m.Identifier != "" {
		_, _, err := mime.ParseMediaType(m.Identifier)
		if err != nil {
			return fmt.Errorf("invalid media type identifier: %s", err)
		}
	}
	return nil
}

// TypeName computes the generated go structure type name for the media type.
func (m *MediaTypeDefinition) TypeName() string {
	if m.Resource != nil {
		return Goify(m.Resource.Name, true)
	}
	return Goify(m.Name, true)
}

// Validate checks that the response definition is consistent: its status is set and the media
// type definition if any is valid.
func (r *ResponseDefinition) Validate() error {
	if r.Status == 0 {
		return fmt.Errorf("response status not defined")
	}
	if r.MediaType != nil {
		if err := r.MediaType.Validate(); err != nil {
			return err
		}
	}
	return nil
}
