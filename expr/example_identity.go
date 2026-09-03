// This file builds repeatable keys for example values from evaluated service,
// method, type, field, and transport names.
package expr

import (
	"encoding/base64"
	"encoding/binary"
)

type (
	// ExampleIdentity selects a repeatable sequence of generated example values.
	// Equal values select the same sequence. Its fields are private so callers
	// must use the constructors below.
	ExampleIdentity struct {
		seed string
	}

	exampleIdentityKind byte
)

const (
	userTypeExampleKind exampleIdentityKind = iota + 1
	methodPayloadExampleKind
	methodResultExampleKind
	methodStreamingPayloadExampleKind
	methodStreamingResultExampleKind
	methodErrorExampleKind
	httpRequestBodyExampleKind
	httpResponseBodyExampleKind
	httpErrorResponseBodyExampleKind
	jsonRPCRequestBodyExampleKind
	jsonRPCResponseBodyExampleKind
	jsonRPCErrorResponseBodyExampleKind
	grpcRequestMessageExampleKind
	grpcResponseMessageExampleKind
	grpcStreamingRequestMessageExampleKind
	grpcStreamingResponseMessageExampleKind
	grpcErrorMessageExampleKind
	grpcArrayWrapperExampleKind
	grpcMapWrapperExampleKind
	memberExampleKind
	arrayElementExampleKind
	mapKeyExampleKind
	mapValueExampleKind
	unionMemberExampleKind
)

// UserTypeExampleIdentity returns the example key for typ.
func UserTypeExampleIdentity(typ UserType) ExampleIdentity {
	if identity, ok := GeneratedUserTypeExampleIdentity(typ); ok {
		return identity
	}
	return newExampleIdentity(userTypeExampleKind, []byte(typ.ID()))
}

// GeneratedUserTypeExampleIdentity returns the example key stored on a user
// type created by Goa. The second result is false for types written in the
// design.
func GeneratedUserTypeExampleIdentity(typ UserType) (ExampleIdentity, bool) {
	var identity ExampleIdentity
	switch generated := typ.(type) {
	case *UserTypeExpr:
		identity = generated.exampleIdentity
	case *ResultTypeExpr:
		identity = generated.UserTypeExpr.exampleIdentity
	}
	return identity, identity.seed != ""
}

// MethodPayloadExampleIdentity returns the example key for method's payload.
func MethodPayloadExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodPayloadExampleKind, method)
}

// MethodResultExampleIdentity returns the example key for method's result.
func MethodResultExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodResultExampleKind, method)
}

// MethodStreamingPayloadExampleIdentity returns the example key for method's
// streaming payload.
func MethodStreamingPayloadExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodStreamingPayloadExampleKind, method)
}

// MethodStreamingResultExampleIdentity returns the example key for method's
// streaming result.
func MethodStreamingResultExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodStreamingResultExampleKind, method)
}

// MethodErrorExampleIdentity returns the example key for err in method.
func MethodErrorExampleIdentity(method *MethodExpr, err *ErrorExpr) ExampleIdentity {
	return newExampleIdentity(
		methodErrorExampleKind,
		[]byte(method.Service.Name),
		[]byte(method.Name),
		[]byte(err.Name),
	)
}

// RequestBodyExampleIdentity returns the example key for endpoint's request
// body. HTTP and JSON-RPC endpoints receive different keys.
func RequestBodyExampleIdentity(endpoint *HTTPEndpointExpr) ExampleIdentity {
	kind := httpRequestBodyExampleKind
	if endpoint.IsJSONRPC() {
		kind = jsonRPCRequestBodyExampleKind
	}
	return newExampleIdentity(
		kind,
		[]byte(endpoint.MethodExpr.Service.Name),
		[]byte(endpoint.MethodExpr.Name),
	)
}

// ResponseBodyExampleIdentity returns the example key for a successful response
// body. HTTP and JSON-RPC endpoints receive different keys, and each successful
// status code receives its own key.
func ResponseBodyExampleIdentity(endpoint *HTTPEndpointExpr, response *HTTPResponseExpr) ExampleIdentity {
	kind := httpResponseBodyExampleKind
	if endpoint.IsJSONRPC() {
		kind = jsonRPCResponseBodyExampleKind
	}
	return newExampleIdentity(
		kind,
		[]byte(endpoint.MethodExpr.Service.Name),
		[]byte(endpoint.MethodExpr.Name),
		exampleIdentityInt(response.StatusCode),
	)
}

// ErrorResponseBodyExampleIdentity returns the example key for an error response
// body. HTTP and JSON-RPC endpoints receive different keys. The error name keeps
// two errors with the same HTTP status separate.
func ErrorResponseBodyExampleIdentity(endpoint *HTTPEndpointExpr, response *HTTPErrorExpr) ExampleIdentity {
	kind := httpErrorResponseBodyExampleKind
	if endpoint.IsJSONRPC() {
		kind = jsonRPCErrorResponseBodyExampleKind
	}
	return newExampleIdentity(
		kind,
		[]byte(endpoint.MethodExpr.Service.Name),
		[]byte(endpoint.MethodExpr.Name),
		[]byte(response.Name),
		exampleIdentityInt(response.Response.StatusCode),
	)
}

