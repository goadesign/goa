// This file checks that JSON-RPC keeps the number and direction of values
// declared by each service method.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestJSONRPCWebSocketFinalizePreservesStreamKinds checks that a method which
// sends many results keeps its one initial request. It must not be changed into
// a method which also receives many requests.
func TestJSONRPCWebSocketFinalizePreservesStreamKinds(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("socket", func() {
			dsl.JSONRPC(func() {
				dsl.Path("/stream")
			})
			dsl.Method("server", func() {
				dsl.Payload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {})
			})
			dsl.Method("bidi", func() {
				dsl.StreamingPayload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {})
			})
		})
	})

	service := root.Service("socket")
	server := service.Method("server")
	require.Equal(t, expr.ServerStreamKind, server.Stream)
	require.Equal(t, expr.String, server.Payload.Type)
	require.Equal(t, expr.Empty, server.StreamingPayload.Type)

	bidi := service.Method("bidi")
	require.Equal(t, expr.BidirectionalStreamKind, bidi.Stream)
	require.Equal(t, expr.String, bidi.StreamingPayload.Type)
}
