package testdata

var PayloadQueryBoolDecodeCode = `// DecodeMethodQueryBoolRequest returns a decoder for requests sent to the
// ServiceQueryBool MethodQueryBool endpoint.
func DecodeMethodQueryBoolRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *bool
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseBool(qRaw)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "boolean"))
				}
				q = &v
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryBoolMethodQueryBoolPayload(q), nil
	}
}
`

var PayloadQueryBoolValidateDecodeCode = `// DecodeMethodQueryBoolValidateRequest returns a decoder for requests sent to
// the ServiceQueryBoolValidate MethodQueryBoolValidate endpoint.
func DecodeMethodQueryBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   bool
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseBool(qRaw)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "boolean"))
			}
			q = v
		}
		if !(q == true) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", q, []interface{}{true}))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryBoolValidateMethodQueryBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryIntDecodeCode = `// DecodeMethodQueryIntRequest returns a decoder for requests sent to the
// ServiceQueryInt MethodQueryInt endpoint.
func DecodeMethodQueryIntRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *int
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseInt(qRaw, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "integer"))
				}
				pv := int(v)
				q = &pv
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryIntMethodQueryIntPayload(q), nil
	}
}
`

var PayloadQueryIntValidateDecodeCode = `// DecodeMethodQueryIntValidateRequest returns a decoder for requests sent to
// the ServiceQueryIntValidate MethodQueryIntValidate endpoint.
func DecodeMethodQueryIntValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   int
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseInt(qRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "integer"))
			}
			q = int(v)
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryIntValidateMethodQueryIntValidatePayload(q), nil
	}
}
`

var PayloadQueryInt32DecodeCode = `// DecodeMethodQueryInt32Request returns a decoder for requests sent to the
// ServiceQueryInt32 MethodQueryInt32 endpoint.
func DecodeMethodQueryInt32Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *int32
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseInt(qRaw, 10, 32)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "integer"))
				}
				pv := int32(v)
				q = &pv
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryInt32MethodQueryInt32Payload(q), nil
	}
}
`

var PayloadQueryInt32ValidateDecodeCode = `// DecodeMethodQueryInt32ValidateRequest returns a decoder for requests sent to
// the ServiceQueryInt32Validate MethodQueryInt32Validate endpoint.
func DecodeMethodQueryInt32ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   int32
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseInt(qRaw, 10, 32)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "integer"))
			}
			q = int32(v)
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryInt32ValidateMethodQueryInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryInt64DecodeCode = `// DecodeMethodQueryInt64Request returns a decoder for requests sent to the
// ServiceQueryInt64 MethodQueryInt64 endpoint.
func DecodeMethodQueryInt64Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *int64
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseInt(qRaw, 10, 64)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "integer"))
				}
				q = &v
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryInt64MethodQueryInt64Payload(q), nil
	}
}
`

var PayloadQueryInt64ValidateDecodeCode = `// DecodeMethodQueryInt64ValidateRequest returns a decoder for requests sent to
// the ServiceQueryInt64Validate MethodQueryInt64Validate endpoint.
func DecodeMethodQueryInt64ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   int64
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseInt(qRaw, 10, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "integer"))
			}
			q = v
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryInt64ValidateMethodQueryInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryUIntDecodeCode = `// DecodeMethodQueryUIntRequest returns a decoder for requests sent to the
// ServiceQueryUInt MethodQueryUInt endpoint.
func DecodeMethodQueryUIntRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *uint
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseUint(qRaw, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "unsigned integer"))
				}
				pv := uint(v)
				q = &pv
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryUIntMethodQueryUIntPayload(q), nil
	}
}
`

var PayloadQueryUIntValidateDecodeCode = `// DecodeMethodQueryUIntValidateRequest returns a decoder for requests sent to
// the ServiceQueryUIntValidate MethodQueryUIntValidate endpoint.
func DecodeMethodQueryUIntValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   uint
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseUint(qRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "unsigned integer"))
			}
			q = uint(v)
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryUIntValidateMethodQueryUIntValidatePayload(q), nil
	}
}
`

var PayloadQueryUInt32DecodeCode = `// DecodeMethodQueryUInt32Request returns a decoder for requests sent to the
// ServiceQueryUInt32 MethodQueryUInt32 endpoint.
func DecodeMethodQueryUInt32Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *uint32
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseUint(qRaw, 10, 32)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "unsigned integer"))
				}
				pv := uint32(v)
				q = &pv
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryUInt32MethodQueryUInt32Payload(q), nil
	}
}
`

var PayloadQueryUInt32ValidateDecodeCode = `// DecodeMethodQueryUInt32ValidateRequest returns a decoder for requests sent
// to the ServiceQueryUInt32Validate MethodQueryUInt32Validate endpoint.
func DecodeMethodQueryUInt32ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   uint32
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseUint(qRaw, 10, 32)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "unsigned integer"))
			}
			q = uint32(v)
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryUInt32ValidateMethodQueryUInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryUInt64DecodeCode = `// DecodeMethodQueryUInt64Request returns a decoder for requests sent to the
// ServiceQueryUInt64 MethodQueryUInt64 endpoint.
func DecodeMethodQueryUInt64Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *uint64
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseUint(qRaw, 10, 64)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "unsigned integer"))
				}
				q = &v
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryUInt64MethodQueryUInt64Payload(q), nil
	}
}
`

