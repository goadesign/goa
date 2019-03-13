package expr

import (
	"testing"

	"goa.design/goa/eval"
)

// RunDSL returns the DSL root resulting from running the given DSL.
// Used only in tests.
func RunDSL(t *testing.T, dsl func()) *RootExpr {
	setupDSLRun()

	// run DSL (first pass)
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}

	// run DSL (second pass)
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}

	// return generated root
	return Root
}

// RunInvalidDSL returns the error resulting from running the given DSL.
// It is used only in tests.
func RunInvalidDSL(t *testing.T, dsl func()) error {
	setupDSLRun()

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

func setupDSLRun() {
	// reset all roots and codegen data structures
	eval.Reset()
	Root = new(RootExpr)
	Root.GeneratedTypes = &GeneratedRoot{}
	eval.Register(Root)
	eval.Register(Root.GeneratedTypes)
	Root.API = NewAPIExpr("test api", func() {})
	Root.API.Servers = []*ServerExpr{Root.API.DefaultServer()}
}
