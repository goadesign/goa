// This file contains expected HTTP streaming sections used to verify that
// websocket client and server code names the exact catalog-owned wire types.
package testdata

var MixedEndpointsConnConfigurerStructCode = `// ConnConfigurer holds the websocket connection configurer functions for the
// streaming endpoints in "StreamingResultService" service.
type ConnConfigurer struct {
	StreamingResultMethodFn goahttp.ConnConfigureFunc
}
`

var MixedEndpointsConnConfigurerInitCode = `// NewConnConfigurer initializes the websocket connection configurer function
// with fn for all the streaming endpoints in "StreamingResultService" service.
func NewConnConfigurer(fn goahttp.ConnConfigureFunc) *ConnConfigurer {
	return &ConnConfigurer{
		StreamingResultMethodFn: fn,
	}
}
`

var StreamingResultServerHandlerInitCode = `// NewStreamingResultMethodHandler creates a HTTP handler which loads the HTTP
// request and calls the "StreamingResultService" service
// "StreamingResultMethod" endpoint.
func NewStreamingResultMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeStreamingResultMethodRequest(mux, decoder)
		encodeError   = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "StreamingResultMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "StreamingResultService")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &streamingresultservice.StreamingResultMethodEndpointInput{
			Stream: &StreamingResultMethodServerStream{
				upgrader:   upgrader,
				configurer: configurer,
				cancel:     cancel,
				w:          w,
				r:          r,
			},
			Payload: payload,
		}
		_, err = endpoint(ctx, v)
		if err != nil {
			var stream *StreamingResultMethodServerStream
			if wrapper, ok := v.Stream.(interface{ Unwrap() any }); ok {
				stream = wrapper.Unwrap().(*StreamingResultMethodServerStream)
			} else {
				stream = v.Stream.(*StreamingResultMethodServerStream)
			}
			if stream != nil && stream.conn != nil {
				// Response writer has been hijacked, do not encode the error
				if errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	})
}
`

var MixedResultsServerHandlerInitCode = `// NewCreateHandler creates a HTTP handler which loads the HTTP request and
// calls the "MixedResultsService" service "Create" endpoint.
func NewCreateHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeCreateRequest(mux, decoder)
		encodeResponse = EncodeCreateResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "Create")
		ctx = context.WithValue(ctx, goa.ServiceKey, "MixedResultsService")

		// Content negotiation for mixed results (standard HTTP vs SSE)
		acceptHeader := r.Header.Get("Accept")
		if strings.Contains(acceptHeader, "text/event-stream") {
			// Handle SSE request
			payload, err := decodeRequest(r)
			if err != nil {
				if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			v := &mixedresultsservice.CreateEndpointInput{
				Stream: &CreateServerStream{
					w: w,
					r: r,
				},
				Payload: payload,
			}
			_, err = endpoint(ctx, v)
			if err != nil {
				stream := v.Stream.(*CreateServerStream)
				if stream.attempted {
					if errhandler != nil {
						errhandler(ctx, w, err)
					}
					return
				}
				if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
					errhandler(ctx, w, err)
				}
			}
		} else {
			// Handle standard HTTP request
			payload, err := decodeRequest(r)
			if err != nil {
				if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			// Mixed results endpoints always use the generated endpoint input struct.
			// In the standard (non-SSE) mode, Stream discards events and the service
			// must return the synchronous result.
			v := &mixedresultsservice.CreateEndpointInput{
				Stream:  &discardCreateServerStream{},
				Payload: payload,
			}
			res, err := endpoint(ctx, v)
			if err != nil {
				if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			if err := encodeResponse(ctx, w, res); err != nil {
				if errhandler != nil {
					errhandler(ctx, w, err)
				}
			}
		}
	})
}

// discardCreateServerStream implements the mixedresultsservice.CreateServerStream
// interface and drops all events. It is used for mixed results endpoints in
// regular HTTP requests so service implementations can use the stream
// parameter without nil checks.
type discardCreateServerStream struct{}

// Send discards the event.
func (s *discardCreateServerStream) Send(v *mixedresultsservice.Event) error {
	return nil
}

// SendWithContext discards the event.
func (s *discardCreateServerStream) SendWithContext(ctx context.Context, v *mixedresultsservice.Event) error {
	return nil
}

// Close is a no-op.
func (s *discardCreateServerStream) Close() error {
	return nil
}
`

var StreamingResultServerStreamSendCode = `// Send streams instances of "streamingresultservice.UserType" to the
// "StreamingResultMethod" endpoint websocket connection.
func (s *StreamingResultMethodServerStream) Send(v *streamingresultservice.UserType) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	body := NewStreamingResultMethodResponseBody(res)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of "streamingresultservice.UserType" to
// the "StreamingResultMethod" endpoint websocket connection with context.
func (s *StreamingResultMethodServerStream) SendWithContext(ctx context.Context, v *streamingresultservice.UserType) error {
	return s.Send(v)
}
`

var StreamingResultServerStreamCloseCode = `// Close closes the "StreamingResultMethod" endpoint websocket connection.
func (s *StreamingResultMethodServerStream) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.close()
	})
	return s.closeErr
}

// close opens the websocket connection when needed, sends its normal close
// message, and closes it.
func (s *StreamingResultMethodServerStream) close() error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Close().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	if err = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var StreamingResultWithViewsServerStreamSendCode = `// Send streams instances of "streamingresultwithviewsservice.Usertype" to the
// "StreamingResultWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultWithViewsMethodServerStream) Send(v *streamingresultwithviewsservice.Usertype) error {
	view := s.view
	if view == "" {
		view = "default"
	}
	if s.sentView != "" && view != s.sentView {
		return goa.InvalidEnumValueError("view", view, []any{s.sentView})
	}
	switch view {
	case "tiny":
	case "extended":
	case "default":
	default:
		return goa.InvalidEnumValueError("view", view, []any{"tiny", "extended", "default"})
	}
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		respHdr := make(http.Header)
		respHdr.Add("goa-view", view)
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, respHdr)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	if s.sentView == "" {
		s.sentView = view
	}
	switch view {
	case "tiny":
		res := streamingresultwithviewsservice.NewViewedUsertype(v, "tiny")
		return s.conn.WriteJSON(NewStreamingResultWithViewsMethodResponseBodyTiny(res.Projected))
	case "extended":
		res := streamingresultwithviewsservice.NewViewedUsertype(v, "extended")
		return s.conn.WriteJSON(NewStreamingResultWithViewsMethodResponseBodyExtended(res.Projected))
	case "default", "":
		res := streamingresultwithviewsservice.NewViewedUsertype(v, "default")
		return s.conn.WriteJSON(NewStreamingResultWithViewsMethodResponseBody(res.Projected))
	default:
		return goa.InvalidEnumValueError("view", view, []any{"tiny", "extended", "default"})
	}
}

// SendWithContext streams instances of
// "streamingresultwithviewsservice.Usertype" to the
// "StreamingResultWithViewsMethod" endpoint websocket connection with context.
func (s *StreamingResultWithViewsMethodServerStream) SendWithContext(ctx context.Context, v *streamingresultwithviewsservice.Usertype) error {
	return s.Send(v)
}
`

var StreamingResultWithViewsServerStreamSetViewCode = `// SetView sets the view to render the streamingresultwithviewsservice.Usertype
// type before sending to the "StreamingResultWithViewsMethod" endpoint
// websocket connection.
func (s *StreamingResultWithViewsMethodServerStream) SetView(view string) {
	s.view = view
}
`

var StreamingResultNoPayloadServerHandlerInitCode = `// NewStreamingResultNoPayloadMethodHandler creates a HTTP handler which loads
// the HTTP request and calls the "StreamingResultNoPayloadService" service
// "StreamingResultNoPayloadMethod" endpoint.
func NewStreamingResultNoPayloadMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		encodeError = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "StreamingResultNoPayloadMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "StreamingResultNoPayloadService")
		var err error
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &streamingresultnopayloadservice.StreamingResultNoPayloadMethodEndpointInput{
			Stream: &StreamingResultNoPayloadMethodServerStream{
				upgrader:   upgrader,
				configurer: configurer,
				cancel:     cancel,
				w:          w,
				r:          r,
			},
		}
		_, err = endpoint(ctx, v)
		if err != nil {
			var stream *StreamingResultNoPayloadMethodServerStream
			if wrapper, ok := v.Stream.(interface{ Unwrap() any }); ok {
				stream = wrapper.Unwrap().(*StreamingResultNoPayloadMethodServerStream)
			} else {
				stream = v.Stream.(*StreamingResultNoPayloadMethodServerStream)
			}
			if stream != nil && stream.conn != nil {
				// Response writer has been hijacked, do not encode the error
				if errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	})
}
`

var StreamingResultClientEndpointCode = `// StreamingResultMethod returns an endpoint that makes HTTP requests to the
// StreamingResultService service StreamingResultMethod server.
func (c *Client) StreamingResultMethod() goa.Endpoint {
	var (
		decodeResponse = DecodeStreamingResultMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildStreamingResultMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingResultService", "StreamingResultMethod", err)
		}
		if c.configurer.StreamingResultMethodFn != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configurer.StreamingResultMethodFn(conn, cancel)
		}
		go func() {
			<-ctx.Done()
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closing connection"),
				time.Now().Add(time.Second),
			)
			conn.Close()
		}()
		stream := &StreamingResultMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var StreamingResultWithViewsServerStreamCloseCode = `// Close closes the "StreamingResultWithViewsMethod" endpoint websocket
// connection.
func (s *StreamingResultWithViewsMethodServerStream) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.close()
	})
	return s.closeErr
}

// close opens the websocket connection when needed, sends its normal close
// message, and closes it.
func (s *StreamingResultWithViewsMethodServerStream) close() error {
	var err error
	view := s.view
	if view == "" {
		view = "default"
	}
	switch view {
	case "tiny":
	case "extended":
	case "default":
	default:
		return goa.InvalidEnumValueError("view", view, []any{"tiny", "extended", "default"})
	}
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Close().
	s.once.Do(func() {
		respHdr := make(http.Header)
		respHdr.Add("goa-view", view)
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, respHdr)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	if err = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var StreamingResultClientStreamRecvCode = `// Recv reads instances of "streamingresultservice.UserType" from the
// "StreamingResultMethod" endpoint websocket connection.
func (s *StreamingResultMethodClientStream) Recv() (*streamingresultservice.UserType, error) {
	var (
		rv   *streamingresultservice.UserType
		body StreamingResultMethodResponseBody
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingResultMethodUserTypeOK(&body)
	return res, nil
}

// RecvWithContext reads instances of "streamingresultservice.UserType" from
// the "StreamingResultMethod" endpoint websocket connection with context.
func (s *StreamingResultMethodClientStream) RecvWithContext(ctx context.Context) (*streamingresultservice.UserType, error) {
	return s.Recv()
}
`

var StreamingResultWithViewsClientEndpointCode = `// StreamingResultWithViewsMethod returns an endpoint that makes HTTP requests
// to the StreamingResultWithViewsService service
// StreamingResultWithViewsMethod server.
func (c *Client) StreamingResultWithViewsMethod() goa.Endpoint {
	var (
		decodeResponse = DecodeStreamingResultWithViewsMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildStreamingResultWithViewsMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingResultWithViewsService", "StreamingResultWithViewsMethod", err)
		}
		if c.configurer.StreamingResultWithViewsMethodFn != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configurer.StreamingResultWithViewsMethodFn(conn, cancel)
		}
		go func() {
			<-ctx.Done()
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closing connection"),
				time.Now().Add(time.Second),
			)
			conn.Close()
		}()
		stream := &StreamingResultWithViewsMethodClientStream{conn: conn}
		view := resp.Header.Get("goa-view")
		stream.SetView(view)
		return stream, nil
	}
}
`

var StreamingResultWithViewsClientStreamRecvCode = `// Recv reads instances of "streamingresultwithviewsservice.Usertype" from the
// "StreamingResultWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultWithViewsMethodClientStream) Recv() (*streamingresultwithviewsservice.Usertype, error) {
	var (
		rv   *streamingresultwithviewsservice.Usertype
		body StreamingResultWithViewsMethodResponseBody
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingResultWithViewsMethodUsertypeOK(&body)
	vres := &streamingresultwithviewsserviceviews.Usertype{Projected: res, View: s.view}
	if err := streamingresultwithviewsserviceviews.ValidateUsertype(vres); err != nil {
		return rv, goahttp.ErrValidationError("StreamingResultWithViewsService", "StreamingResultWithViewsMethod", err)
	}
	return streamingresultwithviewsservice.NewUsertype(vres), nil
}

