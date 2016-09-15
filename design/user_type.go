package design

// NewUserTypeExpr creates a user type expression but does not execute the DSL.
func NewUserTypeExpr(name string, dsl func()) *UserTypeExpr {
	return &UserTypeExpr{
		TypeName:      name,
		AttributeExpr: &AttributeExpr{DSLFunc: dsl},
	}
}

// Kind implements DataKind.
func (u *UserTypeExpr) Kind() Kind { return UserTypeKind }

// Name returns the type name.
func (u *UserTypeExpr) Name() string { return u.TypeName }

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
	}
}
