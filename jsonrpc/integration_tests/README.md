# Goa JSON-RPC Integration Test Framework

A clean, data-driven integration test framework for testing Goa's JSON-RPC implementations.

This framework is designed to be simple and extensible. All test cases are defined in a single `YAML` file, allowing you to add new tests without writing any Go code. The core principle is **client-side testing**: every test is written from the perspective of a client sending a request and expecting a specific response.

## 🚀 Quick Start

### Running Existing Tests

To run all integration tests, navigate to the `integration_tests` directory and use the standard `go test` command.

```bash
# Run all tests in parallel with verbose output
go test -v ./...

# Run a single test by its name from the YAML file
# The format is TestJSONRPC/<scenario_name>
go test -v -run "TestJSONRPC/echo_string_request" ./...

# Filter which tests to run using a regex pattern
FILTER="^echo_.*" go test -v ./...
```

The `FILTER` environment variable is useful for running a specific group of tests (like all `echo` tests) without typing each full name. It matches the regular expression against the `name` field in your `scenarios.yaml` file.

### Adding a New Test

Adding a new test requires only a small addition to the scenarios file; no Go code is needed.

1.  **Open `scenarios/scenarios.yaml`**.

2.  **Add your test case**. For example, to test a method that echoes a map payload:

    ```yaml
    - name: "echo_map_request"
      method: "echo_map"
      transport: "http"
      request:
        id: "map-req-1"
        params:
          key1: "value1"
          key2: 42
      expect:
        id: "map-req-1"
        result:
          key1: "value1"
          key2: 42
    ```

3.  **Run the tests**. The framework will automatically handle the rest.

When you run the test, the framework sees the `method: "echo_map"`. Based on the `echo` action in the name, it dynamically generates a server method that simply returns its input parameters. This is why the `expect.result` in the example is identical to the `request.params`.

## ✨ How It Works

The framework's power comes from dynamically generating a complete Goa service tailored to the tests you define. This ensures we are testing against real, compiled Goa code, not mocks.

The execution flow for `go test` is:

1.  **Scenario Discovery**: The test runner parses `scenarios.yaml`, collects all test cases, and compiles a unique list of all `method` names used (e.g., `echo_string`, `transform_object`). This list informs the next step.

2.  **Dynamic Code Generation**: A temporary directory is created to house a complete Goa service.

      * A `design/design.go` file is generated containing a Goa DSL design for all discovered methods.
      * The framework runs `goa gen` and `goa example` to scaffold the service, server, and client code.
      * Crucially, it **injects service implementations** with predictable behavior based on their names. For example, a method named `transform_string` will be implemented to uppercase its string input. This removes the need for any manual service implementation.

3.  **Server Startup**: The generated Goa server is compiled and started on a random, available port. The test runner waits until the server is responsive before proceeding.

4.  **Test Execution**: For each scenario, the framework sends the defined `request` over the specified `transport`. It prioritizes using the **Goa-generated CLI** for this task to ensure tests closely mimic a real client's behavior. When a test requires sending a payload that the standard CLI cannot produce (e.g., a malformed request, an invalid JSON-RPC structure, or specific protocol-level edge cases), it falls back to a **custom JSON-RPC client** that allows for this fine-grained control. The client then asserts that the response from the server exactly matches the `expect` block in the scenario.

5.  **Cleanup**: Once all tests are complete, the server is shut down, and the temporary directory with all generated code is removed. To inspect the code, you can prevent this step (see **Debugging**).

## 🔧 Writing Test Scenarios

