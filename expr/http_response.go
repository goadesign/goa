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
	// headers and result type.
	HTTPResponseExpr struct {
		// HTTP status
		StatusCode int
		// Response description
		Description string
		// Headers describe the HTTP response headers.
		Headers *MappedAttributeExpr
		// Cookies describe the HTTP response cookies.
		Cookies *MappedAttributeExpr
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
	if r.Cookies == nil {
		r.Cookies = NewEmptyMappedAttributeExpr()
	}
}

// Validate checks that the response definition is consistent: its status is set
// and the result type definition if any is valid.
func (r *HTTPResponseExpr) Validate(e *HTTPEndpointExpr) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	if r.StatusCode == 0 {
		verr.Add(r, "HTTP response status not defined")
	} else if !bodyAllowedForStatus(r.StatusCode) && !e.MethodExpr.IsStreaming() {
		ep, ok := r.Parent.(*HTTPEndpointExpr)
		if ok && httpResponseBody(ep, r).Type != Empty {
			verr.Add(r, "Response body defined for status code %d which does not allow response body.", r.StatusCode)
		}
	}

	// text/html and text/plain can only encode strings so make sure there isn't
	// an explicit conflict with the content-type and response.
	if (r.ContentType == "text/html" || r.ContentType == "text/plain") && !e.SkipRequestBodyEncodeDecode {
		if e.MethodExpr.Result.Type != nil && e.MethodExpr.Result.Type != String && e.MethodExpr.Result.Type != Bytes && r.Body == nil {
			verr.Add(r, fmt.Sprintf("Result type must be String or Bytes when ContentType is '%s'", r.ContentType))
		}
		if r.Body != nil && r.Body.Type != String && r.Body.Type != Bytes {
			verr.Add(r, fmt.Sprintf("Result type must be String or Bytes when ContentType is '%s'", r.ContentType))
		}
	}

	rt, isrt := e.MethodExpr.Result.Type.(*ResultTypeExpr)
	resultAttributeType := func(name string) DataType {
		if !IsObject(e.MethodExpr.Result.Type) {
			return nil
		}
		if isrt {
			if v, ok := e.MethodExpr.Result.Meta["view"]; ok {
				v := rt.View(v[0])
				if v == nil {
					return nil
				}
				return v.AttributeExpr.Find(name).Type
			}
			for _, v := range rt.Views {
				if !rt.ViewHasAttribute(v.Name, name) {
					return nil
				}
			}
		}
		att := e.MethodExpr.Result.Find(name)
		if att == nil || att.Type == nil {
			// nil != nil
			return nil
		}
		return att.Type
	}

	var inview string
	if isrt {
		inview = " all views of"
	}

	if !r.Headers.IsEmpty() {
		verr.Merge(r.Headers.Validate("HTTP response headers", r))
		if isEmpty(e.MethodExpr.Result) {
			verr.Add(r, "response defines headers but result is empty")
		} else if IsObject(e.MethodExpr.Result.Type) {
			mobj := AsObject(r.Headers.Type)
			for _, h := range *mobj {
				t := resultAttributeType(h.Name)
				if t == nil {
					verr.Add(r, "header %q has no equivalent attribute in%s result type, use notation 'attribute_name:header_name' to identify corresponding result type attribute.", h.Name, inview)
				} else if IsArray(t) {
					if !IsPrimitive(AsArray(t).ElemType.Type) {
						verr.Add(e, "attribute %q used in HTTP headers must be a primitive type or an array of primitive types.", h.Name)
					}
				} else if !IsPrimitive(t) {
					verr.Add(e, "attribute %q used in HTTP headers must be a primitive type or an array of primitive types.", h.Name)
				}
			}
		} else if len(*AsObject(r.Headers.Type)) > 1 {
			verr.Add(r, "response defines more than one headers but result type is not an object")
		} else if IsArray(e.MethodExpr.Result.Type) {
			if !IsPrimitive(AsArray(e.MethodExpr.Result.Type).ElemType.Type) {
				verr.Add(e, "Array result is mapped to an HTTP header but is not an array of primitive types.")
			}
		}
	}
	if !r.Cookies.IsEmpty() {
		verr.Merge(r.Cookies.Validate("HTTP response cookies", r))
		if isEmpty(e.MethodExpr.Result) {
			verr.Add(r, "response defines cookies but result is empty")
		} else if IsObject(e.MethodExpr.Result.Type) {
			mobj := AsObject(r.Cookies.Type)
			for _, c := range *mobj {
				t := resultAttributeType(c.Name)
				if t == nil {
					verr.Add(r, "cookie %q has no equivalent attribute in%s result type, use notation 'attribute_name:cookie_name' to identify corresponding result type attribute.", c.Name, inview)
				}
				if !IsPrimitive(t) {
					verr.Add(e, "attribute %q used in HTTP cookies must be a primitive type.", c.Name)
				}
			}
		} else if len(*AsObject(r.Cookies.Type)) > 1 {
			verr.Add(r, "response defines more than one cookies but result type is not an object")
		} else if IsArray(e.MethodExpr.Result.Type) {
			verr.Add(e, "Array result is mapped to an HTTP cookie.")
		}
	}
	if r.Body != nil {
		verr.Merge(r.Body.Validate("HTTP response body", r))
		if e.SkipResponseBodyEncodeDecode {
			verr.Add(r, "Cannot define a response body when endpoint uses SkipResponseBodyEncodeDecode.")
		}
		if att, ok := r.Body.Meta["origin:attribute"]; ok {
			if resultAttributeType(att[0]) == nil {
				verr.Add(r, "body %q has no equivalent attribute in%s result type", att[0], inview)
			}
		} else if bobj := AsObject(r.Body.Type); bobj != nil {
			for _, n := range *bobj {
				if resultAttributeType(n.Name) == nil {
					verr.Add(r, "body %q has no equivalent attribute in%s result type", n.Name, inview)
				}
			}
		}
	} else if e.SkipResponseBodyEncodeDecode {
		body := httpResponseBody(e, r)
		if body.Type != Empty {
			verr.Add(e, "HTTP endpoint response body must be empty when using SkipResponseBodyEncodeDecode. Make sure to define headers and cookies as needed.")
		}
	}
	return verr
}

