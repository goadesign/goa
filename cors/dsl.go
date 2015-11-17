package cors

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/raphael/goa"
)

type (
	// CheckFunc is the signature of the user provided function invoked by the middleware to
	// check whether to handle CORS headers.
	CheckFunc func(*goa.Context) bool

	// ResourceDefinition represents a CORS resource as defined by its path (or path prefix).
	ResourceDefinition struct {
		// Origin defines the origin that may access the CORS resource.
		// One and only one of Origin or OriginRegexp must be set.
		Origin string

		// OriginRegexp defines the origins that may access the CORS resource.
		// One and only one of Origin or OriginRegexp must be set.
		OriginRegexp *regexp.Regexp

		// Path is the resource URL path.
		Path string

		// IsPathPrefix is true if Path is a path prefix, false if it's an exact match.
		IsPathPrefix bool

		// Headers contains the allowed CORS request headers.
		Headers []string

		// Methods contains the allowed CORS request methods.
		Methods []string

		// Expose contains the headers that should be exposed to clients.
		Expose []string

		// MaxAge defines the value of the Access-Control-Max-Age header CORS requeets
		// response header.
		MaxAge int

		// Credentials defines the value of the Access-Control-Allow-Credentials CORS
		// requests response header.
		Credentials bool

		// Vary defines the value of the Vary response header.
		// See https://www.fastly.com/blog/best-practices-for-using-the-vary-header.
		Vary []string

		// Check is an optional user provided functions that causes CORS handling to be
		// bypassed when it return false.
		Check CheckFunc
	}

	// Specification contains the information needed to handle CORS requests.
	Specification []*ResourceDefinition
)

var (
	// spec is the CORS specification being built by the DSL.
	spec Specification

	// dslErrors contain errors encountered when running the DSL.
	dslErrors []error
)

// New runs the given CORS specification DSL and returns the built-up data structure.
func New(dsl func()) (Specification, error) {
	spec = Specification{}
	dslErrors = nil
	if dsl == nil {
		return spec, nil
	}
	dsl()
	if len(dslErrors) > 0 {
		msg := make([]string, len(dslErrors))
		for i, e := range dslErrors {
			msg[i] = e.Error()
		}
		return nil, fmt.Errorf("invalid CORS specification: %s", strings.Join(msg, ", "))
	}
	res := make([]*ResourceDefinition, len(spec))
	for i, r := range spec {
		res[i] = r
	}
	return Specification(res), nil
}

// Origin defines a group of CORS resources for the given origin.
func Origin(origin string, dsl func()) {
	existing := spec
	spec = Specification{}
	dsl()
	for _, res := range spec {
		res.Origin = origin
	}
	spec = append(existing, spec...)
}

// OriginRegex defines a group of CORS resources for the origins matching the given regex.
func OriginRegex(origin *regexp.Regexp, dsl func()) {
	existing := spec
	spec = Specification{}
	dsl()
	for _, res := range spec {
		res.OriginRegexp = origin
	}
	spec = append(existing, spec...)
}

// Resource defines a resource subject to CORS requests. The resource is defined using its URL
// path. The path can finish with the "*" wildcard character to indicate that all path under the
// given prefix target the resource.
func Resource(path string, dsl func()) {
	isPrefix := strings.HasSuffix(path, "*")
	if isPrefix {
		path = path[:len(path)-1]
	}
	res := &ResourceDefinition{Path: path, IsPathPrefix: isPrefix}
	spec = append(spec, res)
	dsl()
}

// Headers defines the HTTP headers that will be allowed in the CORS resource request.
// Use "*" to allow for any headerResources in the actual request.
func Headers(headers ...string) {
	if len(spec) == 0 {
		dslErrors = append(dslErrors, fmt.Errorf("invalid use of Headers, must define Origin and Resource first"))
	} else {
		res := spec[len(spec)-1]
		res.Headers = append(res.Headers, headers...)
	}
}

// Methods defines the HTTP methods allowed for the resource.
func Methods(methods ...string) {
	if len(spec) == 0 {
		dslErrors = append(dslErrors, fmt.Errorf("invalid use of Headers, must define Origin and Resource first"))
	} else {
		res := spec[len(spec)-1]
		for _, m := range methods {
			res.Methods = append(res.Methods, strings.ToUpper(m))
		}
	}
}

// Expose defines the HTTP headers in the resource response that can be exposed to the client.
func Expose(headers ...string) {
	if len(spec) == 0 {
		dslErrors = append(dslErrors, fmt.Errorf("invalid use of Headers, must define Origin and Resource first"))
	} else {
		res := spec[len(spec)-1]
		res.Expose = append(res.Expose, headers...)
	}
}

// MaxAge sets the Access-Control-Max-Age response header.
func MaxAge(age int) {
	if len(spec) == 0 {
		dslErrors = append(dslErrors, fmt.Errorf("invalid use of Headers, must define Origin and Resource first"))
	} else {
		res := spec[len(spec)-1]
		res.MaxAge = age
	}
}

// Credentials sets the Access-Control-Allow-Credentials response header.
func Credentials(val bool) {
	if len(spec) == 0 {
		dslErrors = append(dslErrors, fmt.Errorf("invalid use of Headers, must define Origin and Resource first"))
	} else {
		res := spec[len(spec)-1]
		res.Credentials = val
	}
}

// Vary is a list of HTTP headers to add to the 'Vary' header.
func Vary(headers ...string) {
	if len(spec) == 0 {
		dslErrors = append(dslErrors, fmt.Errorf("invalid use of Headers, must define Origin and Resource first"))
	} else {
		res := spec[len(spec)-1]
		res.Vary = append(res.Vary, headers...)
	}
}

// Check sets a function that must return true if the request is to be treated as a valid CORS
// request.
func Check(check CheckFunc) {
	if len(spec) == 0 {
		dslErrors = append(dslErrors, fmt.Errorf("invalid use of Headers, must define Origin and Resource first"))
	} else {
		res := spec[len(spec)-1]
		res.Check = check
	}
}

// String returns a human friendly representation of the CORS specification.
func (v Specification) String() string {
	if len(v) == 0 {
		return "<empty CORS specification>"
	}
	var origin string
	b := &bytes.Buffer{}
	for _, res := range v {
		o := res.Origin
		if o == "" {
			o = res.OriginRegexp.String()
		}
		if o != origin {
			b.WriteString("Origin: ")
			b.WriteString(o)
			b.WriteString("\n")
			o = origin
		}
		b.WriteString("\tPath: ")
		b.WriteString(res.Path)
		if res.IsPathPrefix {
			b.WriteString("*")
		}
		if len(res.Headers) > 0 {
			b.WriteString("\n\tHeaders: ")
			b.WriteString(strings.Join(res.Headers, ", "))
		}
		if len(res.Methods) > 0 {
			b.WriteString("\n\tMethods: ")
			b.WriteString(strings.Join(res.Methods, ", "))
		}
		if len(res.Expose) > 0 {
			b.WriteString("\n\tExpose: ")
			b.WriteString(strings.Join(res.Expose, ", "))
		}
		if res.MaxAge > 0 {
			b.WriteString(fmt.Sprintf("\n\tMaxAge: %d", res.MaxAge))
		}
		if res.MaxAge > 0 {
			b.WriteString(fmt.Sprintf("\n\tMaxAge: %d", res.MaxAge))
		}
		if len(res.Vary) > 0 {
			b.WriteString("\n\tVary: ")
			b.WriteString(strings.Join(res.Vary, ", "))
		}
	}
	return b.String()
}
