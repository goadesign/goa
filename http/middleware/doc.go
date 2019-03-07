/*Package middleware contains HTTP middlewares that wrap a HTTP handler to
provide additional functionality.

The package contains the following middlewares:

  * Logging server middleware for logging requests and responses.
  * Request ID server middleware to include a unique request ID on receiving
    a HTTP request.
  * Tracing middleware for server and client.
  * AWS X-Ray middleware for server and client that produce X-Ray segments.

Example to use the server middleware:

    var handler http.Handler = goahttp.NewMuxer()
    handler = middleware.RequestID()(handler)

Example to use the client middleware:

    var doer goahttp.Doer = &http.Client{}
    doer = xray.WrapDoer(doer)

*/
package middleware
