package design

import "goa.design/goa.v2/eval"

type (
	// UserTypeExpr is the struct used to describe user defined types.
	UserTypeExpr struct {
		// A user type expression is a field expression.
		*AttributeExpr
		// Name of type
		TypeName string
		// Service this type is the default type for if any
		Service *ServiceExpr
	}
)

// NewUserTypeExpr creates a user type expression but does not execute the DSL.
func NewUserTypeExpr(name string, fn func()) *UserTypeExpr {
	return &UserTypeExpr{
		TypeName:      name,
		AttributeExpr: &AttributeExpr{DSLFunc: fn},
	}
}

// Kind implements DataKind.
func (u *UserTypeExpr) Kind() Kind { return UserTypeKind }

// Name returns the type name.
func (u *UserTypeExpr) Name() string { return u.TypeName }

// Rename changes the type name to the given value.
func (u *UserTypeExpr) Rename(n string) { u.TypeName = n }

// IsCompatible returns true if u describes the (Go) type of val.
func (u *UserTypeExpr) IsCompatible(val interface{}) bool {
	return u.Type == nil || u.Type.IsCompatible(val)
}

// Attribute returns the embedded attribute.
func (u *UserTypeExpr) Attribute() *AttributeExpr {
	return u.AttributeExpr
}

// Dup creates a deep copy of the user type given a deep copy of its attribute.
func (u *UserTypeExpr) Dup(att *AttributeExpr) UserType {
	return &UserTypeExpr{
		AttributeExpr: att,
		TypeName:      u.TypeName,
		Service:       u.Service,
	}
}

// Hash returns a unique hash value for u.
func (u *UserTypeExpr) Hash() string {
	return "_type_+" + u.TypeName
}

// Validate checks that the user type definition is consistent: it has a name
// and the attribute backing the type is valid.
func (u *UserTypeExpr) Validate(ctx string, parent eval.Expression) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if u.TypeName == "" {
		verr.Add(parent, "%s - %s", ctx, "User type must have a name")
	}
	verr.Merge(u.AttributeExpr.Validate(ctx, u))
	return verr
}

// Finalize merges base type attributes.
func (u *UserTypeExpr) Finalize() {
	if u.Reference != nil {
		if bat := u.AttributeExpr; bat != nil {
			u.AttributeExpr.Inherit(bat)
		}
	}
}

// Example produces an example for the user type which is JSON serialization
// compatible.
func (u *UserTypeExpr) Example(r *Random) interface{} {
	return u.AttributeExpr.Type.Example(r)
}
