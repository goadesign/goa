// This file converts HTTP response headers and cookies into OpenAPI v3 values
// without sharing consumed example streams between schema and display fields.
package openapiv3

import (
	"fmt"
	"net/http"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func headersFromAttr(attr *expr.MappedAttributeExpr, parent *expr.AttributeExpr, owner expr.ExampleIdentity, rand *expr.ExampleGenerator) map[string]*HeaderRef {
	o := expr.AsObject(attr.Type)
	if len(*o) == 0 {
		return nil
	}
	headers := make(map[string]*HeaderRef, len(*o))
	expr.WalkMappedAttr(attr, func(name, elem string, hattr *expr.AttributeExpr) error { // nolint: errcheck
		// Anchor the header example stream to the header identity so the
		// example survives generator reorderings.
		identity := exampleFieldIdentity(parent, name, owner)
		header := &Header{
			Description: hattr.Description,
			Required:    hattr.IsRequiredNoDefault(name),
			Schema:      newSchemafier(rand.At(identity)).schemafy(hattr),
			Extensions:  openapi.ExtensionsFromExpr(hattr.Meta),
		}
		initExamples(header, hattr, rand.At(identity))
		headers[elem] = &HeaderRef{Value: header}
		return nil
	})
	return headers
}

func responseFromExpr(r *expr.HTTPResponseExpr, body *openapi.Schema, rand *expr.ExampleGenerator, parent *expr.AttributeExpr, fieldOwner, bodyOwner expr.ExampleIdentity) *Response {
	ct := responseContentType(r)
	headers := headersFromAttr(r.Headers, parent, fieldOwner, rand)
	cookies := headersFromAttr(r.Cookies, parent, fieldOwner, rand)
	if len(cookies) > 0 {
		if headers == nil {
			headers = make(map[string]*HeaderRef)
		}
		if len(cookies) == 1 {
			for _, v := range cookies {
				headers["Set-Cookie"] = v
			}
		} else {
			// Generic cookies header
			headers["Set-Cookie"] = &HeaderRef{
				Value: &Header{
					Description: "Cookies set by the server",
					Required:    true,
					Schema: &openapi.Schema{
						Type: "string",
					},
				},
			}
		}
	}

	var content map[string]*MediaType
	{
		if r.Body.Type != expr.Empty {
			content = make(map[string]*MediaType)
			content[ct] = &MediaType{
				Schema:     body,
				Extensions: openapi.ExtensionsFromExpr(r.Body.Meta),
			}
			initExamples(content[ct], staticViewBody(r), rand.At(bodyOwner))
		} else if r.StatusCode != expr.StatusNoContent &&
			isSkipResponseBodyEncodeDecode(r.Parent) {
			// When SkipResponseBodyEncodeDecode is declared, the response type
			// is Empty, but the response code is not 204 and has content.
			content = make(map[string]*MediaType)
			content[ct] = &MediaType{
				Schema: &openapi.Schema{
					Type:   "string",
					Format: "binary",
				},
				Extensions: openapi.ExtensionsFromExpr(r.Body.Meta),
			}
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

// responseContentType computes the content type of the given response: the
// explicitly defined content type if any, the content type of the response
// result type otherwise, defaulting to application/json. The result type is
// the view-projected one when the design pins the response to a single view;
// projected result types carry no content type.
func responseContentType(r *expr.HTTPResponseExpr) string {
	if r.ContentType != "" {
		return r.ContentType
	}
	if rt, ok := staticViewBody(r).Type.(*expr.ResultTypeExpr); ok && rt.ContentType != "" {
		return rt.ContentType
	}
	return "application/json"
}

// setSSEContent rewrites the content of a successful server-sent events
// response for OpenAPI 3.2 documents. The text/event-stream media type
// describes each streamed event with an itemSchema instead of a whole-stream
// schema. When the method defines mixed results (distinct unary and streaming
// result types) the unary result is documented under its own content type
// (ct) next to the event stream to reflect the content negotiation performed
// by the generated handler.
func setSSEContent(resp *Response, bodies *EndpointBodies, ct string, mixed bool) {
	sse := &MediaType{ItemSchema: bodies.SSEItemSchema}
	if !mixed {
		resp.Content = map[string]*MediaType{"text/event-stream": sse}
		return
	}
	if mt, ok := resp.Content["text/event-stream"]; ok {
		delete(resp.Content, "text/event-stream")
		resp.Content[ct] = mt
	}
	resp.Content["text/event-stream"] = sse
}

func isSkipResponseBodyEncodeDecode(parent eval.Expression) bool {
	ee, ok := parent.(*expr.HTTPEndpointExpr)
	return ok && ee.SkipResponseBodyEncodeDecode
}