// RecvWithContext reads instances of
// "streamingresultwithviewsservice.Usertype" from the
// "StreamingResultWithViewsMethod" endpoint websocket connection with context.
func (s *StreamingResultWithViewsMethodClientStream) RecvWithContext(ctx context.Context) (*streamingresultwithviewsservice.Usertype, error) {
	return s.Recv()
}
`

var StreamingResultWithViewsClientStreamSetViewCode = `// SetView sets the view to render the  type before sending to the
// "StreamingResultWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultWithViewsMethodClientStream) SetView(view string) {
	s.view = view
}
`

var StreamingResultWithExplicitViewClientEndpointCode = `// StreamingResultWithExplicitViewMethod returns an endpoint that makes HTTP
// requests to the StreamingResultWithExplicitViewService service
// StreamingResultWithExplicitViewMethod server.
func (c *Client) StreamingResultWithExplicitViewMethod() goa.Endpoint {
	var (
		decodeResponse = DecodeStreamingResultWithExplicitViewMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildStreamingResultWithExplicitViewMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingResultWithExplicitViewService", "StreamingResultWithExplicitViewMethod", err)
		}
		if c.configurer.StreamingResultWithExplicitViewMethodFn != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configurer.StreamingResultWithExplicitViewMethodFn(conn, cancel)
		}
		go func() {
			<-ctx.Done()
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closing connection"),
				time.Now().Add(time.Second),
			)
			conn.Close()
		}()
		stream := &StreamingResultWithExplicitViewMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var StreamingResultWithExplicitViewClientStreamRecvCode = `// Recv reads instances of "streamingresultwithexplicitviewservice.Usertype"
// from the "StreamingResultWithExplicitViewMethod" endpoint websocket
// connection.
func (s *StreamingResultWithExplicitViewMethodClientStream) Recv() (*streamingresultwithexplicitviewservice.Usertype, error) {
	var (
		rv   *streamingresultwithexplicitviewservice.Usertype
		body StreamingResultWithExplicitViewMethodResponseBodyExtended
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingResultWithExplicitViewMethodUsertypeOK(&body)
	vres := &streamingresultwithexplicitviewserviceviews.Usertype{Projected: res, View: "extended"}
	if err := streamingresultwithexplicitviewserviceviews.ValidateUsertype(vres); err != nil {
		return rv, goahttp.ErrValidationError("StreamingResultWithExplicitViewService", "StreamingResultWithExplicitViewMethod", err)
	}
	return streamingresultwithexplicitviewservice.NewUsertype(vres), nil
}

// RecvWithContext reads instances of
// "streamingresultwithexplicitviewservice.Usertype" from the
// "StreamingResultWithExplicitViewMethod" endpoint websocket connection with
// context.
func (s *StreamingResultWithExplicitViewMethodClientStream) RecvWithContext(ctx context.Context) (*streamingresultwithexplicitviewservice.Usertype, error) {
	return s.Recv()
}
`

var StreamingResultWithExplicitViewServerStreamSendCode = `// Send streams instances of "streamingresultwithexplicitviewservice.Usertype"
// to the "StreamingResultWithExplicitViewMethod" endpoint websocket connection.
func (s *StreamingResultWithExplicitViewMethodServerStream) Send(v *streamingresultwithexplicitviewservice.Usertype) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := streamingresultwithexplicitviewservice.NewViewedUsertype(v, "extended")
	body := NewStreamingResultWithExplicitViewMethodResponseBodyExtended(res.Projected)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "streamingresultwithexplicitviewservice.Usertype" to the
// "StreamingResultWithExplicitViewMethod" endpoint websocket connection with
// context.
func (s *StreamingResultWithExplicitViewMethodServerStream) SendWithContext(ctx context.Context, v *streamingresultwithexplicitviewservice.Usertype) error {
	return s.Send(v)
}
`

var StreamingResultCollectionWithViewsServerStreamSendCode = `// Send streams instances of
// "streamingresultcollectionwithviewsservice.UsertypeCollection" to the
// "StreamingResultCollectionWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultCollectionWithViewsMethodServerStream) Send(v streamingresultcollectionwithviewsservice.UsertypeCollection) error {
	view := s.view
	if view == "" {
		view = "default"
	}
	if s.sentView != "" && view != s.sentView {
		return goa.InvalidEnumValueError("view", view, []any{s.sentView})
	}
	switch view {
	case "tiny":
	case "extended":
	case "default":
	default:
		return goa.InvalidEnumValueError("view", view, []any{"tiny", "extended", "default"})
	}
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		respHdr := make(http.Header)
		respHdr.Add("goa-view", view)
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, respHdr)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	if s.sentView == "" {
		s.sentView = view
	}
	switch view {
	case "tiny":
		res := streamingresultcollectionwithviewsservice.NewViewedUsertypeCollection(v, "tiny")
		return s.conn.WriteJSON(NewUsertypeResponseTinyCollection(res.Projected))
	case "extended":
		res := streamingresultcollectionwithviewsservice.NewViewedUsertypeCollection(v, "extended")
		return s.conn.WriteJSON(NewUsertypeResponseExtendedCollection(res.Projected))
	case "default", "":
		res := streamingresultcollectionwithviewsservice.NewViewedUsertypeCollection(v, "default")
		return s.conn.WriteJSON(NewUsertypeResponseCollection(res.Projected))
	default:
		return goa.InvalidEnumValueError("view", view, []any{"tiny", "extended", "default"})
	}
}

// SendWithContext streams instances of
// "streamingresultcollectionwithviewsservice.UsertypeCollection" to the
// "StreamingResultCollectionWithViewsMethod" endpoint websocket connection
// with context.
func (s *StreamingResultCollectionWithViewsMethodServerStream) SendWithContext(ctx context.Context, v streamingresultcollectionwithviewsservice.UsertypeCollection) error {
	return s.Send(v)
}
`

var StreamingResultCollectionWithViewsServerStreamSetViewCode = `// SetView sets the view to render the
// streamingresultcollectionwithviewsservice.UsertypeCollection type before
// sending to the "StreamingResultCollectionWithViewsMethod" endpoint websocket
// connection.
func (s *StreamingResultCollectionWithViewsMethodServerStream) SetView(view string) {
	s.view = view
}
`

var StreamingResultCollectionWithViewsClientStreamRecvCode = `// Recv reads instances of
// "streamingresultcollectionwithviewsservice.UsertypeCollection" from the
// "StreamingResultCollectionWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultCollectionWithViewsMethodClientStream) Recv() (streamingresultcollectionwithviewsservice.UsertypeCollection, error) {
	var (
		rv   streamingresultcollectionwithviewsservice.UsertypeCollection
		body UsertypeResponseCollection
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingResultCollectionWithViewsMethodUsertypeCollectionOK(body)
	vres := streamingresultcollectionwithviewsserviceviews.UsertypeCollection{Projected: res, View: s.view}
	if err := streamingresultcollectionwithviewsserviceviews.ValidateUsertypeCollection(vres); err != nil {
		return rv, goahttp.ErrValidationError("StreamingResultCollectionWithViewsService", "StreamingResultCollectionWithViewsMethod", err)
	}
	return streamingresultcollectionwithviewsservice.NewUsertypeCollection(vres), nil
}

// RecvWithContext reads instances of
// "streamingresultcollectionwithviewsservice.UsertypeCollection" from the
// "StreamingResultCollectionWithViewsMethod" endpoint websocket connection
// with context.
func (s *StreamingResultCollectionWithViewsMethodClientStream) RecvWithContext(ctx context.Context) (streamingresultcollectionwithviewsservice.UsertypeCollection, error) {
	return s.Recv()
}
`

var StreamingResultCollectionWithViewsClientStreamSetViewCode = `// SetView sets the view to render the  type before sending to the
// "StreamingResultCollectionWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultCollectionWithViewsMethodClientStream) SetView(view string) {
	s.view = view
}
`

var StreamingResultCollectionWithExplicitViewServerStreamSendCode = `// Send streams instances of
// "streamingresultcollectionwithexplicitviewservice.UsertypeCollection" to the
// "StreamingResultCollectionWithExplicitViewMethod" endpoint websocket
// connection.
func (s *StreamingResultCollectionWithExplicitViewMethodServerStream) Send(v streamingresultcollectionwithexplicitviewservice.UsertypeCollection) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := streamingresultcollectionwithexplicitviewservice.NewViewedUsertypeCollection(v, "tiny")
	body := NewUsertypeResponseTinyCollection(res.Projected)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "streamingresultcollectionwithexplicitviewservice.UsertypeCollection" to the
// "StreamingResultCollectionWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *StreamingResultCollectionWithExplicitViewMethodServerStream) SendWithContext(ctx context.Context, v streamingresultcollectionwithexplicitviewservice.UsertypeCollection) error {
	return s.Send(v)
}
`
var StreamingResultCollectionWithExplicitViewClientEndpointCode = `// StreamingResultCollectionWithExplicitViewMethod returns an endpoint that
// makes HTTP requests to the StreamingResultCollectionWithExplicitViewService
// service StreamingResultCollectionWithExplicitViewMethod server.
func (c *Client) StreamingResultCollectionWithExplicitViewMethod() goa.Endpoint {
	var (
		decodeResponse = DecodeStreamingResultCollectionWithExplicitViewMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildStreamingResultCollectionWithExplicitViewMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingResultCollectionWithExplicitViewService", "StreamingResultCollectionWithExplicitViewMethod", err)
		}
		if c.configurer.StreamingResultCollectionWithExplicitViewMethodFn != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configurer.StreamingResultCollectionWithExplicitViewMethodFn(conn, cancel)
		}
		go func() {
			<-ctx.Done()
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closing connection"),
				time.Now().Add(time.Second),
			)
			conn.Close()
		}()
		stream := &StreamingResultCollectionWithExplicitViewMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var StreamingResultCollectionWithExplicitViewClientStreamRecvCode = `// Recv reads instances of
