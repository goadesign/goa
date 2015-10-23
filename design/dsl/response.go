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
	if a, ok := actionDefinition(false); ok {
		if a.Responses == nil {
			a.Responses = make(map[string]*ResponseDefinition)
		}
		if _, ok := a.Responses[name]; ok {
			ReportError("response %s is defined twice", name)
			return
		}
		if resp := executeResponseDSL(name, paramsAndDSL...); resp != nil {
			if resp.Status == 200 && resp.MediaType == "" {
				resp.MediaType = a.Parent.MediaType
			}
			resp.Parent = a
			a.Responses[name] = resp
		}
	} else if r, ok := resourceDefinition(true); ok {
		if r.Responses == nil {
			r.Responses = make(map[string]*ResponseDefinition)
		}
		if _, ok := r.Responses[name]; ok {
			ReportError("response %s is defined twice", name)
			return
		}
		if resp := executeResponseDSL(name, paramsAndDSL...); resp != nil {
			if resp.Status == 200 && resp.MediaType == "" {
				resp.MediaType = r.MediaType
			}
			resp.Parent = r
			r.Responses[name] = resp
		}
	}
}

func executeResponseDSL(name string, paramsAndDSL ...interface{}) *ResponseDefinition {
	var params []string
	var dsl func()
	var ok bool
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
				return nil
			}
		}
	}
	var resp *ResponseDefinition
	if len(params) > 0 {
		if tmpl, ok := Design.ResponseTemplates[name]; ok {
			resp = tmpl.Template(params...)
		} else if tmpl, ok := Design.DefaultResponseTemplates[name]; ok {
			resp = tmpl.Template(params...)
		} else {
			ReportError("no response template named %#v", name)
			return nil
		}
	} else {
		if ar, ok := Design.Responses[name]; ok {
			resp = ar.Dup()
		} else if ar, ok := Design.DefaultResponses[name]; ok {
			resp = ar.Dup()
		} else {
			resp = &ResponseDefinition{Name: name}
		}
	}
	if (dsl != nil) && !executeDSL(dsl, resp) {
		return nil
	}
	return resp
}

// Status sets the Response status
func Status(status int) {
	if r, ok := responseDefinition(true); ok {
		r.Status = status
	}
}
