package expr

import (
	"goa.design/goa/eval"
)

// Register DSL roots.
func init() {
	if err := eval.Register(Root); err != nil {
		panic(err) // bug
	}
	if err := eval.Register(Root.GeneratedTypes); err != nil {
		panic(err) // bug
	}
}
