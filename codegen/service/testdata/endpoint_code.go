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
		return s.A(ctx)
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
		vres := NewViewedViewtype(res)
		switch view {
		case "default":
			vres = vres.AsDefault()
		case "tiny":
			vres = vres.AsTiny()
		default:
			return nil, fmt.Errorf("unknown view %s", view)
		}
		if err := vres.Validate(); err != nil {
			return nil, err
		}
		return vres, nil
	}
}
`
