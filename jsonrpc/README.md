# JSON-RPC in Goa

Goa now provides first-class, type-safe support for JSON-RPC 2.0. You can build
services over HTTP, Server-Sent Events (SSE), and WebSockets using the same Goa
DSL you already know. The framework handles the protocol complexities, letting
you focus on your business logic.

## Key Concepts

### Single Endpoint Multiplexing

A core design principle of Goa's JSON-RPC support is that all methods in a
service are multiplexed over a single endpoint. Unlike REST, where each action
often has a unique URL (`/users`, `/users/{id}`), a JSON-RPC service has one URL
(e.g., `/api/jsonrpc`).

Goa uses the `method` field within the JSON-RPC payload to route incoming
requests to the correct service method. This has a few important implications:

- **Unified Transport**: All methods exposed via JSON-RPC *within a single
  service* must use the same transport. You cannot mix HTTP, SSE, and WebSocket
  JSON-RPC methods in the same service.

- **Mixed Endpoints in One Service**: You can mix JSON-RPC methods and standard
  HTTP endpoints within the same service. A method is only exposed via JSON-RPC
  if you add a `JSONRPC()` block to its design, allowing other methods in the
  same service to function as regular REST endpoints.

- **Payload-Driven**: All parameters are passed inside the JSON payload.
  Standard HTTP features like path parameters and query strings are not used for
  routing method calls.

- **Efficient Connections**: For WebSockets, this design allows multiple,
  concurrent requests and responses to share a single, persistent connection.

## Defining a JSON-RPC Service

You enable JSON-RPC at the service level to define the shared endpoint and at
the method level to expose a specific method.

### 1. Service-Level Configuration

Use the `JSONRPC` function inside a `Service` block to define the common
endpoint for all its JSON-RPC methods.

```go
// design/design.go
Service("calculator", func() {
    Description("A service for basic arithmetic.")
    // All methods in this service will be available over JSON-RPC
    // at the `/jsonrpc` endpoint.
    JSONRPC(func() {
        POST("/jsonrpc")
    })

    // ... methods defined here
})
```

### 2. Method-Level Configuration

Within the service, enable each method by adding a `JSONRPC()` block. This block
is often empty for simple cases but can also be used for mapping custom errors.

```go
// design/design.go
Method("add", func() {
    Description("Adds two integers.")
    Payload(func() {
        Attribute("a", Int, "Left-hand side")
        Attribute("b", Int, "Right-hand side")
        Required("a", "b")
    })
    Result(Int)

    // Expose this method via JSON-RPC
    JSONRPC(func() {})
})
```

### 3. Notification Methods

If a method has a Payload but no Result, Goa treats it as a JSON-RPC
notification. The client sends the request but does not expect a response.

```go
// design/design.go
Method("log", func() {
    Description("Logs a message and returns no response.")
    Payload(String)
    // No Result() makes this a notification.
    JSONRPC(func() {})
})
```

### 4. Handling Request IDs

The JSON-RPC protocol uses an `id` field to correlate requests and responses.
Goa manages this for you automatically, but you can access or override it when
needed using the `ID` function in your Payload and Result definitions.

Rule of thumb for ID attributes:

**WebSocket Services**: The requirement for ID depends on the streaming pattern:

- **Bidirectional Streaming** (StreamingPayload and StreamingResult): ID is
**REQUIRED** in both payload and result. This is crucial for correlating
responses to requests when multiple messages are in-flight on the same
connection.

- **Other Streaming Patterns** (e.g., server-streaming): ID is **OPTIONAL**.
  This allows for server-initiated notifications that are not tied to a specific
  request.

**HTTP Services**: **OPTIONAL**.

- Define an ID in the Payload only if your service logic needs to access the
  request ID (e.g., for logging).

- You generally don't need an ID in the Result, as Goa automatically mirrors the
  request ID in the response. Define one only if you need to explicitly override
  the response ID.

**SSE Services**: **OPTIONAL** but with special behavior.

- If you define an ID field in the StreamingResult, the framework uses it to
  distinguish between notifications and responses.

- Messages with an ID are treated as responses and close the stream after sending.

- Messages without an ID are treated as notifications and keep the stream open.

```go
// design/design.go
// For a bidirectional WebSocket method, IDs are required in both.
Method("echo", func() {
    StreamingPayload(func() {
        ID("request_id", String, "Request identifier for correlation")
        Attribute("data", String)
        Required("request_id", "data")
    })
    StreamingResult(func() {
        ID("request_id", String, "Correlating request identifier")
        Attribute("result", String)
        Required("request_id", "result")
    })
    JSONRPC(func() {})
})
```

## Transports

Goa supports three transports for JSON-RPC services, each suited for different
use cases.

### HTTP: Classic Request-Response

This is the standard, stateless transport for JSON-RPC. It's ideal for simple,
synchronous remote procedure calls.

