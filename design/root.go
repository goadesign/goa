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
		// Services contains the list of services exposed by the API.
		Services []*ServiceExpr
		// Types contains the user and media types described in the DSL.
		Types []UserType
		// GeneratedMediaTypes contains the set of media types created
		// by CollectionOf.
		GeneratedMediaTypes []*MediaTypeExpr
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

// GeneratedMediaType returns the generated media type expression with the given
// id, nil if there isn't one.
func (r *RootExpr) GeneratedMediaType(id string) *MediaTypeExpr {
	for _, mt := range r.GeneratedMediaTypes {
		if mt.Identifier == id {
			return mt
		}
	}
	return nil
}

// Service returns the service with the given name.
func (r *RootExpr) Service(name string) *ServiceExpr {
	for _, s := range r.Services {
		if s.Name == name {
			return s
		}
	}
	return nil
}
