# Goa JSON-RPC Architecture

Goa implements JSON-RPC 2.0 over one HTTP POST route per service. A method
either returns one result in the HTTP response or streams results as JSON-RPC
messages carried by Server-Sent Events (SSE).

JSON-RPC does not support Goa client streams or bidirectional streams. A method
with `StreamingResult` must select `ServerSentEvents`. Ordinary HTTP methods may
still use WebSockets; that transport has its own generated code and does not
change the JSON-RPC service contract.

## Generation layers

The generated service package owns transport-neutral method and stream
interfaces. JSON-RPC uses the same exact typed per-method server stream as HTTP
SSE and gRPC:

```go
type WatchServerStream interface {
	Send(Event) error
	SendWithContext(context.Context, Event) error
	Close() error
}
```

The HTTP generation plan owns JSON request and response body types,
conversions, validation, and file imports. The JSON-RPC generator copies the
finished HTTP plan and adds JSON-RPC request dispatch and message framing. It
does not change the service interface or inspect generated types at runtime.

Generation follows this order:

1. The design is evaluated and rejects stream shapes JSON-RPC cannot carry.
2. The service plan assigns exact method and stream types.
3. The HTTP JSON-RPC plan assigns request and response body types and
   conversions.
4. The JSON-RPC plan writes only the unary or SSE code selected by the design.

## Unary requests

The server reads one JSON-RPC request, decodes its `params` into the designed
payload, calls the service endpoint once, and writes one JSON-RPC response when
the request contains an `id`. A request without an `id` is a notification and
receives no response.

The shared service route dispatches requests by their JSON-RPC `method`. Batch
requests use the same per-method handlers and collect only responses for calls
that contain an `id`.

## Server-sent-event streams

The client sends one JSON-RPC request over HTTP. Each service call to `Send`
writes an SSE event named `notification` whose data is a complete JSON-RPC
notification:

```json
{"jsonrpc":"2.0","method":"watch","params":{"value":"ready"}}
```

The service method return completes the stream. For a request with an `id`, the
transport writes exactly one terminal event:

- a `response` with `result: null` when the method succeeds; or
- an `error` containing the mapped JSON-RPC error when the method returns an
  error.

JSON-RPC rejects a method that defines different `Result` and
`StreamingResult` types because its client stream has no separate operation
that could return the final `Result`. Use one method for the stream and another
method for the final resource. gRPC has the same restriction. Ordinary HTTP
keeps mixed-result support.

A request without an `id` receives streamed notifications but no terminal
response. Request IDs and JSON-RPC messages remain transport details and never
appear in the generated service stream.

The generated client returns each notification value from `Recv`. A successful
terminal response makes the next `Recv` return `io.EOF`. A terminal JSON-RPC
error is returned as an error.

## Result conversion and views

All JSON names, required fields, transport pointers, result conversions, and
view branches are decided while generating code. A variable-view result uses
this JSON-RPC value:

```json
{"view":"summary","body":{"value":"ready"}}
```

The same representation is used for unary results and streamed notification
parameters. A fixed-view or non-viewed result uses only its designed body.

## Errors

Protocol parsing, request validation, method dispatch, designed service errors,
and unexpected service errors are mapped by the JSON-RPC transport. Service
implementations return ordinary errors; they do not write protocol error
messages themselves.

If the request has an `id`, an SSE decode or service error becomes one terminal
JSON-RPC error event. If the request has no `id`, the server writes no response
and passes the error to the configured server error handler.
