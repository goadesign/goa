package testdata

var MultipartPrimitiveDecoderFuncTypeCode = `// ServiceMultipartPrimitiveMethodMultipartPrimitiveDecoderFunc is the type to
// decode multipart request for the "ServiceMultipartPrimitive" service
// "MethodMultipartPrimitive" endpoint.
type ServiceMultipartPrimitiveMethodMultipartPrimitiveDecoderFunc func(*multipart.Reader, *string) error
`

var MultipartUserTypeDecoderFuncTypeCode = `// ServiceMultipartUserTypeMethodMultipartUserTypeDecoderFunc is the type to
// decode multipart request for the "ServiceMultipartUserType" service
// "MethodMultipartUserType" endpoint.
type ServiceMultipartUserTypeMethodMultipartUserTypeDecoderFunc func(*multipart.Reader, **servicemultipartusertype.MethodMultipartUserTypePayload) error
`

var MultipartArrayTypeDecoderFuncTypeCode = `// ServiceMultipartArrayTypeMethodMultipartArrayTypeDecoderFunc is the type to
// decode multipart request for the "ServiceMultipartArrayType" service
// "MethodMultipartArrayType" endpoint.
type ServiceMultipartArrayTypeMethodMultipartArrayTypeDecoderFunc func(*multipart.Reader, *[]*servicemultipartarraytype.PayloadType) error
`

var MultipartMapTypeDecoderFuncTypeCode = `// ServiceMultipartMapTypeMethodMultipartMapTypeDecoderFunc is the type to
// decode multipart request for the "ServiceMultipartMapType" service
// "MethodMultipartMapType" endpoint.
type ServiceMultipartMapTypeMethodMultipartMapTypeDecoderFunc func(*multipart.Reader, *map[string]int) error
`

var MultipartPrimitiveEncoderFuncTypeCode = `// ServiceMultipartPrimitiveMethodMultipartPrimitiveEncoderFunc is the type to
// encode multipart request for the "ServiceMultipartPrimitive" service
// "MethodMultipartPrimitive" endpoint.
type ServiceMultipartPrimitiveMethodMultipartPrimitiveEncoderFunc func(*multipart.Writer, string) error
`

var MultipartUserTypeEncoderFuncTypeCode = `// ServiceMultipartUserTypeMethodMultipartUserTypeEncoderFunc is the type to
// encode multipart request for the "ServiceMultipartUserType" service
// "MethodMultipartUserType" endpoint.
type ServiceMultipartUserTypeMethodMultipartUserTypeEncoderFunc func(*multipart.Writer, *servicemultipartusertype.MethodMultipartUserTypePayload) error
`

var MultipartArrayTypeEncoderFuncTypeCode = `// ServiceMultipartArrayTypeMethodMultipartArrayTypeEncoderFunc is the type to
// encode multipart request for the "ServiceMultipartArrayType" service
// "MethodMultipartArrayType" endpoint.
type ServiceMultipartArrayTypeMethodMultipartArrayTypeEncoderFunc func(*multipart.Writer, []*servicemultipartarraytype.PayloadType) error
`

var MultipartMapTypeEncoderFuncTypeCode = `// ServiceMultipartMapTypeMethodMultipartMapTypeEncoderFunc is the type to
// encode multipart request for the "ServiceMultipartMapType" service
// "MethodMultipartMapType" endpoint.
type ServiceMultipartMapTypeMethodMultipartMapTypeEncoderFunc func(*multipart.Writer, map[string]int) error
`

var MultipartPrimitiveDecoderFuncCode = `// NewServiceMultipartPrimitiveMethodMultipartPrimitiveDecoder returns a
// decoder to decode the multipart request for the "ServiceMultipartPrimitive"
// service "MethodMultipartPrimitive" endpoint.
func NewServiceMultipartPrimitiveMethodMultipartPrimitiveDecoder(mux goahttp.Muxer, ServiceMultipartPrimitiveMethodMultipartPrimitiveDecoderFn ServiceMultipartPrimitiveMethodMultipartPrimitiveDecoderFunc) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v interface{}) error {
			mr, err := r.MultipartReader()
			if err != nil {
				return err
			}
			p := v.(*string)
			if err := ServiceMultipartPrimitiveMethodMultipartPrimitiveDecoderFn(mr, p); err != nil {
				return err
			}
			return nil
		})
	}
}
`

