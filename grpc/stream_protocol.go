package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	// StreamProtocolMetadataKey is the gRPC request metadata key under which
	// generated clients declare the stream protocol they speak for methods
	// that combine a one-shot payload with a streaming payload.
	StreamProtocolMetadataKey = "goa-stream-protocol"

	// StreamProtocolEnvelope identifies the stream protocol that carries the
	// one-shot method payload as a typed initial stream frame. Clients that
	// predate this protocol send no protocol metadata and carry the payload
	// in gRPC request metadata instead.
	StreamProtocolEnvelope = "2"
)

// UsesStreamEnvelope reports whether the incoming request metadata declares
// the stream envelope protocol. Generated servers use it to decide whether
// the client sends the one-shot method payload as a typed initial stream
// frame or, for legacy clients, in gRPC request metadata.
func UsesStreamEnvelope(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	vals := md.Get(StreamProtocolMetadataKey)
	return len(vals) > 0 && vals[0] == StreamProtocolEnvelope
}
