// Package schema supplies reusable declarations without choosing a generated
// package. Consumers choose where their containing service values are written.
package schema

import (
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type (
	// Types identifies the declared values reused by the placement fixtures.
	Types struct {
		// Text is a named string used by required and optional fields.
		Text expr.UserType
		// Nested is the object referenced by Child.
		Nested expr.UserType
		// Child is a recursive object used directly and through a union.
		Child expr.UserType
		// Entry contains the union and a byte field.
		Entry expr.UserType
	}
)

// Define creates the schema through the ordinary DSL in the current design.
// Tests call it after resetting evaluation, just as an imported design creates
// its declarations when the generator loads that design.
func Define() Types {
	text := dsl.Type("TextValue", dsl.String)
	nested := dsl.Type("NestedValue", func() {
		dsl.Field(1, "value", dsl.String)
		dsl.Required("value")
	})
	child := dsl.Type("Child", func() {
		dsl.Field(1, "value", text)
		dsl.Field(2, "optional", text)
		dsl.Field(3, "next", "Child")
		dsl.Field(4, "leaf", nested)
		dsl.Required("value")
	})
	entry := dsl.Type("Entry", func() {
		dsl.OneOf("choice", func() {
			dsl.TypeName("Selection")
			dsl.Field(1, "text", dsl.String)
			dsl.Field(2, "child", child)
			dsl.Field(3, "words", dsl.ArrayOf(dsl.String))
			dsl.Field(4, "table", dsl.MapOf(dsl.String, dsl.ArrayOf(dsl.String)))
		})
		dsl.Field(5, "blob", dsl.Bytes)
		dsl.Required("choice")
	})
	return Types{Text: text, Nested: nested, Child: child, Entry: entry}
}
