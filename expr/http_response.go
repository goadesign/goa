package expr

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/eval"
)

const (
	StatusContinue           = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols = 101 // RFC 7231, 6.2.2
	StatusProcessing         = 102 // RFC 2518, 10.1

	StatusOK                   = 200 // RFC 7231, 6.3.1
	StatusCreated              = 201 // RFC 7231, 6.3.2
	StatusAccepted             = 202 // RFC 7231, 6.3.3
	StatusNonAuthoritativeInfo = 203 // RFC 7231, 6.3.4
	StatusNoContent            = 204 // RFC 7231, 6.3.5
	StatusResetContent         = 205 // RFC 7231, 6.3.6
	StatusPartialContent       = 206 // RFC 7233, 4.1
	StatusMultiStatus          = 207 // RFC 4918, 11.1
	StatusAlreadyReported      = 208 // RFC 5842, 7.1
	StatusIMUsed               = 226 // RFC 3229, 10.4.1

	StatusMultipleChoices  = 300 // RFC 7231, 6.4.1
	StatusMovedPermanently = 301 // RFC 7231, 6.4.2
	StatusFound            = 302 // RFC 7231, 6.4.3
	StatusSeeOther         = 303 // RFC 7231, 6.4.4
	StatusNotModified      = 304 // RFC 7232, 4.1
	StatusUseProxy         = 305 // RFC 7231, 6.4.5

	StatusTemporaryRedirect = 307 // RFC 7231, 6.4.7
	StatusPermanentRedirect = 308 // RFC 7538, 3

	StatusBadRequest                   = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                 = 401 // RFC 7235, 3.1
	StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
	StatusForbidden                    = 403 // RFC 7231, 6.5.3
	StatusNotFound                     = 404 // RFC 7231, 6.5.4
	StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
	StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
	StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
	StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
	StatusConflict                     = 409 // RFC 7231, 6.5.8
	StatusGone                         = 410 // RFC 7231, 6.5.9
	StatusLengthRequired               = 411 // RFC 7231, 6.5.10
	StatusPreconditionFailed           = 412 // RFC 7232, 4.2
	StatusRequestEntityTooLarge        = 413 // RFC 7231, 6.5.11
	StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
	StatusUnsupportedResultType        = 415 // RFC 7231, 6.5.13
	StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
	StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
	StatusTeapot                       = 418 // RFC 7168, 2.3.3
	StatusUnprocessableEntity          = 422 // RFC 4918, 11.2
	StatusLocked                       = 423 // RFC 4918, 11.3
	StatusFailedDependency             = 424 // RFC 4918, 11.4
	StatusUpgradeRequired              = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired         = 428 // RFC 6585, 3
	StatusTooManyRequests              = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   = 451 // RFC 7725, 3

	StatusInternalServerError           = 500 // RFC 7231, 6.6.1
	StatusNotImplemented                = 501 // RFC 7231, 6.6.2
	StatusBadGateway                    = 502 // RFC 7231, 6.6.3
	StatusServiceUnavailable            = 503 // RFC 7231, 6.6.4
	StatusGatewayTimeout                = 504 // RFC 7231, 6.6.5
	StatusHTTPVersionNotSupported       = 505 // RFC 7231, 6.6.6
	StatusVariantAlsoNegotiates         = 506 // RFC 2295, 8.1
	StatusInsufficientStorage           = 507 // RFC 4918, 11.5
	StatusLoopDetected                  = 508 // RFC 5842, 7.2
	StatusNotExtended                   = 510 // RFC 2774, 7
	StatusNetworkAuthenticationRequired = 511 // RFC 6585, 6
)

