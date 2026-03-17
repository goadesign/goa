package internal

import (
	"strings"

	"goa.design/goa/v3/expr"
)

// IsSecurityParameter returns true if the given HTTP transport element is used
// by one of the endpoint security schemes and should therefore not be emitted
// again as a regular OpenAPI parameter.
func IsSecurityParameter(endpoint *expr.HTTPEndpointExpr, in, name string) bool {
	if endpoint == nil {
		return false
	}
	for _, req := range endpoint.Requirements {
		for _, scheme := range req.Schemes {
			if scheme.In != in {
				continue
			}
			if in == "header" {
				if strings.EqualFold(scheme.Name, name) {
					return true
				}
				continue
			}
			if scheme.Name == name {
				return true
			}
		}
	}
	return false
}