// GRPCRequestMessageExampleIdentity returns the example key for method's gRPC
// request message.
func GRPCRequestMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcRequestMessageExampleKind, method)
}

// GRPCResponseMessageExampleIdentity returns the example key for method's gRPC
// response message.
func GRPCResponseMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcResponseMessageExampleKind, method)
}

// GRPCStreamingRequestMessageExampleIdentity returns the example key for
// method's streaming gRPC request message.
func GRPCStreamingRequestMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcStreamingRequestMessageExampleKind, method)
}

// GRPCStreamingResponseMessageExampleIdentity returns the example key for
// method's streaming gRPC response message.
func GRPCStreamingResponseMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcStreamingResponseMessageExampleKind, method)
}

// GRPCErrorMessageExampleIdentity returns the example key for err's gRPC
// message in method.
func GRPCErrorMessageExampleIdentity(method *MethodExpr, err *ErrorExpr) ExampleIdentity {
	return newExampleIdentity(
		grpcErrorMessageExampleKind,
		[]byte(method.Service.Name),
		[]byte(method.Name),
		[]byte(err.Name),
	)
}

// GRPCArrayWrapperExampleIdentity returns the example key for the gRPC message
// that wraps an array type written in the design.
func GRPCArrayWrapperExampleIdentity(typ UserType) ExampleIdentity {
	if !IsArray(typ) {
		panic("gRPC array wrapper identity requires an array user type")
	}
	return newExampleIdentity(grpcArrayWrapperExampleKind, []byte(typ.Origin().ID()))
}

// GRPCMapWrapperExampleIdentity returns the example key for the gRPC message
// that wraps a map type written in the design.
func GRPCMapWrapperExampleIdentity(typ UserType) ExampleIdentity {
	if !IsMap(typ) {
		panic("gRPC map wrapper identity requires a map user type")
	}
	return newExampleIdentity(grpcMapWrapperExampleKind, []byte(typ.Origin().ID()))
}

// Seed returns the complete encoded key passed to custom randomizer factories.
func (i ExampleIdentity) Seed() string {
	return base64.RawURLEncoding.EncodeToString([]byte(i.seed))
}

// Member returns the example key for name within i.
func (i ExampleIdentity) Member(name string) ExampleIdentity {
	return i.append(memberExampleKind, []byte(name))
}

// ArrayElement returns the example key for index within the array at i.
func (i ExampleIdentity) ArrayElement(index int) ExampleIdentity {
	return i.append(arrayElementExampleKind, exampleIdentityInt(index))
}

// MapKey returns the example key for key index within the map at i.
func (i ExampleIdentity) MapKey(index int) ExampleIdentity {
	return i.append(mapKeyExampleKind, exampleIdentityInt(index))
}

// MapValue returns the example key for value index within the map at i.
func (i ExampleIdentity) MapValue(index int) ExampleIdentity {
	return i.append(mapValueExampleKind, exampleIdentityInt(index))
}

// UnionMember returns the example key for name within the union at i.
func (i ExampleIdentity) UnionMember(name string) ExampleIdentity {
	return i.append(unionMemberExampleKind, []byte(name))
}

// A new example key encodes a value kind and each component's byte length so
// different component lists cannot produce the same key.
func newExampleIdentity(kind exampleIdentityKind, components ...[]byte) ExampleIdentity {
	return ExampleIdentity{seed: string(appendExampleIdentitySegment(nil, kind, components...))}
}

// Method example keys are built from the evaluated service and method names.
func methodExampleIdentity(kind exampleIdentityKind, method *MethodExpr) ExampleIdentity {
	return newExampleIdentity(kind, []byte(method.Service.Name), []byte(method.Name))
}

// Integers in example keys are written as eight bytes in big-endian order.
func exampleIdentityInt(value int) []byte {
	return binary.BigEndian.AppendUint64(nil, uint64(value))
}

// Each added part writes its kind, number of components, and every component's
// byte length before the component bytes.
func appendExampleIdentitySegment(seed []byte, kind exampleIdentityKind, components ...[]byte) []byte {
	seed = append(seed, byte(kind))
	seed = binary.BigEndian.AppendUint64(seed, uint64(len(components)))
	for _, component := range components {
		seed = binary.BigEndian.AppendUint64(seed, uint64(len(component)))
		seed = append(seed, component...)
	}
	return seed
}

// append adds one member, array, map, or union part to the current key.
func (i ExampleIdentity) append(kind exampleIdentityKind, components ...[]byte) ExampleIdentity {
	if i.seed == "" {
		panic("example identity must have a semantic owner before structural descent")
	}
	seed := append([]byte(nil), i.seed...)
	seed = appendExampleIdentitySegment(seed, kind, components...)
	return ExampleIdentity{seed: string(seed)}
}
