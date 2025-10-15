package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// initDSL initializes the DSL environment and returns the root.
func initDSL(t *testing.T) *expr.RootExpr {
	// reset all roots and codegen data structures
	eval.Reset()
	expr.Root = new(expr.RootExpr)
	expr.GeneratedResultTypes = new(expr.ResultTypesRoot)
	expr.Root.API = expr.NewAPIExpr("test api", func() {})
	expr.Root.API.Servers = []*expr.ServerExpr{expr.Root.API.DefaultServer()}
	root := expr.Root
	require.NoError(t, eval.Register(root))
	require.NoError(t, eval.Register(expr.GeneratedResultTypes))
	return root
}

// runDSL returns the DSL root resulting from running the given DSL.
func runDSL(t *testing.T, dsl func()) *expr.RootExpr {
	root := initDSL(t)
	require.True(t, eval.Execute(dsl, nil))
	require.NoError(t, eval.RunDSL())
	return root
}

// runDSLWithError returns the DSL root and error from running the given DSL.
func runDSLWithError(t *testing.T, dsl func()) (*expr.RootExpr, error) {
	root := initDSL(t)
	require.True(t, eval.Execute(dsl, nil))
	err := eval.RunDSL()
	require.Error(t, err)
	return root, err
}
