package dsl

import (
	"fmt"

	"bitbucket.org/pkg/inflect"
	. "github.com/raphael/goa/design"
)

// Action defines an action definition DSL.
//
// Action("Update", func() {
//     Description("Update account")
//     Routing(
//         PUT("/:id"),
//         PUT("/organizations/:org/accounts/:id"),
//     )
//     Params(func() {
//         Param("org", String)
//         Param("id", Integer)
//     })
//     Headers(func() {
//         Header("Authorization", String)
//         Header("X-Account", Integer)
//         Required("Authorization", "X-Account")
//     })
//     Payload(UpdatePayload)
//     Response(NoContent)
//     Response(NotFound)
// })
func Action(name string, dsl func()) {
	if r, ok := resourceDefinition(true); ok {
		if r.Actions == nil {
			r.Actions = make(map[string]*ActionDefinition)
		}
		action, ok := r.Actions[name]
		if !ok {
			action = &ActionDefinition{
				Parent: r,
				Name:   name,
			}
		}
		if !executeDSL(dsl, action) {
			return
		}
		r.Actions[name] = action
	}
}

// Routing adds one or more routes to the action
func Routing(routes ...*RouteDefinition) {
	if a, ok := actionDefinition(true); ok {
		for _, r := range routes {
			rwcs := ExtractWildcards(a.Parent.FullPath())
			wcs := ExtractWildcards(r.Path)
			for _, rwc := range rwcs {
				for _, wc := range wcs {
					if rwc == wc {
						ReportError(`duplicate wildcard "%s" in resource base path "%s" and action route "%s"`,
							wc, a.Parent.FullPath(), r.Path)
					}
				}
			}
			r.Parent = a
			a.Routes = append(a.Routes, r)
		}
	}
}

// GET creates a route using the GET HTTP method
func GET(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "GET", Path: path}
}

// HEAD creates a route using the HEAD HTTP method
func HEAD(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "HEAD", Path: path}
}

// POST creates a route using the POST HTTP method
func POST(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "POST", Path: path}
}

// PUT creates a route using the PUT HTTP method
func PUT(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "PUT", Path: path}
}

// DELETE creates a route using the DELETE HTTP method
func DELETE(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "DELETE", Path: path}
}

// TRACE creates a route using the TRACE HTTP method
func TRACE(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "TRACE", Path: path}
}

// CONNECT creates a route using the GET HTTP method
func CONNECT(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "CONNECT", Path: path}
}

// PATCH creates a route using the PATCH HTTP method
func PATCH(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "PATCH", Path: path}
}

// Headers computes the action headers from the given DSL.
// Headers is also used to set the headers on a response.
func Headers(dsl func()) {
	if a, ok := actionDefinition(false); ok {
		headers := newAttribute(a.Parent.MediaType)
		if executeDSL(dsl, headers) {
			a.Headers = headers
		}
	} else if r, ok := resourceDefinition(false); ok {
		headers := newAttribute(r.MediaType)
		if executeDSL(dsl, headers) {
			r.Headers = headers
		}
	} else if r, ok := responseDefinition(true); ok {
		if r.Headers != nil {
			ReportError("headers already defined")
			return
		}
		var mtid string
		if pa, ok := r.Parent.(*ResourceDefinition); ok {
			mtid = pa.MediaType
		} else if pa, ok := r.Parent.(*ActionDefinition); ok {
			mtid = pa.Parent.MediaType
		}
		h := newAttribute(mtid)
		if executeDSL(dsl, h) {
			r.Headers = h
		}
	}
}

// Params computes the action parameters from the given DSL.
func Params(dsl func()) {
	if a, ok := actionDefinition(false); ok {
		params := newAttribute(a.Parent.MediaType)
		if executeDSL(dsl, params) {
			a.Params = params
		}
	} else if r, ok := resourceDefinition(true); ok {
		params := newAttribute(r.MediaType)
		if executeDSL(dsl, params) {
			r.Params = params
		}
	}
}

// Payload defines the action payload DSL.
// This function can be called passing in a data structure, a DSL or both.
// Examples:
//
// Payload(BottlePayload)
//
// Payload(func() {
// 	Member("Name")
// })
//
// Payload(BottlePayload, func() {
// 	Required("Name")
// })
//
func Payload(p interface{}, dsls ...func()) {
	if len(dsls) > 1 {
		ReportError("too many arguments in Payload call")
		return
	}
	if a, ok := actionDefinition(true); ok {
		var att *AttributeDefinition
		var dsl func()
		switch actual := p.(type) {
		case func():
			dsl = actual
			att = newAttribute(a.Parent.MediaType)
			att.Type = Object{}
		case *AttributeDefinition:
			att = actual
		case DataStructure:
			att = actual.Definition()
		}
		if len(dsls) == 1 {
			if dsl != nil {
				ReportError("invalid arguments in Payload call, must be (type), (dsl) or (type, dsl)")
			}
			dsl = dsls[0]
		}
		if dsl != nil {
			executeDSL(dsl, att)
		}
		rn := inflect.Camelize(a.Parent.Name)
		an := inflect.Camelize(a.Name)
		a.Payload = &UserTypeDefinition{
			AttributeDefinition: att,
			TypeName:            fmt.Sprintf("%s%sPayload", an, rn),
		}
	}
}

// newAttribute creates a new attribute definition using the media type with the given identifier
// as base type.
func newAttribute(baseMT string) *AttributeDefinition {
	var base DataType
	if mt, ok := Design.MediaTypes[baseMT]; ok {
		base = mt.Type
	}
	return &AttributeDefinition{BaseType: base}
}
