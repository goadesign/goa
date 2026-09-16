// Package nestedalpha supplies an external field that the design deliberately
// does not map, so conversion planning can prove unused fields reserve nothing.
package nestedalpha

// Child has the same package and type names as a mapped conversion field.
type Child struct {
	Value string
}
