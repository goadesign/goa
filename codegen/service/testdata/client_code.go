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
func (c *Client) A(ctx context.Context, p *AType)(err error) {
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
func (c *Client) B(ctx context.Context, p *BType)(err error) {
	_, err = c.BEndpoint(ctx, p)
	return
}

// C calls the "C" endpoint of the "MultipleEndpoints" service.
func (c *Client) C(ctx context.Context, p *CType)(err error) {
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
func (c *Client) NoPayload(ctx context.Context, )(err error) {
	_, err = c.NoPayloadEndpoint(ctx, nil)
	return
}
`
