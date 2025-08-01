# Goa JSON-RPC Architecture

This document explains the architecture of Goa's JSON-RPC support, covering both basic HTTP and advanced WebSocket-based streaming communication. It details the code generation process, runtime behavior, and recommended usage patterns.

## Core Principle: Composition Over Modification

The fundamental principle behind Goa's JSON-RPC implementation is **composition over modification**. Instead of altering shared HTTP templates to accommodate JSON-RPC, the JSON-RPC code generation layer builds upon the existing HTTP transport infrastructure. This approach ensures a clean separation of concerns, preventing the HTTP layer from becoming coupled to JSON-RPC specifics and allowing both to evolve independently.

## Code Generation

The generation of JSON-RPC enabled services follows a layered process that starts with the standard HTTP transport code.

### HTTP Codegen Foundation

The process begins by generating the transport-agnostic service code, which includes:

*   Service interfaces and endpoints
*   Basic HTTP handlers and middleware
*   Encoding and decoding utilities
*   Error handling infrastructure

### JSON-RPC Composition Layer

The JSON-RPC `codegen` package then composes on top of the generated HTTP code by programmatically manipulating the `codegen.File` data structure before it is rendered. This involves a three-step process:

1.  **Generate Base HTTP Code**: The standard `httpcodegen.ServerEncodeDecodeFile` function is called to produce the initial set of files.
2.  **Modify Sections**: The generated sections are iterated upon to introduce JSON-RPC specific behavior. This includes adding necessary imports and replacing HTTP handler signatures with their JSON-RPC counterparts.
3.  **Add JSON-RPC Sections**: Finally, new sections containing JSON-RPC specific logic, such as server handler initializers, are appended.

This process is exemplified by the following snippet:

```go
// Step 1: Generate base HTTP code
f := httpcodegen.ServerEncodeDecodeFile(genpkg, svc, data)

// Step 2: Modify sections before final code generation
for _, s := range f.SectionTemplates {
    // Add JSON-RPC imports
    if s.Name == "source-header" {
        codegen.AddImport(s, codegen.GoaImport("jsonrpc"))
    }
    
    // Modify signatures for JSON-RPC context
    if s.Name == "request-decoder" {
        s.Source = strings.Replace(s.Source, 
            httpRequestDecoderTemplate, 
            jsonrpcRequestDecoderTemplate, 1)
    }
    
    // Namespace sections to avoid conflicts
    s.Name = "jsonrpc-" + s.Name
}

// Step 3: Add JSON-RPC specific sections
sections = append(sections,
    &codegen.SectionTemplate{
        Name: "jsonrpc-server-handler-init", 
        Source: jsonrpcTemplates.Read(serverHandlerInitT), 
        Data: e
    })
```

### Key Codegen Patterns

Three key patterns enable this compositional approach:

1.  **Template Namespacing**: JSON-RPC sections are prefixed with `jsonrpc-` to prevent name collisions with HTTP sections.
2.  **In-Memory Modification**: Instead of altering the source templates on disk, modifications are made to the `Source` field of the `codegen.SectionTemplate` struct in memory.
3.  **Conditional Template Selection**: The code generation logic dynamically selects the appropriate templates based on the endpoint configuration, for example, adding WebSocket-specific templates only when a WebSocket transport is defined for the service.

### Template Responsibilities

This layered approach results in a clean separation of responsibilities between HTTP and JSON-RPC templates:

*   **HTTP Templates (Shared)**: These are responsible for transport-agnostic service logic. They must not contain any JSON-RPC or WebSocket specific logic.
*   **JSON-RPC Templates (Specialized)**: These handle JSON-RPC protocol specifics and WebSocket streaming. They can specialize HTTP behavior but should do so through composition, not modification of the HTTP templates.

## Runtime Architecture and Usage

The generated code provides two primary mechanisms for JSON-RPC communication: a simple HTTP transport for traditional request-response interactions, and a WebSocket transport for real-time, bidirectional streaming.

### Standard JSON-RPC over HTTP

For services that do not require streaming, JSON-RPC messages are exchanged over standard HTTP. The generated server code includes an HTTP handler that decodes the JSON-RPC request from the HTTP body, invokes the corresponding service method, and writes the JSON-RPC response back to the HTTP response writer.

The handler signatures clearly illustrate the differences between the transport layers:

*   **Regular HTTP**: `func(context.Context, *http.Request, http.ResponseWriter)`
*   **JSON-RPC HTTP**: `func(context.Context, *http.Request, *jsonrpc.RawRequest, http.ResponseWriter)`

### JSON-RPC over WebSocket

Goa provides a powerful abstraction for building streaming services with WebSockets. This allows for full-duplex communication channels that can support a variety of interaction patterns.

#### Architectural Principles

The WebSocket architecture is guided by three principles:

