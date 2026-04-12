# How are default values evaluated in protocol buffers?

Non-nil default values are not supported in protocol buffers
(see https://developers.google.com/protocol-buffers/docs/proto3#default).
Hence, there is no way to figure out whether a field was explicitly set to
the default value or just not set at all. So goa does not initialize such
fields with their default values.

# How goa deals with nested maps and arrays in protocol buffers?

proto3 syntax for protocol buffer does not support nested maps and arrays
(see https://github.com/protocolbuffers/protobuf/issues/4596). In such cases,
goa wraps the inner map/array into a user type having a single attribute named
"field" with RPC tag number 1.

Example:

Type definition
```
Type("MyType", func() {
  Field(3, "nested", MapOf(Int, MapOf(String, ArrayOf(Bool))))
})
```
is transformed into protocol buffer message below
```
message MyType {
  map<int32, MapOfStringArrayOfBool> nested = 3;
}

message MapOfStringArrayOfBool {
  map<string, ArrayOfBool> field = 1;
}

message ArrayOfBool {
  repeated bool field = 1;
}
```
for which protoc generates the following types
```
type MyType struct {
  Nested map[int32]*MapOfStringArrayOfBool
}

type MapOfStringArrayOfBool struct {
  Field map[string]*ArrayOfBool
}

type ArrayOfBool struct {
  Field []bool
}
```

# How does goa encode methods that define both Payload and StreamingPayload?

For gRPC, goa keeps ordinary method payload in the typed message channel.
When a method defines both `Payload(...)` and `StreamingPayload(...)`, goa
generates a streamed request envelope with two variants:

- `initial_payload` carries the one-shot method payload once at stream setup.
- `stream_item` carries each `StreamingPayload` value after that.

This keeps gRPC metadata reserved for explicit `GRPC.Metadata(...)` fields and
security attributes. It also means rich payload types such as objects, maps,
and unions are encoded with the normal protobuf message machinery instead of
being stringified into headers.

# How does goa handle the Any type in gRPC?

Goa supports the `Any` type in gRPC by mapping it to `google.protobuf.Value`, which is specifically designed to represent dynamic JSON-like values. This is simpler and more efficient than using `google.protobuf.Any`.

## Conversion Process

- **Go to Protobuf**: When converting from Go `any` to `*structpb.Value`, Goa uses `structpb.NewValue()` which directly converts Go types to protobuf Value.
- **Protobuf to Go**: When converting from `*structpb.Value` to Go `any`, Goa uses the `AsInterface()` method which returns the corresponding Go value.

## Example Usage

In your Goa design:
```go
Method("echo", func() {
    Payload(func() {
        Field(1, "data", Any, "Any type of data")
    })
    Result(func() {
        Field(1, "data", Any, "Echoed data")
    })
    GRPC(func() {
        Response(CodeOK)
    })
})
```

This generates the following protobuf:
```proto
import "google/protobuf/struct.proto";

message EchoRequest {
    optional google.protobuf.Value data = 1;
}

message EchoResponse {
    optional google.protobuf.Value data = 1;
}
```

## Supported Patterns

- Direct Any fields: `Field(1, "data", Any)`
- Maps with Any values: `MapOf(String, Any)`
- Arrays of Any: `ArrayOf(Any)`
- Nested structures containing Any

## Supported Value Types

The `google.protobuf.Value` type natively supports:
- Null values
- Numbers (integer and floating point)
- Strings
- Booleans
- Structs (maps)
- Lists (arrays)

## Limitations

- Complex Go types (channels, functions, custom structs) need to be JSON-serializable
- Type information is abstracted to basic JSON types
- Precision may be lost for very large integers (uses float64 internally)
