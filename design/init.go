package design

import (
	"goa.design/goa.v2/eval"
)

// Register DSL roots.
func init() {
	if err := eval.Register(Root); err != nil {
		panic(err) // bug
	}
	if err := eval.Register(Root.GeneratedMediaTypes); err != nil {
		panic(err) // bug
	}
}
