# JSON-RPC 2.0 in Goa

Goa provides first-class, type-safe support for JSON-RPC 2.0, enabling you to build robust RPC services with the same powerful DSL used for REST and gRPC. This implementation handles all protocol complexities while preserving Goa's design-first philosophy.

## Table of Contents

- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Protocol Fundamentals](#protocol-fundamentals)
  - [Single Endpoint Architecture](#single-endpoint-architecture)
  - [Request vs Notification](#request-vs-notification)
- [Defining Services](#defining-services)
  - [Service Configuration](#service-configuration)
  - [Method Configuration](#method-configuration)
  - [ID Field Mapping](#id-field-mapping)
- [Transport Options](#transport-options)
  - [HTTP: Request-Response](#http-request-response)
  - [Server-Sent Events: Server Streaming](#server-sent-events-server-streaming)
  - [WebSocket: Bidirectional Streaming](#websocket-bidirectional-streaming)
  - [Mixed Transports: Content Negotiation](#mixed-transports-content-negotiation)
- [Advanced Features](#advanced-features)
  - [Batch Processing](#batch-processing)
  - [Error Handling](#error-handling)
  - [Streaming Patterns](#streaming-patterns)
  - [Mixed Results](#mixed-results)
- [Best Practices](#best-practices)

## Quick Start

Define a simple JSON-RPC calculator service:

```go
// design/design.go
package design

import . "goa.design/goa/v3/dsl"

var _ = API("calculator", func() {
    Title("Calculator Service")
    Description("A simple calculator exposed via JSON-RPC")
})

var _ = Service("calc", func() {
    Description("The calc service performs basic arithmetic")
    
    // Enable JSON-RPC for this service at /rpc endpoint
    JSONRPC(func() {
        POST("/rpc")
    })
    
    // Define an add method
    Method("add", func() {
        Description("Add two numbers")
        Payload(func() {
            Attribute("a", Float64, "First operand")
            Attribute("b", Float64, "Second operand")
            Required("a", "b")
        })
        Result(Float64)
        
        // Expose this method via JSON-RPC
        JSONRPC(func() {})
    })
    
    // Define a divide method with error handling
    Method("divide", func() {
        Description("Divide two numbers")
        Payload(func() {
            Field(1, "dividend", Float64, "The dividend")
            Field(2, "divisor", Float64, "The divisor")
            Required("dividend", "divisor")
        })
        Result(Float64)
        Error("division_by_zero")
        
        JSONRPC(func() {
            Response("division_by_zero", func() {
                Code(-32001) // Custom error code
            })
        })
    })
})
```

Generate the code:

```bash
goa gen calculator/design
```

Implement the service:

```go
// calc.go
package calcapi

import (
    "context"
    calc "calculator/gen/calc"
)

type calcService struct{}

func NewCalc() calc.Service {
    return &calcService{}
}

func (s *calcService) Add(ctx context.Context, p *calc.AddPayload) (float64, error) {
    return p.A + p.B, nil
}

func (s *calcService) Divide(ctx context.Context, p *calc.DividePayload) (float64, error) {
    if p.Divisor == 0 {
        return 0, calc.MakeDivisionByZero("cannot divide by zero")
    }
    return p.Dividend / p.Divisor, nil
}
```

## Core Concepts

### Protocol Fundamentals

JSON-RPC 2.0 is a stateless, lightweight remote procedure call protocol that
uses JSON for encoding. Key characteristics:

1. **Transport Agnostic**: While commonly used over HTTP, the protocol itself doesn't specify transport
2. **Simple Message Format**: All communication uses a consistent JSON structure
3. **Bidirectional**: Supports both client-to-server and server-to-client communication
4. **Batch Support**: Multiple calls can be sent in a single request

Message structure:
```json
// Request
{
    "jsonrpc": "2.0",
    "method": "add",
    "params": {"a": 5, "b": 3},
    "id": 1
}

// Response
{
    "jsonrpc": "2.0",
    "result": 8,
    "id": 1
}
```

### Single Endpoint Architecture

Unlike REST where each resource has its own URL, JSON-RPC services multiplex all
methods through a single endpoint:

- **REST**: `/users` (GET), `/users/{id}` (GET/PUT/DELETE), `/products` (GET/POST)
- **JSON-RPC**: `/rpc` (all methods)

This design provides several benefits:

1. **Simplified Routing**: No complex URL patterns to manage
2. **Protocol Consistency**: All methods follow the same calling convention
3. **Connection Efficiency**: WebSocket/SSE connections can handle multiple methods
4. **Easy Versioning**: Version the entire API at once

The `method` field in the JSON-RPC payload determines which service method to invoke:

```json
{"jsonrpc": "2.0", "method": "add", "params": {"a": 5, "b": 3}, "id": 1}
{"jsonrpc": "2.0", "method": "divide", "params": {"dividend": 10, "divisor": 2}, "id": 2}
```

### Request vs Notification

JSON-RPC distinguishes between two types of messages based on the presence of an ID:

**Requests** (with ID) expect a response:
```json
{"jsonrpc": "2.0", "method": "process", "params": {"data": "hello"}, "id": "req-123"}
// Server MUST send a response with matching ID
```

**Notifications** (without ID) are fire-and-forget:
```json
{"jsonrpc": "2.0", "method": "log", "params": {"message": "user logged in"}}
// Server MUST NOT send a response
```

This behavior is determined at **runtime** by the client, not design time. The
same method can be called as either a request or notification.

## Defining Services

### Service Configuration

Enable JSON-RPC at the service level to define the shared endpoint:

```go
Service("myservice", func() {
    Description("A service exposed via JSON-RPC")
    
    // Define the JSON-RPC endpoint
    JSONRPC(func() {
        POST("/jsonrpc")  // For HTTP and SSE
        // OR
        GET("/ws")        // For WebSocket
    })
    
    // Define error mappings for all methods
    Error("unauthorized", func() {
        Description("Unauthorized access")
    })
    
    JSONRPC(func() {
        Response("unauthorized", func() {
            Code(-32000)  // Map to JSON-RPC error code
        })
    })
})
```

### Method Configuration

Each method needs its own `JSONRPC()` block to be exposed:

```go
Method("process", func() {
    Description("Process data")
    
    Payload(func() {
        Attribute("data", String, "Data to process")
        Attribute("priority", Int, "Processing priority")
        Required("data")
    })
    
    Result(func() {
        Attribute("output", String, "Processed output")
        Attribute("duration", Int, "Processing time in ms")
        Required("output", "duration")
    })
    
    // Enable JSON-RPC for this method
    JSONRPC(func() {
        // Method-specific error mappings (optional)
        Response("invalid_data", func() {
            Code(-32002)
        })
    })
})
```

### ID Field Mapping

Control how JSON-RPC message IDs map to your payload and result types:

```go
Method("track", func() {
    Payload(func() {
        ID("request_id", String, "Tracking ID")  // Maps to JSON-RPC request ID
        Attribute("action", String)
        Required("request_id", "action")
    })
    
    Result(func() {
        ID("request_id", String, "Tracking ID")  // Automatically copied from payload
        Attribute("status", String)
        Required("request_id", "status")
    })
    
    JSONRPC(func() {})
})
```

The `ID()` function marks which field receives the JSON-RPC message ID. Rules:

1. ID fields must be String type
2. Result can only have an ID if Payload has one
3. For non-streaming methods, the ID is automatically copied from payload to result
4. Missing ID at runtime means the message is a notification

## Transport Options

### HTTP: Request-Response

Standard synchronous RPC over HTTP. Best for:
- Simple request-response patterns
- Stateless operations
- RESTful service migration

```go
Service("api", func() {
    JSONRPC(func() {
        POST("/rpc")
    })
    
    Method("query", func() {
        Payload(func() {
            Attribute("sql", String)
            Required("sql")
        })
        Result(ArrayOf(map[string]any))
        JSONRPC(func() {})
    })
})
```

**Client usage:**
```go
client := api.NewClient("http", "localhost:8080", http.DefaultClient, 
    goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

result, err := client.Query(ctx, &api.QueryPayload{SQL: "SELECT * FROM users"})
```

**Wire format:**
```http
POST /rpc HTTP/1.1
Content-Type: application/json

{"jsonrpc":"2.0","method":"query","params":{"sql":"SELECT * FROM users"},"id":1}
```

**How it works internally:**

- The generated server inspects the first byte of the body to route batch
  (`[` starts a JSON array) vs single requests, then decodes a
  `jsonrpc.RawRequest` and validates `jsonrpc:"2.0"`, `method`, and
  `params`.
- Dispatch is by the `method` field to the corresponding generated handler
  for your service method. The handler decodes the typed payload, invokes
  your implementation, and encodes a typed JSON-RPC response via
  `MakeSuccessResponse(id, result)`.
- If the incoming message has no `id` (a notification), the server does not
  send a response, per the spec.
- Batch requests are decoded to `[]jsonrpc.RawRequest` and each entry is
  processed independently; responses are streamed into a JSON array.

### Server-Sent Events: Server Streaming

Unidirectional streaming from server to client. Perfect for:
- Progress updates
- Live notifications
- Real-time feeds
- Long-running operations

```go
Service("monitor", func() {
    JSONRPC(func() {
        POST("/events")  // SSE uses POST for initial payload
    })
    
    Method("watch", func() {
        Description("Watch system metrics")
        
        Payload(func() {
            Attribute("metrics", ArrayOf(String), "Metrics to watch")
            Required("metrics")
        })
        
        StreamingResult(func() {
            Attribute("metric", String)
            Attribute("value", Float64)
            Attribute("timestamp", String, func() {
                Format(FormatDateTime)
            })
            Required("metric", "value", "timestamp")
        })
        
        JSONRPC(func() {
            ServerSentEvents(func() {
                SSEEventType("metric")  // SSE event type field
            })
        })
    })
})
```

**Server implementation:**

```go
func (s *monitorSvc) Watch(ctx context.Context, p *monitor.WatchPayload, 
    stream monitor.WatchServerStream) error {
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            for _, metric := range p.Metrics {
                err := stream.Send(ctx, &monitor.WatchResult{
                    Metric:    metric,
                    Value:     getMetricValue(metric),
                    Timestamp: time.Now().Format(time.RFC3339),
                })
                if err != nil {
                    return err
                }
            }
        }
    }
}
```

**Client usage:**

```go
httpClient := monitorjsonrpc.NewClient(/* ... */)
stream, err := httpClient.Watch(ctx, &monitor.WatchPayload{
    Metrics: []string{"cpu", "memory"},
})

for {
    result, err := stream.Recv()
    if err == io.EOF {
        break
    }
    log.Printf("%s: %f", result.Metric, result.Value)
}
```

**How it works internally:**

- SSE uses a regular HTTP POST to deliver the initial JSON-RPC request. The
  generated handler decodes a `jsonrpc.RawRequest`, validates it, and
  dispatches to the method-specific SSE handler.
- The SSE response is a long-lived HTTP response with
  `Content-Type: text/event-stream`. The generated stream type writes events
  using standard SSE framing (`id:`, `event:`, `data:`, blank line).
- The server stream interface exposes:
  - `Send(ctx, event)`: writes a JSON-RPC notification as an SSE event
    (no response expected). Use this for progress or updates.
  - `SendAndClose(ctx, result)`: sends the final JSON-RPC response (with `id`)
    and closes the stream. The response `id` is taken from the original
    request `id`, or from a result `ID()` field if defined in the design.
  - `SendError(ctx, id, err)`: writes a JSON-RPC error response.
- Notifications vs responses:
  - Notifications omit `id` per JSON-RPC and are represented as SSE events
    with the `data:` being the result body.
  - Final responses include a JSON-RPC envelope; the SSE `id:` field mirrors
    the JSON-RPC response `id` when an ID is present.
- Example on-the-wire SSE frame (simplified):

  ```text
  event: metric
  id: 7
  data: {"jsonrpc":"2.0","result":{"metric":"cpu","value":0.9},"id":"7"}

  ```

### WebSocket: Bidirectional Streaming

Full-duplex, persistent connections for real-time communication. Ideal for:
- Chat applications
- Collaborative editing
- Gaming
- Live bidirectional data exchange

```go
Service("chat", func() {
    JSONRPC(func() {
        GET("/ws")  // WebSocket upgrade
    })
    
    // Client-to-server notifications
    Method("send", func() {
        StreamingPayload(func() {
            Attribute("message", String)
            Required("message")
        })
        JSONRPC(func() {})
    })
    
    // Server-to-client notifications
    Method("broadcast", func() {
        StreamingResult(func() {
            Attribute("from", String)
            Attribute("message", String)
            Required("from", "message")
        })
        JSONRPC(func() {})
    })
    
    // Bidirectional request-response
    Method("echo", func() {
        StreamingPayload(func() {
            ID("msg_id", String)
            Attribute("text", String)
            Required("msg_id", "text")
        })
        StreamingResult(func() {
            ID("msg_id", String)
            Attribute("echo", String)
            Required("msg_id", "echo")
        })
        JSONRPC(func() {})
    })
})
```

**Server implementation:**

```go
type chatSvc struct {
    connections map[string]chat.BroadcastServerStream
    mu          sync.RWMutex
}

func (s *chatSvc) HandleStream(ctx context.Context, stream chat.Stream) error {
    // Register connection
    connID := generateConnID()
    s.mu.Lock()
    s.connections[connID] = stream.(chat.BroadcastServerStream)
    s.mu.Unlock()
    
    defer func() {
        s.mu.Lock()
        delete(s.connections, connID)
        s.mu.Unlock()
        stream.Close()
    }()
    
    // Handle incoming messages
    for {
        _, err := stream.Recv(ctx)
        if err != nil {
            return err
        }
        // Messages are automatically dispatched to method handlers
    }
}

func (s *chatSvc) Send(ctx context.Context, p *chat.SendPayload) error {
    // Broadcast to all connections
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    for _, conn := range s.connections {
        conn.SendNotification(ctx, &chat.BroadcastResult{
            From:    "user",
            Message: p.Message,
        })
    }
    return nil
}

func (s *chatSvc) Echo(ctx context.Context, p *chat.EchoPayload, 
    stream chat.EchoServerStream) error {
    
    return stream.SendResponse(ctx, &chat.EchoResult{
        MsgID: p.MsgID,
        Echo:  "Echo: " + p.Text,
    })
}
```

**How it works internally:**

- Connection lifecycle:
  - The generated server upgrades the HTTP request to a WebSocket and
    constructs a `Stream` implementation, then calls your
    `HandleStream(ctx, stream)`.
  - Your `HandleStream` should defer `stream.Close()` and typically loop on
    `stream.Recv(ctx)`, which reads a JSON-RPC message and dispatches it to
    the appropriate generated handler based on its `method`.
- Dispatch and method invocation:
  - For non-streaming methods, `Recv` decodes the payload, invokes your
    method, and sends the typed JSON-RPC success response via the stream.
  - For streaming methods, `Recv` creates a method-specific stream wrapper
    that implements your generated `XServerStream` interface and calls your
    method implementation with it.
- Sending from your methods:
  - In server or bidirectional streaming, your method receives a stream
    wrapper providing:
    - `SendNotification(ctx, result)`: sends a JSON-RPC notification (no id).
    - `SendResponse(ctx, result)`: sends a JSON-RPC success response using the
      original request `id`. You do not need to pass the id; the wrapper holds
      it for you.
    - `SendError(ctx, err)`: sends a JSON-RPC error response correlated to the
      original request `id` when present.
- Notifications and responses:
  - Messages without `id` are notifications. Use `SendNotification` for
    server-initiated messages that should not expect a response.
  - When replying to a client request that had an `id`, use `SendResponse` to
    correlate via that `id` automatically.
- Error handling:
  - Invalid messages (parse errors, missing method) trigger JSON-RPC error
    responses when an `id` is present; otherwise they are ignored to keep the
    connection alive.
  - Unexpected WebSocket close codes abort the loop and close the connection.

### Mixed Transports: Content Negotiation

Combine HTTP and SSE in a single service using automatic content negotiation:

```go
Service("hybrid", func() {
    JSONRPC(func() {
        POST("/api")
    })
    
    // Standard HTTP method
    Method("status", func() {
        Result(func() {
            Attribute("healthy", Boolean)
            Required("healthy")
        })
        JSONRPC(func() {})
    })
    
    // SSE streaming method
    Method("monitor", func() {
        StreamingResult(func() {
            Attribute("event", String)
            Attribute("data", Any)
        })
        JSONRPC(func() {
            ServerSentEvents(func() {
                SSEEventType("update")
            })
        })
    })
    
    // Mixed results with content negotiation
    Method("flexible", func() {
        Payload(func() {
            Attribute("resource", String)
            Required("resource")
        })
        
        // Return simple result for HTTP
        Result(func() {
            Attribute("data", String)
            Required("data")
        })
        
        // Return stream for SSE
        StreamingResult(func() {
            Attribute("chunk", String)
            Attribute("progress", Int)
        })
        
        JSONRPC(func() {
            ServerSentEvents(func() {
                SSEEventType("progress")
            })
        })
    })
})
```

The server automatically routes based on the `Accept` header:
- `Accept: application/json` → HTTP handler → `Result`
- `Accept: text/event-stream` → SSE handler → `StreamingResult`

Under the hood, the generated handler checks `Accept` at runtime and invokes
the SSE stream only when `text/event-stream` is requested and the method has
`StreamingResult` (including mixed-result shapes). Otherwise, the standard
HTTP request-response path is used.

## Advanced Features

### Batch Processing

JSON-RPC supports sending multiple requests in a single HTTP call:

```json
[
    {"jsonrpc": "2.0", "method": "add", "params": {"a": 1, "b": 2}, "id": 1},
    {"jsonrpc": "2.0", "method": "multiply", "params": {"a": 3, "b": 4}, "id": 2},
    {"jsonrpc": "2.0", "method": "divide", "params": {"dividend": 10, "divisor": 2}, "id": 3}
]
```

The server processes each request independently and returns an array of responses:

```json
[
    {"jsonrpc": "2.0", "result": 3, "id": 1},
    {"jsonrpc": "2.0", "result": 12, "id": 2},
    {"jsonrpc": "2.0", "result": 5, "id": 3}
]
```

Batch processing is automatic - no special configuration needed.

### Error Handling

Goa provides comprehensive error handling with standard JSON-RPC error codes:

```go
Service("api", func() {
    // Define service-level errors
    Error("unauthorized", func() {
        Description("User is not authorized")
    })
    Error("rate_limited", func() {
        Description("Too many requests")
    })
    
    JSONRPC(func() {
        // Map errors to JSON-RPC codes
        Response("unauthorized", func() {
            Code(-32001)  // Custom application code
        })
        Response("rate_limited", func() {
            Code(-32002)
        })
    })
    
    Method("secure", func() {
        // ... method definition ...
        Error("unauthorized")  // Method can return this error
        Error("invalid_token")  // Method-specific error
        
        JSONRPC(func() {
            Response("invalid_token", func() {
                Code(-32003)
            })
        })
    })
})
```

Standard error codes:
- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error
- `-32000` to `-32099`: Reserved for implementation

### Streaming Patterns

#### Client Streaming (WebSocket only)
```go
Method("upload", func() {
    StreamingPayload(func() {
        Attribute("chunk", Bytes)
        Attribute("offset", Int64)
        Required("chunk", "offset")
    })
    Result(func() {
        Attribute("size", Int64)
        Attribute("checksum", String)
    })
    JSONRPC(func() {})
})
```

#### Server Streaming (SSE or WebSocket)
```go
Method("download", func() {
    Payload(func() {
        Attribute("file", String)
        Required("file")
    })
    StreamingResult(func() {
        Attribute("chunk", Bytes)
        Attribute("offset", Int64)
        Required("chunk", "offset")
    })
    JSONRPC(func() {
        ServerSentEvents(func() {})  // Or use WebSocket
    })
})
```

#### Bidirectional Streaming (WebSocket only)
```go
Method("transform", func() {
    StreamingPayload(func() {
        ID("seq", String)
        Attribute("input", String)
        Required("seq", "input")
    })
    StreamingResult(func() {
        ID("seq", String)
        Attribute("output", String)
        Required("seq", "output")
    })
    JSONRPC(func() {})
})
```

### Mixed Results

Support different response types based on content negotiation:

```go
Method("report", func() {
    Payload(func() {
        Attribute("query", String)
        Required("query")
    })
    
    // Simple result for synchronous HTTP
    Result(func() {
        Attribute("summary", String)
        Attribute("count", Int)
        Required("summary", "count")
    })
    
    // Streaming result for SSE
    StreamingResult(func() {
        Attribute("row", Map(String, Any))
        Attribute("progress", Float64)
    })
    
    JSONRPC(func() {
        ServerSentEvents(func() {
            SSEEventType("row")
        })
    })
})
```

Implementation:

```go
// Called for Accept: application/json
func (s *svc) Report(ctx context.Context, p *ReportPayload) (*ReportResult, error) {
    summary, count := generateReport(p.Query)
    return &ReportResult{Summary: summary, Count: count}, nil
}

// Called for Accept: text/event-stream
func (s *svc) ReportStream(ctx context.Context, p *ReportPayload, 
    stream ReportServerStream) error {
    
    rows := queryRows(p.Query)
    for i, row := range rows {
        err := stream.Send(ctx, &ReportStreamingResult{
            Row:      row,
            Progress: float64(i) / float64(len(rows)),
        })
        if err != nil {
            return err
        }
    }
    return nil
}
```

## Best Practices

### 1. Service Design

**DO:**
- Group related methods in the same service
- Use consistent naming conventions
- Define clear error codes and messages
- Document expected behavior

**DON'T:**
- Mix WebSocket with HTTP endpoints in the same service
- Use deeply nested payload structures
- Rely on transport-specific features

### 2. Error Handling

**DO:**
- Map application errors to appropriate JSON-RPC codes
- Provide meaningful error messages
- Use standard codes when applicable
- Include error data when helpful

**DON'T:**
- Use reserved error code ranges
- Return stack traces in production
- Ignore validation errors

### 3. Streaming

**DO:**
- Use SSE for server-push scenarios
- Use WebSocket for bidirectional needs
- Implement proper cleanup in stream handlers
- Handle connection failures gracefully

**DON'T:**
- Keep streams open indefinitely
- Send large payloads in single messages
- Ignore backpressure

### 4. Performance

**DO:**
- Use batch requests for multiple operations
- Implement connection pooling for clients
- Cache frequently accessed data
- Monitor message sizes

**DON'T:**
- Create new connections per request
- Send unnecessary notifications
- Block stream handlers

### Supporting Multiple Transports

Expose the same service over multiple protocols:

```go
Service("universal", func() {
    // JSON-RPC configuration
    JSONRPC(func() {
        POST("/rpc")
    })
    
    Method("process", func() {
        Payload(func() {
            Attribute("data", String)
            Required("data")
        })
        Result(func() {
            Attribute("output", String)
            Required("output")
        })
        
        // Available via JSON-RPC
        JSONRPC(func() {})
        
        // Also available via HTTP REST
        HTTP(func() {
            POST("/process")
        })
        
        // And via gRPC
        GRPC(func() {})
    })
})
```

## Additional Resources

- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [Goa Documentation](https://goa.design)
- [Example Services](https://github.com/goadesign/examples)
- [Integration Tests](../jsonrpc/integration_tests)

## Summary

Goa's JSON-RPC implementation provides:

- **Type Safety**: Full compile-time type checking
- **Code Generation**: Automatic client/server code from DSL
- **Protocol Compliance**: Complete JSON-RPC 2.0 support
- **Transport Flexibility**: HTTP, SSE, and WebSocket options
- **Streaming Support**: Unidirectional and bidirectional patterns
- **Error Handling**: Comprehensive error mapping and codes
- **Content Negotiation**: Mixed results based on Accept headers
- **Batch Processing**: Automatic batch request handling

The implementation seamlessly integrates with Goa's existing features while
maintaining clean separation of concerns and enabling powerful real-time
communication patterns.