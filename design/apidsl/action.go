package apidsl

import (
	"fmt"
	"unicode"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Files defines an API endpoint that serves static assets. The logic for what to do when the
// filename points to a file vs. a directory is the same as the standard http package ServeFile
// function. The path may end with a wildcard that matches the rest of the URL (e.g. *filepath). If
// it does the matching path is appended to filename to form the full file path, so:
//
// 	Files("/index.html", "/www/data/index.html")
//
// Returns the content of the file "/www/data/index.html" when requests are sent to "/index.html"
// and:
//
//	Files("/assets/*filepath", "/www/data/assets")
//
// returns the content of the file "/www/data/assets/x/y/z" when requests are sent to
// "/assets/x/y/z".
// The file path may be specified as a relative path to the current path of the process.
func Files(path, filename string, dsls ...func()) {
	if r, ok := resourceDefinition(); ok {
		server := &design.FileServerDefinition{
			Parent:      r,
			RequestPath: path,
			FilePath:    filename,
		}
		if len(dsls) > 0 {
			if !dslengine.Execute(dsls[0], server) {
				return
			}
		}
		r.FileServers = append(r.FileServers, server)
	}
}

// Action implements the action definition DSL. Action definitions describe specific API endpoints
// including the URL, HTTP method and request parameters (via path wildcards or query strings) and
// payload (data structure describing the request HTTP body). An action belongs to a resource and
// "inherits" default values from the resource definition including the URL path prefix, default
// response media type and default payload attribute properties (inherited from the attribute with
// identical name in the resource default media type). Action definitions also describe all the
// possible responses including the HTTP status, headers and body. Here is an example showing all
// the possible sub-definitions:
//	Action("Update", func() {
//		Description("Update account")
//		Docs(func() {
//			Description("Update docs")
//			URL("http//cellarapi.com/docs/actions/update")
//		})
//		Scheme("http")
//		Routing(
//			PUT("/:id"),				// Full action path is built by appending "/:id" to parent resource base path
//			PUT("//orgs/:org/accounts/:id"),	// The // prefix indicates an absolute path
//		)
//		Params(func() {					// Params describe the action parameters
//			Param("org", String)			// Parameters may correspond to path wildcards
//			Param("id", Integer)
//			Param("sort", func() {			// or URL query string values.
//				Enum("asc", "desc")
//			})
//		})
//		Headers(func() {				// Headers describe relevant action headers
//			Header("Authorization", String)
//			Header("X-Account", Integer)
//			Required("Authorization", "X-Account")
//		})
//		Payload(UpdatePayload)				// Payload describes the HTTP request body (here using a type)
//		OptionalPayload(UpdatePayload)			// You can use OptionalPayload instead of Payload
//		Response(NoContent)				// Each possible HTTP response is described via Response
//		Response(NotFound)
//	})
func Action(name string, dsl func()) {
	if r, ok := resourceDefinition(); ok {
		if r.Actions == nil {
			r.Actions = make(map[string]*design.ActionDefinition)
		}
		action, ok := r.Actions[name]
		if !ok {
			action = &design.ActionDefinition{
				Parent: r,
				Name:   name,
			}
		}
		if !dslengine.Execute(dsl, action) {
			return
		}
		r.Actions[name] = action
	}
}

// Routing lists the action route. Each route is defined with a function named after the HTTP method.
// The route function takes the path as argument. Route paths may use wildcards as described in the
// [httptreemux](https://godoc.org/github.com/dimfeld/httptreemux) package documentation. These
// wildcards define parameters using the `:name` or `*name` syntax where `:name` matches a path
// segment and `*name` is a catch-all that matches the path until the end.
func Routing(routes ...*design.RouteDefinition) {
	if a, ok := actionDefinition(); ok {
		for _, r := range routes {
			r.Parent = a
			a.Routes = append(a.Routes, r)
		}
	}
}

// GET creates a route using the GET HTTP method.
func GET(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "GET", Path: path}
}

