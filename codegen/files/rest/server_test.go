package rest

import (
	"bytes"
	"testing"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

func TestServer(t *testing.T) {

	const (
		userHandlers = `// UserHandlers lists the User service endpoint HTTP handlers.
type UserHandlers struct {
	Show http.Handler
}
`
		userHandlersMultipleActions = `// UserHandlers lists the User service endpoint HTTP handlers.
type UserHandlers struct {
	Show http.Handler
	List http.Handler
}
`

		newUserHandlersConstructor = `// NewUserHandlers instantiates HTTP handlers for all the User service endpoints.
func NewUserHandlers(
	e *endpoints.User,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) *UserHandlers {
	return &UserHandlers{
		Show: NewShowUserHandler(e.Show, dec, enc, logger),
	}
}
`
		newUserHandlersConstructorMultipleActions = `// NewUserHandlers instantiates HTTP handlers for all the User service endpoints.
func NewUserHandlers(
	e *endpoints.User,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) *UserHandlers {
	return &UserHandlers{
		Show: NewShowUserHandler(e.Show, dec, enc, logger),
		List: NewListUserHandler(e.List, dec, enc, logger),
	}
}
`

		mountUserHandlers = `// MountUserHandlers configures the mux to serve the User endpoints.
func MountUserHandlers(mux rest.ServeMux, h *UserHandlers) {
	MountShowUserHandler(mux, h.Show)
}
`

		mountUserHandlersMultipleActions = `// MountUserHandlers configures the mux to serve the User endpoints.
func MountUserHandlers(mux rest.ServeMux, h *UserHandlers) {
	MountShowUserHandler(mux, h.Show)
	MountListUserHandler(mux, h.List)
}
`

		mountShowUserHandler = `// MountShowUserHandler configures the mux to serve the "User" service "Show" endpoint.
func MountShowUserHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("GET", "/account/foo", h)
}
`

		mountShowUserHandlerPathParam = `// MountShowUserHandler configures the mux to serve the "User" service "Show" endpoint.
func MountShowUserHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("GET", "/account/foo/:id", h)
}
`

		mountListUserHandler = `// MountListUserHandler configures the mux to serve the "User" service "List" endpoint.
func MountListUserHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("POST", "/account/bar", h)
}
`

		mountShowUserHandlerMultiplePaths = `// MountShowUserHandler configures the mux to serve the "User" service "Show" endpoint.
func MountShowUserHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("GET", "/account/foo", h)
	mux.Handle("GET", "/bar/baz", h)
}
`

		newShowUserHandlerNoPayload = `// NewShowUserHandler creates a HTTP handler which loads the HTTP request and calls the "User" service "Show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus may access the request state via the rest package ContextXXX functions.
func NewShowUserHandler(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		decodeRequest  = ShowUserDecodeRequest(dec)
		encodeResponse = ShowUserEncodeResponseEncodeResponse(enc)
		encodeError    = EncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, goa.ErrInvalid("request invalid: %s", err))
			return
		}

		res, err := endpoint(r.Context(), payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}
`
		newListUserHandlerNoPayload = `// NewListUserHandler creates a HTTP handler which loads the HTTP request and calls the "User" service "List" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus may access the request state via the rest package ContextXXX functions.
func NewListUserHandler(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		decodeRequest  = ListUserDecodeRequest(dec)
		encodeResponse = ListUserEncodeResponseEncodeResponse(enc)
		encodeError    = EncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, goa.ErrInvalid("request invalid: %s", err))
			return
		}

		res, err := endpoint(r.Context(), payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}
`

		newShowUserHandlerWithCustomError = `// NewShowUserHandler creates a HTTP handler which loads the HTTP request and calls the "User" service "Show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus may access the request state via the rest package ContextXXX functions.
func NewShowUserHandler(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		decodeRequest  = ShowUserDecodeRequest(dec)
		encodeResponse = ShowUserEncodeResponseEncodeResponse(enc)
		encodeError    = ShowUserEncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, goa.ErrInvalid("request invalid: %s", err))
			return
		}

		res, err := endpoint(r.Context(), payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}
`

		showUserDecodeNoPayload = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User endpoint.
func ShowUserDecodeRequest(decoder rest.RequestDecoderFunc) DecodeRequestFunc {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		payload, err := NewShowUserPayload()
		return payload, err
	}
}
`

		showUserDecodePathParams = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User endpoint.
func ShowUserDecodeRequest(decoder rest.RequestDecoderFunc) DecodeRequestFunc {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		params := httptreemux.ContextParams(r.Context())
		var (
			id int
		)

		idRaw := params["id"]
		if v, err := strconv.ParseInt(idRaw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("id must be an integer, got '%s'", idRaw)
		} else {
			id = int(v)
		}

		payload, err := NewShowUserPayload(id)
		return payload, err
	}
}
`

		showUserDecodeQueryParams = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User endpoint.
func ShowUserDecodeRequest(decoder rest.RequestDecoderFunc) DecodeRequestFunc {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		var (
			id int
		)

		idRaw := r.URL.Query().Get("id")
		if v, err := strconv.ParseInt(idRaw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("id must be an integer, got '%s'", idRaw)
		} else {
			id = int(v)
		}

		payload, err := NewShowUserPayload(id)
		return payload, err
	}
}
`

		showUserDecodeBodyPayload = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User endpoint.
func ShowUserDecodeRequest(decoder rest.RequestDecoderFunc) DecodeRequestFunc {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		var (
			body ShowUserBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = fmt.Errorf("empty body")
			}
			return nil, err
		}

		payload, err := NewShowUserPayload(&body)
		return payload, err
	}
}
`

		showUserDecodeBodyAll = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User endpoint.
func ShowUserDecodeRequest(decoder rest.RequestDecoderFunc) DecodeRequestFunc {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		var (
			body ShowUserBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = fmt.Errorf("empty body")
			}
			return nil, err
		}

		params := httptreemux.ContextParams(r.Context())
		var (
			foo int
			id int
		)

		fooRaw := r.URL.Query().Get("foo")
		if v, err := strconv.ParseInt(fooRaw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("foo must be an integer, got '%s'", fooRaw)
		} else {
			foo = int(v)
		}

		idRaw := params["id"]
		if v, err := strconv.ParseInt(idRaw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("id must be an integer, got '%s'", idRaw)
		} else {
			id = int(v)
		}

		payload, err := NewShowUserPayload(&body, foo, id)
		return payload, err
	}
}
`

		showUserEncodeResponseNoResponse = `// ShowUserEncodeResponse returns an encoder for responses returned by the Show User endpoint.
