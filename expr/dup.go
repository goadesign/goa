package expr

import "fmt"

// Dup creates a copy the given data type.
func Dup(d DataType) DataType {
	res := newDupper().DupType(d)
	if rt, ok := d.(*ResultTypeExpr); ok {
		if Root.GeneratedResultType(rt.Identifier) != nil {
			*Root.GeneratedTypes = append(*Root.GeneratedTypes, res.(*ResultTypeExpr))
		}
	}
	return res
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
	var metaDup MetaExpr
	if att.Meta != nil {
		metaDup = att.Meta.Dup()
	}
	dup := AttributeExpr{
		Type:         d.DupType(att.Type),
		Description:  att.Description,
		References:   att.References,
		Bases:        att.Bases,
		Validation:   valDup,
		Meta:         metaDup,
		DefaultValue: att.DefaultValue,
		ZeroValue:    att.ZeroValue,
		DSLFunc:      att.DSLFunc,
		UserExamples: att.UserExamples,
	}
	d.ats[&dup] = struct{}{}
	return &dup
}

// DupType creates a copy of the given data type.
func (d *dupper) DupType(t DataType) DataType {
	if t == Empty {
		// Don't dup Empty so that code may check against it.
		return t
	}
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
		if u, ok := d.uts[actual.ID()]; ok {
			return u
		}
		dp := actual.Dup(nil)
		d.uts[actual.ID()] = dp
		dupAtt := d.DupAttribute(actual.Attribute())
		dp.SetAttribute(dupAtt)
		return dp
	}
	panic("unknown type " + fmt.Sprintf("%T", t))
}
