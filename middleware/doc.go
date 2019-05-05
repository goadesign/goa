/*
Package middleware contains transport independent middlewares. A middleware is a
function that accepts a endpoint function and returns another endpoint function.
The middleware should invoke the endpoint function given as argument and may
apply additional transformations prior to and after calling the original. The
middlewares included in this package include a logger middleware to log incoming
requests, a request ID middleware that makes sure every request as a unique ID
stored in the context and a couple of middlewares used to implement tracing.
*/
package middleware
