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

An optional field selected by `Body` keeps absence through the generated
transport constructor. For an object, map, array, or union, an absent service
field omits `params`; an authored empty value still writes its direct JSON
object or array. Optional primitive fields remain positional, so `[null]`
decodes as absent and an explicit zero value remains present. The server rejects
direct `params: null` instead of treating it as an absent value.
When the selected field has a default, omitted or null input applies that
default. A defaulted scalar is a service value with no absent state, so a
direct generated client always sends its value. A nil defaulted collection
omits `params`; a nonnil empty collection writes an empty array or object.

The generated request encoder returns the exact string ID written into the
request. The endpoint passes that ID to the unary response decoder or SSE
stream so the client can reject a response for another call. Notification
encoders return only an error because notifications do not create an ID.

The shared service route dispatches requests by their JSON-RPC `method`. Batch
requests use the same per-method handlers and collect only responses for calls
that contain an `id`.

Result and error fields stay inside each JSON-RPC message. Design validation
rejects `Header` and `Cookie` inside a JSON-RPC `Response` because one HTTP
response may contain several batch messages or carry a server-sent-event
stream. Request headers and cookies remain available because they describe the
one incoming HTTP request.

## Server-sent-event streams

The client sends one JSON-RPC request over HTTP. When that request has an ID,
each service call to `Send` writes an SSE event whose `data:` value is a
complete JSON-RPC notification:

```json
{"jsonrpc":"2.0","method":"watch","params":{"value":"ready"}}
```

The service method return completes the stream. For a request with an `id`, the
transport writes exactly one terminal JSON-RPC message in an SSE `data:` line:

- `result: null` when the method succeeds; or
- the mapped JSON-RPC error when the method returns an error.

Goa does not add private outer event names for notifications, success, or
failure. `SSEEventType` may map a streaming-result field to the outer `event:`
line. `SSEEventID` and `SSEEventRetry` do the same for `id:` and `retry:`. If a
mapping is absent, the generated server omits that line. The client inspects
the complete JSON-RPC message in `data:` to distinguish a streamed value from
terminal success or failure.

`SSEEventData` may select one streaming-result field for notification params.
Primitive values use one positional parameter; structured values keep their
JSON shape. An optional primitive writes `[null]` when its service pointer is
nil and keeps an explicit zero value. The client maps omitted params and
`[null]` to absence, then applies an authored default when one exists. This
mapping is rejected for viewed results because one selected field would omit
the view name required to rebuild the result. The other outer SSE mappings
remain available for viewed results.

`SSERequestID` maps one initial payload field to the HTTP `Last-Event-ID`
request header. The client writes the header and the server reads it before
payload validation. It is not part of JSON-RPC params.

Generated clients accept only the `text/event-stream` media type, with optional
standard parameters. They join multiple `data:` lines with a newline, accept
CR, LF, and CRLF line endings, remove one byte-order marker at the start of the
stream, and discard a final event that has no blank-line terminator. The last
valid `id:` value applies to later events until another `id:` changes it. An
empty `id:` resets that value, and an `id:` containing NUL is ignored.

Generated servers accept encoded JSON that uses CR, LF, or CRLF and prefix
every resulting physical line with `data:`. They reject event IDs containing
CR, LF, or NUL, event names containing a line break, and negative retry values.

JSON-RPC rejects a method that defines different `Result` and
`StreamingResult` types because its client stream has no separate operation
that could return the final `Result`. Use one method for the stream and another
method for the final resource. gRPC has the same restriction. Ordinary HTTP
keeps mixed-result support.

A request without an `id` receives no HTTP output. The generated server still
runs the service and lets every `Send` perform its normal encoding work, but a
private response writer discards headers and bytes. Decode, encode, and service
failures still reach the configured error handler for logging. Request IDs and
JSON-RPC messages remain transport details and never appear in the generated
service stream.

The generated client returns each notification value from `Recv`. A successful
terminal response makes the next `Recv` return `io.EOF`. A terminal JSON-RPC
error is returned as an error. A designed error is rebuilt as its generated
service error. An unknown error code returns `*jsonrpc.RawErrorResponse`, which
preserves the received code, message, and data.

## Result conversion and views

All JSON names, required fields, transport pointers, result conversions, and
view branches are decided while generating code. A variable-view result uses
this JSON-RPC value:

```json
{"view":"summary","body":{"value":"ready"}}
```

The same representation is used for unary results and streamed notification
parameters. A fixed-view or non-viewed result uses only its designed body.
Malformed JSON is returned as a generated client `decoding_error`. A decoded
body with a missing field, an unknown view, or another failed design rule is
returned as a generated client `validation_error`.

## Errors

Protocol parsing, request validation, method dispatch, designed service errors,
and unexpected service errors are mapped by the JSON-RPC transport. Service
implementations return ordinary errors; they do not write protocol error
messages themselves.

If the request has an `id`, an SSE decode or service error becomes one terminal
JSON-RPC error event. If the request has no `id`, the server writes no response
and passes the error to the configured server error handler.
