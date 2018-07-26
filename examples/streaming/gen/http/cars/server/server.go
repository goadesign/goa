// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// cars HTTP server
//
// Command:
// $ goa gen goa.design/goa/examples/streaming/design -o
// $(GOPATH)/src/goa.design/goa/examples/streaming

package server

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	goa "goa.design/goa"
	carssvc "goa.design/goa/examples/streaming/gen/cars"
	goahttp "goa.design/goa/http"
)

// Server lists the cars service endpoint HTTP handlers.
type Server struct {
	Mounts []*MountPoint
	Login  http.Handler
	List   http.Handler
	Add    http.Handler
	Update http.Handler
}

// ErrorNamer is an interface implemented by generated error structs that
// exposes the name of the error as defined in the design.
type ErrorNamer interface {
	ErrorName() string
}

// MountPoint holds information about the mounted endpoints.
type MountPoint struct {
	// Method is the name of the service method served by the mounted HTTP handler.
	Method string
	// Verb is the HTTP method used to match requests to the mounted handler.
	Verb string
	// Pattern is the HTTP request path pattern used to match requests to the
	// mounted handler.
	Pattern string
}

// listServerStream implements the carssvc.ListServerStream interface.
type listServerStream struct {
	once sync.Once
	// upgrader is the websocket connection upgrader.
	upgrader goahttp.Upgrader
	// connConfigFn is the websocket connection configurer.
	connConfigFn goahttp.ConnConfigureFunc
	// w is the HTTP response writer used in upgrading the connection.
	w http.ResponseWriter
	// r is the HTTP request.
	r *http.Request
	// conn is the underlying websocket connection.
	conn *websocket.Conn
	// view is the view to render carssvc.StoredCar result type before sending to
	// the websocket connection.
	view string
}

// addServerStream implements the carssvc.AddServerStream interface.
type addServerStream struct {
	once sync.Once
	// upgrader is the websocket connection upgrader.
	upgrader goahttp.Upgrader
	// connConfigFn is the websocket connection configurer.
	connConfigFn goahttp.ConnConfigureFunc
	// w is the HTTP response writer used in upgrading the connection.
	w http.ResponseWriter
	// r is the HTTP request.
	r *http.Request
	// conn is the underlying websocket connection.
	conn *websocket.Conn
	// view is the view to render carssvc.StoredCarCollection result type before
	// sending to the websocket connection.
	view string
}

// updateServerStream implements the carssvc.UpdateServerStream interface.
type updateServerStream struct {
	once sync.Once
	// upgrader is the websocket connection upgrader.
	upgrader goahttp.Upgrader
	// connConfigFn is the websocket connection configurer.
	connConfigFn goahttp.ConnConfigureFunc
	// w is the HTTP response writer used in upgrading the connection.
	w http.ResponseWriter
	// r is the HTTP request.
	r *http.Request
	// conn is the underlying websocket connection.
	conn *websocket.Conn
	// view is the view to render carssvc.StoredCarCollection result type before
	// sending to the websocket connection.
	view string
}

// New instantiates HTTP handlers for all the cars service endpoints.
func New(
	e *carssvc.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"Login", "POST", "/cars/login"},
			{"List", "GET", "/cars"},
			{"Add", "GET", "/cars/add"},
			{"Update", "GET", "/cars/update"},
		},
		Login:  NewLoginHandler(e.Login, mux, dec, enc, eh),
		List:   NewListHandler(e.List, mux, dec, enc, eh, up, connConfigFn),
		Add:    NewAddHandler(e.Add, mux, dec, enc, eh, up, connConfigFn),
		Update: NewUpdateHandler(e.Update, mux, dec, enc, eh, up, connConfigFn),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "cars" }

// Use wraps the server handlers with the given middleware.
func (s *Server) Use(m func(http.Handler) http.Handler) {
	s.Login = m(s.Login)
	s.List = m(s.List)
	s.Add = m(s.Add)
	s.Update = m(s.Update)
}

// Mount configures the mux to serve the cars endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountLoginHandler(mux, h.Login)
	MountListHandler(mux, h.List)
	MountAddHandler(mux, h.Add)
	MountUpdateHandler(mux, h.Update)
}

// MountLoginHandler configures the mux to serve the "cars" service "login"
// endpoint.
func MountLoginHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/cars/login", f)
}

// NewLoginHandler creates a HTTP handler which loads the HTTP request and
// calls the "cars" service "login" endpoint.
func NewLoginHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) http.Handler {
	var (
		decodeRequest  = DecodeLoginRequest(mux, dec)
		encodeResponse = EncodeLoginResponse(enc)
		encodeError    = EncodeLoginError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "login")
		ctx = context.WithValue(ctx, goa.ServiceKey, "cars")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		res, err := endpoint(ctx, payload)

		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	})
}

// MountListHandler configures the mux to serve the "cars" service "list"
// endpoint.
func MountListHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/cars", f)
}

// NewListHandler creates a HTTP handler which loads the HTTP request and calls
// the "cars" service "list" endpoint.
func NewListHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeListRequest(mux, dec)
		encodeError   = EncodeListError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "list")
		ctx = context.WithValue(ctx, goa.ServiceKey, "cars")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		v := &carssvc.ListEndpointInput{
			Stream: &listServerStream{
				upgrader:     up,
				connConfigFn: connConfigFn,
				w:            w,
				r:            r,
			},
			Payload: payload.(*carssvc.ListPayload),
		}
		_, err = endpoint(ctx, v)

		if err != nil {
			if _, ok := err.(websocket.HandshakeError); ok {
				return
			}
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
	})
}

// MountAddHandler configures the mux to serve the "cars" service "add"
// endpoint.
func MountAddHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/cars/add", f)
}

