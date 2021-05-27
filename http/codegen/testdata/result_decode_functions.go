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
			res := NewMethodEmptyServerResponseResultOK()
			return res, nil
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
			vres := &servicebodymultipleviewviews.Resulttypemultipleviews{Projected: p, View: view}
			if err = servicebodymultipleviewviews.ValidateResulttypemultipleviews(vres); err != nil {
				return nil, goahttp.ErrValidationError("ServiceBodyMultipleView", "MethodBodyMultipleView", err)
			}
			res := servicebodymultipleview.NewResulttypemultipleviews(vres)
			return res, nil
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
			vres := &serviceemptybodyresultmultipleviewviews.Resulttypemultipleviews{Projected: p, View: view}
			res := serviceemptybodyresultmultipleview.NewResulttypemultipleviews(vres)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceEmptyBodyResultMultipleView", "MethodEmptyBodyResultMultipleView", resp.StatusCode, string(body))
		}
	}
}
`

var ExplicitBodyPrimitiveResultDecodeCode = `// DecodeMethodExplicitBodyPrimitiveResultMultipleViewResponse returns a
// decoder for responses returned by the
// ServiceExplicitBodyPrimitiveResultMultipleView
// MethodExplicitBodyPrimitiveResultMultipleView endpoint. restoreBody controls
// whether the response body should be restored after having been read.
func DecodeMethodExplicitBodyPrimitiveResultMultipleViewResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("ServiceExplicitBodyPrimitiveResultMultipleView", "MethodExplicitBodyPrimitiveResultMultipleView", err)
			}
			if utf8.RuneCountInString(body) < 5 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("body", body, utf8.RuneCountInString(body), 5, true))
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceExplicitBodyPrimitiveResultMultipleView", "MethodExplicitBodyPrimitiveResultMultipleView", err)
			}
			var (
				c *string
			)
			cRaw := resp.Header.Get("Location")
			if cRaw != "" {
				c = &cRaw
			}
			p := NewMethodExplicitBodyPrimitiveResultMultipleViewResulttypemultipleviewsOK(body, c)
			view := resp.Header.Get("goa-view")
			vres := &serviceexplicitbodyprimitiveresultmultipleviewviews.Resulttypemultipleviews{Projected: p, View: view}
			if err = serviceexplicitbodyprimitiveresultmultipleviewviews.ValidateResulttypemultipleviews(vres); err != nil {
				return nil, goahttp.ErrValidationError("ServiceExplicitBodyPrimitiveResultMultipleView", "MethodExplicitBodyPrimitiveResultMultipleView", err)
			}
			res := serviceexplicitbodyprimitiveresultmultipleview.NewResulttypemultipleviews(vres)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceExplicitBodyPrimitiveResultMultipleView", "MethodExplicitBodyPrimitiveResultMultipleView", resp.StatusCode, string(body))
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
				body MethodExplicitBodyUserResultMultipleViewResponseBody
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
			vres := &serviceexplicitbodyuserresultmultipleviewviews.Resulttypemultipleviews{Projected: p, View: view}
			if err = serviceexplicitbodyuserresultmultipleviewviews.ValidateResulttypemultipleviews(vres); err != nil {
				return nil, goahttp.ErrValidationError("ServiceExplicitBodyUserResultMultipleView", "MethodExplicitBodyUserResultMultipleView", err)
			}
			res := serviceexplicitbodyuserresultmultipleview.NewResulttypemultipleviews(vres)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceExplicitBodyUserResultMultipleView", "MethodExplicitBodyUserResultMultipleView", resp.StatusCode, string(body))
		}
	}
}
`

var ExplicitBodyResultCollectionDecodeCode = `// DecodeMethodExplicitBodyResultCollectionResponse returns a decoder for
// responses returned by the ServiceExplicitBodyResultCollection
// MethodExplicitBodyResultCollection endpoint. restoreBody controls whether
// the response body should be restored after having been read.
func DecodeMethodExplicitBodyResultCollectionResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				body ResulttypeCollection
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("ServiceExplicitBodyResultCollection", "MethodExplicitBodyResultCollection", err)
			}
			err = ValidateResulttypeCollection(body)
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceExplicitBodyResultCollection", "MethodExplicitBodyResultCollection", err)
			}
			res := NewMethodExplicitBodyResultCollectionResultOK(body)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceExplicitBodyResultCollection", "MethodExplicitBodyResultCollection", resp.StatusCode, string(body))
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
			cRaw := resp.Header.Get("C")
			if cRaw != "" {
				c = &cRaw
			}
			p := NewMethodTagMultipleViewsResulttypemultipleviewsAccepted(&body, c)
			tmp := "value"
			p.B = &tmp
			view := resp.Header.Get("goa-view")
			vres := &servicetagmultipleviewsviews.Resulttypemultipleviews{Projected: p, View: view}
			if err = servicetagmultipleviewsviews.ValidateResulttypemultipleviews(vres); err != nil {
				return nil, goahttp.ErrValidationError("ServiceTagMultipleViews", "MethodTagMultipleViews", err)
			}
			res := servicetagmultipleviews.NewResulttypemultipleviews(vres)
			return res, nil
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
			vres := &servicetagmultipleviewsviews.Resulttypemultipleviews{Projected: p, View: view}
			if err = servicetagmultipleviewsviews.ValidateResulttypemultipleviews(vres); err != nil {
				return nil, goahttp.ErrValidationError("ServiceTagMultipleViews", "MethodTagMultipleViews", err)
			}
			res := servicetagmultipleviews.NewResulttypemultipleviews(vres)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceTagMultipleViews", "MethodTagMultipleViews", resp.StatusCode, string(body))
		}
	}
}
`

var EmptyServerResponseWithTagsDecodeCode = `// DecodeMethodEmptyServerResponseWithTagsResponse returns a decoder for
// responses returned by the ServiceEmptyServerResponseWithTags
// MethodEmptyServerResponseWithTags endpoint. restoreBody controls whether the
// response body should be restored after having been read.
func DecodeMethodEmptyServerResponseWithTagsResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
		case http.StatusNotModified:
			res := NewMethodEmptyServerResponseWithTagsResultNotModified()
			res.H = "true"
			return res, nil
		case http.StatusNoContent:
			res := NewMethodEmptyServerResponseWithTagsResultNoContent()
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceEmptyServerResponseWithTags", "MethodEmptyServerResponseWithTags", resp.StatusCode, string(body))
		}
	}
}
`

var ResultHeaderStringImplicitResponseDecodeCode = `// DecodeMethodHeaderStringImplicitResponse returns a decoder for responses
// returned by the ServiceHeaderStringImplicit MethodHeaderStringImplicit
// endpoint. restoreBody controls whether the response body should be restored
// after having been read.
func DecodeMethodHeaderStringImplicitResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				h   string
				err error
			)
			hRaw := resp.Header.Get("H")
			if hRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("h", "header"))
			}
			h = hRaw
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceHeaderStringImplicit", "MethodHeaderStringImplicit", err)
			}
			return h, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceHeaderStringImplicit", "MethodHeaderStringImplicit", resp.StatusCode, string(body))
		}
	}
}
`

var ResultHeaderStringArrayResponseDecodeCode = `// DecodeMethodAResponse returns a decoder for responses returned by the
// ServiceHeaderStringArrayResponse MethodA endpoint. restoreBody controls
// whether the response body should be restored after having been read.
func DecodeMethodAResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				array []string
			)
			array = resp.Header["Array"]

			res := NewMethodAResultOK(array)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceHeaderStringArrayResponse", "MethodA", resp.StatusCode, string(body))
		}
	}
}
`

var ResultHeaderStringArrayValidateResponseDecodeCode = `// DecodeMethodAResponse returns a decoder for responses returned by the
// ServiceHeaderStringArrayValidateResponse MethodA endpoint. restoreBody
// controls whether the response body should be restored after having been read.
func DecodeMethodAResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				array []string
				err   error
			)
			array = resp.Header["Array"]

			if len(array) < 5 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("array", array, len(array), 5, true))
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceHeaderStringArrayValidateResponse", "MethodA", err)
			}
			res := NewMethodAResultOK(array)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceHeaderStringArrayValidateResponse", "MethodA", resp.StatusCode, string(body))
		}
	}
}
`

var ResultHeaderArrayResponseDecodeCode = `// DecodeMethodAResponse returns a decoder for responses returned by the
// ServiceHeaderArrayResponse MethodA endpoint. restoreBody controls whether
// the response body should be restored after having been read.
func DecodeMethodAResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				array []uint
				err   error
			)
			{
				arrayRaw := resp.Header["Array"]

				if arrayRaw != nil {
					array = make([]uint, len(arrayRaw))
					for i, rv := range arrayRaw {
						v, err2 := strconv.ParseUint(rv, 10, strconv.IntSize)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("array", arrayRaw, "array of unsigned integers"))
						}
						array[i] = uint(v)
					}
				}
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceHeaderArrayResponse", "MethodA", err)
			}
			res := NewMethodAResultOK(array)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceHeaderArrayResponse", "MethodA", resp.StatusCode, string(body))
		}
	}
}
`

var ResultHeaderArrayValidateResponseDecodeCode = `// DecodeMethodAResponse returns a decoder for responses returned by the
// ServiceHeaderArrayValidateResponse MethodA endpoint. restoreBody controls
// whether the response body should be restored after having been read.
func DecodeMethodAResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				array []int
				err   error
			)
			{
				arrayRaw := resp.Header["Array"]

				if arrayRaw != nil {
					array = make([]int, len(arrayRaw))
					for i, rv := range arrayRaw {
						v, err2 := strconv.ParseInt(rv, 10, strconv.IntSize)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("array", arrayRaw, "array of integers"))
						}
						array[i] = int(v)
					}
				}
			}
			for _, e := range array {
				if e < 5 {
					err = goa.MergeErrors(err, goa.InvalidRangeError("array[*]", e, 5, true))
				}
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceHeaderArrayValidateResponse", "MethodA", err)
			}
			res := NewMethodAResultOK(array)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceHeaderArrayValidateResponse", "MethodA", resp.StatusCode, string(body))
		}
	}
}
`

var WithHeadersBlockResponseDecodeCode = `// DecodeMethodAResponse returns a decoder for responses returned by the
// ServiceWithHeadersBlock MethodA endpoint. restoreBody controls whether the
// response body should be restored after having been read.
func DecodeMethodAResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				required            int
				optional            *float32
				optionalButRequired uint
				err                 error
			)
			{
				requiredRaw := resp.Header.Get("X-Request-Id")
				if requiredRaw == "" {
					return nil, goahttp.ErrValidationError("ServiceWithHeadersBlock", "MethodA", goa.MissingFieldError("X-Request-ID", "header"))
				}
				v, err2 := strconv.ParseInt(requiredRaw, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("required", requiredRaw, "integer"))
				}
				required = int(v)
			}
			{
				optionalRaw := resp.Header.Get("Authorization")
				if optionalRaw != "" {
					v, err2 := strconv.ParseFloat(optionalRaw, 32)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("optional", optionalRaw, "float"))
					}
					pv := float32(v)
					optional = &pv
				}
			}
			{
				optionalButRequiredRaw := resp.Header.Get("Location")
				if optionalButRequiredRaw == "" {
					return nil, goahttp.ErrValidationError("ServiceWithHeadersBlock", "MethodA", goa.MissingFieldError("Location", "header"))
				}
				v, err2 := strconv.ParseUint(optionalButRequiredRaw, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("optionalButRequired", optionalButRequiredRaw, "unsigned integer"))
				}
				optionalButRequired = uint(v)
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceWithHeadersBlock", "MethodA", err)
			}
			res := NewMethodAResultOK(required, optional, optionalButRequired)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceWithHeadersBlock", "MethodA", resp.StatusCode, string(body))
		}
	}
}
`

var WithHeadersBlockViewedResultResponseDecodeCode = `// DecodeMethodAResponse returns a decoder for responses returned by the
// ServiceWithHeadersBlockViewedResult MethodA endpoint. restoreBody controls
// whether the response body should be restored after having been read.
func DecodeMethodAResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				required            int
				optional            *float32
				optionalButRequired uint
				err                 error
			)
			{
				requiredRaw := resp.Header.Get("X-Request-Id")
				if requiredRaw == "" {
					return nil, goahttp.ErrValidationError("ServiceWithHeadersBlockViewedResult", "MethodA", goa.MissingFieldError("X-Request-ID", "header"))
				}
				v, err2 := strconv.ParseInt(requiredRaw, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("required", requiredRaw, "integer"))
				}
				required = int(v)
			}
			{
				optionalRaw := resp.Header.Get("Authorization")
				if optionalRaw != "" {
					v, err2 := strconv.ParseFloat(optionalRaw, 32)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("optional", optionalRaw, "float"))
					}
					pv := float32(v)
					optional = &pv
				}
			}
			{
				optionalButRequiredRaw := resp.Header.Get("Location")
				if optionalButRequiredRaw == "" {
					return nil, goahttp.ErrValidationError("ServiceWithHeadersBlockViewedResult", "MethodA", goa.MissingFieldError("Location", "header"))
				}
				v, err2 := strconv.ParseUint(optionalButRequiredRaw, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("optionalButRequired", optionalButRequiredRaw, "unsigned integer"))
				}
				optionalButRequired = uint(v)
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceWithHeadersBlockViewedResult", "MethodA", err)
			}
			p := NewMethodAAResultOK(required, optional, optionalButRequired)
			view := resp.Header.Get("goa-view")
			vres := &servicewithheadersblockviewedresultviews.AResult{Projected: p, View: view}
			res := servicewithheadersblockviewedresult.NewAResult(vres)
			return res, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceWithHeadersBlockViewedResult", "MethodA", resp.StatusCode, string(body))
		}
	}
}
`

var ValidateErrorResponseTypeDecodeCode = `// DecodeMethodAResponse returns a decoder for responses returned by the
// ValidateErrorResponseType MethodA endpoint. restoreBody controls whether the
// response body should be restored after having been read.
// DecodeMethodAResponse may return the following errors:
//	- "some_error" (type *validateerrorresponsetype.AError): http.StatusBadRequest
//	- error: internal error
func DecodeMethodAResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				required int
				err      error
			)
			{
				requiredRaw := resp.Header.Get("X-Request-Id")
				if requiredRaw == "" {
					return nil, goahttp.ErrValidationError("ValidateErrorResponseType", "MethodA", goa.MissingFieldError("X-Request-ID", "header"))
				}
				v, err2 := strconv.ParseInt(requiredRaw, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("required", requiredRaw, "integer"))
				}
				required = int(v)
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ValidateErrorResponseType", "MethodA", err)
			}
			p := NewMethodAAResultOK(required)
			view := "default"
			vres := &validateerrorresponsetypeviews.AResult{Projected: p, View: view}
			res := validateerrorresponsetype.NewAResult(vres)
			return res, nil
		case http.StatusBadRequest:
			var (
				error    string
				numOccur *int
				err      error
			)
			errorRaw := resp.Header.Get("X-Application-Error")
			if errorRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("X-Application-Error", "header"))
			}
			error = errorRaw
			{
				numOccurRaw := resp.Header.Get("X-Occur")
				if numOccurRaw != "" {
					v, err2 := strconv.ParseInt(numOccurRaw, 10, strconv.IntSize)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("numOccur", numOccurRaw, "integer"))
					}
					pv := int(v)
					numOccur = &pv
				}
			}
			if numOccur != nil {
				if *numOccur < 1 {
					err = goa.MergeErrors(err, goa.InvalidRangeError("numOccur", *numOccur, 1, true))
				}
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ValidateErrorResponseType", "MethodA", err)
			}
			return nil, NewMethodASomeError(error, numOccur)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ValidateErrorResponseType", "MethodA", resp.StatusCode, string(body))
		}
	}
}
`

var EmptyErrorResponseBodyDecodeCode = `// DecodeMethodEmptyErrorResponseBodyResponse returns a decoder for responses
// returned by the ServiceEmptyErrorResponseBody MethodEmptyErrorResponseBody
// endpoint. restoreBody controls whether the response body should be restored
// after having been read.
// DecodeMethodEmptyErrorResponseBodyResponse may return the following errors:
//	- "internal_error" (type *goa.ServiceError): http.StatusInternalServerError
//	- "not_found" (type serviceemptyerrorresponsebody.NotFound): http.StatusNotFound
//	- error: internal error
func DecodeMethodEmptyErrorResponseBodyResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
			return nil, nil
		case http.StatusInternalServerError:
			var (
				name      string
				id        string
				message   string
				temporary bool
				timeout   bool
				fault     bool
				err       error
			)
			nameRaw := resp.Header.Get("Error-Name")
			if nameRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("Error-Name", "header"))
			}
			name = nameRaw
			idRaw := resp.Header.Get("Goa-Attribute-Id")
			if idRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("goa-attribute-id", "header"))
			}
			id = idRaw
			messageRaw := resp.Header.Get("Goa-Attribute-Message")
			if messageRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("goa-attribute-message", "header"))
			}
			message = messageRaw
			{
				temporaryRaw := resp.Header.Get("Goa-Attribute-Temporary")
				if temporaryRaw == "" {
					return nil, goahttp.ErrValidationError("ServiceEmptyErrorResponseBody", "MethodEmptyErrorResponseBody", goa.MissingFieldError("goa-attribute-temporary", "header"))
				}
				v, err2 := strconv.ParseBool(temporaryRaw)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("temporary", temporaryRaw, "boolean"))
				}
				temporary = v
			}
			{
				timeoutRaw := resp.Header.Get("Goa-Attribute-Timeout")
				if timeoutRaw == "" {
					return nil, goahttp.ErrValidationError("ServiceEmptyErrorResponseBody", "MethodEmptyErrorResponseBody", goa.MissingFieldError("goa-attribute-timeout", "header"))
				}
				v, err2 := strconv.ParseBool(timeoutRaw)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("timeout", timeoutRaw, "boolean"))
				}
				timeout = v
			}
			{
				faultRaw := resp.Header.Get("Goa-Attribute-Fault")
				if faultRaw == "" {
					return nil, goahttp.ErrValidationError("ServiceEmptyErrorResponseBody", "MethodEmptyErrorResponseBody", goa.MissingFieldError("goa-attribute-fault", "header"))
				}
				v, err2 := strconv.ParseBool(faultRaw)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("fault", faultRaw, "boolean"))
				}
				fault = v
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceEmptyErrorResponseBody", "MethodEmptyErrorResponseBody", err)
			}
			return nil, NewMethodEmptyErrorResponseBodyInternalError(name, id, message, temporary, timeout, fault)
		case http.StatusNotFound:
			var (
				inHeader string
				err      error
			)
			inHeaderRaw := resp.Header.Get("In-Header")
			if inHeaderRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("in-header", "header"))
			}
			inHeader = inHeaderRaw
			if err != nil {
				return nil, goahttp.ErrValidationError("ServiceEmptyErrorResponseBody", "MethodEmptyErrorResponseBody", err)
			}
			return nil, NewMethodEmptyErrorResponseBodyNotFound(inHeader)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("ServiceEmptyErrorResponseBody", "MethodEmptyErrorResponseBody", resp.StatusCode, string(body))
		}
	}
}
`
