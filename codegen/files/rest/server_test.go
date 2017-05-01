package rest

import (
	"bytes"
	"testing"

	"go/format"

	"github.com/sergi/go-diff/diffmatchpatch"
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

		newUserHandlersConstructor = `// NewUserHandlers instantiates HTTP handlers for all the User service
// endpoints.
func NewUserHandlers(
	e *endpoints.User,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) *UserHandlers {
	return &UserHandlers{
		Show: NewShowUserHandler(e.Show, dec, enc, logger),
	}
}
`
		newUserHandlersConstructorMultipleActions = `// NewUserHandlers instantiates HTTP handlers for all the User service
// endpoints.
func NewUserHandlers(
	e *endpoints.User,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) *UserHandlers {
	return &UserHandlers{
		Show: NewShowUserHandler(e.Show, dec, enc, logger),
		List: NewListUserHandler(e.List, dec, enc, logger),
	}
}
`

		mountUserHandlers = `// MountUserHandlers configures the mux to serve the User endpoints.
func MountUserHandlers(mux rest.Muxer, h *UserHandlers) {
	MountShowUserHandler(mux, h.Show)
}
`

		mountUserHandlersMultipleActions = `// MountUserHandlers configures the mux to serve the User endpoints.
func MountUserHandlers(mux rest.Muxer, h *UserHandlers) {
	MountShowUserHandler(mux, h.Show)
	MountListUserHandler(mux, h.List)
}
`

		mountShowUserHandler = `// MountShowUserHandler configures the mux to serve the "User" service "Show"
// endpoint.
func MountShowUserHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/account/foo", f)
}
`

		mountShowUserHandlerPathParam = `// MountShowUserHandler configures the mux to serve the "User" service "Show"
// endpoint.
func MountShowUserHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/account/foo/{id}", f)
}
`

		mountListUserHandler = `// MountListUserHandler configures the mux to serve the "User" service "List"
// endpoint.
func MountListUserHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/account/bar", f)
}
`

		mountShowUserHandlerMultiplePaths = `// MountShowUserHandler configures the mux to serve the "User" service "Show"
// endpoint.
func MountShowUserHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/account/foo", f)
	mux.Handle("GET", "/bar/baz", f)
}
`

		newShowUserHandlerNoResponse = `// NewShowUserHandler creates a HTTP handler which loads the HTTP request and
// calls the "User" service "Show" endpoint.
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
func NewShowUserHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		decodeRequest  = ShowUserDecodeRequest(dec)
		encodeError    = rest.EncodeError(enc, logger)
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
		w.Write(http.StatusNoContent)
	})
}
`

		newShowUserHandlerNoPayloadAndResponse = `// NewShowUserHandler creates a HTTP handler which loads the HTTP request and
// calls the "User" service "Show" endpoint.
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
func NewShowUserHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		encodeError    = rest.EncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := endpoint(r.Context())

		if err != nil {
			encodeError(w, r, err)
			return
		}
		w.Write(http.StatusNoContent)
	})
}
`
		newShowUserHandlerNoPayload = `// NewShowUserHandler creates a HTTP handler which loads the HTTP request and
// calls the "User" service "Show" endpoint.
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
func NewShowUserHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		encodeResponse = ShowUserEncodeResponseEncodeResponse(enc)
		encodeError    = rest.EncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := endpoint(r.Context())

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

		newShowUserHandler = `// NewShowUserHandler creates a HTTP handler which loads the HTTP request and
// calls the "User" service "Show" endpoint.
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
func NewShowUserHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		decodeRequest  = ShowUserDecodeRequest(dec)
		encodeResponse = ShowUserEncodeResponseEncodeResponse(enc)
		encodeError    = rest.EncodeError(enc, logger)
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

		newListUserHandlerNoPayload = `// NewListUserHandler creates a HTTP handler which loads the HTTP request and
// calls the "User" service "List" endpoint.
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
func NewListUserHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		encodeResponse = ListUserEncodeResponseEncodeResponse(enc)
		encodeError    = rest.EncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := endpoint(r.Context())

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

		newShowUserHandlerWithCustomError = `// NewShowUserHandler creates a HTTP handler which loads the HTTP request and
// calls the "User" service "Show" endpoint.
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
func NewShowUserHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		decodeRequest  = ShowUserDecodeRequest(dec)
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
		w.Write(http.StatusNoContent)
	})
}
`

		showUserDecodeNoPayload = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User
// endpoint.
func ShowUserDecodeRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		payload, err := NewShowUserPayload()
		return payload, err
	}
}
`

		showUserDecodePathParams = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User
