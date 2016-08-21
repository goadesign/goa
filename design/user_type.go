package design

// NewUserTypeDefinition creates a user type definition but does not
// execute the DSL.
func NewUserTypeDefinition(name string, dsl func()) *UserTypeDefinition {
	return &UserTypeDefinition{
		TypeName:        name,
		FieldDefinition: &FieldDefinition{DSLFunc: dsl},
	}
}

// Kind implements DataKind.
func (u *UserTypeDefinition) Kind() Kind { return UserTypeKind }

// Name returns the JSON type name.
func (u *UserTypeDefinition) Name() string { return u.Type.Name() }

// IsPrimitive calls IsPrimitive on the user type underlying data type.
func (u *UserTypeDefinition) IsPrimitive() bool { return u.Type.IsPrimitive() }

// IsObject calls IsObject on the user type underlying data type.
func (u *UserTypeDefinition) IsObject() bool { return u.Type.IsObject() }

// IsArray calls IsArray on the user type underlying data type.
func (u *UserTypeDefinition) IsArray() bool { return u.Type.IsArray() }

// IsMap calls IsMap on the user type underlying data type.
func (u *UserTypeDefinition) IsMap() bool { return u.Type.IsMap() }

// ToObject calls ToObject on the user type underlying data type.
func (u *UserTypeDefinition) ToObject() Object { return u.Type.ToObject() }

// ToArray calls ToArray on the user type underlying data type.
func (u *UserTypeDefinition) ToArray() *Array { return u.Type.ToArray() }

// ToMap calls ToMap on the user type underlying data type.
func (u *UserTypeDefinition) ToMap() *Map { return u.Type.ToMap() }

// IsCompatible returns true if u describes the (Go) type of val.
func (u *UserTypeDefinition) IsCompatible(val interface{}) bool {
	return u.Type == nil || u.Type.IsCompatible(val)
}

// Walk traverses the data structure recursively and calls the given function once on each field
// starting with the field returned by u.Definition.
func (u *UserTypeDefinition) Walk(walker func(*FieldDefinition) error) error {
	return walk(u.FieldDefinition, walker, map[string]bool{u.TypeName: true})
}
