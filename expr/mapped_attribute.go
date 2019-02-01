package expr

import (
	"strings"
)

type (
	// MappedAttributeExpr is an attribute expression of type object that map the
	// object keys to external names (e.g. HTTP header names).
	MappedAttributeExpr struct {
		*AttributeExpr
		nameMap    map[string]string
		reverseMap map[string]string
	}

	// MappedAttributeWalker is the type of functions given to WalkMappedAttr.
	//
	// name is the name of the attribute
	// elem the name of the corresponding transport element
	// a is the corresponding attribute expression
	//
	MappedAttributeWalker func(name, elem string, a *AttributeExpr) error
)

// NewEmptyMappedAttributeExpr creates an empty mapped attribute expression.
func NewEmptyMappedAttributeExpr() *MappedAttributeExpr {
	return NewMappedAttributeExpr(&AttributeExpr{Type: &Object{}})
}

// NewMappedAttributeExpr instantiates a mapped attribute expression for the
// given attribute. The type of att must be Object.
func NewMappedAttributeExpr(att *AttributeExpr) *MappedAttributeExpr {
	if att == nil {
		return NewEmptyMappedAttributeExpr()
	}
	if !IsObject(att.Type) {
		panic("cannot create a mapped attribute with a non object attribute") // bug
	}
	var (
		nameMap    = make(map[string]string)
		reverseMap = make(map[string]string)
		validation *ValidationExpr
	)
	if att.Validation != nil {
		validation = att.Validation.Dup()
	} else if ut, ok := att.Type.(UserType); ok {
		if val := ut.Attribute().Validation; val != nil {
			validation = val.Dup()
		}
	}
	ma := &MappedAttributeExpr{
		AttributeExpr: &AttributeExpr{
			Type:         Dup(att.Type),
			References:   att.References,
			Bases:        att.Bases,
			Description:  att.Description,
			Docs:         att.Docs,
			Meta:         att.Meta,
			DefaultValue: att.DefaultValue,
			UserExamples: att.UserExamples,
			Validation:   validation,
		},
		nameMap:    nameMap,
		reverseMap: reverseMap,
	}
	ma.Remap()
	return ma
}

// Remap recomputes the name mappings from the inner attribute. Use this if
// the underlying attribute is modified after the mapped attribute has been
// initially created.
func (ma *MappedAttributeExpr) Remap() {
	var (
		n = &Object{}
		o = AsObject(ma.Type)
	)
	for _, nat := range *o {
		elems := strings.Split(nat.Name, ":")
		n.Set(elems[0], nat.Attribute)
		if len(elems) > 1 {
			ma.nameMap[elems[0]] = elems[1]
			ma.reverseMap[elems[1]] = elems[0]
		}
	}
	ma.Type = n
}

// DupMappedAtt creates a deep copy of ma.
func DupMappedAtt(ma *MappedAttributeExpr) *MappedAttributeExpr {
	nameMap := make(map[string]string, len(ma.nameMap))
	reverseMap := make(map[string]string, len(ma.reverseMap))
	for k, v := range ma.nameMap {
		nameMap[k] = v
	}
	for k, v := range ma.reverseMap {
		reverseMap[k] = v
	}
	return &MappedAttributeExpr{
		AttributeExpr: DupAtt(ma.AttributeExpr),
		nameMap:       nameMap,
		reverseMap:    reverseMap,
	}
}

// WalkMappedAttr iterates over the mapped attributes. It calls the given
// function giving each attribute as it iterates. WalkMappedAttr stops if there
// is no more attribute to iterate over or if the iterator function returns an
// error in which case it returns the error.
func WalkMappedAttr(ma *MappedAttributeExpr, it MappedAttributeWalker) error {
	o := AsObject(ma.Type)
	for _, nat := range *o {
		if err := it(nat.Name, ma.ElemName(nat.Name), nat.Attribute); err != nil {
			return err
		}
	}
	return nil
}

// Map records the element name of one of the child attributes.
// Map panics if attName is not the name of a child attribute.
func (ma *MappedAttributeExpr) Map(elemName, attName string) {
	if att := AsObject(ma.Type).Attribute(attName); att == nil {
		panic(attName + " is not the name of a child of the mapped attribute") // bug
	}
	ma.nameMap[attName] = elemName
	ma.reverseMap[elemName] = attName
}

// Delete removes a child attribute given its name.
func (ma *MappedAttributeExpr) Delete(attName string) {
	delete(ma.nameMap, attName)
	for k, v := range ma.reverseMap {
		if v == attName {
			delete(ma.reverseMap, k)
			break
		}
	}
	ma.Type.(*Object).Delete(attName)
	if ma.Validation != nil {
		ma.Validation.RemoveRequired(attName)
	}
}

// Attribute returns the original attribute using "att:elem" format for the keys.
func (ma *MappedAttributeExpr) Attribute() *AttributeExpr {
	att := DupAtt(ma.AttributeExpr)
	obj := AsObject(att.Type)
	for _, nat := range *obj {
		if elem := ma.ElemName(nat.Name); elem != nat.Name {
			obj.Rename(nat.Name, nat.Name+":"+elem)
		}
	}
	return att
}

// ElemName returns the transport element name of the given object key. It
// returns keyName if it's a key of the mapped attribute object type. It panics
// if there is no mapping and keyName is not a key.
func (ma *MappedAttributeExpr) ElemName(keyName string) string {
	if n, ok := ma.nameMap[keyName]; ok {
		return n
	}
	if att := AsObject(ma.Type).Attribute(keyName); att != nil {
		return keyName
	}
	panic("Key " + keyName + " is not defined") // bug
}

// KeyName returns the object key of the given transport element name. It
// returns elemName if it's a key of the mapped attribute object type. It panics
// if there is no mapping and elemName is not a key.
func (ma *MappedAttributeExpr) KeyName(elemName string) string {
	if n, ok := ma.reverseMap[elemName]; ok {
		return n
	}
	if att := AsObject(ma.Type).Attribute(elemName); att != nil {
		return elemName
	}
	panic("transport element " + elemName + " is not defined and is not a key") // bug
}

// Merge merges other's attributes into a overriding attributes of a with
// attributes of other with identical names.
func (ma *MappedAttributeExpr) Merge(other *MappedAttributeExpr) {
	if other == nil {
		return
	}
	ma.AttributeExpr.Merge(other.Attribute())
	ma.Remap()
}

// FindKey finds the given key in the mapped attribute expression.
// If key is found, it returns the transport element name of the key and true.
// Otherwise, it returns an empty string and false.
func (ma *MappedAttributeExpr) FindKey(keyName string) (string, bool) {
	obj := AsObject(ma.Type)
	for _, nat := range *obj {
		if nat.Name == keyName {
			return ma.ElemName(keyName), true
		}
	}
	return "", false
}

// IsEmpty returns true if the mapped attribute contains no key.
func (ma *MappedAttributeExpr) IsEmpty() bool {
	return len(*ma.Type.(*Object)) == 0
}
