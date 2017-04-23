package rest

import (
	"goa.design/goa.v2/eval"
)

// Register DSL roots.
func init() {
	if err := eval.Register(Root); err != nil {
		panic(err) // bug
	}
}
