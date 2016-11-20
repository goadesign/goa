package design

import "goa.design/goa.v2/eval"

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
		// Types contains the user types described in the DSL.
		Types []UserType
		// MediaTypes contains the media types described in the DSL.
		MediaTypes []UserType
		// GeneratedMediaTypes contains the set of media types created
		// by CollectionOf.
		GeneratedMediaTypes []UserType
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

// DSLName is the name of the DSL.
func (r *RootExpr) DSLName() string {
	return "API " + r.API.Name
}

// DependsOn returns nil, the core DSL has no dependency.
func (r *RootExpr) DependsOn() []eval.Root { return nil }

// IterateSets returns the expressions in order of evaluation.
func (r *RootExpr) IterateSets(it eval.SetIterator) {
	// First run the top level API DSL
	it(eval.ExpressionSet{r.API})

	// Then run the user type DSLs
	typeAttributes := make([]eval.Expression, len(r.Types))
	for i, ut := range r.Types {
		typeAttributes[i] = ut.Attribute()
		i++
	}
	it(typeAttributes)

	// Next media types
	mts := make([]eval.Expression, len(r.MediaTypes))
	for i, mt := range r.MediaTypes {
		mts[i] = mt.(*MediaTypeExpr)
		i++
	}
	it(mts)

	// Next the services
	services := make([]eval.Expression, len(r.Services))
	for i, s := range r.Services {
		services[i] = s
		i++
	}
	it(services)
}

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
	for _, t := range r.MediaTypes {
		if t.Name() == name {
			return t
		}
	}
	return nil
}

// GeneratedMediaType returns the generated media type expression with the given
// id, nil if there isn't one.
func (r *RootExpr) GeneratedMediaType(id string) *MediaTypeExpr {
	for _, t := range r.GeneratedMediaTypes {
		mt := t.(*MediaTypeExpr)
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
