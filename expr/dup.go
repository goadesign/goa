// This file copies design types while preserving which original declaration
// each copied type came from.
package expr

import (
	"fmt"
)

// Dup creates a copy the given data type.
func Dup(d DataType) DataType {
	return newDupper().DupType(d)
}

// DupForDSL creates a copy of the given data type and registers copied result
// types so Goa evaluates their DSL.
func DupForDSL(dataType DataType) DataType {
	dupper := newDSLDupper()
	result := dupper.DupType(dataType)
	dupper.registerResultTypes()
	return result
}

// DupAtt creates a copy of the given attribute without changing the design.
func DupAtt(att *AttributeExpr) *AttributeExpr {
	return newDupper().DupAttribute(att)
}

// DupAttForDSL creates a copy of the given attribute and registers copied
// result types so Goa evaluates their DSL.
func DupAttForDSL(att *AttributeExpr) *AttributeExpr {
	dupper := newDSLDupper()
	result := dupper.DupAttribute(att)
	dupper.registerResultTypes()
	return result
}

// dupper implements recursive and cycle safe copy of data types.
type dupper struct {
	uts              map[UserType]UserType
	ats              map[*AttributeExpr]struct{}
	registerTypes    bool
	resultTypeCopies []*ResultTypeExpr
}

// newDupper returns a copier that does not change the evaluated design.
func newDupper() *dupper {
	return &dupper{
		uts: make(map[UserType]UserType),
		ats: make(map[*AttributeExpr]struct{}),
	}
}

// newDSLDupper returns a copier that records generated result types whose DSL
// must run before evaluation is complete.
func newDSLDupper() *dupper {
	dupper := newDupper()
	dupper.registerTypes = true
	return dupper
}

// DupAttribute creates a copy of the given attribute with one shared type map
// so repeated and recursive types remain shared in the copy.
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
	var duppedBases []DataType
	if len(att.Bases) > 0 {
		duppedBases = make([]DataType, len(att.Bases))
		for i, b := range att.Bases {
			duppedBases[i] = d.DupType(b)
		}
	}
	var duppedRefs []DataType
	if len(att.References) > 0 {
		duppedRefs = make([]DataType, len(att.References))
		for i, r := range att.References {
			duppedRefs[i] = d.DupType(r)
		}
	}
	dup := AttributeExpr{
		Type:                d.DupType(att.Type),
		Description:         att.Description,
		References:          duppedRefs,
		Bases:               duppedBases,
		Validation:          valDup,
		Meta:                metaDup,
		DefaultValue:        att.DefaultValue,
		DSLFunc:             att.DSLFunc,
		UserExamples:        att.UserExamples,
		inheritanceResolved: att.inheritanceResolved,
		finalized:           att.finalized,
		authored:            att.AuthoredAttribute(),
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
		return &Array{
			ElemType:         d.DupAttribute(actual.ElemType),
			NonNullableElems: actual.NonNullableElems,
		}
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
	case *Union:
		dp := Union{
			TypeName: actual.TypeName,
			Values:   make([]*NamedAttributeExpr, len(actual.Values)),
			TypeKey:  actual.TypeKey,
			ValueKey: actual.ValueKey,
			Flatten:  actual.Flatten,
			Untagged: actual.Untagged,
		}
		for i, nat := range actual.Values {
			dp.Values[i] = &NamedAttributeExpr{Name: nat.Name, Attribute: d.DupAttribute(nat.Attribute)}
		}
		return &dp
	case UserType:
		origin := actual.Origin()
		if u, ok := d.uts[origin]; ok {
			return u
		}
		dp := actual.Dup(nil)
		d.uts[origin] = dp
		dupAtt := d.DupAttribute(actual.Attribute())
		dp.SetAttribute(dupAtt)

		// DSL copies must be evaluated because their DSL may define views
		// used by the attribute that contains the copy.
		if rt, ok := dp.(*ResultTypeExpr); d.registerTypes && ok {
			if GeneratedResultType(rt.Identifier) != nil {
				d.resultTypeCopies = append(d.resultTypeCopies, rt)
			}
		}

		return dp
	}
	panic("unknown type " + fmt.Sprintf("%T", t))
}

// registerResultTypes adds every generated result type found in a DSL copy
// after the complete graph has been copied.
func (d *dupper) registerResultTypes() {
	for _, resultType := range d.resultTypeCopies {
		GeneratedResultTypes.Append(resultType)
	}
}
