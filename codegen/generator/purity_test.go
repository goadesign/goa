// This file verifies through the production lifecycle boundary that only
// preparation and normalization may change evaluated design expressions.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	grpcdata "goa.design/goa/v3/grpc/codegen/testdata"
	httpdata "goa.design/goa/v3/http/codegen/testdata"
	jsonrpcdata "goa.design/goa/v3/jsonrpc/codegen/testdata"
)

// TestGeneratorsTreatDesignAsReadOnly audits the persistent design state after
// generation construction applies the sanctioned normalization. Running every
// generator ("gen" and "example") must leave the prepared semantic design
// unchanged after each callback and completed render. The fixtures cover alias
// chains, result views, websocket streaming, SSE with anonymous object payloads
// and results, mixed HTTP+JSON-RPC transports, and gRPC unions and streaming.
//
// The production lifecycle snapshot owns this audit. Dormant eval.DSLFunc
// closure captures and process-global state outside the prepared roots are not
// evaluated design input and remain outside this assertion.
func TestGeneratorsTreatDesignAsReadOnly(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"alias-chains", httpdata.AliasTypeDSL},
		{"result-views", httpdata.ResultBodyMultipleViewsDSL},
		{"websocket-bidirectional", httpdata.BidirectionalStreamingDSL},
		{"sse-anonymous-object", httpdata.SSEObjectDSL},
		{"jsonrpc-mixed-transport", jsonrpcdata.JSONRPCKitchenSinkDSL},
		{"grpc-union-streaming", grpcdata.ClientStreamingRPCWithUnionPayloadDSL},
		{"grpc-streaming-views", grpcdata.ServerStreamingResultWithViewsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)

			for _, cmd := range []string{"gen", "example"} {
				_, err := executeGeneration("gen", []eval.Root{root}, cmd, newDefaultRegistry())
				require.NoError(t, err)
			}
		})
	}
}

// TestPreparedRootsRejectAttributeMutation proves that generation stops when
// a core planner changes an attribute after the mutable lifecycle phase.
func TestPreparedRootsRejectAttributeMutation(t *testing.T) {
	root := expr.RunDSL(t, httpdata.AliasTypeDSL)
	registry := newRegistry()
	target := root.Types[0].Attribute()
	followingRan := false
	registry.addCommand(
		"test",
		func() coreGenerator {
			return coreGenerator{name: "attribute-mutator", Plan: func(_ *Plan) error {
				target.Description = "changed after preparation"
				return nil
			}}
		},
		func() coreGenerator {
			return coreGenerator{name: "following", Plan: func(_ *Plan) error {
				followingRan = true
				return nil
			}}
		},
	)

	_, err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)
	require.ErrorContains(t, err, `core "attribute-mutator" plan mutated prepared design`)
	require.False(t, followingRan)
}
