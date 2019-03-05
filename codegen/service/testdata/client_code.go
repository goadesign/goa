package testdata

const SingleMethodClient = `// Client is the "SingleEndpoint" service client.
type Client struct {
	AEndpoint goa.Endpoint
}

// NewClient initializes a "SingleEndpoint" service client given the endpoints.
func NewClient(a goa.Endpoint) *Client {
	return &Client{
		AEndpoint: a,
	}
}

// A calls the "A" endpoint of the "SingleEndpoint" service.
func (c *Client) A(ctx context.Context, p *AType) (err error) {
	_, err = c.AEndpoint(ctx, p)
	return
}
`

const MultipleMethodsClient = `// Client is the "MultipleEndpoints" service client.
type Client struct {
	BEndpoint goa.Endpoint
	CEndpoint goa.Endpoint
}

// NewClient initializes a "MultipleEndpoints" service client given the
// endpoints.
func NewClient(b, c goa.Endpoint) *Client {
	return &Client{
		BEndpoint: b,
		CEndpoint: c,
	}
}

// B calls the "B" endpoint of the "MultipleEndpoints" service.
func (c *Client) B(ctx context.Context, p *BType) (err error) {
	_, err = c.BEndpoint(ctx, p)
	return
}

// C calls the "C" endpoint of the "MultipleEndpoints" service.
func (c *Client) C(ctx context.Context, p *CType) (err error) {
	_, err = c.CEndpoint(ctx, p)
	return
}
`

const NoPayloadMethodsClient = `// Client is the "NoPayload" service client.
type Client struct {
	NoPayloadEndpoint goa.Endpoint
}

// NewClient initializes a "NoPayload" service client given the endpoints.
func NewClient(noPayload goa.Endpoint) *Client {
	return &Client{
		NoPayloadEndpoint: noPayload,
	}
}

// NoPayload calls the "NoPayload" endpoint of the "NoPayload" service.
func (c *Client) NoPayload(ctx context.Context) (err error) {
	_, err = c.NoPayloadEndpoint(ctx, nil)
	return
}
`

const StreamingResultMethodClient = `// Client is the "StreamingResultService" service client.
type Client struct {
	StreamingResultMethodEndpoint goa.Endpoint
}

// NewClient initializes a "StreamingResultService" service client given the
// endpoints.
func NewClient(streamingResultMethod goa.Endpoint) *Client {
	return &Client{
		StreamingResultMethodEndpoint: streamingResultMethod,
	}
}

// StreamingResultMethod calls the "StreamingResultMethod" endpoint of the
// "StreamingResultService" service.
func (c *Client) StreamingResultMethod(ctx context.Context, p *APayload) (res StreamingResultMethodClientStream, err error) {
	var ires interface{}
	ires, err = c.StreamingResultMethodEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(StreamingResultMethodClientStream), nil
}
`

const StreamingResultNoPayloadMethodClient = `// Client is the "StreamingResultNoPayloadService" service client.
type Client struct {
	StreamingResultNoPayloadMethodEndpoint goa.Endpoint
}

// NewClient initializes a "StreamingResultNoPayloadService" service client
// given the endpoints.
func NewClient(streamingResultNoPayloadMethod goa.Endpoint) *Client {
	return &Client{
		StreamingResultNoPayloadMethodEndpoint: streamingResultNoPayloadMethod,
	}
}

// StreamingResultNoPayloadMethod calls the "StreamingResultNoPayloadMethod"
// endpoint of the "StreamingResultNoPayloadService" service.
func (c *Client) StreamingResultNoPayloadMethod(ctx context.Context) (res StreamingResultNoPayloadMethodClientStream, err error) {
	var ires interface{}
	ires, err = c.StreamingResultNoPayloadMethodEndpoint(ctx, nil)
	if err != nil {
		return
	}
	return ires.(StreamingResultNoPayloadMethodClientStream), nil
}
`

