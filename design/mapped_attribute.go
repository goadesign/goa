package design

import "strings"

// MappedAttributeExpr is an attribute expression of type object that map the
// object keys to names used in transport specific elements (e.g. HTTP header
// names).
type MappedAttributeExpr struct {
	*AttributeExpr
	nameMap    map[string]string
	reverseMap map[string]string
}

// NewMappedAttributeExpr instantiates a mapped attribute expression for the
// given attribute. The type of att must be Object.
func NewMappedAttributeExpr(att *AttributeExpr) *MappedAttributeExpr {
	if att == nil {
		return NewMappedAttributeExpr(&AttributeExpr{Type: &Object{}})
	}
	if !IsObject(att.Type) {
		panic("cannot create a mapped attribute with a non object attribute") // bug
	}
	var (
		o          = AsObject(att.Type)
		n          = &Object{}
		nameMap    = make(map[string]string)
		reverseMap = make(map[string]string)
		validation *ValidationExpr
	)
	if att.Validation != nil {
		validation = att.Validation.Dup()
	}
	for _, nat := range *o {
		elems := strings.Split(nat.Name, ":")
		n.Set(elems[0], DupAtt(nat.Attribute))
		if len(elems) > 1 {
			nameMap[elems[0]] = elems[1]
			reverseMap[elems[1]] = elems[0]
		}
	}
	if validation == nil {
		if ut, ok := att.Type.(UserType); ok {
			if val := ut.Attribute().Validation; val != nil {
				validation = val.Dup()
			}
		}
	}
	return &MappedAttributeExpr{
		AttributeExpr: &AttributeExpr{
			Type:         n,
			References:   att.References,
			Bases:        att.Bases,
			Description:  att.Description,
			Docs:         att.Docs,
			Metadata:     att.Metadata,
			DefaultValue: att.DefaultValue,
			UserExamples: att.UserExamples,
			Validation:   validation,
		},
		nameMap:    nameMap,
		reverseMap: reverseMap,
	}
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
	ma.AttributeExpr.Merge(other.AttributeExpr)
	for _, nat := range *AsObject(other.AttributeExpr.Type) {
		if en := other.ElemName(nat.Name); en != nat.Name {
			ma.Map(en, nat.Name)
		}
	}
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
