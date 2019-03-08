/*Package xray contains unary and streaming server and client interceptors that
create AWS X-Ray segments from the gRPC requests and responses and send the
segments to an AWS X-ray daemon.

The server interceptor works by extracting the tracing information setup by the
tracing server middleware. The tracing server middleware must be chained before
adding this middleware. It creates a new segment and stores the segment in the
RPC's context. User code can further configure the segment for example to set a
service version or record an error.

The client interceptor works by extracing the segment from the RPC's context
and creates a new sub-segment. It updates the RPC context with the latest trace
information.
*/
package xray
