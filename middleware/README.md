# goa Middlewares

The `middleware` package provides middlewares that do not depend on additional packages other than
the ones already used by `goa`. These middlewares provide functionality that is useful to most
microservices:

* [LogRequest](https://goa.design/reference/goa/middleware#LogRequest) enables logging of
  incoming requests and corresponding responses. The log format is entirely configurable. The default
  format logs the request HTTP method, path and parameters as well as the corresponding
  action and controller names. It also logs the request duration and response length. It also logs
  the request payload if the DEBUG log level is enabled. Finally if the RequestID middleware is
  mounted LogRequest logs the unique request ID with each log entry.

* [LogResponse](https://goa.design/reference/goa/middleware#LogResponse) logs the content
  of the response body if the DEBUG log level is enabled.

* [RequestID](https://goa.design/reference/goa/middleware#RequestID) injects a unique ID
  in the request context. This ID is used by the logger and can be used by controller actions as
  well. The middleware looks for the ID in the [RequestIDHeader](https://goa.design/reference/goa/middleware#RequestIDHeader)
  header and if not found creates one.

* [Recover](https://goa.design/reference/goa/middleware#Recover) recover panics and logs
  the panic object and backtrace.

* [Timeout](https://goa.design/reference/goa/middleware#Timeout) sets a deadline in the
  request context. Controller actions may subscribe to the context channel to get notified when
  the timeout expires.

* [RequireHeader](https://goa.design/reference/goa/middleware#RequireHeader) checks for the
  presence of a header in the request with a value matching a given regular expression. If the
  header is absent or does not match the regexp the middleware sends a HTTP response with a given
  HTTP status.

Other middlewares listed below are provided as separate Go packages.

#### Gzip

Package [gzip](https://goa.design/reference/goa/middleware/gzip.html) contributed by
[@tylerb](https://github.com/tylerb) adds the ability to compress response bodies using gzip format
as specified in RFC 1952.

#### Security

package [security](https://goa.design/reference/goa/middleware/security.html) contains middleware
that should be used in conjunction with the security DSL.
