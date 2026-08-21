// This file defines stable, typed identities for the example values emitted
// from evaluated design expressions.
package expr

import (
	"encoding/base64"
	"encoding/binary"
)

type (
	// ExampleIdentity identifies one semantic example stream. Its representation
	// is opaque so callers cannot manufacture identities by joining names.
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

// UserTypeExampleIdentity returns the example identity owned by typ.
func UserTypeExampleIdentity(typ UserType) ExampleIdentity {
	if identity, ok := GeneratedUserTypeExampleIdentity(typ); ok {
		return identity
	}
	return newExampleIdentity(userTypeExampleKind, []byte(typ.ID()))
}

// GeneratedUserTypeExampleIdentity returns the exact semantic owner retained
// by a synthesized user type. The second result is false for authored types.
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

// MethodPayloadExampleIdentity returns the payload example identity owned by
// method.
func MethodPayloadExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodPayloadExampleKind, method)
}

// MethodResultExampleIdentity returns the result example identity owned by
// method.
func MethodResultExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodResultExampleKind, method)
}

// MethodStreamingPayloadExampleIdentity returns the streaming payload example
// identity owned by method.
func MethodStreamingPayloadExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodStreamingPayloadExampleKind, method)
}

// MethodStreamingResultExampleIdentity returns the streaming result example
// identity owned by method.
func MethodStreamingResultExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(methodStreamingResultExampleKind, method)
}

// MethodErrorExampleIdentity returns the example identity owned by err in
// method.
func MethodErrorExampleIdentity(method *MethodExpr, err *ErrorExpr) ExampleIdentity {
	return newExampleIdentity(
		methodErrorExampleKind,
		[]byte(method.Service.Name),
		[]byte(method.Name),
		[]byte(err.Name),
	)
}

// RequestBodyExampleIdentity returns the request body example identity owned
// by endpoint. HTTP and JSON-RPC mappings receive distinct identities.
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

// ResponseBodyExampleIdentity returns the successful response body example
// identity owned by response in endpoint. HTTP and JSON-RPC mappings receive
// distinct identities. Endpoint validation makes each successful status code
// unique.
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

// ErrorResponseBodyExampleIdentity returns the error response body example
// identity owned by response in endpoint. HTTP and JSON-RPC mappings receive
// distinct identities. Error names distinguish errors that intentionally
// share an HTTP status.
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

// GRPCRequestMessageExampleIdentity returns the gRPC request message example
// identity owned by method.
func GRPCRequestMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcRequestMessageExampleKind, method)
}

// GRPCResponseMessageExampleIdentity returns the gRPC response message example
// identity owned by method.
func GRPCResponseMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcResponseMessageExampleKind, method)
}

// GRPCStreamingRequestMessageExampleIdentity returns the gRPC streaming
// request message example identity owned by method.
func GRPCStreamingRequestMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcStreamingRequestMessageExampleKind, method)
}

// GRPCStreamingResponseMessageExampleIdentity returns the gRPC streaming
// response message example identity owned by method.
func GRPCStreamingResponseMessageExampleIdentity(method *MethodExpr) ExampleIdentity {
	return methodExampleIdentity(grpcStreamingResponseMessageExampleKind, method)
}

// GRPCErrorMessageExampleIdentity returns the gRPC error message example
// identity owned by err in method.
func GRPCErrorMessageExampleIdentity(method *MethodExpr, err *ErrorExpr) ExampleIdentity {
	return newExampleIdentity(
		grpcErrorMessageExampleKind,
		[]byte(method.Service.Name),
		[]byte(method.Name),
		[]byte(err.Name),
	)
}

// GRPCArrayWrapperExampleIdentity returns the stable gRPC wrapper identity for
// an authored array alias shared across message fields.
func GRPCArrayWrapperExampleIdentity(typ UserType) ExampleIdentity {
	if !IsArray(typ) {
		panic("gRPC array wrapper identity requires an array user type")
	}
	return newExampleIdentity(grpcArrayWrapperExampleKind, []byte(typ.Origin().ID()))
}

// GRPCMapWrapperExampleIdentity returns the stable gRPC wrapper identity for
// an authored map alias shared across message fields.
func GRPCMapWrapperExampleIdentity(typ UserType) ExampleIdentity {
	if !IsMap(typ) {
		panic("gRPC map wrapper identity requires a map user type")
	}
	return newExampleIdentity(grpcMapWrapperExampleKind, []byte(typ.Origin().ID()))
}

// Seed returns the complete stable seed material custom randomizer factories
// use to create the stream for this identity.
func (i ExampleIdentity) Seed() string {
	return base64.RawURLEncoding.EncodeToString([]byte(i.seed))
}

// Member returns the identity of the named object member below i.
func (i ExampleIdentity) Member(name string) ExampleIdentity {
	return i.append(memberExampleKind, []byte(name))
}

// ArrayElement returns the identity of the indexed array element below i.
func (i ExampleIdentity) ArrayElement(index int) ExampleIdentity {
	return i.append(arrayElementExampleKind, exampleIdentityInt(index))
}

// MapKey returns the identity of the indexed map key below i.
func (i ExampleIdentity) MapKey(index int) ExampleIdentity {
	return i.append(mapKeyExampleKind, exampleIdentityInt(index))
}

// MapValue returns the identity of the indexed map value below i.
func (i ExampleIdentity) MapValue(index int) ExampleIdentity {
	return i.append(mapValueExampleKind, exampleIdentityInt(index))
}

// UnionMember returns the identity of the named union member below i.
func (i ExampleIdentity) UnionMember(name string) ExampleIdentity {
	return i.append(unionMemberExampleKind, []byte(name))
}

// newExampleIdentity serializes one typed segment with independently framed
// components so punctuation and component boundaries cannot collide.
func newExampleIdentity(kind exampleIdentityKind, components ...[]byte) ExampleIdentity {
	return ExampleIdentity{seed: string(appendExampleIdentitySegment(nil, kind, components...))}
}

// methodExampleIdentity derives a method-owned identity from the evaluated
// service and method names rather than accepting caller-supplied components.
func methodExampleIdentity(kind exampleIdentityKind, method *MethodExpr) ExampleIdentity {
	return newExampleIdentity(kind, []byte(method.Service.Name), []byte(method.Name))
}

// exampleIdentityInt returns a stable fixed-width encoding of value.
func exampleIdentityInt(value int) []byte {
	return binary.BigEndian.AppendUint64(nil, uint64(value))
}

// appendExampleIdentitySegment writes the segment kind, component count, and
// byte length of each component before its data.
func appendExampleIdentitySegment(seed []byte, kind exampleIdentityKind, components ...[]byte) []byte {
	seed = append(seed, byte(kind))
	seed = binary.BigEndian.AppendUint64(seed, uint64(len(components)))
	for _, component := range components {
		seed = binary.BigEndian.AppendUint64(seed, uint64(len(component)))
		seed = append(seed, component...)
	}
	return seed
}

// append adds one structural segment without exposing the serialized form to
// callers.
func (i ExampleIdentity) append(kind exampleIdentityKind, components ...[]byte) ExampleIdentity {
	if i.seed == "" {
		panic("example identity must have a semantic owner before structural descent")
	}
	seed := append([]byte(nil), i.seed...)
	seed = appendExampleIdentitySegment(seed, kind, components...)
	return ExampleIdentity{seed: string(seed)}
}
