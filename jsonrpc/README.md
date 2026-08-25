# JSON-RPC 2.0 in Goa

Goa generates typed JSON-RPC 2.0 clients and servers from the same service
designs used for HTTP and gRPC. Generated code owns request decoding,
validation, method dispatch, response encoding, errors, notifications, batch
requests, and server-sent-event streams.

## Unary methods

Enable JSON-RPC on a service and expose each method that should be callable:

```go
var _ = Service("calc", func() {
	JSONRPC(func() {
		POST("/rpc")
	})

	Method("add", func() {
		Payload(func() {
			Attribute("a", Int)
			Attribute("b", Int)
			Required("a", "b")
		})
		Result(Int)
		JSONRPC(func() {})
	})
})
```

Every JSON-RPC method in the service shares the service route. The `method`
property inside the JSON-RPC request selects the Goa method:

```json
{"jsonrpc":"2.0","id":"sum-1","method":"add","params":{"a":2,"b":3}}
```

The generated server validates `params`, calls the service method once, and
returns the designed result:

```json
{"jsonrpc":"2.0","id":"sum-1","result":5}
```

A JSON-RPC `Response` cannot map result or error fields with `Header` or
`Cookie`. One HTTP response can carry several JSON-RPC messages in a batch or
server-sent-event stream, so an HTTP header or cookie cannot belong to one of
those messages. Put result fields in the JSON-RPC `result` value and error
fields in `error.data`. Request `Header` and `Cookie` mappings remain valid.

Calling `JSONRPC` inside a method automatically enables JSON-RPC for its
service. A service-level `JSONRPC` block is still useful for declaring the
shared route and defaults.

Objects, maps, arrays, and unions keep their designed JSON shape in `params`.
A primitive value, named primitive, byte slice, or `Any` value uses one
positional value. For example, a string payload is sent as:

```json
{"jsonrpc":"2.0","id":"echo-1","method":"echo","params":["hello"]}
```

When `Body("value")` selects an optional object, map, array, or union field,
the generated client omits `params` when that service field is absent. A
present empty array, empty map, or selected empty-message union branch still
writes `params`. An incoming `params: null` is invalid because JSON-RPC requires
parameters to be an object or array. Optional primitives keep the positional
array shape: `[null]` means absent, while `[""]` means a present empty string.
When the selected field has a default, omitted or null input applies that
default. A direct client always sends a defaulted scalar because its service
field has no absent state. A nil defaulted collection omits `params`, while an
explicit empty collection remains empty.

The server rejects a scalar `params` value because JSON-RPC permits only an
object or array there.

## Requests, notifications, and IDs

A JSON-RPC request contains an `id` and receives one response. A notification
omits `id` and receives no response, including when decoding or service work
fails.

Declare a generated client call as a notification explicitly:

```go
Method("record", func() {
	Payload(func() {
		Attribute("message", String)
		Required("message")
	})
	JSONRPC(func() {
		Notification()
	})
})
```

The generated client omits `id`, accepts the HTTP acknowledgement, and does
not decode a JSON-RPC response. The method cannot define a result, stream, or
ID field.

Goa can map the protocol ID to a designed payload field:

```go
Payload(func() {
	ID("request_id", String)
	Attribute("value", String)
	Required("value")
})
```

An ID mapping must be one direct string field. A required field with no default
rejects an absent or null incoming ID. An optional field with no default is a
pointer: nil makes the generated client create an ID, while a non-nil value is
sent exactly, including an empty string. A field with an authored default is a
value; the client sends that value and the server uses the default when the
incoming request has no ID. Without a mapped field, the client generates an
ID. The generated server sets a mapped field before calling the service.

Results cannot declare an `ID` field. The transport owns correlation and copies
the exact incoming ID, including its JSON type, into every response.

Generated clients check the complete response before decoding its result. The
response must use JSON-RPC 2.0, contain exactly one of `result` or `error`, and
repeat the string ID sent by that client call. Parse errors and invalid-request
errors may use a null ID because the server may be unable to recover the
request ID from malformed input. A response for a different call is rejected.

Servers still accept a notification for any JSON-RPC method because the
protocol defines notification per incoming message. The method runs normally,
but the server does not write a JSON-RPC response. A method whose payload marks
its ID field as required rejects such a notification because the service
payload cannot be constructed without an ID.

## Server streaming with SSE

JSON-RPC supports server-to-client streams through Server-Sent Events (SSE).
Define a `StreamingResult` and select `ServerSentEvents` in the method-level
JSON-RPC block:

```go
var Event = Type("Event", func() {
	Attribute("message", String)
	Required("message")
})

var _ = Service("updates", func() {
	JSONRPC(func() {
		POST("/rpc")
	})

	Method("watch", func() {
		Payload(func() {
			Attribute("topic", String)
			Required("topic")
		})
		StreamingResult(Event)
		JSONRPC(func() {
			ServerSentEvents()
		})
	})
})
```

The generated service method receives the same exact typed stream interface as
other Goa transports:

```go
func (s *updatesService) Watch(ctx context.Context, p *updates.WatchPayload, stream updates.WatchServerStream) error {
	if err := stream.Send(&updates.Event{Message: "ready"}); err != nil {
		return err
	}
	return nil
}
```

The stream provides:

