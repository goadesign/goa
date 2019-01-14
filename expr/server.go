package expr

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"goa.design/goa/eval"
)

// WildcardRegex is the regular expression used to capture path parameters.
var WildcardRegex = regexp.MustCompile(`/{\*?([a-zA-Z0-9_]+)}`)

type (
	// ServerExpr contains a single API host information.
	ServerExpr struct {
		// Name of server
		Name string
		// Description of server
		Description string
		// Services list the services hosted by the server.
		Services []string
		// Hosts list the server hosts.
		Hosts []*HostExpr
	}

	// HostExpr describes a server host.
	HostExpr struct {
		// Name of host
		Name string
		// Name of server that uses host.
		ServerName string
		// Description of host
		Description string
		// URIs to host if any, may contain parameter elements using
		// the "{param}" syntax.
		URIs []URIExpr
		// Variables defines the URI variables if any.
		Variables *AttributeExpr
	}

	// URIExpr represents a parameterized URI.
	URIExpr string
)

// ExtractWildcards returns the names of the wildcards that appear in path.
func ExtractWildcards(path string) []string {
	matches := WildcardRegex.FindAllStringSubmatch(path, -1)
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}

// EvalName is the qualified name of the expression.
func (s *ServerExpr) EvalName() string { return "Server " + s.Name }

// Validate validates the server and server hosts.
func (s *ServerExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	for _, h := range s.Hosts {
		verr.Merge(h.Validate().(*eval.ValidationErrors))
	}
	for _, svc := range s.Services {
		if Root.Service(svc) == nil {
			verr.Add(s, "service %q undefined", svc)
		}
	}
	return verr
}

// Finalize initializes the server services and/or host with default values if
// not set explicitly in the design.
func (s *ServerExpr) Finalize() {
	if len(s.Services) == 0 {
		s.Services = make([]string, len(Root.Services))
		for i, svc := range Root.Services {
			s.Services[i] = svc.Name
		}
	}
	if len(s.Hosts) == 0 {
		s.Hosts = []*HostExpr{{
			Name:        "svc",
			Description: "Service host",
			URIs:        []URIExpr{"http://localhost:80", "grpc://localhost:8080"},
		}}
	}
	for _, h := range s.Hosts {
		h.Finalize()
	}
}

// Schemes returns the list of transport schemes used by all the server
// endpoints. The possible values for the elements of the returned slice are
// "http", "https", "grpc" and "grpcs".
func (s *ServerExpr) Schemes() []string {
	schemes := make(map[string]struct{})
	for _, h := range s.Hosts {
		for _, sch := range h.Schemes() {
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

var validSchemes = map[string]struct{}{"http": {}, "https": {}, "grpc": {}, "grpcs": {}}

// Validate validates the host.
func (h *HostExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if len(h.URIs) == 0 {
		verr.Add(h, "host must defined at least one URI")
	}
	for _, u := range h.URIs {
		vu := WildcardRegex.ReplaceAllString(string(u), "/w")
		pu, err := url.Parse(vu)
		if err != nil {
			verr.Add(h, "malformed URI %q", u)
			continue
		}
		if pu.Scheme == "" {
			verr.Add(h, "missing scheme for URI %q, scheme must be one of 'http', 'https', 'grpc' or 'grpcs'", u)
		} else if _, ok := validSchemes[pu.Scheme]; !ok {
			verr.Add(h, "invalid scheme for URI %q, scheme must be one of 'http', 'https', 'grpc' or 'grpcs'", u)
		}
	}
	if h.Variables != nil {
		for _, v := range *(h.Variables.Type.(*Object)) {
			if !IsPrimitive(v.Attribute.Type) {
				verr.Add(h, "invalid type for URI variable %q: type must be a primitive", v.Name)
			}
			if v.Attribute.Validation == nil {
				if v.Attribute.DefaultValue == nil {
					verr.Add(h, "URI variable %q must have a default value or an enum validation", v.Name)
				}
			} else if v.Attribute.DefaultValue == nil && len(v.Attribute.Validation.Values) == 0 {
				verr.Add(h, "URI variable %q must have a default value or an enum validation", v.Name)
			}
		}
	}
	return verr
}

// Finalize makes sure Variables is set.
func (h *HostExpr) Finalize() {
	if h.Variables == nil {
		h.Variables = &AttributeExpr{Type: &Object{}}
	}
}

// EvalName returns the name returned in error messages.
func (h *HostExpr) EvalName() string {
	return fmt.Sprintf("host %q of server %q", h.Name, h.ServerName)
}

// Attribute returns the variables attribute. This implements the CompositeExpr
// interface.
func (h *HostExpr) Attribute() *AttributeExpr {
	if h.Variables == nil {
		h.Variables = &AttributeExpr{Type: &Object{}}
	}
	return h.Variables
}

// Schemes returns the list of transport schemes defined for the host. The
// possible values for the elements of the returned slice are "http", "https",
// "grpc" and "grpcs".
func (h *HostExpr) Schemes() []string {
	schemes := make(map[string]struct{})
	for _, uri := range h.URIs {
		ustr := string(uri)
		// Did not use url package to find scheme because the url may
		// contain params (i.e. http://{version}.example.com) which needs
		// substition for url.Parse to succeed. Also URIs in host must have
		// a scheme otherwise validations would have failed.
		switch {
		case strings.HasPrefix(ustr, "https"):
			schemes["https"] = struct{}{}
		case strings.HasPrefix(ustr, "http"):
			schemes["http"] = struct{}{}
		case strings.HasPrefix(ustr, "grpcs"):
			schemes["grpcs"] = struct{}{}
		case strings.HasPrefix(ustr, "grpc"):
			schemes["grpc"] = struct{}{}
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

// Params return the names of the parameters used in URI if any.
func (u URIExpr) Params() []string {
	r := regexp.MustCompile(`\{([^\{\}]+)\}`)
	matches := r.FindAllStringSubmatch(string(u), -1)
	if len(matches) == 0 {
		return nil
	}
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}