// endpoint.
func ShowUserDecodeRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		params := httptreemux.ContextParams(r.Context())
		var (
			id int
		)

		idRaw := params["id"]
		v, err := strconv.ParseInt(idRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, fmt.Errorf("id must be an integer, got '%s'", idRaw)
		}
		id = int(v)

		payload, err := NewShowUserPayload(id)
		return payload, err
	}
}
`

		showUserDecodeQueryParams = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User
// endpoint.
func ShowUserDecodeRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*service.ShowUserPayload, error) {
		var (
			id int
		)

		idRaw := r.URL.Query().Get("id")
		v, err := strconv.ParseInt(idRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, fmt.Errorf("id must be an integer, got '%s'", idRaw)
		}
		id = int(v)

		payload, err := NewShowUserPayload(id)
		return payload, err
	}
}
`

		showUserDecodeBodyPayload = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User
// endpoint.
func ShowUserDecodeRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

		showUserDecodeBodyAll = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User
// endpoint.
func ShowUserDecodeRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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
		v, err := strconv.ParseInt(fooRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, fmt.Errorf("foo must be an integer, got '%s'", fooRaw)
		}
		foo = int(v)

		idRaw := params["id"]
		v, err := strconv.ParseInt(idRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, fmt.Errorf("id must be an integer, got '%s'", idRaw)
		}
		id = int(v)

		payload, err := NewShowUserPayload(&body, foo, id)
		return payload, err
	}
}
`
		showUserDecodeAllTypes = `// ShowUserDecodeRequest returns a decoder for requests sent to the create User