type (
	// HTTPResponseExpr defines a HTTP response including its status code,
	// headers and result type.
	HTTPResponseExpr struct {
		// HTTP status
		StatusCode int
		// Response description
		Description string
		// Headers describe the HTTP response headers.
		Headers *MappedAttributeExpr
		// Response body if any
		Body *AttributeExpr
		// Response Content-Type header value
		ContentType string
		// Tag the value a field of the result must have for this
		// response to be used.
		Tag [2]string
		// Parent expression, one of EndpointExpr, ServiceExpr or
		// RootExpr.
		Parent eval.Expression
		// Meta is a list of key/value pairs
		Meta MetaExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (r *HTTPResponseExpr) EvalName() string {
	var suffix string
	if r.Parent != nil {
		suffix = fmt.Sprintf(" of %s", r.Parent.EvalName())
	}
	return "HTTP response" + suffix
}

// Prepare makes sure the response is initialized even if not done explicitly
// by
func (r *HTTPResponseExpr) Prepare() {
	if r.Headers == nil {
		r.Headers = NewEmptyMappedAttributeExpr()
	}
}

// Validate checks that the response definition is consistent: its status is set
// and the result type definition if any is valid.
func (r *HTTPResponseExpr) Validate(e *HTTPEndpointExpr) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	if r.StatusCode == 0 {
		verr.Add(r, "HTTP response status not defined")
	} else if !bodyAllowedForStatus(r.StatusCode) && r.bodyExists() && !e.MethodExpr.IsStreaming() {
		verr.Add(r, "Response body defined for status code %d which does not allow response body.", r.StatusCode)
	}

	if e.MethodExpr.Result.Type == Empty {
		if !r.Headers.IsEmpty() {
			verr.Add(r, "response defines headers but result is empty")
		}
		return verr
	}

	rt, isrt := e.MethodExpr.Result.Type.(*ResultTypeExpr)
	var inview string
	if isrt {
		inview = " all views in"
	}

	// text/html can only encode strings so make sure there isn't an explicit conflict with the content-type and response.
	if r.ContentType == "text/html" || r.ContentType == "text/plain" {
		if e.MethodExpr.Result.Type != nil && e.MethodExpr.Result.Type != String && e.MethodExpr.Result.Type != Bytes && r.Body == nil {
			verr.Add(r, fmt.Sprintf("Result type must be String or Bytes when ContentType is '%s'", r.ContentType))
		}
		if r.Body != nil && r.Body.Type != String && r.Body.Type != Bytes {
			verr.Add(r, fmt.Sprintf("Result type must be String or Bytes when ContentType is '%s'", r.ContentType))
		}
	}

	hasAttribute := func(name string) bool {
		if !IsObject(e.MethodExpr.Result.Type) {
			return false
		}
		if !isrt {
			return e.MethodExpr.Result.Find(name) != nil
		}
		if v, ok := e.MethodExpr.Result.Meta["view"]; ok {
			return rt.ViewHasAttribute(v[0], name)
		}
		for _, v := range rt.Views {
			if !rt.ViewHasAttribute(v.Name, name) {
				return false
			}
		}
		return true
	}
	if !r.Headers.IsEmpty() {
		verr.Merge(r.Headers.Validate("HTTP response headers", r))
		if e.MethodExpr.Result.Type == Empty {
			verr.Add(r, "response defines headers but result is empty")
		} else if IsObject(e.MethodExpr.Result.Type) {
			mobj := AsObject(r.Headers.Type)
			for _, h := range *mobj {
				if !hasAttribute(h.Name) {
					verr.Add(r, "header %q has no equivalent attribute in%s result type, use notation 'attribute_name:header_name' to identify corresponding result type attribute.", h.Name, inview)
				}
			}
		} else if len(*AsObject(r.Headers.Type)) > 1 {
			verr.Add(r, "response defines more than one header but result type is not an object")
		}
	}
	if r.Body != nil {
		verr.Merge(r.Body.Validate("HTTP response body", r))
		if att, ok := r.Body.Meta["origin:attribute"]; ok {
			if !hasAttribute(att[0]) {
				verr.Add(r, "body %q has no equivalent attribute in%s result type", att[0], inview)
			}
		} else if bobj := AsObject(r.Body.Type); bobj != nil {
			for _, n := range *bobj {
				if !hasAttribute(n.Name) {
					verr.Add(r, "body %q has no equivalent attribute in%s result type", n.Name, inview)
				}
			}
		}
	}
	return verr
}

// Finalize sets the response result type from its type if the type is a result
// type and no result type is already specified.
func (r *HTTPResponseExpr) Finalize(a *HTTPEndpointExpr, svcAtt *AttributeExpr) {
	r.Parent = a

	// Initialize the body attributes (if an object) with the corresponding
	// result attributes.
	svcObj := AsObject(svcAtt.Type)
	if r.Body != nil {
		if body := AsObject(r.Body.Type); body != nil {
			for _, nat := range *body {
				n := nat.Name
				n = strings.Split(n, ":")[0]
				var att, patt *AttributeExpr
				var required bool
				if svcObj != nil {
					att = svcObj.Attribute(n)
					required = svcAtt.IsRequired(n)
				} else {
					att = svcAtt
					required = svcAtt.Type != Empty
				}
				initAttrFromDesign(att, patt)
				if required {
					if r.Body.Validation == nil {
						r.Body.Validation = &ValidationExpr{}
					}
					r.Body.Validation.Required = append(r.Body.Validation.Required, n)
				}
			}
		}
		if r.Body.Meta == nil {
			r.Body.Meta = svcAtt.Meta
		}
	}
	// Set response content type if empty and if set in the result type
	if r.ContentType == "" {
		if rt, ok := svcAtt.Type.(*ResultTypeExpr); ok && rt.ContentType != "" {
			r.ContentType = rt.ContentType
		}
	}
	initAttr(r.Headers, svcAtt)
}

// Dup creates a copy of the response expression.
func (r *HTTPResponseExpr) Dup() *HTTPResponseExpr {
	res := HTTPResponseExpr{
		StatusCode:  r.StatusCode,
		Description: r.Description,
		ContentType: r.ContentType,
		Parent:      r.Parent,
		Meta:        r.Meta,
	}
	if r.Body != nil {
		res.Body = DupAtt(r.Body)
	}
	res.Headers = DupMappedAtt(r.Headers)
	return &res
}

// bodyAllowedForStatus reports whether a given response status code
// permits a body. See RFC 2616, section 4.4.
// See https://golang.org/src/net/http/transfer.go
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}

// bodyExists returns true if a response body is defined in the
// response expression via Body() or Result() in the method expression.
func (r *HTTPResponseExpr) bodyExists() bool {
	ep, ok := r.Parent.(*HTTPEndpointExpr)
	return ok && httpResponseBody(ep, r).Type != Empty
}
