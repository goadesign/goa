package dsl

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/raphael/goa/design"
)

// API implements the top level API DSL. It defines the API name, description and other global
// properties such as the base path to all the API resource actions. Here is an example showing all
// the possible API sub-definitions:
//
// 	API("API name", func() {
// 		Title("title")                          // API title used in documentation
// 		Description("description")              // API description used in documentation
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
//		Host("goa.design")                      // API hostname
//		Scheme("http")
// 		BasePath("/base/:param")                // Common base path to all API actions
// 		BaseParams(func() {                     // Common parameters to all API actions
// 			Param("param")
// 		})
// 		ResponseTemplate("static", func() {     // Response template for use by actions
// 			Description("description")
// 			Status(404)
// 			MediaType("application/json")
// 		})
// 		ResponseTemplate("dynamic", func(arg1, arg2 string) {
// 			Description(arg1)
// 			Status(200)
// 			MediaType(arg2)
// 		})
// 		Trait("Authenticated", func() {         // Traits define DSL that can be run anywhere
// 			Headers(func() {
// 				Header("header")
// 				Required("header")
// 			})
// 		})
// 	}
//
func API(name string, dsl func()) *design.APIDefinition {
	// We can't rely on this being run first, any of the top level DSL could run
	// in any order. The top level DSLs are API, Resource, MediaType and Type.
	// The first one to be called executes InitDesign.
	// API checks whether that has been called yet (i.e. if the global variable
	// Design is initialized) and if so makes sure that if it has a name it is the
	// same as the one used in the argument: API can be called multiple times as
	// long as it's always to define the same API.
	if design.Design == nil {
		InitDesign()
	} else if design.Design.Name != "" && design.Design.Name != name {
		ReportError("multiple API definitions: %#v and %#v", name, design.Design.Name)
		return nil
	}
	if !topLevelDefinition(true) {
		return nil
	}
	design.Design.Name = name
	design.Design.DSL = dsl
	return design.Design
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
					ReportError(`duplicate wildcard "%s" in API and resource base paths`, wc)
				}
			}
		}
	}
}

// BaseParams defines the API base path parameters. These parameters may correspond to wildcards in
// the BasePath or URL query string values.
// The DSL for describing each Param is the Attribute DSL.
func BaseParams(dsl func()) {
	if a, ok := apiDefinition(false); ok {
		params := new(design.AttributeDefinition)
		if executeDSL(dsl, params) {
			a.BaseParams = params
		}
	} else if r, ok := resourceDefinition(true); ok {
		params := new(design.AttributeDefinition)
		if executeDSL(dsl, params) {
			r.BaseParams = params
		}
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
	if a, ok := apiDefinition(true); ok {
		if !hostnameRegex.MatchString(host) {
			ReportError(`invalid hostname value "%s"`, host)
		} else {
			a.Host = host
		}
	}
}

// Scheme sets the API URL schemes.
func Scheme(vals ...string) {
	if a, ok := apiDefinition(true); ok {
		for _, v := range vals {
			if v != "http" && v != "https" && v != "ws" && v != "wss" {
				ReportError(`invalid scheme "%s", must be one of "http", "https", "ws" or "wss"`, v)
			} else {
				a.Schemes = append(a.Schemes, v)
			}
		}
	}
}

// Contact sets the API contact information.
func Contact(dsl func()) {
	if a, ok := apiDefinition(true); ok {
		contact := new(design.ContactDefinition)
		if executeDSL(dsl, contact) {
			a.Contact = contact
		}
	}
}

// License sets the API license information.
func License(dsl func()) {
	if a, ok := apiDefinition(true); ok {
		license := new(design.LicenseDefinition)
		if executeDSL(dsl, license) {
			a.License = license
		}
	}
}

// Docs provides external documentation pointers.
func Docs(dsl func()) {
	if a, ok := apiDefinition(false); ok {
		docs := new(design.DocsDefinition)
		if executeDSL(dsl, docs) {
			a.Docs = docs
		}
	} else if a, ok := actionDefinition(true); ok {
		docs := new(design.DocsDefinition)
		if executeDSL(dsl, docs) {
			a.Docs = docs
		}
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

// ResponseTemplate defines a response template that action definitions can use to describe their
// responses. The template may specify the HTTP response status, header specification and body media
// type. The template consists of a name and an anonymous function. The function is called when an
// action uses the template to define a response. Response template functions may accept up to 9
// string parameters that they can use to define the response fields. Here is an example of a
// response template definition that uses a function with one argument used to pass in the name of
// the response body media type:
//
//	ResponseTemplate(OK, func(mt string) {
//		Status(200)                             // OK response uses status code 200
//		Media(mt)                               // Media type name set by action definition
//		Headers(func() {
//			Header("X-Request-Id", func() { // X-Request-Id header contains a string
//				Pattern("[0-9A-F]+")    // Regexp used to validate the response header content
//			})
//			Required("X-Request-Id")        // Header is mandatory
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
			ReportError("multiple definitions for response template %s", name)
			return
		}
		if _, ok := a.ResponseTemplates[name]; ok {
			ReportError("multiple definitions for response template %s", name)
			return
		}

		setupResponseTemplate(a, name, p)
	}
}

