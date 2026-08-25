// This file prepares, validates, and finalizes the HTTP transport contract for
// one service method.
package expr

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa/v3/eval"
)

type (
	// HTTPEndpointExpr describes a HTTP endpoint. It embeds a MethodExpr and
	// adds HTTP specific properties.
	//
	// It defines both an HTTP endpoint and the shape of HTTP requests and
	// responses made to that endpoint. The shape of requests is defined via
	// "parameters", there are path parameters (i.e. portions of the URL that
	// define parameter values), query string parameters and a payload parameter
	// (request body).
	HTTPEndpointExpr struct {
		eval.DSLFunc
		// MethodExpr is the underlying method expression.
		MethodExpr *MethodExpr
		// Service is the parent service.
		Service *HTTPServiceExpr
		// Endpoint routes
		Routes []*RouteExpr
		// MapQueryParams - when not nil - indicates that the HTTP
		// request query string parameters are used to build a map.
		//    - If the value is the empty string then the map is stored
		//      in the method payload (which must be of type Map)
		//    - If the value is a non-empty string then the map is
		//      stored in the payload attribute with the corresponding
		//      name (which must of be of type Map)
		MapQueryParams *string
		// Params defines the HTTP request path and query parameters.
		Params *MappedAttributeExpr
		// Headers defines the HTTP request headers.
		Headers *MappedAttributeExpr
		// Cookies defines the HTTP request cookies.
		Cookies *MappedAttributeExpr
		// Body describes the HTTP request body.
		Body *AttributeExpr
		// StreamingBody describes the body transferred through the websocket
		// stream.
		StreamingBody *AttributeExpr
		// PayloadIDAttribute is retained for generator compatibility.
		//
		// Deprecated: inspect the payload field marked with "jsonrpc:id".
		PayloadIDAttribute string
		// ResultIDAttribute is retained for generator compatibility.
		//
		// Deprecated: JSON-RPC results cannot define an ID field.
		ResultIDAttribute string
		// JSONRPCNotification is true when generated clients send this method
		// without an ID and do not wait for a JSON-RPC response.
		JSONRPCNotification bool
		// SkipRequestBodyEncodeDecode indicates that the service method accepts
		// a reader and that the client provides a reader to stream the request
		// body.
		SkipRequestBodyEncodeDecode bool
		// SkipResponseBodyEncodeDecode indicates that the service method
		// returns a reader and that the client accepts a reader to stream the
		// response body.
		SkipResponseBodyEncodeDecode bool
		// Responses is the list of all the possible success HTTP
		// responses.
		Responses []*HTTPResponseExpr
		// HTTPErrors is the list of all the possible error HTTP
		// responses.
		HTTPErrors []*HTTPErrorExpr
		// Requirements contains the security requirements for the HTTP endpoint.
		Requirements []*SecurityExpr
		// MultipartRequest indicates that the request content type for
		// the endpoint is a multipart type.
		MultipartRequest bool
		// Redirect defines a redirect for the endpoint.
		Redirect *HTTPRedirectExpr
		// SSE defines the Server-Sent Events configuration for this endpoint if it's
		// a streaming endpoint. If nil, the endpoint uses WebSockets by default or
		// inherits the service-level SSE configuration if defined.
		SSE *HTTPSSEExpr
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator, see dsl.Meta.
		Meta MetaExpr
		// prepared is true if Prepare has been run. This field is required to
		// avoid infinite recursions.
		prepared bool
	}

	// RouteExpr represents an endpoint route (HTTP endpoint).
	RouteExpr struct {
		// Method is the HTTP method, e.g. "GET", "POST", etc.
		Method string
		// Path is the URL path e.g. "/tasks/{id}"
		Path string
		// Endpoint is the endpoint this route applies to.
		Endpoint *HTTPEndpointExpr
		// Meta is an arbitrary set of key/value pairs, see
		// dsl.Meta
		Meta MetaExpr
	}

	// jsonRPCIDField records where one ID declaration appears so validation can
	// reject declarations that code generation cannot represent.
	jsonRPCIDField struct {
		name      string
		attribute *AttributeExpr
		depth     int
	}
)

// A generated JSON-RPC request ID uses the standard 36-character UUID form.
const generatedJSONRPCRequestIDLength = 36

// Name of HTTP endpoint
func (e *HTTPEndpointExpr) Name() string {
	return e.MethodExpr.Name
}

// Description of HTTP endpoint
func (e *HTTPEndpointExpr) Description() string {
	return e.MethodExpr.Description
}

// EvalName returns the generic expression name used in error messages.
func (e *HTTPEndpointExpr) EvalName() string {
	var prefix, suffix string
	if e.Name() != "" {
		suffix = fmt.Sprintf("HTTP endpoint %#v", e.Name())
	} else {
		suffix = "unnamed HTTP endpoint"
	}
	if e.Service != nil {
		prefix = e.Service.EvalName() + " "
	}
	return prefix + suffix
}

// IsJSONRPC returns true if the endpoint is a JSON-RPC endpoint.
func (e *HTTPEndpointExpr) IsJSONRPC() bool {
	_, ok := e.Meta["jsonrpc"]
	return ok
}

// IsJSONRPCNotification reports whether generated clients call the JSON-RPC
// method without an ID and do not wait for a JSON-RPC response.
func (e *HTTPEndpointExpr) IsJSONRPCNotification() bool {
	return e.IsJSONRPC() && e.JSONRPCNotification
}

// UsesSSE returns true if the endpoint streams result events over Server-Sent
// Events.
func (e *HTTPEndpointExpr) UsesSSE() bool {
	return e.SSE != nil && (e.MethodExpr.IsResultStreaming() || e.MethodExpr.HasMixedResults())
}

// UsesWebSocket returns true if an ordinary HTTP endpoint streams payloads or
// results over a WebSocket connection.
func (e *HTTPEndpointExpr) UsesWebSocket() bool {
	return !e.IsJSONRPC() && e.MethodExpr.IsStreaming() && e.SSE == nil
}

// HasAbsoluteRoutes returns true if all the endpoint routes are absolute.
func (e *HTTPEndpointExpr) HasAbsoluteRoutes() bool {
	for _, r := range e.Routes {
		if !r.IsAbsolute() {
			return false
		}
	}
	return true
}

// PathParams computes a mapped attribute containing the subset of e.Params that
// describe path parameters.
func (e *HTTPEndpointExpr) PathParams() *MappedAttributeExpr {
	obj := Object{}
	v := &ValidationExpr{}
	pat := e.Params.Attribute() // need "attribute:name" style keys
	for _, r := range e.Routes {
		for _, p := range r.Params() {
			att := pat.Find(p)
			if att == nil {
				continue
			}
			obj.Set(p, att)
			if e.Params.IsRequired(p) {
				v.AddRequired(p)
			}
		}
	}
	at := &AttributeExpr{Type: &obj, Validation: v}
	return NewMappedAttributeExpr(at)
}

