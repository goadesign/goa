// This file converts HTTP parameters, headers, and cookies into OpenAPI v3
// values. Each schema and its displayed example use the same new repeatable
// example key.
package openapiv3

import (
	"slices"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiinternal "goa.design/goa/v3/http/codegen/openapi/internal"
)

// paramsFromPath computes the OpenAPI spec parameters for the given endpoint
// HTTP path and query parameters.
func paramsFromPath(endpoint *expr.HTTPEndpointExpr, path string, rand *expr.ExampleGenerator, values openapi.Values) []*Parameter {
	var (
		res       []*Parameter
		params    = endpoint.Params
		wildcards = expr.ExtractHTTPWildcards(path)
		owner     = expr.MethodPayloadExampleIdentity(endpoint.MethodExpr)
	)
	codegen.WalkMappedAttr(params, func(n, pn string, required bool, at *expr.AttributeExpr) error { // nolint: errcheck
		in := "query"
		if slices.Contains(wildcards, n) {
			in = "path"
			required = true
		}
		if in != "path" && openapiinternal.IsSecurityParameter(endpoint, in, pn) {
			return nil
		}
		identity := exampleFieldIdentity(endpoint.MethodExpr.Payload, n, owner)
		res = append(res, paramFor(at, pn, in, required, rand, identity, values))
		return nil
	})
	return res
}

// paramsFromHeadersAndCookies computes the OpenAPI spec parameters for the
// given endpoint HTTP headers and cookies.
func paramsFromHeadersAndCookies(endpoint *expr.HTTPEndpointExpr, rand *expr.ExampleGenerator, values openapi.Values) []*Parameter {
	var params []*Parameter
	owner := expr.MethodPayloadExampleIdentity(endpoint.MethodExpr)

	expr.WalkMappedAttr(endpoint.Headers, func(name, elem string, att *expr.AttributeExpr) error { // nolint: errcheck
		if openapiinternal.IsSecurityParameter(endpoint, "header", elem) {
			return nil
		}
		required := endpoint.Headers.IsRequiredNoDefault(name)
		identity := exampleFieldIdentity(endpoint.MethodExpr.Payload, name, owner)
		params = append(params, paramFor(att, elem, "header", required, rand, identity, values))
		return nil
	})
	expr.WalkMappedAttr(endpoint.Cookies, func(name, elem string, att *expr.AttributeExpr) error { // nolint: errcheck
		if openapiinternal.IsSecurityParameter(endpoint, "cookie", elem) {
			return nil
		}
		required := endpoint.Cookies.IsRequiredNoDefault(name)
		identity := exampleFieldIdentity(endpoint.MethodExpr.Payload, name, owner)
		params = append(params, paramFor(att, elem, "cookie", required, rand, identity, values))
		return nil
	})

	return params
}

// exampleFieldIdentity returns the repeatable example key for a named field.
// Named user types use a key derived from their type; anonymous objects append
// the field name to the key supplied for their parent.
func exampleFieldIdentity(parent *expr.AttributeExpr, name string, owner expr.ExampleIdentity) expr.ExampleIdentity {
	if typ, ok := parent.Type.(expr.UserType); ok {
		owner = expr.UserTypeExampleIdentity(typ)
	}
	return owner.Member(name)
}

// paramFor converts the given attribute into a OpenAPI spec parameter.
func paramFor(att *expr.AttributeExpr, name, in string, required bool, rand *expr.ExampleGenerator, identity expr.ExampleIdentity, values openapi.Values) *Parameter {
	param := &Parameter{
		Name:            name,
		In:              in,
		Description:     values.Description(att.AuthoredAttribute(), att.Description),
		AllowEmptyValue: in == "query",
		Required:        required,
		Schema:          newSchemafier(rand.At(identity), values).schemafy(att),
		Extensions:      openapi.ExtensionsFromExpr(att.Meta),
	}
	initExamples(param, att, rand.At(identity), values)
	return param
}
