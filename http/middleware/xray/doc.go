/*Package xray contains middleware that creates AWS X-Ray segments from the
HTTP requests and responses and send the segments to an AWS X-ray daemon.

The server middleware works by extracting the trace information from the
context using the tracing middleware package. The tracing middleware must be
mounted first on the service. It stores the request segment in the context.
User code can further configure the segment for example to set a service
version or record an error.

The client middleware wraps the client Doer and works by extracing the
segment from the request context. It creates a new sub-segment and updates
the request context with the latest segment before making the request.
*/
package xray