var PayloadQueryUInt64ValidateDecodeCode = `// DecodeMethodQueryUInt64ValidateRequest returns a decoder for requests sent
// to the ServiceQueryUInt64Validate MethodQueryUInt64Validate endpoint.
func DecodeMethodQueryUInt64ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   uint64
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseUint(qRaw, 10, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "unsigned integer"))
			}
			q = v
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryUInt64ValidateMethodQueryUInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryFloat32DecodeCode = `// DecodeMethodQueryFloat32Request returns a decoder for requests sent to the
// ServiceQueryFloat32 MethodQueryFloat32 endpoint.
func DecodeMethodQueryFloat32Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *float32
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseFloat(qRaw, 32)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "float"))
				}
				pv := float32(v)
				q = &pv
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryFloat32MethodQueryFloat32Payload(q), nil
	}
}
`

var PayloadQueryFloat32ValidateDecodeCode = `// DecodeMethodQueryFloat32ValidateRequest returns a decoder for requests sent
// to the ServiceQueryFloat32Validate MethodQueryFloat32Validate endpoint.
func DecodeMethodQueryFloat32ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   float32
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseFloat(qRaw, 32)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "float"))
			}
			q = float32(v)
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryFloat32ValidateMethodQueryFloat32ValidatePayload(q), nil
	}
}
`

var PayloadQueryFloat64DecodeCode = `// DecodeMethodQueryFloat64Request returns a decoder for requests sent to the
// ServiceQueryFloat64 MethodQueryFloat64 endpoint.
func DecodeMethodQueryFloat64Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   *float64
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				v, err2 := strconv.ParseFloat(qRaw, 64)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "float"))
				}
				q = &v
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryFloat64MethodQueryFloat64Payload(q), nil
	}
}
`

var PayloadQueryFloat64ValidateDecodeCode = `// DecodeMethodQueryFloat64ValidateRequest returns a decoder for requests sent
// to the ServiceQueryFloat64Validate MethodQueryFloat64Validate endpoint.
func DecodeMethodQueryFloat64ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   float64
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseFloat(qRaw, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "float"))
			}
			q = v
		}
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryFloat64ValidateMethodQueryFloat64ValidatePayload(q), nil
	}
}
`

var PayloadQueryStringDecodeCode = `// DecodeMethodQueryStringRequest returns a decoder for requests sent to the
// ServiceQueryString MethodQueryString endpoint.
func DecodeMethodQueryStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q *string
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = &qRaw
		}

		return NewMethodQueryStringMethodQueryStringPayload(q), nil
	}
}
`

var PayloadQueryStringValidateDecodeCode = `// DecodeMethodQueryStringValidateRequest returns a decoder for requests sent
// to the ServiceQueryStringValidate MethodQueryStringValidate endpoint.
func DecodeMethodQueryStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   string
			err error
		)
		q = r.URL.Query().Get("q")
		if q == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if !(q == "val") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", q, []interface{}{"val"}))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryStringValidateMethodQueryStringValidatePayload(q), nil
	}
}
`

var PayloadQueryBytesDecodeCode = `// DecodeMethodQueryBytesRequest returns a decoder for requests sent to the
// ServiceQueryBytes MethodQueryBytes endpoint.
func DecodeMethodQueryBytesRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q []byte
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw != "" {
				q = []byte(qRaw)
			}
		}

		return NewMethodQueryBytesMethodQueryBytesPayload(q), nil
	}
}
`

var PayloadQueryBytesValidateDecodeCode = `// DecodeMethodQueryBytesValidateRequest returns a decoder for requests sent to
// the ServiceQueryBytesValidate MethodQueryBytesValidate endpoint.
func DecodeMethodQueryBytesValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []byte
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = []byte(qRaw)
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryBytesValidateMethodQueryBytesValidatePayload(q), nil
	}
}
`

var PayloadQueryAnyDecodeCode = `// DecodeMethodQueryAnyRequest returns a decoder for requests sent to the
// ServiceQueryAny MethodQueryAny endpoint.
func DecodeMethodQueryAnyRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q interface{}
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = qRaw
		}

		return NewMethodQueryAnyMethodQueryAnyPayload(q), nil
	}
}
`

var PayloadQueryAnyValidateDecodeCode = `// DecodeMethodQueryAnyValidateRequest returns a decoder for requests sent to
// the ServiceQueryAnyValidate MethodQueryAnyValidate endpoint.
func DecodeMethodQueryAnyValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   interface{}
			err error
		)
		q = r.URL.Query().Get("q")
		if q == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if !(q == "val" || q == 1) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", q, []interface{}{"val", 1}))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryAnyValidateMethodQueryAnyValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBoolDecodeCode = `// DecodeMethodQueryArrayBoolRequest returns a decoder for requests sent to the
// ServiceQueryArrayBool MethodQueryArrayBool endpoint.
func DecodeMethodQueryArrayBoolRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []bool
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]bool, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseBool(rv)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of booleans"))
					}
					q[i] = v
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayBoolMethodQueryArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryArrayBoolValidateDecodeCode = `// DecodeMethodQueryArrayBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayBoolValidate MethodQueryArrayBoolValidate
// endpoint.
func DecodeMethodQueryArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []bool
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]bool, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseBool(rv)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of booleans"))
				}
				q[i] = v
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if !(e == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayBoolValidateMethodQueryArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayIntDecodeCode = `// DecodeMethodQueryArrayIntRequest returns a decoder for requests sent to the
// ServiceQueryArrayInt MethodQueryArrayInt endpoint.
func DecodeMethodQueryArrayIntRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []int
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]int, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseInt(rv, 10, strconv.IntSize)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of integers"))
					}
					q[i] = int(v)
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayIntMethodQueryArrayIntPayload(q), nil
	}
}
`

var PayloadQueryArrayIntValidateDecodeCode = `// DecodeMethodQueryArrayIntValidateRequest returns a decoder for requests sent
// to the ServiceQueryArrayIntValidate MethodQueryArrayIntValidate endpoint.
func DecodeMethodQueryArrayIntValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []int
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]int, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseInt(rv, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of integers"))
				}
				q[i] = int(v)
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayIntValidateMethodQueryArrayIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt32DecodeCode = `// DecodeMethodQueryArrayInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayInt32 MethodQueryArrayInt32 endpoint.
func DecodeMethodQueryArrayInt32Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []int32
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]int32, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseInt(rv, 10, 32)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of integers"))
					}
					q[i] = int32(v)
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayInt32MethodQueryArrayInt32Payload(q), nil
	}
}
`

var PayloadQueryArrayInt32ValidateDecodeCode = `// DecodeMethodQueryArrayInt32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayInt32Validate MethodQueryArrayInt32Validate
// endpoint.
func DecodeMethodQueryArrayInt32ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []int32
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]int32, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseInt(rv, 10, 32)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of integers"))
				}
				q[i] = int32(v)
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayInt32ValidateMethodQueryArrayInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt64DecodeCode = `// DecodeMethodQueryArrayInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayInt64 MethodQueryArrayInt64 endpoint.
func DecodeMethodQueryArrayInt64Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []int64
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]int64, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseInt(rv, 10, 64)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of integers"))
					}
					q[i] = v
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayInt64MethodQueryArrayInt64Payload(q), nil
	}
}
`

var PayloadQueryArrayInt64ValidateDecodeCode = `// DecodeMethodQueryArrayInt64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayInt64Validate MethodQueryArrayInt64Validate
// endpoint.
func DecodeMethodQueryArrayInt64ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []int64
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]int64, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseInt(rv, 10, 64)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of integers"))
				}
				q[i] = v
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayInt64ValidateMethodQueryArrayInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUIntDecodeCode = `// DecodeMethodQueryArrayUIntRequest returns a decoder for requests sent to the
// ServiceQueryArrayUInt MethodQueryArrayUInt endpoint.
func DecodeMethodQueryArrayUIntRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []uint
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]uint, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseUint(rv, 10, strconv.IntSize)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of unsigned integers"))
					}
					q[i] = uint(v)
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayUIntMethodQueryArrayUIntPayload(q), nil
	}
}
`

var PayloadQueryArrayUIntValidateDecodeCode = `// DecodeMethodQueryArrayUIntValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUIntValidate MethodQueryArrayUIntValidate
// endpoint.
func DecodeMethodQueryArrayUIntValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []uint
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]uint, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseUint(rv, 10, strconv.IntSize)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of unsigned integers"))
				}
				q[i] = uint(v)
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayUIntValidateMethodQueryArrayUIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt32DecodeCode = `// DecodeMethodQueryArrayUInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt32 MethodQueryArrayUInt32 endpoint.
func DecodeMethodQueryArrayUInt32Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []uint32
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]uint32, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseUint(rv, 10, 32)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of unsigned integers"))
					}
					q[i] = int32(v)
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayUInt32MethodQueryArrayUInt32Payload(q), nil
	}
}
`

var PayloadQueryArrayUInt32ValidateDecodeCode = `// DecodeMethodQueryArrayUInt32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUInt32Validate MethodQueryArrayUInt32Validate
// endpoint.
func DecodeMethodQueryArrayUInt32ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []uint32
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]uint32, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseUint(rv, 10, 32)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of unsigned integers"))
				}
				q[i] = int32(v)
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayUInt32ValidateMethodQueryArrayUInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt64DecodeCode = `// DecodeMethodQueryArrayUInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt64 MethodQueryArrayUInt64 endpoint.
func DecodeMethodQueryArrayUInt64Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []uint64
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]uint64, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseUint(rv, 10, 64)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of unsigned integers"))
					}
					q[i] = v
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayUInt64MethodQueryArrayUInt64Payload(q), nil
	}
}
`

