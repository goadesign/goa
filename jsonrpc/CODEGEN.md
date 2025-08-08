# JSON-RPC Code Generation in Goa

This document explains how JSON-RPC code generation is organized in Goa:
what inputs it consumes, what it produces, and how the main pieces fit
together. The focus is on principles, structures, and flows rather than
line-by-line algorithms.

## Design Principles

- Composition over modification: build JSON-RPC on top of HTTP codegen by
  transforming sections in memory and appending JSON-RPC templates instead of
  forking shared HTTP templates.
- Transport isolation: keep SSE and WebSocket concerns in their own template
  stacks. Avoid leaking transport conditionals into shared HTTP templates.
- Single source of truth: reuse HTTP encode/decode, validation, and service
  data so generated code remains consistent across transports.
- Clear separation of responsibilities: design/expr → service data → codegen
  templates → transport handlers; keep each layer narrow and testable.

## Inputs and Outputs

Inputs

- API/design model (`expr.RootExpr`): services, methods, HTTP endpoints,
  JSON-RPC enablement, SSE annotations, and WebSocket presence.
- Per-service data (`httpcodegen.ServicesData` and `codegen/service`): derived
  structures used to render templates and stream interfaces.

Outputs (per JSON-RPC service)

- A server composition file under `jsonrpc/<service>/server/` that wires
  handlers and mounting.
- A transformed copy of the HTTP encode/decode file relocated under `jsonrpc/`
  with JSON-RPC decoding semantics while reusing the HTTP decoder/validator.
- Exactly one transport stack:
  - HTTP only (request/response), or
  - SSE server streaming, or
  - WebSocket streaming (client/server/bidi), or
  - Mixed HTTP + SSE (selected via `Accept`).

## High-Level Pipeline

1. Service data analysis
   - The DSL is analyzed into service/method metadata, including whether a
     method is JSON-RPC, whether it streams, and if SSE or WebSocket is used.
   - `codegen/service` computes stream interfaces (names, kinds, and methods)
     and documents invariants (e.g., when `SendAndClose` exists for SSE).

2. HTTP encode/decode composition
   - Goa generates the standard HTTP encode/decode file for the service.
   - JSON-RPC codegen adjusts the request decoding to read JSON-RPC envelopes
     while keeping the rest of the machinery (type transforms and validation).
   - The file is namespaced and moved from `/http/` to `/jsonrpc/` to avoid
     collisions and to make intent explicit.

3. Transport selection and handler mounting
   - A small set of predicates decides which handler templates to include:
     - HTTP only
     - SSE only
     - WebSocket present
     - Mixed (HTTP + SSE)
   - The server composition file mounts exactly one transport path for JSON-RPC
     requests, or a mixed handler that switches by `Accept`.

4. Service interface and stream APIs
   - The generated service interface exposes standard methods plus, when
     applicable, a WebSocket `HandleStream` entry point.
   - Stream interfaces differ per transport:
     - WebSocket `Stream`: per-method notification/response senders, `Recv`,
       `Close`, and `SendError`.
     - SSE per-method server stream: `Send` (notifications), `SendAndClose`
       (final response with id), and `SendError`.

5. Error model
   - Named errors declared in the DSL are mapped to JSON-RPC error codes by the
     send helpers. Validation/service errors default to `InvalidParams`; others
     default to `InternalError`.

## Transport Flows (Conceptual)

### HTTP (request/response)
- Single endpoint receives a JSON-RPC envelope, dispatches by `method`, uses
  the transformed HTTP decoder to build typed payloads, and returns a JSON-RPC
  success or error envelope. Batch requests are processed independently with
  responses streamed into a JSON array.

### SSE (server streaming)
- The initial JSON-RPC request establishes a long-lived SSE response. The
  per-method server stream can:
  - `Send` JSON-RPC notifications (no id).
  - `SendAndClose` a final JSON-RPC success with id (copied from request or
    taken from a result ID field when present) and close the stream.
  - `SendError` a JSON-RPC error envelope.
- Mixed HTTP + SSE switches on `Accept` (`application/json` vs
  `text/event-stream`).

### WebSocket (streaming)
- The upgrade produces a generated `Stream`. User code implements
  `HandleStream(ctx, stream)` and typically loops on `stream.Recv(ctx)`, which
  decodes JSON-RPC requests and dispatches to generated handlers.
- Method-specific wrappers provide `SendNotification`, `SendResponse`, and
  `SendError` with request id correlation handled for you.

## Template Map (What Generates What)

- Server composition and mounting: `jsonrpc/codegen/server.go` + templates
  under `jsonrpc/codegen/templates/*`.
- HTTP handler (single/batch) and JSON-RPC envelope handling:
  `server_handler.go.tpl`.
- SSE stack: `sse_server_handler.go.tpl`, `sse_server_stream.go.tpl`.
- WebSocket stack: `websocket_server_handler.go.tpl`,
  `websocket_server_stream.go.tpl`, `websocket_server_recv.go.tpl`,
  `websocket_server_send.go.tpl`, `websocket_server_stream_wrapper.go.tpl`.
- Service interface and stream APIs: `codegen/service/templates/service.go.tpl`
  (with JSON-RPC specializations derived from method metadata).

## Extension Points and Invariants

- Keep shared HTTP templates transport-agnostic; JSON-RPC must compose on top.
- The JSON-RPC transform should not change validation or type transforms.
- Each service mounts only one JSON-RPC transport at a time; mixed HTTP + SSE
  is implemented via content negotiation at runtime.
- SSE semantics: `Send` for notifications, `SendAndClose` for the final
  response. WebSocket provides explicit per-method `Send*` helpers.

## Testing Strategy (Concise)

- Integration tests under `jsonrpc/integration_tests` cover HTTP, SSE, and
  WebSocket, including notifications, errors, and mixed flows.
- Golden tests validate that rendered files remain stable across changes.

## Summary

The JSON-RPC code generator reuses HTTP encode/decode and validation, adds a
thin JSON-RPC envelope layer, and selects a transport stack (HTTP, SSE,
WebSocket, or mixed HTTP+SSE). Stream APIs encapsulate JSON-RPC id handling
and error mapping so service implementations can focus on business logic.