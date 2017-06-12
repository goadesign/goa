package codegen

import (
	"sort"

	"goa.design/goa.v2/design"
)

// AttributeWalker is the type of the function given to WalkAttributes.
type AttributeWalker func(string, *design.AttributeExpr) error

// WalkAttributes calls the given iterator passing in each field sorted in
// alphabetical order. Iteration stops if an iterator returns an error and in
// this case WalkObject returns that error.
func WalkAttributes(o design.Object, it AttributeWalker) error {
	names := make([]string, len(o))
	i := 0
	for n := range o {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(n, o[n]); err != nil {
			return err
		}
	}
	return nil
}

// Walk traverses the data structure recursively and calls the given function
// once on each attribute starting with a.
func Walk(a *design.AttributeExpr, walker func(*design.AttributeExpr) error) error {
	return walk(a, walker, make(map[string]bool))
}

// WalkType traverses the data structure recursively and calls the given function
// once on each attribute starting with the user type attribute.
func WalkType(u design.UserType, walker func(*design.AttributeExpr) error) error {
	return walk(u.Attribute(), walker, map[string]bool{u.Name(): true})
}

// Recursive implementation of the Walk methods. Takes care of avoiding infinite
// recursions by keeping track of types that have already been walked.
func walk(at *design.AttributeExpr, walker func(*design.AttributeExpr) error, seen map[string]bool) error {
	if err := walker(at); err != nil {
		return err
	}
	walkUt := func(ut design.UserType) error {
		if _, ok := seen[ut.Name()]; ok {
			return nil
		}
		seen[ut.Name()] = true
		return walk(ut.Attribute(), walker, seen)
	}
	switch actual := at.Type.(type) {
	case design.Primitive:
		return nil
	case *design.Array:
		return walk(actual.ElemType, walker, seen)
	case *design.Map:
		if err := walk(actual.KeyType, walker, seen); err != nil {
			return err
		}
		return walk(actual.ElemType, walker, seen)
	case design.Object:
		for _, cat := range actual {
			if err := walk(cat, walker, seen); err != nil {
				return err
			}
		}
	case *design.UserTypeExpr:
		return walkUt(actual)
	case *design.ResultTypeExpr:
		return walkUt(actual.UserTypeExpr)
	default:
		panic("unknown attribute type") // bug
	}
	return nil
}