func ShowUserEncodeResponse(encoder rest.ResponseEncoderFunc) EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
`
		listUserEncodeResponseNoResponse = `// ListUserEncodeResponse returns an encoder for responses returned by the List User endpoint.
func ListUserEncodeResponse(encoder rest.ResponseEncoderFunc) EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
`

		showUserEncodeResponse = `// ShowUserEncodeResponse returns an encoder for responses returned by the Show User endpoint.
func ShowUserEncodeResponse(encoder rest.ResponseEncoderFunc) EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r))
		w.WriteHeader(http.StatusOK)
		if v != nil {
			return encoder(w, r).Encode(v)
		}
		return nil
	}
}
`
		showUserEncodeMultipleResponses = `// ShowUserEncodeResponse returns an encoder for responses returned by the Show User endpoint.
func ShowUserEncodeResponse(encoder rest.ResponseEncoderFunc) EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		switch t := v.(type) {
		case *UserCreated:
			w.WriteHeader(http.StatusCreated)
			return encoder(w, r).Encode(t)
		case *UserAccepted:
			w.WriteHeader(http.StatusAccepted)
		default:
			return fmt.Errorf("invalid response type")
		}
		return nil
	}
}
`

		showUserEncodeError = `// ShowUserEncodeError returns an encoder for errors returned by the Show User endpoint.
