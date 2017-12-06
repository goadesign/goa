package apidsl

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// API implements the top level API DSL. It defines the API name, default description and other
// default global property values. Here is an example showing all the possible API sub-definitions:
//
//	API("API name", func() {
//		Title("title")				// API title used in documentation
//		Description("description")		// API description used in documentation
//		Version("2.0")				// API version being described
//		TermsOfService("terms")
//		Contact(func() {			// API Contact information
//			Name("contact name")
//			Email("contact email")
//			URL("contact URL")
//		})
//		License(func() {			// API Licensing information
//			Name("license name")
//			URL("license URL")
//		})
//	 	Docs(func() {
//			Description("doc description")
//			URL("doc URL")
//		})
//		Host("goa.design")			// API hostname
//		Scheme("http")
//		BasePath("/base/:param")		// Common base path to all API actions
//		Params(func() {				// Common parameters to all API actions
//			Param("param")
//		})
//		Security("JWT")
//		Origin("http://swagger.goa.design", func() { // Define CORS policy, may be prefixed with "*" wildcard
//			Headers("X-Shared-Secret")           // One or more authorized headers, use "*" to authorize all
//			Methods("GET", "POST")               // One or more authorized HTTP methods
//			Expose("X-Time")                     // One or more headers exposed to clients
//			MaxAge(600)                          // How long to cache a prefligh request response
//			Credentials()                        // Sets Access-Control-Allow-Credentials header
//		})
//		Consumes("application/xml") // Built-in encoders and decoders
//		Consumes("application/json")
//		Produces("application/gob")
//		Produces("application/json", func() {   // Custom encoder
//			Package("github.com/goadesign/goa/encoding/json")
//		})
//		ResponseTemplate("static", func() {	// Response template for use by actions
//			Description("description")
//			Status(404)
//			MediaType("application/json")
//		})
//		ResponseTemplate("dynamic", func(arg1, arg2 string) {
//			Description(arg1)
//			Status(200)
//			MediaType(arg2)
//		})
//              NoExample()                             // Prevent automatic generation of examples
//		Trait("Authenticated", func() {		// Traits define DSL that can be run anywhere
//			Headers(func() {
//				Header("header")
//				Required("header")
//			})
//		})
//	}
//
func API(name string, dsl func()) *design.APIDefinition {
	if design.Design.Name != "" {
		dslengine.ReportError("multiple API definitions, only one is allowed")
		return nil
	}
	if !dslengine.IsTopLevelDefinition() {
		dslengine.IncompatibleDSL()
		return nil
	}

	if name == "" {
		dslengine.ReportError("API name cannot be empty")
	}
	design.Design.Name = name
	design.Design.DSLFunc = dsl
	return design.Design
}

// Version specifies the API version. One design describes one version.
func Version(ver string) {
	if api, ok := apiDefinition(); ok {
		api.Version = ver
	}
}

// Description sets the definition description.
// Description can be called inside API, Resource, Action, MediaType, Attribute, Response or ResponseTemplate
func Description(d string) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition:
		def.Description = d
	case *design.ResourceDefinition:
		def.Description = d
	case *design.FileServerDefinition:
		def.Description = d
	case *design.ActionDefinition:
		def.Description = d
	case *design.MediaTypeDefinition:
		def.Description = d
	case *design.AttributeDefinition:
		def.Description = d
	case *design.ResponseDefinition:
		def.Description = d
	case *design.DocsDefinition:
		def.Description = d
	case *design.SecuritySchemeDefinition:
		def.Description = d
	default:
		dslengine.IncompatibleDSL()
	}
}

// BasePath defines the API base path, i.e. the common path prefix to all the API actions.
// The path may define wildcards (see Routing for a description of the wildcard syntax).
// The corresponding parameters must be described using Params.
func BasePath(val string) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition:
		def.BasePath = val
	case *design.ResourceDefinition:
		def.BasePath = val
		if !strings.HasPrefix(val, "//") {
			awcs := design.ExtractWildcards(design.Design.BasePath)
			wcs := design.ExtractWildcards(val)
			for _, awc := range awcs {
				for _, wc := range wcs {
					if awc == wc {
						dslengine.ReportError(`duplicate wildcard "%s" in API and resource base paths`, wc)
					}
				}
			}
		}
	default:
		dslengine.IncompatibleDSL()
	}
}

