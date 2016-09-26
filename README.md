# v2

This branch is work in progress for goa v2.

v2 brings a host of fixes and has a cleaner more composable overall design. Most notably the DSL
engine assumes less about the DSL and is thus more generic. The top level design package is also
hugely simplified to focus solely on types.

## gRPC Support

The new top level `rest` and `rpc` packages implement the DSL, design objects, code generation and
runtime support for REST and gRPC respectively. The DSLs build on top of the core DSL package to add
transport specific keywords such as request path information for HTTP.

## New Data Types

The primitive types now include `Int32`, `Int64`, `UInt32`, `UInt64`, `Float32`, `Float64` and
`Bytes` making it possible to support gRPC but also helping with making REST interface definitions
crisper. The v1 types `Integer` and `Float` have been removed in favor of these new types.

## Separation of Concern

The generated code produced by `goagen` v2 implements a much stronger separation of concerns where
the transport specific logic is encapsulated in a different layer than the actual service code. This
makes it possible to easily expose the same endpoints via different transport mechanisms such as the
built-in HTTP and gRPC support.
