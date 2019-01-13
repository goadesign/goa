package expr

import (
	"testing"

	"goa.design/goa/eval"
)

// RunHTTPDSL returns the http DSL root resulting from running the given DSL.
func RunHTTPDSL(t *testing.T, dsl func()) *RootExpr {
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

// RunInvalidHTTPDSL returns the error resulting from running the given DSL.
func RunInvalidHTTPDSL(t *testing.T, dsl func()) error {
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

// RunGRPCDSL returns the gRPC DSL root resulting from running the given DSL.
func RunGRPCDSL(t *testing.T, dsl func()) *RootExpr {
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