var PayloadQueryArrayUInt64ValidateDecodeCode = `// DecodeMethodQueryArrayUInt64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUInt64Validate MethodQueryArrayUInt64Validate
// endpoint.
func DecodeMethodQueryArrayUInt64ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []uint64
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]uint64, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseUint(rv, 10, 64)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of unsigned integers"))
				}
				q[i] = v
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayUInt64ValidateMethodQueryArrayUInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat32DecodeCode = `// DecodeMethodQueryArrayFloat32Request returns a decoder for requests sent to
// the ServiceQueryArrayFloat32 MethodQueryArrayFloat32 endpoint.
func DecodeMethodQueryArrayFloat32Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []float32
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]float32, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseFloat(rv, 32)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of floats"))
					}
					q[i] = float32(v)
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayFloat32MethodQueryArrayFloat32Payload(q), nil
	}
}
`

var PayloadQueryArrayFloat32ValidateDecodeCode = `// DecodeMethodQueryArrayFloat32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayFloat32Validate MethodQueryArrayFloat32Validate
// endpoint.
func DecodeMethodQueryArrayFloat32ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []float32
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]float32, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseFloat(rv, 32)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of floats"))
				}
				q[i] = float32(v)
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayFloat32ValidateMethodQueryArrayFloat32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat64DecodeCode = `// DecodeMethodQueryArrayFloat64Request returns a decoder for requests sent to
// the ServiceQueryArrayFloat64 MethodQueryArrayFloat64 endpoint.
func DecodeMethodQueryArrayFloat64Request(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []float64
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]float64, len(qRaw))
				for i, rv := range qRaw {
					v, err2 := strconv.ParseFloat(rv, 64)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of floats"))
					}
					q[i] = v
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayFloat64MethodQueryArrayFloat64Payload(q), nil
	}
}
`

var PayloadQueryArrayFloat64ValidateDecodeCode = `// DecodeMethodQueryArrayFloat64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayFloat64Validate MethodQueryArrayFloat64Validate
// endpoint.
func DecodeMethodQueryArrayFloat64ValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []float64
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]float64, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseFloat(rv, 64)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of floats"))
				}
				q[i] = v
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayFloat64ValidateMethodQueryArrayFloat64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayStringDecodeCode = `// DecodeMethodQueryArrayStringRequest returns a decoder for requests sent to
// the ServiceQueryArrayString MethodQueryArrayString endpoint.
func DecodeMethodQueryArrayStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q []string
		)
		q = r.URL.Query()["q"]

		return NewMethodQueryArrayStringMethodQueryArrayStringPayload(q), nil
	}
}
`

var PayloadQueryArrayStringValidateDecodeCode = `// DecodeMethodQueryArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayStringValidate MethodQueryArrayStringValidate
// endpoint.
func DecodeMethodQueryArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []string
			err error
		)
		q = r.URL.Query()["q"]
		if q == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if !(e == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayStringValidateMethodQueryArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBytesDecodeCode = `// DecodeMethodQueryArrayBytesRequest returns a decoder for requests sent to
// the ServiceQueryArrayBytes MethodQueryArrayBytes endpoint.
func DecodeMethodQueryArrayBytesRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q [][]byte
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([][]byte, len(qRaw))
				for i, rv := range qRaw {
					q[i] = []byte(rv)
				}
			}
		}

		return NewMethodQueryArrayBytesMethodQueryArrayBytesPayload(q), nil
	}
}
`

var PayloadQueryArrayBytesValidateDecodeCode = `// DecodeMethodQueryArrayBytesValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayBytesValidate MethodQueryArrayBytesValidate
// endpoint.
func DecodeMethodQueryArrayBytesValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   [][]byte
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([][]byte, len(qRaw))
			for i, rv := range qRaw {
				q[i] = []byte(rv)
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if len(e) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[*]", e, len(e), 2, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayBytesValidateMethodQueryArrayBytesValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayAnyDecodeCode = `// DecodeMethodQueryArrayAnyRequest returns a decoder for requests sent to the
// ServiceQueryArrayAny MethodQueryArrayAny endpoint.
func DecodeMethodQueryArrayAnyRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q []interface{}
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw != nil {
				q = make([]interface{}, len(qRaw))
				for i, rv := range qRaw {
					q[i] = rv
				}
			}
		}

		return NewMethodQueryArrayAnyMethodQueryArrayAnyPayload(q), nil
	}
}
`

var PayloadQueryArrayAnyValidateDecodeCode = `// DecodeMethodQueryArrayAnyValidateRequest returns a decoder for requests sent
// to the ServiceQueryArrayAnyValidate MethodQueryArrayAnyValidate endpoint.
func DecodeMethodQueryArrayAnyValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []interface{}
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]interface{}, len(qRaw))
			for i, rv := range qRaw {
				q[i] = rv
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if !(e == "val" || e == 1) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{"val", 1}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryArrayAnyValidateMethodQueryArrayAnyValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringStringDecodeCode = `// DecodeMethodQueryMapStringStringRequest returns a decoder for requests sent
// to the ServiceQueryMapStringString MethodQueryMapStringString endpoint.
func DecodeMethodQueryMapStringStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q map[string]string
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[string]string, len(qRaw))
				for key, va := range qRaw {
					var val string
					{
						val = va[0]
					}
					q[key] = val
				}
			}
		}

		return NewMethodQueryMapStringStringMethodQueryMapStringStringPayload(q), nil
	}
}
`

var PayloadQueryMapStringStringValidateDecodeCode = `// DecodeMethodQueryMapStringStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringStringValidate
// MethodQueryMapStringStringValidate endpoint.
func DecodeMethodQueryMapStringStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string]string
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[string]string, len(qRaw))
			for key, va := range qRaw {
				var val string
				{
					val = va[0]
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == "key") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{"key"}))
			}
			if !(v == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[key]", v, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapStringStringValidateMethodQueryMapStringStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringBoolDecodeCode = `// DecodeMethodQueryMapStringBoolRequest returns a decoder for requests sent to
// the ServiceQueryMapStringBool MethodQueryMapStringBool endpoint.
func DecodeMethodQueryMapStringBoolRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[string]bool, len(qRaw))
				for key, va := range qRaw {
					var val bool
					{
						valRaw := va[0]
						v, err2 := strconv.ParseBool(valRaw)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "boolean"))
						}
						val = v
					}
					q[key] = val
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapStringBoolMethodQueryMapStringBoolPayload(q), nil
	}
}
`

var PayloadQueryMapStringBoolValidateDecodeCode = `// DecodeMethodQueryMapStringBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryMapStringBoolValidate
// MethodQueryMapStringBoolValidate endpoint.
func DecodeMethodQueryMapStringBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[string]bool, len(qRaw))
			for key, va := range qRaw {
				var val bool
				{
					valRaw := va[0]
					v, err2 := strconv.ParseBool(valRaw)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "boolean"))
					}
					val = v
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == "key") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{"key"}))
			}
			if !(v == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[key]", v, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapStringBoolValidateMethodQueryMapStringBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolStringDecodeCode = `// DecodeMethodQueryMapBoolStringRequest returns a decoder for requests sent to
// the ServiceQueryMapBoolString MethodQueryMapBoolString endpoint.
func DecodeMethodQueryMapBoolStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool]string
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[bool]string, len(qRaw))
				for keyRaw, va := range qRaw {
					var key bool
					{
						v, err2 := strconv.ParseBool(keyRaw)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
						}
						key = v
					}
					var val string
					{
						val = va[0]
					}
					q[key] = val
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolStringMethodQueryMapBoolStringPayload(q), nil
	}
}
`

var PayloadQueryMapBoolStringValidateDecodeCode = `// DecodeMethodQueryMapBoolStringValidateRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolStringValidate
// MethodQueryMapBoolStringValidate endpoint.
func DecodeMethodQueryMapBoolStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool]string
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[bool]string, len(qRaw))
			for keyRaw, va := range qRaw {
				var key bool
				{
					v, err2 := strconv.ParseBool(keyRaw)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
					}
					key = v
				}
				var val string
				{
					val = va[0]
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{true}))
			}
			if !(v == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[key]", v, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolStringValidateMethodQueryMapBoolStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolBoolDecodeCode = `// DecodeMethodQueryMapBoolBoolRequest returns a decoder for requests sent to
// the ServiceQueryMapBoolBool MethodQueryMapBoolBool endpoint.
func DecodeMethodQueryMapBoolBoolRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[bool]bool, len(qRaw))
				for keyRaw, va := range qRaw {
					var key bool
					{
						v, err2 := strconv.ParseBool(keyRaw)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
						}
						key = v
					}
					var val bool
					{
						valRaw := va[0]
						v, err2 := strconv.ParseBool(valRaw)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "boolean"))
						}
						val = v
					}
					q[key] = val
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolBoolMethodQueryMapBoolBoolPayload(q), nil
	}
}
`

var PayloadQueryMapBoolBoolValidateDecodeCode = `// DecodeMethodQueryMapBoolBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolBoolValidate MethodQueryMapBoolBoolValidate
// endpoint.
func DecodeMethodQueryMapBoolBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[bool]bool, len(qRaw))
			for keyRaw, va := range qRaw {
				var key bool
				{
					v, err2 := strconv.ParseBool(keyRaw)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
					}
					key = v
				}
				var val bool
				{
					valRaw := va[0]
					v, err2 := strconv.ParseBool(valRaw)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "boolean"))
					}
					val = v
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == false) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{false}))
			}
			if !(v == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[key]", v, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolBoolValidateMethodQueryMapBoolBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayStringDecodeCode = `// DecodeMethodQueryMapStringArrayStringRequest returns a decoder for requests
// sent to the ServiceQueryMapStringArrayString MethodQueryMapStringArrayString
// endpoint.
func DecodeMethodQueryMapStringArrayStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q map[string][]string
		)
		q = r.URL.Query()

		return NewMethodQueryMapStringArrayStringMethodQueryMapStringArrayStringPayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayStringValidateDecodeCode = `// DecodeMethodQueryMapStringArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringArrayStringValidate
// MethodQueryMapStringArrayStringValidate endpoint.
func DecodeMethodQueryMapStringArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string][]string
			err error
		)
		q = r.URL.Query()
		if len(q) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == "key") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{"key"}))
			}
			if len(v) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[key]", v, len(v), 2, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapStringArrayStringValidateMethodQueryMapStringArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayBoolDecodeCode = `// DecodeMethodQueryMapStringArrayBoolRequest returns a decoder for requests
