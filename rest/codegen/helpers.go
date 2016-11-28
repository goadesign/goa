package codegen

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/rest/design"
)

// CanonicalTemplate returns the resource URI template as a format string suitable for use in the
// fmt.Printf function family.
func CanonicalTemplate(r *design.ResourceExpr) string {
	return design.WildcardRegex.ReplaceAllLiteralString(r.URITemplate(), "/%v")
}

// CanonicalParams returns the list of parameter names needed to build the canonical href to the
// resource. It returns nil if the resource does not have a canonical action.
func CanonicalParams(r *design.ResourceExpr) []string {
	var params []string
	if ca := r.CanonicalAction(); ca != nil {
		if len(ca.Routes) > 0 {
			params = ca.Routes[0].Params()
		}
		for i, p := range params {
			params[i] = codegen.Goify(p, false)
		}
	}
	return params
}
