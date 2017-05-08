package rest

import (
	"fmt"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
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
	StatusUnsupportedMediaType         = 415 // RFC 7231, 6.5.13
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
	// headers and media type.
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
		// Parent expression, one of ActionExpr, ResourceExpr or
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
		r.headers = &design.AttributeExpr{Type: make(design.Object)}
	}
	return r.headers
}

// MappedHeaders returns the computed response headers attribute map.
func (r *HTTPResponseExpr) MappedHeaders() *MappedAttributeExpr {
	return NewMappedAttributeExpr(r.headers)
}

// MediaType returns the media type describing the response body if any, nil
// otherwise.
func (r *HTTPResponseExpr) MediaType() *design.MediaTypeExpr {
	if r.Body == nil {
		return nil
	}
	if mt, ok := r.Body.Type.(*design.MediaTypeExpr); ok {
		return mt
	}
	return nil
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
// and the media type definition if any is valid.
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
	}
	return verr
}

// Finalize sets the response media type from its type if the type is a media
// type and no media type is already specified.
func (r *HTTPResponseExpr) Finalize() {
	if r.Body == nil {
		return
	}
	if r.ContentType != "" {
		return
	}
	mt, ok := r.Body.Type.(*design.MediaTypeExpr)
	if !ok {
		return
	}
	r.ContentType = mt.Identifier
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
