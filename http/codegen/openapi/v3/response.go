package openapiv3

import (
	"fmt"
	"net/http"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func responseFromExpr(r *expr.HTTPResponseExpr, bodies map[int][]*openapi.Schema, rand *expr.Random) *Response {
	ct := r.ContentType
	rt, ok := r.Body.Type.(*expr.ResultTypeExpr)
	if ok && ct == "" {
		ct = rt.ContentType
	}
	if ct == "" {
		// Default to application/json
		ct = "application/json"
	}
	var headers map[string]*HeaderRef
	o := expr.AsObject(r.Headers.Type)
	if len(*o) > 0 {
		headers = make(map[string]*HeaderRef, len(*o))
		expr.WalkMappedAttr(r.Headers, func(name, elem string, attr *expr.AttributeExpr) error {
			header := &Header{
				Description: attr.Description,
				Required:    r.Headers.IsRequiredNoDefault(name),
				Schema:      newSchemafier(rand).schemafy(attr),
				Example:     attr.Example(rand),
				Extensions:  openapi.ExtensionsFromExpr(attr.Meta),
			}
			initExamples(header, attr, rand)
			headers[elem] = &HeaderRef{Value: header}
			return nil
		})
	}

	var content map[string]*MediaType
	{
		if r.Body.Type != expr.Empty {
			content = make(map[string]*MediaType)
			content[ct] = &MediaType{
				Schema:     bodies[r.StatusCode][0],
				Extensions: openapi.ExtensionsFromExpr(r.Body.Meta),
			}
			initExamples(content[ct], r.Body, rand)
		}
	}
	desc := r.Description
	if desc == "" {
		desc = fmt.Sprintf("%s response.", http.StatusText(r.StatusCode))
	}
	return &Response{
		Description: &desc,
		Headers:     headers,
		Content:     content,
		Extensions:  openapi.ExtensionsFromExpr(r.Meta),
	}
}
