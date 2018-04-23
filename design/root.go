package design

import "goa.design/goa/eval"

// Root is the root object built by the DSL.
var Root = &RootExpr{GeneratedTypes: &GeneratedRoot{}}

type (
	// RootExpr is the struct built by the DSL on process start.
	RootExpr struct {
		// API contains the API expression built by the DSL.
		API *APIExpr
		// Services contains the list of services exposed by the API.
		Services []*ServiceExpr
		// Errors contains the list of errors returned by all the API
		// methods.
		Errors []*ErrorExpr
		// Types contains the user types described in the DSL.
		Types []UserType
		// ResultTypes contains the result types described in the DSL.
		ResultTypes []UserType
		// GeneratedTypes contains the types generated during DSL
		// execution.
		GeneratedTypes *GeneratedRoot
		// Conversions list the user type to external type mappings.
		Conversions []*TypeMap
		// Creations list the external type to user type mappings.
		Creations []*TypeMap
		// Schemes list the registered security schemes.
		Schemes []*SchemeExpr
	}

	// MetadataExpr is a set of key/value pairs
	MetadataExpr map[string][]string

	// TypeMap defines a user to external type mapping.
	TypeMap struct {
		// User is the user type being converted or created.
		User UserType

		// External is an instance of the type being converted from or to.
		External interface{}
	}
)

// WalkSets returns the expressions in order of evaluation.
func (r *RootExpr) WalkSets(walk eval.SetWalker) {
	if r.API == nil {
		r.API = &APIExpr{}
	}

	// First run the top level API DSL
	walk(eval.ExpressionSet{r.API})

	// Then run the user type DSLs
	types := make(eval.ExpressionSet, len(r.Types))
	for i, t := range r.Types {
		types[i] = t.Attribute()
	}
	walk(types)

	// Next result types
	mtypes := make(eval.ExpressionSet, len(r.ResultTypes))
	for i, mt := range r.ResultTypes {
		mtypes[i] = mt.(*ResultTypeExpr)
	}
	walk(mtypes)

	// Next the services
	services := make(eval.ExpressionSet, len(r.Services))
	for i, s := range r.Services {
		services[i] = s
	}
	walk(services)

	// Next the methods
	var methods eval.ExpressionSet
	for _, s := range r.Services {
		for _, e := range s.Methods {
			methods = append(methods, e)
		}
	}
	walk(methods)
}

// DependsOn returns nil, the core DSL has no dependency.
func (r *RootExpr) DependsOn() []eval.Root { return nil }

// Packages returns the Go import path to this and the dsl packages.
func (r *RootExpr) Packages() []string {
	return []string{
		"goa.design/goa/design",
		"goa.design/goa/dsl",
	}
}

// UserType returns the user type expression with the given name if found, nil otherwise.
func (r *RootExpr) UserType(name string) UserType {
	for _, t := range r.Types {
		if t.Name() == name {
			return t
		}
	}
	for _, t := range r.ResultTypes {
		if t.Name() == name {
			return t
		}
	}
	return nil
}

// GeneratedResultType returns the generated result type expression with the given
// id, nil if there isn't one.
func (r *RootExpr) GeneratedResultType(id string) *ResultTypeExpr {
	for _, t := range *r.GeneratedTypes {
		mt := t.(*ResultTypeExpr)
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
