package design

import "fmt"

var (
	// Design is the API definition created via DSL.
	Design *APIDefinition
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
		Prefix          string                       // Common URL prefix to all resource action HTTP requests
		Description     string                       // Optional description
		Version         string                       // Optional version
		MediaType       *MediaTypeDefinition         // Default media type, describes the resource attributes
		Actions         map[string]*ActionDefinition // Exposed resource actions indexed by name
		CanonicalAction string                       // Action with canonical resource path

	}

	// MediaTypeDefinition describes the rendering of a resource using property and link
	// definitions. A property corresponds to a single member of the media type,
	// it has a name and a type as well as optional validation rules. A link has a
	// name and a URL that points to a related resource.
	// Media types also define views which describe which members and links to render when
	// building the response body for the corresponding view.
	MediaTypeDefinition struct {
		Object                                  // A media type definition is a JSON object
		Identifier   string                     // RFC 6838 Media type identifier
		Description  string                     // Optional description
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

	// ActionParam defines an action parameter (path element, query string or payload).
	ActionParam struct {
		Name   string               // Name of parameter
		Member *AttributeDefinition // Type and validations (if any)
	}

	// ActionParams defines a map of action parameters indexed by name.
	ActionParams map[string]*ActionParam

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
)

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
			return err
		}
	}
	if r.CanonicalAction != "" && !found {
		return fmt.Errorf("Unknown canonical action '%s'", r.CanonicalAction)
	}
	return nil
}

// Validate tests whether the action definition is consistent: parameters have unique names and it has at least
//  one response.
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
	}
	if err := a.ValidateParams(); err != nil {
		return err
	}
	return nil
}

// ValidateParams checks the action parameters (make sure they have names, members and types).
func (a *ActionDefinition) ValidateParams() error {
	if a.Params.Type == nil {
		return nil
	}
	params, ok := a.Params.Type.(Object)
	if !ok {
		return fmt.Errorf("invalid params type %s for action %s", a.Params.Type.Name(), a.Name)
	}
	for n, p := range params {
		if n == "" {
			return fmt.Errorf("%s has parameter with no name", a.Name)
		} else if p == nil {
			return fmt.Errorf("Member field of %s parameter :%s cannot be nil",
				a.Name, n)
		} else if p.Type == nil {
			return fmt.Errorf("type of %s parameter :%s cannot be nil",
				a.Name, n)
		}
	}
	return nil
}