// QueryParams computes a mapped attribute containing the subset of e.Params
// that describe query parameters.
func (e *HTTPEndpointExpr) QueryParams() *MappedAttributeExpr {
	obj := Object{}
	v := &ValidationExpr{}
	pp := make(map[string]struct{})
	for _, r := range e.Routes {
		for _, p := range r.Params() {
			pp[p] = struct{}{}
		}
	}
	pat := e.Params.Attribute() // need "attribute:name" style keys
	for _, at := range *(pat.Type.(*Object)) {
		found := false
		for n := range pp {
			if n == at.Name {
				found = true
				break
			}
		}
		if !found {
			obj.Set(at.Name, at.Attribute)
			// when looking for required attributes we need the unmapped keys
			// (i.e. without the "attribute:name" syntax)
			attName := strings.Split(at.Name, ":")[0]
			if e.Params.IsRequired(attName) {
				v.AddRequired(attName)
			}
		}
	}
	at := &AttributeExpr{Type: &obj, Validation: v}
	return NewMappedAttributeExpr(at)
}

// Prepare computes the request path and query string parameters as well as the
// headers and body taking into account the inherited values from the service.
func (e *HTTPEndpointExpr) Prepare() {
	// Avoid infinite recursions when traversing parents.
	if e.prepared {
		return
	}
	e.prepared = true
	if e.Headers == nil {
		e.Headers = NewEmptyMappedAttributeExpr()
	}
	if e.Cookies == nil {
		e.Cookies = NewEmptyMappedAttributeExpr()
	}
	if e.Params == nil {
		e.Params = NewEmptyMappedAttributeExpr()
	}

	// Inherit headers, cookies and params from parent service and API
	headers := NewEmptyMappedAttributeExpr()
	headers.Merge(e.MethodExpr.Service.design.API.HTTP.Headers)
	headers.Merge(e.Service.Headers)

	cookies := NewEmptyMappedAttributeExpr()
	cookies.Merge(e.MethodExpr.Service.design.API.HTTP.Cookies)
	cookies.Merge(e.Service.Cookies)

	params := NewEmptyMappedAttributeExpr()
	params.Merge(e.MethodExpr.Service.design.API.HTTP.Params)
	params.Merge(e.Service.Params)

	if p := e.Service.Parent(); p != nil {
		if c := p.CanonicalEndpoint(); c != nil {
			c.Prepare()
			if !e.HasAbsoluteRoutes() {
				headers.Merge(c.Headers)
				cookies.Merge(c.Cookies)
				cpp := c.PathParams()
				params.Merge(cpp)

				// Inherit attributes for path params from parent service
				WalkMappedAttr(cpp, func(name, _ string, _ *AttributeExpr) error { // nolint: errcheck
					if att := c.MethodExpr.Payload.Find(name); att != nil {
						if e.MethodExpr.Payload.Type == Empty {
							e.MethodExpr.Payload.Type = &Object{}
						}
						if o := AsObject(e.MethodExpr.Payload.Type); o != nil && o.Attribute(name) == nil {
							if c.MethodExpr.Payload.IsRequired(name) {
								if e.MethodExpr.Payload.Validation == nil {
									e.MethodExpr.Payload.Validation = &ValidationExpr{}
								}
								e.MethodExpr.Payload.Validation.AddRequired(name)
							}
							o.Set(name, att)
						}
					}
					return nil
				})
			}
		}
	}
	headers.Merge(e.Headers)
	cookies.Merge(e.Cookies)
	params.Merge(e.Params)

	e.Headers = headers
	e.Cookies = cookies
	e.Params = params

	// Initialize path params that are not defined explicitly in
	for _, r := range e.Routes {
		for _, p := range r.Params() {
			if a := params.Find(p); a == nil {
				params.Merge(NewMappedAttributeExpr(&AttributeExpr{
					Type: &Object{
						&NamedAttributeExpr{
							Name:      p,
							Attribute: &AttributeExpr{Type: String},
						},
					},
				}))
			}
		}
	}

	// Make sure there's a default success response if none define explicitly.
	if len(e.Responses) == 0 {
		status := StatusOK
		if e.Redirect != nil {
			status = e.Redirect.StatusCode
		} else if e.MethodExpr.Result.Type == Empty && !e.SkipResponseBodyEncodeDecode {
			status = StatusNoContent
		}
		e.Responses = []*HTTPResponseExpr{{StatusCode: status}}
	}

	// Inherit SSE configuration from service or API level for streaming endpoints
	if e.MethodExpr.Stream == ServerStreamKind && e.SSE == nil {
		if e.Service.SSE != nil {
			e.SSE = e.Service.SSE
		} else if e.MethodExpr.Service.design.API.HTTP.SSE != nil {
			e.SSE = e.MethodExpr.Service.design.API.HTTP.SSE
		}
	}

	// Error -> ResponseError
	methodErrors := map[string]struct{}{}
	for _, v := range e.HTTPErrors {
		methodErrors[v.Name] = struct{}{}
	}
	for _, me := range e.MethodExpr.Errors {
		if _, ok := methodErrors[me.Name]; ok {
			continue
		}
		methodErrors[me.Name] = struct{}{}
		var found bool
		for _, v := range e.Service.HTTPErrors {
			if me.Name == v.Name {
				e.HTTPErrors = append(e.HTTPErrors, v.Dup())
				found = true
				break
			}
		}
		if found {
			continue
		}
		// Lookup undefined HTTP errors in API.
		for _, v := range e.Service.Root.Errors {
			if me.Name == v.Name {
				e.HTTPErrors = append(e.HTTPErrors, v.Dup())
			}
		}
	}
	// Inherit HTTP errors from service if the error has not added.
	for _, se := range e.Service.ServiceExpr.Errors {
		if _, ok := methodErrors[se.Name]; ok {
			continue
		}
		var found bool
		for _, resp := range e.Service.HTTPErrors {
			if se.Name == resp.Name {
				found = true
				e.HTTPErrors = append(e.HTTPErrors, resp.Dup())
				break
			}
		}
		if !found {
			for _, ae := range e.MethodExpr.Service.design.API.HTTP.Errors {
				if se.Name == ae.Name {
					e.HTTPErrors = append(e.HTTPErrors, ae.Dup())
					break
				}
			}
		}
	}

	// WebSocket endpoints use GET for the HTTP upgrade request.
	if e.UsesWebSocket() && len(e.Routes) > 0 {
		e.Routes[0].Method = "GET"
	}

	// Prepare responses
	for _, r := range e.Responses {
		r.Prepare()
	}
	for _, er := range e.HTTPErrors {
		er.Response.Prepare()
	}
}