// "streamingresultcollectionwithexplicitviewservice.UsertypeCollection" from
// the "StreamingResultCollectionWithExplicitViewMethod" endpoint websocket
// connection.
func (s *StreamingResultCollectionWithExplicitViewMethodClientStream) Recv() (streamingresultcollectionwithexplicitviewservice.UsertypeCollection, error) {
	var (
		rv   streamingresultcollectionwithexplicitviewservice.UsertypeCollection
		body UsertypeResponseTinyCollection
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingResultCollectionWithExplicitViewMethodUsertypeCollectionOK(body)
	vres := streamingresultcollectionwithexplicitviewserviceviews.UsertypeCollection{Projected: res, View: "tiny"}
	if err := streamingresultcollectionwithexplicitviewserviceviews.ValidateUsertypeCollection(vres); err != nil {
		return rv, goahttp.ErrValidationError("StreamingResultCollectionWithExplicitViewService", "StreamingResultCollectionWithExplicitViewMethod", err)
	}
	return streamingresultcollectionwithexplicitviewservice.NewUsertypeCollection(vres), nil
}

// RecvWithContext reads instances of
// "streamingresultcollectionwithexplicitviewservice.UsertypeCollection" from
// the "StreamingResultCollectionWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *StreamingResultCollectionWithExplicitViewMethodClientStream) RecvWithContext(ctx context.Context) (streamingresultcollectionwithexplicitviewservice.UsertypeCollection, error) {
	return s.Recv()
}
`

var StreamingResultPrimitiveServerStreamSendCode = `// Send streams instances of "string" to the "StreamingResultPrimitiveMethod"
// endpoint websocket connection.
func (s *StreamingResultPrimitiveMethodServerStream) Send(v string) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	return s.conn.WriteJSON(res)
}

// SendWithContext streams instances of "string" to the
// "StreamingResultPrimitiveMethod" endpoint websocket connection with context.
func (s *StreamingResultPrimitiveMethodServerStream) SendWithContext(ctx context.Context, v string) error {
	return s.Send(v)
}
`

var StreamingResultPrimitiveClientStreamRecvCode = `// Recv reads instances of "string" from the "StreamingResultPrimitiveMethod"
// endpoint websocket connection.
func (s *StreamingResultPrimitiveMethodClientStream) Recv() (string, error) {
	var (
		rv   string
		body string
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// RecvWithContext reads instances of "string" from the
// "StreamingResultPrimitiveMethod" endpoint websocket connection with context.
func (s *StreamingResultPrimitiveMethodClientStream) RecvWithContext(ctx context.Context) (string, error) {
	return s.Recv()
}
`

var StreamingResultPrimitiveArrayServerStreamSendCode = `// Send streams instances of "[]int32" to the
// "StreamingResultPrimitiveArrayMethod" endpoint websocket connection.
func (s *StreamingResultPrimitiveArrayMethodServerStream) Send(v []int32) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	return s.conn.WriteJSON(res)
}

// SendWithContext streams instances of "[]int32" to the
// "StreamingResultPrimitiveArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingResultPrimitiveArrayMethodServerStream) SendWithContext(ctx context.Context, v []int32) error {
	return s.Send(v)
}
`

var StreamingResultPrimitiveArrayClientStreamRecvCode = `// Recv reads instances of "[]int32" from the
// "StreamingResultPrimitiveArrayMethod" endpoint websocket connection.
func (s *StreamingResultPrimitiveArrayMethodClientStream) Recv() ([]int32, error) {
	var (
		rv   []int32
		body []int32
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// RecvWithContext reads instances of "[]int32" from the
// "StreamingResultPrimitiveArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingResultPrimitiveArrayMethodClientStream) RecvWithContext(ctx context.Context) ([]int32, error) {
	return s.Recv()
}
`

var StreamingResultPrimitiveMapServerStreamSendCode = `// Send streams instances of "map[int32]string" to the
// "StreamingResultPrimitiveMapMethod" endpoint websocket connection.
func (s *StreamingResultPrimitiveMapMethodServerStream) Send(v map[int32]string) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	return s.conn.WriteJSON(res)
}

// SendWithContext streams instances of "map[int32]string" to the
// "StreamingResultPrimitiveMapMethod" endpoint websocket connection with
// context.
func (s *StreamingResultPrimitiveMapMethodServerStream) SendWithContext(ctx context.Context, v map[int32]string) error {
	return s.Send(v)
}
`

var StreamingResultPrimitiveMapClientStreamRecvCode = `// Recv reads instances of "map[int32]string" from the
// "StreamingResultPrimitiveMapMethod" endpoint websocket connection.
func (s *StreamingResultPrimitiveMapMethodClientStream) Recv() (map[int32]string, error) {
	var (
		rv   map[int32]string
		body map[int32]string
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// RecvWithContext reads instances of "map[int32]string" from the
// "StreamingResultPrimitiveMapMethod" endpoint websocket connection with
// context.
func (s *StreamingResultPrimitiveMapMethodClientStream) RecvWithContext(ctx context.Context) (map[int32]string, error) {
	return s.Recv()
}
`

var StreamingResultUserTypeArrayServerStreamSendCode = `// Send streams instances of "[]*streamingresultusertypearrayservice.UserType"
// to the "StreamingResultUserTypeArrayMethod" endpoint websocket connection.
func (s *StreamingResultUserTypeArrayMethodServerStream) Send(v []*streamingresultusertypearrayservice.UserType) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	body := NewStreamingResultUserTypeArrayMethodResponseBody(res)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "[]*streamingresultusertypearrayservice.UserType" to the
// "StreamingResultUserTypeArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingResultUserTypeArrayMethodServerStream) SendWithContext(ctx context.Context, v []*streamingresultusertypearrayservice.UserType) error {
	return s.Send(v)
}
`

var StreamingResultUserTypeArrayClientStreamRecvCode = `// Recv reads instances of "[]*streamingresultusertypearrayservice.UserType"
// from the "StreamingResultUserTypeArrayMethod" endpoint websocket connection.
func (s *StreamingResultUserTypeArrayMethodClientStream) Recv() ([]*streamingresultusertypearrayservice.UserType, error) {
	var (
		rv   []*streamingresultusertypearrayservice.UserType
		body []*UserTypeResponse
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingResultUserTypeArrayMethodUserTypeOK(body)
	return res, nil
}

// RecvWithContext reads instances of
// "[]*streamingresultusertypearrayservice.UserType" from the
// "StreamingResultUserTypeArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingResultUserTypeArrayMethodClientStream) RecvWithContext(ctx context.Context) ([]*streamingresultusertypearrayservice.UserType, error) {
	return s.Recv()
}
`

var StreamingResultUserTypeMapServerStreamSendCode = `// Send streams instances of
// "map[string]*streamingresultusertypemapservice.UserType" to the
// "StreamingResultUserTypeMapMethod" endpoint websocket connection.
func (s *StreamingResultUserTypeMapMethodServerStream) Send(v map[string]*streamingresultusertypemapservice.UserType) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	body := NewStreamingResultUserTypeMapMethodResponseBody(res)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "map[string]*streamingresultusertypemapservice.UserType" to the
// "StreamingResultUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *StreamingResultUserTypeMapMethodServerStream) SendWithContext(ctx context.Context, v map[string]*streamingresultusertypemapservice.UserType) error {
	return s.Send(v)
}
`

var StreamingResultUserTypeMapClientStreamRecvCode = `// Recv reads instances of
// "map[string]*streamingresultusertypemapservice.UserType" from the
// "StreamingResultUserTypeMapMethod" endpoint websocket connection.
func (s *StreamingResultUserTypeMapMethodClientStream) Recv() (map[string]*streamingresultusertypemapservice.UserType, error) {
	var (
		rv   map[string]*streamingresultusertypemapservice.UserType
		body map[string]*UserTypeResponse
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingResultUserTypeMapMethodMapStringUserTypeOK(body)
	return res, nil
}

// RecvWithContext reads instances of
// "map[string]*streamingresultusertypemapservice.UserType" from the
// "StreamingResultUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *StreamingResultUserTypeMapMethodClientStream) RecvWithContext(ctx context.Context) (map[string]*streamingresultusertypemapservice.UserType, error) {
	return s.Recv()
}
`

var StreamingResultNoPayloadClientEndpointCode = `// StreamingResultNoPayloadMethod returns an endpoint that makes HTTP requests
// to the StreamingResultNoPayloadService service
// StreamingResultNoPayloadMethod server.
func (c *Client) StreamingResultNoPayloadMethod() goa.Endpoint {
	var (
		decodeResponse = DecodeStreamingResultNoPayloadMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildStreamingResultNoPayloadMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingResultNoPayloadService", "StreamingResultNoPayloadMethod", err)
		}
		if c.configurer.StreamingResultNoPayloadMethodFn != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configurer.StreamingResultNoPayloadMethodFn(conn, cancel)
		}
		go func() {
			<-ctx.Done()
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closing connection"),
				time.Now().Add(time.Second),
			)
			conn.Close()
		}()
		stream := &StreamingResultNoPayloadMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var StreamingPayloadServerHandlerInitCode = `// NewStreamingPayloadMethodHandler creates a HTTP handler which loads the HTTP
// request and calls the "StreamingPayloadService" service
// "StreamingPayloadMethod" endpoint.
func NewStreamingPayloadMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeStreamingPayloadMethodRequest(mux, decoder)
		encodeError   = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "StreamingPayloadMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "StreamingPayloadService")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &streamingpayloadservice.StreamingPayloadMethodEndpointInput{
			Stream: &StreamingPayloadMethodServerStream{
				upgrader:   upgrader,
				configurer: configurer,
				cancel:     cancel,
				w:          w,
				r:          r,
			},
			Payload: payload,
		}
		_, err = endpoint(ctx, v)
		if err != nil {
			var stream *StreamingPayloadMethodServerStream
			if wrapper, ok := v.Stream.(interface{ Unwrap() any }); ok {
				stream = wrapper.Unwrap().(*StreamingPayloadMethodServerStream)
			} else {
				stream = v.Stream.(*StreamingPayloadMethodServerStream)
			}
			if stream != nil && stream.conn != nil {
				// Response writer has been hijacked, do not encode the error
				if errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	})
}
`

