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

// Name returns the JSON type name.
func (u *UserTypeExpr) Name() string { return u.Type.Name() }

// IsPrimitive calls IsPrimitive on the user type underlying data type.
func (u *UserTypeExpr) IsPrimitive() bool { return u.Type.IsPrimitive() }

// IsObject calls IsObject on the user type underlying data type.
func (u *UserTypeExpr) IsObject() bool { return u.Type.IsObject() }

// IsArray calls IsArray on the user type underlying data type.
func (u *UserTypeExpr) IsArray() bool { return u.Type.IsArray() }

// IsMap calls IsMap on the user type underlying data type.
func (u *UserTypeExpr) IsMap() bool { return u.Type.IsMap() }

// ToObject calls ToObject on the user type underlying data type.
func (u *UserTypeExpr) ToObject() Object { return u.Type.ToObject() }

// ToArray calls ToArray on the user type underlying data type.
func (u *UserTypeExpr) ToArray() *Array { return u.Type.ToArray() }

// ToMap calls ToMap on the user type underlying data type.
func (u *UserTypeExpr) ToMap() *Map { return u.Type.ToMap() }

// IsCompatible returns true if u describes the (Go) type of val.
func (u *UserTypeExpr) IsCompatible(val interface{}) bool {
	return u.Type == nil || u.Type.IsCompatible(val)
}

// Walk traverses the data structure recursively and calls the given function once on each field
// starting with the field returned by u.Expr.
func (u *UserTypeExpr) Walk(walker func(*AttributeExpr) error) error {
	return walk(u.AttributeExpr, walker, map[string]bool{u.TypeName: true})
}
