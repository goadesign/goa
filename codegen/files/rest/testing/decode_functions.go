package testing

var PayloadQueryBoolDecodeCode = `// DecodeEndpointQueryBoolRequest returns a decoder for requests sent to the
// ServiceQueryBool EndpointQueryBool endpoint.
func DecodeEndpointQueryBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryBoolPayload, error) {
		var (
			q *bool
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseBool(qRaw)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "boolean")
			}
			q = &v
		}

		return NewEndpointQueryBoolPayload(q), nil
	}
}
`

var PayloadQueryBoolValidateDecodeCode = `// DecodeEndpointQueryBoolValidateRequest returns a decoder for requests sent
// to the ServiceQueryBoolValidate EndpointQueryBoolValidate endpoint.
func DecodeEndpointQueryBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryBoolValidatePayload, error) {
		var (
			q *bool
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseBool(qRaw)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "boolean")
			}
			q = &v
		}
		if q != nil {
			if !(*q == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", *q, []interface{}{true}))
			}
		}

		return NewEndpointQueryBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryIntDecodeCode = `// DecodeEndpointQueryIntRequest returns a decoder for requests sent to the
// ServiceQueryInt EndpointQueryInt endpoint.
func DecodeEndpointQueryIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryIntPayload, error) {
		var (
			q *int
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseInt(qRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
			}
			pv := int(v)
			q = &pv
		}

		return NewEndpointQueryIntPayload(q), nil
	}
}
`

var PayloadQueryIntValidateDecodeCode = `// DecodeEndpointQueryIntValidateRequest returns a decoder for requests sent to
// the ServiceQueryIntValidate EndpointQueryIntValidate endpoint.
func DecodeEndpointQueryIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryIntValidatePayload, error) {
		var (
			q *int
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseInt(qRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
			}
			pv := int(v)
			q = &pv
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryIntValidatePayload(q), nil
	}
}
`

var PayloadQueryInt32DecodeCode = `// DecodeEndpointQueryInt32Request returns a decoder for requests sent to the
// ServiceQueryInt32 EndpointQueryInt32 endpoint.
func DecodeEndpointQueryInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryInt32Payload, error) {
		var (
			q *int32
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseInt(qRaw, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
			}
			pv := int32(v)
			q = &pv
		}

		return NewEndpointQueryInt32Payload(q), nil
	}
}
`

var PayloadQueryInt32ValidateDecodeCode = `// DecodeEndpointQueryInt32ValidateRequest returns a decoder for requests sent
// to the ServiceQueryInt32Validate EndpointQueryInt32Validate endpoint.
func DecodeEndpointQueryInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryInt32ValidatePayload, error) {
		var (
			q *int32
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseInt(qRaw, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
			}
			pv := int32(v)
			q = &pv
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryInt64DecodeCode = `// DecodeEndpointQueryInt64Request returns a decoder for requests sent to the
// ServiceQueryInt64 EndpointQueryInt64 endpoint.
func DecodeEndpointQueryInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryInt64Payload, error) {
		var (
			q *int64
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseInt(qRaw, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
			}
			q = &v
		}

		return NewEndpointQueryInt64Payload(q), nil
	}
}
`

var PayloadQueryInt64ValidateDecodeCode = `// DecodeEndpointQueryInt64ValidateRequest returns a decoder for requests sent
// to the ServiceQueryInt64Validate EndpointQueryInt64Validate endpoint.
func DecodeEndpointQueryInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryInt64ValidatePayload, error) {
		var (
			q *int64
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseInt(qRaw, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
			}
			q = &v
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryUIntDecodeCode = `// DecodeEndpointQueryUIntRequest returns a decoder for requests sent to the
// ServiceQueryUInt EndpointQueryUInt endpoint.
func DecodeEndpointQueryUIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryUIntPayload, error) {
		var (
			q *uint
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseUint(qRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
			}
			pv := uint(v)
			q = &pv
		}

		return NewEndpointQueryUIntPayload(q), nil
	}
}
`

var PayloadQueryUIntValidateDecodeCode = `// DecodeEndpointQueryUIntValidateRequest returns a decoder for requests sent
// to the ServiceQueryUIntValidate EndpointQueryUIntValidate endpoint.
func DecodeEndpointQueryUIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryUIntValidatePayload, error) {
		var (
			q *uint
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseUint(qRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
			}
			pv := uint(v)
			q = &pv
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryUIntValidatePayload(q), nil
	}
}
`

var PayloadQueryUInt32DecodeCode = `// DecodeEndpointQueryUInt32Request returns a decoder for requests sent to the
// ServiceQueryUInt32 EndpointQueryUInt32 endpoint.
func DecodeEndpointQueryUInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryUInt32Payload, error) {
		var (
			q *uint32
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseUint(qRaw, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
			}
			pv := uint32(v)
			q = &pv
		}

		return NewEndpointQueryUInt32Payload(q), nil
	}
}
`

var PayloadQueryUInt32ValidateDecodeCode = `// DecodeEndpointQueryUInt32ValidateRequest returns a decoder for requests sent
// to the ServiceQueryUInt32Validate EndpointQueryUInt32Validate endpoint.
func DecodeEndpointQueryUInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryUInt32ValidatePayload, error) {
		var (
			q *uint32
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseUint(qRaw, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
			}
			pv := uint32(v)
			q = &pv
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryUInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryUInt64DecodeCode = `// DecodeEndpointQueryUInt64Request returns a decoder for requests sent to the
// ServiceQueryUInt64 EndpointQueryUInt64 endpoint.
func DecodeEndpointQueryUInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryUInt64Payload, error) {
		var (
			q *uint64
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseUint(qRaw, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
			}
			q = &v
		}

		return NewEndpointQueryUInt64Payload(q), nil
	}
}
`

var PayloadQueryUInt64ValidateDecodeCode = `// DecodeEndpointQueryUInt64ValidateRequest returns a decoder for requests sent
// to the ServiceQueryUInt64Validate EndpointQueryUInt64Validate endpoint.
func DecodeEndpointQueryUInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryUInt64ValidatePayload, error) {
		var (
			q *uint64
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseUint(qRaw, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
			}
			q = &v
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryUInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryFloat32DecodeCode = `// DecodeEndpointQueryFloat32Request returns a decoder for requests sent to the
// ServiceQueryFloat32 EndpointQueryFloat32 endpoint.
func DecodeEndpointQueryFloat32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryFloat32Payload, error) {
		var (
			q *float32
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseFloat(qRaw, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "float")
			}
			pv := float32(v)
			q = &pv
		}

		return NewEndpointQueryFloat32Payload(q), nil
	}
}
`

var PayloadQueryFloat32ValidateDecodeCode = `// DecodeEndpointQueryFloat32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryFloat32Validate EndpointQueryFloat32Validate
// endpoint.
func DecodeEndpointQueryFloat32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryFloat32ValidatePayload, error) {
		var (
			q *float32
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseFloat(qRaw, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "float")
			}
			pv := float32(v)
			q = &pv
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryFloat32ValidatePayload(q), nil
	}
}
`

var PayloadQueryFloat64DecodeCode = `// DecodeEndpointQueryFloat64Request returns a decoder for requests sent to the
// ServiceQueryFloat64 EndpointQueryFloat64 endpoint.
func DecodeEndpointQueryFloat64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryFloat64Payload, error) {
		var (
			q *float64
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseFloat(qRaw, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "float")
			}
			q = &v
		}

		return NewEndpointQueryFloat64Payload(q), nil
	}
}
`

var PayloadQueryFloat64ValidateDecodeCode = `// DecodeEndpointQueryFloat64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryFloat64Validate EndpointQueryFloat64Validate
// endpoint.
func DecodeEndpointQueryFloat64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryFloat64ValidatePayload, error) {
		var (
			q *float64
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseFloat(qRaw, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "float")
			}
			q = &v
		}
		if q != nil {
			if *q < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q", *q, 1, true))
			}
		}

		return NewEndpointQueryFloat64ValidatePayload(q), nil
	}
}
`

var PayloadQueryStringDecodeCode = `// DecodeEndpointQueryStringRequest returns a decoder for requests sent to the
// ServiceQueryString EndpointQueryString endpoint.
func DecodeEndpointQueryStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryStringPayload, error) {
		var (
			q *string
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = &qRaw
		}

		return NewEndpointQueryStringPayload(q), nil
	}
}
`

var PayloadQueryStringValidateDecodeCode = `// DecodeEndpointQueryStringValidateRequest returns a decoder for requests sent
// to the ServiceQueryStringValidate EndpointQueryStringValidate endpoint.
func DecodeEndpointQueryStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryStringValidatePayload, error) {
		var (
			q *string
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = &qRaw
		}
		if q != nil {
			if !(*q == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", *q, []interface{}{"val"}))
			}
		}

		return NewEndpointQueryStringValidatePayload(q), nil
	}
}
`

var PayloadQueryBytesDecodeCode = `// DecodeEndpointQueryBytesRequest returns a decoder for requests sent to the
// ServiceQueryBytes EndpointQueryBytes endpoint.
func DecodeEndpointQueryBytesRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryBytesPayload, error) {
		var (
			q []byte
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = []byte(qRaw)
		}

		return NewEndpointQueryBytesPayload(q), nil
	}
}
`

var PayloadQueryBytesValidateDecodeCode = `// DecodeEndpointQueryBytesValidateRequest returns a decoder for requests sent
// to the ServiceQueryBytesValidate EndpointQueryBytesValidate endpoint.
func DecodeEndpointQueryBytesValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryBytesValidatePayload, error) {
		var (
			q []byte
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = []byte(qRaw)
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}

		return NewEndpointQueryBytesValidatePayload(q), nil
	}
}
`

var PayloadQueryAnyDecodeCode = `// DecodeEndpointQueryAnyRequest returns a decoder for requests sent to the
// ServiceQueryAny EndpointQueryAny endpoint.
func DecodeEndpointQueryAnyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryAnyPayload, error) {
		var (
			q interface{}
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = qRaw
		}

		return NewEndpointQueryAnyPayload(q), nil
	}
}
`

var PayloadQueryAnyValidateDecodeCode = `// DecodeEndpointQueryAnyValidateRequest returns a decoder for requests sent to
// the ServiceQueryAnyValidate EndpointQueryAnyValidate endpoint.
func DecodeEndpointQueryAnyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryAnyValidatePayload, error) {
		var (
			q interface{}
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = qRaw
		}
		if q != nil {
			if !(q == "val" || q == 1) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", q, []interface{}{"val", 1}))
			}
		}

		return NewEndpointQueryAnyValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBoolDecodeCode = `// DecodeEndpointQueryArrayBoolRequest returns a decoder for requests sent to
// the ServiceQueryArrayBool EndpointQueryArrayBool endpoint.
func DecodeEndpointQueryArrayBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayBoolPayload, error) {
		var (
			q []bool
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]bool, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseBool(rv)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of booleans")
				}
				q[i] = v
			}
		}

		return NewEndpointQueryArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryArrayBoolValidateDecodeCode = `// DecodeEndpointQueryArrayBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayBoolValidate EndpointQueryArrayBoolValidate
