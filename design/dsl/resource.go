package dsl

import . "github.com/raphael/goa/design"

// Resource defines a resource DSL.
//
// var _ = Resource("bottle", func() {
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
func Resource(name string, dsl func()) *ResourceDefinition {
	var resource *ResourceDefinition
	if a, ok := apiDefinition(true); ok {
		resource, ok = a.Resources[name]
		if !ok {
			resource = &ResourceDefinition{Name: name}
		}
		if ok := executeDSL(dsl, resource); ok {
			a.Resources[name] = resource
		}
	}
	return resource
}

// Parent defines the resource parent.
// The parent resource is used to compute the path to the resource actions.
func Parent(p string) {
	if r, ok := resourceDefinition(true); ok {
		r.ParentName = p
	}
}

// CanonicalAction sets the name of the action with canonical href.
func CanonicalAction(a string) {
	if r, ok := resourceDefinition(true); ok {
		r.CanonicalAction = a
	}
}