var StreamingPayloadServerStreamSendCode = `// SendAndClose streams instances of "streamingpayloadservice.UserType" to the
// "StreamingPayloadMethod" endpoint websocket connection and closes the
// connection.
func (s *StreamingPayloadMethodServerStream) SendAndClose(v *streamingpayloadservice.UserType) error {
	defer s.conn.Close()
	res := v
	body := NewStreamingPayloadMethodResponseBody(res)
	return s.conn.WriteJSON(body)
}

// SendAndCloseWithContext streams instances of
// "streamingpayloadservice.UserType" to the "StreamingPayloadMethod" endpoint
// websocket connection with context and closes the connection.
func (s *StreamingPayloadMethodServerStream) SendAndCloseWithContext(ctx context.Context, v *streamingpayloadservice.UserType) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadServerStreamRecvCode = `// Recv reads instances of "streamingpayloadservice.Request" from the
// "StreamingPayloadMethod" endpoint websocket connection.
func (s *StreamingPayloadMethodServerStream) Recv() (*streamingpayloadservice.Request, error) {
	var (
		rv  *streamingpayloadservice.Request
		msg *StreamingPayloadMethodStreamingBody
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return NewStreamingPayloadMethodStreamingBody(msg), nil
}

// RecvWithContext reads instances of "streamingpayloadservice.Request" from
// the "StreamingPayloadMethod" endpoint websocket connection with context.
func (s *StreamingPayloadMethodServerStream) RecvWithContext(ctx context.Context) (*streamingpayloadservice.Request, error) {
	return s.Recv()
}
`

var StreamingPayloadClientEndpointCode = `// StreamingPayloadMethod returns an endpoint that makes HTTP requests to the
// StreamingPayloadService service StreamingPayloadMethod server.
func (c *Client) StreamingPayloadMethod() goa.Endpoint {
	var (
		encodeRequest  = EncodeStreamingPayloadMethodRequest(c.encoder)
		decodeResponse = DecodeStreamingPayloadMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildStreamingPayloadMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingPayloadService", "StreamingPayloadMethod", err)
		}
		if c.configurer.StreamingPayloadMethodFn != nil {
			conn = c.configurer.StreamingPayloadMethodFn(conn, nil)
		}
		stream := &StreamingPayloadMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var StreamingPayloadClientStreamSendCode = `// Send streams instances of "streamingpayloadservice.Request" to the
// "StreamingPayloadMethod" endpoint websocket connection.
func (s *StreamingPayloadMethodClientStream) Send(v *streamingpayloadservice.Request) error {
	body := NewStreamingPayloadMethodStreamingBody(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of "streamingpayloadservice.Request" to
// the "StreamingPayloadMethod" endpoint websocket connection with context.
func (s *StreamingPayloadMethodClientStream) SendWithContext(ctx context.Context, v *streamingpayloadservice.Request) error {
	return s.Send(v)
}
`

var StreamingPayloadClientStreamRecvCode = `// CloseAndRecv stops sending messages to the "StreamingPayloadMethod" endpoint
// websocket connection and reads instances of
// "streamingpayloadservice.UserType" from the connection.
func (s *StreamingPayloadMethodClientStream) CloseAndRecv() (*streamingpayloadservice.UserType, error) {
	var (
		rv   *streamingpayloadservice.UserType
		body StreamingPayloadMethodResponseBody
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingPayloadMethodUserTypeOK(&body)
	return res, nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadMethod" endpoint websocket connection and reads instances
// of "streamingpayloadservice.UserType" from the connection with context.
func (s *StreamingPayloadMethodClientStream) CloseAndRecvWithContext(ctx context.Context) (*streamingpayloadservice.UserType, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadNoPayloadServerHandlerInitCode = `// NewStreamingPayloadNoPayloadMethodHandler creates a HTTP handler which loads
// the HTTP request and calls the "StreamingPayloadNoPayloadService" service
// "StreamingPayloadNoPayloadMethod" endpoint.
func NewStreamingPayloadNoPayloadMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		encodeError = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "StreamingPayloadNoPayloadMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "StreamingPayloadNoPayloadService")
		var err error
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &streamingpayloadnopayloadservice.StreamingPayloadNoPayloadMethodEndpointInput{
			Stream: &StreamingPayloadNoPayloadMethodServerStream{
				upgrader:   upgrader,
				configurer: configurer,
				cancel:     cancel,
				w:          w,
				r:          r,
			},
		}
		_, err = endpoint(ctx, v)
		if err != nil {
			var stream *StreamingPayloadNoPayloadMethodServerStream
			if wrapper, ok := v.Stream.(interface{ Unwrap() any }); ok {
				stream = wrapper.Unwrap().(*StreamingPayloadNoPayloadMethodServerStream)
			} else {
				stream = v.Stream.(*StreamingPayloadNoPayloadMethodServerStream)
			}
			if stream != nil && stream.conn != nil {
				// Response writer has been hijacked, do not encode the error
				if errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	})
}
`

var StreamingPayloadNoPayloadClientEndpointCode = `// StreamingPayloadNoPayloadMethod returns an endpoint that makes HTTP requests
// to the StreamingPayloadNoPayloadService service
// StreamingPayloadNoPayloadMethod server.
func (c *Client) StreamingPayloadNoPayloadMethod() goa.Endpoint {
	var (
		decodeResponse = DecodeStreamingPayloadNoPayloadMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildStreamingPayloadNoPayloadMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingPayloadNoPayloadService", "StreamingPayloadNoPayloadMethod", err)
		}
		if c.configurer.StreamingPayloadNoPayloadMethodFn != nil {
			conn = c.configurer.StreamingPayloadNoPayloadMethodFn(conn, nil)
		}
		stream := &StreamingPayloadNoPayloadMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var StreamingPayloadNoPayloadClientStreamSendCode = `// Send streams instances of "streamingpayloadnopayloadservice.Request" to the
// "StreamingPayloadNoPayloadMethod" endpoint websocket connection.
func (s *StreamingPayloadNoPayloadMethodClientStream) Send(v *streamingpayloadnopayloadservice.Request) error {
	body := NewStreamingPayloadNoPayloadMethodStreamingBody(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "streamingpayloadnopayloadservice.Request" to the
// "StreamingPayloadNoPayloadMethod" endpoint websocket connection with context.
func (s *StreamingPayloadNoPayloadMethodClientStream) SendWithContext(ctx context.Context, v *streamingpayloadnopayloadservice.Request) error {
	return s.Send(v)
}
`

var StreamingPayloadNoPayloadClientStreamRecvCode = `// CloseAndRecv stops sending messages to the "StreamingPayloadNoPayloadMethod"
// endpoint websocket connection and reads instances of
// "streamingpayloadnopayloadservice.UserType" from the connection.
func (s *StreamingPayloadNoPayloadMethodClientStream) CloseAndRecv() (*streamingpayloadnopayloadservice.UserType, error) {
	var (
		rv   *streamingpayloadnopayloadservice.UserType
		body StreamingPayloadNoPayloadMethodResponseBody
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingPayloadNoPayloadMethodUserTypeOK(&body)
	return res, nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadNoPayloadMethod" endpoint websocket connection and reads
// instances of "streamingpayloadnopayloadservice.UserType" from the connection
// with context.
func (s *StreamingPayloadNoPayloadMethodClientStream) CloseAndRecvWithContext(ctx context.Context) (*streamingpayloadnopayloadservice.UserType, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadNoResultServerStreamRecvCode = `// Recv reads instances of "string" from the "StreamingPayloadNoResultMethod"
// endpoint websocket connection.
func (s *StreamingPayloadNoResultMethodServerStream) Recv() (string, error) {
	var (
		rv  string
		msg *string
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return *msg, nil
}

// RecvWithContext reads instances of "string" from the
// "StreamingPayloadNoResultMethod" endpoint websocket connection with context.
func (s *StreamingPayloadNoResultMethodServerStream) RecvWithContext(ctx context.Context) (string, error) {
	return s.Recv()
}
`

var StreamingPayloadNoResultServerStreamCloseCode = `// Close closes the "StreamingPayloadNoResultMethod" endpoint websocket
// connection.
func (s *StreamingPayloadNoResultMethodServerStream) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.close()
	})
	return s.closeErr
}

// close opens the websocket connection when needed, sends its normal close
// message, and closes it.
func (s *StreamingPayloadNoResultMethodServerStream) close() error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Close().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	if err = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var StreamingPayloadNoResultClientStreamSendCode = `// Send streams instances of "string" to the "StreamingPayloadNoResultMethod"
// endpoint websocket connection.
func (s *StreamingPayloadNoResultMethodClientStream) Send(v string) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "string" to the
// "StreamingPayloadNoResultMethod" endpoint websocket connection with context.
func (s *StreamingPayloadNoResultMethodClientStream) SendWithContext(ctx context.Context, v string) error {
	return s.Send(v)
}
`

var StreamingPayloadNoResultClientStreamCloseCode = `// Close closes the "StreamingPayloadNoResultMethod" endpoint websocket
// connection.
func (s *StreamingPayloadNoResultMethodClientStream) Close() error {
	var err error
	// Send a nil payload to the server implying client closing connection.
	if err = s.conn.WriteJSON(nil); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var StreamingPayloadResultWithExplicitViewServerStreamSendCode = `// SendAndClose streams instances of
// "streamingpayloadresultwithexplicitviewservice.Usertype" to the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection
// and closes the connection.
func (s *StreamingPayloadResultWithExplicitViewMethodServerStream) SendAndClose(v *streamingpayloadresultwithexplicitviewservice.Usertype) error {
	defer s.conn.Close()
	res := streamingpayloadresultwithexplicitviewservice.NewViewedUsertype(v, "extended")
	body := NewStreamingPayloadResultWithExplicitViewMethodResponseBodyExtended(res.Projected)
	return s.conn.WriteJSON(body)
}

// SendAndCloseWithContext streams instances of
// "streamingpayloadresultwithexplicitviewservice.Usertype" to the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection
// with context and closes the connection.
func (s *StreamingPayloadResultWithExplicitViewMethodServerStream) SendAndCloseWithContext(ctx context.Context, v *streamingpayloadresultwithexplicitviewservice.Usertype) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadResultWithExplicitViewServerStreamRecvCode = `// Recv reads instances of "float32" from the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection.
func (s *StreamingPayloadResultWithExplicitViewMethodServerStream) Recv() (float32, error) {
	var (
		rv  float32
		msg *float32
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return *msg, nil
}

// RecvWithContext reads instances of "float32" from the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection
// with context.
func (s *StreamingPayloadResultWithExplicitViewMethodServerStream) RecvWithContext(ctx context.Context) (float32, error) {
	return s.Recv()
}
`

var StreamingPayloadResultWithExplicitViewClientStreamSendCode = `// Send streams instances of "float32" to the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection.
func (s *StreamingPayloadResultWithExplicitViewMethodClientStream) Send(v float32) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "float32" to the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection
// with context.
func (s *StreamingPayloadResultWithExplicitViewMethodClientStream) SendWithContext(ctx context.Context, v float32) error {
	return s.Send(v)
}
`

