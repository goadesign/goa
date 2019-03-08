package testdata

const (
	MultipleEndpointsClientInitCode = `// NewClient instantiates HTTP clients for all the ServiceMultiEndpoints
// service servers.
func NewClient(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
) *Client {
	return &Client{
		MethodMultiEndpoints1Doer: doer,
		MethodMultiEndpoints2Doer: doer,
		RestoreResponseBody:       restoreBody,
		scheme:                    scheme,
		host:                      host,
		decoder:                   dec,
		encoder:                   enc,
	}
}
`

	StreamingClientInitCode = `// NewClient instantiates HTTP clients for all the StreamingResultService
// service servers.
func NewClient(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
	dialer goahttp.Dialer,
	cfn *ConnConfigurer,
) *Client {
	if cfn == nil {
		cfn = &ConnConfigurer{}
	}
	return &Client{
		StreamingResultMethodDoer: doer,
		RestoreResponseBody:       restoreBody,
		scheme:                    scheme,
		host:                      host,
		decoder:                   dec,
		encoder:                   enc,
		dialer:                    dialer,
		configurer:                cfn,
	}
}
`
)
