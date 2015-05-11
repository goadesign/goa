package design

import "fmt"

// Action defines an action definition DSL
func Action(name string, dsl func()) {
	if r, ok := resourceDefinition(); ok {
		action := &ActionDefinition{Name: name}
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

// Get creates a route using the GET HTTP method
func Get(path string) *Route {
	return &Route{Verb: "GET", Path: path}
}

// Head creates a route using the HEAD HTTP method
func Head(path string) *Route {
	return &Route{Verb: "HEAD", Path: path}
}

// Post creates a route using the POST HTTP method
func Post(path string) *Route {
	return &Route{Verb: "POST", Path: path}
}

// Put creates a route using the PUT HTTP method
func Put(path string) *Route {
	return &Route{Verb: "PUT", Path: path}
}

// Delete creates a route using the DELETE HTTP method
func Delete(path string) *Route {
	return &Route{Verb: "DELETE", Path: path}
}

// Trace creates a route using the TRACE HTTP method
func Trace(path string) *Route {
	return &Route{Verb: "TRACE", Path: path}
}

// Connect creates a route using the GEt HTTP method
func Connect(path string) *Route {
	return &Route{Verb: "CONNECT", Path: path}
}

// Patch creates a route using the PATCH HTTP method
func Patch(path string) *Route {
	return &Route{Verb: "PATCH", Path: path}
}

// Payload sets the action params attributes.
func Params(attributes ActionParams) {
	if a, ok := actionDefinition(); ok {
		a.QueryParams = attributes
	}
}

// Payload sets the action payload attributes.
func Payload(attribute *AttributeDefinition) {
	if a, ok := actionDefinition(); ok {
		a.Payload = attribute
	}
}

// Response records the response and template parameters.
func Response(status int, params ...interface{}) {
	if a, ok := actionDefinition(); ok {
		if len(params) > 2 {
			appendError(fmt.Errorf("too many arguments in call to Response"))
			return
		}
		var mediaType *MediaTypeDefinition
		if len(params) >= 1 {
			var ok bool
			if mediaType, ok = params[0].(*MediaTypeDefinition); !ok {
				invalidArgError("MediaTypeDefinition", params[0])
				return
			}
		}
		response := ResponseDefinition{Status: status, MediaType: mediaType}
		a.Responses = append(a.Responses, &response)
	}
}
