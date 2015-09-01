package dsl

import (
	"fmt"

	. "github.com/raphael/goa/design"
)

const (
	Continue           = "StatusContinue"
	SwitchingProtocols = "StatusSwitchingProtocols"

	OK                   = "StatusOK"
	Created              = "StatusCreated"
	Accepted             = "StatusAccepted"
	NonAuthoritativeInfo = "StatusNonAuthoritativeInfo"
	NoContent            = "StatusNoContent"
	ResetContent         = "StatusResetContent"
	PartialContent       = "StatusPartialContent"

	MultipleChoices   = "StatusMultipleChoices"
	MovedPermanently  = "StatusMovedPermanently"
	Found             = "StatusFound"
	SeeOther          = "StatusSeeOther"
	NotModified       = "StatusNotModified"
	UseProxy          = "StatusUseProxy"
	TemporaryRedirect = "StatusTemporaryRedirect"

	BadRequest                   = "StatusBadRequest"
	Unauthorized                 = "StatusUnauthorized"
	PaymentRequired              = "StatusPaymentRequired"
	Forbidden                    = "StatusForbidden"
	NotFound                     = "StatusNotFound"
	MethodNotAllowed             = "StatusMethodNotAllowed"
	NotAcceptable                = "StatusNotAcceptable"
	ProxyAuthRequired            = "StatusProxyAuthRequired"
	RequestTimeout               = "StatusRequestTimeout"
	Conflict                     = "StatusConflict"
	Gone                         = "StatusGone"
	LengthRequired               = "StatusLengthRequired"
	PreconditionFailed           = "StatusPreconditionFailed"
	RequestEntityTooLarge        = "StatusRequestEntityTooLarge"
	RequestURITooLong            = "StatusRequestURITooLong"
	UnsupportedMediaType         = "StatusUnsupportedMediaType"
	RequestedRangeNotSatisfiable = "StatusRequestedRangeNotSatisfiable"
	ExpectationFailed            = "StatusExpectationFailed"
	Teapot                       = "StatusTeapot"

	InternalServerError     = "StatusInternalServerError"
	NotImplemented          = "StatusNotImplemented"
	BadGateway              = "StatusBadGateway"
	ServiceUnavailable      = "StatusServiceUnavailable"
	GatewayTimeout          = "StatusGatewayTimeout"
	HTTPVersionNotSupported = "StatusHTTPVersionNotSupported"
)

