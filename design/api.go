package design

import (
	"net/url"
	"regexp"
	"sort"

	"goa.design/goa/eval"
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
		// Metadata is a list of key/value pairs.
		Metadata MetadataExpr
		// Requirements contains the security requirements that apply to
		// all the API service methods. One requirement is composed of
		// potentially multiple schemes. Incoming requests must validate
		// at least one requirement to be authorized.
		Requirements []*SecurityExpr

		// random generator used to build examples for the API types.
		random *Random
	}

	// ServerExpr contains a single API host information.
	ServerExpr struct {
		// Description of host
		Description string
		// URL to host, may contain parameter elements using the
		// "{param}" syntax.
		URL string
		// Params defines the URL parameters if any.
		Params *AttributeExpr
	}

	// ServerParamExpr defines a single server URL parameter.
	ServerParamExpr struct {
		*AttributeExpr
		// Name is the parameter name
		Name string
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

// Schemes returns the list of HTTP methods used by all the API servers.
func (a *APIExpr) Schemes() []string {
	schemes := make(map[string]bool)
	for _, s := range a.Servers {
		if u, err := url.Parse(s.URL); err == nil && u.Scheme != "" {
			schemes[u.Scheme] = true
		}
	}
	if len(schemes) == 0 {
		return []string{"http"}
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

// EvalName is the qualified name of the expression.
func (a *APIExpr) EvalName() string { return "API " + a.Name }

// Hash returns a unique hash value for a.
func (a *APIExpr) Hash() string { return "_api_+" + a.Name }

// Finalize makes sure there's one server definition.
func (a *APIExpr) Finalize() {
	if len(a.Servers) == 0 {
		a.Servers = []*ServerExpr{{URL: "http://localhost:8080"}}
	}
}

// EvalName is the qualified name of the expression.
func (s *ServerExpr) EvalName() string { return "Server " + s.URL }

// Attribute returns the embedded attribute.
func (s *ServerExpr) Attribute() *AttributeExpr {
	return s.Params
}

// Validate makes sure the server expression defines all the parameters and that
// for each parameter there is a default value.
func (s *ServerExpr) Validate() error {
	var (
		verr   = new(eval.ValidationErrors)
		params = URLParams(s.URL)
	)
	if s.Params == nil {
		if len(params) > 0 {
			verr.Add(s, "missing Param expressions")
		}
		return verr
	}
	o := s.Params.Type.(*Object)
	if params := URLParams(s.URL); params != nil {
		if len(params) != len(*o) {
			verr.Add(s, "invalid parameter count, expected %d, got %d",
				len(params), len(*o))
		} else {
			for _, p := range params {
				found := false
				for _, nat := range *o {
					if nat.Name == p {
						found = true
						break
					}
				}
				if !found {
					verr.Add(s, "parameter %s is not defined", p)
				}
			}
		}
	}
	for _, nat := range *o {
		if nat.Attribute.DefaultValue == nil {
			verr.Add(s, "parameter %s has no default value", nat.Name)
		}
	}

	return verr
}

// EvalName is the qualified name of the expression.
func (p *ServerParamExpr) EvalName() string { return "URL parameter " + p.Name }

// EvalName is the qualified name of the expression.
func (l *LicenseExpr) EvalName() string { return "License " + l.Name }

// EvalName is the qualified name of the expression.
func (d *DocsExpr) EvalName() string { return "Documentation " + d.URL }

// EvalName is the qualified name of the expression.
func (c *ContactExpr) EvalName() string { return "Contact " + c.Name }

// URLParamsRegexp is the regular expression used to capture the parameters
// present in a URL.
var URLParamsRegexp = regexp.MustCompile(`\{([^\{\}]+)\}`)

// URLParams returns the list of parameters present in the given URL if any.
func URLParams(url string) []string {
	matches := URLParamsRegexp.FindAllStringSubmatch(url, -1)
	if len(matches) == 0 {
		return nil
	}
	params := make([]string, len(matches))
	for i, m := range matches {
		params[i] = m[1]
	}
	return params
}
