package testdata

var StreamingResultServerHandlerInitCode = `// NewStreamingResultMethodHandler creates a HTTP handler which loads the HTTP
// request and calls the "StreamingResultService" service
// "StreamingResultMethod" endpoint.
func NewStreamingResultMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeStreamingResultMethodRequest(mux, dec)
		encodeError   = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "StreamingResultMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "StreamingResultService")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		v := &streamingresultservice.StreamingResultMethodEndpointInput{
			Stream: &StreamingResultMethodServerStream{
				upgrader:     up,
				connConfigFn: connConfigFn,
				w:            w,
				r:            r,
			},
			Payload: payload.(*streamingresultservice.Request),
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
`

var StreamingResultServerStreamSendCode = `// Send sends streamingresultservice.UserType type to the
// "StreamingResultMethod" endpoint websocket connection.
func (s *StreamingResultMethodServerStream) Send(v *streamingresultservice.UserType) error {
	// Upgrade the HTTP connection to a websocket connection only once before
	// sending result. Connection upgrade is done here so that authorization logic
	// in the endpoint is executed before calling the actual service method which
	// may call Send().
	s.once.Do(func() {
		conn, err := s.upgrader.Upgrade(s.w, s.r, nil)
		if err != nil {
			s.Lock()
			s.sendErr = err
			s.Unlock()
			return
		}
		if s.connConfigFn != nil {
			conn = s.connConfigFn(conn)
		}
		s.Lock()
		s.conn = conn
		s.Unlock()
	})
	if s.sendErr != nil {
		if s.conn != nil {
			return s.Close()
		}
		return s.sendErr
	}
	res := v
	body := NewStreamingResultMethodResponseBody(res)
	if err := s.conn.WriteJSON(body); err != nil {
		s.Lock()
		s.sendErr = err
		s.Unlock()
		return s.sendErr
	}
	return nil
}
`

var StreamingResultServerStreamCloseCode = `// Close closes the "StreamingResultMethod" endpoint websocket connection after
// sending a close control message.
func (s *StreamingResultMethodServerStream) Close() error {
	if s.conn == nil {
		return nil
	}
	if err := s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "end of message"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	if err := s.conn.Close(); err != nil {
		return err
	}
	s.Lock()
	defer s.Unlock()
	s.conn = nil
	return nil
}
`

var StreamingResultWithViewsServerHandlerInitCode = `// NewStreamingResultWithViewsMethodHandler creates a HTTP handler which loads
// the HTTP request and calls the "StreamingResultWithViewsService" service
// "StreamingResultWithViewsMethod" endpoint.
func NewStreamingResultWithViewsMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeStreamingResultWithViewsMethodRequest(mux, dec)
		encodeError   = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "StreamingResultWithViewsMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "StreamingResultWithViewsService")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		v := &streamingresultwithviewsservice.StreamingResultWithViewsMethodEndpointInput{
			Stream: &StreamingResultWithViewsMethodServerStream{
				upgrader:     up,
				connConfigFn: connConfigFn,
				w:            w,
				r:            r,
			},
			Payload: payload.(*streamingresultwithviewsservice.Request),
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
`

var StreamingResultWithViewsServerStreamSendCode = `// Send sends streamingresultwithviewsservice.Usertype type to the
// "StreamingResultWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultWithViewsMethodServerStream) Send(v *streamingresultwithviewsservice.Usertype) error {
	// Upgrade the HTTP connection to a websocket connection only once before
	// sending result. Connection upgrade is done here so that authorization logic
	// in the endpoint is executed before calling the actual service method which
	// may call Send().
	s.once.Do(func() {
		respHdr := make(http.Header)
		respHdr.Add("goa-view", s.view)
		conn, err := s.upgrader.Upgrade(s.w, s.r, respHdr)
		if err != nil {
			s.Lock()
			s.sendErr = err
			s.Unlock()
			return
		}
		if s.connConfigFn != nil {
			conn = s.connConfigFn(conn)
		}
		s.Lock()
		s.conn = conn
		s.Unlock()
	})
	if s.sendErr != nil {
		if s.conn != nil {
			return s.Close()
		}
		return s.sendErr
	}
	res := streamingresultwithviewsservice.NewViewedUsertype(v, s.view)
	body := NewStreamingResultWithViewsMethodResponseBody(res.Projected)
	if err := s.conn.WriteJSON(body); err != nil {
		s.Lock()
		s.sendErr = err
		s.Unlock()
		return s.sendErr
	}
	return nil
}
`

var StreamingResultWithViewsServerStreamSetViewCode = `// SetView sets the view to render the streamingresultwithviewsservice.Usertype
// type before sending to the "StreamingResultWithViewsMethod" endpoint
// websocket connection.
func (s *StreamingResultWithViewsMethodServerStream) SetView(view string) {
	s.Lock()
	defer s.Unlock()
	s.view = view
}
`

