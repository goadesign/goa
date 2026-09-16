package dsl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"goa.design/goa/v3/expr"
)

// TestCollectionOfReusesGeneratedType verifies that CollectionOf reuses the
// generated collection type when it is called more than once for the same
// result type. The lookup is made with a canonical identifier, so a result type
// whose identifier carries a suffix such as "+json" must still match. Creating a
// second type makes two declarations ask for one generated Go name.
func TestCollectionOfReusesGeneratedType(t *testing.T) {
	var first, second *expr.ResultTypeExpr
	expr.RunDSL(t, func() {
		rt := ResultType("application/vnd.item+json", func() {
			Attributes(func() {
				Attribute("id", expr.Int)
				Required("id")
			})
		})
		Service("svc", func() {
			Method("first", func() {
				first = CollectionOf(rt)
				Result(first)
				HTTP(func() { GET("/first") })
			})
			Method("second", func() {
				second = CollectionOf(rt)
				Result(second)
				HTTP(func() { GET("/second") })
			})
		})
	})

	assert.Same(t, first, second)
	assert.Len(t, *expr.GeneratedResultTypes, 1)
}
