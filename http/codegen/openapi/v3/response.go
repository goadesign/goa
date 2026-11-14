// This file converts HTTP response headers and cookies into OpenAPI v3 values
// without sharing consumed example streams between schema and display fields.
// It also merges the responses of errors that share one status code.
package openapiv3

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

type (
	// errorResponse pairs the OpenAPI response built for one design error with
	// the error name that labels its example once merged.
	errorResponse struct {
		name     string
		response *Response
	}

	// errorMediaType pairs the media type of one error response body with the
	// error name that labels its example once merged.
	errorMediaType struct {
		name      string
		mediaType *MediaType
	}
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

// mergeErrorResponses returns the response documenting every error an endpoint
// maps to one HTTP status code, given the errors' responses in declaration
// order. OpenAPI allows a single response per status so the errors share it:
// its description lists each error description, each content type lists the
// distinct body schemas of the errors using it, and a header is required only
// when every error requires it. The content comes from the errors that define
// a body, so it is documented even when another error of the status has none.
// A status used by a single error keeps that error's response unchanged.
func mergeErrorResponses(ers []*errorResponse) *Response {
	if len(ers) == 1 {
		return ers[0].response
	}
	descriptions := make([]string, len(ers))
	headers := make(map[string]*HeaderRef)
	contents := make(map[string][]*errorMediaType)
	var extensions map[string]any
	for i, er := range ers {
		descriptions[i] = *er.response.Description
		for name, ref := range er.response.Headers {
			if _, ok := headers[name]; !ok {
				header := *ref.Value
				// Errors declared earlier do not set the header.
				header.Required = header.Required && i == 0
				headers[name] = &HeaderRef{Value: &header}
			}
		}
		for name, ref := range headers {
			if header, ok := er.response.Headers[name]; !ok || !header.Value.Required {
				ref.Value.Required = false
			}
		}
		for ct, mt := range er.response.Content {
			contents[ct] = append(contents[ct], &errorMediaType{name: er.name, mediaType: mt})
		}
		extensions = mergeExtensions(extensions, er.response.Extensions)
	}
	description := strings.Join(descriptions, "\n\n")
	merged := &Response{Description: &description, Extensions: extensions}
	if len(headers) > 0 {
		merged.Headers = headers
	}
	if len(contents) > 0 {
		merged.Content = make(map[string]*MediaType, len(contents))
		for ct, mts := range contents {
			merged.Content[ct] = mergeErrorMediaTypes(mts)
		}
	}
	return merged
}

// mergeErrorMediaTypes returns the media type describing the bodies of the
// errors that share one status code and content type, given in declaration
// order. Identical body schemas are listed once and different ones become
// anyOf alternatives. Each error example is keyed by the error name, prefixed
// to the example summary when the error defines several examples.
func mergeErrorMediaTypes(mts []*errorMediaType) *MediaType {
	if len(mts) == 1 {
		return mts[0].mediaType
	}
	var (
		schemas    []*openapi.Schema
		examples   = make(map[string]*ExampleRef)
		extensions map[string]any
	)
	for _, mt := range mts {
		schema := mt.mediaType.Schema
		if !slices.ContainsFunc(schemas, func(s *openapi.Schema) bool { return reflect.DeepEqual(s, schema) }) {
			schemas = append(schemas, schema)
		}
		if mt.mediaType.Example != nil {
			examples[mt.name] = &ExampleRef{Value: &Example{Value: mt.mediaType.Example}}
		}
		for summary, example := range mt.mediaType.Examples {
			examples[mt.name+": "+summary] = example
		}
		extensions = mergeExtensions(extensions, mt.mediaType.Extensions)
	}
	merged := &MediaType{Schema: schemas[0], Extensions: extensions}
	if len(schemas) > 1 {
		merged.Schema = &openapi.Schema{AnyOf: schemas}
	}
	if len(examples) > 0 {
		merged.Examples = examples
	}
	return merged
}

// mergeExtensions returns dst with the extensions of src it does not define
// yet so the error declared first wins when two errors set one extension.
func mergeExtensions(dst, src map[string]any) map[string]any {
	for key, value := range src {
		if dst == nil {
			dst = make(map[string]any, len(src))
		}
		if _, ok := dst[key]; !ok {
			dst[key] = value
		}
	}
	return dst
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