var MultipartUserTypeDecoderFuncCode = `// NewServiceMultipartUserTypeMethodMultipartUserTypeDecoder returns a decoder
// to decode the multipart request for the "ServiceMultipartUserType" service
// "MethodMultipartUserType" endpoint.
func NewServiceMultipartUserTypeMethodMultipartUserTypeDecoder(mux goahttp.Muxer, ServiceMultipartUserTypeMethodMultipartUserTypeDecoderFn ServiceMultipartUserTypeMethodMultipartUserTypeDecoderFunc) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v interface{}) error {
			mr, err := r.MultipartReader()
			if err != nil {
				return err
			}
			p := v.(**servicemultipartusertype.MethodMultipartUserTypePayload)
			if err := ServiceMultipartUserTypeMethodMultipartUserTypeDecoderFn(mr, p); err != nil {
				return err
			}
			return nil
		})
	}
}
`

var MultipartArrayTypeDecoderFuncCode = `// NewServiceMultipartArrayTypeMethodMultipartArrayTypeDecoder returns a
// decoder to decode the multipart request for the "ServiceMultipartArrayType"
// service "MethodMultipartArrayType" endpoint.
func NewServiceMultipartArrayTypeMethodMultipartArrayTypeDecoder(mux goahttp.Muxer, ServiceMultipartArrayTypeMethodMultipartArrayTypeDecoderFn ServiceMultipartArrayTypeMethodMultipartArrayTypeDecoderFunc) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v interface{}) error {
			mr, err := r.MultipartReader()
			if err != nil {
				return err
			}
			p := v.(*[]*servicemultipartarraytype.PayloadType)
			if err := ServiceMultipartArrayTypeMethodMultipartArrayTypeDecoderFn(mr, p); err != nil {
				return err
			}
			return nil
		})
	}
}
`

var MultipartMapTypeDecoderFuncCode = `// NewServiceMultipartMapTypeMethodMultipartMapTypeDecoder returns a decoder to
// decode the multipart request for the "ServiceMultipartMapType" service
// "MethodMultipartMapType" endpoint.
func NewServiceMultipartMapTypeMethodMultipartMapTypeDecoder(mux goahttp.Muxer, ServiceMultipartMapTypeMethodMultipartMapTypeDecoderFn ServiceMultipartMapTypeMethodMultipartMapTypeDecoderFunc) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v interface{}) error {
			mr, err := r.MultipartReader()
			if err != nil {
				return err
			}
			p := v.(*map[string]int)
			if err := ServiceMultipartMapTypeMethodMultipartMapTypeDecoderFn(mr, p); err != nil {
				return err
			}
			return nil
		})
	}
}
`

var MultipartWithParamsDecoderFuncCode = `// NewServiceMultipartWithParamsMethodMultipartWithParamsDecoder returns a
// decoder to decode the multipart request for the "ServiceMultipartWithParams"
// service "MethodMultipartWithParams" endpoint.
func NewServiceMultipartWithParamsMethodMultipartWithParamsDecoder(mux goahttp.Muxer, ServiceMultipartWithParamsMethodMultipartWithParamsDecoderFn ServiceMultipartWithParamsMethodMultipartWithParamsDecoderFunc) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v interface{}) error {
			mr, err := r.MultipartReader()
			if err != nil {
				return err
			}
			p := v.(**servicemultipartwithparams.PayloadType)
			if err := ServiceMultipartWithParamsMethodMultipartWithParamsDecoderFn(mr, p); err != nil {
				return err
			}
			var (
				a   string
				c   map[int][]string
				b   *string
				err error

				params = mux.Vars(r)
			)
			a = params["a"]
			err = goa.MergeErrors(err, goa.ValidatePattern("a", a, "patterna"))
			{
				cRaw := r.URL.Query()
				if len(cRaw) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError("c", "query string"))
				}
				c = make(map[int][]string, len(cRaw))
				for keyRaw, val := range cRaw {
					var key int
					{
						v, err2 := strconv.ParseInt(keyRaw, 10, strconv.IntSize)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "integer"))
						}
						key = int(v)
					}
					c[key] = val
				}
			}
			bRaw := r.Header.Get("Authorization")
			if bRaw != "" {
				b = &bRaw
			}
			if b != nil {
				err = goa.MergeErrors(err, goa.ValidatePattern("b", *b, "patternb"))
			}
			if err != nil {
				return err
			}
			(*p).A = a
			(*p).C = c
			(*p).B = b
			return nil
		})
	}
}
`

