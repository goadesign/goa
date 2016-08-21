package design

// Root is the root object built by the DSL.
var Root = new(RootExpr)

type (
	// RootExpr is the struct built by the DSL on process start.
	RootExpr struct {
		// Services contains the service expressions built by the DSL.
		Services []*ServiceExpr
		// Traits contains the trait expressions built by the DSL.
		Traits []*TraitExpr
		// Types contains the user types described in the DSL.
		Types []*UserTypeExpr
		// MediaTypes contains the media types described in the DSL.
		MediaTypes []*MediaTypeExpr
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

// Service returns the service expression with the given name and true if found, nil and false
// otherwise.
func (r *RootExpr) Service(name string) (*ServiceExpr, bool) {
	for _, s := range r.Services {
		if s.Name == name {
			return s, true
		}
	}
	return nil, false
}

// Trait returns the trait expression with the given name and true if found, nil and false
// otherwise.
func (r *RootExpr) Trait(name string) (*TraitExpr, bool) {
	for _, t := range r.Traits {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

// UserType returns the user type expression with the given name and true if found, nil and false
// otherwise.
func (r *RootExpr) UserType(name string) (*UserTypeExpr, bool) {
	for _, t := range r.Types {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

// MediaType returns the media type expression with the given id and true if found, nil and false
// otherwise.
func (r *RootExpr) MediaType(id string) (*MediaTypeExpr, bool) {
	for _, mt := range r.MediaTypes {
		if mt.ID == id {
			return mt, true
		}
	}
	return nil, false
}