// endpoint.
func DecodeEndpointQueryArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayBoolValidatePayload, error) {
		var (
			q []bool
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]bool, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseBool(rv)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of booleans")
				}
				q[i] = v
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if !(e == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{true}))
			}
		}

		return NewEndpointQueryArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayIntDecodeCode = `// DecodeEndpointQueryArrayIntRequest returns a decoder for requests sent to
// the ServiceQueryArrayInt EndpointQueryArrayInt endpoint.
func DecodeEndpointQueryArrayIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayIntPayload, error) {
		var (
			q []int
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]int, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseInt(rv, 10, strconv.IntSize)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
				}
				q[i] = int(v)
			}
		}

		return NewEndpointQueryArrayIntPayload(q), nil
	}
}
`

var PayloadQueryArrayIntValidateDecodeCode = `// DecodeEndpointQueryArrayIntValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayIntValidate EndpointQueryArrayIntValidate
// endpoint.
func DecodeEndpointQueryArrayIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayIntValidatePayload, error) {
		var (
			q []int
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]int, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseInt(rv, 10, strconv.IntSize)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
				}
				q[i] = int(v)
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt32DecodeCode = `// DecodeEndpointQueryArrayInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayInt32 EndpointQueryArrayInt32 endpoint.
func DecodeEndpointQueryArrayInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayInt32Payload, error) {
		var (
			q []int32
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]int32, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseInt(rv, 10, 32)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
				}
				q[i] = int32(v)
			}
		}

		return NewEndpointQueryArrayInt32Payload(q), nil
	}
}
`

var PayloadQueryArrayInt32ValidateDecodeCode = `// DecodeEndpointQueryArrayInt32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayInt32Validate EndpointQueryArrayInt32Validate
// endpoint.
func DecodeEndpointQueryArrayInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayInt32ValidatePayload, error) {
		var (
			q []int32
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]int32, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseInt(rv, 10, 32)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
				}
				q[i] = int32(v)
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt64DecodeCode = `// DecodeEndpointQueryArrayInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayInt64 EndpointQueryArrayInt64 endpoint.
func DecodeEndpointQueryArrayInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayInt64Payload, error) {
		var (
			q []int64
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]int64, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseInt(rv, 10, 64)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
				}
				q[i] = v
			}
		}

		return NewEndpointQueryArrayInt64Payload(q), nil
	}
}
`