#### Design

```go
// design/design.go
Service("calculator", func() {
    JSONRPC(func() { POST("/jsonrpc") })

    Method("add", func() {
        Payload(func() {
            Attribute("a", Int); Attribute("b", Int)
            Required("a", "b")
        })
        Result(Int)
        JSONRPC(func() {})
    })
})
```

#### Server Implementation

The implementation is straightforward. Goa handles the JSON-RPC protocol
wrapping.

```go
// calculator.go
func (s *calculatorSvc) Add(ctx context.Context, p *calculator.AddPayload) (res int, err error) {
    return p.A + p.B, nil
}
```

#### Client Usage

The generated client provides a simple, function-call interface.

```go
// main.go
client := calculator.NewClient(
    "http", "localhost:8080", http.DefaultClient,
    goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
)
result, err := client.Add(ctx, &calculator.AddPayload{A: 10, B: 5})
// result == 15
```

### Server-Sent Events (SSE): Server-to-Client Streaming

SSE enables unidirectional streaming from the server to the client. This is
perfect for progress updates, notifications, and real-time data feeds. The
connection is initiated with a POST request to send the initial payload.

SSE in Goa's JSON-RPC implementation uses a unified `Send` method that can send
both notifications and responses. The distinction is made automatically based on
the presence of an ID field in the message.

#### Design

Use `StreamingResult` to define the stream's data type. Here, we use `OneOf` to
send different kinds of messages on the same stream: progress updates and a
final completion event.

```go
// design/design.go
Service("processor", func() {
    JSONRPC(func() { POST("/process") }) // SSE uses POST

    Method("process_file", func() {
        Payload(func() { /* ... */ })
        StreamingResult(func() {
            // Optional: Define an ID field to enable response messages
            ID("request_id", String, "Request ID for final response")
            OneOf("status", func() {
                Attribute("progress", Progress) // Progress notification
                Attribute("complete", Complete) // Final result
            })
            Required("status")
        })
        JSONRPC(func() {
            ServerSentEvents(func() { SSEEventType("status") })
        })
    })
})
```

#### Server Implementation

Your method receives a stream object with a unified `Send` method that handles
both notifications and responses. The framework automatically determines whether
a message is a notification or response based on the presence of an ID field.

```go
// processor.go
func (s *processorSvc) ProcessFile(
    ctx context.Context,
    p *processor.ProcessFilePayload,
    stream processor.ProcessFileServerStream,
) error {
    // Send progress notifications (no ID field)
    err := stream.Send(ctx, &processor.ProcessFileResult{
        Status: &processor.ProcessFileStatus{Progress: &Progress{Percent: 50}},
    })
    if err != nil {
        return err
    }

    // ... do more work ...

    // Send the final response (with ID field if defined in the Result)
    return stream.Send(ctx, &processor.ProcessFileResult{
        Status: &processor.ProcessFileStatus{Complete: &Complete{URL: "/done.zip"}},
    })
}
```

Note: SSE streams automatically close after sending a response with an ID.
Notifications (messages without ID) keep the stream open for additional messages.

#### Client Usage

The client calls the method to get a stream object, then receives messages in a
loop until the stream is closed.

```go
// main.go
client := processor.NewClient(
    "http", "localhost:8080", http.DefaultClient,
    goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
)

// 1. Call the endpoint to get the stream
stream, err := client.ProcessFile(ctx, &processor.ProcessFilePayload{File: "my-data.csv"})
if err != nil { /* handle error */ }

// 2. Loop to receive messages
for {
    res, err := stream.Recv()
    if err == io.EOF {
        // Stream was closed cleanly by the server.
        break
    }
    if err != nil {
        // An unexpected error occurred.
        log.Fatalf("receive error: %s", err)
    }

    // 3. Process the received message
    if p := res.Status.Progress; p != nil {
        log.Printf("Progress: %d%%", p.Percent)
    }
    if c := res.Status.Complete; c != nil {
        log.Printf("Done! Result at %s", c.URL)
    }
}
```
### WebSocket: Full Bidirectional Streaming

WebSockets provide a persistent, full-duplex connection for true real-time
communication. This is the most powerful transport, supporting client-streaming,
server-streaming, and fully bidirectional interactions.

#### WebSocket Architecture

- **HandleStream Method**: Every WebSocket service requires you to implement a
  `HandleStream` method. This method manages the entire lifecycle of the
  connection.

- **stream.Recv()**: Inside `HandleStream`, you call `stream.Recv()` in a loop.
  This call blocks, waits for an incoming client message, and automatically
  dispatches it to the correct service method implementation (e.g., `subscribe`,
  `echo`).

- **Method Signatures**: The signature of your service methods changes based on
  the streaming pattern defined in the DSL:

  - **Non-streaming / Client-streaming**: `func(ctx, payload) (result, error)`

  - **Server-streaming / Bidirectional**: `func(ctx, payload, stream) error`

