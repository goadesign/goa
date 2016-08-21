package dsl

import "regexp"

// Service defines a network service. It provides the service name, description and other global
// properties.
//
// Service is a top level DSL.
// Service takes two arguments: the name of the service and the defining DSL.
//
// Example:
//
//    var _ = Service("adder", func() {
//        Title("title")                // Title used in documentation
//        Description("description")    // Description used in documentation
//        Version("2.0")                // Version of service
//        TermsOfService("terms")
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
//        Host("goa.design")            // Hostname of service
//    }
//
func Service(name string, dsl func()) *ServiceExpr {
	if name == "" {
		eval.ReportError("Service first argument cannot be empty")
		return nil
	}
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	service := &ServiceExpr{Name: name, DSLFunc: dsl}
	Root.Services = append(Root.Services, service)
	return service
}

// Title sets the service title used by the generated documentation and code comments.
func Title(val string) {
	if s, ok := eval.Current().(*ServiceExpr); ok {
		s.Title = val
	} else {
		eval.IncompatibleDSL()
	}
}

// Version specifies the service version. One design describes one version.
func Version(ver string) {
	if s, ok := eval.Current().(*ServiceExpr); ok {
		s.Version = ver
	} else {
		eval.IncompatibleDSL()
	}
}

// Contact sets the service contact information.
func Contact(dsl func()) {
	contact := new(design.ContactExpr)
	if !eval.Execute(dsl, contact) {
		return
	}
	if a, ok := apiExpr(); ok {
		a.Contact = contact
	}
}

// License sets the service license information.
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
	case *design.serviceExpr:
		expr.Docs = docs
	case *design.ActionExpr:
		expr.Docs = docs
	case *design.FileServerExpr:
		expr.Docs = docs
	exprault:
		eval.IncompatibleDSL()
	}
}

// TermsOfService describes the service terms of services or links to them.
func TermsOfService(terms string) {
	if s, ok := eval.Current().(*ServiceExpr); ok {
		s.TermsOfService = terms
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

	if s, ok := eval.Current().(*ServiceExpr); ok {
		s.Host = host
	} else {
		eval.IncompatibleDSL()
	}
}
