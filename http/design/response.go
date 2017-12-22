package design

import (
	"fmt"
	"strings"

	"goa.design/goa/design"
	"goa.design/goa/eval"
)

const (
	// StatusContinue refers to HTTP code 100 (RFC 7231, 6.2.1)
	StatusContinue = 100
	// StatusSwitchingProtocols refers to HTTP code 101 (RFC 7231, 6.2.2)
	StatusSwitchingProtocols = 101
	// StatusProcessing refers to HTTP code 102 (RFC 2518, 10.1)
	StatusProcessing = 102

	// StatusOK refers to HTTP code 200 (RFC 7231, 6.3.1)
	StatusOK = 200
	// StatusCreated refers to HTTP code 201 (RFC 7231, 6.3.2)
	StatusCreated = 201
	// StatusAccepted refers to HTTP code 202 (RFC 7231, 6.3.3)
	StatusAccepted = 202
	// StatusNonAuthoritativeInfo refers to HTTP code 203 (RFC 7231, 6.3.4)
	StatusNonAuthoritativeInfo = 203
	// StatusNoContent refers to HTTP code 204 (RFC 7231, 6.3.5)
	StatusNoContent = 204
	// StatusResetContent refers to HTTP code 205 (RFC 7231, 6.3.6)
	StatusResetContent = 205
	// StatusPartialContent refers to HTTP code 206 (RFC 7233, 4.1)
	StatusPartialContent = 206
	// StatusMultiStatus refers to HTTP code 207 (RFC 4918, 11.1)
	StatusMultiStatus = 207
	// StatusAlreadyReported refers to HTTP code 208 (RFC 5842, 7.1)
	StatusAlreadyReported = 208
	// StatusIMUsed refers to HTTP code 226 (RFC 3229, 10.4.1)
	StatusIMUsed = 226

	// StatusMultipleChoices refers to HTTP code 300 (RFC 7231, 6.4.1)
	StatusMultipleChoices = 300
	// StatusMovedPermanently refers to HTTP code 301 (RFC 7231, 6.4.2)
	StatusMovedPermanently = 301
	// StatusFound refers to HTTP code 302 (RFC 7231, 6.4.3)
	StatusFound = 302
	// StatusSeeOther refers to HTTP code 303 (RFC 7231, 6.4.4)
	StatusSeeOther = 303
	// StatusNotModified refers to HTTP code 304 (RFC 7232, 4.1)
	StatusNotModified = 304
	// StatusUseProxy refers to HTTP code 305 (RFC 7231, 6.4.5)
	StatusUseProxy = 305
	// StatusTemporaryRedirect refers to HTTP code 307 (RFC 7231, 6.4.7)
	StatusTemporaryRedirect = 307
	// StatusPermanentRedirect refers to HTTP code 308 (RFC 7538, 3)
	StatusPermanentRedirect = 308

	// StatusBadRequest refers to HTTP code 400 (RFC 7231, 6.5.1)
	StatusBadRequest = 400
	// StatusUnauthorized refers to HTTP code 401 (RFC 7235, 3.1)
	StatusUnauthorized = 401
	// StatusPaymentRequired refers to HTTP code 402 (RFC 7231, 6.5.2)
	StatusPaymentRequired = 402
	// StatusForbidden refers to HTTP code 403 (RFC 7231, 6.5.3)
	StatusForbidden = 403
	// StatusNotFound refers to HTTP code 404 (RFC 7231, 6.5.4)
	StatusNotFound = 404
	// StatusMethodNotAllowed refers to HTTP code 405 (RFC 7231, 6.5.5)
	StatusMethodNotAllowed = 405
	// StatusNotAcceptable refers to HTTP code 406 (RFC 7231, 6.5.6)
	StatusNotAcceptable = 406
	// StatusProxyAuthRequired refers to HTTP code 407 (RFC 7235, 3.2)
	StatusProxyAuthRequired = 407
	// StatusRequestTimeout refers to HTTP code 408 (RFC 7231, 6.5.7)
	StatusRequestTimeout = 408
	// StatusConflict refers to HTTP code 409 (RFC 7231, 6.5.8)
	StatusConflict = 409
	// StatusGone refers to HTTP code 410 (RFC 7231, 6.5.9)
	StatusGone = 410
	// StatusLengthRequired refers to HTTP code 411 (RFC 7231, 6.5.10)
	StatusLengthRequired = 411
	// StatusPreconditionFailed refers to HTTP code 412 (RFC 7232, 4.2)
	StatusPreconditionFailed = 412
	// StatusRequestEntityTooLarge refers to HTTP code 413 (RFC 7231, 6.5.11)
	StatusRequestEntityTooLarge = 413
	// StatusRequestURITooLong refers to HTTP code 414 (RFC 7231, 6.5.12)
	StatusRequestURITooLong = 414
	// StatusUnsupportedResultType refers to HTTP code 415 (RFC 7231, 6.5.13)
	StatusUnsupportedResultType = 415
	// StatusRequestedRangeNotSatisfiable refers to HTTP code 416 (RFC 7233, 4.4)
	StatusRequestedRangeNotSatisfiable = 416
	// StatusExpectationFailed refers to HTTP code 417 (RFC 7231, 6.5.14)
	StatusExpectationFailed = 417
	// StatusTeapot refers to HTTP code 418 (RFC 7168, 2.3.3)
	StatusTeapot = 418
	// StatusUnprocessableEntity refers to HTTP code 422 (RFC 4918, 11.2)
	StatusUnprocessableEntity = 422
	// StatusLocked refers to HTTP code 423 (RFC 4918, 11.3)
	StatusLocked = 423
	// StatusFailedDependency refers to HTTP code 424 (RFC 4918, 11.4)
	StatusFailedDependency = 424
	// StatusUpgradeRequired refers to HTTP code 426 (RFC 7231, 6.5.15)
	StatusUpgradeRequired = 426
	// StatusPreconditionRequired refers to HTTP code 428 (RFC 6585, 3)
	StatusPreconditionRequired = 428
	// StatusTooManyRequests refers to HTTP code 429 (RFC 6585, 4)
	StatusTooManyRequests = 429
	// StatusRequestHeaderFieldsTooLarge refers to HTTP code 431 (RFC 6585, 5)
	StatusRequestHeaderFieldsTooLarge = 431
	// StatusUnavailableForLegalReasons refers to HTTP code 451 (RFC 7725, 3)
	StatusUnavailableForLegalReasons = 451

	// StatusInternalServerError refers to HTTP code 500 (RFC 7231, 6.6.1)
	StatusInternalServerError = 500
	// StatusNotImplemented refers to HTTP code 501 (RFC 7231, 6.6.2)
	StatusNotImplemented = 501
	// StatusBadGateway refers to HTTP code 502 (RFC 7231, 6.6.3)
	StatusBadGateway = 502
	// StatusServiceUnavailable refers to HTTP code 503 (RFC 7231, 6.6.4)
	StatusServiceUnavailable = 503
	// StatusGatewayTimeout refers to HTTP code 504 (RFC 7231, 6.6.5)
	StatusGatewayTimeout = 504
	// StatusHTTPVersionNotSupported refers to HTTP code 505 (RFC 7231, 6.6.6)
	StatusHTTPVersionNotSupported = 505
	// StatusVariantAlsoNegotiates refers to HTTP code 506 (RFC 2295, 8.1)
	StatusVariantAlsoNegotiates = 506
	// StatusInsufficientStorage refers to HTTP code 507 (RFC 4918, 11.5)
	StatusInsufficientStorage = 507
	// StatusLoopDetected refers to HTTP code 508 (RFC 5842, 7.2)
	StatusLoopDetected = 508
	// StatusNotExtended refers to HTTP code 510 (RFC 2774, 7)
	StatusNotExtended = 510
	// StatusNetworkAuthenticationRequired refers to HTTP code 511 (RFC 6585, 6)
	StatusNetworkAuthenticationRequired = 511
)

