package design

// Dup creates a copy the given data type.
func Dup(d DataType) DataType {
	return newDupper().DupType(d)
}

// DupAtt creates a copy of the given attribute.
func DupAtt(att *AttributeExpr) *AttributeExpr {
	return newDupper().DupAttribute(att)
}

// dupper implements recursive and cycle safe copy of data types.
type dupper struct {
	uts map[string]UserType
	ats map[*AttributeExpr]struct{}
}

// newDupper returns a new initialized dupper.
func newDupper() *dupper {
	return &dupper{
		uts: make(map[string]UserType),
		ats: make(map[*AttributeExpr]struct{}),
	}
}

// DupAttribute creates a copy of the given attribute.
func (d *dupper) DupAttribute(att *AttributeExpr) *AttributeExpr {
	if _, ok := d.ats[att]; ok {
		return att
	}
	var valDup *ValidationExpr
	if att.Validation != nil {
		valDup = att.Validation.Dup()
	}
	dup := AttributeExpr{
		Type:         d.DupType(att.Type),
		Description:  att.Description,
		Validation:   valDup,
		Metadata:     att.Metadata,
		DefaultValue: att.DefaultValue,
		DSLFunc:      att.DSLFunc,
	}
	d.ats[&dup] = struct{}{}
	return &dup
}

// DupType creates a copy of the given data type.
func (d *dupper) DupType(t DataType) DataType {
	switch actual := t.(type) {
	case Primitive:
		return t
	case *Array:
		return &Array{ElemType: d.DupAttribute(actual.ElemType)}
	case *Object:
		res := &Object{}
		for _, nat := range *actual {
			res.Set(nat.Name, d.DupAttribute(nat.Attribute))
		}
		return res
	case *Map:
		return &Map{
			KeyType:  d.DupAttribute(actual.KeyType),
			ElemType: d.DupAttribute(actual.ElemType),
		}
	case UserType:
		if u, ok := d.uts[actual.Name()]; ok {
			return u
		}
		return actual.Dup(d.DupAttribute(actual.Attribute()))
	}
	panic("unknown type " + t.Name())
}
