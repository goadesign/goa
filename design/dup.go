package design

import "github.com/goadesign/goa/dslengine"

// Dup creates a copy the given data type.
func Dup(d DataType) DataType {
	return newDupper().DupType(d)
}

// DupAtt creates a copy of the given attribute.
func DupAtt(att *AttributeDefinition) *AttributeDefinition {
	return newDupper().DupAttribute(att)
}

// dupper implements recursive and cycle safe copy of data types.
type dupper struct {
	dts  map[string]*UserTypeDefinition
	dmts map[string]*MediaTypeDefinition
}

// newDupper returns a new initialized dupper.
func newDupper() *dupper {
	return &dupper{
		dts:  make(map[string]*UserTypeDefinition),
		dmts: make(map[string]*MediaTypeDefinition),
	}
}

// DupUserType creates a copy of the given user type.
func (d *dupper) DupUserType(ut *UserTypeDefinition) *UserTypeDefinition {
	return &UserTypeDefinition{
		AttributeDefinition: d.DupAttribute(ut.AttributeDefinition),
		TypeName:            ut.TypeName,
	}
}

// DupAttribute creates a copy of the given attribute.
func (d *dupper) DupAttribute(att *AttributeDefinition) *AttributeDefinition {
	var valDup *dslengine.ValidationDefinition
	if att.Validation != nil {
		valDup = att.Validation.Dup()
	}
	dup := AttributeDefinition{
		Type:              att.Type,
		Description:       att.Description,
		Validation:        valDup,
		Metadata:          att.Metadata,
		DefaultValue:      att.DefaultValue,
		NonZeroAttributes: att.NonZeroAttributes,
		View:              att.View,
		DSLFunc:           att.DSLFunc,
		Example:           att.Example,
	}
	return &dup
}

// DupType creates a copy of the given data type.
func (d *dupper) DupType(t DataType) DataType {
	switch actual := t.(type) {
	case Primitive:
		return t
	case *Array:
		return &Array{ElemType: d.DupAttribute(actual.ElemType)}
	case Object:
		res := make(Object, len(actual))
		for n, att := range actual {
			res[n] = d.DupAttribute(att)
		}
		return res
	case *Hash:
		return &Hash{
			KeyType:  d.DupAttribute(actual.KeyType),
			ElemType: d.DupAttribute(actual.ElemType),
		}
	case *UserTypeDefinition:
		if u, ok := d.dts[actual.TypeName]; ok {
			return u
		}
		u := &UserTypeDefinition{
			TypeName: actual.TypeName,
		}
		d.dts[u.TypeName] = u
		u.AttributeDefinition = d.DupAttribute(actual.AttributeDefinition)
		return u
	case *MediaTypeDefinition:
		if m, ok := d.dmts[actual.Identifier]; ok {
			return m
		}
		m := &MediaTypeDefinition{
			Identifier: actual.Identifier,
			Links:      actual.Links,
			Views:      actual.Views,
			Resource:   actual.Resource,
		}
		d.dmts[actual.Identifier] = m
		m.UserTypeDefinition = d.DupUserType(actual.UserTypeDefinition)
		return m
	}
	panic("unknown type " + t.Name())
}