1.  **Single WebSocket Connection**: A single WebSocket connection is used to handle all JSON-RPC communication for a given service, including multiplexing different method calls and streaming patterns.
2.  **User Code Owns Streaming Logic**: The core streaming logic is implemented by the developer in the `HandleStream` method. Goa provides the infrastructure and the `Stream` interface, but the implementation of the streaming strategy is left to the user.
3.  **Clean Separation of Concerns**: The architecture separates the business logic (in service methods), the transport layer (JSON-RPC protocol and WebSocket management), and the streaming logic (in `HandleStream`).

#### Core Components

The WebSocket support is built around three core components:

*   **`HandleStream` Method**: This method is the entry point for all WebSocket communication. It is where the developer implements the application-specific streaming logic.

    ```go
    func (s *serviceImpl) HandleStream(ctx context.Context, stream ServiceName.Stream) error {
        // User implements their streaming strategy here.
        // Can listen to channels, timers, events, etc.
        // Can call stream.Recv() to process incoming JSON-RPC requests.
        // Can call stream.SendMethodName() to send responses or notifications.
    }
    ```

*   **`Stream` Interface**: This generated interface provides the methods for interacting with the WebSocket connection, including receiving requests (`Recv`), sending responses (`SendMethodName`), sending errors (`SendError`), and closing the connection (`Close`).

*   **Service Methods**: These are the regular service methods with standard Go signatures. They are called automatically when `Recv()` processes a matching JSON-RPC request and can also be called directly from `HandleStream` for server-initiated communication.

The handler signature for WebSocket streaming endpoints reflects the asynchronous nature of the communication:

```go
func(context.Context, *http.Request, *jsonrpc.RawRequest) (any, error)
```

The handler returns a result and an error because responses are sent asynchronously via the `Stream` interface rather than being written directly to an `http.ResponseWriter`.

#### Streaming Patterns

The flexibility of the `HandleStream` method allows for a variety of streaming patterns:

*   **Request-Response**: The traditional JSON-RPC pattern can be implemented by simply calling `stream.Recv()` in a loop. When `Recv()` is called, it reads a JSON-RPC request from the WebSocket, dispatches it to the appropriate service method, and automatically sends the response back.

    ```go
    func (s *serviceImpl) HandleStream(ctx context.Context, stream ServiceName.Stream) error {
        defer stream.Close()
        
        for {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                if err := stream.Recv(ctx); err != nil {
                    return err
                }
            }
        }
    }
    ```

*   **Server Streaming**: To push data from the server to the client, the `HandleStream` method can initiate a goroutine that sends data at regular intervals or in response to events.

*   **Client Streaming**: To receive a stream of data from a client, the `HandleStream` method can repeatedly call `stream.Recv()` and accumulate the results.

*   **Bidirectional Streaming**: For interactive communication, `HandleStream` can combine both server and client streaming patterns, for example by launching a goroutine to handle outgoing messages while the main loop processes incoming messages.

#### Advanced Patterns

The `HandleStream` method can also be used to implement more advanced patterns, such as:

*   **Mixed Request-Response and Streaming**: A service can handle both traditional request-response interactions and asynchronous, server-initiated notifications within the same WebSocket connection.
*   **Conditional Streaming**: The streaming strategy can be determined dynamically based on the properties of the connection or the initial messages exchanged.

#### Method Dispatch and Results

When `stream.Recv()` is called, it automatically handles the parsing of the incoming JSON-RPC request, validation, routing to the appropriate service method, and marshalling of the response. Service methods can also be invoked manually from within `HandleStream` for server-initiated communication.

#### Error Handling

The architecture provides mechanisms for handling various types of errors:

*   **Connection Errors**: Errors at the WebSocket connection level will cause `HandleStream` to terminate.
*   **JSON-RPC Protocol Errors**: Invalid requests will result in the automatic sending of a JSON-RPC error response.
*   **Streaming Errors**: Errors that occur while sending or receiving data can be handled within the `HandleStream` implementation.

#### Testing Strategies

The separation of concerns in the architecture simplifies testing:

*   **Integration Tests**: The `HandleStream` implementation can be overridden in tests to simulate specific streaming behaviors.
*   **Unit Tests**: Service methods can be tested independently as standard Go functions.

## Maintenance Guidelines

To maintain the clean separation of concerns and the long-term health of the codebase, it is important to adhere to the following guidelines:

### DO:

*   ✅ Modify JSON-RPC templates for JSON-RPC specific behavior.
*   ✅ Use `codegen.File` section manipulation for signature changes.
*   ✅ Add JSON-RPC specific sections for specialized functionality.
*   ✅ Compose on top of HTTP-generated code.

### DON'T:

*   ❌ Modify HTTP templates with JSON-RPC specific logic.
*   ❌ Add WebSocket conditionals to shared HTTP templates.
*   ❌ Break the transport independence of the HTTP layer.
*   ❌ Couple the HTTP codegen to JSON-RPC requirements.
