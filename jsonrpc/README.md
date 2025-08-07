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

- **Transport Flexibility**: JSON-RPC methods within a single service can use
  different transports with some limitations. You can mix HTTP and SSE methods
  using content negotiation based on the `Accept` header. WebSocket methods
  require a dedicated service due to their persistent connection nature.

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

### 3. Methods Without Results  

Non-streaming methods that don't define a Result can still be called as either requests or
notifications, depending on whether an ID is provided at runtime. When called
with an ID, they return an empty success response. When called without an ID,
they behave as notifications.

**Note**: This runtime behavior applies to non-streaming methods only. WebSocket streaming
methods use explicit `SendNotification`, `SendResponse`, and `SendError` methods to control
message types (see WebSocket section below).

```go
// design/design.go
Method("log", func() {
    Description("Logs a message.")
    Payload(func() {
        Field(1, "message", String)
        Field(2, "id", String, "Optional ID")
        Meta("jsonrpc:id", "2")
    })
    // No Result() - can be request or notification
    JSONRPC(func() {})
})
```

Note: This applies to non-streaming methods only. Streaming methods have different
behavior based on their streaming pattern and transport (see the Transports section below).

### 4. Request vs Notification: Runtime Determination

In Goa's JSON-RPC implementation, whether a message is a request (expecting a response) or a notification (fire-and-forget) is determined at runtime by the presence of an ID:

- **With ID**: The message is a request and expects a response
- **Without ID or empty string ID**: The message is a notification and no response is sent

This applies to ALL methods, regardless of whether they return a result. Even methods that only return errors will behave as notifications when called without an ID.

#### Client-to-Server Messages

Any method can be called as either a request or notification by controlling the ID field:

```go
// Design
Method("process", func() {
    Payload(func() {
        Field(1, "data", String)
        Field(2, "request_id", String, "Optional request ID")
        Meta("jsonrpc:id", "2")  // Mark as JSON-RPC ID field
    })
    Result(String)
    JSONRPC(func() {})
})

// Client usage
// As request (expects response)
err := client.Process(ctx, &ProcessPayload{
    Data: "hello",
    RequestID: "123",  // ID present = request
})

// As notification (no response expected)
err := client.Process(ctx, &ProcessPayload{
    Data: "hello",
    // No RequestID = notification
})
```

#### Server-to-Client Messages (WebSocket/SSE/Mixed)

For streaming methods, servers can send both responses and notifications:

```go
// Design
Method("updates", func() {
    Payload(String)
    StreamingResult(func() {
        Field(1, "event", String)
        Field(2, "id", String, "Optional ID for responses")
        Meta("jsonrpc:id", "2")
    })
    JSONRPC(func() {})
})

// Server implementation
func (s *svc) Updates(ctx context.Context, p string, stream Updates) error {
    // Send as notification (no ID)
    stream.Send(ctx, &UpdateResult{Event: "progress 50%"})
    
    // Send as response (with ID)
    stream.Send(ctx, &UpdateResult{Event: "complete", ID: "123"})
    
    return nil
}
```

#### ID Field Design Rules

1. **Validation**: Result may only define an ID field if the corresponding Payload (or StreamingPayload) also defines one
2. **Field Naming**: Use the `Meta("jsonrpc:id", "position")` tag to mark which field is the JSON-RPC ID
3. **Field Type**: ID fields should be String type (required or optional via pointer)
4. **Required vs Optional**: Control whether ID is required using standard Goa field definitions

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
use cases. Additionally, you can combine HTTP and SSE transports within a single
service using automatic content negotiation.

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

For server-only streaming (SSE), the client initiates the stream at the service level,
but the actual stream handling happens at the transport layer. The generated HTTP
client provides access to the SSE stream.

```go
// main.go
// Use the HTTP client directly for SSE streaming
httpClient := processorjsonrpc.NewClient(
    "http", "localhost:8080", http.DefaultClient,
    goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
)

// The HTTP client's method returns the SSE stream
stream, err := httpClient.ProcessFile(ctx, &processor.ProcessFilePayload{File: "my-data.csv"})
if err != nil { /* handle error */ }

// Loop to receive messages from the SSE stream
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

    // Process the received message
    if p := res.Status.Progress; p != nil {
        log.Printf("Progress: %d%%", p.Percent)
    }
    if c := res.Status.Complete; c != nil {
        log.Printf("Done! Result at %s", c.URL)
    }
}
```

Note: The service-level client method only returns an error for server-only streaming,
as the actual stream handling is a transport concern. Use the generated HTTP/JSON-RPC
client to access the SSE stream functionality.
### WebSocket: Full Bidirectional Streaming

WebSockets provide a persistent, full-duplex connection for true real-time
communication. This is the most powerful transport, supporting client-streaming,
server-streaming, and fully bidirectional interactions.

#### Three-Method Pattern for WebSocket Streaming

Unlike non-streaming methods that determine request/notification behavior at runtime,
WebSocket streaming methods use three explicit methods to control message types:

- **`SendNotification`**: Sends a JSON-RPC notification (no response expected)
- **`SendResponse`**: Sends a JSON-RPC response with the original request ID
- **`SendError`**: Sends a JSON-RPC error response

