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

Calling `JSONRPC` inside a method automatically enables JSON-RPC for its
service. A service-level `JSONRPC` block is still useful for declaring the
shared route and defaults.

## Requests, notifications, and IDs

A JSON-RPC request contains an `id` and receives one response. A notification
omits `id` and receives no response, including when decoding or service work
fails.

Goa can map the protocol ID to a designed payload field:

```go
Payload(func() {
	ID("request_id", String)
	Attribute("value", String)
	Required("value")
})
```

If `request_id` is required, callers must send a request ID. If it is optional,
the generated client omits the JSON-RPC `id` when the field is empty and sends
a notification. The generated server sets the field from the incoming ID
before calling the service.

Unary result types may also declare an `ID` field. When the service returns a
non-empty result ID, the generated server uses it as the response ID. Otherwise
it uses the request ID.

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

Each `Send` writes one SSE `notification` event containing a complete JSON-RPC
notification. When the method returns, the transport writes one terminal event
for a request with an ID:

- success writes `result: null`;
- a returned error writes a JSON-RPC error.

A request without an ID receives streamed notifications but no terminal
response. The generated client returns notification values from `Recv`; after
a successful terminal response it returns `io.EOF`, and after an error response
it returns that error.

JSON-RPC does not accept `StreamingPayload` or bidirectional streaming. Use
gRPC or an ordinary HTTP WebSocket method when the client must send a stream of
values.

### Last-Event-ID

`SSERequestID` maps the incoming HTTP `Last-Event-ID` header to a string field
in the initial method payload:

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

See [ARCHITECTURE.md](ARCHITECTURE.md) for generator ownership and the exact SSE
message lifecycle.
