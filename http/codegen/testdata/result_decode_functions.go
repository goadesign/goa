package testdata

var EmptyServerResponseDecodeCode = `// DecodeMethodEmptyServerResponseResponse returns a decoder for responses
// returned by the ServiceEmptyServerResponse MethodEmptyServerResponse
// endpoint. restoreBody controls whether the response body should be restored
// after having been read.
func DecodeMethodEmptyServerResponseResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			return NewMethodEmptyServerResponseMethodEmptyServerResponseResultOK(), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceEmptyServerResponse", "MethodEmptyServerResponse", resp.StatusCode, string(body))
		}
	}
}
`

var ResultBodyMultipleViewsDecodeCode = `// DecodeMethodBodyMultipleViewResponse returns a decoder for responses
// returned by the ServiceBodyMultipleView MethodBodyMultipleView endpoint.
// restoreBody controls whether the response body should be restored after
// having been read.
func DecodeMethodBodyMultipleViewResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body MethodBodyMultipleViewResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("ServiceBodyMultipleView", "MethodBodyMultipleView", err)
			}
			var (
				c *string
			)
			cRaw := resp.Header.Get("Location")
			if cRaw != "" {
				c = &cRaw
			}
			p := NewMethodBodyMultipleViewResulttypemultipleviewsOK(&body, c)
			view := resp.Header.Get("goa-view")
			vres := &servicebodymultipleviewviews.Resulttypemultipleviews{p, view}
			if err = vres.Validate(); err != nil {
				return nil, goahttp.ErrValidationError("ServiceBodyMultipleView", "MethodBodyMultipleView", err)
			}
			return servicebodymultipleview.NewResulttypemultipleviews(vres), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceBodyMultipleView", "MethodBodyMultipleView", resp.StatusCode, string(body))
		}
	}
}
`

var EmptyBodyResultMultipleViewsDecodeCode = `// DecodeMethodEmptyBodyResultMultipleViewResponse returns a decoder for
// responses returned by the ServiceEmptyBodyResultMultipleView
// MethodEmptyBodyResultMultipleView endpoint. restoreBody controls whether the
// response body should be restored after having been read.
func DecodeMethodEmptyBodyResultMultipleViewResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				c *string
			)
			cRaw := resp.Header.Get("Location")
			if cRaw != "" {
				c = &cRaw
			}
			p := NewMethodEmptyBodyResultMultipleViewResulttypemultipleviewsOK(c)
			view := resp.Header.Get("goa-view")
			vres := &serviceemptybodyresultmultipleviewviews.Resulttypemultipleviews{p, view}
			return serviceemptybodyresultmultipleview.NewResulttypemultipleviews(vres), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceEmptyBodyResultMultipleView", "MethodEmptyBodyResultMultipleView", resp.StatusCode, string(body))
		}
	}
}
`

var ExplicitBodyUserResultMultipleViewsDecodeCode = `// DecodeMethodExplicitBodyUserResultMultipleViewResponse returns a decoder for
// responses returned by the ServiceExplicitBodyUserResultMultipleView
// MethodExplicitBodyUserResultMultipleView endpoint. restoreBody controls
// whether the response body should be restored after having been read.
func DecodeMethodExplicitBodyUserResultMultipleViewResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body UserType
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("ServiceExplicitBodyUserResultMultipleView", "MethodExplicitBodyUserResultMultipleView", err)
			}
			var (
				c *string
			)
			cRaw := resp.Header.Get("Location")
			if cRaw != "" {
				c = &cRaw
			}
			p := NewMethodExplicitBodyUserResultMultipleViewResulttypemultipleviewsOK(&body, c)
			view := resp.Header.Get("goa-view")
			vres := &serviceexplicitbodyuserresultmultipleviewviews.Resulttypemultipleviews{p, view}
			if err = vres.Validate(); err != nil {
				return nil, goahttp.ErrValidationError("ServiceExplicitBodyUserResultMultipleView", "MethodExplicitBodyUserResultMultipleView", err)
			}
			return serviceexplicitbodyuserresultmultipleview.NewResulttypemultipleviews(vres), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceExplicitBodyUserResultMultipleView", "MethodExplicitBodyUserResultMultipleView", resp.StatusCode, string(body))
		}
	}
}
`

var ResultMultipleViewsTagDecodeCode = `// DecodeMethodTagMultipleViewsResponse returns a decoder for responses
// returned by the ServiceTagMultipleViews MethodTagMultipleViews endpoint.
// restoreBody controls whether the response body should be restored after
// having been read.
func DecodeMethodTagMultipleViewsResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			var (
				body MethodTagMultipleViewsAcceptedResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("ServiceTagMultipleViews", "MethodTagMultipleViews", err)
			}
			var (
				c *string
			)
			cRaw := resp.Header.Get("c")
			if cRaw != "" {
				c = &cRaw
			}
			p := NewMethodTagMultipleViewsResulttypemultipleviewsAccepted(&body, c)
			view := resp.Header.Get("goa-view")
			vres := &servicetagmultipleviewsviews.Resulttypemultipleviews{p, view}
			if err = vres.Validate(); err != nil {
				return nil, goahttp.ErrValidationError("ServiceTagMultipleViews", "MethodTagMultipleViews", err)
			}
			return servicetagmultipleviews.NewResulttypemultipleviews(vres), nil
		case http.StatusOK:
			var (
				body MethodTagMultipleViewsOKResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("ServiceTagMultipleViews", "MethodTagMultipleViews", err)
			}
			p := NewMethodTagMultipleViewsResulttypemultipleviewsOK(&body)
			view := resp.Header.Get("goa-view")
			vres := &servicetagmultipleviewsviews.Resulttypemultipleviews{p, view}
			if err = vres.Validate(); err != nil {
				return nil, goahttp.ErrValidationError("ServiceTagMultipleViews", "MethodTagMultipleViews", err)
			}
			return servicetagmultipleviews.NewResulttypemultipleviews(vres), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceTagMultipleViews", "MethodTagMultipleViews", resp.StatusCode, string(body))
		}
	}
}
`
