// This file implements inheritance resolution and finalization for Goa attributes.
// It applies type references and resolves base type merges in dependency order
// before traversing child value attributes, preventing premature finalization
// across circular or back-edge type graphs.
package expr

// Finalize applies reference type attributes, merges base type attributes in
// dependency order, and finalizes child value attributes.
func (a *AttributeExpr) Finalize() {
	if a.finalized {
		return // Avoid infinite recursion.
	}
	a.resolveInheritance(make(map[*AttributeExpr]struct{}))
	if a.finalized {
		return
	}
	a.finalized = true
	if ut, ok := a.Type.(UserType); ok {
		ut.Finalize()
	}
	switch {
	case IsObject(a.Type):
		for _, nat := range *AsObject(a.Type) {
			nat.Attribute.Finalize()
		}
	case IsUnion(a.Type):
		for _, nat := range AsUnion(a.Type).Values {
			nat.Attribute.Finalize()
		}
	case IsArray(a.Type):
		AsArray(a.Type).ElemType.Finalize()
	case IsMap(a.Type):
		m := AsMap(a.Type)
		m.ElemType.Finalize()
		m.KeyType.Finalize()
	}
}

// Merge merges other's attributes into a overriding attributes of a with
// attributes of other with identical names.
//
// This only applies to attributes of type Object and Merge panics if the
// argument or the target is not of type Object.
func (a *AttributeExpr) Merge(other *AttributeExpr) {
	if other == nil {
		return
	}
	left := AsObject(a.Type)
	right := AsObject(other.Type)
	if left == nil || right == nil {
		panic("cannot merge non object attributes") // bug
	}
	if a.Type == Empty && len(*right) > 0 {
		a.Type = &Object{}
		left = AsObject(a.Type)
	}
	if other.Validation != nil {
		if a.Validation == nil {
			a.Validation = other.Validation.Dup()
		} else {
			a.Validation.Merge(other.Validation)
		}
	}
	for _, nat := range *right {
		left.Set(nat.Name, nat.Attribute)
	}
}

// Inherit merges the properties of existing target type attributes with the
// argument's. The algorithm is recursive so that child attributes are also
// merged.
func (a *AttributeExpr) Inherit(parent *AttributeExpr) {
	if !a.shouldInherit(parent) {
		return
	}
	pobj := AsObject(parent.Type)
	if a.Type == Empty && len(*pobj) > 0 {
		a.Type = &Object{}
	}
	a.inheritValidations(parent)
	a.inheritRecursive(parent, make(map[*AttributeExpr]struct{}))
}

// resolveInheritance applies References and merges Bases in dependency order
// before Finalize traverses child value attributes. Separating base-type
// inheritance resolution from value traversal ensures that a base attribute
// whose fields reference a downstream consumer does not finalize that consumer
// before intermediate base types have merged their own bases.
func (a *AttributeExpr) resolveInheritance(resolving map[*AttributeExpr]struct{}) {
	if a == nil || a.finalized || a.inheritanceResolved {
		return
	}
	if _, ok := resolving[a]; ok {
		return
	}
	resolving[a] = struct{}{}
	defer delete(resolving, a)

	if ut, ok := a.Type.(UserType); ok && ut.Attribute() != nil {
		ut.Attribute().resolveInheritance(resolving)
	}
	if IsObject(a.Type) {
		for _, ref := range a.References {
			if ru, ok := ref.(UserType); ok {
				a.Inherit(ru.Attribute())
			}
		}
		for _, base := range a.Bases {
			ru, ok := base.(UserType)
			if !ok || ru.Attribute() == nil {
				continue
			}
			ru.Attribute().resolveInheritance(resolving)
			a.Merge(ru.Attribute())
		}
		// Now that we've merged the bases, we can clear them. This
		// avoids issues where the bases are dupped and modifications
		// made to the originals are not reflected in the dups.
		a.Bases = nil
	}
	a.inheritanceResolved = true
}

// inheritRecursive recursively copies descriptions, validations, default values,
// and types from matching fields on parent into a.
func (a *AttributeExpr) inheritRecursive(parent *AttributeExpr, seen map[*AttributeExpr]struct{}) {
	if !a.shouldInherit(parent) {
		return
	}
	for _, nat := range *AsObject(a.Type) {
		if patt := AsObject(parent.Type).Attribute(nat.Name); patt != nil {
			att := nat.Attribute
			if att.Description == "" {
				att.Description = patt.Description
			}
			att.inheritValidations(patt)
			if att.DefaultValue == nil {
				att.DefaultValue = patt.DefaultValue
			}
			if att.Type == nil {
				att.Type = patt.Type
			} else if att.shouldInherit(patt) {
				if _, ok := seen[att]; ok {
					continue
				}
				seen[att] = struct{}{}
				for _, nat := range *AsObject(att.Type) {
					child := nat.Attribute
					parent := AsObject(patt.Type).Attribute(nat.Name)
					if parent != nil {
						child.inheritValidations(parent)
						child.inheritRecursive(parent, seen)
					}
				}
			}
		}
	}
}

// inheritValidations copies required validation rules from parent into a.
func (a *AttributeExpr) inheritValidations(parent *AttributeExpr) {
	if parent.Validation == nil {
		return
	}
	if a.Validation == nil {
		a.Validation = &ValidationExpr{}
	}
	a.Validation.AddRequired(parent.Validation.Required...)
}

// shouldInherit reports whether both a and parent are non-nil objects that can
// participate in inheritance.
func (a *AttributeExpr) shouldInherit(parent *AttributeExpr) bool {
	return a != nil && AsObject(a.Type) != nil &&
		parent != nil && AsObject(parent.Type) != nil
}
