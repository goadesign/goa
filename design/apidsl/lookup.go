package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"strings"
)

// Lookup defines a lookup template who is beeing called after the context is created
func Lookup(name string, userType design.DataStructure, dsl func()) {

	key := mapKey(name)

	if design.Design.Lookups == nil {
		design.Design.Lookups = make(map[string]*design.LookupDefinition)
	} else if _, ok := design.Design.Lookups[key]; ok {
		dslengine.ReportError("redefinition of lookup", name)
		return
	}

	_, ok := dslengine.CurrentDefinition().(*design.APIDefinition)
	if !ok {
		dslengine.IncompatibleDSL()
		return
	}

	var ut *design.UserTypeDefinition
	switch t := userType.(type) {
	case *design.UserTypeDefinition:
		ut = t
	case *design.MediaTypeDefinition:
		ut = t.UserTypeDefinition
	default:
		dslengine.ReportError("provided type must be a Type or a Media")
	}

	lookup := &design.LookupDefinition{
		Name:       key,
		ReturnType: ut,
	}
	dslengine.Execute(dsl, lookup)

	design.Design.Lookups[key] = lookup
}

// UseLookup can be used inside a Resource or Action to define wich lookup template to use
func UseLookup(name string, dsl ...func()) {

	var lookups *[]*design.LookupDefinition

	switch d := dslengine.CurrentDefinition().(type) {
	case *design.ActionDefinition:
		lookups = &d.Lookups

	case *design.ResourceDefinition:
		lookups = &d.Lookups

	default:
		dslengine.IncompatibleDSL()
		return
	}

	//if _, ok := lookups[key]; ok {
	//	dslengine.ReportError("duplicate uselookup with same name", name)
	//	return
	//}

	lookup := &design.LookupDefinition{
		Name: mapKey(name),
	}

	if len(dsl) >= 1 {
		dslengine.Execute(dsl[0], lookup)
	}

	*lookups = append(*lookups, lookup)
}

// MustResolve indicates that the lookup must return a non nil value
// otherwise the execution of the chain will stop and return a 404 not found
func MustResolve() {
	if l, ok := lookupDefinition(); ok {
		l.MustResolve = true
	}
}

// MapRouteParam reroutes a route param to the argument of the lookup
func MapRouteParam(from, to string) {
	if l, ok := lookupDefinition(); ok {
		if l.Remap == nil {
			l.Remap = make(map[string]string)
		}
		l.Remap[mapKey(to)] = mapKey(from)
	}
}

func mapKey(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}
