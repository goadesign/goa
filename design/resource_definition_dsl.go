package design

import "fmt"

// Resource defines a resource
func Resource(name string, dsl func()) error {
	if a, ok := apiDefinition(); ok {
		if _, ok := a.Resources[name]; ok {
			appendError(fmt.Errorf("multiple definitions for resource %s", name))
			return nil
		}
		resource := &ResourceDefinition{Name: name}
		if ok := executeDSL(dsl, resource); ok {
			a.Resources[name] = resource
		}
	}
	return nil // to allow for 'var _ = ' trick
}

// MediaType sets the resource media type
func MediaType(val *MediaTypeDefinition) {
	if r, ok := resourceDefinition(); ok {
		r.MediaType = val
	}
}

// Prefix sets the resource path prefix
func Prefix(p string) {
	if r, ok := resourceDefinition(); ok {
		r.Prefix = p
	}
}

// Prefix sets the resource path prefix
func CanonicalAction(a string) {
	if r, ok := resourceDefinition(); ok {
		r.CanonicalAction = a
	}
}
