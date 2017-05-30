package rest

import (
	"sort"
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
		return NewMappedAttributeExpr(&design.AttributeExpr{Type: make(design.Object)})
	}
	if !design.IsObject(att.Type) {
		panic("cannot create a mapped attribute with a non object attribute") // bug
	}
	var (
		o          = design.AsObject(att.Type)
		n          = make(design.Object, len(o))
		nameMap    = make(map[string]string)
		reverseMap = make(map[string]string)
	)
	for k, v := range o {
		elems := strings.Split(k, ":")
		n[elems[0]] = v
		if len(elems) > 1 {
			nameMap[elems[0]] = elems[1]
			reverseMap[elems[1]] = elems[0]
		}
	}
	return &MappedAttributeExpr{
		AttributeExpr: &design.AttributeExpr{
			Type:       n,
			Validation: att.Validation,
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
	if _, ok := design.AsObject(ma.Type)[attName]; !ok {
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
	delete(ma.Type.(design.Object), attName)
	if ma.Validation != nil {
		if req := ma.Validation.Required; len(req) > 0 {
			for i, r := range req {
				if r == attName {
					ma.Validation.Required = append(req[:i], req[i+1:len(req)]...)
					break
				}
			}
		}
	}
}

// Attribute returns the original attribute using "att:elem" format for the keys.
func (ma *MappedAttributeExpr) Attribute() *design.AttributeExpr {
	att := design.DupAtt(ma.AttributeExpr)
	obj := design.AsObject(att.Type)
	for k, v := range obj {
		if elem := ma.ElemName(k); elem != k {
			delete(obj, k)
			obj[k+":"+elem] = v
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
	if _, ok := design.AsObject(ma.Type)[keyName]; ok {
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
	if _, ok := design.AsObject(ma.Type)[elemName]; ok {
		return elemName
	}
	panic("HTTP element " + elemName + " is not defined and is not a key") // bug
}

// Keys returns the attribute keys sorted alphabetically.
func (ma *MappedAttributeExpr) Keys() []string {
	o := design.AsObject(ma.Type)
	keys := make([]string, len(o))
	i := 0
	for key := range design.AsObject(ma.Type) {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// Merge merges other's attributes into a overriding attributes of a with
// attributes of other with identical names.
func (ma *MappedAttributeExpr) Merge(other *MappedAttributeExpr) {
	ma.AttributeExpr.Merge(other.AttributeExpr)
	for n := range design.AsObject(other.AttributeExpr.Type) {
		if en := other.ElemName(n); en != n {
			ma.Map(en, n)
		}
	}
}
