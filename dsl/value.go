package dsl

import "goa.design/goa.v2/design"
import "goa.design/goa.v2/eval"

// Value sets the example value.
//
// Value may appear in Example.
// Value takes one argument: the example value.
//
// Example:
//
//	Example("A simple bottle", func() {
//		Description("This bottle has an ID set to 1")
//		Value(Val{"ID": 1})
//	})
//
func Value(val interface{}) {
	switch expr := eval.Current().(type) {
	case *design.ExampleExpr:
		if v, ok := val.(design.Val); ok {
			val = map[string]interface{}(v)
		}
		expr.Value = val
	default:
		eval.IncompatibleDSL()
	}
}
