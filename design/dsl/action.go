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
		a.Routes = append(a.Routes, routes...)
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
func Headers(dsl func()) {
	if a, ok := actionDefinition(true); ok {
		headers := new(AttributeDefinition)
		if executeDSL(dsl, headers) {
			a.Headers = headers
		}
	}
}

// Params computes the action parameters from the given DSL.
func Params(dsl func()) {
	if a, ok := actionDefinition(true); ok {
		params := new(AttributeDefinition)
		if executeDSL(dsl, params) {
			a.Params = params
		}
	}
}

// Payload defines the action payload DSL.
func Payload(p interface{}) {
	if a, ok := actionDefinition(true); ok {
		var at *AttributeDefinition
		if dsl, ok := p.(func()); ok {
			executeDSL(dsl, at)
		} else {
			at, _ = p.(*AttributeDefinition)
		}
		rn := inflect.Camelize(a.Parent.Name)
		an := inflect.Camelize(a.Name)
		a.Payload = &UserTypeDefinition{
			AttributeDefinition: at,
			TypeName:            fmt.Sprintf("%s%sPayload", an, rn),
		}
	}
}

// Response records a possible action response.
func Response(name string, paramsAndDSL ...interface{}) {
	if a, ok := actionDefinition(true); ok {
		if a.Responses == nil {
			a.Responses = make(map[string]*ResponseDefinition)
		}
		if _, ok := a.Responses[name]; ok {
			RecordError(fmt.Errorf("response %s is defined twice", name))
			return
		}
		var params []string
		var dsl func()
		if len(paramsAndDSL) > 0 {
			d := paramsAndDSL[len(paramsAndDSL)-1]
			if dsl, ok = d.(func()); ok {
				paramsAndDSL = paramsAndDSL[:len(paramsAndDSL)-1]
			}
			params = make([]string, len(paramsAndDSL))
			for i, p := range paramsAndDSL {
				params[i], ok = p.(string)
				if !ok {
					RecordError(fmt.Errorf("invalid response template parameter %#v, must be a string", p))
					return
				}
			}
		}
		var resp *ResponseDefinition
		if len(params) > 0 {
			if tmpl, ok := Design.ResponseTemplates[name]; ok {
				resp = tmpl.Template(params...)
			} else {
				RecordError(fmt.Errorf("no response template named %#v", name))
				return
			}
		} else {
			if ar, ok := Design.Responses[name]; ok {
				resp = ar.Dup()
			} else {
				resp = &ResponseDefinition{Name: name}
			}
		}
		if (dsl != nil) && !executeDSL(dsl, resp) {
			return
		}
		a.Responses[name] = resp
	}
}