// init loads the built-in response templates.
func init() {
	Design.ResponseTemplateFuncs["OK"] = func(params ...string) *ResponseTemplateDefinition {
		if len(params) < 1 {
			appendError(fmt.Errorf("expected media type as argument when invoking response template OK"))
			return nil
		} else {
			return &ResponseTemplateDefinition{
				Name:      "OK",
				Status:    200,
				MediaType: params[0],
			}
		}
	}

	Design.ResponseTemplates["Continue"] = &ResponseTemplateDefinition{
		Name:   "Continue",
		Status: 100,
	}

	Design.ResponseTemplates["SwitchingProtocols"] = &ResponseTemplateDefinition{
		Name:   "SwitchingProtocols",
		Status: 101,
	}

	Design.ResponseTemplates["Created"] = &ResponseTemplateDefinition{
		Name:   "Created",
		Status: 201,
	}

	Design.ResponseTemplates["Accepted"] = &ResponseTemplateDefinition{
		Name:   "Accepted",
		Status: 202,
	}

	Design.ResponseTemplates["NonAuthoritativeInfo"] = &ResponseTemplateDefinition{
		Name:   "NonAuthoritativeInfo",
		Status: 203,
	}

	Design.ResponseTemplates["NoContent"] = &ResponseTemplateDefinition{
		Name:   "NoContent",
		Status: 204,
	}

	Design.ResponseTemplates["ResetContent"] = &ResponseTemplateDefinition{
		Name:   "ResetContent",
		Status: 205,
	}

	Design.ResponseTemplates["PartialContent"] = &ResponseTemplateDefinition{
		Name:   "PartialContent",
		Status: 206,
	}

	Design.ResponseTemplates["MultipleChoices"] = &ResponseTemplateDefinition{
		Name:   "MultipleChoices",
		Status: 300,
	}

	Design.ResponseTemplates["MovedPermanently"] = &ResponseTemplateDefinition{
		Name:   "MovedPermanently",
		Status: 301,
	}

	Design.ResponseTemplates["Found"] = &ResponseTemplateDefinition{
		Name:   "Found",
		Status: 302,
	}

	Design.ResponseTemplates["SeeOther"] = &ResponseTemplateDefinition{
		Name:   "SeeOther",
		Status: 303,
	}

	Design.ResponseTemplates["NotModified"] = &ResponseTemplateDefinition{
		Name:   "NotModified",
		Status: 304,
	}

	Design.ResponseTemplates["UseProxy"] = &ResponseTemplateDefinition{
		Name:   "UseProxy",
		Status: 305,
	}

	Design.ResponseTemplates["TemporaryRedirect"] = &ResponseTemplateDefinition{
		Name:   "TemporaryRedirect",
		Status: 307,
	}

	Design.ResponseTemplates["BadRequest"] = &ResponseTemplateDefinition{
		Name:   "BadRequest",
		Status: 400,
	}

	Design.ResponseTemplates["Unauthorized"] = &ResponseTemplateDefinition{
		Name:   "Unauthorized",
		Status: 401,
	}

	Design.ResponseTemplates["PaymentRequired"] = &ResponseTemplateDefinition{
		Name:   "PaymentRequired",
		Status: 402,
	}

	Design.ResponseTemplates["Forbidden"] = &ResponseTemplateDefinition{
		Name:   "Forbidden",
		Status: 403,
	}

	Design.ResponseTemplates["NotFound"] = &ResponseTemplateDefinition{
		Name:   "NotFound",
		Status: 404,
	}

	Design.ResponseTemplates["MethodNotAllowed"] = &ResponseTemplateDefinition{
		Name:   "MethodNotAllowed",
		Status: 405,
	}

	Design.ResponseTemplates["NotAcceptable"] = &ResponseTemplateDefinition{
		Name:   "NotAcceptable",
		Status: 406,
	}

	Design.ResponseTemplates["ProxyAuthRequired"] = &ResponseTemplateDefinition{
		Name:   "ProxyAuthRequired",
		Status: 407,
	}

	Design.ResponseTemplates["RequestTimeout"] = &ResponseTemplateDefinition{
		Name:   "RequestTimeout",
		Status: 408,
	}

	Design.ResponseTemplates["Conflict"] = &ResponseTemplateDefinition{
		Name:   "Conflict",
		Status: 409,
	}

	Design.ResponseTemplates["Gone"] = &ResponseTemplateDefinition{
		Name:   "Gone",
		Status: 410,
	}

	Design.ResponseTemplates["LengthRequired"] = &ResponseTemplateDefinition{
		Name:   "LengthRequired",
		Status: 411,
	}

	Design.ResponseTemplates["PreconditionFailed"] = &ResponseTemplateDefinition{
		Name:   "PreconditionFailed",
		Status: 412,
	}

	Design.ResponseTemplates["RequestEntityTooLarge"] = &ResponseTemplateDefinition{
		Name:   "RequestEntityTooLarge",
		Status: 413,
	}

	Design.ResponseTemplates["RequestURITooLong"] = &ResponseTemplateDefinition{
		Name:   "RequestURITooLong",
		Status: 414,
	}

	Design.ResponseTemplates["UnsupportedMediaType"] = &ResponseTemplateDefinition{
		Name:   "UnsupportedMediaType",
		Status: 415,
	}

	Design.ResponseTemplates["RequestedRangeNotSatisfiable"] = &ResponseTemplateDefinition{
		Name:   "RequestedRangeNotSatisfiable",
		Status: 416,
	}

	Design.ResponseTemplates["ExpectationFailed"] = &ResponseTemplateDefinition{
		Name:   "ExpectationFailed",
		Status: 417,
	}

	Design.ResponseTemplates["Teapot"] = &ResponseTemplateDefinition{
		Name:   "Teapot",
		Status: 418,
	}

	Design.ResponseTemplates["InternalServerError"] = &ResponseTemplateDefinition{
		Name:   "InternalServerError",
		Status: 500,
	}

	Design.ResponseTemplates["NotImplemented"] = &ResponseTemplateDefinition{
		Name:   "NotImplemented",
		Status: 501,
	}

	Design.ResponseTemplates["BadGateway"] = &ResponseTemplateDefinition{
		Name:   "BadGateway",
		Status: 502,
	}

	Design.ResponseTemplates["ServiceUnavailable"] = &ResponseTemplateDefinition{
		Name:   "ServiceUnavailable",
		Status: 503,
	}

	Design.ResponseTemplates["GatewayTimeout"] = &ResponseTemplateDefinition{
		Name:   "GatewayTimeout",
		Status: 504,
	}

	Design.ResponseTemplates["HTTPVersionNotSupported"] = &ResponseTemplateDefinition{
		Name:   "HTTPVersionNotSupported",
		Status: 505,
	}
}

// Status sets the ResponseTemplate status
func Status(status int) error {
	if r, ok := responseDefinition(true); ok {
		r.Status = status
	}
	return nil
}