// Finalize sets the response result type from its type if the type is a result
// type and no result type is already specified.
func (r *HTTPResponseExpr) Finalize(a *HTTPEndpointExpr, svcAtt *AttributeExpr) {
	r.Parent = a

	if r.Body != nil && r.Body.Type != Empty {
		bodyAtt := svcAtt
		if o, ok := r.Body.Meta["origin:attribute"]; ok {
			bodyAtt = svcAtt.Find(o[0])
		}
		bodyObj := AsObject(bodyAtt.Type)
		if body := AsObject(r.Body.Type); body != nil {
			for _, nat := range *body {
				n := nat.Name
				n = strings.Split(n, ":")[0]
				var att, patt *AttributeExpr
				var required bool
				if bodyObj != nil {
					att = bodyObj.Attribute(n)
					required = bodyAtt.IsRequired(n)
				} else {
					att = bodyAtt
					required = bodyAtt.Type != Empty
				}
				initAttrFromDesign(att, patt)
				if required {
					if r.Body.Validation == nil {
						r.Body.Validation = &ValidationExpr{}
					}
					r.Body.Validation.AddRequired(n)
				}
			}
			// Remember original name for example to generate friendlier OpenAPI specs.
			if t, ok := r.Body.Type.(UserType); ok {
				t.Attribute().AddMeta("name:original", t.Name())
			}
			// Wrap object with user type to simplify response rendering code.
			r.Body.Type = &UserTypeExpr{
				AttributeExpr: DupAtt(r.Body),
				TypeName:      fmt.Sprintf("%s%sResponseBody", a.Service.Name(), a.Name()),
			}
		}
		if r.Body.Meta == nil {
			r.Body.Meta = bodyAtt.Meta
		}
	}

	// Set response content type if empty and if set in the result type
	if r.ContentType == "" {
		if rt, ok := svcAtt.Type.(*ResultTypeExpr); ok && rt.ContentType != "" {
			r.ContentType = rt.ContentType
		}
	}

	initAttr(r.Headers, svcAtt)
	initAttr(r.Cookies, svcAtt)
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
	if r.Headers != nil {
		res.Headers = DupMappedAtt(r.Headers)
	}
	if r.Cookies != nil {
		res.Cookies = DupMappedAtt(r.Cookies)
	}
	return &res
}

// mapUnmappedAttrs maps any unmapped attributes in ErrorResult type to the
// response headers. Unmapped attributes refer to the attributes in ErrorResult
// type that are not mapped to response body or headers. Such unmapped
// attributes are mapped to special goa headers in the form of
// "Goa-Attribute(-<Attribute Name>)".
func (r *HTTPResponseExpr) mapUnmappedAttrs(svcAtt *AttributeExpr) {
	if svcAtt.Type != ErrorResult {
		return
	}

	// map attributes to headers that are not explicitly mapped
	switch {
	case IsObject(svcAtt.Type):
		// map the attribute names in the service type to response headers if
		// not mapped explicitly.

		var originAttr string
		{
			if r.Body != nil {
				if o, ok := r.Body.Meta["origin:attribute"]; ok {
					originAttr = o[0]
				}
			}
		}
		// if response body was mapped explicitly using Body(<attribute name>) then
		// we must make sure we map all the other unmapped attributes to headers.
		if r.Body == nil || r.Body.Type == Empty || originAttr != "" {
			for _, nat := range *(AsObject(svcAtt.Type)) {
				if originAttr == nat.Name {
					continue
				}
				if _, ok := r.Headers.FindKey(nat.Name); ok {
					continue
				}
				r.Headers.Type.(*Object).Set(nat.Name, nat.Attribute)
				r.Headers.Map("goa-attribute-"+nat.Name, nat.Name)
				if svcAtt.IsRequired(nat.Name) {
					if r.Headers.Validation == nil {
						r.Headers.Validation = &ValidationExpr{}
					}
					r.Headers.Validation.AddRequired(nat.Name)
				}
			}
		}
	default:
		if r.Headers.IsEmpty() && (r.Body == nil || r.Body.Type == Empty) {
			r.Headers.Type.(*Object).Set("goa-attribute", svcAtt)
		}
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