var PayloadQueryArrayInt64ValidateDecodeCode = `// DecodeEndpointQueryArrayInt64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayInt64Validate EndpointQueryArrayInt64Validate
// endpoint.
func DecodeEndpointQueryArrayInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayInt64ValidatePayload, error) {
		var (
			q []int64
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]int64, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseInt(rv, 10, 64)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
				}
				q[i] = v
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUIntDecodeCode = `// DecodeEndpointQueryArrayUIntRequest returns a decoder for requests sent to
// the ServiceQueryArrayUInt EndpointQueryArrayUInt endpoint.
func DecodeEndpointQueryArrayUIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUIntPayload, error) {
		var (
			q []uint
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]uint, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseUint(rv, 10, strconv.IntSize)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
				}
				q[i] = uint(v)
			}
		}

		return NewEndpointQueryArrayUIntPayload(q), nil
	}
}
`

var PayloadQueryArrayUIntValidateDecodeCode = `// DecodeEndpointQueryArrayUIntValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUIntValidate EndpointQueryArrayUIntValidate
// endpoint.
func DecodeEndpointQueryArrayUIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUIntValidatePayload, error) {
		var (
			q []uint
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]uint, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseUint(rv, 10, strconv.IntSize)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
				}
				q[i] = uint(v)
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayUIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt32DecodeCode = `// DecodeEndpointQueryArrayUInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt32 EndpointQueryArrayUInt32 endpoint.
func DecodeEndpointQueryArrayUInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUInt32Payload, error) {
		var (
			q []uint32
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]uint32, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseUint(rv, 10, 32)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
				}
				q[i] = int32(v)
			}
		}

		return NewEndpointQueryArrayUInt32Payload(q), nil
	}
}
`

