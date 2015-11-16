package goa

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	origin             = "HTTP_ORIGIN"
	originX            = "HTTP_X_ORIGIN"
	pathInfo           = "PATH_INFO"
	vary               = "Vary"
	defaultVary        = "Origin"
	acRequestMethod    = "Http-Access-Control-Request-Method"
	acRequestHeaders   = "Http-Access-Control-Request-Headers"
	acAllowOrigin      = "Access-Control-Allow-Origin"
	acAllowMethods     = "Access-Control-Allow-Methods"
	acExposeHeaders    = "Access-Control-Expose-Headers"
	acMaxAge           = "Access-Control-Max-Age"
	acAllowCredentials = "Access-Control-Allow-Credentials"
)

type (
	// corsValidation contains the information needed to handle CORS requests from a given
	// origin.
	corsValidation []*corsResource

	// corsCheck is the signature of the user provided function invoked by the middleware to
	// check whether to handle CORS headers.
	corsCheck func(*Context) bool

	// corsResource represents a CORS resource as defined by its path (or path prefix).
	corsResource struct {
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
		Check corsCheck
		// Origin defines the origin that may access the CORS resource.
		// One and only one of Origin or OriginRegexp must be set.
		Origin string
		// OriginRegexp defines the origins that may access the CORS resource.
		// One and only one of Origin or OriginRegexp must be set.
		OriginRegexp *regexp.Regexp
	}
)

// corsV is the data structure being built by the CORS middleware DSL.
var corsV corsValidation

// CORS is a middleware that provides support for Cross-Origin Resource Sharing (CORS) as defined
// by W3C (http://www.w3.org/TR/access-control/).
// Each call to this method defines CORS resources for the given origin.
// The dsl for definining CORS resources is:
//
//	CORS(func() {
//		CORSOrigin("https://goa.design", func () {
//			CORSResource("/private", func() {
//				CORSHeaders("X-Shared-Secret")
//				CORSMethods("GET", "POST")
//				CORSExpose("X-Time")
//				CORSMaxAge(600)
//				CORSCredentials(true)
//				CORSVary("Http-Origin")
//				CORSCheck(func(ctx *Context) bool {
//					if ctx.Request.Header().Get("X-Client") == "api" {
//						return false
//					}
//					return true
//				})
//			})
//		})
//		CORSOrignRegex(regexp.MustCompile("^https?://([^\.]\.)?goa.design$"), func () {
//			CORSResource("/public/*", func() {
//				CORSMethods("GET")
//			})
//			CORSResource("/public/actions/*", func() {
//				CORSMethods("GET", "POST", "PUT", "DELETE")
//			})
//		})
//	}}
//
// Where:
//
// * CORSOrigin and CORSOriginRegex define the CORS resources for the given origin (note: the
//   semantic of resources here is the one defined in the CORS W3C standard - not to be confused
//   with REST / goa resources). The value of the first argument is matched against the incoming
//   request Http-Origin or X-Http-Origin header (if Http-Origin is absent). There can be 0 or more
//   CORSOrigin and 0 or more CORSOriginRegex function call in a CORS middleware definition, the
//   first one to match an incoming request is used.
//
// * CORSResource defines a CORS request addressable resource. The first argument defines the URL
//   path to the resource. The path may finish with the wildcard character "*" in which case it
//   matches all the URLs with the given prefix. The second argument defines the fields of the
//   resource using a simple DSL.
//
// * CORSHeaders defines the HTTP headers that will be allowed in the CORS resource request.
//   Use "*" to allow any header.
//
// * CORSMethods defines the HTTP methods allowed for the resource.
//
// * CORSExpose defines the HTTP headers in the resource response that can be exposed to the client.
//
// * CORSMaxAge sets the Access-Control-Max-Age response header.
//
// * CORSCredentials sets the Access-Control-Allow-Credentials response header.
//
// * CORSVary is a list of HTTP headers to add to the 'Vary' header.
//
// * CORSCheck is a function that returns true if the request is to be treated as a valid CORS request.
func CORS(dsl func()) Middleware {
	corsV = corsValidation{}
	dsl()
	// Copy so that further use of the CORS function doesn't affect the middleware.
	validation := corsV
	return func(h Handler) Handler {
		return func(ctx *Context) error {
			header := ctx.Request().Header
			origin := header.Get(origin)
			if origin == "" {
				origin = header.Get(originX)
			}
			var err error
			var res *corsResource
			var originHeader string
			if origin != "" {
				originHeader = origin
				if res.Origin == "*" && !res.Credentials {
					originHeader = "*"
				}
				acMethod := strings.ToUpper(header.Get(acRequestMethod))
				res = validation.RequestResource(ctx, origin)
				if ctx.Request().Method == "OPTIONS" && acMethod != "" {
					if res != nil {
						found := false
						for _, m := range res.Methods {
							if m == acMethod {
								found = true
							}
							break
						}
						if found {
							// We are responding to a preflight request.
							ctx.Header().Set("Content-Type", "text/plain")
							res.FillHeaders(originHeader, ctx.Header())
							if reqHeaders := header.Get(acRequestHeaders); reqHeaders != "" {
								ctx.Header().Set(acRequestHeaders, reqHeaders)
							}
							return ctx.Respond(200, nil)
						}
					}
				}
			}
			err = h(ctx)
			if res != nil {
				res.FillHeaders(originHeader, ctx.Header())
			} else {
				res = corsV.PathResource(ctx.Request().URL.Path)
			}
			if res != nil {
				v := ctx.Request().Header[vary]
				if len(res.Vary) > 0 {
					v = append(v, res.Vary...)
				} else {
					v = append(v, "Origin")
				}
				ctx.Header()[vary] = v
			}
			return err
		}
	}
}