// endpoint.
func ShowUserDecodeRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

		var (
			boolVar bool
			float32Var float32
			float64Var float64
			int32Var int32
			int64Var int64
			intVar int
			sliceBoolVar []bool
			sliceFloat32Var []float32
			sliceFloat64Var []float64
			sliceInt32Var []int32
			sliceInt64Var []int64
			sliceIntVar []int
			sliceString []string
			sliceUint32Var []uint32
			sliceUint64Var []uint64
			sliceUintVar []uint
			stringVar string
			uint32Var uint32
			uint64Var uint64
			uintVar uint
		)

		boolVarRaw := r.URL.Query().Get("bool_var")
		v, err := strconv.ParseBool(boolVarRaw)
		if err != nil {
			return nil, fmt.Errorf("bool_var must be a boolean (true or false), got '%s'", boolVarRaw)
		}
		boolVar = v

		float32VarRaw := r.URL.Query().Get("float32_var")
		v, err := strconv.ParseFloat(float32VarRaw, 32)
		if err != nil {
			return nil, fmt.Errorf("float32_var must be a float, got '%s'", float32VarRaw)
		}
		float32Var = float32(v)

		float64VarRaw := r.URL.Query().Get("float64_var")
		v, err := strconv.ParseFloat(float64VarRaw, 64)
		if err != nil {
			return nil, fmt.Errorf("float64_var must be a float, got '%s'", float64VarRaw)
		}
		float64Var = v

		int32VarRaw := r.URL.Query().Get("int32_var")
		v, err := strconv.ParseInt(int32VarRaw, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("int32_var must be an integer, got '%s'", int32VarRaw)
		}
		int32Var = int32(v)

		int64VarRaw := r.URL.Query().Get("int64_var")
		v, err := strconv.ParseInt(int64VarRaw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("int64_var must be an integer, got '%s'", int64VarRaw)
		}
		int64Var = v

		intVarRaw := r.URL.Query().Get("int_var")
		v, err := strconv.ParseInt(intVarRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, fmt.Errorf("int_var must be an integer, got '%s'", intVarRaw)
		}
		intVar = int(v)

		sliceBoolVarRaw := r.URL.Query().Get("slice_bool_var")
		sliceBoolVarRawSlice := strings.Split(sliceBoolVarRaw, ",")
		sliceBoolVar = make([]bool, len(sliceBoolVarRawSlice))
		for i, rv := range sliceBoolVarRawSlice {
			v, err := strconv.ParseBool(rv)
			if err != nil {
				return nil, fmt.Errorf("slice_bool_var must be an set of booleans (true, false, 1 or 0), got value '%s' in set '%s'", rv, sliceBoolVarRaw)
			}
			sliceBoolVar[i] = v
		}

		sliceFloat32VarRaw := r.URL.Query().Get("slice_float32_var")
		sliceFloat32VarRawSlice := strings.Split(sliceFloat32VarRaw, ",")
		sliceFloat32Var = make([]float32, len(sliceFloat32VarRawSlice))
		for i, rv := range sliceFloat32VarRawSlice {
			v, err := strconv.ParseFloat(rv, 32)
			if err != nil {
				return nil, fmt.Errorf("slice_float32_var must be an set of floats, got value '%s' in set '%s'", rv, sliceFloat32VarRaw)
			}
			sliceFloat32Var[i] = float32(v)
		}

		sliceFloat64VarRaw := r.URL.Query().Get("slice_float64_var")
		sliceFloat64VarRawSlice := strings.Split(sliceFloat64VarRaw, ",")
		sliceFloat64Var = make([]float64, len(sliceFloat64VarRawSlice))
		for i, rv := range sliceFloat64VarRawSlice {
			v, err := strconv.ParseFloat(rv, 64)
			if err != nil {
				return nil, fmt.Errorf("slice_float64_var must be an set of floats, got value '%s' in set '%s'", rv, sliceFloat64VarRaw)
			}
			sliceFloat64Var[i] = v
		}

		sliceInt32VarRaw := r.URL.Query().Get("slice_int32_var")
		sliceInt32VarRawSlice := strings.Split(sliceInt32VarRaw, ",")
		sliceInt32Var = make([]int32, len(sliceInt32VarRawSlice))
		for i, rv := range sliceInt32VarRawSlice {
			v, err := strconv.ParseInt(rv, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("slice_int32_var must be an set of integers, got value '%s' in set '%s'", rv, sliceInt32VarRaw)
			}
			sliceInt32Var[i] = int32(v)
		}

		sliceInt64VarRaw := r.URL.Query().Get("slice_int64_var")
		sliceInt64VarRawSlice := strings.Split(sliceInt64VarRaw, ",")
		sliceInt64Var = make([]int64, len(sliceInt64VarRawSlice))
		for i, rv := range sliceInt64VarRawSlice {
			v, err := strconv.ParseInt(rv, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("slice_int64_var must be an set of integers, got value '%s' in set '%s'", rv, sliceInt64VarRaw)
			}
			sliceInt64Var[i] = v
		}

		sliceIntVarRaw := r.URL.Query().Get("slice_int_var")
		sliceIntVarRawSlice := strings.Split(sliceIntVarRaw, ",")
		sliceIntVar = make([]int, len(sliceIntVarRawSlice))
		for i, rv := range sliceIntVarRawSlice {
			v, err := strconv.ParseInt(rv, 10, strconv.IntSize)
			if err != nil {
				return nil, fmt.Errorf("slice_int_var must be an set of integers, got value '%s' in set '%s'", rv, sliceIntVarRaw)
			}
			sliceIntVar[i] = int(v)
		}

		sliceStringRaw := r.URL.Query().Get("slice_string")
		sliceStringRawSlice := strings.Split(sliceStringRaw, ",")
		sliceString = make([]string, len(sliceStringRawSlice))
		for i, rv := range sliceStringRawSlice {
			sliceString[i] = url.QueryUnescape(rv)
		}

		sliceUint32VarRaw := r.URL.Query().Get("slice_uint32_var")
		sliceUint32VarRawSlice := strings.Split(sliceUint32VarRaw, ",")
		sliceUint32Var = make([]uint32, len(sliceUint32VarRawSlice))
		for i, rv := range sliceUint32VarRawSlice {
			v, err := strconv.ParseUint(rv, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("slice_uint32_var must be an set of unsigned integers, got value '%s' in set '%s'", rv, sliceUint32VarRaw)
			}
			sliceUint32Var[i] = int32(v)
		}

		sliceUint64VarRaw := r.URL.Query().Get("slice_uint64_var")
		sliceUint64VarRawSlice := strings.Split(sliceUint64VarRaw, ",")
		sliceUint64Var = make([]uint64, len(sliceUint64VarRawSlice))
		for i, rv := range sliceUint64VarRawSlice {
			v, err := strconv.ParseUint(rv, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("slice_uint64_var must be an set of unsigned integers, got value '%s' in set '%s'", rv, sliceUint64VarRaw)
			}
			sliceUint64Var[i] = v
		}

		sliceUintVarRaw := r.URL.Query().Get("slice_uint_var")
		sliceUintVarRawSlice := strings.Split(sliceUintVarRaw, ",")
		sliceUintVar = make([]uint, len(sliceUintVarRawSlice))
		for i, rv := range sliceUintVarRawSlice {
			v, err := strconv.ParseUint(rv, 10, strconv.IntSize)
			if err != nil {
				return nil, fmt.Errorf("slice_uint_var must be an set of unsigned integers, got value '%s' in set '%s'", rv, sliceUintVarRaw)
			}
			sliceUintVar[i] = uint(v)
		}

		stringVar = r.URL.Query().Get("mappedvar")

		uint32VarRaw := r.URL.Query().Get("uint32_var")
		v, err := strconv.ParseUint(uint32VarRaw, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("uint32_var must be an unsigned integer, got '%s'", uint32VarRaw)
		}
		uint32Var = int32(v)

		uint64VarRaw := r.URL.Query().Get("uint64_var")
		v, err := strconv.ParseUint(uint64VarRaw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("uint64_var must be an unsigned integer, got '%s'", uint64VarRaw)
		}
		uint64Var = v

		uintVarRaw := r.URL.Query().Get("uint_var")
		v, err := strconv.ParseUint(uintVarRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, fmt.Errorf("uint_var must be an unsigned integer, got '%s'", uintVarRaw)
		}
		uintVar = uint(v)

		payload, err := NewShowUserPayload(&body, boolVar, float32Var, float64Var, int32Var, int64Var, intVar, sliceBoolVar, sliceFloat32Var, sliceFloat64Var, sliceInt32Var, sliceInt64Var, sliceIntVar, sliceString, sliceUint32Var, sliceUint64Var, sliceUintVar, stringVar, uint32Var, uint64Var, uintVar)
		return payload, err
	}
}
`

		showUserEncodeResponseNoResponse = `// ShowUserEncodeResponse returns an encoder for responses returned by the Show
