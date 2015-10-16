package dsl

import . "github.com/raphael/goa/design"

// Resource defines a resource DSL.
//
// var _ = Resource("bottle", func() {
//      Description("A wine bottle") // Resource description
// 	DefaultMedia(BottleMediaType)   // Resource actions default media type
// 	BasePath("/bottles")         // Resource actions path prefix if not ""
//      Parent("account")            // Name of parent resource if any
// 	CanonicalActionName("show")      // Action that returns canonical representation
// 	Trait("Authenticated")       // Included trait if any, can appear more than once
// 	Action("show", func() {      // Action definition, can appear more than once
//        // ... Action DSL
// 	})
// })
func Resource(name string, dsl func()) *ResourceDefinition {
	if Design == nil {
		InitDesign()
	}
	if Design.Resources == nil {
		Design.Resources = make(map[string]*ResourceDefinition)
	}
	var resource *ResourceDefinition
	if topLevelDefinition(true) {
		if _, ok := Design.Resources[name]; ok {
			ReportError("resource %#v is defined twice", name)
			return nil
		}
		resource = NewResourceDefinition(name, dsl)
		Design.Resources[name] = resource
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

// CanonicalActionName sets the name of the action with canonical href.
func CanonicalActionName(a string) {
	if r, ok := resourceDefinition(true); ok {
		r.CanonicalActionName = a
	}
}