type (
	// HTTPResponseExpr defines a HTTP response including its status code,
	// headers and result type.
	HTTPResponseExpr struct {
		// HTTP status
		StatusCode int
		// Response description
		Description string
		// Response body if any
		Body *design.AttributeExpr
		// Response Content-Type header value
		ContentType string
		// Tag the value a field of the result must have for this
		// response to be used.
		Tag [2]string
		// Parent expression, one of EndpointExpr, ServiceExpr or
		// RootExpr.
		Parent eval.Expression
		// Metadata is a list of key/value pairs
		Metadata design.MetadataExpr
		// Response header attribute, access with Headers method
		headers *design.AttributeExpr
	}
)

// Headers returns the raw response headers attribute.
func (r *HTTPResponseExpr) Headers() *design.AttributeExpr {
	if r.headers == nil {
		r.headers = &design.AttributeExpr{Type: &design.Object{}}
	}
	return r.headers
}

// MappedHeaders returns the computed response headers attribute map.
func (r *HTTPResponseExpr) MappedHeaders() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(r.headers)
}

// EvalName returns the generic definition name used in error messages.
func (r *HTTPResponseExpr) EvalName() string {
	var suffix string
	if r.Parent != nil {
		suffix = fmt.Sprintf(" of %s", r.Parent.EvalName())
	}
	return "HTTP response" + suffix
}

