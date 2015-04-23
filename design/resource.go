package design

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"bitbucket.org/pkg/inflect"
)

// A REST resource
// Defines a media type and a set of actions that can be executed through HTTP requests.
// A resource is versioned so that multiple versions of the same resource may
// be exposed by the API.
type Resource struct {
	Name        string             // Resource name
	BasePath    string             // Common URL prefix to all resource action HTTP requests
	Description string             // Optional description
	Version     string             // Optional version
	MediaType   *MediaType         // Default media type, describes the resource attributes
	Actions     map[string]*Action // Exposed resource actions indexed by name
}

// Action adds or retrieves an action with the given name.
// Use returned value to define description, http method, path, parameters and responses.
func (r *Resource) Action(name string) *Action {
	if action, ok := r.Actions[name]; ok {
		return action
	}
	a := Action{
		Name:        name,
		PathParams:  make(ActionParams),
		QueryParams: make(ActionParams),
	}
	r.Actions[name] = &a
	return &a
}

// Index is an helper method that defines a REST "index" action.
// Sets HTTP method to GET and initializes path with resource base path.
// Also defines default response with status 200 and media type that is a
// collection of media types defined by resource.
func (r *Resource) Index(p string) *Action {
	a := r.Action("Index").Get(path.Join(r.BasePath, p))
	a.Description = fmt.Sprintf("List all %s.", strings.ToLower(inflect.Pluralize(r.Name)))
	a.Respond(CollectionOf(r.MediaType))
	return a
}

// Show is an helper method that defines a REST "show" action.
// Sets HTTP method to GET and initializes path resource base path appended with
// ":id" path parameter.
// Also defines default response with status 200 and same media type as resource.
func (r *Resource) Show(p string) *Action {
	a := r.Action("Show").Get(path.Join(r.BasePath, p))
	a.Description = fmt.Sprintf("Retrieve %s.",
		strings.ToLower(inflect.Singularize(r.Name)))
	a.Respond(r.MediaType)
	return a
}

// Create is an helper method that defines a REST "create" action.
// Sets HTTP method to POST and initializes path with resource base path.
// Also defines default response with status 201 and "Location" header.
func (r *Resource) Create(p string) *Action {
	a := r.Action("Create").Post(path.Join(r.BasePath, p))
	a.Description = fmt.Sprintf("Create new %s.", strings.ToLower(inflect.Singularize(r.Name)))
	loc := regexp.MustCompile(fmt.Sprintf("^%s", regexp.QuoteMeta(r.BasePath)))
	a.RespondNoContent().WithLocation(loc)
	return a
}

// Update is an helper method that defines a REST "update" action.
// Sets HTTP method to PUT and initializes path with resource base path appended
// with ":id" path parameter.
// Also defines default response with status 204 and no media type.
func (r *Resource) Update(p string) *Action {
	a := r.Action("Update").Put(path.Join(r.BasePath, p))
	a.Description = fmt.Sprintf("Replace content of %s.",
		strings.ToLower(inflect.Singularize(r.Name)))
	a.RespondNoContent()
	return a
}

// Path is an helper method that creates a REST "patch" action.
// Sets HTTP method to PATCH and initializes path with resource base path
// appended with ":id" path parameter.
// Also defines default response with status 204 and no media type.
func (r *Resource) Patch(p string) *Action {
	a := r.Action("Patch").Patch(path.Join(r.BasePath, p))
	a.Description = fmt.Sprintf("Update given fields of %s.",
		strings.ToLower(inflect.Singularize(r.Name)))
	a.RespondNoContent()
	return a
}

// Delete is an helper method that creates a REST "delete" action.
// Sets HTTP method to DELETE and initializes path with resource base path appended
// with ":id" path parameter.
// Also defines default response with status 204 and no media type.
func (r *Resource) Delete(p string) *Action {
	a := r.Action("Delete").Delete(path.Join(r.BasePath, p))
	a.Description = fmt.Sprintf("Delete %s.",
		strings.ToLower(inflect.Singularize(r.Name)))
	a.RespondNoContent()
	return a
}

// Validates that resource definition is consistent: action names are valid and each action is
// valid.
func (r *Resource) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("Resource name cannot be empty")
	}
	for _, a := range r.Actions {
		if err := a.validate(); err != nil {
			return err
		}
	}
	return nil
}