// CORSOrigin defines a group of CORS resources for the given origin.
func CORSOrigin(origin string, dsl func()) {
	existing := corsV
	corsV = corsValidation{}
	dsl()
	for _, res := range corsV {
		res.Origin = origin
	}
	corsV = append(existing, corsV...)
}

// CORSOriginRegex defines a group of CORS resources for the origins matching the given regex.
func CORSOriginRegex(origin *regexp.Regexp, dsl func()) {
	existing := corsV
	corsV = corsValidation{}
	dsl()
	for _, res := range corsV {
		res.OriginRegexp = origin
	}
	corsV = append(existing, corsV...)
}

// CORSResource defines a resource subject to CORS requests. The resource is defined using its URL
// path. The path can finish with the "*" wildcard character to indicate that all path under the
// given prefix target the resource.
func CORSResource(path string, dsl func()) {
	isPrefix := strings.HasSuffix(path, "*")
	if isPrefix {
		path = path[:len(path)-1]
	}
	res := &corsResource{Path: path, IsPathPrefix: isPrefix}
	corsV = append(corsV, res)
	dsl()
}

// CORSHeaders defines the HTTP headers that will be allowed in the CORS resource request.
// Use "*" to allow for any headerResources in the actual request.
func CORSHeaders(headers ...string) {
	res := corsV[len(corsV)-1]
	res.Headers = append(res.Headers, headers...)
}

// CORSMethods defines the HTTP methods allowed for the resource.
func CORSMethods(methods ...string) {
	res := corsV[len(corsV)-1]
	for _, m := range methods {
		res.Methods = append(res.Methods, strings.ToUpper(m))
	}
}

// CORSExpose defines the HTTP headers in the resource response that can be exposed to the client.
func CORSExpose(headers ...string) {
	res := corsV[len(corsV)-1]
	res.Expose = append(res.Expose, headers...)
}

// CORSMaxAge sets the Access-Control-Max-Age response header.
func CORSMaxAge(age int) {
	res := corsV[len(corsV)-1]
	res.MaxAge = age
}

// CORSCredentials sets the Access-Control-Allow-Credentials response header.
func CORSCredentials(val bool) {
	res := corsV[len(corsV)-1]
	res.Credentials = val
}

// CORSVary is a list of HTTP headers to add to the 'Vary' header.
func CORSVary(headers ...string) {
	res := corsV[len(corsV)-1]
	res.Vary = append(res.Vary, headers...)
}

// CORSCheck is a function that returns true if the request is to be treated as a valid CORS request.
func CORSCheck(check corsCheck) {
	res := corsV[len(corsV)-1]
	res.Check = check
}

// RequestResource returns the resource targetted by the CORS request defined in ctx.
func (v corsValidation) RequestResource(ctx *Context, origin string) *corsResource {
	path := ctx.Request().URL.Path
	var match *corsResource
	for _, res := range v {
		if res.OriginAllowed(origin) && res.PathMatches(path) {
			if res.Check == nil || res.Check(ctx) {
				match = res
				break
			}
		}
	}
	return match
}

// PathResource returns the resource under the given path if any.
func (v corsValidation) PathResource(path string) *corsResource {
	var res *corsResource
	for _, r := range v {
		if r.IsPathPrefix {
			if strings.HasPrefix(path, r.Path) {
				res = r
				break
			}
		} else if r.Path == path {
			res = r
			break
		}
	}
	return res
}

// FillHeaders initializes the given header with the resource CORS headers.
func (res *corsResource) FillHeaders(origin string, header http.Header) {
	header.Set(acAllowOrigin, origin)
	header.Set(acAllowMethods, strings.Join(res.Methods, ", "))
	header.Set(acExposeHeaders, strings.Join(res.Expose, ", "))
	header.Set(acMaxAge, strconv.Itoa(res.MaxAge))
	if res.Credentials {
		header.Set(acAllowCredentials, "true")
	}
}

// OriginAllowed returns true if the resource is accessible to the given origin.
func (res *corsResource) OriginAllowed(origin string) bool {
	if res.Origin != "" {
		return res.Origin == origin
	}
	return res.OriginRegexp.MatchString(origin)
}

// PatchMatches returns true if the resource lives under the given path.
func (res *corsResource) PathMatches(path string) bool {
	if res.IsPathPrefix {
		return strings.HasPrefix(path, res.Path)
	}
	return path == res.Path
}
