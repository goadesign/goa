package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// NOTE: the following functions are in this file so that IncompatibleDSL can compute the stack
// depth correcly when looking up the name of the caller DSL function. IncompatibleDSL in used in
// two scenarios: in a type switch statement or via one of the functions below. In the case of a
// switch statement the name of the DSL function is 2 levels up the call to IncompatibleDSL while in
// the case of the functions below it's 3 levels up. Using a different file allows IncompatibleDSL
// to correctly compute the stack depth. A little bit dirty but seems to be the lesser evil.

// apiDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func apiDefinition() (*design.APIDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.APIDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return a, ok
}

// encodingDefinition returns true and current context if it is an EncodingDefinition,
// nil and false otherwise.
func encodingDefinition() (*design.EncodingDefinition, bool) {
	e, ok := dslengine.CurrentDefinition().(*design.EncodingDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return e, ok
}

// contactDefinition returns true and current context if it is an ContactDefinition,
// nil and false otherwise.
func contactDefinition() (*design.ContactDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.ContactDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return a, ok
}

// licenseDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func licenseDefinition() (*design.LicenseDefinition, bool) {
	l, ok := dslengine.CurrentDefinition().(*design.LicenseDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return l, ok
}

// docsDefinition returns true and current context if it is a DocsDefinition,
// nil and false otherwise.
func docsDefinition() (*design.DocsDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.DocsDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return a, ok
}

// mediaTypeDefinition returns true and current context if it is a MediaTypeDefinition,
// nil and false otherwise.
func mediaTypeDefinition() (*design.MediaTypeDefinition, bool) {
	m, ok := dslengine.CurrentDefinition().(*design.MediaTypeDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return m, ok
}

// typeDefinition returns true and current context if it is a UserTypeDefinition,
// nil and false otherwise.
func typeDefinition() (*design.UserTypeDefinition, bool) {
	m, ok := dslengine.CurrentDefinition().(*design.UserTypeDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return m, ok
}

// attributeDefinition returns true and current context if it is an Attribute,
// nil and false otherwise.
func attributeDefinition() (*design.AttributeDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.AttributeDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return a, ok
}

// resourceDefinition returns true and current context if it is a ResourceDefinition,
// nil and false otherwise.
func resourceDefinition() (*design.ResourceDefinition, bool) {
	r, ok := dslengine.CurrentDefinition().(*design.ResourceDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return r, ok
}

// corsDefinition returns true and current context if it is a CORSDefinition, nil And
// false otherwise.
func corsDefinition() (*design.CORSDefinition, bool) {
	cors, ok := dslengine.CurrentDefinition().(*design.CORSDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return cors, ok
}

// actionDefinition returns true and current context if it is an ActionDefinition,
// nil and false otherwise.
func actionDefinition() (*design.ActionDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.ActionDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return a, ok
}

// responseDefinition returns true and current context if it is a ResponseDefinition,
// nil and false otherwise.
func responseDefinition() (*design.ResponseDefinition, bool) {
	r, ok := dslengine.CurrentDefinition().(*design.ResponseDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
	}
	return r, ok
}
