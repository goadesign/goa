package dsl

import "github.com/raphael/goa/design"

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
//     Params(func() {
//         Param("id", Integer, "Account ID")
//         Required("id")
//     })
//     Payload(func() {
//         Member("name")
//         Member("year")
//     })
//     Responses(
//         NoContent(),
//         NotFound(),
//     )
// })
func Action(name string, dsl func()) {
	if r, ok := resourceDefinition(); ok {
		action := &design.ActionDefinition{Name: name}
		if !executeDSL(dsl, action) {
			return
		}
		r.Actions[name] = action
	}
}

// Routing adds one or more routes to the action
func Routing(routes ...*Route) {
	if a, ok := actionDefinition(); ok {
		a.Routes = append(a.Routes, routes...)
	}
}

// GET creates a route using the GET HTTP method
func GET(path string) *Route {
	return &Route{Verb: "GET", Path: path}
}

// HEAD creates a route using the HEAD HTTP method
func HEAD(path string) *Route {
	return &Route{Verb: "HEAD", Path: path}
}

// POST creates a route using the POST HTTP method
func POST(path string) *Route {
	return &Route{Verb: "POST", Path: path}
}

// PUT creates a route using the PUT HTTP method
func PUT(path string) *Route {
	return &Route{Verb: "PUT", Path: path}
}

// DELETE creates a route using the DELETE HTTP method
func DELETE(path string) *Route {
	return &Route{Verb: "DELETE", Path: path}
}

// TRACE creates a route using the TRACE HTTP method
func TRACE(path string) *Route {
	return &Route{Verb: "TRACE", Path: path}
}

// CONNECT creates a route using the GET HTTP method
func CONNECT(path string) *Route {
	return &Route{Verb: "CONNECT", Path: path}
}

// PATCH creates a route using the PATCH HTTP method
func PATCH(path string) *Route {
	return &Route{Verb: "PATCH", Path: path}
}

// Headers computes the action headers from the given DSL.
func Headers(dsl func()) {
	if a, ok := actionDefinition(); ok {
		headers := new(AttributeDefinition)
		if executeDSL(dsl, headers) {
			a.Headers = headers
		}
	}
}

// Params computes the action parameters from the given DSL.
func Params(dsl func()) {
	if a, ok := actionDefinition(); ok {
		params := new(AttributeDefinition)
		if executeDSL(dsl, params) {
			a.Params = params
		}
	}
}

// Payload sets the action payload attributes.
func Payload(dsl func()) {
	if a, ok := actionDefinition(); ok {
		payload := new(AttributeDefinition)
		if executeDSL(dsl, payload) {
			a.Payload = payload
		}
	}
}

func Responses(resps ...*ResponseDefinition) {
	if a, ok := actionDefinition(); ok {
		a.Responses = resps
	}
}

//// Response records the response and template parameters.
//func Response(status int, params ...interface{}) {
//if a, ok := actionDefinition(); ok {
//if len(params) > 2 {
//appendError(fmt.Errorf("too many arguments in call to Response"))
//return
//}
//var mediaType *MediaTypeDefinition
//if len(params) >= 1 {
//var ok bool
//if mediaType, ok = params[0].(*MediaTypeDefinition); !ok {
//invalidArgError("MediaTypeDefinition", params[0])
//return
//}
//}
//response := ResponseDefinition{Status: status, MediaType: mediaType}
//a.Responses = append(a.Responses, &response)
//}
//}

//// Regular expression used to capture path parameters
//var pathRegex = regexp.MustCompile("/:([^/]+)")

//// Internal helper method that sets HTTP method, path and path params
//func (a *ActionDefinition) method(method, path string) *ActionDefinition {
//r := Route{Verb: method, Path: path}
//a.Routes = append(a.Routes, &r)
//var matches = pathRegex.FindAllStringSubmatch(path, -1)
//a.PathParams = make(map[string]*ActionParam, len(matches))
//for _, m := range matches {
//mem := AttributeDefinition{Type: String}
//a.PathParams[m[1]] = &ActionParam{Name: m[1], Member: &mem}
//}
//return a
//}