// Origin defines the CORS policy for a given origin. The origin can use a wildcard prefix
// such as "https://*.mydomain.com". The special value "*" defines the policy for all origins
// (in which case there should be only one Origin DSL in the parent resource).
// The origin can also be a regular expression wrapped into "/".
// Example:
//
//        Origin("http://swagger.goa.design", func() { // Define CORS policy, may be prefixed with "*" wildcard
//                Headers("X-Shared-Secret")           // One or more authorized headers, use "*" to authorize all
//                Methods("GET", "POST")               // One or more authorized HTTP methods
//                Expose("X-Time")                     // One or more headers exposed to clients
//                MaxAge(600)                          // How long to cache a prefligh request response
//                Credentials()                        // Sets Access-Control-Allow-Credentials header
//        })
//
//        Origin("/(api|swagger)[.]goa[.]design/", func() {}) // Define CORS policy with a regular expression
func Origin(origin string, dsl func()) {
	cors := &design.CORSDefinition{Origin: origin}

	if strings.HasPrefix(origin, "/") && strings.HasSuffix(origin, "/") {
		cors.Regexp = true
		cors.Origin = strings.Trim(origin, "/")
	}

	if !dslengine.Execute(dsl, cors) {
		return
	}
	var parent dslengine.Definition
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition:
		parent = def
		if def.Origins == nil {
			def.Origins = make(map[string]*design.CORSDefinition)
		}
		def.Origins[origin] = cors
	case *design.ResourceDefinition:
		parent = def
		if def.Origins == nil {
			def.Origins = make(map[string]*design.CORSDefinition)
		}
		def.Origins[origin] = cors
	default:
		dslengine.IncompatibleDSL()
		return
	}
	cors.Parent = parent
}

// Methods sets the origin allowed methods. Used in Origin DSL.
func Methods(vals ...string) {
	if cors, ok := corsDefinition(); ok {
		cors.Methods = vals
	}
}

// Expose sets the origin exposed headers. Used in Origin DSL.
func Expose(vals ...string) {
	if cors, ok := corsDefinition(); ok {
		cors.Exposed = vals
	}
}

// MaxAge sets the cache expiry for preflight request responses. Used in Origin DSL.
func MaxAge(val uint) {
	if cors, ok := corsDefinition(); ok {
		cors.MaxAge = val
	}
}

// Credentials sets the allow credentials response header. Used in Origin DSL.
func Credentials() {
	if cors, ok := corsDefinition(); ok {
		cors.Credentials = true
	}
}

// TermsOfService describes the API terms of services or links to them.
func TermsOfService(terms string) {
	if a, ok := apiDefinition(); ok {
		a.TermsOfService = terms
	}
}

// Regular expression used to validate RFC1035 hostnames*/
var hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

// Host sets the API hostname.
func Host(host string) {
	if !hostnameRegex.MatchString(host) {
		dslengine.ReportError(`invalid hostname value "%s"`, host)
		return
	}

	if a, ok := apiDefinition(); ok {
		a.Host = host
	}
}

// Scheme sets the API URL schemes.
func Scheme(vals ...string) {
	ok := true
	for _, v := range vals {
		if v != "http" && v != "https" && v != "ws" && v != "wss" {
			dslengine.ReportError(`invalid scheme "%s", must be one of "http", "https", "ws" or "wss"`, v)
			ok = false
		}
	}
	if !ok {
		return
	}

	switch def := dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition:
		def.Schemes = append(def.Schemes, vals...)
	case *design.ResourceDefinition:
		def.Schemes = append(def.Schemes, vals...)
	case *design.ActionDefinition:
		def.Schemes = append(def.Schemes, vals...)
	default:
		dslengine.IncompatibleDSL()
	}
}

// Contact sets the API contact information.
func Contact(dsl func()) {
	contact := new(design.ContactDefinition)
	if !dslengine.Execute(dsl, contact) {
		return
	}
	if a, ok := apiDefinition(); ok {
		a.Contact = contact
	}
}

// License sets the API license information.
func License(dsl func()) {
	license := new(design.LicenseDefinition)
	if !dslengine.Execute(dsl, license) {
		return
	}
	if a, ok := apiDefinition(); ok {
		a.License = license
	}
}

// Docs provides external documentation pointers.
func Docs(dsl func()) {
	docs := new(design.DocsDefinition)
	if !dslengine.Execute(dsl, docs) {
		return
	}

	switch def := dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition:
		def.Docs = docs
	case *design.ActionDefinition:
		def.Docs = docs
	case *design.FileServerDefinition:
		def.Docs = docs
	default:
		dslengine.IncompatibleDSL()
	}
}