const StreamingPayloadMethodClient = `// Client is the "StreamingPayloadService" service client.
type Client struct {
	StreamingPayloadMethodEndpoint goa.Endpoint
}

// NewClient initializes a "StreamingPayloadService" service client given the
// endpoints.
func NewClient(streamingPayloadMethod goa.Endpoint) *Client {
	return &Client{
		StreamingPayloadMethodEndpoint: streamingPayloadMethod,
	}
}

// StreamingPayloadMethod calls the "StreamingPayloadMethod" endpoint of the
// "StreamingPayloadService" service.
func (c *Client) StreamingPayloadMethod(ctx context.Context, p *BPayload) (res StreamingPayloadMethodClientStream, err error) {
	var ires interface{}
	ires, err = c.StreamingPayloadMethodEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(StreamingPayloadMethodClientStream), nil
}
`

const StreamingPayloadNoPayloadMethodClient = `// Client is the "StreamingPayloadNoPayloadService" service client.
type Client struct {
	StreamingPayloadNoPayloadMethodEndpoint goa.Endpoint
}

// NewClient initializes a "StreamingPayloadNoPayloadService" service client
// given the endpoints.
func NewClient(streamingPayloadNoPayloadMethod goa.Endpoint) *Client {
	return &Client{
		StreamingPayloadNoPayloadMethodEndpoint: streamingPayloadNoPayloadMethod,
	}
}

// StreamingPayloadNoPayloadMethod calls the "StreamingPayloadNoPayloadMethod"
// endpoint of the "StreamingPayloadNoPayloadService" service.
func (c *Client) StreamingPayloadNoPayloadMethod(ctx context.Context) (res StreamingPayloadNoPayloadMethodClientStream, err error) {
	var ires interface{}
	ires, err = c.StreamingPayloadNoPayloadMethodEndpoint(ctx, nil)
	if err != nil {
		return
	}
	return ires.(StreamingPayloadNoPayloadMethodClientStream), nil
}
`

const BidirectionalStreamingMethodClient = `// Client is the "BidirectionalStreamingService" service client.
type Client struct {
	BidirectionalStreamingMethodEndpoint goa.Endpoint
}

// NewClient initializes a "BidirectionalStreamingService" service client given
// the endpoints.
func NewClient(bidirectionalStreamingMethod goa.Endpoint) *Client {
	return &Client{
		BidirectionalStreamingMethodEndpoint: bidirectionalStreamingMethod,
	}
}

// BidirectionalStreamingMethod calls the "BidirectionalStreamingMethod"
// endpoint of the "BidirectionalStreamingService" service.
func (c *Client) BidirectionalStreamingMethod(ctx context.Context, p *BPayload) (res BidirectionalStreamingMethodClientStream, err error) {
	var ires interface{}
	ires, err = c.BidirectionalStreamingMethodEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(BidirectionalStreamingMethodClientStream), nil
}
`

const BidirectionalStreamingNoPayloadMethodClient = `// Client is the "BidirectionalStreamingNoPayloadService" service client.
type Client struct {
	BidirectionalStreamingNoPayloadMethodEndpoint goa.Endpoint
}

// NewClient initializes a "BidirectionalStreamingNoPayloadService" service
// client given the endpoints.
func NewClient(bidirectionalStreamingNoPayloadMethod goa.Endpoint) *Client {
	return &Client{
		BidirectionalStreamingNoPayloadMethodEndpoint: bidirectionalStreamingNoPayloadMethod,
	}
}

// BidirectionalStreamingNoPayloadMethod calls the
// "BidirectionalStreamingNoPayloadMethod" endpoint of the
// "BidirectionalStreamingNoPayloadService" service.
func (c *Client) BidirectionalStreamingNoPayloadMethod(ctx context.Context) (res BidirectionalStreamingNoPayloadMethodClientStream, err error) {
	var ires interface{}
	ires, err = c.BidirectionalStreamingNoPayloadMethodEndpoint(ctx, nil)
	if err != nil {
		return
	}
	return ires.(BidirectionalStreamingNoPayloadMethodClientStream), nil
}
`
