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

func headersFromAttr(attr *expr.MappedAttributeExpr, parent *expr.AttributeExpr, owner expr.ExampleIdentity, rand *expr.ExampleGenerator, values openapi.Values) map[string]*HeaderRef {
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
			Description: values.Description(hattr.AuthoredAttribute(), hattr.Description),
			Required:    hattr.IsRequiredNoDefault(name),
			Schema:      newSchemafier(rand.At(identity), values).schemafy(hattr),
			Extensions:  openapi.ExtensionsFromExpr(hattr.Meta),
		}
		initExamples(header, hattr, rand.At(identity), values)
		headers[elem] = &HeaderRef{Value: header}
		return nil
	})
	return headers
}

func responseFromExpr(r *expr.HTTPResponseExpr, body *openapi.Schema, rand *expr.ExampleGenerator, parent *expr.AttributeExpr, fieldOwner, bodyOwner expr.ExampleIdentity, fallbackDescription string, values openapi.Values) *Response {
	ct := openapi.ResponseContentType(r)
	headers := headersFromAttr(r.Headers, parent, fieldOwner, rand, values)
	cookies := headersFromAttr(r.Cookies, parent, fieldOwner, rand, values)
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
			initExamples(content[ct], staticViewBody(r), rand.At(bodyOwner), values)
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
	desc := values.Description(r, r.Description)
	if desc == "" {
		desc = fallbackDescription
	}
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

// setSSEContent rewrites the content of a successful server-sent events
// response for OpenAPI 3.2 documents. The text/event-stream media type
// describes each streamed event with an itemSchema instead of a whole-stream
// schema. When the method defines separate normal and streaming results, the
// normal result is documented under its own content type next to the event
// stream to match the response selected by the generated handler.
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
