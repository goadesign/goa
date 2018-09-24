package dsl

import (
	"goa.design/goa/design"
	"goa.design/goa/eval"
)

// API provides the API name, description and other properties. API also lists
// the servers that expose the services describe in the design. There may only
// be one API declaration in a given design package.
//
// API is a top level DSL. API takes two arguments: the name of the API and the
// defining DSL.
//
// The API properties are leveraged by the OpenAPI specification. The server
// expressions are also used by the server and the client tool code generators.
//
// Example:
//
//    var _ = API("adder", func() {
//        Title("title")                // Title used in documentation
//        Description("description")    // Description used in documentation
//        Version("2.0")                // Version of API
//        TermsOfService("terms")       // Terms of use
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
//    }
//
func API(name string, fn func()) *design.APIExpr {
	if name == "" {
		eval.ReportError("API first argument cannot be empty")
		return nil
	}
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	design.Root.API = &design.APIExpr{Name: name, DSLFunc: fn}
	return design.Root.API
}

// Title sets the API title. It is used by the generated OpenAPI specification.
//
// Title must appear in a API expression.
//
// Title accepts a single string argument.
//
// Example:
//
//    var _ = API("divider", func() {
//        Title("divider API")
//    })
//
func Title(val string) {
	if a, ok := eval.Current().(*design.APIExpr); ok {
		a.Title = val
		return
	}
	eval.IncompatibleDSL()
}

// Version sets the API version. It is used by the generated OpenAPI
// specification.
//
// Version must appear in a API expression.
//
// Version accepts a single string argument.
//
// Example:
//
//    var _ = API("divider", func() {
//        Version("1.0")
//    })
//
func Version(ver string) {
	if a, ok := eval.Current().(*design.APIExpr); ok {
		a.Version = ver
		return
	}
	eval.IncompatibleDSL()
}

// Docs provides external documentation URLs. It is used by the generated
// OpenAPI specification.
//
// Docs must appear in a API, Service, Method or Attribute expression.
//
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
func Docs(fn func()) {
	docs := new(design.DocsExpr)
	if !eval.Execute(fn, docs) {
		return
	}
	switch e := eval.Current().(type) {
	case *design.APIExpr:
		e.Docs = docs
	case *design.ServiceExpr:
		e.Docs = docs
	case *design.MethodExpr:
		e.Docs = docs
	case *design.AttributeExpr:
		e.Docs = docs
	default:
		eval.IncompatibleDSL()
	}
}

// Contact sets the API contact information. It is used by the generated OpenAPI
// specification.
//
// Contact must appear in a API expression.
//
// Contact takes a single argument which is the defining DSL.
//
// Example:
//
//    var _ = API("divider", func() {
//        Contact(func() {
//            Name("support")
//            Email("support@goa.design")
//            URL("https://goa.design")
//        })
//    })
//
func Contact(fn func()) {
	contact := new(design.ContactExpr)
	if !eval.Execute(fn, contact) {
		return
	}
	if a, ok := eval.Current().(*design.APIExpr); ok {
		a.Contact = contact
		return
	}
	eval.IncompatibleDSL()
}

// License sets the API license. It is used by the generated OpenAPI
// specification.
//
// License must appear in a API expression.
//
// License takes a single argument which is the defining DSL.
//
// Example:
//
//    var _ = API("divider", func() {
//        License(func() {
//            Name("MIT")
//            URL("https://github.com/goadesign/goa/blob/master/LICENSE")
//        })
//    })
//
func License(fn func()) {
	license := new(design.LicenseExpr)
	if !eval.Execute(fn, license) {
		return
	}
	if a, ok := eval.Current().(*design.APIExpr); ok {
		a.License = license
		return
	}
	eval.IncompatibleDSL()
}

// TermsOfService sets the terms of service of the API. It is used by the
// generated OpenAPI specification.
//
// TermsOfService must appear in a API expression.
//
// TermsOfService takes a single argument which is the TOS text or URL.
//
// Example:
//
//    var _ = API("github", func() {
//        TermsOfService("https://help.github.com/articles/github-terms-of-API/"
//    })
//
func TermsOfService(terms string) {
	if a, ok := eval.Current().(*design.APIExpr); ok {
		a.TermsOfService = terms
		return
	}
	eval.IncompatibleDSL()
}

// Name sets the contact or license name.
//
// Name must appear in a Contact or License expression.
//
// Name takes a single argument which is the contact or license name.
//
// Example:
//
//    var _ = API("divider", func() {
//        License(func() {
//            Name("MIT")
//            URL("https://github.com/goadesign/goa/blob/master/LICENSE")
//        })
//    })
//
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
//
// Email must appear in a Contact expression.
//
// Email takes a single argument which is the email address.
//
// Example:
//
//    var _ = API("divider", func() {
//        Contact(func() {
//            Email("support@goa.design")
//        })
//    })
//
func Email(email string) {
	if c, ok := eval.Current().(*design.ContactExpr); ok {
		c.Email = email
	}
}

// URL sets the contact, license or external documentation URL.
//
// URL must appear in Contact, License or Docs.
//
// URL accepts a single argument which is the URL.
//
// Example:
//
//    Docs(func() {
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
