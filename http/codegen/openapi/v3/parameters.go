package openapiv3

import (
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

// paramsFromPath computes the OpenAPI spec parameters for the given API,
// service or endpoint HTTP path and query parameters.
func paramsFromPath(params *expr.MappedAttributeExpr, path string, rand *expr.Random) []*Parameter {
	var (
		res       []*Parameter
		wildcards = expr.ExtractHTTPWildcards(path)
	)
	codegen.WalkMappedAttr(params, func(n, pn string, required bool, at *expr.AttributeExpr) error {
		in := "query"
		for _, w := range wildcards {
			if n == w {
				in = "path"
				required = true
				break
			}
		}
		res = append(res, paramFor(at, pn, in, required, rand))
		return nil
	})
	return res
}

// paramsFromHeadersAndCookies computes the OpenAPI spec parameters for the
// given endpoint HTTP headers and cookies.
func paramsFromHeadersAndCookies(endpoint *expr.HTTPEndpointExpr, rand *expr.Random) []*Parameter {
	params := []*Parameter{}
	expr.WalkMappedAttr(endpoint.Headers, func(name, elem string, att *expr.AttributeExpr) error {
		if strings.ToLower(elem) == "authorization" {
			// Headers named "Authorization" are ignored by OpenAPI v3.
			// Instead it uses the security and securitySchemes sections to
			// define authorization.
			return nil
		}
		required := endpoint.Headers.IsRequiredNoDefault(name)
		params = append(params, paramFor(att, elem, "header", required, rand))
		return nil
	})
	expr.WalkMappedAttr(endpoint.Cookies, func(name, elem string, att *expr.AttributeExpr) error {
		required := endpoint.Cookies.IsRequiredNoDefault(name)
		params = append(params, paramFor(att, elem, "cookie", required, rand))
		return nil
	})

	return params
}

// paramFor converts the given attribute into a OpenAPI spec parameter.
func paramFor(att *expr.AttributeExpr, name, in string, required bool, rand *expr.Random) *Parameter {
	param := &Parameter{
		Name:            name,
		In:              in,
		Description:     att.Description,
		AllowEmptyValue: in != "path",
		Required:        required,
		Schema:          newSchemafier(rand).schemafy(att),
		Extensions:      openapi.ExtensionsFromExpr(att.Meta),
	}
	initExamples(param, att, rand)
	return param
}
