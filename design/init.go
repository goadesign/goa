package design

import (
	"goa.design/goa.v2/eval"
)

// Register DSL roots.
func init() {
	eval.Register(Root)
	eval.Register(Root.GeneratedMediaTypes)
}