// sent to the ServiceQueryMapStringArrayBool MethodQueryMapStringArrayBool
// endpoint.
func DecodeMethodQueryMapStringArrayBoolRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string][]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[string][]bool, len(qRaw))
				for key, valRaw := range qRaw {
					var val []bool
					{
						val = make([]bool, len(valRaw))
						for i, rv := range valRaw {
							v, err2 := strconv.ParseBool(rv)
							if err2 != nil {
								err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "array of booleans"))
							}
							val[i] = v
						}
					}
					q[key] = val
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapStringArrayBoolMethodQueryMapStringArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayBoolValidateDecodeCode = `// DecodeMethodQueryMapStringArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringArrayBoolValidate
// MethodQueryMapStringArrayBoolValidate endpoint.
func DecodeMethodQueryMapStringArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string][]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[string][]bool, len(qRaw))
			for key, valRaw := range qRaw {
				var val []bool
				{
					val = make([]bool, len(valRaw))
					for i, rv := range valRaw {
						v, err2 := strconv.ParseBool(rv)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "array of booleans"))
						}
						val[i] = v
					}
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == "key") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{"key"}))
			}
			if len(v) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[key]", v, len(v), 2, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapStringArrayBoolValidateMethodQueryMapStringArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayStringDecodeCode = `// DecodeMethodQueryMapBoolArrayStringRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolArrayString MethodQueryMapBoolArrayString
// endpoint.
func DecodeMethodQueryMapBoolArrayStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool][]string
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[bool][]string, len(qRaw))
				for keyRaw, val := range qRaw {
					var key bool
					{
						v, err2 := strconv.ParseBool(keyRaw)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
						}
						key = v
					}
					q[key] = val
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolArrayStringMethodQueryMapBoolArrayStringPayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayStringValidateDecodeCode = `// DecodeMethodQueryMapBoolArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapBoolArrayStringValidate
// MethodQueryMapBoolArrayStringValidate endpoint.
func DecodeMethodQueryMapBoolArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool][]string
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[bool][]string, len(qRaw))
				for keyRaw, val := range qRaw {
					var key bool
					{
						v, err2 := strconv.ParseBool(keyRaw)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
						}
						key = v
					}
					q[key] = val
				}
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{true}))
			}
			if len(v) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[key]", v, len(v), 2, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolArrayStringValidateMethodQueryMapBoolArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayBoolDecodeCode = `// DecodeMethodQueryMapBoolArrayBoolRequest returns a decoder for requests sent
// to the ServiceQueryMapBoolArrayBool MethodQueryMapBoolArrayBool endpoint.
func DecodeMethodQueryMapBoolArrayBoolRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool][]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) != 0 {
				q = make(map[bool][]bool, len(qRaw))
				for keyRaw, valRaw := range qRaw {
					var key bool
					{
						v, err2 := strconv.ParseBool(keyRaw)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
						}
						key = v
					}
					var val []bool
					{
						val = make([]bool, len(valRaw))
						for i, rv := range valRaw {
							v, err2 := strconv.ParseBool(rv)
							if err2 != nil {
								err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "array of booleans"))
							}
							val[i] = v
						}
					}
					q[key] = val
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolArrayBoolMethodQueryMapBoolArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayBoolValidateDecodeCode = `// DecodeMethodQueryMapBoolArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapBoolArrayBoolValidate
// MethodQueryMapBoolArrayBoolValidate endpoint.
func DecodeMethodQueryMapBoolArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool][]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[bool][]bool, len(qRaw))
			for keyRaw, valRaw := range qRaw {
				var key bool
				{
					v, err2 := strconv.ParseBool(keyRaw)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
					}
					key = v
				}
				var val []bool
				{
					val = make([]bool, len(valRaw))
					for i, rv := range valRaw {
						v, err2 := strconv.ParseBool(rv)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "array of booleans"))
						}
						val[i] = v
					}
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{true}))
			}
			if len(v) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[key]", v, len(v), 2, true))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodQueryMapBoolArrayBoolValidateMethodQueryMapBoolArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryPrimitiveStringValidateDecodeCode = `// DecodeMethodQueryPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveStringValidate
