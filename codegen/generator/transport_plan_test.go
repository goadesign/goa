// This file checks that plugins can find the gRPC and JSON-RPC plans retained
// for the exact prepared design root they received.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

func TestTransportPlansUseExactRoot(t *testing.T) {
	grpcRoot := &expr.RootExpr{}
	jsonrpcRoot := &expr.RootExpr{}
	grpcPlan := &grpccodegen.Plan{}
	jsonrpcPlan := &jsonrpccodegen.Plan{}
	plan := &Plan{
		grpc:    map[*expr.RootExpr]*grpccodegen.Plan{grpcRoot: grpcPlan},
		jsonrpc: map[*expr.RootExpr]*jsonrpccodegen.Plan{jsonrpcRoot: jsonrpcPlan},
	}

	gotGRPC, ok := plan.GRPC(grpcRoot)
	require.True(t, ok)
	require.Same(t, grpcPlan, gotGRPC)
	gotGRPC, ok = plan.GRPC(&expr.RootExpr{})
	require.False(t, ok)
	require.Nil(t, gotGRPC)

	gotJSONRPC, ok := plan.JSONRPC(jsonrpcRoot)
	require.True(t, ok)
	require.Same(t, jsonrpcPlan, gotJSONRPC)
	gotJSONRPC, ok = plan.JSONRPC(&expr.RootExpr{})
	require.False(t, ok)
	require.Nil(t, gotJSONRPC)
}