func ShowUserEncodeError(encoder rest.ResponseEncoderFunc, logger goa.Logger) EncodeErrorFunc {
	encodeError := EncodeError(encoder, logger)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		w.Header().Set("Content-Type", ResponseContentType(r))
		switch t := v.(type) {
		case *service.NameAlreadyTaken:
			w.WriteHeader(http.StatusConflict)
			encoder(w, r).Encode(t)
		default:
			encodeError(w, r, v)
		}
	}
}
`
	)

	var (
		accountType = design.UserTypeExpr{
			AttributeExpr: &design.AttributeExpr{},
			TypeName:      "Account",
		}

		arrayAccountType = design.UserTypeExpr{
			AttributeExpr: &design.AttributeExpr{
				Type: &design.Array{
					ElemType: &design.AttributeExpr{Type: &accountType},
				},
			},
		}

		payload = design.UserTypeExpr{
			AttributeExpr: &design.AttributeExpr{},
			TypeName:      "FooUserPayload",
		}

		errorNameAlreadyTaken = design.ErrorExpr{
			AttributeExpr: &design.AttributeExpr{},
			Name:          "name_already_taken",
		}

		service = design.ServiceExpr{
			Name: "User",
		}

		endpointWithErrorAndPayload = design.EndpointExpr{
			Name:    "Show",
			Payload: &payload,
			Result:  &accountType,
			Errors:  []*design.ErrorExpr{&errorNameAlreadyTaken},
			Service: &service,
		}

		endpointWithPayload = design.EndpointExpr{
			Name:    "Show",
			Payload: &payload,
			Result:  &arrayAccountType,
			Service: &service,
		}

		endpointPlain = design.EndpointExpr{
			Name:    "Show",
			Payload: design.Empty,
			Result:  design.Empty,
			Service: &service,
		}

		endpointPlainOther = design.EndpointExpr{
			Name:    "List",
			Payload: design.Empty,
			Result:  design.Empty,
			Service: &service,
		}

		actionWithNoPayloadAndResponse = rest.ActionExpr{
			EndpointExpr: &endpointPlain,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
		}

		actionWithEmptyResponse = rest.ActionExpr{
			EndpointExpr: &endpointPlain,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
			Responses: []*rest.HTTPResponseExpr{
				{
					StatusCode: rest.StatusNoContent,
					Body:       &design.AttributeExpr{Type: design.Empty},
				},
			},
		}

		actionWithMultiplePaths = rest.ActionExpr{
			EndpointExpr: &endpointPlain,
			Routes: []*rest.RouteExpr{
				{Path: "/foo", Method: "GET"},
				{Path: "//bar/baz", Method: "GET"},
			},
		}

		actionWithCustomErrorResponses = rest.ActionExpr{
			EndpointExpr: &endpointWithErrorAndPayload,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
			HTTPErrors: []*rest.HTTPErrorExpr{
				{
					ErrorExpr: &errorNameAlreadyTaken,
					Name:      "name_already_taken",
					Response: &rest.HTTPResponseExpr{
						StatusCode: rest.StatusConflict,
					},
				},
			},
		}

		actionWithMultipleResponses = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
			Responses: []*rest.HTTPResponseExpr{
				{
					StatusCode: rest.StatusCreated,
					Body:       &design.AttributeExpr{Type: &accountType},
				}, {
					StatusCode: rest.StatusAccepted,
					Body:       &design.AttributeExpr{Type: design.Empty},
				},
			},
		}

		actionWithResponse = rest.ActionExpr{
			EndpointExpr: &endpointPlain,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
			Responses: []*rest.HTTPResponseExpr{
				{
					StatusCode: rest.StatusOK,
					Body:       &design.AttributeExpr{Type: &accountType},
				},
			},
		}

		actionEmptyResponseOther = rest.ActionExpr{
			EndpointExpr: &endpointPlainOther,
			Routes:       []*rest.RouteExpr{{Path: "/bar", Method: "POST"}},
			Responses: []*rest.HTTPResponseExpr{
				{
					StatusCode: rest.StatusNoContent,
					Body:       &design.AttributeExpr{Type: design.Empty},
				},
			},
		}

		actionWithPayloadBody = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Body:         endpointWithPayload.Payload.Attribute(),
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
		}

		actionWithPayloadPathParams = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Routes:       []*rest.RouteExpr{{Path: "/foo/:id", Method: "GET"}},
		}

		actionWithPayloadQueryParams = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
		}

		actionWithPayloadBodyAndParams = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Body:         endpointWithPayload.Payload.Attribute(),
			Routes:       []*rest.RouteExpr{{Path: "/foo/:id", Method: "GET"}},
		}

		setParams = func(r *rest.ResourceExpr, obj *design.Object) {
			//list := map[string]*design.AttributeExpr{
			//	,
			//	"view":            {Type: design.String},
			//	"slice_string":    {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.String}}},
			//	"slice_int32":     {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int32}}},
			//	"slice_int64":     {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int64}}},
			//	"slice_uint32":    {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt32}}},
			//	"slice_uint64":    {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt64}}},
			//	"slice_float32":   {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Float32}}},
			//	"slice_float64":   {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Float64}}},
			//	"slice_bool":      {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Boolean}}},
			//	"slice_interface": {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Any}}},
			//}

			for _, a := range r.Actions {
				a.Params().Type = *obj
			}
		}

		resource = func(actions ...*rest.ActionExpr) *rest.ResourceExpr {
			r := &rest.ResourceExpr{
				ServiceExpr: &service,
				Path:        "/account",
				Actions:     actions,
			}

			for _, a := range actions {
				a.Resource = r
				for _, r := range a.Routes {
					r.Action = a
				}
			}

			return r
		}
	)

	// testcases
	cases := map[string]struct {
		Resource *rest.ResourceExpr
		Params   design.Object
		Expected []string
	}{
		"multiple-actions": {
			Resource: resource(&actionWithNoPayloadAndResponse, &actionEmptyResponseOther),
			Expected: []string{
				userHandlersMultipleActions,
				newUserHandlersConstructorMultipleActions,
				mountUserHandlersMultipleActions,
				mountShowUserHandler,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
				mountListUserHandler,
				newListUserHandlerNoPayload,
				listUserEncodeResponseNoResponse,
			},
		},
		"multiple-paths": {
			Resource: resource(&actionWithMultiplePaths),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandlerMultiplePaths,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
			},
		},
		"no-payload-and-response": {
			Resource: resource(&actionWithNoPayloadAndResponse),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
			},
		},
		"with-empty-response": {
			Resource: resource(&actionWithEmptyResponse),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
			},
		},
		"with-payload-in-body": {
			Resource: resource(&actionWithPayloadBody),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
				showUserDecodeBodyPayload,
			},
		},
		"with-payload-query-params": {
			Resource: resource(&actionWithPayloadQueryParams),
			Params: design.Object{
				"id": {Type: design.Int},
			},
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
				showUserDecodeQueryParams,
			},
		},
		"with-payload-path-params": {
			Resource: resource(&actionWithPayloadPathParams),
			Params: design.Object{
				"id": {Type: design.Int},
			},
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandlerPathParam,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
				showUserDecodePathParams,
			},
		},
		"with-payload-in-body-and-params": {
			Resource: resource(&actionWithPayloadBodyAndParams),
			Params: design.Object{
				"id":  {Type: design.Int},
				"foo": {Type: design.Int},
			},
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandlerPathParam,
				newShowUserHandlerNoPayload,
				showUserEncodeResponseNoResponse,
				showUserDecodeBodyAll,
			},
		},
		"with-response": {
			Resource: resource(&actionWithResponse),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayload,
				showUserEncodeResponse,
			},
		},
		"with-multiple-responses": {
			Resource: resource(&actionWithMultipleResponses),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayload,
				showUserEncodeMultipleResponses,
				showUserDecodeNoPayload,
			},
		},
		"with-custom-errors": {
			Resource: resource(&actionWithCustomErrorResponses),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerWithCustomError,
				showUserEncodeResponseNoResponse,
				showUserDecodeNoPayload,
				showUserEncodeError,
			},
		},
	}

	for k, tc := range cases {

		//for testing only the test i want
		//if k != "with-payload-in-body-and-params" && k != "with-payload-path-params" && k != "with-payload-query-params" && k != "with-payload-in-body" {
		//	continue
		//}

		if tc.Params != nil {
			setParams(tc.Resource, &tc.Params)
		}

		buf := new(bytes.Buffer)
		ss := Server(tc.Resource)

		if len(ss) != len(tc.Expected) {
			t.Errorf("%s: got %d sections but expected %d", k, len(ss), len(tc.Expected))
			continue
		}

		for i, s := range ss {
			buf.Reset()

			e := s.Write(buf)
			if e != nil {
				t.Errorf("%s: failed to execute template, error '%s' for section @index %d", k, e, i)
				continue
			}
			actual := buf.String()

			if actual != tc.Expected[i] {
				t.Errorf("%s: got `%s`, expected `%s` for section @index %d", k, actual, tc.Expected[i], i)
			}
		}
	}
}
