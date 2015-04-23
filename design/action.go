package design

import (
	"fmt"
	"regexp"
	"sort"
)

// A resource action
// Defines an HTTP endpoint and the shape of HTTP requests and responses made to
// that endpoint.
// The shape of requests is defined via "parameters", there are path parameters
// (i.e. portions of the URL that define parameter values), query string
// parameters and a payload parameter (request body).
type Action struct {
	Name        string       // Action name, e.g. "create"
	Description string       // Action description, e.g. "Creates a task"
	HttpMethod  string       // HTTP method, e.g. "POST"
	Path        string       // HTTP URL suffix (appended to parent resource path)
	Responses   []*Response  // Set of possible response definitions
	PathParams  ActionParams // Path parameters if any
	QueryParams ActionParams // Query string parameters if any
	Payload     *Member      // Payload blueprint (request body) if any
}

// Get initializes the action HTTP method to GET and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Get(path string) *Action {
	return a.method("Get", path)
}

// Post initializes the action HTTP method to POST and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Post(path string) *Action {
	return a.method("Post", path)
}

// Put initializes the action HTTP method to PUT and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Put(path string) *Action {
	return a.method("Put", path)
}

// Patch initializes the action HTTP method to PATCH and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Patch(path string) *Action {
	return a.method("Patch", path)
}

// Delete initializes the action HTTP method to DELETE and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Delete(path string) *Action {
	return a.method("Delete", path)
}

// WithParam creates a new query string parameter and returns it.
// Type is inherited from the resource media type member with the same name.
// If the resource media type does not define a member with the param name then the type must be
// set explicitly (with e.g. 'WithParam("foo").Integer()').
func (a *Action) WithParam(name string) *ActionParam {
	m := Member{Type: String}
	param := &ActionParam{Name: name, Member: &m}
	a.QueryParams[name] = param
	return param
}

// WithPayload sets the request payload type.
// Note: Object members may be nil in which case the definition for the member with the same name
// in the resource media type is used to load and validate request bodies.
func (a *Action) WithPayload(payload *Member) *Action {
	a.Payload = payload
	return a
}

// Respond adds a new action response using the given media type and a
// status code of 200.
func (a *Action) Respond(media *MediaType) *Response {
	r := Response{Status: 200, MediaType: media}
	a.Responses = append(a.Responses, &r)
	return &r
}

// RespondNoContent adds a new action response with no media type and a status
// code of 204.
func (a *Action) RespondNoContent() *Response {
	r := Response{Status: 204}
	a.Responses = append(a.Responses, &r)
	return &r
}

// Query parameter names sorted alphavetically
func (a *Action) QueryParamNames() []string {
	return sortedKeys(a.QueryParams)
}

// Path parameter names sorted alphavetically
func (a *Action) PathParamNames() []string {
	return sortedKeys(a.PathParams)
}

// Sort map keys alphabetically
func sortedKeys(params ActionParams) []string {
	names := make([]string, len(params))
	i := 0
	for n, _ := range params {
		names[i] = n
		i += 1
	}
	sort.Strings(names)
	return names
}

// Regular expression used to capture path parameters
var pathRegex = regexp.MustCompile("/:([^/]+)")

// Internal helper method that sets HTTP method, path and path params
func (a *Action) method(method, path string) *Action {
	a.HttpMethod = method
	a.Path = path
	var matches = pathRegex.FindAllStringSubmatch(path, -1)
	a.PathParams = make(map[string]*ActionParam, len(matches))
	for _, m := range matches {
		mem := Member{Type: String}
		a.PathParams[m[1]] = &ActionParam{Name: m[1], Member: &mem}
	}
	return a
}

// Validates that action definition is consistent: parameters have unique names, has at least one
// response.
func (a *Action) validate() error {
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
	for _, p := range a.PathParams {
		for _, q := range a.QueryParams {
			if p.Name == q.Name {
				return fmt.Errorf("Action has both path parameter and query parameter named %s",
					p.Name)
			}
		}
	}
	if err := a.validateParams(true); err != nil {
		return err
	}
	if err := a.validateParams(false); err != nil {
		return err
	}
	return nil
}

// Validate action parameters (make sure they have names, members and types)
func (a *Action) validateParams(isPath bool) error {
	var params ActionParams
	if isPath {
		params = a.PathParams
	} else {
		params = a.QueryParams
	}
	for n, p := range params {
		if n == "" {
			return fmt.Errorf("%s has parameter with no name", a.Name)
		} else if p.Member == nil {
			return fmt.Errorf("Member field of %s parameter :%s cannot be nil",
				a.Name, n)
		} else if p.Member.Type == nil {
			return fmt.Errorf("type of %s parameter :%s cannot be nil",
				a.Name, n)
		}
	}
	return nil
}
