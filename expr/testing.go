package expr

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
)

// RunDSL returns the DSL root resulting from running the given DSL.
// Used only in tests.
func RunDSL(t *testing.T, dsl func()) *RootExpr {
	t.Helper()
	ResetDSL(t)

	// run DSL (first pass)
	require.True(t, eval.Execute(dsl, nil), eval.Context.Error())

	// run DSL (second pass)
	require.NoError(t, eval.RunDSL())

	// return generated root
	return Root
}

// RunInvalidDSL returns the error resulting from running the given DSL.
// It is used only in tests.
func RunInvalidDSL(t *testing.T, dsl func()) error {
	t.Helper()
	ResetDSL(t)

	// run DSL (first pass)
	if !eval.Execute(dsl, nil) {
		return eval.Context.Errors
	}

	// run DSL (second pass)
	if err := eval.RunDSL(); err != nil {
		return err
	}

	// expected an error - didn't get one
	t.Fatal("expected a DSL evaluation error - got none")

	return nil
}

// CreateTempFile creates a temporary file and writes the given content.
// It is used only for testing.
func CreateTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		_ = os.Remove(f.Name())
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

// ResetDSL resets the global expression state for testing and initializes
// a default API. This function should be called before running any DSL that
// modifies the global Root or GeneratedResultTypes variables.
//
// Usage in tests:
//
//	func TestMyDSL(t *testing.T) {
//	    // Option 1: Use expr.RunDSL which calls ResetDSL automatically
//	    root := expr.RunDSL(t, func() {
//	        Service("my-service", func() { /* ... */ })
//	    })
//
//	    // Option 2: Call ResetDSL manually when running DSL directly
//	    expr.ResetDSL(t)
//	    eval.Execute(myDSL, nil)
//	    eval.RunDSL()
//	}
//
// Note: RunDSL and RunInvalidDSL automatically call ResetDSL, so you
// only need to call it manually when executing DSL code directly.
func ResetDSL(t *testing.T) {
	t.Helper()
	// reset all roots and codegen data structures
	eval.Reset()
	Root = new(RootExpr)
	GeneratedResultTypes = new(ResultTypesRoot)
	require.NoError(t, eval.Register(Root))
	require.NoError(t, eval.Register(GeneratedResultTypes))

	// Initialize default API for DSL execution
	Root.API = NewAPIExpr("test api", func() {})
	Root.API.Servers = []*ServerExpr{Root.API.DefaultServer()}
}
