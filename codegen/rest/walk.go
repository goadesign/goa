package rest

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// EndpointWalker is the type of functions given to WalkEndpoints.
type EndpointWalker func(a *rest.HTTPEndpointExpr) error

// WalkHeaders iterates over the endpoint and service headers in alphabetical
// order. It calls the given function giving each header as it iterates.
// WalkHeaders stops if there is no more header to iterate over or if the
// iterator function returns an error in which case it returns the error.
func WalkHeaders(a *rest.HTTPEndpointExpr, it codegen.MappedAttributeWalker) error {
	return walk(a, it, a.MappedHeaders(), a.Service.MappedHeaders())
}

// WalkParams iterates over the endpoint and service parameters in alphabetical
// order. It calls the given function giving each parameter as it iterates.
// WalkParams stops if there is no more parameter to iterate over or if the
// iterator function returns an error in which case it returns the error.
func WalkParams(a *rest.HTTPEndpointExpr, it codegen.MappedAttributeWalker) error {
	return walk(a, it, a.MappedParams(), a.Service.MappedParams())
}

// type MappedAttributeWalker func(name, elem string, required bool, a *design.AttributeExpr) error
func walk(a *rest.HTTPEndpointExpr, it codegen.MappedAttributeWalker, ma, rma *design.MappedAttributeExpr) error {
	if ma == nil && rma == nil {
		return nil
	}

	var (
		merged *design.MappedAttributeExpr
	)
	{
		if rma == nil {
			merged = ma
		} else if ma == nil {
			merged = rma
		} else {
			merged = design.DupMappedAtt(rma)
			merged.Merge(ma)
		}
	}

	for _, n := range *design.AsObject(merged.Type) {
		if err := it(n.Name, merged.ElemName(n.Name), merged.IsRequired(n.Name), n.Attribute); err != nil {
			return err
		}
	}
	return nil
}
