// This file converts HTTP parameters, headers, and cookies into OpenAPI v3
// values whose schema and displayed examples use the same fresh identity.
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
func paramsFromPath(endpoint *expr.HTTPEndpointExpr, path string, rand *expr.ExampleGenerator) []*Parameter {
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
		res = append(res, paramFor(at, pn, in, required, rand, identity))
		return nil
	})
	return res
}

// paramsFromHeadersAndCookies computes the OpenAPI spec parameters for the
// given endpoint HTTP headers and cookies.
func paramsFromHeadersAndCookies(endpoint *expr.HTTPEndpointExpr, rand *expr.ExampleGenerator) []*Parameter {
	var params []*Parameter
	owner := expr.MethodPayloadExampleIdentity(endpoint.MethodExpr)

	expr.WalkMappedAttr(endpoint.Headers, func(name, elem string, att *expr.AttributeExpr) error { // nolint: errcheck
		if openapiinternal.IsSecurityParameter(endpoint, "header", elem) {
			return nil
		}
		required := endpoint.Headers.IsRequiredNoDefault(name)
		identity := exampleFieldIdentity(endpoint.MethodExpr.Payload, name, owner)
		params = append(params, paramFor(att, elem, "header", required, rand, identity))
		return nil
	})
	expr.WalkMappedAttr(endpoint.Cookies, func(name, elem string, att *expr.AttributeExpr) error { // nolint: errcheck
		if openapiinternal.IsSecurityParameter(endpoint, "cookie", elem) {
			return nil
		}
		required := endpoint.Cookies.IsRequiredNoDefault(name)
		identity := exampleFieldIdentity(endpoint.MethodExpr.Payload, name, owner)
		params = append(params, paramFor(att, elem, "cookie", required, rand, identity))
		return nil
	})

	return params
}

// exampleFieldGenerator anchors a detached transport field to its named user
// type or to the explicit semantic owner of an anonymous parent.
func exampleFieldIdentity(parent *expr.AttributeExpr, name string, owner expr.ExampleIdentity) expr.ExampleIdentity {
	if typ, ok := parent.Type.(expr.UserType); ok {
		owner = expr.UserTypeExampleIdentity(typ)
	}
	return owner.Member(name)
}

// paramFor converts the given attribute into a OpenAPI spec parameter.
func paramFor(att *expr.AttributeExpr, name, in string, required bool, rand *expr.ExampleGenerator, identity expr.ExampleIdentity) *Parameter {
	param := &Parameter{
		Name:            name,
		In:              in,
		Description:     att.Description,
		AllowEmptyValue: in == "query",
		Required:        required,
		Schema:          newSchemafier(rand.At(identity)).schemafy(att),
		Extensions:      openapi.ExtensionsFromExpr(att.Meta),
	}
	initExamples(param, att, rand.At(identity))
	return param
}
