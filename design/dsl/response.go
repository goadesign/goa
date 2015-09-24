package dsl

import . "github.com/raphael/goa/design"

// List of all built-in response names.
const (
	Continue           = "Continue"
	SwitchingProtocols = "SwitchingProtocols"

	OK                   = "OK"
	Created              = "Created"
	Accepted             = "Accepted"
	NonAuthoritativeInfo = "NonAuthoritativeInfo"
	NoContent            = "NoContent"
	ResetContent         = "ResetContent"
	PartialContent       = "PartialContent"

	MultipleChoices   = "MultipleChoices"
	MovedPermanently  = "MovedPermanently"
	Found             = "Found"
	SeeOther          = "SeeOther"
	NotModified       = "NotModified"
	UseProxy          = "UseProxy"
	TemporaryRedirect = "TemporaryRedirect"

	BadRequest                   = "BadRequest"
	Unauthorized                 = "Unauthorized"
	PaymentRequired              = "PaymentRequired"
	Forbidden                    = "Forbidden"
	NotFound                     = "NotFound"
	MethodNotAllowed             = "MethodNotAllowed"
	NotAcceptable                = "NotAcceptable"
	ProxyAuthRequired            = "ProxyAuthRequired"
	RequestTimeout               = "RequestTimeout"
	Conflict                     = "Conflict"
	Gone                         = "Gone"
	LengthRequired               = "LengthRequired"
	PreconditionFailed           = "PreconditionFailed"
	RequestEntityTooLarge        = "RequestEntityTooLarge"
	RequestURITooLong            = "RequestURITooLong"
	UnsupportedMediaType         = "UnsupportedMediaType"
	RequestedRangeNotSatisfiable = "RequestedRangeNotSatisfiable"
	ExpectationFailed            = "ExpectationFailed"
	Teapot                       = "Teapot"

	InternalServerError     = "InternalServerError"
	NotImplemented          = "NotImplemented"
	BadGateway              = "BadGateway"
	ServiceUnavailable      = "ServiceUnavailable"
	GatewayTimeout          = "GatewayTimeout"
	HTTPVersionNotSupported = "HTTPVersionNotSupported"
)

// Response records a possible action response.
func Response(name string, paramsAndDSL ...interface{}) {
	if a, ok := actionDefinition(true); ok {
		if a.Responses == nil {
			a.Responses = make(map[string]*ResponseDefinition)
		}
		if _, ok := a.Responses[name]; ok {
			ReportError("response %s is defined twice", name)
			return
		}
		var params []string
		var dsl func()
		if len(paramsAndDSL) > 0 {
			d := paramsAndDSL[len(paramsAndDSL)-1]
			if dsl, ok = d.(func()); ok {
				paramsAndDSL = paramsAndDSL[:len(paramsAndDSL)-1]
			}
			params = make([]string, len(paramsAndDSL))
			for i, p := range paramsAndDSL {
				params[i], ok = p.(string)
				if !ok {
					ReportError("invalid response template parameter %#v, must be a string", p)
					return
				}
			}
		}
		var resp *ResponseDefinition
		if len(params) > 0 {
			if tmpl, ok := Design.ResponseTemplates[name]; ok {
				resp = tmpl.Template(params...)
			} else {
				ReportError("no response template named %#v", name)
				return
			}
		} else {
			if ar, ok := Design.Responses[name]; ok {
				resp = ar.Dup()
			} else {
				resp = &ResponseDefinition{Name: name}
			}
		}
		if (dsl != nil) && !executeDSL(dsl, resp) {
			return
		}
		resp.Parent = a
		a.Responses[name] = resp
	}
}

// Status sets the Response status
func Status(status int) {
	if r, ok := responseDefinition(true); ok {
		r.Status = status
	}
}

// Name sets the name of the response.
// Useful when using response templates to override the template name.
func Name(name string) {
	if r, ok := responseDefinition(true); ok {
		delete(Design.Responses, r.Name)
		r.Name = name
		Design.Responses[name] = r
	}
}
