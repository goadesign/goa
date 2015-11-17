// Package cors provides a goa middleware that implements the Cross-Origin Resource Sharing (CORS)
// standard as defined by W3C (http://www.w3.org/TR/access-control/). CORS implements a mechanism
// to enable client-side cross-origin requests.
//
// Middleware DSL
//
// This package implements a DSL that allows goa applications to define precisely all aspects of
// CORS request handling. The DSL makes it possible to define the set of CORS resources exposed to
// given origins. For each CORS resource it is possible to specify the allowed CORS request HTTP
// methods and headers and all other CORS response properties. Additionnally CORS resources may be
// equipped with a "Check" function which gets invoked by the middleware prior to handling a CORS
// request. If this function returns false then the entire middleware is bypassed.
//
// Here is an example of a CORS specification:
//
//	New(func() {
//		Origin("https://goa.design", func () {     // This function defines CORS resources for the https://goa.design origin.
//			Resource("/private", func() {      // "/private" is the path of the CORS resource
//				Headers("X-Shared-Secret") // One or more authorized headers
//				Methods("GET", "POST")     // One or more authorized HTTP methods
//				Expose("X-Time")           // One or more headers exposed to clients
//				MaxAge(600)                // How long to cache a prefligh request response
//				Credentials(true)          // Sets Access-Control-Allow-Credentials header
//				Vary("Http-Origin")        // Sets Vary header
//				Check(func(ctx *Context) bool { // Optional function that causes the middleware to be bypassed when returning false.
//					if ctx.Request.Header().Get("X-Client") == "api" {
//						return false
//					}
//					return true
//				})
//			})
//		})
//		// Origins can be defined using regular expression with OriginRegex:
//		OrignRegex(regexp.MustCompile("^https?://([^\.]\.)?goa.design$"), func () {
//			Resource("/public/*", func() {
//				Methods("GET")
//			})
//			// Each origin may expose any number of CORS resources.
//			Resource("/public/actions/*", func() {
//				Methods("GET", "POST", "PUT", "DELETE")
//			})
//		})
//	}}
//
// CORS Middleware and the Vary HTTP Header
//
// The middleware automatically sets the "Vary" header to "Origin" unless the DSL defines a custom
// value for it. The idea is to prevent caching of responses coming from different origins. Ideally
// the application should make an effort at normalizing the value used in the  "Vary" header. See
// https://www.fastly.com/blog/best-practices-for-using-the-vary-header.
//
// CORS Usage in goa
//
// A goa service wanting to leverage this package to add support for CORS requests needs to do two
// things. First the service should mount the CORS middleware using for example:
//
//	spec := cors.New(func() {
//		// ... CORS DSL goes here
//	})
//	service.Use(cors.Middleware(spec))
//
// Secondly the service should mount the preflight controller. This controller takes care of
// handling CORS preflight requests. It should be mounted *last* to avoid collisions in the low
// level router between the service OPTIONS handler and the preflight controller handlers.
//
//	cors.MountPreflightController(service, spec)
//
package cors