- `Send(T) error`
- `SendWithContext(context.Context, T) error`
- `Close() error`
- `SetView(string)` when the streaming result has selectable views

For a request with an ID, each `Send` writes one SSE event whose `data:` value
is a complete JSON-RPC notification. Goa does not invent private `event:` names
for streamed values, success, or failure. When the method returns, the
transport writes one terminal event:

- success writes `result: null`;
- a returned error writes a JSON-RPC error.

A request without an ID is a JSON-RPC notification and receives no HTTP output.
The service still runs, and its `Send` calls still encode their values so the
service observes the same success or failure, but the transport discards those
bytes. Decode, encode, and service failures still reach the configured server
error handler for logging.

The generated client returns streamed notification values from `Recv`. After a
successful terminal response it returns `io.EOF`; after an error response it
returns that error. Designed error codes return the generated service error
type. An unknown error code returns `*jsonrpc.RawErrorResponse` with the exact
code, message, and data received from the server.

`SSEEventData` selects one streaming-result field for notification `params`.
A primitive selected value uses one positional value; a structured value keeps
its JSON shape. For an optional primitive, a nil service pointer is `[null]`
and an explicit empty string is `[""]`. The client treats omitted `params` and
`[null]` as absent, while preserving explicit zero values. If the field has a
default, absence uses that default. The generated client rebuilds the service
streaming result from that value. A viewed streaming result cannot use
`SSEEventData` because one selected field would omit the view name required to
decode the result.

`SSEEventID`, `SSEEventType`, and `SSEEventRetry` map streaming-result fields to
the outer `id:`, `event:`, and `retry:` lines. When a mapping is absent, Goa
omits that line. The JSON-RPC message in `data:` determines whether the event is
a streamed notification, terminal success, or terminal error; the outer event
name does not.

Generated clients require the response media type to be `text/event-stream`;
standard parameters such as `charset=utf-8` are accepted. Multiple `data:`
lines are joined with a newline. The parser accepts CR, LF, and CRLF line
endings, removes one byte-order marker at the start of the stream, and discards
an incomplete final event. A valid `id:` applies to later events until another
one changes it. An empty `id:` resets the value, while an `id:` containing NUL
is ignored. Generated servers prefix every physical encoded JSON line with
`data:` for all three line-ending forms. Event IDs sent by a generated server
cannot contain CR, LF, or NUL, event names cannot contain a line break, and
retry values must be non-negative decimal integers.

JSON-RPC does not accept `StreamingPayload` or bidirectional streaming. Use
gRPC or an ordinary HTTP WebSocket method when the client must send a stream of
values.

### Last-Event-ID

`SSERequestID` maps a string field in the initial method payload to the HTTP
`Last-Event-ID` request header. Generated clients write the header, and
generated servers read it before validating the payload:

```go
Payload(func() {
	Attribute("last_event_id", String)
})

JSONRPC(func() {
	ServerSentEvents(func() {
		SSERequestID("last_event_id")
	})
})
```

The payload field stays optional unless the design marks it required.
An absent optional field stays absent unless the design gives it a default. An
explicitly empty header remains an explicitly empty string. A required field
can be satisfied by the header because the server reads it before validating
the payload. The field is not repeated in JSON-RPC `params`.

JSON-RPC rejects a method that defines different `Result` and
`StreamingResult` types because the generated client cannot receive both from
one call. Define one method for the stream and another method for the final
resource. The two methods may share the same service and JSON-RPC path.

## Errors

Map designed errors to JSON-RPC codes in the method design:

```go
Method("divide", func() {
	Error("division_by_zero")
	JSONRPC(func() {
		Response("division_by_zero", func() {
			Code(-32001)
		})
	})
})
```

Goa also uses the standard JSON-RPC codes:

- `-32700` for malformed JSON;
- `-32600` for an invalid request;
- `-32601` for an unknown method;
- `-32602` for invalid parameters; and
- `-32603` for an unexpected service error.

Service implementations return ordinary designed or unexpected errors. The
generated transport writes the JSON-RPC error and preserves the request ID.
Generated clients report malformed response JSON as `decoding_error` and
decoded values that fail a field or view rule as `validation_error`.

## Batch requests

Unary JSON-RPC methods accept a JSON array of requests and notifications. The
generated server dispatches each item and returns an array containing responses
only for items with an ID. An all-notification batch receives no JSON-RPC
response body.

SSE streams are opened by one request and are not batch operations.

## Using JSON-RPC with other transports

A Goa method may also have ordinary `HTTP` or `GRPC` transport mappings. Each
generated transport implements the same service method contract. JSON-RPC
methods share their JSON-RPC route; ordinary HTTP routes and gRPC procedures
remain independent.

## Generation

Run Goa against the design package import path:

```bash
goa gen example.com/project/design
```

Do not edit generated files. Change the design or the owning generator and run
generation again.

Code that calls generated transport helpers directly must pass the request ID
to JSON-RPC response decoders and SSE stream constructors. Request encoders for
ordinary calls now return `(string, error)`, where the string is the ID written
to the request. Generated service method and `Send`/`Recv` interfaces are
unchanged.

See [ARCHITECTURE.md](ARCHITECTURE.md) for generator ownership and the exact SSE
message lifecycle.
