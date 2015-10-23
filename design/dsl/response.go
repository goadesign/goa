package dsl

import . "github.com/raphael/goa/design"

// Response implements the response definition DSL. Response takes the name of the response as
// first parameter. goa defines all the standard HTTP status name as global variables so they can be
// readily used as response names. Response also accepts optional arguments that correspond to the
// arguments defined by the corresponding response template (the response template with the same
// name) if there is one, see ResponseTemplate.
//
// A response may also optionally use an anonymous function as last argument to specify the response
// status code, media type and headers overriding what the default response or response template
// specifies:
//
// 	Response(OK, "vnd.goa.bottle", func() {  // OK response template accepts one argument: the media type identifier
// 		Headers(func() {                 // Headers list the response HTTP headers, see Headers
// 			Header("X-Request-Id")
// 		})
// 	})
//
//	Response(NotFound, func() {
//		Status(404)               // Not necessary as defined by default NotFound response.
//		Media("application/json") // Override NotFound response default of "text/plain"
//	})
//
//	Response(Created, func() {
//		Media(BottleMedia)
//	})
//
// goa defines a default response for all the HTTP status code. The default response simply sets
// the status code. So if an action can return NotFound for example all it has to do is specify
// Response(NotFound) - there is no need to specify the status code as the default response already
// does it, in other words:
//
//	Response(NotFound)
//
// is equivalent to:
//
//	Response(NotFound, func() {
//		Status(404)
//	})
//
// goa also defines a default response template for the OK response which takes a single argument:
// the identifier of the media type used to render the response. The API DSL can define additional
// response templates or override the default OK response template using ResponseTemplate.
//
// The media type identifier specified in a response definition via the Media function can be
// "generic" such as "text/plain" or "application/json" or can correspond to the identifier of a
// media type defined in the API DSL. In this latter case goa uses the media type definition to
// generate helper response methods. These methods know how to render the views defined on the media
// type and run the validations defined in the media type during rendering.
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

// Status sets the Response status.
func Status(status int) {
	if r, ok := responseDefinition(true); ok {
		r.Status = status
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
