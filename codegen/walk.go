package codegen

import "goa.design/goa/v3/expr"

// MappedAttributeWalker is the type of functions given to WalkMappedAttr. name
// is the name of the attribute, elem the name of the corresponding transport
// element (e.g. HTTP header). required is true if the attribute is required.
type MappedAttributeWalker func(name, elem string, required bool, a *expr.AttributeExpr) error

// Walk traverses the data structure recursively and calls the given function
// once on each attribute starting with a.
func Walk(a *expr.AttributeExpr, walker func(*expr.AttributeExpr) error) error {
	return walk(a, walker, make(map[string]bool))
}

// WalkType traverses the data structure recursively and calls the given function
// once on each attribute starting with the user type attribute.
func WalkType(u expr.UserType, walker func(*expr.AttributeExpr) error) error {
	return walk(u.Attribute(), walker, map[string]bool{u.ID(): true})
}

// WalkMappedAttr iterates over the mapped attributes. It calls the given
// function giving each attribute as it iterates. WalkMappedAttr stops if there
// is no more attribute to iterate over or if the iterator function returns an
// error in which case it returns the error.
func WalkMappedAttr(ma *expr.MappedAttributeExpr, it MappedAttributeWalker) error {
	o := expr.AsObject(ma.Type)
	for _, nat := range *o {
		if err := it(nat.Name, ma.ElemName(nat.Name), ma.IsRequired(nat.Name), nat.Attribute); err != nil {
			return err
		}
	}
	return nil
}

// Recursive implementation of the Walk methods. Takes care of avoiding infinite
// recursions by keeping track of types that have already been walked.
func walk(at *expr.AttributeExpr, walker func(*expr.AttributeExpr) error, seen map[string]bool) error {
	if err := walker(at); err != nil {
		return err
	}
	walkUt := func(ut expr.UserType) error {
		if _, ok := seen[ut.ID()]; ok {
			return nil
		}
		seen[ut.ID()] = true
		return walk(ut.Attribute(), walker, seen)
	}
	switch actual := at.Type.(type) {
	case expr.Primitive:
		return nil
	case *expr.Array:
		return walk(actual.ElemType, walker, seen)
	case *expr.Map:
		if err := walk(actual.KeyType, walker, seen); err != nil {
			return err
		}
		return walk(actual.ElemType, walker, seen)
	case *expr.Union:
		for _, nat := range actual.Values {
			if err := walk(nat.Attribute, walker, seen); err != nil {
				return err
			}
		}
	case *expr.Object:
		for _, cat := range *actual {
			if err := walk(cat.Attribute, walker, seen); err != nil {
				return err
			}
		}
	case *expr.UserTypeExpr:
		return walkUt(actual)
	case *expr.ResultTypeExpr:
		return walkUt(actual.UserTypeExpr)
	default:
		panic("unknown attribute type") // bug
	}
	return nil
}