// User endpoint.
func ShowUserEncodeResponse(encoder func(http.ResponseWriter, *http.Request) rest.Encoder) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
`
		listUserEncodeResponseNoResponse = `// ListUserEncodeResponse returns an encoder for responses returned by the List
// User endpoint.
func ListUserEncodeResponse(encoder func(http.ResponseWriter, *http.Request) rest.Encoder) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
`

		showUserEncodeResponse = `// ShowUserEncodeResponse returns an encoder for responses returned by the Show
// User endpoint.
func ShowUserEncodeResponse(encoder func(http.ResponseWriter, *http.Request) rest.Encoder) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		t := v.(*UserOK)
		w.Header().Set("Content-Type", ResponseContentType(r))
		w.Header().Set("Location", t.Href)
		w.Header().Set("Request", t.Request)
		w.WriteHeader(http.StatusOK)
		if t != nil {
			return encoder(w, r).Encode(t)
		}
		return nil
	}
}
`
		showUserEncodeMultipleResponses = `// ShowUserEncodeResponse returns an encoder for responses returned by the Show
// User endpoint.
func ShowUserEncodeResponse(encoder func(http.ResponseWriter, *http.Request) rest.Encoder) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		switch t := v.(type) {
		case *UserCreated:
			w.Header().Set("Location", t.Href)
			w.Header().Set("Request", t.Request)
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

		showUserEncodeError = `// ShowUserEncodeError returns an encoder for errors returned by the Show User
// endpoint.
func ShowUserEncodeError(encoder func(http.ResponseWriter, *http.Request) rest.Encoder, logger goa.LogAdapter) func(http.ResponseWriter, *http.Request, error) {
	encodeError := rest.EncodeError(encoder, logger)
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
		accountAttr = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{},
				TypeName:      "Account",
			}}

		arrayAccountAttr = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{
					Type: &design.Array{
						ElemType: &accountAttr,
					},
				}},
		}

		payload = design.AttributeExpr{
			Type: &design.UserTypeExpr{
				AttributeExpr: &design.AttributeExpr{},
				TypeName:      "FooUserPayload",
			}}

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
			Result:  &accountAttr,
			Errors:  []*design.ErrorExpr{&errorNameAlreadyTaken},
			Service: &service,
		}

		endpointWithPayload = design.EndpointExpr{
			Name:    "Show",
			Payload: &payload,
			Result:  &arrayAccountAttr,
			Service: &service,
		}

		endpointPlain = design.EndpointExpr{
			Name:    "Show",
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result:  &design.AttributeExpr{Type: design.Empty},
			Service: &service,
		}

		endpointPlainOther = design.EndpointExpr{
			Name:    "List",
			Payload: &design.AttributeExpr{Type: design.Empty},
			Result:  &design.AttributeExpr{Type: design.Empty},
			Service: &service,
		}

		actionWithNoPayloadAndResponse = rest.ActionExpr{
			EndpointExpr: &endpointPlain,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
		}

		// actionWithEmptyResponse is the testcase when  the user defined at response with no content as status code
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

		// actionWithNilResponse is the testcase when no response is defined at all in the design
		actionWithNilResponse = rest.ActionExpr{
			EndpointExpr: &endpointPlain,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
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
					Body:       &accountAttr,
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
					Body:       &accountAttr,
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
			Body:         endpointWithPayload.Payload,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
		}

		actionWithPayloadPathParams = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Routes:       []*rest.RouteExpr{{Path: "/foo/{id}", Method: "GET"}},
		}

		actionWithPayloadQueryParams = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Routes:       []*rest.RouteExpr{{Path: "/foo", Method: "GET"}},
		}

		actionWithPayloadBodyAndParams = rest.ActionExpr{
			EndpointExpr: &endpointWithPayload,
			Body:         endpointWithPayload.Payload,
			Routes:       []*rest.RouteExpr{{Path: "/foo/{id}", Method: "GET"}},
		}

		setParams = func(r *rest.ResourceExpr, obj *design.Object) {
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

	h := actionWithResponse.Responses[0].Headers()
	h.Type.(design.Object)["Href:Location"] = &design.AttributeExpr{Type: design.String}
	h.Type.(design.Object)["Request"] = &design.AttributeExpr{Type: design.String}
	h.Validation = &design.ValidationExpr{Required: []string{"Href", "Request"}}

	h = actionWithMultipleResponses.Responses[0].Headers()
	h.Type.(design.Object)["Href:Location"] = &design.AttributeExpr{Type: design.String}
	h.Type.(design.Object)["Request"] = &design.AttributeExpr{Type: design.String}
	h.Validation = &design.ValidationExpr{Required: []string{"Href", "Request"}}

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
				newShowUserHandlerNoPayloadAndResponse,
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
				newShowUserHandlerNoPayloadAndResponse,
			},
		},
		"no-payload-and-response": {
			Resource: resource(&actionWithNoPayloadAndResponse),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayloadAndResponse,
			},
		},
		"with-empty-no-content-defined-response": {
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
		"with-empty-nil-response": {
			Resource: resource(&actionWithNilResponse),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoPayloadAndResponse,
			},
		},
		"with-payload-in-body": {
			Resource: resource(&actionWithPayloadBody),
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoResponse,
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
				newShowUserHandlerNoResponse,
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
				newShowUserHandlerNoResponse,
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
				newShowUserHandlerNoResponse,
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
				newShowUserHandler,
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
				showUserDecodeNoPayload,
				showUserEncodeError,
			},
		},

		"with-all-payload-types": {
			Resource: resource(&actionWithPayloadBody),
			Params: design.Object{
				"string_var:mappedvar": {Type: design.String},
				"int_var":              {Type: design.Int},
				"int32_var":            {Type: design.Int32},
				"int64_var":            {Type: design.Int64},
				"uint_var":             {Type: design.UInt},
				"uint32_var":           {Type: design.UInt32},
				"uint64_var":           {Type: design.UInt64},
				"float32_var":          {Type: design.Float32},
				"float64_var":          {Type: design.Float64},
				"bool_var":             {Type: design.Boolean},
				"slice_string":         {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.String}}},
				"slice_int_var":        {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int}}},
				"slice_int32_var":      {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int32}}},
				"slice_int64_var":      {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int64}}},
				"slice_uint_var":       {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt}}},
				"slice_uint32_var":     {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt32}}},
				"slice_uint64_var":     {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt64}}},
				"slice_float32_var":    {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Float32}}},
				"slice_float64_var":    {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Float64}}},
				"slice_bool_var":       {Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Boolean}}},
			},
			Expected: []string{
				userHandlers,
				newUserHandlersConstructor,
				mountUserHandlers,
				mountShowUserHandler,
				newShowUserHandlerNoResponse,
				showUserDecodeAllTypes,
			},
		},
	}

	for k, tc := range cases {
		if tc.Params != nil {
			setParams(tc.Resource, &tc.Params)
		} else {
			setParams(tc.Resource, &design.Object{})
		}

		ss := Server(tc.Resource).Sections("")
		if len(ss)-1 != len(tc.Expected) {
			t.Errorf("%s: got %d sections but expected %d", k, len(ss)-1, len(tc.Expected))
			continue
		}

		for i, s := range ss[1:] {
			buf := new(bytes.Buffer)

			e := s.Write(buf)
			if e != nil {
				t.Fatalf("%s: failed to execute template, error '%s' for section @index %d", k, e, i)
			}

			actual, err := format.Source(buf.Bytes())
			if err != nil {
				t.Fatal(err)
			}

			expected, err := format.Source([]byte(tc.Expected[i]))
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(string(actual), string(expected), false)
				diff := dmp.DiffPrettyText(diffs)
				t.Errorf("%s: diff:\n%s\nfor section @index %d", k, diff, i)
			}
		}
	}
}