// Validate validates the endpoint expression.
func (e *HTTPEndpointExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	// Name cannot be empty
	if e.Name() == "" {
		verr.Add(e, "Endpoint name cannot be empty")
	}

	// SkipRequestBodyEncodeDecode is not compatible with gRPC or WebSocket
	if e.SkipRequestBodyEncodeDecode {
		if s := e.MethodExpr.Service.design.API.GRPC.Service(e.Service.Name()); s != nil {
			if s.Endpoint(e.Name()) != nil {
				verr.Add(e, "Endpoint cannot use SkipRequestBodyEncodeDecode and define a gRPC transport.")
			}
		}
		if e.MethodExpr.IsPayloadStreaming() {
			verr.Add(e, "Endpoint cannot use SkipRequestBodyEncodeDecode when method defines a StreamingPayload.")
		}
		if e.MethodExpr.IsResultStreaming() {
			verr.Add(e, "Endpoint cannot use SkipRequestBodyEncodeDecode when method defines a StreamingResult. Use SkipResponseBodyEncodeDecode instead.")
		}
	}

	// SkipResponseBodyEncodeDecode is not compatible with gRPC or WebSocket.
	if e.SkipResponseBodyEncodeDecode {
		if s := e.MethodExpr.Service.design.API.GRPC.Service(e.Service.Name()); s != nil {
			if s.Endpoint(e.Name()) != nil {
				verr.Add(e, "Endpoint response cannot use SkipResponseBodyEncodeDecode and define a gRPC transport.")
			}
		}
		if e.MethodExpr.IsPayloadStreaming() {
			verr.Add(e, "Endpoint cannot use SkipResponseBodyEncodeDecode when method defines a StreamingPayload. Use SkipRequestBodyEncodeDecode instead.")
		}
		if e.MethodExpr.IsResultStreaming() {
			verr.Add(e, "Endpoint cannot use SkipResponseBodyEncodeDecode when method defines a StreamingResult.")
		}
		if rt, ok := e.MethodExpr.Result.Type.(*ResultTypeExpr); ok {
			if len(rt.Views) > 1 {
				verr.Add(e, "Endpoint cannot use SkipResponseBodyEncodeDecode when method result type defines multiple views.")
			}
		}
	}

	// A WebSocket client learns the result view from the connection handshake.
	// Receiving a streamed payload starts that connection before the service can
	// choose a result view, so the design must select the view in advance.
	if e.UsesWebSocket() && e.MethodExpr.IsPayloadStreaming() && !e.MethodExpr.HasMixedResults() {
		if result, ok := e.MethodExpr.Result.Type.(*ResultTypeExpr); ok {
			viewCount := len(result.Views)
			if result.View(DefaultView) == nil {
				viewCount++
			}
			_, selectedByMethod := e.MethodExpr.Result.Meta.Last(ViewMetaKey)
			_, selectedByType := result.Meta.Last(ViewMetaKey)
			if viewCount > 1 && !selectedByMethod && !selectedByType {
				verr.Add(e, "Endpoint cannot choose a result view at runtime when the method defines StreamingPayload because the WebSocket connection starts before the result view is known. Select a view in Result or StreamingResult.")
			}
		}
	}

	// Validate streaming endpoints for SSE compatibility
	if e.MethodExpr.Stream == ServerStreamKind {
		if e.SSE != nil {
			if err := e.SSE.Validate(e.MethodExpr); err != nil {
				var valErr *eval.ValidationErrors
				if errors.As(err, &valErr) {
					verr.Merge(valErr)
				}
			}
			if e.IsJSONRPC() && e.SSE.DataField != "" {
				result := e.MethodExpr.StreamingResult
				if result == nil {
					result = e.MethodExpr.Result
				}
				if _, viewed := result.Type.(*ResultTypeExpr); viewed {
					verr.Add(e, "SSE event data cannot select one field from a viewed streaming result because the selected data would omit the result view needed to decode the stream")
				}
			}
			if e.SSE.RequestIDField != "" {
				for _, field := range *AsObject(e.Headers.Type) {
					header := e.Headers.ElemName(field.Name)
					if field.Name == e.SSE.RequestIDField {
						verr.Add(e, "SSE request ID field %q cannot also be mapped with Header because ServerSentEvents maps it to Last-Event-ID", field.Name)
						continue
					}
					if strings.EqualFold(header, "Last-Event-ID") {
						verr.Add(e, "HTTP header %q is reserved for SSE request ID field %q", header, e.SSE.RequestIDField)
					}
				}
			}
		}
	}

	// Validate mixed results configuration
	if e.MethodExpr.HasMixedResults() {
		// A separate streaming result requires SSE.
		if e.SSE == nil {
			verr.Add(e, "Methods with both Result and StreamingResult must use ServerSentEvents()")
		}
		// Cannot have bidirectional streaming with mixed results
		if e.MethodExpr.IsPayloadStreaming() {
			verr.Add(e, "Methods with both Result and StreamingResult cannot have StreamingPayload")
		}
	} else if e.SSE != nil {
		// Error if SSE is defined but endpoint is not server streaming or mixed results
		switch e.MethodExpr.Stream {
		case BidirectionalStreamKind:
			verr.Add(e, "Server-Sent Events cannot be used with bidirectional streaming endpoints")
		case ClientStreamKind:
			verr.Add(e, "Server-Sent Events cannot be used with client-to-server streaming endpoints")
		case NoStreamKind:
			// SSE requires either server streaming or mixed results
			if !e.MethodExpr.HasMixedResults() {
				verr.Add(e, "Server-Sent Events can only be used with endpoints that have a streaming result or mixed results")
			}
			// case ServerStreamKind is valid, no error
		}
	}

	// JSON-RPC validation
	if e.IsJSONRPC() {
		if strings.HasPrefix(e.MethodExpr.Name, "rpc.") {
			verr.Add(e, "JSON-RPC method %q cannot begin with %q because JSON-RPC reserves that namespace", e.MethodExpr.Name, "rpc.")
		}
		requestIDs := jsonRPCIDFields(e.MethodExpr.Payload)
		if len(requestIDs) > 1 {
			verr.Add(e, "JSON-RPC method %q cannot define more than one request ID field", e.MethodExpr.Name)
		}
		for _, id := range requestIDs {
			if id.depth != 1 {
				verr.Add(e, "JSON-RPC request ID field %q must be a direct payload field", id.name)
			}
			if !isJSONRPCIDString(id.attribute.Type) {
				verr.Add(e, "JSON-RPC request ID field %q must be a string", id.name)
			}
		}
		if len(requestIDs) == 1 {
			id := requestIDs[0]
			name := id.name
			if id.depth == 1 && isJSONRPCIDString(id.attribute.Type) &&
				!e.MethodExpr.Payload.IsRequired(name) && e.MethodExpr.Payload.GetDefault(name) == nil {
				if incompatible := incompatibleGeneratedJSONRPCRequestIDValidations(id.attribute); len(incompatible) > 0 {
					verr.Add(e, "JSON-RPC request ID field %q is optional and has no default, so generated clients create a UUID when it is absent; the following validation rules may reject that UUID: %s; make the field required or give it a default", name, strings.Join(incompatible, ", "))
				}
			}
			if e.Body != nil && jsonRPCBodyContainsID(e.Body, name) {
				verr.Add(e, "JSON-RPC request ID field %q cannot also appear in params", name)
			}
			if _, ok := e.Params.FindKey(name); ok {
				verr.Add(e, "JSON-RPC request ID field %q cannot also be mapped as an HTTP parameter", name)
			}
			if _, ok := e.Headers.FindKey(name); ok {
				verr.Add(e, "JSON-RPC request ID field %q cannot also be mapped as an HTTP header", name)
			}
			if _, ok := e.Cookies.FindKey(name); ok {
				verr.Add(e, "JSON-RPC request ID field %q cannot also be mapped as an HTTP cookie", name)
			}
			if e.SSE != nil && e.SSE.RequestIDField == name {
				verr.Add(e, "JSON-RPC request ID field %q cannot also be mapped to the Last-Event-ID header", name)
			}
		}
		if len(jsonRPCIDFields(e.MethodExpr.Result)) > 0 || len(jsonRPCIDFields(e.MethodExpr.StreamingResult)) > 0 {
			verr.Add(e, "JSON-RPC method %q cannot define an ID field in its result because the transport copies the request ID", e.MethodExpr.Name)
		}
		seenErrors := make(map[string]struct{})
		for _, mapped := range e.HTTPErrors {
			if _, ok := seenErrors[mapped.Name]; ok {
				continue
			}
			seenErrors[mapped.Name] = struct{}{}
			designError := e.MethodExpr.Error(mapped.Name)
			if designError != nil && len(jsonRPCIDFields(designError.AttributeExpr)) > 0 {
				verr.Add(e, "JSON-RPC error %q cannot define an ID field because the transport copies the request ID", mapped.Name)
			}
		}
		if e.IsJSONRPCNotification() {
			if e.MethodExpr.Result.Type != Empty {
				verr.Add(e, "JSON-RPC notification %q cannot define a result because notifications receive no response", e.MethodExpr.Name)
			}
			if e.MethodExpr.IsStreaming() {
				verr.Add(e, "JSON-RPC notification %q cannot stream because notifications send one message and receive no response", e.MethodExpr.Name)
			}
			if hasJSONRPCIDField(e.MethodExpr.Payload) || hasJSONRPCIDField(e.MethodExpr.StreamingPayload) {
				verr.Add(e, "JSON-RPC notification %q cannot define an ID field because notifications omit the request ID", e.MethodExpr.Name)
			}
		}
		if e.MethodExpr.HasMixedResults() {
			verr.Add(e, "JSON-RPC method %q cannot define both Result and StreamingResult because its client stream cannot return a separate final result", e.MethodExpr.Name)
		}
		switch e.MethodExpr.Stream {
		case ClientStreamKind:
			verr.Add(e, "JSON-RPC method %q cannot use client streaming because one JSON-RPC request contains one params value", e.MethodExpr.Name)
		case BidirectionalStreamKind:
			verr.Add(e, "JSON-RPC method %q cannot use bidirectional streaming because one JSON-RPC request contains one params value", e.MethodExpr.Name)
		case ServerStreamKind:
			if e.SSE == nil {
				verr.Add(e, "JSON-RPC method %q with a streaming result must use ServerSentEvents()", e.MethodExpr.Name)
			}
		}
	}

	// Redirect is not compatible with Response.
	if e.Redirect != nil {
		found := false
		for _, r := range e.Responses {
			if r.StatusCode != e.Redirect.StatusCode {
				found = true
				break
			}
		}
		if found {
			verr.Add(e, "Endpoint cannot use Response when using Redirect.")
		}
	}

	// Validate routes

	// Routes cannot be empty
	if len(e.Routes) == 0 {
		verr.Add(e, "No route defined for HTTP endpoint")
	} else {
		for _, r := range e.Routes {
			verr.Merge(r.Validate())
		}
		// Make sure that the same parameters are used in all routes
		params := e.Routes[0].Params()
		for _, r := range e.Routes[1:] {
			for _, p := range params {
				found := slices.Contains(r.Params(), p)
				if !found {
					verr.Add(e, "Param %q does not appear in all routes", p)
				}
			}
			for _, p2 := range r.Params() {
				found := slices.Contains(params, p2)
				if !found {
					verr.Add(e, "Param %q does not appear in all routes", p2)
				}
			}
		}
	}

	// Validate responses

	// All responses but one must have tags for the same status code
	hasTags := false
	allTagged := true
	successResp := false
	for i, r := range e.Responses {
		verr.Merge(r.Validate(e))
		for j, r2 := range e.Responses {
			if i != j && r.StatusCode == r2.StatusCode {
				verr.Add(r, "Multiple response definitions with status code %d", r.StatusCode)
			}
		}
		if r.Tag[0] == "" {
			allTagged = false
		} else {
			hasTags = true
		}
		if r.StatusCode < 400 {
			if e.MethodExpr.IsStreaming() {
				if !r.Headers.IsEmpty() {
					verr.Add(r, "streaming success response cannot map result attributes to HTTP headers")
				}
				if !r.Cookies.IsEmpty() {
					verr.Add(r, "streaming success response cannot map result attributes to HTTP cookies")
				}
			}
			if successResp && e.MethodExpr.Stream == ServerStreamKind {
				verr.Add(r, "At most one success response can be defined for a streaming endpoint.")
				if r.Body != nil && r.Body.Type == Empty {
					verr.Add(r, "Response body empty but endpoint defines streaming WebSocket response.")
				}
			} else if successResp && e.SkipResponseBodyEncodeDecode {
				verr.Add(r, "At most one success response can be defined for a endpoint using SkipResponseBodyEncodeDecode.")
			}
			successResp = true
		}
	}
	if hasTags && allTagged {
		verr.Add(e, "All responses define a Tag, at least one response must define no Tag.")
	}
	if hasTags && !IsObject(e.MethodExpr.Result.Type) {
		verr.Add(e, "Some responses define a Tag but the method Result type is not an object.")
	}

	// Make sure parameters and headers use compatible types
	verr.Merge(e.validateParams())
	verr.Merge(e.validateHeadersAndCookies())

	// Validate body attribute (required fields exist etc.)
	if e.Body != nil {
		verr.Merge(e.Body.Validate("HTTP body", e))
		if e.SkipRequestBodyEncodeDecode {
			verr.Add(e, "Cannot define a request body when using SkipRequestBodyEncodeDecode.")
		}
		// Make sure Body does not require attribute that are not required in
		// payload.
		if v := e.Body.Validation; v != nil {
			var preqs, missing []string
			if e.MethodExpr.Payload != nil && e.MethodExpr.Payload.Validation != nil {
				preqs = e.MethodExpr.Payload.Validation.Required
			}
			for _, req := range v.Required {
				found := slices.Contains(preqs, req)
				if !found {
					missing = append(missing, req)
				}
			}
			if len(missing) > 0 {
				is := "is"
				s := ""
				if len(missing) > 1 {
					is = "are"
					s = "s"
				}
				verr.Add(e, "The following HTTP request body attribute%s %s required but the corresponding method payload attribute%s %s not: %s. Use 'Required' to make the attribute%s required in the method payload as well.",
					s, is, s, is, strings.Join(missing, ", "), s)
			}
		}
	}

	// Validate errors
	for _, er := range e.HTTPErrors {
		verr.Merge(er.Validate())
	}
	verr.Merge(e.validateErrorMappings())

	// Validate definitions of params, headers and bodies against definition of payload
	var (
		hasParams  = !e.Params.IsEmpty()
		hasHeaders = !e.Headers.IsEmpty()
		hasCookies = !e.Cookies.IsEmpty()
	)
	if isEmpty(e.MethodExpr.Payload) {
		if e.MapQueryParams != nil {
			verr.Add(e, "MapParams is set but Payload is not defined")
		}
		if e.MultipartRequest {
			verr.Add(e, "MultipartRequest is set but Payload is not defined")
		}
		if !e.Params.IsEmpty() {
			verr.Add(e, "Params are set but Payload is not defined.")
		}
		if !e.Headers.IsEmpty() {
			verr.Add(e, "Headers are set but Payload is not defined.")
		}
		return verr
	}
	if IsArray(e.MethodExpr.Payload.Type) {
		if e.MapQueryParams != nil {
			verr.Add(e, "MapParams is set but Payload type is array. Payload type must be map or an object with a map attribute")
		}
		if hasParams && e.MultipartRequest {
			verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
		}
		if hasHeaders {
			if hasCookies || e.MultipartRequest {
				verr.Add(e, "Payload type is array but HTTP endpoint defines headers and MultipartRequest or cookies. At most one of these must be defined.")
			}
			if hasParams {
				verr.Add(e, "Payload type is array but HTTP endpoint defines both route or query string parameters and headers. At most one parameter or header must be defined and it must be of type array.")
			}
			if !IsPrimitive(AsArray(e.MethodExpr.Payload.Type).ElemType.Type) {
				verr.Add(e, "Array payloads used in HTTP headers must be of arrays of primitive types.")
			}
		}
		if e.Body != nil && e.Body.Type != Empty {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
			}
			if !IsArray(e.Body.Type) {
				verr.Add(e, "Payload type is array but HTTP endpoint body is not.")
			}
			if hasParams {
				verr.Add(e, "Payload type is array but HTTP endpoint defines both a body and route or query string parameters. At most one of these must be defined and it must be an array.")
			}
			if hasHeaders {
				verr.Add(e, "Payload type is array but HTTP endpoint defines both a body and headers. At most one of these must be defined and it must be an array.")
			}
		}
		if !hasParams && !hasHeaders && e.SkipRequestBodyEncodeDecode {
			verr.Add(e, "Payload type is array but HTTP endpoint uses SkipRequestBodyEncodeDecode and does not define headers or params.")
		}
	}

	if pMap := AsMap(e.MethodExpr.Payload.Type); pMap != nil {
		if e.MapQueryParams != nil {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and MapParams. At most one of these must be defined.")
			}
			if *e.MapQueryParams != "" {
				verr.Add(e, "MapParams is set to an attribute in the Payload but Payload is a map. Payload must be an object with an attribute of map type")
			}
			if !IsPrimitive(pMap.KeyType.Type) {
				verr.Add(e, "MapParams is set and Payload type is map. But payload key type must be a primitive")
			}
			if !IsPrimitive(pMap.ElemType.Type) && !IsArray(pMap.ElemType.Type) {
				verr.Add(e, "MapParams is set and Payload type is map. But payload element type must be a primitive or array")
			}
			if IsArray(pMap.ElemType.Type) && !IsPrimitive(AsArray(pMap.ElemType.Type).ElemType.Type) {
				verr.Add(e, "MapParams is set and Payload type is map. But array elements in payload element type must be primitive")
			}
		}
		if hasParams && e.MultipartRequest {
			verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
		}
		if e.Body != nil && e.Body.Type != Empty {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
			}
			if !IsMap(e.Body.Type) {
				verr.Add(e, "Payload type is map but HTTP endpoint body is not.")
			}
			if hasParams {
				verr.Add(e, "Payload type is map but HTTP endpoint defines both a body and route or query string parameters. At most one of these must be defined and it must be a map.")
			}
		}
		if !hasParams && e.SkipRequestBodyEncodeDecode {
			verr.Add(e, "Payload type is map but HTTP endpoint uses SkipRequestBodyEncodeDecode and does not define headers.")
		}
	}

	if IsObject(e.MethodExpr.Payload.Type) {
		if e.MapQueryParams != nil {
			if pAttr := *e.MapQueryParams; pAttr == "" {
				verr.Add(e, "MapParams is set to map entire payload but payload is an object. Payload must be a map.")
			} else if e.MethodExpr.Payload.Find(pAttr) == nil {
				verr.Add(e, "MapParams is set to an attribute in Payload. But payload has no attribute with type map and name %s", pAttr)
			}
		}
		if e.Body != nil {
			if e.MultipartRequest {
				verr.Add(e, "HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
			}
			if bObj := AsObject(e.Body.Type); bObj != nil {
				var props []string
				props, ok := e.Body.Meta["origin:attribute"]
				if !ok {
					for _, nat := range *bObj {
						name := strings.Split(nat.Name, ":")[0]
						props = append(props, name)
					}
				}
				for _, prop := range props {
					if e.MethodExpr.Payload.Find(prop) == nil {
						verr.Add(e, "Body %q is not found in Payload.", prop)
					}
				}
			}
		}
	}

	body := httpRequestBody(e)
	if e.MultipartRequest && body.Type == Empty {
		verr.Add(e, "MultipartRequest requires a request body.")
	}
	if e.SkipRequestBodyEncodeDecode && body.Type != Empty {
		verr.Add(e, "HTTP endpoint request body must be empty when using SkipRequestBodyEncodeDecode but not all method payload attributes are mapped to headers and params. Make sure to define Headers and Params as needed.")
	}

	// WebSocket upgrade requests cannot carry a request body.
	if e.MethodExpr.IsStreaming() && body.Type != Empty {
		if e.UsesWebSocket() {
			verr.Add(e, "HTTP endpoint request body must be empty when the endpoint uses streaming. Payload attributes must be mapped to headers and/or params.")
		}
	}

	return verr
}

