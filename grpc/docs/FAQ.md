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

# How does goa handle the Any type in gRPC?

Goa supports the `Any` type in gRPC by mapping it to `google.protobuf.Any`. The conversion between Go's `any` type and protobuf's `Any` is done using JSON marshaling/unmarshaling.

## Conversion Process

- **Go to Protobuf**: When converting from Go `any` to `*anypb.Any`, the value is JSON-marshaled and wrapped in a `structpb.Value`.
- **Protobuf to Go**: When converting from `*anypb.Any` to Go `any`, the value is unwrapped and JSON-unmarshaled.

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
import "google/protobuf/any.proto";

message EchoRequest {
    optional google.protobuf.Any data = 1;
}

message EchoResponse {
    optional google.protobuf.Any data = 1;
}
```

## Supported Patterns

- Direct Any fields: `Field(1, "data", Any)`
- Maps with Any values: `MapOf(String, Any)`
- Arrays of Any: `ArrayOf(Any)`
- Nested structures containing Any

## Limitations

- The JSON conversion means that complex Go types may not roundtrip perfectly
- Only JSON-serializable values are supported
- Type information is lost during conversion
