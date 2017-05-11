package restgen

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// GoTypeDef returns the Go code that defines the struct corresponding to ma.
func GoTypeDef(ma *rest.MappedAttributeExpr, public bool) string {
	obj := make(design.Object, len(design.AsObject(ma.Type)))
	var required []string
	WalkMappedAttr(ma, func(name, elem string, req bool, a *design.AttributeExpr) error {
		obj[elem] = a
		if req {
			required = append(required, elem)
		}
		return nil
	})
	att := design.DupAtt(ma.Attribute())
	att.Type = obj
	att.Validation = &design.ValidationExpr{Required: required}
	return codegen.GoTypeDef(att, public)
}
