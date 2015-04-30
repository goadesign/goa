package design

// Action defines an action definition DSL
func Action(name string, dsl func()) {
	if r, ok := resourceDefinition(); ok {
		action := &ActionDefinition{Name: name}
		err := executeDSL(dsl, action)
		if err != nil {
			return
		}
		r.Actions = append(r.Actions, action)
	}
}

// Routing adds one or more routes to the action
func Routing(routes ...Route) {
	if a, ok := actionDefinition; ok {
		a.Routes = append(a.Routes, routes...)
	}
}

func Get(path string) *Route {
	return &Route{Verb: "GET", Path: path}
}

func Head(path string) *Route {
	return &Route{Verb: "HEAD", Path: path}
}

func Post(path string) *Route {
	return &Route{Verb: "POST", Path: path}
}

func Put(path string) *Route {
	return &Route{Verb: "PUT", Path: path}
}

func Delete(path string) *Route {
	return &Route{Verb: "DELETE", Path: path}
}

func Trace(path string) *Route {
	return &Route{Verb: "TRACE", Path: path}
}

func Connect(path string) *Route {
	return &Route{Verb: "CONNECT", Path: path}
}

func Patch(path string) *Route {
	return &Route{Verb: "PATCH", Path: path}
}

// Payload sets the action params attributes.
func Params(attributes ...Attribute) {
	if a, ok := actionDefinition(); ok {
		a.Params = append(a.Params, routes...)
	}
}

// Payload sets the action payload attributes.
func Payload(attributes ...Attribute) {
	if a, ok := actionDefinition(); ok {
		a.Payload = append(a.Payload, attributes...)
	}
}

// Response records the response and template parameters.
func Response(name string, params ...interface{}) {
	if a, ok := actionDefinition(); ok {
		a.responseParams[name] = params
	}
}

// Routing appends the given routes to the action routes.
func Routing(routes ...Routes) {
	if a, ok := actionDefinition(); ok {
		a.Routes = append(a.Routes, routes)
	}
}
