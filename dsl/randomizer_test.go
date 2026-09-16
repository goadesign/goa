// This file verifies that the API DSL stores immutable example factory
// configuration instead of a mutable random stream.
package dsl

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestRandomizerStoresFactory(t *testing.T) {
	factory := expr.NewDeterministicRandomizerFactory()
	api := expr.NewAPIExpr("test", func() {})
	eval.Context = &eval.DSLContext{}

	eval.Execute(func() {
		Randomizer(factory)
	}, api)

	require.Empty(t, eval.Context.Errors)
	require.Equal(t, factory, api.RandomizerFactory)
}

func TestRandomizerRejectsNilFactory(t *testing.T) {
	api := expr.NewAPIExpr("test", func() {})
	eval.Context = &eval.DSLContext{}

	eval.Execute(func() {
		Randomizer(nil)
	}, api)

	require.Len(t, eval.Context.Errors, 1)
	require.Contains(t, eval.Context.Errors[0].Error(), "non-nil randomizer factory")
}
