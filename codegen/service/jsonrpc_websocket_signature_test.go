// This file checks where generated JSON-RPC WebSocket methods receive their
// request values.
package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestJSONRPCWebSocketBidirectionalPayloadIsStreamOwned checks that a method
// which receives many values reads them with Recv. A method which receives one
// value gets it as its first argument.
func TestJSONRPCWebSocketBidirectionalPayloadIsStreamOwned(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("socket", func() {
			dsl.JSONRPC(func() {
				dsl.Path("/stream")
			})
			dsl.Method("bidi", func() {
				dsl.StreamingPayload(func() {
					dsl.Attribute("value", dsl.String)
				})
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {})
			})
			dsl.Method("server", func() {
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.String)
				})
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {})
			})
		})
	})
	plan := mustServicePlan(t, root)
	facts := plan.facts.services[0]

	serviceCode := renderSignatureFile(t, serviceFiles(plan, facts)[0])
	require.Contains(t, serviceCode, "Bidi(context.Context, BidiServerStream) (err error)")
	require.Contains(t, serviceCode, "Server(context.Context, *ServerPayload, ServerServerStream) (err error)")

	endpointCode := renderSignatureFile(t, endpointFile(plan, facts))
	require.Contains(t, endpointCode, "return nil, s.Bidi(ctx, ep.Stream)")
	require.Contains(t, endpointCode, "return nil, s.Server(ctx, ep.Payload, ep.Stream)")
}

// renderSignatureFile returns the Go source produced for one file.
func renderSignatureFile(t *testing.T, file *codegen.File) string {
	t.Helper()
	var source strings.Builder
	for _, section := range file.SectionTemplates {
		require.NoError(t, section.Write(&source))
	}
	return source.String()
}