// Name sets the contact or license name.
func Name(name string) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.ContactDefinition:
		def.Name = name
	case *design.LicenseDefinition:
		def.Name = name
	default:
		dslengine.IncompatibleDSL()
	}
}

// Email sets the contact email.
func Email(email string) {
	if c, ok := contactDefinition(); ok {
		c.Email = email
	}
}

// URL can be used in: Contact, License, Docs
//
// URL sets the contact, license, or Docs URL.
func URL(url string) {
	switch def := dslengine.CurrentDefinition().(type) {
	case *design.ContactDefinition:
		def.URL = url
	case *design.LicenseDefinition:
		def.URL = url
	case *design.DocsDefinition:
		def.URL = url
	default:
		dslengine.IncompatibleDSL()
	}
}

// Consumes adds a MIME type to the list of MIME types the APIs supports when accepting requests.
// Consumes may also specify the path of the decoding package.
// The package must expose a DecoderFactory method that returns an object which implements
// goa.DecoderFactory.
func Consumes(args ...interface{}) {
	if a, ok := apiDefinition(); ok {
		if def := buildEncodingDefinition(false, args...); def != nil {
			a.Consumes = append(a.Consumes, def)
		}
	}
}

// Produces adds a MIME type to the list of MIME types the APIs can encode responses with.
// Produces may also specify the path of the encoding package.
// The package must expose a EncoderFactory method that returns an object which implements
// goa.EncoderFactory.
func Produces(args ...interface{}) {
	if a, ok := apiDefinition(); ok {
		if def := buildEncodingDefinition(true, args...); def != nil {
			a.Produces = append(a.Produces, def)
		}
	}
}

// buildEncodingDefinition builds up an encoding definition.
func buildEncodingDefinition(encoding bool, args ...interface{}) *design.EncodingDefinition {
	var dsl func()
	var ok bool
	funcName := "Consumes"
	if encoding {
		funcName = "Produces"
	}
	if len(args) == 0 {
		dslengine.ReportError("missing argument in call to %s", funcName)
		return nil
	}
	if _, ok = args[0].(string); !ok {
		dslengine.ReportError("first argument to %s must be a string (MIME type)", funcName)
		return nil
	}
	last := len(args)
	if dsl, ok = args[len(args)-1].(func()); ok {
		last = len(args) - 1
	}
	mimeTypes := make([]string, last)
	for i := 0; i < last; i++ {
		var mimeType string
		if mimeType, ok = args[i].(string); !ok {
			dslengine.ReportError("argument #%d of %s must be a string (MIME type)", i, funcName)
			return nil
		}
		mimeTypes[i] = mimeType
	}
	d := &design.EncodingDefinition{MIMETypes: mimeTypes, Encoder: encoding}
	if dsl != nil {
		dslengine.Execute(dsl, d)
	}
	return d
}

// Package sets the Go package path to the encoder or decoder. It must be used inside a
// Consumes or Produces DSL.
func Package(path string) {
	if e, ok := encodingDefinition(); ok {
		e.PackagePath = path
	}
}

// Function sets the Go function name used to instantiate the encoder or decoder. Defaults to
// NewEncoder / NewDecoder.
func Function(fn string) {
	if e, ok := encodingDefinition(); ok {
		e.Function = fn
	}
}

// ResponseTemplate defines a response template that action definitions can use to describe their
// responses. The template may specify the HTTP response status, header specification and body media
// type. The template consists of a name and an anonymous function. The function is called when an
// action uses the template to define a response. Response template functions accept string
// parameters they can use to define the response fields. Here is an example of a response template
// definition that uses a function with one argument corresponding to the name of the response body
// media type:
//
//	ResponseTemplate(OK, func(mt string) {
//		Status(200)				// OK response uses status code 200
//		Media(mt)				// Media type name set by action definition
//		Headers(func() {
//			Header("X-Request-Id", func() {	// X-Request-Id header contains a string
//				Pattern("[0-9A-F]+")	// Regexp used to validate the response header content
//			})
//			Required("X-Request-Id")	// Header is mandatory
//		})
//	})
//
// This template can the be used by actions to define the OK response as follows:
//
//	Response(OK, "vnd.goa.example")
//
// goa comes with a set of predefined response templates (one per standard HTTP status code). The
// OK template is the only one that accepts an argument. It is used as shown in the example above to
// set the response media type. Other predefined templates do not use arguments. ResponseTemplate
// makes it possible to define additional response templates specific to the API.
func ResponseTemplate(name string, p interface{}) {
	if a, ok := apiDefinition(); ok {
		if a.Responses == nil {
			a.Responses = make(map[string]*design.ResponseDefinition)
		}
		if a.ResponseTemplates == nil {
			a.ResponseTemplates = make(map[string]*design.ResponseTemplateDefinition)
		}
		if _, ok := a.Responses[name]; ok {
			dslengine.ReportError("multiple definitions for response template %s", name)
			return
		}
		if _, ok := a.ResponseTemplates[name]; ok {
			dslengine.ReportError("multiple definitions for response template %s", name)
			return
		}
		setupResponseTemplate(a, name, p)
	}
}

