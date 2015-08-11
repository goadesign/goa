package dsl

import . "github.com/raphael/goa/design"

// Resource defines a resource DSL.
//
// Resource("bottle", func() {
//      Description("A wine bottle") // Resource description
// 	MediaType(BottleMediaType)   // Resource actions default media type
// 	Prefix("/bottles")           // Resource actions path prefix if not ""
//      Parent("account")            // Name of parent resource if any
// 	CanonicalAction("show")      // Action that returns canonical representation
// 	Trait("Authenticated")       // Included trait if any, can appear more than once
// 	Action("show", func() {      // Action definition, can appear more than once
//        // ... Action DSL
// 	})
// })
func Resource(name string, dsl func()) {
	if a, ok := apiDefinition(); ok {
		resource, ok := a.Resources[name]
		if !ok {
			resource = &ResourceDefinition{Name: name}
		}
		if ok := executeDSL(dsl, resource); ok {
			a.Resources[name] = resource
		}
	}
}

// MediaType sets the resource media type
func MediaType(m *MediaTypeDefinition) {
	if r, ok := ctxStack.current().(*ResourceDefinition); ok {
		r.MediaType = m
		m.Resource = r
	} else if r, ok := responseDefinition(); ok {
		r.MediaType = m
	}
}

// Parent defines the resource parent.
// The parent resource is used to compute the path to the resource actions.
func Parent(p string) {
	if r, ok := resourceDefinition(); ok {
		r.Parent = p
	}
}

// Prefix sets the resource path prefix
func Prefix(p string) {
	if r, ok := resourceDefinition(); ok {
		r.Prefix = p
	}
}

// CanonicalAction sets the name of the action with canonical href.
func CanonicalAction(a string) {
	if r, ok := resourceDefinition(); ok {
		r.CanonicalAction = a
	}
}
