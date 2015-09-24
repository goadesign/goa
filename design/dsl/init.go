package dsl

import . "github.com/raphael/goa/design"

// InitDesign initializes the Design global variable and loads the built-in
// response templates.
func InitDesign() {
	Design = &APIDefinition{}
	Design.ResponseTemplates = make(map[string]*ResponseTemplateDefinition)
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
	Design.ResponseTemplates[OK] = &ResponseTemplateDefinition{
		Name:     OK,
		Template: t,
	}

	Design.Responses = make(map[string]*ResponseDefinition)
	Design.Responses[Continue] = &ResponseDefinition{
		Name:   Continue,
		Status: 100,
	}

	Design.Responses[SwitchingProtocols] = &ResponseDefinition{
		Name:   SwitchingProtocols,
		Status: 101,
	}

	Design.Responses[Created] = &ResponseDefinition{
		Name:   Created,
		Status: 201,
	}

	Design.Responses[Accepted] = &ResponseDefinition{
		Name:   Accepted,
		Status: 202,
	}

	Design.Responses[NonAuthoritativeInfo] = &ResponseDefinition{
		Name:   NonAuthoritativeInfo,
		Status: 203,
	}

	Design.Responses[NoContent] = &ResponseDefinition{
		Name:   NoContent,
		Status: 204,
	}

	Design.Responses[ResetContent] = &ResponseDefinition{
		Name:   ResetContent,
		Status: 205,
	}

	Design.Responses[PartialContent] = &ResponseDefinition{
		Name:   PartialContent,
		Status: 206,
	}

	Design.Responses[MultipleChoices] = &ResponseDefinition{
		Name:   MultipleChoices,
		Status: 300,
	}

	Design.Responses[MovedPermanently] = &ResponseDefinition{
		Name:   MovedPermanently,
		Status: 301,
	}

	Design.Responses[Found] = &ResponseDefinition{
		Name:   Found,
		Status: 302,
	}

	Design.Responses[SeeOther] = &ResponseDefinition{
		Name:   SeeOther,
		Status: 303,
	}

	Design.Responses[NotModified] = &ResponseDefinition{
		Name:   NotModified,
		Status: 304,
	}

	Design.Responses[UseProxy] = &ResponseDefinition{
		Name:   UseProxy,
		Status: 305,
	}

	Design.Responses[TemporaryRedirect] = &ResponseDefinition{
		Name:   TemporaryRedirect,
		Status: 307,
	}

	Design.Responses[BadRequest] = &ResponseDefinition{
		Name:   BadRequest,
		Status: 400,
	}

	Design.Responses[Unauthorized] = &ResponseDefinition{
		Name:   Unauthorized,
		Status: 401,
	}

	Design.Responses[PaymentRequired] = &ResponseDefinition{
		Name:   PaymentRequired,
		Status: 402,
	}

	Design.Responses[Forbidden] = &ResponseDefinition{
		Name:   Forbidden,
		Status: 403,
	}

	Design.Responses[NotFound] = &ResponseDefinition{
		Name:   NotFound,
		Status: 404,
	}

	Design.Responses[MethodNotAllowed] = &ResponseDefinition{
		Name:   MethodNotAllowed,
		Status: 405,
	}

	Design.Responses[NotAcceptable] = &ResponseDefinition{
		Name:   NotAcceptable,
		Status: 406,
	}

	Design.Responses[ProxyAuthRequired] = &ResponseDefinition{
		Name:   ProxyAuthRequired,
		Status: 407,
	}

	Design.Responses[RequestTimeout] = &ResponseDefinition{
		Name:   RequestTimeout,
		Status: 408,
	}

	Design.Responses[Conflict] = &ResponseDefinition{
		Name:   Conflict,
		Status: 409,
	}

	Design.Responses[Gone] = &ResponseDefinition{
		Name:   Gone,
		Status: 410,
	}

	Design.Responses[LengthRequired] = &ResponseDefinition{
		Name:   LengthRequired,
		Status: 411,
	}

	Design.Responses[PreconditionFailed] = &ResponseDefinition{
		Name:   PreconditionFailed,
		Status: 412,
	}

	Design.Responses[RequestEntityTooLarge] = &ResponseDefinition{
		Name:   RequestEntityTooLarge,
		Status: 413,
	}

	Design.Responses[RequestURITooLong] = &ResponseDefinition{
		Name:   RequestURITooLong,
		Status: 414,
	}

	Design.Responses[UnsupportedMediaType] = &ResponseDefinition{
		Name:   UnsupportedMediaType,
		Status: 415,
	}

	Design.Responses[RequestedRangeNotSatisfiable] = &ResponseDefinition{
		Name:   RequestedRangeNotSatisfiable,
		Status: 416,
	}

	Design.Responses[ExpectationFailed] = &ResponseDefinition{
		Name:   ExpectationFailed,
		Status: 417,
	}

	Design.Responses[Teapot] = &ResponseDefinition{
		Name:   Teapot,
		Status: 418,
	}

	Design.Responses[InternalServerError] = &ResponseDefinition{
		Name:   InternalServerError,
		Status: 500,
	}

	Design.Responses[NotImplemented] = &ResponseDefinition{
		Name:   NotImplemented,
		Status: 501,
	}

	Design.Responses[BadGateway] = &ResponseDefinition{
		Name:   BadGateway,
		Status: 502,
	}

	Design.Responses[ServiceUnavailable] = &ResponseDefinition{
		Name:   ServiceUnavailable,
		Status: 503,
	}

	Design.Responses[GatewayTimeout] = &ResponseDefinition{
		Name:   GatewayTimeout,
		Status: 504,
	}

	Design.Responses[HTTPVersionNotSupported] = &ResponseDefinition{
		Name:   HTTPVersionNotSupported,
		Status: 505,
	}
}