// HEAD creates a route using the HEAD HTTP method.
func HEAD(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "HEAD", Path: path}
}

// POST creates a route using the POST HTTP method.
func POST(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "POST", Path: path}
}

// PUT creates a route using the PUT HTTP method.
func PUT(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "PUT", Path: path}
}

// DELETE creates a route using the DELETE HTTP method.
func DELETE(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "DELETE", Path: path}
}

// OPTIONS creates a route using the OPTIONS HTTP method.
func OPTIONS(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "OPTIONS", Path: path}
}

// TRACE creates a route using the TRACE HTTP method.
func TRACE(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "TRACE", Path: path}
}

// CONNECT creates a route using the GET HTTP method.
func CONNECT(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "CONNECT", Path: path}
}

// PATCH creates a route using the PATCH HTTP method.
func PATCH(path string) *design.RouteDefinition {
	return &design.RouteDefinition{Verb: "PATCH", Path: path}
}

// Headers implements the DSL for describing HTTP headers. The DSL syntax is identical to the one
// of Attribute. Here is an example defining a couple of headers with validations:
//
//	Headers(func() {
//		Header("Authorization")
//		Header("X-Account", Integer, func() {
//			Minimum(1)
//		})
//		Required("Authorization")
//	})
//
// Headers can be used inside Action to define the action request headers, Response to define the
// response headers or Resource to define common request headers to all the resource actions.
func Headers(params ...interface{}) {
	if len(params) == 0 {
		dslengine.ReportError("missing parameter")
		return
	}
	dsl, ok := params[0].(func())
	if ok {
		switch def := dslengine.CurrentDefinition().(type) {
		case *design.ActionDefinition:
			headers := newAttribute(def.Parent.MediaType)
			if dslengine.Execute(dsl, headers) {
				def.Headers = headers
			}

		case *design.ResourceDefinition:
			headers := newAttribute(def.MediaType)
			if dslengine.Execute(dsl, headers) {
				def.Headers = headers
			}

		case *design.ResponseDefinition:
			if def.Headers != nil {
				dslengine.ReportError("headers already defined")
				return
			}
			var h *design.AttributeDefinition
			switch actual := def.Parent.(type) {
			case *design.ResourceDefinition:
				h = newAttribute(actual.MediaType)
			case *design.ActionDefinition:
				h = newAttribute(actual.Parent.MediaType)
			case nil: // API ResponseTemplate
				h = &design.AttributeDefinition{}
			default:
				dslengine.ReportError("invalid use of Response or ResponseTemplate")
			}
			if dslengine.Execute(dsl, h) {
				def.Headers = h
			}

		default:
			dslengine.IncompatibleDSL()
		}
	} else if cors, ok := corsDefinition(); ok {
		vals := make([]string, len(params))
		for i, p := range params {
			if v, ok := p.(string); ok {
				vals[i] = v
			} else {
				dslengine.ReportError("invalid parameter at position %d: must be a string", i)
				return
			}
		}
		cors.Headers = vals
	} else {
		dslengine.IncompatibleDSL()
	}
}

// Params describe the action parameters, either path parameters identified via wildcards or query
// string parameters. Each parameter is described via the `Param` function which uses the same DSL
// as the Attribute DSL. Here is an example:
//
//	Params(func() {
//		Param("id", Integer)		// A path parameter defined using e.g. GET("/:id")
//		Param("sort", String, func() {	// A query string parameter
//			Enum("asc", "desc")
//		})
//	})
//
// Params can be used inside Action to define the action parameters or Resource to define common
// parameters to all the resource actions.
func Params(dsl func()) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.ActionDefinition:
		params := newAttribute(def.Parent.MediaType)
		params.Type = make(design.Object)
		if dslengine.Execute(dsl, params) {
			def.Params = params
		}

	case *design.ResourceDefinition:
		params := newAttribute(def.MediaType)
		params.Type = make(design.Object)
		if dslengine.Execute(dsl, params) {
			def.Params = params
		}

	default:
		dslengine.IncompatibleDSL()
	}
}

