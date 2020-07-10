package openapiv3

import (
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func responseFromExpr(r *expr.HTTPResponseExpr, bodies map[int][]*openapi.Schema, rand *expr.Random) *Response {
	ct := r.ContentType
	rt, ok := r.Body.Type.(*expr.ResultTypeExpr)
	if ok && ct == "" {
		ct = rt.ContentType
	}
	var headers map[string]*HeaderRef
	o := expr.AsObject(r.Headers.Type)
	if len(*o) > 0 {
		headers = make(map[string]*HeaderRef, len(*o))
		expr.WalkMappedAttr(r.Headers, func(name, elem string, attr *expr.AttributeExpr) error {
			headers[elem] = &HeaderRef{Value: &Header{
				Description: attr.Description,
				Required:    r.Headers.IsRequiredNoDefault(name),
				Schema:      newSchemafier(rand).schemafy(attr),
				Example:     attr.Example(rand),
				Extensions:  openapi.ExtensionsFromExpr(attr.Meta),
			}}
			return nil
		})
	}
	mt := &MediaType{
		Schema:     bodies[r.StatusCode][0],
		Example:    r.Body.Example(rand),
		Extensions: openapi.ExtensionsFromExpr(r.Body.Meta),
	}
	return &Response{
		Description: &r.Description,
		Headers:     headers,
		Content:     map[string]*MediaType{ct: mt},
		Extensions:  openapi.ExtensionsFromExpr(r.Meta),
	}
}
