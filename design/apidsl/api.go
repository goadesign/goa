package apidsl

import (
	"fmt"
	"reflect"
	"regexp"

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
//		BaseParams(func() {			// Common parameters to all API actions
//			Param("param")
//		})
//		Consumes("application/xml", "text/xml") // Built-in encoders and decoders
//		Consumes("application/json")
//		Produces("application/gob")
//		Produces("application/json", func() {   // Custom encoder
//			Package("github.com/goadesign/encoding/json")
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
	if !dslengine.TopLevelDefinition(true) {
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
	if api, ok := apiDefinition(true); ok {
		api.Version = ver
	}
}

// Description sets the definition description.
// Description can be called inside API, Resource, Action or MediaType.
func Description(d string) {
	if a, ok := apiDefinition(false); ok {
		a.Description = d
	} else if r, ok := resourceDefinition(false); ok {
		r.Description = d
	} else if a, ok := actionDefinition(false); ok {
		a.Description = d
	} else if m, ok := mediaTypeDefinition(false); ok {
		m.Description = d
	} else if a, ok := attributeDefinition(false); ok {
		a.Description = d
	} else if r, ok := responseDefinition(false); ok {
		r.Description = d
	} else if do, ok := docsDefinition(true); ok {
		do.Description = d
	}
}

// BasePath defines the API base path, i.e. the common path prefix to all the API actions.
// The path may define wildcards (see Routing for a description of the wildcard syntax).
// The corresponding parameters must be described using BaseParams.
func BasePath(val string) {
	if a, ok := apiDefinition(false); ok {
		a.BasePath = val
	} else if r, ok := resourceDefinition(true); ok {
		r.BasePath = val
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
}

// BaseParams defines the API base path parameters. These parameters may correspond to wildcards in
// the BasePath or URL query string values.
// The DSL for describing each Param is the Attribute DSL.
func BaseParams(dsl func()) {
	params := new(design.AttributeDefinition)
	if !dslengine.Execute(dsl, params) {
		return
	}
	params.NonZeroAttributes = make(map[string]bool)
	for n := range params.Type.ToObject() {
		params.NonZeroAttributes[n] = true
	}
	if a, ok := apiDefinition(false); ok {
		a.BaseParams = params
	} else if r, ok := resourceDefinition(true); ok {
		r.BaseParams = params
	}
}

// TermsOfService describes the API terms of services or links to them.
func TermsOfService(terms string) {
	if a, ok := apiDefinition(true); ok {
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
	if a, ok := apiDefinition(true); ok {
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
	if a, ok := apiDefinition(false); ok {
		a.Schemes = append(a.Schemes, vals...)
	} else if a, ok := actionDefinition(true); ok {
		a.Schemes = append(a.Schemes, vals...)
	}
}

// Contact sets the API contact information.
func Contact(dsl func()) {
	contact := new(design.ContactDefinition)
	if !dslengine.Execute(dsl, contact) {
		return
	}
	if a, ok := apiDefinition(true); ok {
		a.Contact = contact
	}
}

// License sets the API license information.
func License(dsl func()) {
	license := new(design.LicenseDefinition)
	if !dslengine.Execute(dsl, license) {
		return
	}
	if a, ok := apiDefinition(true); ok {
		a.License = license
	}
}

// Docs provides external documentation pointers.
func Docs(dsl func()) {
	docs := new(design.DocsDefinition)
	if !dslengine.Execute(dsl, docs) {
		return
	}
	if a, ok := apiDefinition(false); ok {
		a.Docs = docs
	} else if a, ok := actionDefinition(true); ok {
		a.Docs = docs
	}
}

// Name sets the contact or license name.
func Name(name string) {
	if c, ok := contactDefinition(false); ok {
		c.Name = name
	} else if l, ok := licenseDefinition(true); ok {
		l.Name = name
	}
}

// Email sets the contact email.
func Email(email string) {
	if c, ok := contactDefinition(true); ok {
		c.Email = email
	}
}

// URL sets the contact or license URL.
func URL(url string) {
	if c, ok := contactDefinition(false); ok {
		c.URL = url
	} else if l, ok := licenseDefinition(false); ok {
		l.URL = url
	} else if d, ok := docsDefinition(true); ok {
		d.URL = url
	}
}

// Consumes adds a MIME type to the list of MIME types the APIs supports when accepting requests.
// Consumes may also specify the path of the decoding package.
// The package must expose a DecoderFactory method that returns an object which implements
// goa.DecoderFactory.
func Consumes(args ...interface{}) {
	if a, ok := apiDefinition(true); ok {
		if def := buildEncodingDefinition(args...); def != nil {
			a.Consumes = append(a.Consumes, def)
		}
	}
}

// Produces adds a MIME type to the list of MIME types the APIs can encode responses with.
// Produces may also specify the path of the encoding package.
// The package must expose a EncoderFactory method that returns an object which implements
// goa.EncoderFactory.
func Produces(args ...interface{}) {
	if a, ok := apiDefinition(true); ok {
		if def := buildEncodingDefinition(args...); def != nil {
			a.Produces = append(a.Produces, def)
		}
	}
}

// buildEncodingDefinition builds up an encoding definition.
func buildEncodingDefinition(args ...interface{}) *design.EncodingDefinition {
	var dsl func()
	var ok bool
	if len(args) == 0 {
		dslengine.ReportError("missing argument in call to Consumes")
		return nil
	}
	if _, ok := args[0].(string); !ok {
		dslengine.ReportError("first argument to Consumes must be a string (MIME type)")
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
			dslengine.ReportError("argument #%d of Consumes must be a string (MIME type)", i)
			return nil
		}
		mimeTypes[i] = mimeType
	}
	d := &design.EncodingDefinition{MIMETypes: mimeTypes}
	if dsl != nil {
		dslengine.Execute(dsl, d)
	}
	return d
}

// Package sets the Go package path to the encoder or decoder. It must be used inside a
// Consumes or Produces DSL.
func Package(path string) {
	if e, ok := encodingDefinition(true); ok {
		e.PackagePath = path
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
	if a, ok := apiDefinition(true); ok {
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
	if a, ok := apiDefinition(false); ok {
		a.Title = val
	}
}

// Trait defines an API trait. A trait encapsulates arbitrary DSL that gets executed wherever the
// trait is called via the UseTrait function.
func Trait(name string, val ...func()) {
	if a, ok := apiDefinition(true); ok {
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
// Action or Attribute DSL.
func UseTrait(name string) {
	var def dslengine.Definition
	if r, ok := resourceDefinition(false); ok {
		def = r
	} else if a, ok := actionDefinition(false); ok {
		def = a
	} else if a, ok := attributeDefinition(true); ok {
		def = a
	}
	if def != nil {
		if trait, ok := design.Design.Traits[name]; ok {
			dslengine.Execute(trait.DSLFunc, def)
		} else {
			dslengine.ReportError("unknown trait %s", name)
		}
	}
}

// apiDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func apiDefinition(failIfNotAPI bool) (*design.APIDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.APIDefinition)
	if !ok && failIfNotAPI {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return a, ok
}

// encodingDefinition returns true and current context if it is an EncodingDefinition,
// nil and false otherwise.
func encodingDefinition(failIfNotEnc bool) (*design.EncodingDefinition, bool) {
	e, ok := dslengine.CurrentDefinition().(*design.EncodingDefinition)
	if !ok && failIfNotEnc {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return e, ok
}

// contactDefinition returns true and current context if it is an ContactDefinition,
// nil and false otherwise.
func contactDefinition(failIfNotContact bool) (*design.ContactDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.ContactDefinition)
	if !ok && failIfNotContact {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return a, ok
}

// licenseDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func licenseDefinition(failIfNotLicense bool) (*design.LicenseDefinition, bool) {
	l, ok := dslengine.CurrentDefinition().(*design.LicenseDefinition)
	if !ok && failIfNotLicense {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return l, ok
}

// docsDefinition returns true and current context if it is a DocsDefinition,
// nil and false otherwise.
func docsDefinition(failIfNotDocs bool) (*design.DocsDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.DocsDefinition)
	if !ok && failIfNotDocs {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return a, ok
}

// mediaTypeDefinition returns true and current context if it is a MediaTypeDefinition,
// nil and false otherwise.
func mediaTypeDefinition(failIfNotMT bool) (*design.MediaTypeDefinition, bool) {
	m, ok := dslengine.CurrentDefinition().(*design.MediaTypeDefinition)
	if !ok && failIfNotMT {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return m, ok
}

// typeDefinition returns true and current context if it is a UserTypeDefinition,
// nil and false otherwise.
func typeDefinition(failIfNotMT bool) (*design.UserTypeDefinition, bool) {
	m, ok := dslengine.CurrentDefinition().(*design.UserTypeDefinition)
	if !ok && failIfNotMT {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return m, ok
}

// attribute returns true and current context if it is an Attribute,
// nil and false otherwise.
func attributeDefinition(failIfNotAttribute bool) (*design.AttributeDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.AttributeDefinition)
	if !ok && failIfNotAttribute {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return a, ok
}

// resourceDefinition returns true and current context if it is a ResourceDefinition,
// nil and false otherwise.
func resourceDefinition(failIfNotResource bool) (*design.ResourceDefinition, bool) {
	r, ok := dslengine.CurrentDefinition().(*design.ResourceDefinition)
	if !ok && failIfNotResource {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return r, ok
}

// actionDefinition returns true and current context if it is an ActionDefinition,
// nil and false otherwise.
func actionDefinition(failIfNotAction bool) (*design.ActionDefinition, bool) {
	a, ok := dslengine.CurrentDefinition().(*design.ActionDefinition)
	if !ok && failIfNotAction {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return a, ok
}

// responseDefinition returns true and current context if it is a ResponseDefinition,
// nil and false otherwise.
func responseDefinition(failIfNotResponse bool) (*design.ResponseDefinition, bool) {
	r, ok := dslengine.CurrentDefinition().(*design.ResponseDefinition)
	if !ok && failIfNotResponse {
		dslengine.IncompatibleDSL(dslengine.Caller())
	}
	return r, ok
}