- **Server-Initiated Messages**: The stream object given to `HandleStream` can
  also be used to send messages to the client at any time, not just in response
  to a request.

#### Design

A single WebSocket service can contain methods for different streaming patterns.

```go
// design/design.go
Service("chat", func() {
    JSONRPC(func() { GET("/ws") }) // WebSocket connection starts with GET

    // Notifications (client streaming)
    Method("notify", func() {
        StreamingPayload(func() {
                Attribute("Event")
                Attribute("Data")
                Required("Event", "Data")
        })
        JSONRPC(func() {})
    })

    // Streaming Response
    Method("listen", func() {
        Payload(func() {
                Attribute("Topic")
                Required("Topic")
        })
        StreamingResult(func() {
                ID("id")
                Attribute("data")
                Required("id", "response")
        })
        JSONRPC(func() {})
    })

    // Bidirectional streaming
    Method("echo", func() {
        StreamingPayload(func() {
                ID("id")
                Attribute("message")
                Required("id", "message")
        })
        StreamingResult(func() {
                ID("id")
                Attribute("response")
                Required("id", "response")
        })
        JSONRPC(func() {})
    })

    // Server-initiated broadcast (no payload)
    Method("broadcast", func() {
        StreamingResult(String)
        JSONRPC(func() {})
    })
})
```

#### Server Implementation

Implement `HandleStream` to manage the connection and individual methods to
handle the logic.

```go
// chat.go

// HandleStream manages the connection lifecycle.
func (s *chatSvc) HandleStream(ctx context.Context, stream chat.Stream) error {
    defer stream.Close()

    // Example: Start a goroutine for server-initiated broadcasts
    go func() {
        for {
            time.Sleep(30 * time.Second)
            // This sends a message without a client request
            stream.Send(&chat.BroadcastResult{Message: "Server announcement!"})
        }
    }()

    // Loop to receive and dispatch client messages to `echo`, etc.
    for {
        if _, err := stream.Recv(ctx); err != nil {
            return err // On error (e.g., connection closed), return to exit.
        }
    }
}

// Echo implements the bidirectional "echo" method.
func (s *chatSvc) Echo(ctx context.Context, p *chat.EchoPayload, stream chat.EchoServerStream) error {
    // Echo the message back to the client.
    return stream.Send(&chat.EchoResult{
        RequestID: p.RequestID,
        Message: "You said: " + p.Message,
    })
}
```

#### Client Usage

The client gets a stream object that can both send and receive messages.
Goroutines are commonly used to handle this concurrently.

```go
// main.go
client := chat.NewClient(
    "ws", "localhost:8080", http.DefaultClient,
    goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
    websocket.DefaultDialer, nil,
)

// 1. Call the endpoint to get the bidirectional stream
stream, err := client.Echo(ctx)
if err != nil { /* handle error */ }

// 2. Start a goroutine to send messages to the server
go func() {
    for i := 0; i < 5; i++ {
        log.Printf("client: sending message %d", i)
        err := stream.Send(&chat.EchoPayload{
            RequestID: fmt.Sprintf("req-%d", i),
            Message: "hello",
        })
        if err != nil { /* handle error */ }
        time.Sleep(1 * time.Second)
    }
    // Close the send direction of the stream.
    stream.Close()
}()

// 3. Loop on the main goroutine to receive messages from the server
for {
    res, err := stream.Recv()
    if err == io.EOF {
        break // Stream was closed.
    }
    if err != nil {
        log.Fatalf("client: receive error: %v", err)
    }
    // Received message could be an echo response or a server broadcast
    log.Printf("client: received '%s'", res)
}
```

## Error Handling

Goa automatically handles standard JSON-RPC protocol errors (-32700, -32600,
etc.). For your application-specific errors, define them in the DSL using the
`Error` function.

### Design

You can optionally assign a custom `Code` to your error. If you do, avoid the
reserved range from -32000 to -32768.

```go
// design/design.go
Error("division_by_zero", func() {
    Description("Returned when the divisor is zero.")
    Code(-1001) // Custom application error code
})
```

### Server Implementation

Return an instance of the generated error struct from your service method.

```go
// calculator.go
func (s *calculatorSvc) Divide(ctx context.Context, p *calculator.DividePayload) (float64, error) {
    if p.B == 0 {
        return 0, &calculator.DivisionByZero{Message: "Cannot divide by zero."}
    }
    return p.A / p.B, nil
}
```

### Resulting JSON-RPC Error

Goa will serialize the error into a valid JSON-RPC error response, which looks
like this on the wire:

```json
{
    "jsonrpc": "2.0",
    "error": {
        "code": -1001,
        "message": "Cannot divide by zero.",
        "data": null
    },
    "id": "some-request-id"
}
```