var PayloadQueryArrayUInt32ValidateDecodeCode = `// DecodeEndpointQueryArrayUInt32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUInt32Validate EndpointQueryArrayUInt32Validate
// endpoint.
func DecodeEndpointQueryArrayUInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUInt32ValidatePayload, error) {
		var (
			q []uint32
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]uint32, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseUint(rv, 10, 32)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
				}
				q[i] = int32(v)
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayUInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt64DecodeCode = `// DecodeEndpointQueryArrayUInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt64 EndpointQueryArrayUInt64 endpoint.
func DecodeEndpointQueryArrayUInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUInt64Payload, error) {
		var (
			q []uint64
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]uint64, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseUint(rv, 10, 64)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
				}
				q[i] = v
			}
		}

		return NewEndpointQueryArrayUInt64Payload(q), nil
	}
}
`

var PayloadQueryArrayUInt64ValidateDecodeCode = `// DecodeEndpointQueryArrayUInt64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUInt64Validate EndpointQueryArrayUInt64Validate
// endpoint.
func DecodeEndpointQueryArrayUInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUInt64ValidatePayload, error) {
		var (
			q []uint64
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]uint64, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseUint(rv, 10, 64)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
				}
				q[i] = v
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayUInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat32DecodeCode = `// DecodeEndpointQueryArrayFloat32Request returns a decoder for requests sent
// to the ServiceQueryArrayFloat32 EndpointQueryArrayFloat32 endpoint.
func DecodeEndpointQueryArrayFloat32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayFloat32Payload, error) {
		var (
			q []float32
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]float32, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseFloat(rv, 32)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of floats")
				}
				q[i] = float32(v)
			}
		}

		return NewEndpointQueryArrayFloat32Payload(q), nil
	}
}
`