// validateErrorMappings ensures inherited HTTP response policy describes the
// same concrete error value returned by the endpoint method.
func (e *HTTPEndpointExpr) validateErrorMappings() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	for _, mapping := range e.HTTPErrors {
		mapped, owner := mapping.mappedError()
		method := e.MethodExpr.Error(mapping.Name)
		if mapped == nil || method == nil || equivalentErrorAttributes(mapped.AttributeExpr, method.AttributeExpr) {
			continue
		}
		verr.Add(
			mapping.Response,
			`HTTP error mapping %q inherited from the %s uses error type %q, but method %q of service %q uses %q; both definitions must define the same error attribute; %s`,
			mapping.Name,
			owner,
			mapped.Type.Name(),
			e.MethodExpr.Name,
			e.MethodExpr.Service.Name,
			method.Type.Name(),
			errorAttributeDifference(mapped.AttributeExpr, method.AttributeExpr),
		)
	}
	return verr
}

// errorAttributeDifference names qualifier settings when they are the reason
// two reusable error definitions disagree.
func errorAttributeDifference(first, second *AttributeExpr) string {
	if settings := differingErrorQualifierSettings(first, second); len(settings) > 0 {
		return "the " + strings.Join(settings, ", ") + " setting differs"
	}
	return "their type, validations, defaults, or metadata differ"
}