// Validate checks that the response definition is consistent: its status is set
// and the result type definition if any is valid.
func (r *HTTPResponseExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if r.headers != nil {
		verr.Merge(r.headers.Validate("HTTP response headers", r))
	}
	if r.Body != nil {
		verr.Merge(r.Body.Validate("HTTP response body", r))
	}
	if r.StatusCode == 0 {
		verr.Add(r, "HTTP response status not defined")
	} else if !bodyAllowedForStatus(r.StatusCode) && r.bodyExists() {
		verr.Add(r, "Response body defined for status code %d which does not allow response body.", r.StatusCode)
	}
	return verr
}

// Finalize sets the response result type from its type if the type is a result
// type and no result type is already specified.
func (r *HTTPResponseExpr) Finalize(a *EndpointExpr, svcAtt *design.AttributeExpr) {
	r.Parent = a

	// Initialize the headers with the corresponding result attributes.
	svcObj := design.AsObject(svcAtt.Type)
	if r.headers != nil {
		for _, nat := range *design.AsObject(r.headers.Type) {
			n := nat.Name
			att := nat.Attribute
			n = strings.Split(n, ":")[0]
			var patt *design.AttributeExpr
			var required bool
			if svcObj != nil {
				patt = svcObj.Attribute(n)
				required = svcAtt.IsRequired(n)
			} else {
				patt = svcAtt
				required = svcAtt.Type != design.Empty
			}
			initAttrFromDesign(att, patt)
			if required {
				if r.headers.Validation == nil {
					r.headers.Validation = &design.ValidationExpr{}
				}
				r.headers.Validation.Required = append(r.headers.Validation.Required, n)
			}
		}
	}

	// Initialize the body attributes (if an object) with the corresponding
	// payload attributes.
	if r.Body != nil {
		if body := design.AsObject(r.Body.Type); body != nil {
			for _, nat := range *body {
				n := nat.Name
				att := nat.Attribute
				n = strings.Split(n, ":")[0]
				var patt *design.AttributeExpr
				var required bool
				if svcObj != nil {
					att = svcObj.Attribute(n)
					required = svcAtt.IsRequired(n)
				} else {
					att = svcAtt
					required = svcAtt.Type != design.Empty
				}
				initAttrFromDesign(att, patt)
				if required {
					if r.Body.Validation == nil {
						r.Body.Validation = &design.ValidationExpr{}
					}
					r.Body.Validation.Required = append(r.Body.Validation.Required, n)
				}
			}
		}
	}
}

// Dup creates a copy of the response expression.
func (r *HTTPResponseExpr) Dup() *HTTPResponseExpr {
	return &HTTPResponseExpr{
		StatusCode:  r.StatusCode,
		Description: r.Description,
		Body:        design.DupAtt(r.Body),
		ContentType: r.ContentType,
		Parent:      r.Parent,
		Metadata:    r.Metadata,
		headers:     design.DupAtt(r.headers),
	}
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
	ep, ok := r.Parent.(*EndpointExpr)
	return ok && ResponseBody(ep, r).Type != design.Empty
}
