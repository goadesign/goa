package service

import (
	"bytes"
	"strings"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

func TestClient(t *testing.T) {
	const (
		singleMethod = `// Client is the "Single" service client.
type Client struct {
	AEndpoint goa.Endpoint
}
// NewClient initializes a "Single" service client given the endpoints.
func NewClient(a goa.Endpoint) *Client {
	return &Client{
		AEndpoint: a,
	}
}

// A calls the "A" endpoint of the "Single" service.
func (c *Client) A(ctx context.Context, p *AType)(err error) {
	_, err = c.AEndpoint(ctx, p)
	return
}
`

		multipleMethods = `// Client is the "Multiple" service client.
type Client struct {
	BEndpoint goa.Endpoint
	CEndpoint goa.Endpoint
}
// NewClient initializes a "Multiple" service client given the endpoints.
func NewClient(b, c goa.Endpoint) *Client {
	return &Client{
		BEndpoint: b,
		CEndpoint: c,
	}
}

// B calls the "B" endpoint of the "Multiple" service.
func (c *Client) B(ctx context.Context, p *BType)(err error) {
	_, err = c.BEndpoint(ctx, p)
	return
}

// C calls the "C" endpoint of the "Multiple" service.
func (c *Client) C(ctx context.Context, p *CType)(err error) {
	_, err = c.CEndpoint(ctx, p)
	return
}
`

		nopayloadMethods = `// Client is the "NoPayload" service client.
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
	)
	cases := map[string]struct {
		Service  *design.ServiceExpr
		Expected string
	}{
		"single":    {Service: &singleEndpoint, Expected: singleMethod},
		"multiple":  {Service: &multipleEndpoints, Expected: multipleMethods},
		"nopayload": {Service: &nopayloadEndpoint, Expected: nopayloadMethods},
	}
	for k, tc := range cases {
		buf := new(bytes.Buffer)
		Services = make(ServicesData)
		design.Root.Services = []*design.ServiceExpr{tc.Service}
		design.Root.API = &design.APIExpr{Name: "test"}
		File("", tc.Service) // to initialize ServiceScope
		cf := ClientFile(tc.Service)
		for _, s := range cf.SectionTemplates {
			if err := s.Write(buf); err != nil {
				t.Fatal(err)
			}
		}
		actual := buf.String()
		if !strings.Contains(actual, tc.Expected) {
			d := codegen.Diff(t, actual, tc.Expected)
			t.Errorf("%s: got\n%v\n=============\nexpected to contain\n%v\ndiff\n%v", k, actual, tc.Expected, d)
		}
	}
}
