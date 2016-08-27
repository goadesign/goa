package dsl

import (
	"regexp"

	"github.com/goadesign/goa/design"
)

// API defines a API exposed by a single host . It provides the API name, description and other
// global properties. There may be only one API declaration in a given design package.
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
//        TermsOfAPI("terms")
//        Contact(func() {              // Contact information
//            Name("contact name")
//            Email("contact email")
//            URL("contact URL")
//        })
//        License(func() {              // Licensing information
//            Name("license name")
//            URL("license URL")
//        })
//        Docs(func() {                 // Documentation information
//            Description("doc description")
//            URL("doc URL")
//        })
//        Host("goa.design")            // Hostname of API
//    }
//
func API(name string, dsl func()) *APIExpr {
	if name == "" {
		eval.ReportError("API first argument cannot be empty")
		return nil
	}
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	api := &APIExpr{Name: name, DSLFunc: dsl}
	design.Root.APIs = append(Root.APIs, api)
	return api
}

// Title sets the API title used by the generated documentation and code comments.
func Title(val string) {
	if s, ok := eval.Current().(*APIExpr); ok {
		s.Title = val
	} else {
		eval.IncompatibleDSL()
	}
}

// Version specifies the API version. One design describes one version.
func Version(ver string) {
	if s, ok := eval.Current().(*APIExpr); ok {
		s.Version = ver
	} else {
		eval.IncompatibleDSL()
	}
}

// Contact sets the API contact information.
func Contact(dsl func()) {
	contact := new(design.ContactExpr)
	if !eval.Execute(dsl, contact) {
		return
	}
	if a, ok := apiExpr(); ok {
		a.Contact = contact
	}
}

// License sets the API license information.
func License(dsl func()) {
	license := new(design.LicenseExpr)
	if !eval.Execute(dsl, license) {
		return
	}
	if a, ok := apiExpr(); ok {
		a.License = license
	}
}

// Docs provides external documentation pointers.
func Docs(dsl func()) {
	docs := new(design.DocsExpr)
	if !eval.Execute(dsl, docs) {
		return
	}

	switch expr := eval.Current().(type) {
	case *design.APIExpr:
		expr.Docs = docs
	case *design.ActionExpr:
		expr.Docs = docs
	case *design.FileServerExpr:
		expr.Docs = docs
	exprault:
		eval.IncompatibleDSL()
	}
}

// TermsOfAPI describes the API terms of services or links to them.
func TermsOfAPI(terms string) {
	if s, ok := eval.Current().(*APIExpr); ok {
		s.TermsOfAPI = terms
	} else {
		eval.IncompatibleDSL()
	}
}

// Regular expression used to validate RFC1035 hostnames
var hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

// Host sets the API hostname.
func Host(host string) {
	if !hostnameRegex.MatchString(host) {
		eval.ReportError(`invalid hostname value "%s"`, host)
		return
	}

	if s, ok := eval.Current().(*APIExpr); ok {
		s.Host = host
	} else {
		eval.IncompatibleDSL()
	}
}
