package design

import (
	"sort"

	"goa.design/goa.v2/eval"
)

type (
	// GeneratedRoot records the generated media types and is a DSL root
	// evaluated after Root.
	GeneratedRoot []UserType
)

// EvalName is the name of the expression used by eval.
func (r GeneratedRoot) EvalName() string {
	return "generated media types"
}

// WalkSets returns the generated media types for evaluation.
func (r GeneratedRoot) WalkSets(w eval.SetWalker) {
	ids := make([]string, len(r))
	for i, t := range r {
		mt := t.(*MediaTypeExpr)
		id := CanonicalIdentifier(mt.Identifier)
		Root.MediaTypes = append(Root.MediaTypes, mt)
		ids[i] = id
	}
	sort.Strings(ids)
	set := make(eval.ExpressionSet, len(ids))
	for i, id := range ids {
		set[i] = Root.UserType(id)
	}
	w(set)
}

// DependsOn ensures that Root executes first.
func (r GeneratedRoot) DependsOn() []eval.Root {
	return []eval.Root{Root}
}

// Packages returns the Go import path to this and the dsl packages.
func (r GeneratedRoot) Packages() []string {
	return []string{
		"goa.design/goa.v2/design",
		"goa.design/goa.v2/dsl",
	}
}

// Used returns true if the DSL makes use of CollectionOf.
func (r GeneratedRoot) Used() bool {
	return len(r) > 0
}