// Payload implements the action payload DSL. An action payload describes the HTTP request body
// data structure. The function accepts either a type or a DSL that describes the payload members
// using the Member DSL which accepts the same syntax as the Attribute DSL. This function can be
// called passing in a type, a DSL or both. Examples:
//
//	Payload(BottlePayload)		// Request payload is described by the BottlePayload type
//
//	Payload(func() {		// Request payload is an object and is described inline
//		Member("Name")
//	})
//
//	Payload(BottlePayload, func() {	// Request payload is described by merging the inline
//		Required("Name")	// definition into the BottlePayload type.
//	})
//
func Payload(p interface{}, dsls ...func()) {
	payload(false, p, dsls...)
}

// OptionalPayload implements the action optional payload DSL. The function works identically to the
// Payload DSL except it sets a bit in the action definition to denote that the payload is not
// required. Example:
//
//	OptionalPayload(BottlePayload)		// Request payload is described by the BottlePayload type and is optional
//
func OptionalPayload(p interface{}, dsls ...func()) {
	payload(true, p, dsls...)
}

func payload(isOptional bool, p interface{}, dsls ...func()) {
	if len(dsls) > 1 {
		dslengine.ReportError("too many arguments given to Payload")
		return
	}
	if a, ok := actionDefinition(); ok {
		var att *design.AttributeDefinition
		var dsl func()
		switch actual := p.(type) {
		case func():
			dsl = actual
			att = newAttribute(a.Parent.MediaType)
			att.Type = design.Object{}
		case *design.AttributeDefinition:
			att = design.DupAtt(actual)
		case design.DataStructure:
			att = design.DupAtt(actual.Definition())
		case string:
			ut, ok := design.Design.Types[actual]
			if !ok {
				dslengine.ReportError("unknown payload type %s", actual)
			}
			att = design.DupAtt(ut.AttributeDefinition)
		case *design.Array:
			att = &design.AttributeDefinition{Type: actual}
		case *design.Hash:
			att = &design.AttributeDefinition{Type: actual}
		case design.Primitive:
			att = &design.AttributeDefinition{Type: actual}
		default:
			dslengine.ReportError("invalid Payload argument, must be a type, a media type or a DSL building a type")
			return
		}
		if len(dsls) == 1 {
			if dsl != nil {
				dslengine.ReportError("invalid arguments in Payload call, must be (type), (dsl) or (type, dsl)")
			}
			dsl = dsls[0]
		}
		if dsl != nil {
			dslengine.Execute(dsl, att)
		}
		rn := camelize(a.Parent.Name)
		an := camelize(a.Name)
		a.Payload = &design.UserTypeDefinition{
			AttributeDefinition: att,
			TypeName:            fmt.Sprintf("%s%sPayload", an, rn),
		}
		a.PayloadOptional = isOptional
	}
}

// newAttribute creates a new attribute definition using the media type with the given identifier
// as base type.
func newAttribute(baseMT string) *design.AttributeDefinition {
	var base design.DataType
	if mt := design.Design.MediaTypeWithIdentifier(baseMT); mt != nil {
		base = mt.Type
	}
	return &design.AttributeDefinition{Reference: base}
}

func camelize(str string) string {
	runes := []rune(str)
	w, i := 0, 0
	for i+1 <= len(runes) {
		eow := false
		if i+1 == len(runes) {
			eow = true
		} else if !validIdentifier(runes[i]) {
			runes = append(runes[:i], runes[i+1:]...)
		} else if spacer(runes[i+1]) {
			eow = true
			n := 1
			for i+n+1 < len(runes) && spacer(runes[i+n+1]) {
				n++
			}
			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			eow = true
		}
		i++
		if !eow {
			continue
		}
		runes[w] = unicode.ToUpper(runes[w])
		w = i
	}
	return string(runes)
}

// validIdentifier returns true if the rune is a letter or number
func validIdentifier(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func spacer(c rune) bool {
	switch c {
	case '_', ' ', ':', '-':
		return true
	}
	return false
}
