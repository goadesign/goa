// This file binds reusable HTTP error-response policy to the concrete error
// returned by each endpoint method.
package expr

import (
	"goa.design/goa/v3/eval"
)

type (
	// HTTPErrorExpr defines a HTTP error response including its name,
	// status, headers and result type.
	HTTPErrorExpr struct {
		// ErrorExpr is the underlying goa design error expression.
		*ErrorExpr
		// Name of error, we need a separate copy of the name to match it
		// up with the appropriate ErrorExpr.
		Name string
		// Response is the corresponding HTTP response.
		Response *HTTPResponseExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (e *HTTPErrorExpr) EvalName() string {
	return "HTTP error " + e.Name
}

// IsJSONRPC reports whether the response maps an error for a JSON-RPC API,
// service, or method.
func (e *HTTPErrorExpr) IsJSONRPC() bool {
	switch parent := e.Response.Parent.(type) {
	case *JSONRPCExpr:
		return true
	case *HTTPServiceExpr:
		return parent.IsJSONRPC()
	case *HTTPEndpointExpr:
		return parent.IsJSONRPC()
	default:
		return false
	}
}

// Validate makes sure there is a error expression that matches the HTTP error
// expression.
func (e *HTTPErrorExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if e.IsJSONRPC() {
		if jsonRPCErrorCodeReserved(e.Response.StatusCode) {
			verr.Add(
				e.Response,
				"JSON-RPC error code %d is reserved; use a standard protocol code, a code from -32099 through -32000, or an application code outside -32768 through -32000",
				e.Response.StatusCode,
			)
		}
		// One JSON-RPC error message cannot add HTTP headers or cookies to the
		// shared response used for a batch or server-sent-event connection.
		if e.Response.Headers != nil && !e.Response.Headers.IsEmpty() {
			verr.Add(e.Response, "JSON-RPC error response cannot map error attributes to HTTP headers")
		}
		if e.Response.Cookies != nil && !e.Response.Cookies.IsEmpty() {
			verr.Add(e.Response, "JSON-RPC error response cannot map error attributes to HTTP cookies")
		}
	}
	switch p := e.Response.Parent.(type) {
	case *HTTPEndpointExpr:
		if p.MethodExpr.Error(e.Name) == nil {
			verr.Add(e.Response, "Error %#v does not match an error defined in the method", e.Name)
		}
	case *HTTPServiceExpr:
		if p.Error(e.Name) == nil {
			verr.Add(e.Response, "Error %#v does not match an error defined in the service", e.Name)
		}
	case *RootExpr:
		if p.Error(e.Name) == nil {
			verr.Add(e.Response, "Error %#v does not match an error defined in the API", e.Name)
		}
	case *JSONRPCExpr:
		if Root.Error(e.Name) == nil {
			verr.Add(e.Response, "Error %#v does not match an error defined in the API", e.Name)
		}
	}

	var ee *ErrorExpr
	switch p := e.Response.Parent.(type) {
	case *HTTPEndpointExpr:
		ee = p.MethodExpr.Error(e.Name)
	case *HTTPServiceExpr:
		ee = p.Error(e.Name)
	case *RootExpr:
		ee = p.Error(e.Name)
	case *JSONRPCExpr:
		ee = Root.Error(e.Name)
	}
	if ee == nil {
		return verr
	}

	// validate headers
	if e.Response.Headers != nil && !e.Response.Headers.IsEmpty() {
		verr.Merge(e.Response.Headers.Validate("HTTP error response headers", e.Response))
		switch {
		case ee.Type == Empty:
			verr.Add(e.Response, "response defines headers but error type is empty")
		case IsObject(ee.Type):
			for _, h := range *AsObject(e.Response.Headers.Type) {
				att := ee.Find(h.Name)
				switch {
				case att == nil:
					verr.Add(e.Response, "header %q has no equivalent attribute in error type, use notation 'attribute_name:header_name' to identify corresponding error type attribute.", h.Name)
				case IsArray(att.Type):
					if !IsPrimitive(AsArray(att.Type).ElemType.Type) {
						verr.Add(e.Response, "attribute %q used in HTTP headers must be a primitive type or an array of primitive types.", h.Name)
					}
				case !IsPrimitive(att.Type):
					verr.Add(e.Response, "attribute %q used in HTTP headers must be a primitive type or an array of primitive types.", h.Name)
				}
			}
		case len(*AsObject(e.Response.Headers.Type)) > 1:
			verr.Add(e.Response, "response defines more than one headers but error type is not an object")
		case IsArray(ee.Type):
			if !IsPrimitive(AsArray(ee.Type).ElemType.Type) {
				verr.Add(e.Response, "Array error type is mapped to an HTTP header but is not an array of primitive types.")
			}
		case IsMap(ee.Type):
			verr.Add(e.Response, "error type must be a primitive type or an array of primitive types.")
		}
	}
	return verr
}

// jsonRPCErrorCodeReserved reports whether code belongs to the part of the
// JSON-RPC reserved range that an application cannot use.
func jsonRPCErrorCodeReserved(code int) bool {
	if code < -32768 || code > -32000 {
		return false
	}
	if code >= -32099 {
		return false
	}
	switch code {
	case RPCParseError, RPCInvalidRequest, RPCMethodNotFound, RPCInvalidParams, RPCInternalError:
		return false
	default:
		return true
	}
}

// Finalize looks up the corresponding method error expression.
func (e *HTTPErrorExpr) Finalize(a *HTTPEndpointExpr) {
	e.ErrorExpr = a.MethodExpr.Error(e.Name)
	e.Response.Finalize(a, e.AttributeExpr)
	if e.Response.Body == nil {
		e.Response.Body = httpErrorResponseBody(a, e)
		e.Response.Body.Finalize()
	}
	// map any unmapped attributes in ErrorResult type to response headers
	e.Response.mapUnmappedAttrs(e.AttributeExpr)

	// Initialize response content type if result is media type.
	if e.Response.Body.Type == Empty {
		return
	}
	if e.Response.ContentType != "" {
		return
	}
	mt, ok := e.Response.Body.Type.(*ResultTypeExpr)
	if !ok {
		return
	}
	e.Response.ContentType = mt.Identifier
}

// mappedError returns the error declaration that owns this reusable HTTP
// response policy before the policy is applied to an endpoint method.
func (e *HTTPErrorExpr) mappedError() (*ErrorExpr, string) {
	switch parent := e.Response.Parent.(type) {
	case *HTTPEndpointExpr:
		return parent.MethodExpr.Error(e.Name), "method"
	case *HTTPServiceExpr:
		return parent.Error(e.Name), "service"
	case *RootExpr:
		return parent.Error(e.Name), "API"
	case *JSONRPCExpr:
		return Root.Error(e.Name), "API"
	}
	return nil, ""
}

// Dup creates a copy of the error expression.
func (e *HTTPErrorExpr) Dup() *HTTPErrorExpr {
	return &HTTPErrorExpr{
		ErrorExpr: e.ErrorExpr,
		Name:      e.Name,
		Response:  e.Response.Dup(),
	}
}
