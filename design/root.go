package design

// Root is the root object built by the DSL.
var Root = new(RootExpr)

type (
	// RootExpr is the struct built by the DSL on process start.
	RootExpr struct {
		// API contains the API expression built by the DSL.
		API *APIExpr
		// Traits contains the trait expressions built by the DSL.
		Traits []*TraitExpr
		// Types contains the user types described in the DSL.
		Types []UserType
		// NoExamples is a boolean that indicates whether to generate random examples
		// (false) or not (true).
		NoExamples bool
	}

	// MetadataExpr is a set of key/value pairs
	MetadataExpr map[string][]string

	// TraitExpr defines a set of reusable properties.
	TraitExpr struct {
		// Trait name
		Name string
		// Trait DSL
		DSLFunc func()
	}
)

// Trait returns the trait expression with the given name if found, nil otherwise.
func (r *RootExpr) Trait(name string) *TraitExpr {
	for _, t := range r.Traits {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// UserType returns the user type expression with the given name if found, nil otherwise.
func (r *RootExpr) UserType(name string) UserType {
	for _, t := range r.Types {
		if t.Name() == name {
			return t
		}
	}
	return nil
}
