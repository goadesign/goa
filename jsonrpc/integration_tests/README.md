# Goa JSON-RPC Integration Test Framework

A clean, data-driven integration test framework for testing Goa's JSON-RPC implementations.

This framework is designed to be simple and extensible. All test cases are defined in a single `YAML` file, allowing you to add new tests without writing any Go code. The core principle is **client-side testing**: every test is written from the perspective of a client sending a request and expecting a specific response.

**📍 File Location**: All tests are defined in `integration_tests/scenarios/scenarios.yaml`

## 🚀 Quick Start

### Running Existing Tests

To run all integration tests, navigate to the `integration_tests` directory and use the standard `go test` command.

```bash
# Run all tests in parallel with verbose output (bypasses test cache)
go test -count=1 -v ./...

# Run a single test by its name from the YAML file
# The format is TestJSONRPC/<scenario_name>
go test -count=1 -v -run "TestJSONRPC/echo_string_request" ./...

# Filter which tests to run using a regex pattern
FILTER="^echo_.*" go test -count=1 -v ./...
```

The `FILTER` environment variable is useful for running a specific group of tests (like all `echo` tests) without typing each full name. It matches the regular expression against the `name` field in your `scenarios.yaml` file.

### Adding a New Test - Three Complete Examples

Adding a new test requires only a small addition to the scenarios file; no Go code is needed.

1. **Open `integration_tests/scenarios/scenarios.yaml`** (from the project root)

2. **Add your test case** at the end of the `scenarios:` list

3. **Run the tests** - the framework automatically handles the rest

#### Example 1: Simple HTTP Test
```yaml
# Add this to scenarios.yaml
- name: "my_echo_test"
  method: "echo_string"  # action_type format
  transport: "http"
  request:
    params: "hello world"  # What to send
    id: 123                # Required unless the method name ends in _notify
  expect:
    result: "hello world"  # Echo returns the same value
    id: 123                # Response has same ID
```

#### Example 2: SSE Streaming Test
```yaml
# SSE uses request to initiate, then sequence for the stream
- name: "my_stream_test"
  method: "stream_string_sse"
  transport: "sse"
  request:
    params: "hi"           # 2 characters = 2 notifications
    id: "sse-1"
  sequence:                 # What we expect to receive
    - type: "receive"
      expect:
        jsonrpc: "2.0"
        method: "stream_string_sse"
        params:
          value: "Stream 1 of 2"
    - type: "receive"
      expect:
        jsonrpc: "2.0"
        method: "stream_string_sse"
        params:
          value: "Stream 2 of 2"
```

#### Example 3: Transform Test with Object
```yaml
# Objects have fixed field names: field1, field2, field3
- name: "my_transform_test"
  method: "transform_object"
  transport: "http"
  request:
    params:
      field1: "hello"      # Will be uppercased
      field2: 10           # Will be doubled
      field3: true         # Will be negated
    id: "transform-1"
  expect:
    result:
      field1: "HELLO"      # Uppercased
      field2: 20           # Doubled
      field3: false        # Negated
    id: "transform-1"
```

## 💡 Common Patterns and Pitfalls

### Arrays Need Wrapper Objects
Arrays aren't sent directly - they need an `items` wrapper:
```yaml
# ❌ Wrong
params: ["one", "two"]

# ✅ Correct
params:
  items: ["one", "two"]
```

### Object Fields Are Fixed
The `object` type always uses these exact field names:
- `field1` (string)
- `field2` (integer) 
- `field3` (boolean)

```yaml
params:
  field1: "text"    # Must be field1, not myField
  field2: 42         # Must be field2, not count
  field3: true       # Must be field3, not enabled
```

### Maps Use a Data Wrapper
Maps need a `data` field to hold the key-value pairs:
```yaml
params:
  data:
    any_key: "any_value"   # Keys are flexible
    another: 123
```

### SSE Always Uses Request + Sequence
SSE tests need both:
- `request`: Initiates the SSE connection
- `sequence`: Defines expected stream events

### Declared Notifications Have No ID
For a fire-and-forget method whose name ends in `_notify`:
```yaml
request:
  params: "notification"
  # A declared notification must omit the id field.
expect:
  no_response: true
```

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

All tests live in `integration_tests/scenarios/scenarios.yaml`. Each scenario defines a single client-server interaction or a sequence of interactions.

### Transport-Specific Patterns