All tests live in `scenarios/scenarios.yaml`. Each scenario defines a single client-server interaction or a sequence of interactions. For a complete reference of all YAML fields and structures, see the **[YAML Schema Reference](https://www.google.com/search?q=SCHEMA.md)**.

### Basic Scenario Structure

Each scenario defines a single request-response cycle. The `request` block describes the JSON-RPC payload the client sends, and the `expect` block describes the exact payload the client must receive back for the test to pass. The `method` field links this scenario to the corresponding generated server method.

```yaml
- name: "unique_test_case_name" # A descriptive name for the test. Used with `go test -run`.
  method: "action_type_modifier" # Maps to a generated server method. See naming convention.
  transport: "http"             # 'http', 'websocket', or 'sse'.
  request:
    id: "req-1"                 # JSON-RPC request ID. Omit for notifications.
    params: ["hello"]           # The parameters for the method call.
  expect:
    id: "req-1"                 # The expected ID in the response.
    result: "HELLO"             # The expected result payload.
```

### Streaming Scenario Structure

For stateful protocols like WebSockets and Server-Sent Events (SSE), where multiple messages can be exchanged over a single connection, the `sequence` block is used. It defines an ordered list of actions the test client will perform.

```yaml
- name: "websocket_stream_and_collect"
  method: "collect_string"
  transport: "websocket"
  sequence:
    - type: "send" # Client sends a message to the server.
      data:
        id: "ws-req-1"
        params: ["one", "two", "three"]
    - type: "receive" # Client waits to receive a message from the server.
      expect:
        id: "ws-req-1"
        result: "onetwothree"
    - type: "close" # Client closes the connection.
```

## 📜 Method Naming Convention

Server behavior is determined entirely by the method name, which follows the pattern: `[action]_[type]_[modifier]`.

#### Actions

  * `echo`: Returns the `params` payload exactly as it was received.
  * `transform`: Returns a predictably modified version of the `params`.
  * `generate`: Ignores `params` and returns a fixed, predictable value.
  * `stream`: (SSE/WebSocket) Sends a stream of messages to the client. Ideal for testing server-streaming RPC.
  * `collect`: (WebSocket) Receives a stream of messages from a client and returns a single summary response after the stream is closed. Useful for testing client-streaming RPC.
  * `broadcast`: (WebSocket) Tests the server's ability to send unsolicited messages to a client (server-initiated notifications).

#### Types

  * `string`, `int`, `bool`
  * `array`: An array of simple types.
  * `object`: A structured JSON object.
  * `map`: A key-value map.
  * `user`: A Goa user-defined type with built-in validations.

#### Modifiers (Optional)

  * `_notify`: Indicates a JSON-RPC notification (no response expected).
  * `_error`: The method is hardcoded to always return a predefined JSON-RPC error.
  * `_validate`: The method includes Goa validation logic on the payload, which will return an error if the payload is invalid.
  * `_final`: (SSE) The method sends several notifications before sending a final, ID-tagged response.

## 🔬 Debugging

When a test fails, you can use the following tools to diagnose the issue:

  * **Keep Generated Code**: To inspect the dynamically generated Goa service, set the `KEEP_GENERATED` environment variable. The path to the generated code will be printed in the test logs.

    ```bash
    KEEP_GENERATED=true go test -v ./...
    # Look for: "Generated code kept in: /tmp/jsonrpc-test-XXXXX"
    ```

    Once you have the code, inspect `design/design.go` to see how your method was translated into Goa DSL and `http/server/server.go` to see its actual Go implementation.

  * **Verbose Output**: Always use the `-v` flag. It provides detailed logs showing the exact JSON payloads being sent and received, which is invaluable for debugging discrepancies between your `expect` block and the actual server response.

## 🏗️ Framework Architecture

The framework is organized into several packages, each with a clear responsibility.

```
integration_tests/
├── README.md
├── SCHEMA.md       # Detailed reference for the scenarios.yaml file structure.
├── go.mod
├── framework/      # The core engine: discovers scenarios, orchestrates code generation, and runs tests.
├── harness/        # The "physical" parts: a transport-aware client and server management code.
├── scenarios/      # The data-driven heart of the framework. This is where you'll spend most of your time.
└── tests/          # The Go test entrypoint (*_test.go) that kicks off the framework runner.
```

# YAML Schema Reference

This document provides a detailed reference for the structure and validation rules of the `scenarios.yaml` file used for JSON-RPC integration testing.

## 📂 Top-Level Structure

The `scenarios.yaml` file has two top-level keys: `scenarios` and `settings`.

```yaml
scenarios:
  - # A list of Scenario Objects, defined below.
  - # ...

settings:
  # A map of global settings for the test run.
```

## 📝 The `Scenario` Object

Each item in the top-level `scenarios` list is a `Scenario` object. It defines a single, self-contained test case.

| Key | Type | Required? | Description |
| :--- | :--- | :--- | :--- |
| **`name`** | `string` | **Yes** | A unique, human-readable name for the test. Used in test runner output. |
| **`method`** | `string` | **Yes** | The name of the server method to test. Must follow the `action_type_modifier` convention. |
| **`transport`** | `string` | **Yes** | The transport protocol. Must be one of `"http"`, `"websocket"`, or `"sse"`. |
| `request` | `object` | Conditional | An object describing the request to send. **Required** for non-streaming (`http`) tests. |
| `expect` | `object` | Conditional | An object describing the expected response. **Required** for non-streaming (`http`) tests. |
| `sequence` | `list` | Conditional | A list of steps for stateful interactions. **Required** for streaming (`websocket`, `sse`) tests. |

> A `Scenario` object must contain **either** a `request`/`expect` pair **or** a `sequence`, but not both.

### The `request` Object

Describes a single JSON-RPC request.

| Key | Type | Description |
| :--- | :--- | :--- |
| `id` | `any` | The JSON-RPC request ID. If omitted, the request is treated as a notification. |
| `params` | `array` or `object` | The parameters for the method call. See **Parameter Structures** below. |
| `method` | `string` | An optional override for the JSON-RPC `method` field in the payload. Defaults to the scenario's `method` value. |

### The `expect` Object

Describes the expected outcome of a request.

| Key | Type | Description |
| :--- | :--- | :--- |
| `id` | `any` | The expected ID in the response. Must match the `request.id`.
| `result` | `any` | The expected `result` payload. Mutually exclusive with `error`. |
| `error` | `object` | The expected `error` object. Mutually exclusive with `result`. Contains `code` (number), `message` (string), and optional `data` (any). |
| `no_response` | `boolean` | If `true`, the framework asserts that no response is received. Used for notifications. |

### The `Sequence Step` Object

Each item in a `sequence` list is a step object that defines a single action in a streaming test.

| Key | Type | Required? | Description |
| :--- | :--- | :--- | :--- |
| **`type`** | `string` | **Yes** | The type of action. Must be one of `"send"`, `"receive"`, or `"close"`. |
| `data` | `object` | Conditional | The JSON-RPC payload to send. **Required** for `type: "send"`. |
| `expect`| `object` | Conditional | The expected JSON-RPC payload to receive. **Required** for `type: "receive"`. |
| `delay` | `string` | No | A duration to wait before executing this step (e.g., `"100ms"`, `"1s"`). |

## 🧩 Parameter Structures (`params`)

The `params` field must follow one of the two standard JSON-RPC 2.0 formats.

1.  **By-Position (`array`)**: Parameters are passed as a list of values.

    ```yaml
    # The server method receives these values in the specified order.
    params: ["first_param", "second_param", 42]
    ```

2.  **By-Name (`object`)**: Parameters are passed as a map of key-value pairs.

    ```yaml
    # The server method receives these values by their key name.
    params:
      name: "example"
      value: 123
      is_active: true
    ```

## ✅ Semantic and Validation Rules

In addition to the structure, the content of the YAML file must adhere to these rules.

  * **Uniqueness**: The `name` of each scenario must be unique within the file.
  * **Exclusivity**: A scenario cannot have both `request`/`expect` and `sequence` defined.
  * **ID Matching**: If a `request.id` is present, the corresponding `expect.id` must be identical.
  * **Result vs. Error**: An `expect` object cannot define both a `result` and an `error`.
  * **Method Convention**: The `method` field must follow the `[action]_[type]_[modifier]` pattern, which determines the generated server's behavior.

## 🌐 Complete Examples

### HTTP Request with an Error Response

```yaml
scenarios:
  - name: "generate_object_error_http"
    method: "generate_object_error" # This method is designed to always fail
    transport: "http"
    request:
      id: "err-req-01"
      params: {} # Params can be empty
    expect:
      id: "err-req-01"
      error:
        code: -32000
        message: "A simulated server error occurred"
```

### WebSocket Bidirectional Streaming

This example shows a client subscribing to a channel and then receiving a server-initiated broadcast.

```yaml
scenarios:
  - name: "broadcast_websocket_interaction"
    method: "broadcast_string"
    transport: "websocket"
    sequence:
      # 1. Client sends a subscription request
      - type: "send"
        data:
          jsonrpc: "2.0"
          method: "broadcast_string" # Method to call on the server
          params: { "channel": "news" }
          id: "sub-1"

      # 2. Client expects a confirmation response
      - type: "receive"
        expect:
          jsonrpc: "2.0"
          id: "sub-1"
          result: { "status": "subscribed", "channel": "news" }

      # 3. Client waits to receive an unsolicited broadcast from the server
      - type: "receive"
        expect:
          jsonrpc: "2.0"
          method: "broadcast" # Note: This is a server-initiated method, not a response
          params: { "message": "Server update!" }
```


Review each file one by one and each function one by one and think of ways it can be streamlined, improved, simplify, made more tuitive and follow Go best practice. 