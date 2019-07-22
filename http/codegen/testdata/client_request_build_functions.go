package testdata

const PathStringRequestBuildCode = `// BuildMethodPathStringRequest instantiates a HTTP request object with method
// and path set to call the "ServicePathString" service "MethodPathString"
// endpoint
func (c *Client) BuildMethodPathStringRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		p string
	)
	{
		p, ok := v.(*servicepathstring.MethodPathStringPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("ServicePathString", "MethodPathString", "*servicepathstring.MethodPathStringPayload", v)
		}
		if p.P != nil {
			p = *p.P
		}
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: MethodPathStringServicePathStringPath(p)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("ServicePathString", "MethodPathString", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}
`

const PathStringRequiredRequestBuildCode = `// BuildMethodPathStringValidateRequest instantiates a HTTP request object with
// method and path set to call the "ServicePathStringValidate" service
// "MethodPathStringValidate" endpoint
func (c *Client) BuildMethodPathStringValidateRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		p string
	)
	{
		p, ok := v.(*servicepathstringvalidate.MethodPathStringValidatePayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("ServicePathStringValidate", "MethodPathStringValidate", "*servicepathstringvalidate.MethodPathStringValidatePayload", v)
		}
		p = p.P
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: MethodPathStringValidateServicePathStringValidatePath(p)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("ServicePathStringValidate", "MethodPathStringValidate", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}
`

const PathStringDefaultRequestBuildCode = `// BuildMethodPathStringDefaultRequest instantiates a HTTP request object with
// method and path set to call the "ServicePathStringDefault" service
// "MethodPathStringDefault" endpoint
func (c *Client) BuildMethodPathStringDefaultRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		p string
	)
	{
		p, ok := v.(*servicepathstringdefault.MethodPathStringDefaultPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("ServicePathStringDefault", "MethodPathStringDefault", "*servicepathstringdefault.MethodPathStringDefaultPayload", v)
		}
		p = p.P
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: MethodPathStringDefaultServicePathStringDefaultPath(p)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("ServicePathStringDefault", "MethodPathStringDefault", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}
`