#### HTTP Tests (Request → Response)
Use `request` and `expect` for single request-response:
```yaml
- name: "http_test"
  method: "echo_string"
  transport: "http"
  request:                  # What we send
    params: "hello"
    id: 1
  expect:                   # What we receive
    result: "hello"
    id: 1
```

#### SSE Tests (Request → Stream of Events)
Use `request` to initiate, `sequence` for the event stream:
```yaml
- name: "sse_test"
  method: "stream_string_sse"
  transport: "sse"
  request:                  # Initiates SSE connection
    params: "test"
    id: "sse-1"
  sequence:                 # Stream of events we expect
    - type: "receive"
      expect:
        method: "stream_string_sse"  # Notifications include method
        params:
          value: "Stream 1 of 4"
    # ... more events
```

## 📜 Method Naming Convention

Server behavior is determined entirely by the method name. Unary methods use
`[action]_[type]_[modifier]`; SSE methods add the `_sse` suffix.

### Quick Reference Table

| Action | Type | Input Example | Output Example |
|--------|------|--------------|----------------|
| **echo** | string | `"hello"` | `"hello"` |
| **echo** | array | `{items: ["a", "b"]}` | `{items: ["a", "b"]}` |
| **echo** | object | `{field1: "x", field2: 1, field3: true}` | Same as input |
| **echo** | map | `{data: {k: "v"}}` | `{data: {k: "v"}}` |
| **transform** | string | `"hello"` | `"HELLO"` |
| **transform** | array | `{items: ["a", "b", "c"]}` | `{items: ["c", "b", "a"]}` |
| **transform** | object | `{field1: "x", field2: 5, field3: true}` | `{field1: "X", field2: 10, field3: false}` |
| **transform** | map | `{data: {key: "val"}}` | `{data: {transformed_key: "val"}}` |
| **generate** | string | (ignored) | `"generated-string"` |
| **generate** | array | (ignored) | `{items: ["item1", "item2", "item3"]}` |
| **generate** | object | (ignored) | `{field1: "generated-value1", field2: 42, field3: true}` |
| **generate** | map | (ignored) | `{data: {generated: true, count: 3, status: "ok"}}` |

### Actions

  * `echo`: Returns the `params` payload exactly as it was received.
  * `transform`: Returns a predictably modified version of the `params`.
  * `generate`: Ignores `params` and returns a fixed, predictable value.
  * `stream`: Sends a stream of server-sent events to the client.

### Types and Their Structure

#### Primitive Types
  * `string`: Plain string value
  * `int`: Integer value  
  * `bool`: Boolean value

#### Structured Types
  * `array`: Must use `items` wrapper
    ```yaml
    params:
      items: ["one", "two", "three"]
    ```
  
  * `object`: Must use exactly these field names
    ```yaml
    params:
      field1: "string value"   # string
      field2: 42               # integer
      field3: true             # boolean
    ```
  
  * `map`: Must use `data` wrapper with flexible keys
    ```yaml
    params:
      data:
        any_key_name: "value"
        another_key: 123
    ```
  
  * `user`: A Goa user-defined type with built-in validations

### Modifiers (Optional)

  * `_notify`: Indicates a JSON-RPC notification (no response expected).
  * `_error`: The method is hardcoded to always return a predefined JSON-RPC error.
  * `_validate`: The method includes Goa validation logic on the payload, which will return an error if the payload is invalid.

## 📊 Data-Driven Behavior

The framework generates predictable server behavior based on the method name and **payload data**. This is crucial to understand when writing tests, especially for streaming scenarios.

### Action Behaviors

#### `echo` Action
Returns the payload exactly as received. For SSE, sends the payload as a notification.

```yaml
# Example: echo_string_sse
request:
  params: "hello world"
expect:
  params:
    value: "hello world"  # Exact echo
```

#### `transform` Action  
Applies predictable transformations to the payload:
- **string**: Converts to uppercase
- **array**: Reverses the order
- **object**: Uppercases field1, doubles field2, negates field3
- **map**: Prefixes all keys with "transformed_"

```yaml
# Example: transform_string_sse
request:
  params: "hello"
expect:
  params:
    value: "HELLO"  # Uppercase transformation
```

#### `generate` Action
Ignores the payload and returns fixed values. For SSE, always sends 3 generated notifications.

```yaml
# Example: generate_string_sse
request:
  params: "ignored"  # Payload is ignored
sequence:
  - expect:
      params:
        value: "generated-1"  # Fixed sequence
  - expect:
      params:
        value: "generated-2"
  - expect:
      params:
        value: "generated-3"
```

