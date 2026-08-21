// This file defines user-authored type declarations and the distinction
// between their stable semantic IDs and in-memory copy provenance.
package expr

type (
	// UserTypeExpr describes user defined types. While a given design must
	// ensure that the names are unique the code used to generate code can
	// create multiple user types that share the same name (for example because
	// generated in different packages). When supplied, UID is a stable semantic
	// identifier used by authored examples and media-type behavior; generated
	// types retain a separate opaque example owner. Origin identifies copied
	// in-memory declarations.
	UserTypeExpr struct {
		// The embedded attribute expression.
		*AttributeExpr
		// Name of type
		TypeName string
		// UID is the optional stable semantic identifier of the type.
		UID string
		// origin is the earliest declaration copied to create this type.
		origin UserType
		// exampleIdentity is the semantic owner of a type synthesized by a
		// transport generator. Authored types leave it empty and use ID.
		exampleIdentity ExampleIdentity
	}
)

// NewGeneratedUserType creates a synthesized user type whose stable ID and
// examples are derived from identity. Code generators use this constructor so
// a copied wire type cannot accidentally inherit an authored type's identity.
func NewGeneratedUserType(name string, attribute *AttributeExpr, identity ExampleIdentity) *UserTypeExpr {
	if identity.seed == "" {
		panic("generated user type requires an example identity")
	}
	return &UserTypeExpr{
		AttributeExpr:   attribute,
		TypeName:        name,
		UID:             "generated:" + identity.Seed(),
		exampleIdentity: identity,
	}
}

// ID returns the unique identifier for the user type.
func (u *UserTypeExpr) ID() string {
	if u.UID != "" {
		return u.UID
	}
	return u.Name()
}

// Origin returns the earliest user type declaration from which u was copied.
func (u *UserTypeExpr) Origin() UserType {
	if u.origin != nil {
		return u.origin
	}
	return u
}

// Kind implements DataKind.
func (*UserTypeExpr) Kind() Kind { return UserTypeKind }

// Name returns the type name.
func (u *UserTypeExpr) Name() string {
	if u.AttributeExpr == nil {
		return u.TypeName
	}
	if n, ok := u.Meta["struct:type:name"]; ok {
		return n[0]
	}
	return u.TypeName
}

// Rename changes the type name to the given value.
func (u *UserTypeExpr) Rename(n string) {
	// Remember original name for example to generate friendly docs.
	u.AddMeta("name:original", u.TypeName)
	delete(u.Meta, "struct:type:name")
	u.TypeName = n
	u.origin = nil
}

// IsCompatible returns true if u describes the (Go) type of val.
func (u *UserTypeExpr) IsCompatible(val any) bool {
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
		AttributeExpr:   att,
		TypeName:        u.TypeName,
		UID:             u.UID,
		origin:          u.Origin(),
		exampleIdentity: u.exampleIdentity,
	}
}

// Hash returns a unique hash value for u.
func (u *UserTypeExpr) Hash() string {
	return Hash(u, true, false, true)
}

// Example produces an example for the user type which is JSON serialization
// compatible.
func (u *UserTypeExpr) Example(r *ExampleGenerator) any {
	if ex := u.recExample(r); ex != nil {
		return *ex
	}
	return nil
}

func (u *UserTypeExpr) recExample(r *ExampleGenerator) *any {
	if ex, ok := r.previouslySeen(u); ok {
		return ex
	}
	var ex any
	pex := &ex
	r.haveSeen(u, pex)
	// Anchor the value stream to the type identity so the example depends
	// only on the type definition, not on how many examples were computed
	// before it nor on which design path reached the type first.
	actual := u.AttributeExpr.Example(r.At(UserTypeExampleIdentity(u)))
	*pex = actual
	return pex
}
