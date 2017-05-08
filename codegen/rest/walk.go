package restgen

import (
	"sort"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

type (
	// ActionWalker is the type of functions given to WalkActions.
	ActionWalker func(a *rest.ActionExpr) error

	// MappedAttributeWalker is the type of functions given to WalkHeaders
	// and WalkParams.
	// name is the name of the attribute, elem the name of the corresponding
	// HTTP element (header or parameter). required is true if the attribute
	// is required.
	MappedAttributeWalker func(name, elem string, required bool, a *design.AttributeExpr) error
)

// WalkMappedAttr iterates over the mapped attributes in alphabetical order. It
// calls the given function giving each attribute as it iterates. WalkMappedAttr
// stops if there is no more attribute to iterate over or if the iterator
// function returns an error in which case it returns the error.
func WalkMappedAttr(ma *rest.MappedAttributeExpr, it MappedAttributeWalker) error {
	o := design.AsObject(ma.Type)
	keys := make([]string, len(o))
	i := 0
	for key := range design.AsObject(ma.Type) {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := it(k, ma.ElemName(k), ma.IsRequired(k), o[k]); err != nil {
			return err
		}
	}
	return nil
}

// WalkHeaders iterates over the action and resource headers in alphabetical
// order. It calls the given function giving each header as it iterates.
// WalkHeaders stops if there is no more header to iterate over or if the
// iterator function returns an error in which case it returns the error.
func WalkHeaders(a *rest.ActionExpr, it MappedAttributeWalker) error {
	return walk(a, it, a.MappedHeaders(), a.Resource.MappedHeaders())
}

// WalkParams iterates over the action and resource parameters in alphabetical
// order. It calls the given function giving each parameter as it iterates.
// WalkParams stops if there is no more parameter to iterate over or if the
// iterator function returns an error in which case it returns the error.
func WalkParams(a *rest.ActionExpr, it MappedAttributeWalker) error {
	return walk(a, it, a.MappedParams(), a.Resource.MappedParams())
}

func walk(a *rest.ActionExpr, it MappedAttributeWalker, ma, rma *rest.MappedAttributeExpr) error {
	if ma == nil && rma == nil {
		return nil
	}

	var (
		merged    *rest.MappedAttributeExpr
		mergedMap design.Object
		elemNames []string
		nameMap   map[string]string
	)
	{
		if rma == nil {
			merged = ma
		} else if ma == nil {
			merged = rma
		} else {
			merged = rest.DupMappedAtt(rma)
			merged.Merge(ma)
		}

		mergedMap = merged.Type.(design.Object)
		nameMap = make(map[string]string, len(mergedMap))
		elemNames = make([]string, len(mergedMap))
		i := 0
		for n := range mergedMap {
			en := merged.ElemName(n)
			nameMap[en] = n
			elemNames[i] = en
			i++
		}
		sort.Strings(elemNames)
	}

	for _, n := range elemNames {
		attName := nameMap[n]
		header := mergedMap[attName]
		if err := it(attName, n, merged.IsRequired(attName), header); err != nil {
			return err
		}
	}
	return nil
}
