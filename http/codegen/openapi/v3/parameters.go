package openapiv3

import (
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
		required := endpoint.Headers.IsRequiredNoDefault(name)
		params = append(params, paramFor(att, elem, "header", required, rand))
		return nil
	})
	expr.WalkMappedAttr(endpoint.Cookies, func(name, elem string, att *expr.AttributeExpr) error {
		required := endpoint.Cookies.IsRequiredNoDefault(name)
		params = append(params, paramFor(att, elem, "cookie", required, rand))
		return nil
	})

	// Add basic auth to headers
	if att := expr.TaggedAttribute(endpoint.MethodExpr.Payload, "security:username"); att != "" {
		// Basic Auth is always encoded in the Authorization header
		// https://golang.org/pkg/net/http/#Request.SetBasicAuth
		s := openapi.NewSchema()
		s.Type = openapi.Type("string")
		params = append(params, &Parameter{
			Name:            "Authorization",
			In:              "header",
			Description:     "Basic Auth security using Basic scheme (https://tools.ietf.org/html/rfc7617)",
			AllowEmptyValue: false,
			Required:        endpoint.MethodExpr.Payload.IsRequired(att),
			Schema:          s,
			Example:         "Basic Z29hOmRlc2lnbg==",
		})
	}
	return params
}

// paramFor converts the given attribute into a OpenAPI spec parameter.
func paramFor(att *expr.AttributeExpr, name, in string, required bool, rand *expr.Random) *Parameter {
	return &Parameter{
		Name:            name,
		In:              in,
		Description:     att.Description,
		AllowEmptyValue: in != "path",
		Required:        required,
		Schema:          newSchemafier(rand).schemafy(att),
		Example:         att.Example(rand),
		Extensions:      openapi.ExtensionsFromExpr(att.Meta),
	}
}