// Finalize is run post DSL execution. It merges response definitions, creates
// implicit endpoint parameters and initializes querystring parameters. It also
// flattens the error responses and makes sure the error types are all user
// types so that the response encoding code can properly use the type to infer
// the response that it needs to build.
func (e *HTTPEndpointExpr) Finalize() {
	// Compute security scheme attribute name and corresponding HTTP location
	requirements := EffectiveSecurityRequirements(e.MethodExpr.Requirements)
	if reqLen := len(requirements); reqLen > 0 {
		e.Requirements = make([]*SecurityExpr, 0, reqLen)
		for _, req := range requirements {
			dupReq := DupRequirement(req)
			for _, sch := range dupReq.Schemes {
				var field string
				switch sch.Kind {
				case NoKind:
					continue
				case BasicAuthKind:
					sch.In = "header"
					sch.Name = "Authorization"
					continue
				case APIKeyKind:
					field = TaggedAttribute(e.MethodExpr.Payload, "security:apikey:"+sch.SchemeName)
				case BearerKind:
					field = TaggedAttribute(e.MethodExpr.Payload, "security:bearer")
				case JWTKind:
					field = TaggedAttribute(e.MethodExpr.Payload, "security:token")
				case OAuth2Kind:
					field = TaggedAttribute(e.MethodExpr.Payload, "security:accesstoken")
				}
				sch.Name, sch.In = findKey(e, field)
				if sch.Name == "" {
					// Initialize Authorization header implicitly defined via
					// security DSL if mapping isn't explicit.
					sch.Name = "Authorization"
					attr := e.MethodExpr.Payload.Find(field)
					e.Headers.Type.(*Object).Set(field, attr)
					e.Headers.Map(sch.Name, field)
					if e.MethodExpr.Payload.IsRequired(field) {
						if e.Headers.Validation == nil {
							e.Headers.Validation = &ValidationExpr{}
						}
						e.Headers.Validation.AddRequired(field)
					}
				}
			}
			e.Requirements = append(e.Requirements, dupReq)
		}
	}

	// Initialize the HTTP specific attributes with the corresponding
	// payload attributes.
	initAttr(e.Params, e.MethodExpr.Payload)
	initAttr(e.Headers, e.MethodExpr.Payload)
	initAttr(e.Cookies, e.MethodExpr.Payload)
	if e.SSE != nil && e.SSE.RequestIDField != "" {
		name := e.SSE.RequestIDField
		attribute := DupAtt(e.MethodExpr.Payload.Find(name))
		if attribute.Meta == nil {
			attribute.Meta = make(MetaExpr)
		}
		attribute.Meta["sse:last-event-id"] = []string{"true"}
		AsObject(e.Headers.Type).Set(name, attribute)
		e.Headers.Map("Last-Event-ID", name)
		if e.MethodExpr.Payload.IsRequired(name) {
			if e.Headers.Validation == nil {
				e.Headers.Validation = &ValidationExpr{}
			}
			e.Headers.Validation.AddRequired(name)
		}
	}

	e.Body = httpRequestBody(e)
	e.Body.Finalize()

	e.StreamingBody = httpStreamingBody(e)
	if e.StreamingBody != nil {
		e.StreamingBody.Finalize()
	}

	// Initialize responses parent, headers and body
	for _, r := range e.Responses {
		r.Finalize(e, e.MethodExpr.Result)
		r.Body = httpResponseBody(e, r)
		r.Body.Finalize()
	}

	// Make sure all error types are user types and have a body.
	for _, herr := range e.HTTPErrors {
		herr.Finalize(e)
	}
}