// MethodQueryPrimitiveStringValidate endpoint.
func DecodeMethodQueryPrimitiveStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   string
			err error
		)
		q = r.URL.Query().Get("q")
		if q == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if !(q == "val") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", q, []interface{}{"val"}))
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadQueryPrimitiveBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryPrimitiveBoolValidate
// MethodQueryPrimitiveBoolValidate endpoint.
func DecodeMethodQueryPrimitiveBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   bool
			err error
		)
		{
			qRaw := r.URL.Query().Get("q")
			if qRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			v, err2 := strconv.ParseBool(qRaw)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "boolean"))
			}
			q = v
		}
		if !(q == true) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", q, []interface{}{true}))
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadQueryPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodQueryPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveArrayStringValidate
// MethodQueryPrimitiveArrayStringValidate endpoint.
func DecodeMethodQueryPrimitiveArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []string
			err error
		)
		q = r.URL.Query()["q"]
		if q == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if !(e == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadQueryPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveArrayBoolValidate
// MethodQueryPrimitiveArrayBoolValidate endpoint.
func DecodeMethodQueryPrimitiveArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   []bool
			err error
		)
		{
			qRaw := r.URL.Query()["q"]
			if qRaw == nil {
				return goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make([]bool, len(qRaw))
			for i, rv := range qRaw {
				v, err2 := strconv.ParseBool(rv)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("q", qRaw, "array of booleans"))
				}
				q[i] = v
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for _, e := range q {
			if !(e == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadQueryPrimitiveMapStringArrayStringValidateDecodeCode = `// DecodeMethodQueryPrimitiveMapStringArrayStringValidateRequest returns a
// decoder for requests sent to the
// ServiceQueryPrimitiveMapStringArrayStringValidate
// MethodQueryPrimitiveMapStringArrayStringValidate endpoint.
func DecodeMethodQueryPrimitiveMapStringArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string][]string
			err error
		)
		q = r.URL.Query()
		if len(q) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			err = goa.MergeErrors(err, goa.ValidatePattern("q.key", k, "key"))
			if len(v) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[key]", v, len(v), 2, true))
			}
			for _, e := range v {
				err = goa.MergeErrors(err, goa.ValidatePattern("q[key][*]", e, "val"))
			}
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadQueryPrimitiveMapStringBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveMapStringBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveMapStringBoolValidate
// MethodQueryPrimitiveMapStringBoolValidate endpoint.
func DecodeMethodQueryPrimitiveMapStringBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[string]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[string]bool, len(qRaw))
			for key, va := range qRaw {
				var val bool
				{
					valRaw := va[0]
					v, err2 := strconv.ParseBool(valRaw)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "boolean"))
					}
					val = v
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			err = goa.MergeErrors(err, goa.ValidatePattern("q.key", k, "key"))
			if !(v == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[key]", v, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadQueryPrimitiveMapBoolArrayBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveMapBoolArrayBoolValidateRequest returns a decoder
// for requests sent to the ServiceQueryPrimitiveMapBoolArrayBoolValidate
// MethodQueryPrimitiveMapBoolArrayBoolValidate endpoint.
func DecodeMethodQueryPrimitiveMapBoolArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   map[bool][]bool
			err error
		)
		{
			qRaw := r.URL.Query()
			if len(qRaw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
			}
			q = make(map[bool][]bool, len(qRaw))
			for keyRaw, valRaw := range qRaw {
				var key bool
				{
					v, err2 := strconv.ParseBool(keyRaw)
					if err2 != nil {
						err = goa.MergeErrors(err, goa.InvalidFieldTypeError("key", keyRaw, "boolean"))
					}
					key = v
				}
				var val []bool
				{
					val = make([]bool, len(valRaw))
					for i, rv := range valRaw {
						v, err2 := strconv.ParseBool(rv)
						if err2 != nil {
							err = goa.MergeErrors(err, goa.InvalidFieldTypeError("val", valRaw, "array of booleans"))
						}
						val[i] = v
					}
				}
				q[key] = val
			}
		}
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}
		for k, v := range q {
			if !(k == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q.key", k, []interface{}{true}))
			}
			if len(v) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[key]", v, len(v), 2, true))
			}
			for _, e := range v {
				if !(e == false) {
					err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[key][*]", e, []interface{}{false}))
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadQueryStringMappedDecodeCode = `// DecodeMethodQueryStringMappedRequest returns a decoder for requests sent to
// the ServiceQueryStringMapped MethodQueryStringMapped endpoint.
func DecodeMethodQueryStringMappedRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			query *string
		)
		queryRaw := r.URL.Query().Get("q")
		if queryRaw != "" {
			query = &queryRaw
		}

		return NewMethodQueryStringMappedMethodQueryStringMappedPayload(query), nil
	}
}
`

var PayloadQueryStringDefaultDecodeCode = `// DecodeMethodQueryStringDefaultRequest returns a decoder for requests sent to
// the ServiceQueryStringDefault MethodQueryStringDefault endpoint.
func DecodeMethodQueryStringDefaultRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q string
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = qRaw
		}

		return NewMethodQueryStringDefaultMethodQueryStringDefaultPayload(q), nil
	}
}
`

var PayloadQueryPrimitiveStringDefaultDecodeCode = `// DecodeMethodQueryPrimitiveStringDefaultRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveStringDefault
// MethodQueryPrimitiveStringDefault endpoint.
func DecodeMethodQueryPrimitiveStringDefaultRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			q   string
			err error
		)
		q = r.URL.Query().Get("q")
		if q == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("q", "query string"))
		}
		if err != nil {
			return nil, err
		}

		return q, nil
	}
}
`

var PayloadPathStringDecodeCode = `// DecodeMethodPathStringRequest returns a decoder for requests sent to the
// ServicePathString MethodPathString endpoint.
func DecodeMethodPathStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p string

			params = mux.Vars(r)
		)
		p = params["p"]

		return NewMethodPathStringMethodPathStringPayload(p), nil
	}
}
`

var PayloadPathStringValidateDecodeCode = `// DecodeMethodPathStringValidateRequest returns a decoder for requests sent to
// the ServicePathStringValidate MethodPathStringValidate endpoint.
func DecodeMethodPathStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p   string
			err error

			params = mux.Vars(r)
		)
		p = params["p"]
		if !(p == "val") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", p, []interface{}{"val"}))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodPathStringValidateMethodPathStringValidatePayload(p), nil
	}
}
`

var PayloadPathArrayStringDecodeCode = `// DecodeMethodPathArrayStringRequest returns a decoder for requests sent to
// the ServicePathArrayString MethodPathArrayString endpoint.
func DecodeMethodPathArrayStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p []string

			params = mux.Vars(r)
		)
		{
			pRaw := params["p"]
			pRawSlice := strings.Split(pRaw, ",")
			p = make([]string, len(pRawSlice))
			for i, rv := range pRawSlice {
				p[i] = rv
			}
		}

		return NewMethodPathArrayStringMethodPathArrayStringPayload(p), nil
	}
}
`

var PayloadPathArrayStringValidateDecodeCode = `// DecodeMethodPathArrayStringValidateRequest returns a decoder for requests
// sent to the ServicePathArrayStringValidate MethodPathArrayStringValidate
// endpoint.
func DecodeMethodPathArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p   []string
			err error

			params = mux.Vars(r)
		)
		{
			pRaw := params["p"]
			pRawSlice := strings.Split(pRaw, ",")
			p = make([]string, len(pRawSlice))
			for i, rv := range pRawSlice {
				p[i] = rv
			}
		}
		if !(p == []string{"val"}) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", p, []interface{}{[]string{"val"}}))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodPathArrayStringValidateMethodPathArrayStringValidatePayload(p), nil
	}
}
`

var PayloadPathPrimitiveStringValidateDecodeCode = `// DecodeMethodPathPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveStringValidate
// MethodPathPrimitiveStringValidate endpoint.
func DecodeMethodPathPrimitiveStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p   string
			err error

			params = mux.Vars(r)
		)
		p = params["p"]
		if !(p == "val") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", p, []interface{}{"val"}))
		}
		if err != nil {
			return nil, err
		}

		return p, nil
	}
}
`

var PayloadPathPrimitiveBoolValidateDecodeCode = `// DecodeMethodPathPrimitiveBoolValidateRequest returns a decoder for requests
// sent to the ServicePathPrimitiveBoolValidate MethodPathPrimitiveBoolValidate
// endpoint.
func DecodeMethodPathPrimitiveBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p   bool
			err error

			params = mux.Vars(r)
		)
		{
			pRaw := params["p"]
			v, err2 := strconv.ParseBool(pRaw)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("p", pRaw, "boolean"))
			}
			p = v
		}
		if !(p == true) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", p, []interface{}{true}))
		}
		if err != nil {
			return nil, err
		}

		return p, nil
	}
}
`

var PayloadPathPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodPathPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveArrayStringValidate
// MethodPathPrimitiveArrayStringValidate endpoint.
func DecodeMethodPathPrimitiveArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p   []string
			err error

			params = mux.Vars(r)
		)
		{
			pRaw := params["p"]
			pRawSlice := strings.Split(pRaw, ",")
			p = make([]string, len(pRawSlice))
			for i, rv := range pRawSlice {
				p[i] = rv
			}
		}
		if len(p) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("p", p, len(p), 1, true))
		}
		for _, e := range p {
			if !(e == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("p[*]", e, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return p, nil
	}
}
`

var PayloadPathPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodPathPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveArrayBoolValidate
// MethodPathPrimitiveArrayBoolValidate endpoint.
func DecodeMethodPathPrimitiveArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			p   []bool
			err error

			params = mux.Vars(r)
		)
		{
			pRaw := params["p"]
			pRawSlice := strings.Split(pRaw, ",")
			p = make([]bool, len(pRawSlice))
			for i, rv := range pRawSlice {
				v, err2 := strconv.ParseBool(rv)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("p", pRaw, "array of booleans"))
				}
				p[i] = v
			}
		}
		if len(p) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("p", p, len(p), 1, true))
		}
		for _, e := range p {
			if !(e == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("p[*]", e, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return p, nil
	}
}
`

var PayloadHeaderStringDecodeCode = `// DecodeMethodHeaderStringRequest returns a decoder for requests sent to the
// ServiceHeaderString MethodHeaderString endpoint.
func DecodeMethodHeaderStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h *string
		)
		hRaw := r.Header.Get("h")
		if hRaw != "" {
			h = &hRaw
		}

		return NewMethodHeaderStringMethodHeaderStringPayload(h), nil
	}
}
`

var PayloadHeaderStringValidateDecodeCode = `// DecodeMethodHeaderStringValidateRequest returns a decoder for requests sent
// to the ServiceHeaderStringValidate MethodHeaderStringValidate endpoint.
func DecodeMethodHeaderStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h   *string
			err error
		)
		hRaw := r.Header.Get("h")
		if hRaw != "" {
			h = &hRaw
		}
		if h != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("h", *h, "header"))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodHeaderStringValidateMethodHeaderStringValidatePayload(h), nil
	}
}
`

var PayloadHeaderArrayStringDecodeCode = `// DecodeMethodHeaderArrayStringRequest returns a decoder for requests sent to
// the ServiceHeaderArrayString MethodHeaderArrayString endpoint.
func DecodeMethodHeaderArrayStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h []string
		)
		h = r.Header["H"]

		return NewMethodHeaderArrayStringMethodHeaderArrayStringPayload(h), nil
	}
}
`

var PayloadHeaderArrayStringValidateDecodeCode = `// DecodeMethodHeaderArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceHeaderArrayStringValidate MethodHeaderArrayStringValidate
// endpoint.
func DecodeMethodHeaderArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h   []string
			err error
		)
		h = r.Header["H"]

		for _, e := range h {
			if !(e == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("h[*]", e, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodHeaderArrayStringValidateMethodHeaderArrayStringValidatePayload(h), nil
	}
}
`

var PayloadHeaderPrimitiveStringValidateDecodeCode = `// DecodeMethodHeaderPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveStringValidate
// MethodHeaderPrimitiveStringValidate endpoint.
func DecodeMethodHeaderPrimitiveStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h   string
			err error
		)
		h = r.Header.Get("h")
		if h == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("h", "header"))
		}
		if !(h == "val") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("h", h, []interface{}{"val"}))
		}
		if err != nil {
			return nil, err
		}

		return h, nil
	}
}
`

var PayloadHeaderPrimitiveBoolValidateDecodeCode = `// DecodeMethodHeaderPrimitiveBoolValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveBoolValidate
// MethodHeaderPrimitiveBoolValidate endpoint.
func DecodeMethodHeaderPrimitiveBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h   bool
			err error
		)
		{
			hRaw := r.Header.Get("h")
			if hRaw == "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("h", "header"))
			}
			v, err2 := strconv.ParseBool(hRaw)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("h", hRaw, "boolean"))
			}
			h = v
		}
		if !(h == true) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("h", h, []interface{}{true}))
		}
		if err != nil {
			return nil, err
		}

		return h, nil
	}
}
`

var PayloadHeaderPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodHeaderPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveArrayStringValidate
// MethodHeaderPrimitiveArrayStringValidate endpoint.
func DecodeMethodHeaderPrimitiveArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h   []string
			err error
		)
		h = r.Header["H"]

		if h == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("h", "header"))
		}
		if len(h) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("h", h, len(h), 1, true))
		}
		for _, e := range h {
			err = goa.MergeErrors(err, goa.ValidatePattern("h[*]", e, "val"))
		}
		if err != nil {
			return nil, err
		}

		return h, nil
	}
}
`

var PayloadHeaderPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodHeaderPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveArrayBoolValidate
// MethodHeaderPrimitiveArrayBoolValidate endpoint.
func DecodeMethodHeaderPrimitiveArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h   []bool
			err error
		)
		{
			hRaw := r.Header["H"]
			if hRaw == nil {
				err = goa.MergeErrors(err, goa.MissingFieldError("h", "header"))
			}
			h = make([]bool, len(hRaw))
			for i, rv := range hRaw {
				v, err2 := strconv.ParseBool(rv)
				if err2 != nil {
					err = goa.MergeErrors(err, goa.InvalidFieldTypeError("h", hRaw, "array of booleans"))
				}
				h[i] = v
			}
		}
		if len(h) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("h", h, len(h), 1, true))
		}
		for _, e := range h {
			if !(e == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("h[*]", e, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return h, nil
	}
}
`

var PayloadHeaderStringDefaultDecodeCode = `// DecodeMethodHeaderStringDefaultRequest returns a decoder for requests sent
// to the ServiceHeaderStringDefault MethodHeaderStringDefault endpoint.
func DecodeMethodHeaderStringDefaultRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h string
		)
		hRaw := r.Header.Get("h")
		if hRaw != "" {
			h = hRaw
		}

		return NewMethodHeaderStringDefaultMethodHeaderStringDefaultPayload(h), nil
	}
}
`

var PayloadHeaderPrimitiveStringDefaultDecodeCode = `// DecodeMethodHeaderPrimitiveStringDefaultRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveStringDefault
// MethodHeaderPrimitiveStringDefault endpoint.
func DecodeMethodHeaderPrimitiveStringDefaultRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			h   string
			err error
		)
		h = r.Header.Get("h")
		if h == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("h", "header"))
		}
		if err != nil {
			return nil, err
		}

		return h, nil
	}
}
`

var PayloadBodyStringDecodeCode = `// DecodeMethodBodyStringRequest returns a decoder for requests sent to the
// ServiceBodyString MethodBodyString endpoint.
func DecodeMethodBodyStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyStringRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		return NewMethodBodyStringMethodBodyStringPayload(&body), nil
	}
}
`

var PayloadBodyStringValidateDecodeCode = `// DecodeMethodBodyStringValidateRequest returns a decoder for requests sent to
// the ServiceBodyStringValidate MethodBodyStringValidate endpoint.
func DecodeMethodBodyStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyStringValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyStringValidateMethodBodyStringValidatePayload(&body), nil
	}
}
`

var PayloadBodyUserDecodeCode = `// DecodeMethodBodyUserRequest returns a decoder for requests sent to the
// ServiceBodyUser MethodBodyUser endpoint.
func DecodeMethodBodyUserRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyUserRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		return NewMethodBodyUserPayloadType(&body), nil
	}
}
`

var PayloadBodyUserValidateDecodeCode = `// DecodeMethodBodyUserValidateRequest returns a decoder for requests sent to
// the ServiceBodyUserValidate MethodBodyUserValidate endpoint.
func DecodeMethodBodyUserValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyUserValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyUserValidatePayloadType(&body), nil
	}
}
`

var PayloadBodyArrayStringDecodeCode = `// DecodeMethodBodyArrayStringRequest returns a decoder for requests sent to
// the ServiceBodyArrayString MethodBodyArrayString endpoint.
func DecodeMethodBodyArrayStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyArrayStringRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		return NewMethodBodyArrayStringMethodBodyArrayStringPayload(&body), nil
	}
}
`

var PayloadBodyArrayStringValidateDecodeCode = `// DecodeMethodBodyArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceBodyArrayStringValidate MethodBodyArrayStringValidate
// endpoint.
func DecodeMethodBodyArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyArrayStringValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyArrayStringValidateMethodBodyArrayStringValidatePayload(&body), nil
	}
}
`

var PayloadBodyArrayUserDecodeCode = `// DecodeMethodBodyArrayUserRequest returns a decoder for requests sent to the
// ServiceBodyArrayUser MethodBodyArrayUser endpoint.
func DecodeMethodBodyArrayUserRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyArrayUserRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyArrayUserMethodBodyArrayUserPayload(&body), nil
	}
}
`

var PayloadBodyArrayUserValidateDecodeCode = `// DecodeMethodBodyArrayUserValidateRequest returns a decoder for requests sent
// to the ServiceBodyArrayUserValidate MethodBodyArrayUserValidate endpoint.
func DecodeMethodBodyArrayUserValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyArrayUserValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyArrayUserValidateMethodBodyArrayUserValidatePayload(&body), nil
	}
}
`

var PayloadBodyMapStringDecodeCode = `// DecodeMethodBodyMapStringRequest returns a decoder for requests sent to the
// ServiceBodyMapString MethodBodyMapString endpoint.
func DecodeMethodBodyMapStringRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyMapStringRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		return NewMethodBodyMapStringMethodBodyMapStringPayload(&body), nil
	}
}
`

var PayloadBodyMapStringValidateDecodeCode = `// DecodeMethodBodyMapStringValidateRequest returns a decoder for requests sent
// to the ServiceBodyMapStringValidate MethodBodyMapStringValidate endpoint.
func DecodeMethodBodyMapStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyMapStringValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyMapStringValidateMethodBodyMapStringValidatePayload(&body), nil
	}
}
`

var PayloadBodyMapUserDecodeCode = `// DecodeMethodBodyMapUserRequest returns a decoder for requests sent to the
// ServiceBodyMapUser MethodBodyMapUser endpoint.
func DecodeMethodBodyMapUserRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyMapUserRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyMapUserMethodBodyMapUserPayload(&body), nil
	}
}
`

var PayloadBodyMapUserValidateDecodeCode = `// DecodeMethodBodyMapUserValidateRequest returns a decoder for requests sent
// to the ServiceBodyMapUserValidate MethodBodyMapUserValidate endpoint.
func DecodeMethodBodyMapUserValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyMapUserValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		return NewMethodBodyMapUserValidateMethodBodyMapUserValidatePayload(&body), nil
	}
}
`

var PayloadBodyPrimitiveStringValidateDecodeCode = `// DecodeMethodBodyPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveStringValidate
// MethodBodyPrimitiveStringValidate endpoint.
func DecodeMethodBodyPrimitiveStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		if body != nil {
			if !(*body == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("body", *body, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadBodyPrimitiveBoolValidateDecodeCode = `// DecodeMethodBodyPrimitiveBoolValidateRequest returns a decoder for requests
// sent to the ServiceBodyPrimitiveBoolValidate MethodBodyPrimitiveBoolValidate
// endpoint.
func DecodeMethodBodyPrimitiveBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body bool
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		if body != nil {
			if !(*body == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("body", *body, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadBodyPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodBodyPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayStringValidate
// MethodBodyPrimitiveArrayStringValidate endpoint.
func DecodeMethodBodyPrimitiveArrayStringValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body []string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		if len(body) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body", body, len(body), 1, true))
		}
		for _, e := range body {
			if !(e == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("body[*]", e, []interface{}{"val"}))
			}
		}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadBodyPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodBodyPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayBoolValidate
// MethodBodyPrimitiveArrayBoolValidate endpoint.
func DecodeMethodBodyPrimitiveArrayBoolValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body []bool
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		if len(body) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body", body, len(body), 1, true))
		}
		for _, e := range body {
			if !(e == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("body[*]", e, []interface{}{true}))
			}
		}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadBodyPrimitiveArrayUserValidateDecodeCode = `// DecodeMethodBodyPrimitiveArrayUserValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayUserValidate
// MethodBodyPrimitiveArrayUserValidate endpoint.
func DecodeMethodBodyPrimitiveArrayUserValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body []*PayloadTypeRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		if len(body) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body", body, len(body), 1, true))
		}
		for _, e := range body {
			if e != nil {
				if err2 := e.Validate(); err2 != nil {
					err = goa.MergeErrors(err, err2)
				}
			}
		}
		if err != nil {
			return nil, err
		}

		return NewMethodBodyPrimitiveArrayUserValidatePayloadType(body), nil
	}
}
`

var PayloadBodyPrimitiveFieldArrayUserDecodeCode = `// DecodeMethodBodyPrimitiveArrayUserRequest returns a decoder for requests
// sent to the ServiceBodyPrimitiveArrayUser MethodBodyPrimitiveArrayUser
// endpoint.
func DecodeMethodBodyPrimitiveArrayUserRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body []string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		return NewMethodBodyPrimitiveArrayUserPayloadType(body), nil
	}
}
`

var PayloadBodyPrimitiveFieldArrayUserValidateDecodeCode = `// DecodeMethodBodyPrimitiveArrayUserValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayUserValidate
// MethodBodyPrimitiveArrayUserValidate endpoint.
func DecodeMethodBodyPrimitiveArrayUserValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body []string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		if len(body) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body", body, len(body), 1, true))
		}
		for _, e := range body {
			err = goa.MergeErrors(err, goa.ValidatePattern("body[*]", e, "pattern"))
		}
		if err != nil {
			return nil, err
		}

		return NewMethodBodyPrimitiveArrayUserValidatePayloadType(body), nil
	}
}
`

var PayloadBodyQueryObjectDecodeCode = `// DecodeMethodBodyQueryObjectRequest returns a decoder for requests sent to
// the ServiceBodyQueryObject MethodBodyQueryObject endpoint.
func DecodeMethodBodyQueryObjectRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryObjectRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		var (
			b *string
		)
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewMethodBodyQueryObjectMethodBodyQueryObjectPayload(&body, b), nil
	}
}
`

var PayloadBodyQueryObjectValidateDecodeCode = `// DecodeMethodBodyQueryObjectValidateRequest returns a decoder for requests
// sent to the ServiceBodyQueryObjectValidate MethodBodyQueryObjectValidate
// endpoint.
func DecodeMethodBodyQueryObjectValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryObjectValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		var (
			b string
		)
		b = r.URL.Query().Get("b")
		if b == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "query string"))
		}
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))
		if err != nil {
			return nil, err
		}

		return NewMethodBodyQueryObjectValidateMethodBodyQueryObjectValidatePayload(&body, b), nil
	}
}
`

var PayloadBodyQueryUserDecodeCode = `// DecodeMethodBodyQueryUserRequest returns a decoder for requests sent to the
// ServiceBodyQueryUser MethodBodyQueryUser endpoint.
func DecodeMethodBodyQueryUserRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryUserRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		var (
			b *string
		)
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewMethodBodyQueryUserPayloadType(&body, b), nil
	}
}
`

var PayloadBodyQueryUserValidateDecodeCode = `// DecodeMethodBodyQueryUserValidateRequest returns a decoder for requests sent
// to the ServiceBodyQueryUserValidate MethodBodyQueryUserValidate endpoint.
func DecodeMethodBodyQueryUserValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryUserValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		var (
			b string
		)
		b = r.URL.Query().Get("b")
		if b == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "query string"))
		}
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))
		if err != nil {
			return nil, err
		}

		return NewMethodBodyQueryUserValidatePayloadType(&body, b), nil
	}
}
`

var PayloadBodyPathObjectDecodeCode = `// DecodeMethodBodyPathObjectRequest returns a decoder for requests sent to the
// ServiceBodyPathObject MethodBodyPathObject endpoint.
func DecodeMethodBodyPathObjectRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyPathObjectRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		var (
			b string

			params = mux.Vars(r)
		)
		b = params["b"]

		return NewMethodBodyPathObjectMethodBodyPathObjectPayload(&body, b), nil
	}
}
`

var PayloadBodyPathObjectValidateDecodeCode = `// DecodeMethodBodyPathObjectValidateRequest returns a decoder for requests
// sent to the ServiceBodyPathObjectValidate MethodBodyPathObjectValidate
// endpoint.
func DecodeMethodBodyPathObjectValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyPathObjectValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		var (
			b string

			params = mux.Vars(r)
		)
		b = params["b"]
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))
		if err != nil {
			return nil, err
		}

		return NewMethodBodyPathObjectValidateMethodBodyPathObjectValidatePayload(&body, b), nil
	}
}
`

var PayloadBodyPathUserDecodeCode = `// DecodeMethodBodyPathUserRequest returns a decoder for requests sent to the
// ServiceBodyPathUser MethodBodyPathUser endpoint.
func DecodeMethodBodyPathUserRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyPathUserRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		var (
			b string

			params = mux.Vars(r)
		)
		b = params["b"]

		return NewMethodBodyPathUserPayloadType(&body, b), nil
	}
}
`

var PayloadBodyPathUserValidateDecodeCode = `// DecodeMethodUserBodyPathValidateRequest returns a decoder for requests sent
// to the ServiceBodyPathUserValidate MethodUserBodyPathValidate endpoint.
func DecodeMethodUserBodyPathValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodUserBodyPathValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		var (
			b string

			params = mux.Vars(r)
		)
		b = params["b"]
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))
		if err != nil {
			return nil, err
		}

		return NewMethodUserBodyPathValidatePayloadType(&body, b), nil
	}
}
`

var PayloadBodyQueryPathObjectDecodeCode = `// DecodeMethodBodyQueryPathObjectRequest returns a decoder for requests sent
// to the ServiceBodyQueryPathObject MethodBodyQueryPathObject endpoint.
func DecodeMethodBodyQueryPathObjectRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryPathObjectRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		var (
			c string
			b *string

			params = mux.Vars(r)
		)
		c = params["c"]
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewMethodBodyQueryPathObjectMethodBodyQueryPathObjectPayload(&body, c, b), nil
	}
}
`

var PayloadBodyQueryPathObjectValidateDecodeCode = `// DecodeMethodBodyQueryPathObjectValidateRequest returns a decoder for
// requests sent to the ServiceBodyQueryPathObjectValidate
// MethodBodyQueryPathObjectValidate endpoint.
func DecodeMethodBodyQueryPathObjectValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryPathObjectValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		var (
			c string
			b string

			params = mux.Vars(r)
		)
		c = params["c"]
		err = goa.MergeErrors(err, goa.ValidatePattern("c", c, "patternc"))
		b = r.URL.Query().Get("b")
		if b == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "query string"))
		}
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))
		if err != nil {
			return nil, err
		}

		return NewMethodBodyQueryPathObjectValidateMethodBodyQueryPathObjectValidatePayload(&body, c, b), nil
	}
}
`

var PayloadBodyQueryPathUserDecodeCode = `// DecodeMethodBodyQueryPathUserRequest returns a decoder for requests sent to
// the ServiceBodyQueryPathUser MethodBodyQueryPathUser endpoint.
func DecodeMethodBodyQueryPathUserRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryPathUserRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		var (
			c string
			b *string

			params = mux.Vars(r)
		)
		c = params["c"]
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewMethodBodyQueryPathUserPayloadType(&body, c, b), nil
	}
}
`

var PayloadBodyQueryPathUserValidateDecodeCode = `// DecodeMethodBodyQueryPathUserValidateRequest returns a decoder for requests
// sent to the ServiceBodyQueryPathUserValidate MethodBodyQueryPathUserValidate
// endpoint.
func DecodeMethodBodyQueryPathUserValidateRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MethodBodyQueryPathUserValidateRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = body.Validate()
		if err != nil {
			return nil, err
		}

		var (
			c string
			b string

			params = mux.Vars(r)
		)
		c = params["c"]
		err = goa.MergeErrors(err, goa.ValidatePattern("c", c, "patternc"))
		b = r.URL.Query().Get("b")
		if b == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "query string"))
		}
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))
		if err != nil {
			return nil, err
		}

		return NewMethodBodyQueryPathUserValidatePayloadType(&body, c, b), nil
	}
}
`
