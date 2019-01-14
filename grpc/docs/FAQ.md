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
