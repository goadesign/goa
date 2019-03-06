package expr

type (
	// UserTypeExpr describes user defined types.
	UserTypeExpr struct {
		// The embedded attribute expression.
		*AttributeExpr
		// Name of type
		TypeName string
	}
)

// NewUserTypeExpr creates a user type expression but does not execute the DSL.
func NewUserTypeExpr(name string, fn func()) *UserTypeExpr {
	return &UserTypeExpr{
		TypeName:      name,
		AttributeExpr: &AttributeExpr{DSLFunc: fn},
	}
}

// ID returns the identifier (type name) for the user type.
func (u *UserTypeExpr) ID() string {
	return u.Name()
}

// Kind implements DataKind.
func (u *UserTypeExpr) Kind() Kind { return UserTypeKind }

// Name returns the type name.
func (u *UserTypeExpr) Name() string {
	if u.AttributeExpr == nil {
		return u.TypeName
	}
	if n, ok := u.AttributeExpr.Meta["struct:type:name"]; ok {
		return n[0]
	}
	return u.TypeName
}

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

// SetAttribute sets the embedded attribute.
func (u *UserTypeExpr) SetAttribute(att *AttributeExpr) {
	u.AttributeExpr = att
}

// Dup creates a deep copy of the user type given a deep copy of its attribute.
func (u *UserTypeExpr) Dup(att *AttributeExpr) UserType {
	if u == Empty {
		// Don't dup Empty so that code may check against it.
		return u
	}
	return &UserTypeExpr{
		AttributeExpr: att,
		TypeName:      u.TypeName,
	}
}

// Hash returns a unique hash value for u.
func (u *UserTypeExpr) Hash() string {
	return "_type_+" + u.TypeName
}

// Example produces an example for the user type which is JSON serialization
// compatible.
func (u *UserTypeExpr) Example(r *Random) interface{} {
	if ex := u.recExample(r); ex != nil {
		return *ex
	}
	return nil
}

func (u *UserTypeExpr) recExample(r *Random) *interface{} {
	if ex, ok := r.Seen[u.ID()]; ok {
		return ex
	}
	if r.Seen == nil {
		r.Seen = make(map[string]*interface{})
	}
	var ex interface{}
	pex := &ex
	r.Seen[u.ID()] = pex
	actual := u.Type.Example(r)
	*pex = actual
	return pex
}