#### `stream` Action (SSE)
The payload data controls the streaming behavior:

**For `string` type:**
- Payload length determines the number of messages (max 10)
- Empty string or no payload: sends 3 messages by default

```yaml
# Example: 5 characters = 5 messages
request:
  params: "12345"  # Length 5
sequence:
  - expect:
      params:
        value: "Stream 1 of 5"
  # ... continues to "Stream 5 of 5"
```

**For `array` type:**
- Each array item generates one notification
- Empty array: sends single "empty" notification

```yaml
# Example: Each item is processed
request:
  params:
    items: ["first", "second"]
sequence:
  - expect:
      params:
        items: ["Processing: first"]
  - expect:
      params:
        items: ["Processing: second"]
```

**For `object` type:**
- `field2` value controls the number of notifications (max 10)
- Default is 3 if field2 is 0 or missing

```yaml
# Example: field2 controls count
request:
  params:
    field1: "test"
    field2: 2  # Will send 2 notifications
    field3: false
sequence:
  - expect:
      params:
        field1: "test-1"
        field2: 1
        field3: false
  - expect:
      params:
        field1: "test-2"
        field2: 2
        field3: true  # Last item is true
```

**For `map` type:**
- Each key-value pair generates one notification
- Empty map: sends single notification with `{"status": "empty"}`

```yaml
# Example: Each key-value becomes a notification
request:
  params:
    data:
      key1: "value1"
      key2: "value2"
sequence:
  - expect:
      params:
        data:
          key: "key1"
          value: "value1"
  - expect:
      params:
        data:
          key: "key2"
          value: "value2"
```

### Modifier Effects

**`_error` modifier:**
For streaming, sends notifications then returns an error:

```yaml
# Example: stream_string_error_sse  
sequence:
  - expect:  # Some notifications first
      params:
        value: "Stream 1 of 2"
  - expect:  # Then error with ID
      id: "req-1"
      error:
        code: -32602
        message: "Streaming error occurred"
```

### Writing Effective Tests

#### SSE Tests - Key Points
1. **Always use both `request` and `sequence`**: Request initiates the connection, sequence defines expected events
2. **Method names determine behavior**: Use the right action for your test case
3. **Payload data controls streaming**: The actual values in your request determine what gets streamed
4. **No payload = defaults**: Methods without payloads send 3 default messages
5. **Notifications vs responses**: Notifications have no ID, final responses include the request ID

#### Common Test Patterns
```yaml
# Testing an error condition
method: "echo_string_error"  # _error modifier
expect:
  error:
    code: -32602
    message: "Invalid params"

# Testing a notification (no response)
method: "echo_string_notify"  # _notify modifier
request:
  params: "fire and forget"
  # No id field
expect:
  no_response: true

# Testing validation
method: "echo_string_validate"  # _validate modifier
request:
  params:
    value: ""  # Empty string might fail validation
expect:
  error:
    code: -32602
    message: "validation error"
```

## 🔬 Debugging

When a test fails, you can use the following tools to diagnose the issue:

  * **Keep Generated Code**: To inspect the dynamically generated Goa service, set the `KEEP_GENERATED` environment variable. The path to the generated code will be printed in the test logs.

    ```bash
    KEEP_GENERATED=true go test -count=1 -v ./...
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
| **`method`** | `string` | **Yes** | The server method. SSE method names end with `_sse`. |
| **`transport`** | `string` | **Yes** | The transport protocol. Must be either `"http"` or `"sse"`. |
| `request` | `object` | Conditional | An object describing the request to send. **Required** for non-streaming (`http`) tests. |
| `expect` | `object` | Conditional | An object describing the expected response. **Required** for non-streaming (`http`) tests. |
| `sequence` | `list` | Conditional | The events expected from an `sse` stream. |

> A `Scenario` object must contain **either** a `request`/`expect` pair **or** a `sequence`, but not both.

### The `request` Object

Describes a single JSON-RPC request.

| Key | Type | Description |
| :--- | :--- | :--- |
| `id` | `any` | The JSON-RPC request ID. It is required and non-null unless the generated method is declared as a notification with the `_notify` suffix. Declared notifications must omit it. |
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
| **`type`** | `string` | **Yes** | The event action. SSE sequences use `"receive"`. |
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
  * **Method Convention**: Unary methods follow `[action]_[type]_[modifier]`; SSE methods end with `_sse`.

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
