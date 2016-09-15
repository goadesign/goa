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
}

// newDupper returns a new initialized dupper.
func newDupper() *dupper {
	return &dupper{make(map[string]UserType)}
}

// DupAttribute creates a copy of the given attribute.
func (d *dupper) DupAttribute(att *AttributeExpr) *AttributeExpr {
	var valDup *ValidationExpr
	if att.Validation != nil {
		valDup = att.Validation.Dup()
	}
	dup := AttributeExpr{
		Type:         att.Type,
		Description:  att.Description,
		Validation:   valDup,
		Metadata:     att.Metadata,
		DefaultValue: att.DefaultValue,
		DSLFunc:      att.DSLFunc,
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
