package dsl

import (
	"net/http"

	"github.com/raphael/goa/design"
)

// Call InitDesign by default.
func init() {
	InitDesign()
}

// InitDesign initializes the Design global variable and loads the built-in
// response templates. This is a public function mainly so it can be used in tests.
func InitDesign() {
	ctxStack = nil // mostly for tests
	api := &design.APIDefinition{
		APIVersionDefinition: &design.APIVersionDefinition{
			DefaultResponseTemplates: make(map[string]*design.ResponseTemplateDefinition),
		},
	}
	t := func(params ...string) *design.ResponseDefinition {
		if len(params) < 1 {
			ReportError("expected media type as argument when invoking response template OK")
			return nil
		}
		return &design.ResponseDefinition{
			Name:      OK,
			Status:    200,
			MediaType: params[0],
		}
	}
	api.DefaultResponseTemplates[OK] = &design.ResponseTemplateDefinition{
		Name:     OK,
		Template: t,
	}

	api.DefaultResponses = make(map[string]*design.ResponseDefinition)
	for _, p := range []struct {
		status int
		name   string
	}{
		{100, Continue},
		{101, SwitchingProtocols},
		{200, OK},
		{201, Created},
		{202, Accepted},
		{203, NonAuthoritativeInfo},
		{204, NoContent},
		{205, ResetContent},
		{206, PartialContent},
		{300, MultipleChoices},
		{301, MovedPermanently},
		{302, Found},
		{303, SeeOther},
		{304, NotModified},
		{305, UseProxy},
		{307, TemporaryRedirect},
		{400, BadRequest},
		{401, Unauthorized},
		{402, PaymentRequired},
		{403, Forbidden},
		{404, NotFound},
		{405, MethodNotAllowed},
		{406, NotAcceptable},
		{407, ProxyAuthRequired},
		{408, RequestTimeout},
		{409, Conflict},
		{410, Gone},
		{411, LengthRequired},
		{412, PreconditionFailed},
		{413, RequestEntityTooLarge},
		{414, RequestURITooLong},
		{415, UnsupportedMediaType},
		{416, RequestedRangeNotSatisfiable},
		{417, ExpectationFailed},
		{418, Teapot},
		{500, InternalServerError},
		{501, NotImplemented},
		{502, BadGateway},
		{503, ServiceUnavailable},
		{504, GatewayTimeout},
		{505, HTTPVersionNotSupported},
	} {
		api.DefaultResponses[p.name] = &design.ResponseDefinition{
			Name:        p.name,
			Description: http.StatusText(p.status),
			Status:      p.status,
		}
	}

	// Initialize package variables
	design.Design = api
	design.Roots = []design.Root{api}
	design.GeneratedMediaTypes = nil
}