func setupResponseTemplate(a *design.APIDefinition, name string, p interface{}) {
	if f, ok := p.(func()); ok {
		r := &design.ResponseDefinition{Name: name}
		if executeDSL(f, r) {
			a.Responses[name] = r
		}
	} else if tmpl, ok := p.(func(...string)); ok {
		t := func(params ...string) *design.ResponseDefinition {
			r := &design.ResponseDefinition{Name: name}
			executeDSL(func() { tmpl(params...) }, r)
			return r
		}
		a.ResponseTemplates[name] = &design.ResponseTemplateDefinition{
			Name:     name,
			Template: t,
		}
	} else {
		typ := reflect.TypeOf(p)
		if kind := typ.Kind(); kind != reflect.Func {
			ReportError("dsl must be a function but got %s", kind)
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
				ReportError("expected at least %s when invoking response template %s", args, name)
				return nil
			}
			r := &design.ResponseDefinition{Name: name}

			in := make([]reflect.Value, num)
			for i := 0; i < num; i++ {
				// type checking
				if t := typ.In(i); t.Kind() != reflect.String {
					ReportError("ResponseTemplate parameters must be strings but type of parameter at position %d is %s", i, t)
					return nil
				}
				// append input arguments
				in[i] = reflect.ValueOf(params[i])
			}
			executeDSL(func() { val.Call(in) }, r)
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
	if a, ok := apiDefinition(true); ok {
		a.Title = val
	}
}

// Trait defines an API trait. A trait encapsulates arbitrary DSL that gets executed wherever the
// trait is called via the UseTrait function.
func Trait(name string, val ...func()) {
	if a, ok := apiDefinition(true); ok {
		if len(val) < 1 {
			ReportError("missing trait DSL for %s", name)
			return
		} else if len(val) > 1 {
			ReportError("too many arguments given to Trait")
			return
		}
		if _, ok := a.Traits[name]; ok {
			ReportError("multiple definitions for trait %s", name)
			return
		}
		trait := &design.TraitDefinition{Name: name, DSL: val[0]}
		if a.Traits == nil {
			a.Traits = make(map[string]*design.TraitDefinition)
		}
		a.Traits[name] = trait
	}
}

// UseTrait executes the API trait with the given name. UseTrait can be used inside a Resource,
// Action or Attribute DSL.
func UseTrait(name string) {
	var def design.DSLDefinition
	if r, ok := resourceDefinition(false); ok {
		def = r
	} else if a, ok := actionDefinition(false); ok {
		def = a
	} else if a, ok := attributeDefinition(true); ok {
		def = a
	}
	if def != nil {
		if trait, ok := design.Design.Traits[name]; ok {
			executeDSL(trait.DSL, def)
		} else {
			ReportError("unknown trait %s", name)
		}
	}
}
