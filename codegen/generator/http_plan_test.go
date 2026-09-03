// This file checks that plugins can find only the ordinary HTTP plan belonging
// to the exact prepared design root they received.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

func TestHTTPReturnsOnlyExactOrdinaryPlan(t *testing.T) {
	root := &expr.RootExpr{}
	sameContents := &expr.RootExpr{}
	httpPlan := &httpcodegen.Plan{}
	jsonrpcPlan := &httpcodegen.Plan{}
	plan := &Plan{
		http:        map[*expr.RootExpr]*httpcodegen.Plan{root: httpPlan},
		jsonrpcHTTP: map[*expr.RootExpr]*httpcodegen.Plan{sameContents: jsonrpcPlan},
	}

	got, ok := plan.HTTP(root)
	require.True(t, ok)
	require.Same(t, httpPlan, got)
	got, ok = plan.HTTP(sameContents)
	require.False(t, ok)
	require.Nil(t, got)
	got, ok = plan.HTTP(&expr.RootExpr{})
	require.False(t, ok)
	require.Nil(t, got)
}
