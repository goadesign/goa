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

// Walk traverses the data structure recursively and calls the given function
// once on each attribute starting with the user type attribute.
func (u *UserTypeExpr) Walk(walker func(*AttributeExpr) error) error {
	return walk(u.AttributeExpr, walker, map[string]bool{u.TypeName: true})
}

// Recursive implementation of the Walk methods. Takes care of avoiding infinite
// recursions by keeping track of types that have already been walked.
func walk(at *AttributeExpr, walker func(*AttributeExpr) error, seen map[string]bool) error {
	if err := walker(at); err != nil {
		return err
	}
	walkUt := func(ut *UserTypeExpr) error {
		if _, ok := seen[ut.TypeName]; ok {
			return nil
		}
		seen[ut.TypeName] = true
		return walk(ut.AttributeExpr, walker, seen)
	}
	switch actual := at.Type.(type) {
	case Primitive:
		return nil
	case *Array:
		return walk(actual.ElemType, walker, seen)
	case *Map:
		if err := walk(actual.KeyType, walker, seen); err != nil {
			return err
		}
		return walk(actual.ElemType, walker, seen)
	case Object:
		for _, cat := range actual {
			if err := walk(cat, walker, seen); err != nil {
				return err
			}
		}
	case *UserTypeExpr:
		return walkUt(actual)
	case *MediaTypeExpr:
		return walkUt(actual.UserTypeExpr)
	default:
		panic("unknown attribute type") // bug
	}
	return nil
}