// NewAddHandler creates a HTTP handler which loads the HTTP request and calls
// the "cars" service "add" endpoint.
func NewAddHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeAddRequest(mux, dec)
		encodeError   = EncodeAddError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "add")
		ctx = context.WithValue(ctx, goa.ServiceKey, "cars")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		v := &carssvc.AddEndpointInput{
			Stream: &addServerStream{
				upgrader:     up,
				connConfigFn: connConfigFn,
				w:            w,
				r:            r,
			},
			Payload: payload.(*carssvc.AddPayload),
		}
		_, err = endpoint(ctx, v)

		if err != nil {
			if _, ok := err.(websocket.HandshakeError); ok {
				return
			}
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
	})
}

// MountUpdateHandler configures the mux to serve the "cars" service "update"
// endpoint.
func MountUpdateHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/cars/update", f)
}

// NewUpdateHandler creates a HTTP handler which loads the HTTP request and
// calls the "cars" service "update" endpoint.
func NewUpdateHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeUpdateRequest(mux, dec)
		encodeError   = EncodeUpdateError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "update")
		ctx = context.WithValue(ctx, goa.ServiceKey, "cars")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		v := &carssvc.UpdateEndpointInput{
			Stream: &updateServerStream{
				upgrader:     up,
				connConfigFn: connConfigFn,
				w:            w,
				r:            r,
			},
			Payload: payload.(*carssvc.UpdatePayload),
		}
		_, err = endpoint(ctx, v)

		if err != nil {
			if _, ok := err.(websocket.HandshakeError); ok {
				return
			}
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
	})
}

// Send streams instances of "carssvc.StoredCar" to the "list" endpoint
// websocket connection.
func (s *listServerStream) Send(v *carssvc.StoredCar) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		respHdr := make(http.Header)
		respHdr.Add("goa-view", s.view)
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, respHdr)
		if err != nil {
			return
		}
		if s.connConfigFn != nil {
			conn = s.connConfigFn(conn)
		}
		s.conn = conn
	})
	if err != nil {
		return err
	}
	res := carssvc.NewViewedStoredCar(v, s.view)
	body := NewListResponseBody(res.Projected)
	return s.conn.WriteJSON(body)
}

// Close closes the "list" endpoint websocket connection after sending a close
// control message.
func (s *listServerStream) Close() error {
	if s.conn == nil {
		return nil
	}
	err := s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "end of message"),
		time.Now().Add(time.Second),
	)
	if err == websocket.ErrCloseSent {
		return nil
	}
	if err != nil {
		return err
	}
	err = s.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// SetView sets the view to render the carssvc.StoredCar type before sending to
// the "list" endpoint websocket connection.
func (s *listServerStream) SetView(view string) {
	s.view = view
}

// SendAndClose streams instances of "carssvc.StoredCarCollection" to the "add"
// endpoint websocket connection and closes the connection.
func (s *addServerStream) SendAndClose(v carssvc.StoredCarCollection) error {
	defer s.conn.Close()
	res := carssvc.NewViewedStoredCarCollection(v, s.view)
	body := NewAddResponseBody(res.Projected)
	return s.conn.WriteJSON(body)
}

// Recv reads instances of "carssvc.AddStreamingPayload" from the "add"
// endpoint websocket connection.
func (s *addServerStream) Recv() (*carssvc.AddStreamingPayload, error) {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			return
		}
		if s.connConfigFn != nil {
			conn = s.connConfigFn(conn)
		}
		s.conn = conn
	})
	if err != nil {
		return nil, err
	}
	var v **AddStreamingBody
	if err = s.conn.ReadJSON(&v); err != nil {
		return nil, err
	}
	if v == nil {
		return nil, io.EOF
	}
	body := *v
	err = body.Validate()
	if err != nil {
		return nil, err
	}
	return NewAddStreamingBody(body), nil
}

// SetView sets the view to render the carssvc.StoredCarCollection type before
// sending to the "add" endpoint websocket connection.
func (s *addServerStream) SetView(view string) {
	s.view = view
}

// Send streams instances of "carssvc.StoredCarCollection" to the "update"
// endpoint websocket connection.
func (s *updateServerStream) Send(v carssvc.StoredCarCollection) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		respHdr := make(http.Header)
		respHdr.Add("goa-view", s.view)
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, respHdr)
		if err != nil {
			return
		}
		if s.connConfigFn != nil {
			conn = s.connConfigFn(conn)
		}
		s.conn = conn
	})
	if err != nil {
		return err
	}
	res := carssvc.NewViewedStoredCarCollection(v, s.view)
	body := NewUpdateResponseBody(res.Projected)
	return s.conn.WriteJSON(body)
}

// Recv reads instances of "[]*carssvc.Car" from the "update" endpoint
// websocket connection.
func (s *updateServerStream) Recv() ([]*carssvc.Car, error) {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			return
		}
		if s.connConfigFn != nil {
			conn = s.connConfigFn(conn)
		}
		s.conn = conn
	})
	if err != nil {
		return nil, err
	}
	var v *UpdateStreamingBody
	if err = s.conn.ReadJSON(&v); err != nil {
		return nil, err
	}
	if v == nil {
		return nil, io.EOF
	}
	body := *v
	return NewUpdateStreamingBody(body), nil
}

// Close closes the "update" endpoint websocket connection after sending a
// close control message.
func (s *updateServerStream) Close() error {
	if s.conn == nil {
		return nil
	}
	err := s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "end of message"),
		time.Now().Add(time.Second),
	)
	if err == websocket.ErrCloseSent {
		return nil
	}
	if err != nil {
		return err
	}
	err = s.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// SetView sets the view to render the carssvc.StoredCarCollection type before
// sending to the "update" endpoint websocket connection.
func (s *updateServerStream) SetView(view string) {
	s.view = view
}
