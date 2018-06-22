package testdata

const SingleEndpoint = `// Endpoints wraps the "SingleEndpoint" service endpoints.
type Endpoints struct {
	A goa.Endpoint
}

// NewEndpoints wraps the methods of the "SingleEndpoint" service with
// endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		A: NewAEndpoint(s),
	}
}

// Use applies the given middleware to all the "SingleEndpoint" service
// endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.A = m(e.A)
}

// NewAEndpoint returns an endpoint function that calls the method "A" of
// service "SingleEndpoint".
func NewAEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*AType)
		return nil, s.A(ctx, p)
	}
}
`

const MultipleEndpoints = `// Endpoints wraps the "MultipleEndpoints" service endpoints.
type Endpoints struct {
	B goa.Endpoint
	C goa.Endpoint
}

// NewEndpoints wraps the methods of the "MultipleEndpoints" service with
// endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		B: NewBEndpoint(s),
		C: NewCEndpoint(s),
	}
}

// Use applies the given middleware to all the "MultipleEndpoints" service
// endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.B = m(e.B)
	e.C = m(e.C)
}

// NewBEndpoint returns an endpoint function that calls the method "B" of
// service "MultipleEndpoints".
func NewBEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*BType)
		return nil, s.B(ctx, p)
	}
}

// NewCEndpoint returns an endpoint function that calls the method "C" of
// service "MultipleEndpoints".
func NewCEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*CType)
		return nil, s.C(ctx, p)
	}
}
`

const NoPayloadEndpoint = `// Endpoints wraps the "NoPayload" service endpoints.
type Endpoints struct {
	NoPayload goa.Endpoint
}

// NewEndpoints wraps the methods of the "NoPayload" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		NoPayload: NewNoPayloadEndpoint(s),
	}
}

// Use applies the given middleware to all the "NoPayload" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.NoPayload = m(e.NoPayload)
}

// NewNoPayloadEndpoint returns an endpoint function that calls the method
// "NoPayload" of service "NoPayload".
func NewNoPayloadEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, s.NoPayload(ctx)
	}
}
`

const WithResultEndpoint = `// Endpoints wraps the "WithResult" service endpoints.
type Endpoints struct {
	A goa.Endpoint
}

// NewEndpoints wraps the methods of the "WithResult" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		A: NewAEndpoint(s),
	}
}

// Use applies the given middleware to all the "WithResult" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.A = m(e.A)
}

// NewAEndpoint returns an endpoint function that calls the method "A" of
// service "WithResult".
func NewAEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		res, err := s.A(ctx)
		if err != nil {
			return nil, err
		}
		vres := NewViewedRtype(res, "default")
		return vres, nil
	}
}
`

const WithResultMultipleViewsEndpoint = `// Endpoints wraps the "WithResultMultipleViews" service endpoints.
type Endpoints struct {
	A goa.Endpoint
}

// NewEndpoints wraps the methods of the "WithResultMultipleViews" service with
// endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		A: NewAEndpoint(s),
	}
}

// Use applies the given middleware to all the "WithResultMultipleViews"
// service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.A = m(e.A)
}

// NewAEndpoint returns an endpoint function that calls the method "A" of
// service "WithResultMultipleViews".
func NewAEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		res, view, err := s.A(ctx)
		if err != nil {
			return nil, err
		}
		vres := NewViewedViewtype(res, view)
		return vres, nil
	}
}
`

const StreamingResultMethodEndpoint = `// Endpoints wraps the "StreamingResultEndpoint" service endpoints.
type Endpoints struct {
	StreamingResultMethod goa.Endpoint
}

// StreamingResultMethodEndpointInput is the input type of
// "StreamingResultMethod" endpoint that holds the method payload and the
// server stream.
type StreamingResultMethodEndpointInput struct {
	// Payload is the method payload.
	Payload *AType
	// Stream is the server stream used by the "StreamingResultMethod" method to
	// send data.
	Stream StreamingResultMethodServerStream
}

// NewEndpoints wraps the methods of the "StreamingResultEndpoint" service with
// endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		StreamingResultMethod: NewStreamingResultMethodEndpoint(s),
	}
}

// Use applies the given middleware to all the "StreamingResultEndpoint"
// service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.StreamingResultMethod = m(e.StreamingResultMethod)
}

// NewStreamingResultMethodEndpoint returns an endpoint function that calls the
// method "StreamingResultMethod" of service "StreamingResultEndpoint".
func NewStreamingResultMethodEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		ep := req.(*StreamingResultMethodEndpointInput)
		p := ep.Payload
		return nil, s.StreamingResultMethod(ctx, ep.Payload, ep.Stream)
	}
}
`

const StreamingResultNoPayloadMethodEndpoint = `// Endpoints wraps the "StreamingResultNoPayloadEndpoint" service endpoints.
type Endpoints struct {
	StreamingResultNoPayloadMethod goa.Endpoint
}

// StreamingResultNoPayloadMethodEndpointInput is the input type of
// "StreamingResultNoPayloadMethod" endpoint that holds the method payload and
// the server stream.
type StreamingResultNoPayloadMethodEndpointInput struct {
	// Stream is the server stream used by the "StreamingResultNoPayloadMethod"
	// method to send data.
	Stream StreamingResultNoPayloadMethodServerStream
}

// NewEndpoints wraps the methods of the "StreamingResultNoPayloadEndpoint"
// service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		StreamingResultNoPayloadMethod: NewStreamingResultNoPayloadMethodEndpoint(s),
	}
}

// Use applies the given middleware to all the
// "StreamingResultNoPayloadEndpoint" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.StreamingResultNoPayloadMethod = m(e.StreamingResultNoPayloadMethod)
}

// NewStreamingResultNoPayloadMethodEndpoint returns an endpoint function that
// calls the method "StreamingResultNoPayloadMethod" of service
// "StreamingResultNoPayloadEndpoint".
func NewStreamingResultNoPayloadMethodEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		ep := req.(*StreamingResultNoPayloadMethodEndpointInput)
		return nil, s.StreamingResultNoPayloadMethod(ctx, ep.Stream)
	}
}
`