var PayloadQueryArrayFloat32ValidateDecodeCode = `// DecodeEndpointQueryArrayFloat32ValidateRequest returns a decoder for
// requests sent to the ServiceQueryArrayFloat32Validate
// EndpointQueryArrayFloat32Validate endpoint.
func DecodeEndpointQueryArrayFloat32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayFloat32ValidatePayload, error) {
		var (
			q []float32
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]float32, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseFloat(rv, 32)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of floats")
				}
				q[i] = float32(v)
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayFloat32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat64DecodeCode = `// DecodeEndpointQueryArrayFloat64Request returns a decoder for requests sent
// to the ServiceQueryArrayFloat64 EndpointQueryArrayFloat64 endpoint.
func DecodeEndpointQueryArrayFloat64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayFloat64Payload, error) {
		var (
			q []float64
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]float64, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseFloat(rv, 64)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of floats")
				}
				q[i] = v
			}
		}

		return NewEndpointQueryArrayFloat64Payload(q), nil
	}
}
`

var PayloadQueryArrayFloat64ValidateDecodeCode = `// DecodeEndpointQueryArrayFloat64ValidateRequest returns a decoder for
// requests sent to the ServiceQueryArrayFloat64Validate
// EndpointQueryArrayFloat64Validate endpoint.
func DecodeEndpointQueryArrayFloat64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayFloat64ValidatePayload, error) {
		var (
			q []float64
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]float64, len(qRaw))
			for i, rv := range qRaw {
				v, err := strconv.ParseFloat(rv, 64)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(qRaw, q, "array of floats")
				}
				q[i] = v
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if e < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("q[*]", e, 1, true))
			}
		}

		return NewEndpointQueryArrayFloat64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayStringDecodeCode = `// DecodeEndpointQueryArrayStringRequest returns a decoder for requests sent to
// the ServiceQueryArrayString EndpointQueryArrayString endpoint.
func DecodeEndpointQueryArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayStringPayload, error) {
		var (
			q []string
		)
		q = r.URL.Query()["q"]

		return NewEndpointQueryArrayStringPayload(q), nil
	}
}
`

var PayloadQueryArrayStringValidateDecodeCode = `// DecodeEndpointQueryArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayStringValidate EndpointQueryArrayStringValidate
// endpoint.
func DecodeEndpointQueryArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayStringValidatePayload, error) {
		var (
			q []string
		)
		q = r.URL.Query()["q"]
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if !(e == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{"val"}))
			}
		}

		return NewEndpointQueryArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBytesDecodeCode = `// DecodeEndpointQueryArrayBytesRequest returns a decoder for requests sent to