var StreamingResultNoPayloadServerHandlerInitCode = `// NewStreamingResultNoPayloadMethodHandler creates a HTTP handler which loads
// the HTTP request and calls the "StreamingResultNoPayloadService" service
// "StreamingResultNoPayloadMethod" endpoint.
func NewStreamingResultNoPayloadMethodHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		encodeError = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "StreamingResultNoPayloadMethod")
		ctx = context.WithValue(ctx, goa.ServiceKey, "StreamingResultNoPayloadService")

		v := &streamingresultnopayloadservice.StreamingResultNoPayloadMethodEndpointInput{
			Stream: &StreamingResultNoPayloadMethodServerStream{
				upgrader:     up,
				connConfigFn: connConfigFn,
				w:            w,
				r:            r,
			},
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
`

var StreamingResultClientEndpointCode = `// StreamingResultMethod returns an endpoint that makes HTTP requests to the
// StreamingResultService service StreamingResultMethod server.
func (c *Client) StreamingResultMethod() goa.Endpoint {
	var (
		encodeRequest  = EncodeStreamingResultMethodRequest(c.encoder)
		decodeResponse = DecodeStreamingResultMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildStreamingResultMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.Dial(req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingResultService", "StreamingResultMethod", err)
		}
		if c.connConfigFn != nil {
			conn = c.connConfigFn(conn)
		}
		stream := &StreamingResultMethodClientStream{conn: conn}
		return stream, nil
	}
}
`

var StreamingResultWithViewsServerStreamCloseCode = `// Close closes the "StreamingResultWithViewsMethod" endpoint websocket
// connection after sending a close control message.
func (s *StreamingResultWithViewsMethodServerStream) Close() error {
	if s.conn == nil {
		return nil
	}
	if err := s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "end of message"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	if err := s.conn.Close(); err != nil {
		return err
	}
	s.Lock()
	defer s.Unlock()
	s.conn = nil
	return nil
}
`

var StreamingResultClientStreamRecvCode = `// Recv receives a streamingresultservice.UserType type from the
// "StreamingResultMethod" endpoint websocket connection.
func (s *StreamingResultMethodClientStream) Recv() (*streamingresultservice.UserType, error) {
	if s.conn == nil {
		return nil, nil
	}
	var body StreamingResultMethodResponseBody
	err := s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, goahttp.NormalSocketCloseErrors...) {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}
	res := NewStreamingResultMethodUserTypeOK(&body)
	return res, nil
}
`

var StreamingResultWithViewsClientEndpointCode = `// StreamingResultWithViewsMethod returns an endpoint that makes HTTP requests
// to the StreamingResultWithViewsService service
// StreamingResultWithViewsMethod server.
func (c *Client) StreamingResultWithViewsMethod() goa.Endpoint {
	var (
		encodeRequest  = EncodeStreamingResultWithViewsMethodRequest(c.encoder)
		decodeResponse = DecodeStreamingResultWithViewsMethodResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildStreamingResultWithViewsMethodRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		conn, resp, err := c.dialer.Dial(req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("StreamingResultWithViewsService", "StreamingResultWithViewsMethod", err)
		}
		if c.connConfigFn != nil {
			conn = c.connConfigFn(conn)
		}
		stream := &StreamingResultWithViewsMethodClientStream{conn: conn}
		view := resp.Header.Get("goa-view")
		stream.SetView(view)
		return stream, nil
	}
}
`

var StreamingResultWithViewsClientStreamRecvCode = `// Recv receives a streamingresultwithviewsservice.Usertype type from the
// "StreamingResultWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultWithViewsMethodClientStream) Recv() (*streamingresultwithviewsservice.Usertype, error) {
	if s.conn == nil {
		return nil, nil
	}
	var body StreamingResultWithViewsMethodResponseBody
	err := s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, goahttp.NormalSocketCloseErrors...) {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}
	res := NewStreamingResultWithViewsMethodUsertypeOK(&body)
	vres := &streamingresultwithviewsserviceviews.Usertype{res, s.view}
	if err := vres.Validate(); err != nil {
		return nil, goahttp.ErrValidationError("StreamingResultWithViewsService", "StreamingResultWithViewsMethod", err)
	}
	return streamingresultwithviewsservice.NewUsertype(vres), nil
}
`

var StreamingResultWithViewsClientStreamSetViewCode = `// SetView sets the view to render the  type before sending to the
// "StreamingResultWithViewsMethod" endpoint websocket connection.
func (s *StreamingResultWithViewsMethodClientStream) SetView(view string) {
	s.Lock()
	defer s.Unlock()
	s.view = view
}
`