var MultipartPrimitiveEncoderFuncCode = `// NewServiceMultipartPrimitiveMethodMultipartPrimitiveEncoder returns an
// encoder to encode the multipart request for the "ServiceMultipartPrimitive"
// service "MethodMultipartPrimitive" endpoint.
func NewServiceMultipartPrimitiveMethodMultipartPrimitiveEncoder(encoderFn ServiceMultipartPrimitiveMethodMultipartPrimitiveEncoderFunc) func(r *http.Request) goahttp.Encoder {
	return func(r *http.Request) goahttp.Encoder {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		return goahttp.EncodingFunc(func(v interface{}) error {
			p := v.(string)
			if err := encoderFn(mw, p); err != nil {
				return err
			}
			r.Body = ioutil.NopCloser(body)
			r.Header.Set("Content-Type", mw.FormDataContentType())
			return mw.Close()
		})
	}
}
`

var MultipartUserTypeEncoderFuncCode = `// NewServiceMultipartUserTypeMethodMultipartUserTypeEncoder returns an encoder
// to encode the multipart request for the "ServiceMultipartUserType" service
// "MethodMultipartUserType" endpoint.
func NewServiceMultipartUserTypeMethodMultipartUserTypeEncoder(encoderFn ServiceMultipartUserTypeMethodMultipartUserTypeEncoderFunc) func(r *http.Request) goahttp.Encoder {
	return func(r *http.Request) goahttp.Encoder {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		return goahttp.EncodingFunc(func(v interface{}) error {
			p := v.(*servicemultipartusertype.MethodMultipartUserTypePayload)
			if err := encoderFn(mw, p); err != nil {
				return err
			}
			r.Body = ioutil.NopCloser(body)
			r.Header.Set("Content-Type", mw.FormDataContentType())
			return mw.Close()
		})
	}
}
`
var MultipartArrayTypeEncoderFuncCode = `// NewServiceMultipartArrayTypeMethodMultipartArrayTypeEncoder returns an
// encoder to encode the multipart request for the "ServiceMultipartArrayType"
// service "MethodMultipartArrayType" endpoint.
func NewServiceMultipartArrayTypeMethodMultipartArrayTypeEncoder(encoderFn ServiceMultipartArrayTypeMethodMultipartArrayTypeEncoderFunc) func(r *http.Request) goahttp.Encoder {
	return func(r *http.Request) goahttp.Encoder {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		return goahttp.EncodingFunc(func(v interface{}) error {
			p := v.([]*servicemultipartarraytype.PayloadType)
			if err := encoderFn(mw, p); err != nil {
				return err
			}
			r.Body = ioutil.NopCloser(body)
			r.Header.Set("Content-Type", mw.FormDataContentType())
			return mw.Close()
		})
	}
}
`

var MultipartMapTypeEncoderFuncCode = `// NewServiceMultipartMapTypeMethodMultipartMapTypeEncoder returns an encoder
// to encode the multipart request for the "ServiceMultipartMapType" service
// "MethodMultipartMapType" endpoint.
func NewServiceMultipartMapTypeMethodMultipartMapTypeEncoder(encoderFn ServiceMultipartMapTypeMethodMultipartMapTypeEncoderFunc) func(r *http.Request) goahttp.Encoder {
	return func(r *http.Request) goahttp.Encoder {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		return goahttp.EncodingFunc(func(v interface{}) error {
			p := v.(map[string]int)
			if err := encoderFn(mw, p); err != nil {
				return err
			}
			r.Body = ioutil.NopCloser(body)
			r.Header.Set("Content-Type", mw.FormDataContentType())
			return mw.Close()
		})
	}
}
`

var MultipartWithParamsEncoderFuncCode = `// NewServiceMultipartWithParamsMethodMultipartWithParamsEncoder returns an
// encoder to encode the multipart request for the "ServiceMultipartWithParams"
// service "MethodMultipartWithParams" endpoint.
func NewServiceMultipartWithParamsMethodMultipartWithParamsEncoder(encoderFn ServiceMultipartWithParamsMethodMultipartWithParamsEncoderFunc) func(r *http.Request) goahttp.Encoder {
	return func(r *http.Request) goahttp.Encoder {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		return goahttp.EncodingFunc(func(v interface{}) error {
			p := v.(*servicemultipartwithparams.PayloadType)
			if err := encoderFn(mw, p); err != nil {
				return err
			}
			r.Body = ioutil.NopCloser(body)
			r.Header.Set("Content-Type", mw.FormDataContentType())
			return mw.Close()
		})
	}
}
`
