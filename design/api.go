package design

import (
	"regexp"

	"goa.design/goa.v2/eval"
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
		// Servers list the API hosts
		Servers []*ServerExpr
		// TermsOfAPI describes or links to the API terms of API
		TermsOfAPI string
		// Contact provides the API users with contact information
		Contact *ContactExpr
		// License describes the API license
		License *LicenseExpr
		// Docs points to the API external documentation
		Docs *DocsExpr
		// Metadata is a list of key/value pairs
		Metadata MetadataExpr
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

// EvalName is the qualified name of the expression.
func (a *APIExpr) EvalName() string     { return "API " + a.Name }
func (a *ContactExpr) EvalName() string { return "Contact " + a.Name }
func (l *LicenseExpr) EvalName() string { return "License " + l.Name }
func (d *DocsExpr) EvalName() string    { return "Documentation " + d.URL }
func (s *ServerExpr) EvalName() string  { return "Server " + s.URL }

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
	o := s.Params.Type.(Object)
	if params := URLParams(s.URL); params != nil {
		if len(params) != len(o) {
			verr.Add(s, "invalid parameter count, expected %d, got %d",
				len(params), len(o))
		} else {
			for _, p := range params {
				found := false
				for n := range o {
					if n == p {
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
	for n, p := range o {
		if p.DefaultValue == nil {
			verr.Add(s, "parameter %s has no default value", n)
		}
	}

	return verr
}

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