func setupResponseTemplate(a *design.APIDefinition, name string, p interface{}) {
	if f, ok := p.(func()); ok {
		r := &design.ResponseDefinition{Name: name}
		if dslengine.Execute(f, r) {
			a.Responses[name] = r
		}
	} else if tmpl, ok := p.(func(...string)); ok {
		t := func(params ...string) *design.ResponseDefinition {
			r := &design.ResponseDefinition{Name: name}
			dslengine.Execute(func() { tmpl(params...) }, r)
			return r
		}
		a.ResponseTemplates[name] = &design.ResponseTemplateDefinition{
			Name:     name,
			Template: t,
		}
	} else {
		typ := reflect.TypeOf(p)
		if kind := typ.Kind(); kind != reflect.Func {
			dslengine.ReportError("dsl must be a function but got %s", kind)
			return
		}

		num := typ.NumIn()
		val := reflect.ValueOf(p)
		t := func(params ...string) *design.ResponseDefinition {
			if len(params) < num {
				args := "1 argument"
				if num > 0 {
					args = fmt.Sprintf("%d arguments", num)
				}
				dslengine.ReportError("expected at least %s when invoking response template %s", args, name)
				return nil
			}
			r := &design.ResponseDefinition{Name: name}

			in := make([]reflect.Value, num)
			for i := 0; i < num; i++ {
				// type checking
				if t := typ.In(i); t.Kind() != reflect.String {
					dslengine.ReportError("ResponseTemplate parameters must be strings but type of parameter at position %d is %s", i, t)
					return nil
				}
				// append input arguments
				in[i] = reflect.ValueOf(params[i])
			}
			dslengine.Execute(func() { val.Call(in) }, r)
			return r
		}
		a.ResponseTemplates[name] = &design.ResponseTemplateDefinition{
			Name:     name,
			Template: t,
		}
	}
}

// Title sets the API title used by generated documentation, JSON Hyper-schema, code comments etc.
func Title(val string) {
	if a, ok := apiDefinition(); ok {
		a.Title = val
	}
}

// Trait defines an API trait. A trait encapsulates arbitrary DSL that gets executed wherever the
// trait is called via the UseTrait function.
func Trait(name string, val ...func()) {
	if a, ok := apiDefinition(); ok {
		if len(val) < 1 {
			dslengine.ReportError("missing trait DSL for %s", name)
			return
		} else if len(val) > 1 {
			dslengine.ReportError("too many arguments given to Trait")
			return
		}
		if _, ok := design.Design.Traits[name]; ok {
			dslengine.ReportError("multiple definitions for trait %s%s", name, design.Design.Context())
			return
		}
		trait := &dslengine.TraitDefinition{Name: name, DSLFunc: val[0]}
		if a.Traits == nil {
			a.Traits = make(map[string]*dslengine.TraitDefinition)
		}
		a.Traits[name] = trait
	}
}

// UseTrait executes the API trait with the given name. UseTrait can be used inside a Resource,
// Action, Type, MediaType or Attribute DSL.  UseTrait takes a variable number
// of trait names.
func UseTrait(names ...string) {
	var def dslengine.Definition

	switch typedDef := dslengine.CurrentDefinition().(type) {
	case *design.ResourceDefinition:
		def = typedDef
	case *design.ActionDefinition:
		def = typedDef
	case *design.AttributeDefinition:
		def = typedDef
	case *design.MediaTypeDefinition:
		def = typedDef
	default:
		dslengine.IncompatibleDSL()
	}

	if def != nil {
		for _, name := range names {
			if trait, ok := design.Design.Traits[name]; ok {
				dslengine.Execute(trait.DSLFunc, def)
			} else {
				dslengine.ReportError("unknown trait %s", name)
			}
		}
	}
}