var StreamingPayloadResultWithExplicitViewClientStreamRecvCode = `// CloseAndRecv stops sending messages to the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection
// and reads instances of
// "streamingpayloadresultwithexplicitviewservice.Usertype" from the connection.
func (s *StreamingPayloadResultWithExplicitViewMethodClientStream) CloseAndRecv() (*streamingpayloadresultwithexplicitviewservice.Usertype, error) {
	var (
		rv   *streamingpayloadresultwithexplicitviewservice.Usertype
		body StreamingPayloadResultWithExplicitViewMethodResponseBodyExtended
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingPayloadResultWithExplicitViewMethodUsertypeOK(&body)
	vres := &streamingpayloadresultwithexplicitviewserviceviews.Usertype{Projected: res, View: "extended"}
	if err := streamingpayloadresultwithexplicitviewserviceviews.ValidateUsertype(vres); err != nil {
		return rv, goahttp.ErrValidationError("StreamingPayloadResultWithExplicitViewService", "StreamingPayloadResultWithExplicitViewMethod", err)
	}
	return streamingpayloadresultwithexplicitviewservice.NewUsertype(vres), nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadResultWithExplicitViewMethod" endpoint websocket connection
// and reads instances of
// "streamingpayloadresultwithexplicitviewservice.Usertype" from the connection
// with context.
func (s *StreamingPayloadResultWithExplicitViewMethodClientStream) CloseAndRecvWithContext(ctx context.Context) (*streamingpayloadresultwithexplicitviewservice.Usertype, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadResultCollectionWithExplicitViewServerStreamSendCode = `// SendAndClose streams instances of
// "streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection"
// to the "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint
// websocket connection and closes the connection.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodServerStream) SendAndClose(v streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection) error {
	defer s.conn.Close()
	res := streamingpayloadresultcollectionwithexplicitviewservice.NewViewedUsertypeCollection(v, "tiny")
	body := NewUsertypeResponseTinyCollection(res.Projected)
	return s.conn.WriteJSON(body)
}

// SendAndCloseWithContext streams instances of
// "streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection"
// to the "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint
// websocket connection with context and closes the connection.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodServerStream) SendAndCloseWithContext(ctx context.Context, v streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadResultCollectionWithExplicitViewServerStreamRecvCode = `// Recv reads instances of "any" from the
// "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint websocket
// connection.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodServerStream) Recv() (any, error) {
	var (
		rv  any
		msg *any
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return *msg, nil
}

