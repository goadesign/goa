package rest

import (
	"strings"

	"goa.design/goa.v2/design"
)

// MappedAttributeExpr is an attribute expression of type object that map the
// object keys to names used in HTTP elements (e.g. header names, param names)
type MappedAttributeExpr struct {
	*design.AttributeExpr
	nameMap    map[string]string
	reverseMap map[string]string
}

// NewMappedAttributeExpr instantiates a mapped attribute expression for the
// given attribute. The type of att must be Object.
func NewMappedAttributeExpr(att *design.AttributeExpr) *MappedAttributeExpr {
	if att == nil {
		return NewMappedAttributeExpr(&design.AttributeExpr{Type: &design.Object{}})
	}
	if !design.IsObject(att.Type) {
		panic("cannot create a mapped attribute with a non object attribute") // bug
	}
	var (
		o          = design.AsObject(att.Type)
		n          = &design.Object{}
		nameMap    = make(map[string]string)
		reverseMap = make(map[string]string)
		validation *design.ValidationExpr
	)
	if att.Validation != nil {
		validation = att.Validation.Dup()
	}
	for _, nat := range *o {
		elems := strings.Split(nat.Name, ":")
		n.Set(elems[0], design.DupAtt(nat.Attribute))
		if len(elems) > 1 {
			nameMap[elems[0]] = elems[1]
			reverseMap[elems[1]] = elems[0]
		}
	}
	if ut, ok := att.Type.(design.UserType); ok {
		if val := ut.Attribute().Validation; val != nil {
			validation = val.Dup()
		}
	}
	return &MappedAttributeExpr{
		AttributeExpr: &design.AttributeExpr{
			Type:       n,
			Validation: validation,
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
		AttributeExpr: design.DupAtt(ma.AttributeExpr),
		nameMap:       nameMap,
		reverseMap:    reverseMap,
	}
}

// Map records the element name of one of the child attributes.
// Map panics if attName is not the name of a child attribute.
func (ma *MappedAttributeExpr) Map(elemName, attName string) {
	if att := design.AsObject(ma.Type).Attribute(attName); att == nil {
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
	ma.Type.(*design.Object).Delete(attName)
	if ma.Validation != nil {
		ma.Validation.RemoveRequired(attName)
	}
}

// Attribute returns the original attribute using "att:elem" format for the keys.
func (ma *MappedAttributeExpr) Attribute() *design.AttributeExpr {
	att := design.DupAtt(ma.AttributeExpr)
	obj := design.AsObject(att.Type)
	for _, nat := range *obj {
		if elem := ma.ElemName(nat.Name); elem != nat.Name {
			obj.Rename(nat.Name, nat.Name+":"+elem)
		}
	}
	return att
}

// ElemName returns the HTTP element name of the given object key. It returns
// keyName if it's a key of the mapped attribute object type. It panics if there
// is no mapping and keyName is not a key.
func (ma *MappedAttributeExpr) ElemName(keyName string) string {
	if n, ok := ma.nameMap[keyName]; ok {
		return n
	}
	if att := design.AsObject(ma.Type).Attribute(keyName); att != nil {
		return keyName
	}
	panic("Key " + keyName + " is not defined") // bug
}

// KeyName returns the object key of the given HTTP element name. It returns
// elemName if it's a key of the mapped attribute object type. It panics if
// there is no mapping and elemName is not a key.
func (ma *MappedAttributeExpr) KeyName(elemName string) string {
	if n, ok := ma.reverseMap[elemName]; ok {
		return n
	}
	if att := design.AsObject(ma.Type).Attribute(elemName); att != nil {
		return elemName
	}
	panic("HTTP element " + elemName + " is not defined and is not a key") // bug
}

// Merge merges other's attributes into a overriding attributes of a with
// attributes of other with identical names.
func (ma *MappedAttributeExpr) Merge(other *MappedAttributeExpr) {
	ma.AttributeExpr.Merge(other.AttributeExpr)
	for _, nat := range *design.AsObject(other.AttributeExpr.Type) {
		if en := other.ElemName(nat.Name); en != nat.Name {
			ma.Map(en, nat.Name)
		}
	}
}
