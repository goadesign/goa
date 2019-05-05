package expr

import "goa.design/goa/v3/eval"

type (
	// GeneratedRoot records the generated result types and is a DSL root
	// evaluated after Root.
	GeneratedRoot []UserType
)

// EvalName is the name of the expression used by eval.
func (r *GeneratedRoot) EvalName() string {
	return "generated result types"
}

// WalkSets returns the generated result types for evaluation.
func (r *GeneratedRoot) WalkSets(w eval.SetWalker) {
	if r == nil {
		return
	}
	set := make(eval.ExpressionSet, len(*r))
	for i, t := range *r {
		rt := t.(*ResultTypeExpr)
		Root.ResultTypes = append(Root.ResultTypes, rt)
		set[i] = rt
	}
	w(set)
}

// DependsOn ensures that Root executes first.
func (r *GeneratedRoot) DependsOn() []eval.Root {
	return []eval.Root{Root}
}

// Packages returns the Go import path to this and the dsl packages.
func (r *GeneratedRoot) Packages() []string {
	return []string{
		"goa.design/goa/v3/expr",
		"goa.design/goa/v3/dsl",
	}
}