// RecvWithContext reads instances of "any" from the
// "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodServerStream) RecvWithContext(ctx context.Context) (any, error) {
	return s.Recv()
}
`

var StreamingPayloadResultCollectionWithExplicitViewClientStreamSendCode = `// Send streams instances of "any" to the
// "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint websocket
// connection.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodClientStream) Send(v any) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "any" to the
// "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodClientStream) SendWithContext(ctx context.Context, v any) error {
	return s.Send(v)
}
`

var StreamingPayloadResultCollectionWithExplicitViewClientStreamRecvCode = `// CloseAndRecv stops sending messages to the
// "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint websocket
// connection and reads instances of
// "streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection"
// from the connection.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodClientStream) CloseAndRecv() (streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection, error) {
	var (
		rv   streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection
		body UsertypeResponseTinyCollection
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewStreamingPayloadResultCollectionWithExplicitViewMethodUsertypeCollectionOK(body)
	vres := streamingpayloadresultcollectionwithexplicitviewserviceviews.UsertypeCollection{Projected: res, View: "tiny"}
	if err := streamingpayloadresultcollectionwithexplicitviewserviceviews.ValidateUsertypeCollection(vres); err != nil {
		return rv, goahttp.ErrValidationError("StreamingPayloadResultCollectionWithExplicitViewService", "StreamingPayloadResultCollectionWithExplicitViewMethod", err)
	}
	return streamingpayloadresultcollectionwithexplicitviewservice.NewUsertypeCollection(vres), nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadResultCollectionWithExplicitViewMethod" endpoint websocket
// connection and reads instances of
// "streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection"
// from the connection with context.
func (s *StreamingPayloadResultCollectionWithExplicitViewMethodClientStream) CloseAndRecvWithContext(ctx context.Context) (streamingpayloadresultcollectionwithexplicitviewservice.UsertypeCollection, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadPrimitiveServerStreamSendCode = `// SendAndClose streams instances of "string" to the
// "StreamingPayloadPrimitiveMethod" endpoint websocket connection and closes
// the connection.
func (s *StreamingPayloadPrimitiveMethodServerStream) SendAndClose(v string) error {
	defer s.conn.Close()
	res := v
	return s.conn.WriteJSON(res)
}

// SendAndCloseWithContext streams instances of "string" to the
// "StreamingPayloadPrimitiveMethod" endpoint websocket connection with context
// and closes the connection.
func (s *StreamingPayloadPrimitiveMethodServerStream) SendAndCloseWithContext(ctx context.Context, v string) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadPrimitiveServerStreamRecvCode = `// Recv reads instances of "string" from the "StreamingPayloadPrimitiveMethod"
// endpoint websocket connection.
func (s *StreamingPayloadPrimitiveMethodServerStream) Recv() (string, error) {
	var (
		rv  string
		msg *string
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return *msg, nil
}

// RecvWithContext reads instances of "string" from the
// "StreamingPayloadPrimitiveMethod" endpoint websocket connection with context.
func (s *StreamingPayloadPrimitiveMethodServerStream) RecvWithContext(ctx context.Context) (string, error) {
	return s.Recv()
}
`

var StreamingPayloadPrimitiveClientStreamSendCode = `// Send streams instances of "string" to the "StreamingPayloadPrimitiveMethod"
// endpoint websocket connection.
func (s *StreamingPayloadPrimitiveMethodClientStream) Send(v string) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "string" to the
// "StreamingPayloadPrimitiveMethod" endpoint websocket connection with context.
func (s *StreamingPayloadPrimitiveMethodClientStream) SendWithContext(ctx context.Context, v string) error {
	return s.Send(v)
}
`

var StreamingPayloadPrimitiveClientStreamRecvCode = `// CloseAndRecv stops sending messages to the "StreamingPayloadPrimitiveMethod"
// endpoint websocket connection and reads instances of "string" from the
// connection.
func (s *StreamingPayloadPrimitiveMethodClientStream) CloseAndRecv() (string, error) {
	var (
		rv   string
		body string
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadPrimitiveMethod" endpoint websocket connection and reads
// instances of "string" from the connection with context.
func (s *StreamingPayloadPrimitiveMethodClientStream) CloseAndRecvWithContext(ctx context.Context) (string, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadPrimitiveArrayServerStreamSendCode = `// SendAndClose streams instances of "[]string" to the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection and
// closes the connection.
func (s *StreamingPayloadPrimitiveArrayMethodServerStream) SendAndClose(v []string) error {
	defer s.conn.Close()
	res := v
	return s.conn.WriteJSON(res)
}

// SendAndCloseWithContext streams instances of "[]string" to the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection with
// context and closes the connection.
func (s *StreamingPayloadPrimitiveArrayMethodServerStream) SendAndCloseWithContext(ctx context.Context, v []string) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadPrimitiveArrayServerStreamRecvCode = `// Recv reads instances of "[]int32" from the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection.
func (s *StreamingPayloadPrimitiveArrayMethodServerStream) Recv() ([]int32, error) {
	var (
		rv   []int32
		body []int32
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return body, nil
}

// RecvWithContext reads instances of "[]int32" from the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadPrimitiveArrayMethodServerStream) RecvWithContext(ctx context.Context) ([]int32, error) {
	return s.Recv()
}
`

var StreamingPayloadPrimitiveArrayClientStreamSendCode = `// Send streams instances of "[]int32" to the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection.
func (s *StreamingPayloadPrimitiveArrayMethodClientStream) Send(v []int32) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "[]int32" to the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadPrimitiveArrayMethodClientStream) SendWithContext(ctx context.Context, v []int32) error {
	return s.Send(v)
}
`

var StreamingPayloadPrimitiveArrayClientStreamRecvCode = `// CloseAndRecv stops sending messages to the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection and
// reads instances of "[]string" from the connection.
func (s *StreamingPayloadPrimitiveArrayMethodClientStream) CloseAndRecv() ([]string, error) {
	var (
		rv   []string
		body []string
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadPrimitiveArrayMethod" endpoint websocket connection and
// reads instances of "[]string" from the connection with context.
func (s *StreamingPayloadPrimitiveArrayMethodClientStream) CloseAndRecvWithContext(ctx context.Context) ([]string, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadPrimitiveMapServerStreamSendCode = `// SendAndClose streams instances of "map[int]int" to the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection and
// closes the connection.
func (s *StreamingPayloadPrimitiveMapMethodServerStream) SendAndClose(v map[int]int) error {
	defer s.conn.Close()
	res := v
	return s.conn.WriteJSON(res)
}

// SendAndCloseWithContext streams instances of "map[int]int" to the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection with
// context and closes the connection.
func (s *StreamingPayloadPrimitiveMapMethodServerStream) SendAndCloseWithContext(ctx context.Context, v map[int]int) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadPrimitiveMapServerStreamRecvCode = `// Recv reads instances of "map[string]int32" from the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection.
func (s *StreamingPayloadPrimitiveMapMethodServerStream) Recv() (map[string]int32, error) {
	var (
		rv   map[string]int32
		body map[string]int32
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return body, nil
}

// RecvWithContext reads instances of "map[string]int32" from the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadPrimitiveMapMethodServerStream) RecvWithContext(ctx context.Context) (map[string]int32, error) {
	return s.Recv()
}
`

var StreamingPayloadPrimitiveMapClientStreamSendCode = `// Send streams instances of "map[string]int32" to the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection.
func (s *StreamingPayloadPrimitiveMapMethodClientStream) Send(v map[string]int32) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "map[string]int32" to the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadPrimitiveMapMethodClientStream) SendWithContext(ctx context.Context, v map[string]int32) error {
	return s.Send(v)
}
`

var StreamingPayloadPrimitiveMapClientStreamRecvCode = `// CloseAndRecv stops sending messages to the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection and reads
// instances of "map[int]int" from the connection.
func (s *StreamingPayloadPrimitiveMapMethodClientStream) CloseAndRecv() (map[int]int, error) {
	var (
		rv   map[int]int
		body map[int]int
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadPrimitiveMapMethod" endpoint websocket connection and reads
// instances of "map[int]int" from the connection with context.
func (s *StreamingPayloadPrimitiveMapMethodClientStream) CloseAndRecvWithContext(ctx context.Context) (map[int]int, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadUserTypeArrayServerStreamSendCode = `// SendAndClose streams instances of "string" to the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection and
// closes the connection.
func (s *StreamingPayloadUserTypeArrayMethodServerStream) SendAndClose(v string) error {
	defer s.conn.Close()
	res := v
	return s.conn.WriteJSON(res)
}

// SendAndCloseWithContext streams instances of "string" to the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection with
// context and closes the connection.
func (s *StreamingPayloadUserTypeArrayMethodServerStream) SendAndCloseWithContext(ctx context.Context, v string) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadUserTypeArrayServerStreamRecvCode = `// Recv reads instances of
// "[]*streamingpayloadusertypearrayservice.RequestType" from the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection.
func (s *StreamingPayloadUserTypeArrayMethodServerStream) Recv() ([]*streamingpayloadusertypearrayservice.RequestType, error) {
	var (
		rv   []*streamingpayloadusertypearrayservice.RequestType
		body []*RequestType
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return NewStreamingPayloadUserTypeArrayMethodArray(body), nil
}

// RecvWithContext reads instances of
// "[]*streamingpayloadusertypearrayservice.RequestType" from the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadUserTypeArrayMethodServerStream) RecvWithContext(ctx context.Context) ([]*streamingpayloadusertypearrayservice.RequestType, error) {
	return s.Recv()
}
`

var StreamingPayloadUserTypeArrayClientStreamSendCode = `// Send streams instances of
// "[]*streamingpayloadusertypearrayservice.RequestType" to the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection.
func (s *StreamingPayloadUserTypeArrayMethodClientStream) Send(v []*streamingpayloadusertypearrayservice.RequestType) error {
	body := NewRequestType(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "[]*streamingpayloadusertypearrayservice.RequestType" to the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadUserTypeArrayMethodClientStream) SendWithContext(ctx context.Context, v []*streamingpayloadusertypearrayservice.RequestType) error {
	return s.Send(v)
}
`

var StreamingPayloadUserTypeArrayClientStreamRecvCode = `// CloseAndRecv stops sending messages to the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection and
// reads instances of "string" from the connection.
func (s *StreamingPayloadUserTypeArrayMethodClientStream) CloseAndRecv() (string, error) {
	var (
		rv   string
		body string
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadUserTypeArrayMethod" endpoint websocket connection and
// reads instances of "string" from the connection with context.
func (s *StreamingPayloadUserTypeArrayMethodClientStream) CloseAndRecvWithContext(ctx context.Context) (string, error) {
	return s.CloseAndRecv()
}
`

var StreamingPayloadUserTypeMapServerStreamSendCode = `// SendAndClose streams instances of "[]string" to the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection and closes
// the connection.
func (s *StreamingPayloadUserTypeMapMethodServerStream) SendAndClose(v []string) error {
	defer s.conn.Close()
	res := v
	return s.conn.WriteJSON(res)
}

// SendAndCloseWithContext streams instances of "[]string" to the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection with
// context and closes the connection.
func (s *StreamingPayloadUserTypeMapMethodServerStream) SendAndCloseWithContext(ctx context.Context, v []string) error {
	return s.SendAndClose(v)
}
`

var StreamingPayloadUserTypeMapServerStreamRecvCode = `// Recv reads instances of
// "map[string]*streamingpayloadusertypemapservice.RequestType" from the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection.
func (s *StreamingPayloadUserTypeMapMethodServerStream) Recv() (map[string]*streamingpayloadusertypemapservice.RequestType, error) {
	var (
		rv   map[string]*streamingpayloadusertypemapservice.RequestType
		body map[string]*RequestType
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return NewStreamingPayloadUserTypeMapMethodMap(body), nil
}

// RecvWithContext reads instances of
// "map[string]*streamingpayloadusertypemapservice.RequestType" from the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadUserTypeMapMethodServerStream) RecvWithContext(ctx context.Context) (map[string]*streamingpayloadusertypemapservice.RequestType, error) {
	return s.Recv()
}
`

var StreamingPayloadUserTypeMapClientStreamSendCode = `// Send streams instances of
// "map[string]*streamingpayloadusertypemapservice.RequestType" to the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection.
func (s *StreamingPayloadUserTypeMapMethodClientStream) Send(v map[string]*streamingpayloadusertypemapservice.RequestType) error {
	body := NewMapStringRequestType(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "map[string]*streamingpayloadusertypemapservice.RequestType" to the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *StreamingPayloadUserTypeMapMethodClientStream) SendWithContext(ctx context.Context, v map[string]*streamingpayloadusertypemapservice.RequestType) error {
	return s.Send(v)
}
`

var StreamingPayloadUserTypeMapClientStreamRecvCode = `// CloseAndRecv stops sending messages to the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection and reads
// instances of "[]string" from the connection.
func (s *StreamingPayloadUserTypeMapMethodClientStream) CloseAndRecv() ([]string, error) {
	var (
		rv   []string
		body []string
		err  error
	)
	defer s.conn.Close()
	// Send a nil payload to the server implying end of message
	if err = s.conn.WriteJSON(nil); err != nil {
		return rv, err
	}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		s.conn.Close()
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// CloseAndRecvWithContext stops sending messages to the
// "StreamingPayloadUserTypeMapMethod" endpoint websocket connection and reads
// instances of "[]string" from the connection with context.
func (s *StreamingPayloadUserTypeMapMethodClientStream) CloseAndRecvWithContext(ctx context.Context) ([]string, error) {
	return s.CloseAndRecv()
}
`

var BidirectionalStreamingServerHandlerInitCode = `// NewBidirectionalStreamingMethodHandler creates a HTTP handler which loads
// the HTTP request and calls the "BidirectionalStreamingService" service
// "BidirectionalStreamingMethod" endpoint.
func NewBidirectionalStreamingMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeBidirectionalStreamingMethodRequest(mux, decoder)
		encodeError   = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "BidirectionalStreamingMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "BidirectionalStreamingService")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &bidirectionalstreamingservice.BidirectionalStreamingMethodEndpointInput{
			Stream: &BidirectionalStreamingMethodServerStream{
				upgrader:   upgrader,
				configurer: configurer,
				cancel:     cancel,
				w:          w,
				r:          r,
			},
			Payload: payload,
		}
		_, err = endpoint(ctx, v)
		if err != nil {
			var stream *BidirectionalStreamingMethodServerStream
			if wrapper, ok := v.Stream.(interface{ Unwrap() any }); ok {
				stream = wrapper.Unwrap().(*BidirectionalStreamingMethodServerStream)
			} else {
				stream = v.Stream.(*BidirectionalStreamingMethodServerStream)
			}
			if stream != nil && stream.conn != nil {
				// Response writer has been hijacked, do not encode the error
				if errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	})
}
`

var BidirectionalStreamingServerStreamSendCode = `// Send streams instances of "bidirectionalstreamingservice.UserType" to the
// "BidirectionalStreamingMethod" endpoint websocket connection.
func (s *BidirectionalStreamingMethodServerStream) Send(v *bidirectionalstreamingservice.UserType) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	body := NewBidirectionalStreamingMethodResponseBody(res)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "bidirectionalstreamingservice.UserType" to the
// "BidirectionalStreamingMethod" endpoint websocket connection with context.
func (s *BidirectionalStreamingMethodServerStream) SendWithContext(ctx context.Context, v *bidirectionalstreamingservice.UserType) error {
	return s.Send(v)
}
`

var BidirectionalStreamingServerStreamRecvCode = `// Recv reads instances of "bidirectionalstreamingservice.Request" from the
// "BidirectionalStreamingMethod" endpoint websocket connection.
func (s *BidirectionalStreamingMethodServerStream) Recv() (*bidirectionalstreamingservice.Request, error) {
	var (
		rv  *bidirectionalstreamingservice.Request
		msg *BidirectionalStreamingMethodStreamingBody
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return NewBidirectionalStreamingMethodStreamingBody(msg), nil
}

// RecvWithContext reads instances of "bidirectionalstreamingservice.Request"
// from the "BidirectionalStreamingMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingMethodServerStream) RecvWithContext(ctx context.Context) (*bidirectionalstreamingservice.Request, error) {
	return s.Recv()
}
`

var BidirectionalStreamingServerStreamCloseCode = `// Close closes the "BidirectionalStreamingMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingMethodServerStream) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.close()
	})
	return s.closeErr
}

// close opens the websocket connection when needed, sends its normal close
// message, and closes it.
func (s *BidirectionalStreamingMethodServerStream) close() error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Close().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	if err = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var BidirectionalStreamingClientEndpointCode = `// BidirectionalStreamingMethod returns an endpoint that makes HTTP requests to
// the BidirectionalStreamingService service BidirectionalStreamingMethod
// server.
func (c *Client) BidirectionalStreamingMethod() goa.Endpoint {
	var (
		encodeRequest  = EncodeBidirectionalStreamingMethodRequest(c.encoder)
		decodeResponse = DecodeBidirectionalStreamingMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildBidirectionalStreamingMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("BidirectionalStreamingService", "BidirectionalStreamingMethod", err)
		}
		if c.configurer.BidirectionalStreamingMethodFn != nil {
			conn = c.configurer.BidirectionalStreamingMethodFn(conn, nil)
		}
		stream := &BidirectionalStreamingMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var BidirectionalStreamingClientStreamSendCode = `// Send streams instances of "bidirectionalstreamingservice.Request" to the
// "BidirectionalStreamingMethod" endpoint websocket connection.
func (s *BidirectionalStreamingMethodClientStream) Send(v *bidirectionalstreamingservice.Request) error {
	body := NewBidirectionalStreamingMethodStreamingBody(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of "bidirectionalstreamingservice.Request"
// to the "BidirectionalStreamingMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingMethodClientStream) SendWithContext(ctx context.Context, v *bidirectionalstreamingservice.Request) error {
	return s.Send(v)
}
`

var BidirectionalStreamingClientStreamRecvCode = `// Recv reads instances of "bidirectionalstreamingservice.UserType" from the
// "BidirectionalStreamingMethod" endpoint websocket connection.
func (s *BidirectionalStreamingMethodClientStream) Recv() (*bidirectionalstreamingservice.UserType, error) {
	var (
		rv   *bidirectionalstreamingservice.UserType
		body BidirectionalStreamingMethodResponseBody
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewBidirectionalStreamingMethodUserTypeOK(&body)
	return res, nil
}

// RecvWithContext reads instances of "bidirectionalstreamingservice.UserType"
// from the "BidirectionalStreamingMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingMethodClientStream) RecvWithContext(ctx context.Context) (*bidirectionalstreamingservice.UserType, error) {
	return s.Recv()
}
`

var BidirectionalStreamingClientStreamCloseCode = `// Close closes the "BidirectionalStreamingMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingMethodClientStream) Close() error {
	var err error
	// Send a nil payload to the server implying client closing connection.
	if err = s.conn.WriteJSON(nil); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var BidirectionalStreamingNoPayloadServerHandlerInitCode = `// NewBidirectionalStreamingNoPayloadMethodHandler creates a HTTP handler which
// loads the HTTP request and calls the
// "BidirectionalStreamingNoPayloadService" service
// "BidirectionalStreamingNoPayloadMethod" endpoint.
func NewBidirectionalStreamingNoPayloadMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		encodeError = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "BidirectionalStreamingNoPayloadMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "BidirectionalStreamingNoPayloadService")
		var err error
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &bidirectionalstreamingnopayloadservice.BidirectionalStreamingNoPayloadMethodEndpointInput{
			Stream: &BidirectionalStreamingNoPayloadMethodServerStream{
				upgrader:   upgrader,
				configurer: configurer,
				cancel:     cancel,
				w:          w,
				r:          r,
			},
		}
		_, err = endpoint(ctx, v)
		if err != nil {
			var stream *BidirectionalStreamingNoPayloadMethodServerStream
			if wrapper, ok := v.Stream.(interface{ Unwrap() any }); ok {
				stream = wrapper.Unwrap().(*BidirectionalStreamingNoPayloadMethodServerStream)
			} else {
				stream = v.Stream.(*BidirectionalStreamingNoPayloadMethodServerStream)
			}
			if stream != nil && stream.conn != nil {
				// Response writer has been hijacked, do not encode the error
				if errhandler != nil {
					errhandler(ctx, w, err)
				}
				return
			}
			if err := encodeError(ctx, w, err); err != nil && errhandler != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	})
}
`

var BidirectionalStreamingNoPayloadServerStreamCloseCode = `// Close closes the "BidirectionalStreamingNoPayloadMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingNoPayloadMethodServerStream) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.close()
	})
	return s.closeErr
}

// close opens the websocket connection when needed, sends its normal close
// message, and closes it.
func (s *BidirectionalStreamingNoPayloadMethodServerStream) close() error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Close().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	if err = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var BidirectionalStreamingNoPayloadClientEndpointCode = `// BidirectionalStreamingNoPayloadMethod returns an endpoint that makes HTTP
// requests to the BidirectionalStreamingNoPayloadService service
// BidirectionalStreamingNoPayloadMethod server.
func (c *Client) BidirectionalStreamingNoPayloadMethod() goa.Endpoint {
	var (
		decodeResponse = DecodeBidirectionalStreamingNoPayloadMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.BuildBidirectionalStreamingNoPayloadMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("BidirectionalStreamingNoPayloadService", "BidirectionalStreamingNoPayloadMethod", err)
		}
		if c.configurer.BidirectionalStreamingNoPayloadMethodFn != nil {
			conn = c.configurer.BidirectionalStreamingNoPayloadMethodFn(conn, nil)
		}
		stream := &BidirectionalStreamingNoPayloadMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var BidirectionalStreamingNoPayloadClientStreamSendCode = `// Send streams instances of "bidirectionalstreamingnopayloadservice.Request"
// to the "BidirectionalStreamingNoPayloadMethod" endpoint websocket connection.
func (s *BidirectionalStreamingNoPayloadMethodClientStream) Send(v *bidirectionalstreamingnopayloadservice.Request) error {
	body := NewBidirectionalStreamingNoPayloadMethodStreamingBody(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "bidirectionalstreamingnopayloadservice.Request" to the
// "BidirectionalStreamingNoPayloadMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingNoPayloadMethodClientStream) SendWithContext(ctx context.Context, v *bidirectionalstreamingnopayloadservice.Request) error {
	return s.Send(v)
}
`

var BidirectionalStreamingNoPayloadClientStreamRecvCode = `// Recv reads instances of "bidirectionalstreamingnopayloadservice.UserType"
// from the "BidirectionalStreamingNoPayloadMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingNoPayloadMethodClientStream) Recv() (*bidirectionalstreamingnopayloadservice.UserType, error) {
	var (
		rv   *bidirectionalstreamingnopayloadservice.UserType
		body BidirectionalStreamingNoPayloadMethodResponseBody
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewBidirectionalStreamingNoPayloadMethodUserTypeOK(&body)
	return res, nil
}

// RecvWithContext reads instances of
// "bidirectionalstreamingnopayloadservice.UserType" from the
// "BidirectionalStreamingNoPayloadMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingNoPayloadMethodClientStream) RecvWithContext(ctx context.Context) (*bidirectionalstreamingnopayloadservice.UserType, error) {
	return s.Recv()
}
`

var BidirectionalStreamingNoPayloadClientStreamCloseCode = `// Close closes the "BidirectionalStreamingNoPayloadMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingNoPayloadMethodClientStream) Close() error {
	var err error
	// Send a nil payload to the server implying client closing connection.
	if err = s.conn.WriteJSON(nil); err != nil {
		return err
	}
	return s.conn.Close()
}
`

var BidirectionalStreamingResultWithExplicitViewServerStreamSendCode = `// Send streams instances of
// "bidirectionalstreamingresultwithexplicitviewservice.Usertype" to the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingResultWithExplicitViewMethodServerStream) Send(v *bidirectionalstreamingresultwithexplicitviewservice.Usertype) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := bidirectionalstreamingresultwithexplicitviewservice.NewViewedUsertype(v, "extended")
	body := NewBidirectionalStreamingResultWithExplicitViewMethodResponseBodyExtended(res.Projected)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "bidirectionalstreamingresultwithexplicitviewservice.Usertype" to the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *BidirectionalStreamingResultWithExplicitViewMethodServerStream) SendWithContext(ctx context.Context, v *bidirectionalstreamingresultwithexplicitviewservice.Usertype) error {
	return s.Send(v)
}
`

var BidirectionalStreamingResultWithExplicitViewServerStreamRecvCode = `// Recv reads instances of "float32" from the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingResultWithExplicitViewMethodServerStream) Recv() (float32, error) {
	var (
		rv  float32
		msg *float32
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return *msg, nil
}

// RecvWithContext reads instances of "float32" from the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *BidirectionalStreamingResultWithExplicitViewMethodServerStream) RecvWithContext(ctx context.Context) (float32, error) {
	return s.Recv()
}
`

var BidirectionalStreamingResultWithExplicitViewClientStreamSendCode = `// Send streams instances of "float32" to the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingResultWithExplicitViewMethodClientStream) Send(v float32) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "float32" to the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *BidirectionalStreamingResultWithExplicitViewMethodClientStream) SendWithContext(ctx context.Context, v float32) error {
	return s.Send(v)
}
`

var BidirectionalStreamingResultWithExplicitViewClientStreamRecvCode = `// Recv reads instances of
// "bidirectionalstreamingresultwithexplicitviewservice.Usertype" from the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection.
func (s *BidirectionalStreamingResultWithExplicitViewMethodClientStream) Recv() (*bidirectionalstreamingresultwithexplicitviewservice.Usertype, error) {
	var (
		rv   *bidirectionalstreamingresultwithexplicitviewservice.Usertype
		body BidirectionalStreamingResultWithExplicitViewMethodResponseBodyExtended
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewBidirectionalStreamingResultWithExplicitViewMethodUsertypeOK(&body)
	vres := &bidirectionalstreamingresultwithexplicitviewserviceviews.Usertype{Projected: res, View: "extended"}
	if err := bidirectionalstreamingresultwithexplicitviewserviceviews.ValidateUsertype(vres); err != nil {
		return rv, goahttp.ErrValidationError("BidirectionalStreamingResultWithExplicitViewService", "BidirectionalStreamingResultWithExplicitViewMethod", err)
	}
	return bidirectionalstreamingresultwithexplicitviewservice.NewUsertype(vres), nil
}

// RecvWithContext reads instances of
// "bidirectionalstreamingresultwithexplicitviewservice.Usertype" from the
// "BidirectionalStreamingResultWithExplicitViewMethod" endpoint websocket
// connection with context.
func (s *BidirectionalStreamingResultWithExplicitViewMethodClientStream) RecvWithContext(ctx context.Context) (*bidirectionalstreamingresultwithexplicitviewservice.Usertype, error) {
	return s.Recv()
}
`

var BidirectionalStreamingResultCollectionWithExplicitViewServerStreamSendCode = `// Send streams instances of
// "bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection"
// to the "BidirectionalStreamingResultCollectionWithExplicitViewMethod"
// endpoint websocket connection.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodServerStream) Send(v bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := bidirectionalstreamingresultcollectionwithexplicitviewservice.NewViewedUsertypeCollection(v, "tiny")
	body := NewUsertypeResponseTinyCollection(res.Projected)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection"
// to the "BidirectionalStreamingResultCollectionWithExplicitViewMethod"
// endpoint websocket connection with context.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodServerStream) SendWithContext(ctx context.Context, v bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection) error {
	return s.Send(v)
}
`

var BidirectionalStreamingResultCollectionWithExplicitViewServerStreamRecvCode = `// Recv reads instances of "any" from the
// "BidirectionalStreamingResultCollectionWithExplicitViewMethod" endpoint
// websocket connection.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodServerStream) Recv() (any, error) {
	var (
		rv  any
		msg *any
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return *msg, nil
}

// RecvWithContext reads instances of "any" from the
// "BidirectionalStreamingResultCollectionWithExplicitViewMethod" endpoint
// websocket connection with context.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodServerStream) RecvWithContext(ctx context.Context) (any, error) {
	return s.Recv()
}
`

var BidirectionalStreamingResultCollectionWithExplicitViewClientStreamSendCode = `// Send streams instances of "any" to the
// "BidirectionalStreamingResultCollectionWithExplicitViewMethod" endpoint
// websocket connection.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodClientStream) Send(v any) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "any" to the
// "BidirectionalStreamingResultCollectionWithExplicitViewMethod" endpoint
// websocket connection with context.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodClientStream) SendWithContext(ctx context.Context, v any) error {
	return s.Send(v)
}
`

var BidirectionalStreamingResultCollectionWithExplicitViewClientStreamRecvCode = `// Recv reads instances of
// "bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection"
// from the "BidirectionalStreamingResultCollectionWithExplicitViewMethod"
// endpoint websocket connection.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodClientStream) Recv() (bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection, error) {
	var (
		rv   bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection
		body UsertypeResponseTinyCollection
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewBidirectionalStreamingResultCollectionWithExplicitViewMethodUsertypeCollectionOK(body)
	vres := bidirectionalstreamingresultcollectionwithexplicitviewserviceviews.UsertypeCollection{Projected: res, View: "tiny"}
	if err := bidirectionalstreamingresultcollectionwithexplicitviewserviceviews.ValidateUsertypeCollection(vres); err != nil {
		return rv, goahttp.ErrValidationError("BidirectionalStreamingResultCollectionWithExplicitViewService", "BidirectionalStreamingResultCollectionWithExplicitViewMethod", err)
	}
	return bidirectionalstreamingresultcollectionwithexplicitviewservice.NewUsertypeCollection(vres), nil
}

// RecvWithContext reads instances of
// "bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection"
// from the "BidirectionalStreamingResultCollectionWithExplicitViewMethod"
// endpoint websocket connection with context.
func (s *BidirectionalStreamingResultCollectionWithExplicitViewMethodClientStream) RecvWithContext(ctx context.Context) (bidirectionalstreamingresultcollectionwithexplicitviewservice.UsertypeCollection, error) {
	return s.Recv()
}
`

var BidirectionalStreamingPrimitiveServerStreamSendCode = `// Send streams instances of "string" to the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMethodServerStream) Send(v string) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	return s.conn.WriteJSON(res)
}

// SendWithContext streams instances of "string" to the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingPrimitiveMethodServerStream) SendWithContext(ctx context.Context, v string) error {
	return s.Send(v)
}
`

var BidirectionalStreamingPrimitiveServerStreamRecvCode = `// Recv reads instances of "string" from the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMethodServerStream) Recv() (string, error) {
	var (
		rv  string
		msg *string
		err error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	return *msg, nil
}

// RecvWithContext reads instances of "string" from the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingPrimitiveMethodServerStream) RecvWithContext(ctx context.Context) (string, error) {
	return s.Recv()
}
`

var BidirectionalStreamingPrimitiveClientStreamSendCode = `// Send streams instances of "string" to the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMethodClientStream) Send(v string) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "string" to the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingPrimitiveMethodClientStream) SendWithContext(ctx context.Context, v string) error {
	return s.Send(v)
}
`

var BidirectionalStreamingPrimitiveClientStreamRecvCode = `// Recv reads instances of "string" from the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMethodClientStream) Recv() (string, error) {
	var (
		rv   string
		body string
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// RecvWithContext reads instances of "string" from the
// "BidirectionalStreamingPrimitiveMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingPrimitiveMethodClientStream) RecvWithContext(ctx context.Context) (string, error) {
	return s.Recv()
}
`

var BidirectionalStreamingPrimitiveArrayServerStreamSendCode = `// Send streams instances of "[]string" to the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveArrayMethodServerStream) Send(v []string) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	return s.conn.WriteJSON(res)
}

// SendWithContext streams instances of "[]string" to the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveArrayMethodServerStream) SendWithContext(ctx context.Context, v []string) error {
	return s.Send(v)
}
`

var BidirectionalStreamingPrimitiveArrayServerStreamRecvCode = `// Recv reads instances of "[]int32" from the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveArrayMethodServerStream) Recv() ([]int32, error) {
	var (
		rv   []int32
		body []int32
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return body, nil
}

// RecvWithContext reads instances of "[]int32" from the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveArrayMethodServerStream) RecvWithContext(ctx context.Context) ([]int32, error) {
	return s.Recv()
}
`

var BidirectionalStreamingPrimitiveArrayClientStreamSendCode = `// Send streams instances of "[]int32" to the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveArrayMethodClientStream) Send(v []int32) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "[]int32" to the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveArrayMethodClientStream) SendWithContext(ctx context.Context, v []int32) error {
	return s.Send(v)
}
`

var BidirectionalStreamingPrimitiveArrayClientStreamRecvCode = `// Recv reads instances of "[]string" from the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveArrayMethodClientStream) Recv() ([]string, error) {
	var (
		rv   []string
		body []string
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// RecvWithContext reads instances of "[]string" from the
// "BidirectionalStreamingPrimitiveArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveArrayMethodClientStream) RecvWithContext(ctx context.Context) ([]string, error) {
	return s.Recv()
}
`

var BidirectionalStreamingPrimitiveMapServerStreamSendCode = `// Send streams instances of "map[int]int" to the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMapMethodServerStream) Send(v map[int]int) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	return s.conn.WriteJSON(res)
}

// SendWithContext streams instances of "map[int]int" to the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveMapMethodServerStream) SendWithContext(ctx context.Context, v map[int]int) error {
	return s.Send(v)
}
`

var BidirectionalStreamingPrimitiveMapServerStreamRecvCode = `// Recv reads instances of "map[string]int32" from the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMapMethodServerStream) Recv() (map[string]int32, error) {
	var (
		rv   map[string]int32
		body map[string]int32
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return body, nil
}

// RecvWithContext reads instances of "map[string]int32" from the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveMapMethodServerStream) RecvWithContext(ctx context.Context) (map[string]int32, error) {
	return s.Recv()
}
`

var BidirectionalStreamingPrimitiveMapClientStreamSendCode = `// Send streams instances of "map[string]int32" to the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMapMethodClientStream) Send(v map[string]int32) error {
	return s.conn.WriteJSON(v)
}

// SendWithContext streams instances of "map[string]int32" to the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveMapMethodClientStream) SendWithContext(ctx context.Context, v map[string]int32) error {
	return s.Send(v)
}
`

var BidirectionalStreamingPrimitiveMapClientStreamRecvCode = `// Recv reads instances of "map[int]int" from the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingPrimitiveMapMethodClientStream) Recv() (map[int]int, error) {
	var (
		rv   map[int]int
		body map[int]int
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	return body, nil
}

// RecvWithContext reads instances of "map[int]int" from the
// "BidirectionalStreamingPrimitiveMapMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingPrimitiveMapMethodClientStream) RecvWithContext(ctx context.Context) (map[int]int, error) {
	return s.Recv()
}
`

var BidirectionalStreamingUserTypeArrayServerStreamSendCode = `// Send streams instances of
// "[]*bidirectionalstreamingusertypearrayservice.ResultType" to the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeArrayMethodServerStream) Send(v []*bidirectionalstreamingusertypearrayservice.ResultType) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	body := NewBidirectionalStreamingUserTypeArrayMethodResponseBody(res)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "[]*bidirectionalstreamingusertypearrayservice.ResultType" to the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingUserTypeArrayMethodServerStream) SendWithContext(ctx context.Context, v []*bidirectionalstreamingusertypearrayservice.ResultType) error {
	return s.Send(v)
}
`

var BidirectionalStreamingUserTypeArrayServerStreamRecvCode = `// Recv reads instances of
// "[]*bidirectionalstreamingusertypearrayservice.RequestType" from the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeArrayMethodServerStream) Recv() ([]*bidirectionalstreamingusertypearrayservice.RequestType, error) {
	var (
		rv   []*bidirectionalstreamingusertypearrayservice.RequestType
		body []*RequestType
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return NewBidirectionalStreamingUserTypeArrayMethodArray(body), nil
}

// RecvWithContext reads instances of
// "[]*bidirectionalstreamingusertypearrayservice.RequestType" from the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingUserTypeArrayMethodServerStream) RecvWithContext(ctx context.Context) ([]*bidirectionalstreamingusertypearrayservice.RequestType, error) {
	return s.Recv()
}
`

var BidirectionalStreamingUserTypeArrayClientStreamSendCode = `// Send streams instances of
// "[]*bidirectionalstreamingusertypearrayservice.RequestType" to the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeArrayMethodClientStream) Send(v []*bidirectionalstreamingusertypearrayservice.RequestType) error {
	body := NewRequestType(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "[]*bidirectionalstreamingusertypearrayservice.RequestType" to the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingUserTypeArrayMethodClientStream) SendWithContext(ctx context.Context, v []*bidirectionalstreamingusertypearrayservice.RequestType) error {
	return s.Send(v)
}
`

var BidirectionalStreamingUserTypeArrayClientStreamRecvCode = `// Recv reads instances of
// "[]*bidirectionalstreamingusertypearrayservice.ResultType" from the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeArrayMethodClientStream) Recv() ([]*bidirectionalstreamingusertypearrayservice.ResultType, error) {
	var (
		rv   []*bidirectionalstreamingusertypearrayservice.ResultType
		body []*ResultTypeResponse
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewBidirectionalStreamingUserTypeArrayMethodResultTypeOK(body)
	return res, nil
}

// RecvWithContext reads instances of
// "[]*bidirectionalstreamingusertypearrayservice.ResultType" from the
// "BidirectionalStreamingUserTypeArrayMethod" endpoint websocket connection
// with context.
func (s *BidirectionalStreamingUserTypeArrayMethodClientStream) RecvWithContext(ctx context.Context) ([]*bidirectionalstreamingusertypearrayservice.ResultType, error) {
	return s.Recv()
}
`

var BidirectionalStreamingUserTypeMapServerStreamSendCode = `// Send streams instances of
// "map[string]*bidirectionalstreamingusertypemapservice.ResultType" to the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeMapMethodServerStream) Send(v map[string]*bidirectionalstreamingusertypemapservice.ResultType) error {
	var err error
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Send().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return s.upgradeErr
	}
	res := v
	body := NewBidirectionalStreamingUserTypeMapMethodResponseBody(res)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "map[string]*bidirectionalstreamingusertypemapservice.ResultType" to the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingUserTypeMapMethodServerStream) SendWithContext(ctx context.Context, v map[string]*bidirectionalstreamingusertypemapservice.ResultType) error {
	return s.Send(v)
}
`

var BidirectionalStreamingUserTypeMapServerStreamRecvCode = `// Recv reads instances of
// "map[string]*bidirectionalstreamingusertypemapservice.RequestType" from the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeMapMethodServerStream) Recv() (map[string]*bidirectionalstreamingusertypemapservice.RequestType, error) {
	var (
		rv   map[string]*bidirectionalstreamingusertypemapservice.RequestType
		body map[string]*RequestType
		err  error
	)
	// Upgrade the HTTP connection to a websocket connection only once. Connection
	// upgrade is done here so that authorization logic in the endpoint is executed
	// before calling the actual service method which may call Recv().
	s.once.Do(func() {
		var conn *websocket.Conn
		conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.upgradeErr = err
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if s.upgradeErr != nil {
		return rv, s.upgradeErr
	}
	if err = s.conn.ReadJSON(&body); err != nil {
		return rv, err
	}
	if body == nil {
		return rv, io.EOF
	}
	return NewBidirectionalStreamingUserTypeMapMethodMap(body), nil
}

// RecvWithContext reads instances of
// "map[string]*bidirectionalstreamingusertypemapservice.RequestType" from the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingUserTypeMapMethodServerStream) RecvWithContext(ctx context.Context) (map[string]*bidirectionalstreamingusertypemapservice.RequestType, error) {
	return s.Recv()
}
`

var BidirectionalStreamingUserTypeMapClientStreamSendCode = `// Send streams instances of
// "map[string]*bidirectionalstreamingusertypemapservice.RequestType" to the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeMapMethodClientStream) Send(v map[string]*bidirectionalstreamingusertypemapservice.RequestType) error {
	body := NewMapStringRequestType(v)
	return s.conn.WriteJSON(body)
}

// SendWithContext streams instances of
// "map[string]*bidirectionalstreamingusertypemapservice.RequestType" to the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingUserTypeMapMethodClientStream) SendWithContext(ctx context.Context, v map[string]*bidirectionalstreamingusertypemapservice.RequestType) error {
	return s.Send(v)
}
`

var BidirectionalStreamingUserTypeMapClientStreamRecvCode = `// Recv reads instances of
// "map[string]*bidirectionalstreamingusertypemapservice.ResultType" from the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection.
func (s *BidirectionalStreamingUserTypeMapMethodClientStream) Recv() (map[string]*bidirectionalstreamingusertypemapservice.ResultType, error) {
	var (
		rv   map[string]*bidirectionalstreamingusertypemapservice.ResultType
		body map[string]*ResultTypeResponse
		err  error
	)
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	res := NewBidirectionalStreamingUserTypeMapMethodMapStringResultTypeOK(body)
	return res, nil
}

// RecvWithContext reads instances of
// "map[string]*bidirectionalstreamingusertypemapservice.ResultType" from the
// "BidirectionalStreamingUserTypeMapMethod" endpoint websocket connection with
// context.
func (s *BidirectionalStreamingUserTypeMapMethodClientStream) RecvWithContext(ctx context.Context) (map[string]*bidirectionalstreamingusertypemapservice.ResultType, error) {
	return s.Recv()
}
`