// validateParams checks the endpoint parameters are of an allowed type and the
// method payload contains the parameters.
func (e *HTTPEndpointExpr) validateParams() *eval.ValidationErrors {
	if e.Params.IsEmpty() {
		return nil
	}

	var (
		pparams = DupMappedAtt(e.PathParams())
		qparams = DupMappedAtt(e.QueryParams())
	)
	// We have to figure out the actual type for the params because the actual
	// type is initialized only during the finalize phase. In the validation
	// phase, all param types are string type by default unless specified
	// explicitly.
	initAttr(pparams, e.MethodExpr.Payload)
	initAttr(qparams, e.MethodExpr.Payload)

	invalidTypeErr := func(verr *eval.ValidationErrors, e *HTTPEndpointExpr, name string) {
		verr.Add(e, "path parameter %s cannot be an object, path parameter types must be primitive, array or map (query string only)", name)
	}
	verr := new(eval.ValidationErrors)
	WalkMappedAttr(pparams, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
		switch {
		case IsObject(a.Type), IsMap(a.Type), IsUnion(a.Type):
			invalidTypeErr(verr, e, name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array path parameter %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("path parameter %s", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	WalkMappedAttr(qparams, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
		switch {
		case IsObject(a.Type), IsUnion(a.Type):
			invalidTypeErr(verr, e, name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array query parameter %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("query parameter %s", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	if e.MethodExpr.Payload != nil {
		switch e.MethodExpr.Payload.Type.(type) {
		case *Object, UserType:
			WalkMappedAttr(pparams, func(name, _ string, _ *AttributeExpr) error { // nolint: errcheck
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Path parameter %q not found in payload.", name)
				}
				return nil
			})
			WalkMappedAttr(qparams, func(name, _ string, _ *AttributeExpr) error { // nolint: errcheck
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Query string parameter %q not found in payload.", name)
				}
				return nil
			})
		case *Array:
			if len(*AsObject(pparams.Type))+len(*AsObject(qparams.Type)) > 1 {
				verr.Add(e, "Payload type is array but HTTP endpoint defines multiple parameters. At most one parameter must be defined and it must be an array.")
			}
		case *Map:
			if len(*AsObject(pparams.Type))+len(*AsObject(qparams.Type)) > 1 {
				verr.Add(e, "Payload type is map but HTTP endpoint defines multiple parameters. At most one query string parameter must be defined and it must be a map.")
			}
		}
	}
	return verr
}

// validateHeadersAndCookies makes sure headers and cookies are of an allowed
// type and the method payload defines the corresponding attributes.
func (e *HTTPEndpointExpr) validateHeadersAndCookies() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	// We have to figure out the actual type because it is initialized during
	// the finalize phase. In the validation phase, all param types are string
	// type by default unless specified explicitly.
	headers := DupMappedAtt(e.Headers)
	cookies := DupMappedAtt(e.Cookies)
	initAttr(headers, e.MethodExpr.Payload)
	initAttr(cookies, e.MethodExpr.Payload)
	WalkMappedAttr(headers, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
		switch {
		case IsObject(a.Type), IsUnion(a.Type):
			verr.Add(e, "header %q must be primitive or array", name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array header %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("header %q", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	WalkMappedAttr(cookies, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
		switch {
		case IsObject(a.Type), IsUnion(a.Type), IsArray(a.Type):
			verr.Add(e, "cookie %q must be primitive", name)
		default:
			ctx := fmt.Sprintf("cookie %q", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	switch e.MethodExpr.Payload.Type.(type) {
	case *Object, UserType:
		hasBasicAuth := TaggedAttribute(e.MethodExpr.Payload, "security:username") != ""
		WalkMappedAttr(headers, func(name, elem string, _ *AttributeExpr) error { // nolint: errcheck
			if e.MethodExpr.Payload.Find(name) == nil {
				verr.Add(e, "header %q not found in payload.", name)
			}
			if elem == "Authorization" && hasBasicAuth {
				// BasicAuth security implicitly sets the Authorization header. If any
				// payload attribute is mapped to Authorization header, raise a
				// validation error.
				verr.Add(e, "Attribute %q is mapped to \"Authorization\" header in the endpoint secured by BasicAuth which also sets \"Authorization\" header. Specify a different header to map attribute %q.", name, name)
			}
			return nil
		})
		WalkMappedAttr(cookies, func(name, _ string, _ *AttributeExpr) error { // nolint: errcheck
			if e.MethodExpr.Payload.Find(name) == nil {
				verr.Add(e, "cookie %q not found in payload.", name)
			}
			return nil
		})
	case *Array:
		if len(*AsObject(headers.Type)) > 1 {
			verr.Add(e, "Payload type is array but HTTP endpoint defines multiple headers. At most one header must be defined and it must be an array.")
		}
	case *Map:
		if len(*AsObject(headers.Type))+len(*AsObject(cookies.Type)) > 0 {
			verr.Add(e, "Payload type is map but HTTP endpoint defines headers or cookies. Map payloads can only be decoded from HTTP request bodies or query strings.")
		}
	}
	return verr
}

// EvalName returns the generic definition name used in error messages.
func (r *RouteExpr) EvalName() string {
	return fmt.Sprintf(`route %s %q of %s`, r.Method, r.Path, r.Endpoint.EvalName())
}

// Validate validates a route expression by ensuring that the route parameters
// can be inferred from the method payload and there is no duplicate parameters
// in an absolute route.
func (r *RouteExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	// Make sure route params are defined in the method payload
	if rparams := r.Params(); len(rparams) > 0 {
		if r.Endpoint.MethodExpr.Payload == nil {
			verr.Add(r, "Route parameters are defined, but method payload is not defined.")
		} else {
			switch r.Endpoint.MethodExpr.Payload.Type.(type) {
			case *Map:
				verr.Add(r, "Route parameters are defined, but method payload is a map. Method payload must be a primitive or an object.")
			case *Object, UserType:
				for _, p := range rparams {
					if r.Endpoint.MethodExpr.Payload.Find(p) == nil {
						verr.Add(r, "Route param %q not found in method payload", p)
					}
				}
			}
			if len(rparams) > 1 && IsPrimitive(r.Endpoint.MethodExpr.Payload.Type) {
				verr.Add(r, "Multiple route parameters are defined, but method payload is a primitive. Only one router parameter can be defined if payload is primitive.")
			}
		}
	}

	// Make sure there's no duplicate params in absolute route
	paths := r.FullPaths()
	for _, path := range paths {
		matches := HTTPWildcardRegex.FindAllStringSubmatch(path, -1)
		wcs := make(map[string]struct{}, len(matches))
		for _, match := range matches {
			if _, ok := wcs[match[1]]; ok {
				verr.Add(r, "Wildcard %q appears multiple times in full path %q", match[1], path)
			}
			wcs[match[1]] = struct{}{}
		}
	}

	// For WebSocket streaming endpoints, only GET is supported
	// SSE endpoints can use both GET and POST (JSON-RPC SSE uses POST)
	if r.Endpoint.UsesWebSocket() && len(r.Endpoint.Responses) > 0 {
		if r.Method != "GET" {
			verr.Add(r, "WebSocket endpoint supports only \"GET\" method. Got %q.", r.Method)
		}
	}

	// HEAD method must not return a response body as per RFC 2616 section 9.4
	if r.Method == "HEAD" {
		disallowBody := func(resp *HTTPResponseExpr) {
			if httpResponseBody(r.Endpoint, resp).Type != Empty {
				verr.Add(r, "HTTP status %d: Response body defined for HEAD method which does not allow response body.", resp.StatusCode)
			}
		}
		for _, resp := range r.Endpoint.Responses {
			disallowBody(resp)
		}
		for _, e := range r.Endpoint.HTTPErrors {
			disallowBody(e.Response)
		}
	}
	return verr
}

// Params returns all the route parameters across all the base paths. For
// example for the route "GET /foo/{fooID:foo_id}" Params returns
// []string{"fooID:foo_id"}.
func (r *RouteExpr) Params() []string {
	paths := r.FullPaths()
	var res []string
	for _, p := range paths {
		ws := ExtractHTTPWildcards(p)
		for _, w := range ws {
			found := slices.Contains(res, w)
			if !found {
				res = append(res, w)
			}
		}
	}
	return res
}

// FullPaths returns the endpoint full paths computed by concatenating the
// service base paths with the route specific path.
func (r *RouteExpr) FullPaths() []string {
	if r.IsAbsolute() {
		return []string{httppath.Clean(r.Path[1:])}
	}
	bases := r.Endpoint.Service.FullPaths()
	res := make([]string, len(bases))
	for i, b := range bases {
		res[i] = httppath.Clean(path.Join(b, r.Path))
		if res[i] == "/" {
			continue
		}
		// path has trailing slash
		if r.Path == "/" && strings.HasSuffix(b, "/") {
			res[i] += "/"
		} else if r.Path != "/" && strings.HasSuffix(r.Path, "/") {
			res[i] += "/"
		}
	}
	return res
}

// IsAbsolute returns true if the endpoint path should not be concatenated to
// the service and API base paths.
func (r *RouteExpr) IsAbsolute() bool {
	return strings.HasPrefix(r.Path, "//")
}

// initAttr initializes the given mapped attribute with the given service
// attribute.
func initAttr(ma *MappedAttributeExpr, svcAtt *AttributeExpr) {
	svcObj := AsObject(svcAtt.Type)
	for _, nat := range *AsObject(ma.Type) {
		var patt *AttributeExpr
		var required bool
		if svcObj != nil {
			patt = svcObj.Attribute(nat.Name)
			required = svcAtt.IsRequired(nat.Name)
		} else {
			patt = svcAtt
			required = true
		}
		initAttrFromDesign(nat.Attribute, patt)
		if required {
			if ma.Validation == nil {
				ma.Validation = &ValidationExpr{}
			}
			ma.Validation.AddRequired(nat.Name)
		}
	}
}

// initAttrFromDesign overrides the type of att with the one of patt and
// initializes other non-initialized fields of att with the one of patt except
// Meta.
func initAttrFromDesign(att, patt *AttributeExpr) {
	if patt == nil || patt.Type == Empty {
		return
	}
	att.authored = patt.AuthoredAttribute()
	att.Type = patt.Type
	if att.Description == "" {
		att.Description = patt.Description
	}
	if att.Docs == nil {
		att.Docs = patt.Docs
	}
	if att.Validation == nil {
		att.Validation = patt.Validation
	}
	if att.DefaultValue == nil {
		att.DefaultValue = patt.DefaultValue
	}
	if att.UserExamples == nil {
		att.UserExamples = patt.UserExamples
	}
	if att.Meta == nil {
		att.Meta = patt.Meta
	}
}

// isEmpty returns true if an attribute is Empty type and it has no bases and
// references, or if an attribute is an empty object.
func isEmpty(a *AttributeExpr) bool {
	if !IsObject(a.Type) {
		return false
	}
	if obj := AsObject(a.Type); obj != nil && len(*obj) != 0 {
		if a.Type == Empty {
			panic("Empty should have no attribute") // bug
		}
		return false
	}
	if len(a.Bases) != 0 || len(a.References) != 0 {
		return false
	}
	if ut, ok := a.Type.(UserType); ok {
		uatt := ut.Attribute()
		if len(uatt.Bases) != 0 || len(uatt.References) != 0 {
			return false
		}
	}
	return true
}

// hasJSONRPCIDField reports whether an attribute graph contains an ID
// declaration.
func hasJSONRPCIDField(attr *AttributeExpr) bool {
	return len(jsonRPCIDFields(attr)) > 0
}

// jsonRPCIDFields lists every ID declaration and its distance from the root
// attribute. A direct payload field has depth one.
func jsonRPCIDFields(attr *AttributeExpr) []jsonRPCIDField {
	var fields []jsonRPCIDField
	collectJSONRPCIDFields(attr, "", 0, make(map[*AttributeExpr]struct{}), &fields)
	return fields
}

// collectJSONRPCIDFields walks one attribute path at a time so a shared type
// can be checked at each place it is used while recursive types still stop.
func collectJSONRPCIDFields(attr *AttributeExpr, name string, depth int, active map[*AttributeExpr]struct{}, fields *[]jsonRPCIDField) {
	if attr == nil || attr.Type == Empty {
		return
	}
	if _, ok := active[attr]; ok {
		return
	}
	active[attr] = struct{}{}
	defer delete(active, attr)

	if _, ok := attr.Meta["jsonrpc:id"]; ok {
		*fields = append(*fields, jsonRPCIDField{name: name, attribute: attr, depth: depth})
	}
	if userType, ok := attr.Type.(UserType); ok {
		collectJSONRPCIDFields(userType.Attribute(), name, depth, active, fields)
		return
	}
	if object := AsObject(attr.Type); object != nil {
		for _, field := range *object {
			collectJSONRPCIDFields(field.Attribute, field.Name, depth+1, active, fields)
		}
		return
	}
	if array := AsArray(attr.Type); array != nil {
		collectJSONRPCIDFields(array.ElemType, name, depth+1, active, fields)
		return
	}
	if mapped := AsMap(attr.Type); mapped != nil {
		collectJSONRPCIDFields(mapped.KeyType, name, depth+1, active, fields)
		collectJSONRPCIDFields(mapped.ElemType, name, depth+1, active, fields)
		return
	}
	if union := AsUnion(attr.Type); union != nil {
		for _, field := range union.Values {
			collectJSONRPCIDFields(field.Attribute, field.Name, depth+1, active, fields)
		}
	}
}

// isJSONRPCIDString reports whether an ID field is a string or a named string.
func isJSONRPCIDString(dataType DataType) bool {
	switch actual := dataType.(type) {
	case Primitive:
		return actual == String
	case UserType:
		return isJSONRPCIDString(actual.Attribute().Type)
	default:
		return false
	}
}

// incompatibleGeneratedJSONRPCRequestIDValidations lists the rules that are
// not guaranteed by the UUID generated when an optional request ID is absent.
func incompatibleGeneratedJSONRPCRequestIDValidations(attribute *AttributeExpr) []string {
	validation := EffectiveValidation(attribute)
	if validation == nil {
		return nil
	}
	var incompatible []string
	if len(validation.Values) > 0 {
		incompatible = append(incompatible, "enum")
	}
	if validation.Format != "" && validation.Format != FormatUUID {
		incompatible = append(incompatible, fmt.Sprintf("format %q", validation.Format))
	}
	if validation.Pattern != "" {
		incompatible = append(incompatible, "pattern")
	}
	if validation.MinLength != nil && *validation.MinLength > generatedJSONRPCRequestIDLength {
		incompatible = append(incompatible, fmt.Sprintf("minimum length %d", *validation.MinLength))
	}
	if validation.MaxLength != nil && *validation.MaxLength < generatedJSONRPCRequestIDLength {
		incompatible = append(incompatible, fmt.Sprintf("maximum length %d", *validation.MaxLength))
	}
	return incompatible
}

// jsonRPCBodyContainsID reports whether an explicitly designed params body
// includes the payload field carried by the JSON-RPC envelope.
func jsonRPCBodyContainsID(body *AttributeExpr, name string) bool {
	if origin, ok := body.Meta["origin:attribute"]; ok {
		return len(origin) > 0 && origin[0] == name
	}
	object := AsObject(body.Type)
	if object == nil {
		return false
	}
	for _, field := range *object {
		if strings.Split(field.Name, ":")[0] == name {
			return true
		}
	}
	return false
}
