package design

import (
	"goa.design/goa.v2/eval"
)

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
		// Errors contains the list of errors returned by all the API
		// endpoints.
		Errors []*ErrorExpr
		// Types contains the user types described in the DSL.
		Types []UserType
		// MediaTypes contains the media types described in the DSL.
		MediaTypes []UserType
		// GeneratedTypes contains the types generated during DSL
		// execution.
		GeneratedTypes GeneratedRoot
	}

	// MetadataExpr is a set of key/value pairs
	MetadataExpr map[string][]string

	// TraitExpr defines a set of reusable properties.
	TraitExpr struct {
		// Trait name
		Name string
		// Trait DSL
		DSL interface{}
	}
)

// WalkSets returns the expressions in order of evaluation.
func (r *RootExpr) WalkSets(walk eval.SetWalker) {
	// First run the top level API DSL
	walk(eval.ExpressionSet{r.API})

	// Then run the user type DSLs
	types := make(eval.ExpressionSet, len(r.Types))
	for i, t := range r.Types {
		types[i] = t.Attribute()
	}
	walk(types)

	// Next media types
	mtypes := make(eval.ExpressionSet, len(r.MediaTypes))
	for i, mt := range r.MediaTypes {
		mtypes[i] = mt.(*MediaTypeExpr)
	}
	walk(mtypes)

	// Next the services
	services := make(eval.ExpressionSet, len(r.Services))
	for i, s := range r.Services {
		services[i] = s
	}
	walk(services)

	// Next the endpoints
	var endpoints eval.ExpressionSet
	for _, s := range r.Services {
		for _, e := range s.Endpoints {
			endpoints = append(endpoints, e)
		}
	}
	walk(endpoints)
}

// DependsOn returns nil, the core DSL has no dependency.
func (r *RootExpr) DependsOn() []eval.Root { return nil }

// Packages returns the Go import path to this and the dsl packages.
func (r *RootExpr) Packages() []string {
	return []string{
		"goa.design/goa.v2/design",
		"goa.design/goa.v2/dsl",
	}
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
	for _, t := range r.GeneratedTypes {
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

// Error returns the error with the given name.
func (r *RootExpr) Error(name string) *ErrorExpr {
	for _, e := range r.Errors {
		if e.Name == name {
			return e
		}
	}
	return nil
}

// EvalName is the name of the DSL.
func (r *RootExpr) EvalName() string {
	return "design"
}

// Validate makes sure the root expression is valid for code generation.
func (r *RootExpr) Validate() error {
	var verr eval.ValidationErrors
	if r.API == nil {
		verr.Add(r, "Missing API declaration")
	}
	return &verr
}

// Dup creates a new map from the given expression.
func (m MetadataExpr) Dup() MetadataExpr {
	d := make(MetadataExpr, len(m))
	for k, v := range m {
		d[k] = v
	}
	return d
}