// the ServiceQueryArrayBytes EndpointQueryArrayBytes endpoint.
func DecodeEndpointQueryArrayBytesRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayBytesPayload, error) {
		var (
			q [][]byte
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([][]byte, len(qRaw))
			for i, rv := range qRaw {
				q[i] = []byte(rv)
			}
		}

		return NewEndpointQueryArrayBytesPayload(q), nil
	}
}
`

var PayloadQueryArrayBytesValidateDecodeCode = `// DecodeEndpointQueryArrayBytesValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayBytesValidate EndpointQueryArrayBytesValidate
// endpoint.
func DecodeEndpointQueryArrayBytesValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayBytesValidatePayload, error) {
		var (
			q [][]byte
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([][]byte, len(qRaw))
			for i, rv := range qRaw {
				q[i] = []byte(rv)
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if len(e) < 2 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q[*]", e, len(e), 2, true))
			}
		}

		return NewEndpointQueryArrayBytesValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayAnyDecodeCode = `// DecodeEndpointQueryArrayAnyRequest returns a decoder for requests sent to
// the ServiceQueryArrayAny EndpointQueryArrayAny endpoint.
func DecodeEndpointQueryArrayAnyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayAnyPayload, error) {
		var (
			q []interface{}
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]interface{}, len(qRaw))
			for i, rv := range qRaw {
				q[i] = rv
			}
		}

		return NewEndpointQueryArrayAnyPayload(q), nil
	}
}
`

var PayloadQueryArrayAnyValidateDecodeCode = `// DecodeEndpointQueryArrayAnyValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayAnyValidate EndpointQueryArrayAnyValidate
// endpoint.
func DecodeEndpointQueryArrayAnyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayAnyValidatePayload, error) {
		var (
			q []interface{}
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]interface{}, len(qRaw))
			for i, rv := range qRaw {
				q[i] = rv
			}
		}
		if q != nil {
			if len(q) < 1 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
			}
		}
		for _, e := range q {
			if !(e == "val" || e == 1) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("q[*]", e, []interface{}{"val", 1}))
			}
		}

		return NewEndpointQueryArrayAnyValidatePayload(q), nil
	}
}
`

var PayloadPathStringDecodeCode = `// DecodeEndpointPathStringRequest returns a decoder for requests sent to the
// ServicePathString EndpointPathString endpoint.
func DecodeEndpointPathStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointPathStringPayload, error) {
		var (
			p *string

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		if pRaw != "" {
			p = &pRaw
		}

		return NewEndpointPathStringPayload(p), nil
	}
}
`

var PayloadPathStringValidateDecodeCode = `// DecodeEndpointPathStringValidateRequest returns a decoder for requests sent
// to the ServicePathStringValidate EndpointPathStringValidate endpoint.
func DecodeEndpointPathStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointPathStringValidatePayload, error) {
		var (
			p *string

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		if pRaw != "" {
			p = &pRaw
		}
		if p != nil {
			if !(*p == "val") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", *p, []interface{}{"val"}))
			}
		}

		return NewEndpointPathStringValidatePayload(p), nil
	}
}
`

var PayloadPathArrayStringDecodeCode = `// DecodeEndpointPathArrayStringRequest returns a decoder for requests sent to
// the ServicePathArrayString EndpointPathArrayString endpoint.
func DecodeEndpointPathArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointPathArrayStringPayload, error) {
		var (
			p []string

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		if pRaw != "" {
			pRawSlice := strings.Split(pRaw, ",")
			p = make([]string, len(pRawSlice))
			for i, rv := range pRawSlice {
				p[i] = rv
			}
		}

		return NewEndpointPathArrayStringPayload(p), nil
	}
}
`

var PayloadPathArrayStringValidateDecodeCode = `// DecodeEndpointPathArrayStringValidateRequest returns a decoder for requests
// sent to the ServicePathArrayStringValidate EndpointPathArrayStringValidate
// endpoint.
func DecodeEndpointPathArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointPathArrayStringValidatePayload, error) {
		var (
			p []string

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		if pRaw != "" {
			pRawSlice := strings.Split(pRaw, ",")
			p = make([]string, len(pRawSlice))
			for i, rv := range pRawSlice {
				p[i] = rv
			}
		}
		if p != nil {
			if !(p == []string{"val"}) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", p, []interface{}{[]string{"val"}}))
			}
		}

		return NewEndpointPathArrayStringValidatePayload(p), nil
	}
}
`

var PayloadBodyStringDecodeCode = `// DecodeEndpointBodyStringRequest returns a decoder for requests sent to the
// ServiceBodyString EndpointBodyString endpoint.
func DecodeEndpointBodyStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyStringPayload, error) {
		var (
			body EndpointBodyStringPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyStringValidateDecodeCode = `// DecodeEndpointBodyStringValidateRequest returns a decoder for requests sent
// to the ServiceBodyStringValidate EndpointBodyStringValidate endpoint.
func DecodeEndpointBodyStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyStringValidatePayload, error) {
		var (
			body EndpointBodyStringValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyObjectDecodeCode = `// DecodeEndpointBodyObjectRequest returns a decoder for requests sent to the
// ServiceObjectBody EndpointBodyObject endpoint.
func DecodeEndpointBodyObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyObjectPayload, error) {
		var (
			body EndpointBodyObjectPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadObjectBodyValidateDecodeCode = `// DecodeEndpointBodyObjectValidateRequest returns a decoder for requests sent
// to the ServiceObjectBodyValidate EndpointBodyObjectValidate endpoint.
func DecodeEndpointBodyObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyObjectValidatePayload, error) {
		var (
			body EndpointBodyObjectValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyUserDecodeCode = `// DecodeEndpointBodyUserRequest returns a decoder for requests sent to the
// ServiceBodyUser EndpointBodyUser endpoint.
func DecodeEndpointBodyUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadUserBodyValidateDecodeCode = `// DecodeEndpointBodyUserValidateRequest returns a decoder for requests sent to
// the ServiceBodyUserValidate EndpointBodyUserValidate endpoint.
func DecodeEndpointBodyUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyArrayStringDecodeCode = `// DecodeEndpointBodyArrayStringRequest returns a decoder for requests sent to
// the ServiceBodyArrayString EndpointBodyArrayString endpoint.
func DecodeEndpointBodyArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyArrayStringPayload, error) {
		var (
			body EndpointBodyArrayStringPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyArrayStringValidateDecodeCode = `// DecodeEndpointBodyArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceBodyArrayStringValidate EndpointBodyArrayStringValidate
// endpoint.
func DecodeEndpointBodyArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyArrayStringValidatePayload, error) {
		var (
			body EndpointBodyArrayStringValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyArrayUserDecodeCode = `// DecodeEndpointBodyArrayUserRequest returns a decoder for requests sent to
// the ServiceBodyArrayUser EndpointBodyArrayUser endpoint.
func DecodeEndpointBodyArrayUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyArrayUserPayload, error) {
		var (
			body EndpointBodyArrayUserPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyArrayUserValidateDecodeCode = `// DecodeEndpointBodyArrayUserValidateRequest returns a decoder for requests
// sent to the ServiceBodyArrayUserValidate EndpointBodyArrayUserValidate
// endpoint.
func DecodeEndpointBodyArrayUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyArrayUserValidatePayload, error) {
		var (
			body EndpointBodyArrayUserValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyMapStringDecodeCode = `// DecodeEndpointBodyMapStringRequest returns a decoder for requests sent to
// the ServiceBodyMapString EndpointBodyMapString endpoint.
func DecodeEndpointBodyMapStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyMapStringPayload, error) {
		var (
			body EndpointBodyMapStringPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyMapStringValidateDecodeCode = `// DecodeEndpointBodyMapStringValidateRequest returns a decoder for requests
// sent to the ServiceBodyMapStringValidate EndpointBodyMapStringValidate
// endpoint.
func DecodeEndpointBodyMapStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyMapStringValidatePayload, error) {
		var (
			body EndpointBodyMapStringValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyMapUserDecodeCode = `// DecodeEndpointBodyMapUserRequest returns a decoder for requests sent to the
// ServiceBodyMapUser EndpointBodyMapUser endpoint.
func DecodeEndpointBodyMapUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyMapUserPayload, error) {
		var (
			body EndpointBodyMapUserPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyMapUserValidateDecodeCode = `// DecodeEndpointBodyMapUserValidateRequest returns a decoder for requests sent
// to the ServiceBodyMapUserValidate EndpointBodyMapUserValidate endpoint.
func DecodeEndpointBodyMapUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyMapUserValidatePayload, error) {
		var (
			body EndpointBodyMapUserValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyQueryObjectDecodeCode = `// DecodeEndpointBodyQueryObjectRequest returns a decoder for requests sent to
// the ServiceBodyQueryObject EndpointBodyQueryObject endpoint.
func DecodeEndpointBodyQueryObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyQueryObjectPayload, error) {
		var (
			body EndpointBodyQueryObjectServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		var (
			b *string
		)
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewEndpointBodyQueryObjectPayload(&body, b), nil
	}
}
`

var PayloadBodyQueryObjectValidateDecodeCode = `// DecodeEndpointBodyQueryObjectValidateRequest returns a decoder for requests
// sent to the ServiceBodyQueryObjectValidate EndpointBodyQueryObjectValidate
// endpoint.
func DecodeEndpointBodyQueryObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyQueryObjectValidatePayload, error) {
		var (
			body EndpointBodyQueryObjectValidateServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		var (
			b *string
		)
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}
		if b != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("b", *b, "patternb"))
		}

		return NewEndpointBodyQueryObjectValidatePayload(&body, b), nil
	}
}
`

var PayloadBodyQueryUserDecodeCode = `// DecodeEndpointBodyQueryUserRequest returns a decoder for requests sent to
// the ServiceBodyQueryUser EndpointBodyQueryUser endpoint.
func DecodeEndpointBodyQueryUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body EndpointBodyQueryUserServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		var (
			b *string
		)
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewPayloadType(&body, b), nil
	}
}
`

var PayloadBodyQueryUserValidateDecodeCode = `// DecodeEndpointBodyQueryUserValidateRequest returns a decoder for requests
// sent to the ServiceBodyQueryUserValidate EndpointBodyQueryUserValidate
// endpoint.
func DecodeEndpointBodyQueryUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body EndpointBodyQueryUserValidateServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		var (
			b *string
		)
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}
		if b != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("b", *b, "patternb"))
		}

		return NewPayloadType(&body, b), nil
	}
}
`

var PayloadBodyPathObjectDecodeCode = `// DecodeEndpointBodyPathObjectRequest returns a decoder for requests sent to
// the ServiceBodyPathObject EndpointBodyPathObject endpoint.
func DecodeEndpointBodyPathObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyPathObjectPayload, error) {
		var (
			body EndpointBodyPathObjectPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyPathObjectValidateDecodeCode = `// DecodeEndpointBodyPathObjectValidateRequest returns a decoder for requests
// sent to the ServiceBodyPathObjectValidate EndpointBodyPathObjectValidate
// endpoint.
func DecodeEndpointBodyPathObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyPathObjectValidatePayload, error) {
		var (
			body EndpointBodyPathObjectValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyPathUserDecodeCode = `// DecodeEndpointBodyPathUserRequest returns a decoder for requests sent to the
// ServiceBodyPathUser EndpointBodyPathUser endpoint.
func DecodeEndpointBodyPathUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyPathUserValidateDecodeCode = `// DecodeEndpointUserBodyPathValidateRequest returns a decoder for requests
// sent to the ServiceBodyPathUserValidate EndpointUserBodyPathValidate
// endpoint.
func DecodeEndpointUserBodyPathValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadBodyQueryPathObjectDecodeCode = `// DecodeEndpointBodyQueryPathObjectRequest returns a decoder for requests sent
// to the ServiceBodyQueryPathObject EndpointBodyQueryPathObject endpoint.
func DecodeEndpointBodyQueryPathObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyQueryPathObjectPayload, error) {
		var (
			body EndpointBodyQueryPathObjectServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		var (
			b *string
			c *string

			params = rest.ContextParams(r.Context())
		)
		cRaw := params["c"]
		if cRaw != "" {
			c = &cRaw
		}
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewEndpointBodyQueryPathObjectPayload(&body, b, c), nil
	}
}
`

var PayloadBodyQueryPathObjectValidateDecodeCode = `// DecodeEndpointBodyQueryPathObjectValidateRequest returns a decoder for
// requests sent to the ServiceBodyQueryPathObjectValidate
// EndpointBodyQueryPathObjectValidate endpoint.
func DecodeEndpointBodyQueryPathObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointBodyQueryPathObjectValidatePayload, error) {
		var (
			body EndpointBodyQueryPathObjectValidateServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		var (
			b *string
			c *string

			params = rest.ContextParams(r.Context())
		)
		cRaw := params["c"]
		if cRaw != "" {
			c = &cRaw
		}
		if c != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("c", *c, "patternc"))
		}
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}
		if b != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("b", *b, "patternb"))
		}

		return NewEndpointBodyQueryPathObjectValidatePayload(&body, b, c), nil
	}
}
`

var PayloadBodyQueryPathUserDecodeCode = `// DecodeEndpointBodyQueryPathUserRequest returns a decoder for requests sent
// to the ServiceBodyQueryPathUser EndpointBodyQueryPathUser endpoint.
func DecodeEndpointBodyQueryPathUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body EndpointBodyQueryPathUserServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		var (
			b *string
			c *string

			params = rest.ContextParams(r.Context())
		)
		cRaw := params["c"]
		if cRaw != "" {
			c = &cRaw
		}
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		return NewPayloadType(&body, b, c), nil
	}
}
`

var PayloadBodyQueryPathUserValidateDecodeCode = `// DecodeEndpointBodyQueryPathUserValidateRequest returns a decoder for
// requests sent to the ServiceBodyQueryPathUserValidate
// EndpointBodyQueryPathUserValidate endpoint.
func DecodeEndpointBodyQueryPathUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body EndpointBodyQueryPathUserValidateServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		var (
			b *string
			c *string

			params = rest.ContextParams(r.Context())
		)
		cRaw := params["c"]
		if cRaw != "" {
			c = &cRaw
		}
		if c != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("c", *c, "patternc"))
		}
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}
		if b != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("b", *b, "patternb"))
		}

		return NewPayloadType(&body, b, c), nil
	}
}
`
