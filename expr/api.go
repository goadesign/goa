package expr

import (
	"sort"

	"goa.design/goa/v3/eval"
)

type (
	// APIExpr contains the global properties for a API expression.
	APIExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of API
		Name string
		// Title of API
		Title string
		// Description of API
		Description string
		// Version is the version of the API described by this DSL.
		Version string
		// Servers lists the API hosts.
		Servers []*ServerExpr
		// TermsOfService describes or links to the service terms of API.
		TermsOfService string
		// Contact provides the API users with contact information.
		Contact *ContactExpr
		// License describes the API license.
		License *LicenseExpr
		// Docs points to the API external documentation.
		Docs *DocsExpr
		// Meta is a list of key/value pairs.
		Meta MetaExpr
		// Requirements contains the security requirements that apply to
		// all the API service methods. One requirement is composed of
		// potentially multiple schemes. Incoming requests must validate
		// at least one requirement to be authorized.
		Requirements []*SecurityExpr
		// HTTP contains the HTTP specific API level expressions.
		HTTP *HTTPExpr
		// GRPC contains the gRPC specific API level expressions.
		GRPC *GRPCExpr

		// random generator used to build examples for the API types.
		random *Random
	}

	// ContactExpr contains the API contact information.
	ContactExpr struct {
		// Name of the contact person/organization
		Name string `json:"name,omitempty"`
		// Email address of the contact person/organization
		Email string `json:"email,omitempty"`
		// URL pointing to the contact information
		URL string `json:"url,omitempty"`
	}

	// LicenseExpr contains the license information for the API.
	LicenseExpr struct {
		// Name of license used for the API
		Name string `json:"name,omitempty"`
		// URL to the license used for the API
		URL string `json:"url,omitempty"`
	}

	// DocsExpr points to external documentation.
	DocsExpr struct {
		// Description of documentation.
		Description string `json:"description,omitempty"`
		// URL to documentation.
		URL string `json:"url,omitempty"`
	}
)

// NewAPIExpr initializes an API expression.
func NewAPIExpr(name string, dsl func()) *APIExpr {
	return &APIExpr{
		Name:    name,
		HTTP:    new(HTTPExpr),
		GRPC:    new(GRPCExpr),
		DSLFunc: dsl,
	}
}

// Schemes returns the list of transport schemes used by all the API servers.
// The possible values for the elements of the returned slice are "http",
// "https", "grpc" and "grpcs".
func (a *APIExpr) Schemes() []string {
	schemes := make(map[string]struct{})
	for _, s := range a.Servers {
		for _, sch := range s.Schemes() {
			schemes[sch] = struct{}{}
		}
	}
	ss := make([]string, len(schemes))
	i := 0
	for s := range schemes {
		ss[i] = s
		i++
	}
	sort.Strings(ss)
	return ss
}

// Random returns the random generator associated with a. APIs with identical
// names return generators that return the same sequence of pseudo random values.
func (a *APIExpr) Random() *Random {
	if a.random == nil {
		a.random = NewRandom(a.Name)
	}
	return a.random
}

// DefaultServer returns a server expression that describes a server which
// exposes all the services in the design and listens on localhost port 80 for
// HTTP requests and port 8080 for gRPC requests.
func (a *APIExpr) DefaultServer() *ServerExpr {
	svcs := make([]string, len(Root.Services))
	for i, svc := range Root.Services {
		svcs[i] = svc.Name
	}
	return &ServerExpr{
		Name:        a.Name,
		Description: "Default server for " + a.Name,
		Services:    svcs,
		Hosts: []*HostExpr{{
			Name:       "localhost",
			ServerName: a.Name,
			URIs:       []URIExpr{URIExpr("http://localhost:80"), URIExpr("grpc://localhost:8080")},
		}},
	}
}

// EvalName is the qualified name of the expression.
func (a *APIExpr) EvalName() string { return "API " + a.Name }

// Hash returns a unique hash value for a.
func (a *APIExpr) Hash() string { return "_api_+" + a.Name }

// Finalize makes sure that the API name is initialized and there is at least
// one server definition (if none exists, it creates a default server). If API
// name is empty, it sets the name of the first service definition as API name.
func (a *APIExpr) Finalize() {
	if a.Name == "" {
		a.Name = "api"
		if len(Root.Services) > 0 {
			a.Name = Root.Services[0].Name
		}
	}
	if len(a.Servers) == 0 {
		a.Servers = []*ServerExpr{a.DefaultServer()}
	}
}

// EvalName is the qualified name of the expression.
func (l *LicenseExpr) EvalName() string { return "License " + l.Name }

// EvalName is the qualified name of the expression.
func (d *DocsExpr) EvalName() string { return "Documentation " + d.URL }

// EvalName is the qualified name of the expression.
func (c *ContactExpr) EvalName() string { return "Contact " + c.Name }
