package dsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

// API defines a network service API. It provides the API name, description and other global
// properties. There may only be one API declaration in a given design package.
//
// API is a top level DSL.
// API takes two arguments: the name of the API and the defining DSL.
//
// Example:
//
//    var _ = API("adder", func() {
//        Title("title")                // Title used in documentation
//        Description("description")    // Description used in documentation
//        Version("2.0")                // Version of API
//        TermsOfAPI("terms")           // Terms of use
//        Contact(func() {              // Contact info
//            Name("contact name")
//            Email("contact email")
//            URL("contact URL")
//        })
//        License(func() {              // License
//            Name("license name")
//            URL("license URL")
//        })
//        Docs(func() {                 // Documentation links
//            Description("doc description")
//            URL("doc URL")
//        })
//        Host("goa.design")            // Hostname used by OpenAPI spec
//    }
//
func API(name string, dsl func()) *design.APIExpr {
	if name == "" {
		eval.ReportError("API first argument cannot be empty")
		return nil
	}
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	design.Root.API = &design.APIExpr{Name: name, DSLFunc: dsl}
	return design.Root.API
}

// Title sets the API title used by the generated documentation and code comments.
func Title(val string) {
	if s, ok := eval.Current().(*design.APIExpr); ok {
		s.Title = val
		return
	}
	eval.IncompatibleDSL()
}

// Version specifies the API version. One design describes one version.
func Version(ver string) {
	if s, ok := eval.Current().(*design.APIExpr); ok {
		s.Version = ver
		return
	}
	eval.IncompatibleDSL()
}

// Contact sets the API contact information.
func Contact(dsl func()) {
	contact := new(design.ContactExpr)
	if !eval.Execute(dsl, contact) {
		return
	}
	if a, ok := eval.Current().(*design.APIExpr); ok {
		a.Contact = contact
		return
	}
	eval.IncompatibleDSL()
}

// License sets the API license information.
func License(dsl func()) {
	license := new(design.LicenseExpr)
	if !eval.Execute(dsl, license) {
		return
	}
	if a, ok := eval.Current().(*design.APIExpr); ok {
		a.License = license
		return
	}
	eval.IncompatibleDSL()
}

// Docs provides external documentation pointers.
//
// Docs may appear in an API, Service, Endpoint or Attribute expressions.
// Docs takes a single argument which is the defining DSL.
//
// Example:
//
//    var _ = API("cellar", func() {
//        Docs(func() {
//            Description("Additional documentation")
//            URL("https://goa.design")
//        })
//    })
//
func Docs(dsl func()) {
	docs := new(design.DocsExpr)
	if !eval.Execute(dsl, docs) {
		return
	}
	switch e := eval.Current().(type) {
	case *design.APIExpr:
		e.Docs = docs
	case *design.ServiceExpr:
		e.Docs = docs
	case *design.EndpointExpr:
		e.Docs = docs
	case *design.AttributeExpr:
		e.Docs = docs
	default:
		eval.IncompatibleDSL()
	}
}

// TermsOfAPI describes the API terms of services or links to them.
func TermsOfAPI(terms string) {
	if s, ok := eval.Current().(*design.APIExpr); ok {
		s.TermsOfAPI = terms
		return
	}
	eval.IncompatibleDSL()
}

// Server defines an API host.
func Server(url string, dsl ...func()) {
	if len(dsl) > 1 {
		eval.ReportError("too many arguments given to Server")
	}
	s, ok := eval.Current().(*design.APIExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if url == "" {
		eval.ReportError("Server URL cannot be empty")
	}
	server := &design.ServerExpr{
		Params: new(design.AttributeExpr),
		URL:    url,
	}
	if len(dsl) > 0 {
		eval.Execute(dsl[0], server)
	}
	s.Servers = append(s.Servers, server)
}

// Param defines a server URL parameter.
func Param(name string, args ...interface{}) {
	if _, ok := eval.Current().(*design.ServerExpr); !ok {
		eval.IncompatibleDSL()
		return
	}
	Attribute(name, args...)
}

// Name sets the contact or license name.
func Name(name string) {
	switch def := eval.Current().(type) {
	case *design.ContactExpr:
		def.Name = name
	case *design.LicenseExpr:
		def.Name = name
	default:
		eval.IncompatibleDSL()
	}
}

// Email sets the contact email.
func Email(email string) {
	if c, ok := eval.Current().(*design.ContactExpr); ok {
		c.Email = email
	}
}

// URL sets the contact, license or external documentation URL.
//
// URL may appear in Contact, License or Docs
// URL accepts a single argument which is the URL.
//
// Example:
//
//    Docs(func() {
//        Description("Additional information")
//        URL("https://goa.design")
//    })
//
func URL(url string) {
	switch def := eval.Current().(type) {
	case *design.ContactExpr:
		def.URL = url
	case *design.LicenseExpr:
		def.URL = url
	case *design.DocsExpr:
		def.URL = url
	default:
		eval.IncompatibleDSL()
	}
}
