package design

import "fmt"

// A REST resource
// Defines a media type and a set of actions that can be executed through HTTP requests.
// A resource is versioned so that multiple versions of the same resource may
// be exposed by the API.
type ResourceDefinition struct {
	Name        string             // Resource name
	BasePath    string             // Common URL prefix to all resource action HTTP requests
	Description string             // Optional description
	Version     string             // Optional version
	MediaType   *MediaType         // Default media type, describes the resource attributes
	Actions     map[string]*Action // Exposed resource actions indexed by name
}

// Validates that resource definition is consistent: action names are valid and each action is
// valid.
func (r *ResourceDefinition) Validate() error {
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
