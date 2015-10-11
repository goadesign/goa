package dsl

import . "github.com/raphael/goa/design"

// InitDesign initializes the Design global variable and loads the built-in
// response templates.
func InitDesign() {
	Design = &APIDefinition{}
	Design.DefaultResponseTemplates = make(map[string]*ResponseTemplateDefinition)
	t := func(params ...string) *ResponseDefinition {
		if len(params) < 1 {
			ReportError("expected media type as argument when invoking response template OK")
			return nil
		}
		return &ResponseDefinition{
			Name:      OK,
			Status:    200,
			MediaType: params[0],
		}
	}
	Design.DefaultResponseTemplates[OK] = &ResponseTemplateDefinition{
		Name:     OK,
		Template: t,
	}

	Design.DefaultResponses = make(map[string]*ResponseDefinition)
	Design.DefaultResponses[Continue] = &ResponseDefinition{
		Name:   Continue,
		Status: 100,
	}

	Design.DefaultResponses[SwitchingProtocols] = &ResponseDefinition{
		Name:   SwitchingProtocols,
		Status: 101,
	}

	Design.DefaultResponses[OK] = &ResponseDefinition{
		Name:   OK,
		Status: 200,
	}

	Design.DefaultResponses[Created] = &ResponseDefinition{
		Name:   Created,
		Status: 201,
	}

	Design.DefaultResponses[Accepted] = &ResponseDefinition{
		Name:   Accepted,
		Status: 202,
	}

	Design.DefaultResponses[NonAuthoritativeInfo] = &ResponseDefinition{
		Name:   NonAuthoritativeInfo,
		Status: 203,
	}

	Design.DefaultResponses[NoContent] = &ResponseDefinition{
		Name:   NoContent,
		Status: 204,
	}

	Design.DefaultResponses[ResetContent] = &ResponseDefinition{
		Name:   ResetContent,
		Status: 205,
	}

	Design.DefaultResponses[PartialContent] = &ResponseDefinition{
		Name:   PartialContent,
		Status: 206,
	}

	Design.DefaultResponses[MultipleChoices] = &ResponseDefinition{
		Name:   MultipleChoices,
		Status: 300,
	}

	Design.DefaultResponses[MovedPermanently] = &ResponseDefinition{
		Name:   MovedPermanently,
		Status: 301,
	}

	Design.DefaultResponses[Found] = &ResponseDefinition{
		Name:   Found,
		Status: 302,
	}

	Design.DefaultResponses[SeeOther] = &ResponseDefinition{
		Name:   SeeOther,
		Status: 303,
	}

	Design.DefaultResponses[NotModified] = &ResponseDefinition{
		Name:   NotModified,
		Status: 304,
	}

	Design.DefaultResponses[UseProxy] = &ResponseDefinition{
		Name:   UseProxy,
		Status: 305,
	}

	Design.DefaultResponses[TemporaryRedirect] = &ResponseDefinition{
		Name:   TemporaryRedirect,
		Status: 307,
	}

	Design.DefaultResponses[BadRequest] = &ResponseDefinition{
		Name:   BadRequest,
		Status: 400,
	}

	Design.DefaultResponses[Unauthorized] = &ResponseDefinition{
		Name:   Unauthorized,
		Status: 401,
	}

	Design.DefaultResponses[PaymentRequired] = &ResponseDefinition{
		Name:   PaymentRequired,
		Status: 402,
	}

	Design.DefaultResponses[Forbidden] = &ResponseDefinition{
		Name:   Forbidden,
		Status: 403,
	}

	Design.DefaultResponses[NotFound] = &ResponseDefinition{
		Name:   NotFound,
		Status: 404,
	}

	Design.DefaultResponses[MethodNotAllowed] = &ResponseDefinition{
		Name:   MethodNotAllowed,
		Status: 405,
	}

	Design.DefaultResponses[NotAcceptable] = &ResponseDefinition{
		Name:   NotAcceptable,
		Status: 406,
	}

	Design.DefaultResponses[ProxyAuthRequired] = &ResponseDefinition{
		Name:   ProxyAuthRequired,
		Status: 407,
	}

	Design.DefaultResponses[RequestTimeout] = &ResponseDefinition{
		Name:   RequestTimeout,
		Status: 408,
	}

	Design.DefaultResponses[Conflict] = &ResponseDefinition{
		Name:   Conflict,
		Status: 409,
	}

	Design.DefaultResponses[Gone] = &ResponseDefinition{
		Name:   Gone,
		Status: 410,
	}

	Design.DefaultResponses[LengthRequired] = &ResponseDefinition{
		Name:   LengthRequired,
		Status: 411,
	}

	Design.DefaultResponses[PreconditionFailed] = &ResponseDefinition{
		Name:   PreconditionFailed,
		Status: 412,
	}

	Design.DefaultResponses[RequestEntityTooLarge] = &ResponseDefinition{
		Name:   RequestEntityTooLarge,
		Status: 413,
	}

	Design.DefaultResponses[RequestURITooLong] = &ResponseDefinition{
		Name:   RequestURITooLong,
		Status: 414,
	}

	Design.DefaultResponses[UnsupportedMediaType] = &ResponseDefinition{
		Name:   UnsupportedMediaType,
		Status: 415,
	}

	Design.DefaultResponses[RequestedRangeNotSatisfiable] = &ResponseDefinition{
		Name:   RequestedRangeNotSatisfiable,
		Status: 416,
	}

	Design.DefaultResponses[ExpectationFailed] = &ResponseDefinition{
		Name:   ExpectationFailed,
		Status: 417,
	}

	Design.DefaultResponses[Teapot] = &ResponseDefinition{
		Name:   Teapot,
		Status: 418,
	}

	Design.DefaultResponses[InternalServerError] = &ResponseDefinition{
		Name:   InternalServerError,
		Status: 500,
	}

	Design.DefaultResponses[NotImplemented] = &ResponseDefinition{
		Name:   NotImplemented,
		Status: 501,
	}

	Design.DefaultResponses[BadGateway] = &ResponseDefinition{
		Name:   BadGateway,
		Status: 502,
	}

	Design.DefaultResponses[ServiceUnavailable] = &ResponseDefinition{
		Name:   ServiceUnavailable,
		Status: 503,
	}

	Design.DefaultResponses[GatewayTimeout] = &ResponseDefinition{
		Name:   GatewayTimeout,
		Status: 504,
	}

	Design.DefaultResponses[HTTPVersionNotSupported] = &ResponseDefinition{
		Name:   HTTPVersionNotSupported,
		Status: 505,
	}
}