This explicit control allows precise handling of the JSON-RPC protocol in streaming contexts.

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

    // Server-side streaming (server can push messages anytime)
    Method("subscribe", func() {
        Payload(func() {
            Attribute("topic", String)
            Required("topic")
        })
        StreamingResult(func() {
            Attribute("event", String)
            Attribute("data", Any)
            Required("event", "data")
        })
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

    // Loop to receive and dispatch client messages
    for {
        if _, err := stream.Recv(ctx); err != nil {
            return err // On error (e.g., connection closed), return to exit.
        }
    }
}

// Echo implements the bidirectional "echo" method.
func (s *chatSvc) Echo(ctx context.Context, p *chat.EchoPayload, stream chat.EchoServerStream) error {
    // Echo the message back to the client.
    return stream.SendResponse(ctx, &chat.EchoResult{
        ID: p.ID,
        Response: "You said: " + p.Message,
    })
}

// Subscribe implements server-side streaming.
// Once subscribed, the server can push messages at any time.
func (s *chatSvc) Subscribe(ctx context.Context, p *chat.SubscribePayload, stream chat.SubscribeServerStream) error {
    // Register this stream for the topic
    s.registerSubscriber(p.Topic, stream)
    defer s.unregisterSubscriber(p.Topic, stream)
    
    // Keep the stream alive
    <-ctx.Done()
    return nil
}

// In another part of your service, you can push messages to subscribers
func (s *chatSvc) publishEvent(topic string, event string, data any) {
    subscribers := s.getSubscribers(topic)
    for _, stream := range subscribers {
        // Send notification to each subscriber
        stream.SendNotification(ctx, &chat.SubscribeResult{
            Event: event,
            Data: data,
        })
    }
}
```

#### Client Usage

For WebSocket connections, the transport client manages the connection and provides
different interfaces based on the streaming pattern:

**Bidirectional Streaming** - Client gets a stream interface for both sending and receiving:

```go
// main.go
// Use the WebSocket transport client
wsClient := chatws.NewClient(
    "ws", "localhost:8080", http.DefaultClient,
    goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
    websocket.DefaultDialer, nil,
)

// For bidirectional streaming, get a stream object
stream, err := wsClient.Echo(ctx)
if err != nil { /* handle error */ }

// Send and receive concurrently
go func() {
    for i := 0; i < 5; i++ {
        err := stream.Send(&chat.EchoPayload{
            ID: fmt.Sprintf("req-%d", i),
            Message: "hello",
        })
        if err != nil { /* handle error */ }
        time.Sleep(1 * time.Second)
    }
    stream.Close()
}()

for {
    res, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatalf("receive error: %v", err)
    }
    log.Printf("received: %s", res.Response)
}
```

**Server-Side Streaming** - Client initiates subscription, then receives pushed messages:

```go
// At the service level, the method just returns an error
serviceClient := chat.NewClient(/* endpoints */)
err := serviceClient.Subscribe(ctx, &chat.SubscribePayload{Topic: "news"})
if err != nil { /* handle error */ }

// The actual stream handling happens at the transport level
// The WebSocket connection receives the pushed messages through the main stream
// established by the transport client
```

Note: For server-only streaming over WebSocket, the service-level client method returns
just an error, as receiving streamed messages is handled at the transport layer through
the persistent WebSocket connection.

### Mixed HTTP/SSE Transports

As mentioned in the Key Concepts section, Goa supports mixed transports for services 
that need both standard HTTP request-response and Server-Sent Events streaming for 
different methods. This allows you to define some methods as regular HTTP JSON-RPC 
calls and others as SSE streaming within the same service.

The server automatically handles content negotiation based on the `Accept` header:
- Requests with `Accept: text/event-stream` are routed to SSE handlers for streaming methods
- All other requests are handled as standard HTTP JSON-RPC calls

#### Design

Define methods with both HTTP and SSE transport patterns in the same service:

```go
// design/design.go
Service("processor", func() {
    JSONRPC(func() { POST("/process") })

    // Standard HTTP method - non-streaming
    Method("validate", func() {
        Payload(func() {
            Attribute("data", String)
            Required("data")
        })
        Result(func() {
            Attribute("valid", Boolean)
            Required("valid")
        })
        JSONRPC(func() {})
    })

    // SSE streaming method
    Method("process", func() {
        Payload(func() {
            Attribute("file", String)
            Required("file")
        })
        StreamingResult(func() {
            OneOf("event", func() {
                Attribute("progress", Progress)
                Attribute("complete", Complete)
            })
            Required("event")
        })
        JSONRPC(func() {
            ServerSentEvents(func() {
                SSEEventType("event")
            })
        })
    })
})
```

#### Server Implementation

Implement both types of methods normally. The framework handles the routing:

```go
// processor.go

// Regular HTTP method
func (s *processorSvc) Validate(ctx context.Context, p *processor.ValidatePayload) (*processor.ValidateResult, error) {
    // Standard synchronous processing
    valid := validateData(p.Data)
    return &processor.ValidateResult{Valid: valid}, nil
}

// SSE streaming method
func (s *processorSvc) Process(
    ctx context.Context,
    p *processor.ProcessPayload,
    stream processor.ProcessServerStream,
) error {
    // Send progress updates via SSE
    err := stream.Send(ctx, &processor.ProcessResult{
        Event: &processor.ProcessEvent{Progress: &Progress{Percent: 50}},
    })
    if err != nil {
        return err
    }

    // ... do work ...

    // Send completion
    return stream.Send(ctx, &processor.ProcessResult{
        Event: &processor.ProcessEvent{Complete: &Complete{URL: "/result"}},
    })
}
```

#### Client Usage

Use the appropriate client method for each transport:

```go
// Standard HTTP call - no special headers needed
client := processor.NewClient(/* ... */)
result, err := client.Validate(ctx, &processor.ValidatePayload{Data: "test"})

// SSE streaming call - client sets Accept header automatically
httpClient := processorjsonrpc.NewClient(/* ... */)
stream, err := httpClient.Process(ctx, &processor.ProcessPayload{File: "data.csv"})
for {
    res, err := stream.Recv()
    if err == io.EOF {
        break
    }
    // Handle streaming response
}
```

The generated client automatically sets the correct `Accept: text/event-stream` header
for SSE streaming methods, while regular methods use standard JSON content negotiation.

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