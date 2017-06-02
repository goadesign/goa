package testing

var PayloadQueryBoolDecodeCode = `// DecodeEndpointQueryBoolRequest returns a decoder for requests sent to the
// ServiceQueryBool EndpointQueryBool endpoint.
func DecodeEndpointQueryBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryBoolPayload, error) {
		var (
			q   *bool
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseBool(qRaw)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "boolean")
			}
			q = &v
		}

		if err != nil {
			return nil, err
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
			q   bool
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseBool(qRaw)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "boolean")
		}
		q = v
		if !(q == true) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("q", q, []interface{}{true}))
		}

		if err != nil {
			return nil, err
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
			q   *int
			err error
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

		if err != nil {
			return nil, err
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
			q   int
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseInt(qRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
		}
		q = int(v)
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *int32
			err error
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

		if err != nil {
			return nil, err
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
			q   int32
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseInt(qRaw, 10, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
		}
		q = int32(v)
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *int64
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseInt(qRaw, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
			}
			q = &v
		}

		if err != nil {
			return nil, err
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
			q   int64
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseInt(qRaw, 10, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "integer")
		}
		q = v
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *uint
			err error
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

		if err != nil {
			return nil, err
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
			q   uint
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseUint(qRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
		}
		q = uint(v)
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *uint32
			err error
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

		if err != nil {
			return nil, err
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
			q   uint32
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseUint(qRaw, 10, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
		}
		q = uint32(v)
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *uint64
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseUint(qRaw, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
			}
			q = &v
		}

		if err != nil {
			return nil, err
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
			q   uint64
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseUint(qRaw, 10, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "unsigned integer")
		}
		q = v
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *float32
			err error
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

		if err != nil {
			return nil, err
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
			q   float32
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseFloat(qRaw, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "float")
		}
		q = float32(v)
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *float64
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			v, err := strconv.ParseFloat(qRaw, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "float")
			}
			q = &v
		}

		if err != nil {
			return nil, err
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
			q   float64
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseFloat(qRaw, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "float")
		}
		q = v
		if q < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("q", q, 1, true))
		}

		if err != nil {
			return nil, err
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
			q   *string
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = &qRaw
		}

		if err != nil {
			return nil, err
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
		return NewEndpointQueryStringValidatePayload(q), nil
	}
}
`

var PayloadQueryBytesDecodeCode = `// DecodeEndpointQueryBytesRequest returns a decoder for requests sent to the
// ServiceQueryBytes EndpointQueryBytes endpoint.
func DecodeEndpointQueryBytesRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryBytesPayload, error) {
		var (
			q   []byte
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = []byte(qRaw)
		}

		if err != nil {
			return nil, err
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
			q   []byte
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = []byte(qRaw)
		if len(q) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("q", q, len(q), 1, true))
		}

		if err != nil {
			return nil, err
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
			q   interface{}
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = qRaw
		}

		if err != nil {
			return nil, err
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
		return NewEndpointQueryAnyValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBoolDecodeCode = `// DecodeEndpointQueryArrayBoolRequest returns a decoder for requests sent to
// the ServiceQueryArrayBool EndpointQueryArrayBool endpoint.
func DecodeEndpointQueryArrayBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayBoolPayload, error) {
		var (
			q   []bool
			err error
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

		if err != nil {
			return nil, err
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
			q   []bool
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]bool, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseBool(rv)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of booleans")
			}
			q[i] = v
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
		return NewEndpointQueryArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayIntDecodeCode = `// DecodeEndpointQueryArrayIntRequest returns a decoder for requests sent to
// the ServiceQueryArrayInt EndpointQueryArrayInt endpoint.
func DecodeEndpointQueryArrayIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayIntPayload, error) {
		var (
			q   []int
			err error
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

		if err != nil {
			return nil, err
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
			q   []int
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]int, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseInt(rv, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
			}
			q[i] = int(v)
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
		return NewEndpointQueryArrayIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt32DecodeCode = `// DecodeEndpointQueryArrayInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayInt32 EndpointQueryArrayInt32 endpoint.
func DecodeEndpointQueryArrayInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayInt32Payload, error) {
		var (
			q   []int32
			err error
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

		if err != nil {
			return nil, err
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
			q   []int32
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]int32, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseInt(rv, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
			}
			q[i] = int32(v)
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
		return NewEndpointQueryArrayInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt64DecodeCode = `// DecodeEndpointQueryArrayInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayInt64 EndpointQueryArrayInt64 endpoint.
func DecodeEndpointQueryArrayInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayInt64Payload, error) {
		var (
			q   []int64
			err error
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

		if err != nil {
			return nil, err
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
			q   []int64
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]int64, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseInt(rv, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of integers")
			}
			q[i] = v
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
		return NewEndpointQueryArrayInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUIntDecodeCode = `// DecodeEndpointQueryArrayUIntRequest returns a decoder for requests sent to
// the ServiceQueryArrayUInt EndpointQueryArrayUInt endpoint.
func DecodeEndpointQueryArrayUIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUIntPayload, error) {
		var (
			q   []uint
			err error
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

		if err != nil {
			return nil, err
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
			q   []uint
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]uint, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseUint(rv, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
			}
			q[i] = uint(v)
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
		return NewEndpointQueryArrayUIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt32DecodeCode = `// DecodeEndpointQueryArrayUInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt32 EndpointQueryArrayUInt32 endpoint.
func DecodeEndpointQueryArrayUInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUInt32Payload, error) {
		var (
			q   []uint32
			err error
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

		if err != nil {
			return nil, err
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
			q   []uint32
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]uint32, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseUint(rv, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
			}
			q[i] = int32(v)
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
		return NewEndpointQueryArrayUInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt64DecodeCode = `// DecodeEndpointQueryArrayUInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt64 EndpointQueryArrayUInt64 endpoint.
func DecodeEndpointQueryArrayUInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayUInt64Payload, error) {
		var (
			q   []uint64
			err error
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

		if err != nil {
			return nil, err
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
			q   []uint64
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]uint64, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseUint(rv, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of unsigned integers")
			}
			q[i] = v
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
		return NewEndpointQueryArrayUInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat32DecodeCode = `// DecodeEndpointQueryArrayFloat32Request returns a decoder for requests sent
// to the ServiceQueryArrayFloat32 EndpointQueryArrayFloat32 endpoint.
func DecodeEndpointQueryArrayFloat32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayFloat32Payload, error) {
		var (
			q   []float32
			err error
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

		if err != nil {
			return nil, err
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
			q   []float32
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]float32, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseFloat(rv, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of floats")
			}
			q[i] = float32(v)
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
		return NewEndpointQueryArrayFloat32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat64DecodeCode = `// DecodeEndpointQueryArrayFloat64Request returns a decoder for requests sent
// to the ServiceQueryArrayFloat64 EndpointQueryArrayFloat64 endpoint.
func DecodeEndpointQueryArrayFloat64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayFloat64Payload, error) {
		var (
			q   []float64
			err error
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

		if err != nil {
			return nil, err
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
			q   []float64
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]float64, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseFloat(rv, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of floats")
			}
			q[i] = v
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
		return NewEndpointQueryArrayFloat64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayStringDecodeCode = `// DecodeEndpointQueryArrayStringRequest returns a decoder for requests sent to
// the ServiceQueryArrayString EndpointQueryArrayString endpoint.
func DecodeEndpointQueryArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayStringPayload, error) {
		var (
			q   []string
			err error
		)
		q = r.URL.Query()["q"]

		if err != nil {
			return nil, err
		}
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
		return NewEndpointQueryArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBytesDecodeCode = `// DecodeEndpointQueryArrayBytesRequest returns a decoder for requests sent to
// the ServiceQueryArrayBytes EndpointQueryArrayBytes endpoint.
func DecodeEndpointQueryArrayBytesRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayBytesPayload, error) {
		var (
			q   [][]byte
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([][]byte, len(qRaw))
			for i, rv := range qRaw {
				q[i] = []byte(rv)
			}
		}

		if err != nil {
			return nil, err
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
			q   [][]byte
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([][]byte, len(qRaw))
		for i, rv := range qRaw {
			q[i] = []byte(rv)
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
		return NewEndpointQueryArrayBytesValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayAnyDecodeCode = `// DecodeEndpointQueryArrayAnyRequest returns a decoder for requests sent to
// the ServiceQueryArrayAny EndpointQueryArrayAny endpoint.
func DecodeEndpointQueryArrayAnyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryArrayAnyPayload, error) {
		var (
			q   []interface{}
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw != nil {
			q = make([]interface{}, len(qRaw))
			for i, rv := range qRaw {
				q[i] = rv
			}
		}

		if err != nil {
			return nil, err
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
			q   []interface{}
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]interface{}, len(qRaw))
		for i, rv := range qRaw {
			q[i] = rv
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
		return NewEndpointQueryArrayAnyValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringStringDecodeCode = `// DecodeEndpointQueryMapStringStringRequest returns a decoder for requests
// sent to the ServiceQueryMapStringString EndpointQueryMapStringString
// endpoint.
func DecodeEndpointQueryMapStringStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringStringPayload, error) {
		var (
			q   map[string]string
			err error
		)
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

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapStringStringPayload(q), nil
	}
}
`

var PayloadQueryMapStringStringValidateDecodeCode = `// DecodeEndpointQueryMapStringStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringStringValidate
// EndpointQueryMapStringStringValidate endpoint.
func DecodeEndpointQueryMapStringStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringStringValidatePayload, error) {
		var (
			q   map[string]string
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[string]string, len(qRaw))
		for key, va := range qRaw {
			var val string
			{
				val = va[0]
			}
			q[key] = val
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
		return NewEndpointQueryMapStringStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringBoolDecodeCode = `// DecodeEndpointQueryMapStringBoolRequest returns a decoder for requests sent
// to the ServiceQueryMapStringBool EndpointQueryMapStringBool endpoint.
func DecodeEndpointQueryMapStringBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringBoolPayload, error) {
		var (
			q   map[string]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) != 0 {
			q = make(map[string]bool, len(qRaw))
			for key, va := range qRaw {
				var val bool
				{
					valRaw := va[0]
					v, err := strconv.ParseBool(valRaw)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(valRaw, "query", "boolean")
					}
					val = v
				}
				q[key] = val
			}
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapStringBoolPayload(q), nil
	}
}
`

var PayloadQueryMapStringBoolValidateDecodeCode = `// DecodeEndpointQueryMapStringBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringBoolValidate
// EndpointQueryMapStringBoolValidate endpoint.
func DecodeEndpointQueryMapStringBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringBoolValidatePayload, error) {
		var (
			q   map[string]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[string]bool, len(qRaw))
		for key, va := range qRaw {
			var val bool
			{
				valRaw := va[0]
				v, err := strconv.ParseBool(valRaw)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(valRaw, "query", "boolean")
				}
				val = v
			}
			q[key] = val
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
		return NewEndpointQueryMapStringBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolStringDecodeCode = `// DecodeEndpointQueryMapBoolStringRequest returns a decoder for requests sent
// to the ServiceQueryMapBoolString EndpointQueryMapBoolString endpoint.
func DecodeEndpointQueryMapBoolStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolStringPayload, error) {
		var (
			q   map[bool]string
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) != 0 {
			q = make(map[bool]string, len(qRaw))
			for keyRaw, va := range qRaw {
				var key bool
				{
					v, err := strconv.ParseBool(keyRaw)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
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

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapBoolStringPayload(q), nil
	}
}
`

var PayloadQueryMapBoolStringValidateDecodeCode = `// DecodeEndpointQueryMapBoolStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapBoolStringValidate
// EndpointQueryMapBoolStringValidate endpoint.
func DecodeEndpointQueryMapBoolStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolStringValidatePayload, error) {
		var (
			q   map[bool]string
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[bool]string, len(qRaw))
		for keyRaw, va := range qRaw {
			var key bool
			{
				v, err := strconv.ParseBool(keyRaw)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
				}
				key = v
			}
			var val string
			{
				val = va[0]
			}
			q[key] = val
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
		return NewEndpointQueryMapBoolStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolBoolDecodeCode = `// DecodeEndpointQueryMapBoolBoolRequest returns a decoder for requests sent to
// the ServiceQueryMapBoolBool EndpointQueryMapBoolBool endpoint.
func DecodeEndpointQueryMapBoolBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolBoolPayload, error) {
		var (
			q   map[bool]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) != 0 {
			q = make(map[bool]bool, len(qRaw))
			for keyRaw, va := range qRaw {
				var key bool
				{
					v, err := strconv.ParseBool(keyRaw)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
					}
					key = v
				}
				var val bool
				{
					valRaw := va[0]
					v, err := strconv.ParseBool(valRaw)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(valRaw, "query", "boolean")
					}
					val = v
				}
				q[key] = val
			}
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapBoolBoolPayload(q), nil
	}
}
`

var PayloadQueryMapBoolBoolValidateDecodeCode = `// DecodeEndpointQueryMapBoolBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolBoolValidate EndpointQueryMapBoolBoolValidate
// endpoint.
func DecodeEndpointQueryMapBoolBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolBoolValidatePayload, error) {
		var (
			q   map[bool]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[bool]bool, len(qRaw))
		for keyRaw, va := range qRaw {
			var key bool
			{
				v, err := strconv.ParseBool(keyRaw)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
				}
				key = v
			}
			var val bool
			{
				valRaw := va[0]
				v, err := strconv.ParseBool(valRaw)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(valRaw, "query", "boolean")
				}
				val = v
			}
			q[key] = val
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
		return NewEndpointQueryMapBoolBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayStringDecodeCode = `// DecodeEndpointQueryMapStringArrayStringRequest returns a decoder for
// requests sent to the ServiceQueryMapStringArrayString
// EndpointQueryMapStringArrayString endpoint.
func DecodeEndpointQueryMapStringArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringArrayStringPayload, error) {
		var (
			q   map[string][]string
			err error
		)
		q = r.URL.Query()

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapStringArrayStringPayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayStringValidateDecodeCode = `// DecodeEndpointQueryMapStringArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringArrayStringValidate
// EndpointQueryMapStringArrayStringValidate endpoint.
func DecodeEndpointQueryMapStringArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringArrayStringValidatePayload, error) {
		var (
			q   map[string][]string
			err error
		)
		q = r.URL.Query()
		if len(q) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
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
		return NewEndpointQueryMapStringArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayBoolDecodeCode = `// DecodeEndpointQueryMapStringArrayBoolRequest returns a decoder for requests
// sent to the ServiceQueryMapStringArrayBool EndpointQueryMapStringArrayBool
// endpoint.
func DecodeEndpointQueryMapStringArrayBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringArrayBoolPayload, error) {
		var (
			q   map[string][]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) != 0 {
			q = make(map[string][]bool, len(qRaw))
			for key, valRaw := range qRaw {
				var val []bool
				{
					val = make([]bool, len(valRaw))
					for i, rv := range valRaw {
						v, err := strconv.ParseBool(rv)
						if err != nil {
							return nil, goa.InvalidFieldTypeError(valRaw, "query", "array of booleans")
						}
						val[i] = v
					}
				}
				q[key] = val
			}
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapStringArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayBoolValidateDecodeCode = `// DecodeEndpointQueryMapStringArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringArrayBoolValidate
// EndpointQueryMapStringArrayBoolValidate endpoint.
func DecodeEndpointQueryMapStringArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapStringArrayBoolValidatePayload, error) {
		var (
			q   map[string][]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[string][]bool, len(qRaw))
		for key, valRaw := range qRaw {
			var val []bool
			{
				val = make([]bool, len(valRaw))
				for i, rv := range valRaw {
					v, err := strconv.ParseBool(rv)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(valRaw, "query", "array of booleans")
					}
					val[i] = v
				}
			}
			q[key] = val
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
		return NewEndpointQueryMapStringArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayStringDecodeCode = `// DecodeEndpointQueryMapBoolArrayStringRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolArrayString EndpointQueryMapBoolArrayString
// endpoint.
func DecodeEndpointQueryMapBoolArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolArrayStringPayload, error) {
		var (
			q   map[bool][]string
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) != 0 {
			q = make(map[bool][]string, len(qRaw))
			for keyRaw, val := range qRaw {
				var key bool
				{
					v, err := strconv.ParseBool(keyRaw)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
					}
					key = v
				}
				q[key] = val
			}
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapBoolArrayStringPayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayStringValidateDecodeCode = `// DecodeEndpointQueryMapBoolArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapBoolArrayStringValidate
// EndpointQueryMapBoolArrayStringValidate endpoint.
func DecodeEndpointQueryMapBoolArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolArrayStringValidatePayload, error) {
		var (
			q   map[bool][]string
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) != 0 {
			q = make(map[bool][]string, len(qRaw))
			for keyRaw, val := range qRaw {
				var key bool
				{
					v, err := strconv.ParseBool(keyRaw)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
					}
					key = v
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
		return NewEndpointQueryMapBoolArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayBoolDecodeCode = `// DecodeEndpointQueryMapBoolArrayBoolRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolArrayBool EndpointQueryMapBoolArrayBool
// endpoint.
func DecodeEndpointQueryMapBoolArrayBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolArrayBoolPayload, error) {
		var (
			q   map[bool][]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) != 0 {
			q = make(map[bool][]bool, len(qRaw))
			for keyRaw, valRaw := range qRaw {
				var key bool
				{
					v, err := strconv.ParseBool(keyRaw)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
					}
					key = v
				}
				var val []bool
				{
					val = make([]bool, len(valRaw))
					for i, rv := range valRaw {
						v, err := strconv.ParseBool(rv)
						if err != nil {
							return nil, goa.InvalidFieldTypeError(valRaw, "query", "array of booleans")
						}
						val[i] = v
					}
				}
				q[key] = val
			}
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryMapBoolArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayBoolValidateDecodeCode = `// DecodeEndpointQueryMapBoolArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapBoolArrayBoolValidate
// EndpointQueryMapBoolArrayBoolValidate endpoint.
func DecodeEndpointQueryMapBoolArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryMapBoolArrayBoolValidatePayload, error) {
		var (
			q   map[bool][]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[bool][]bool, len(qRaw))
		for keyRaw, valRaw := range qRaw {
			var key bool
			{
				v, err := strconv.ParseBool(keyRaw)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
				}
				key = v
			}
			var val []bool
			{
				val = make([]bool, len(valRaw))
				for i, rv := range valRaw {
					v, err := strconv.ParseBool(rv)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(valRaw, "query", "array of booleans")
					}
					val[i] = v
				}
			}
			q[key] = val
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
		return NewEndpointQueryMapBoolArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryPrimitiveStringValidateDecodeCode = `// DecodeEndpointQueryPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveStringValidate
// EndpointQueryPrimitiveStringValidate endpoint.
func DecodeEndpointQueryPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
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

var PayloadQueryPrimitiveBoolValidateDecodeCode = `// DecodeEndpointQueryPrimitiveBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveBoolValidate
// EndpointQueryPrimitiveBoolValidate endpoint.
func DecodeEndpointQueryPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (bool, error) {
		var (
			q   bool
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw == "" {
			return nil, goa.MissingFieldError("q", "query string")
		}
		v, err := strconv.ParseBool(qRaw)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(qRaw, q, "boolean")
		}
		q = v
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

var PayloadQueryPrimitiveArrayStringValidateDecodeCode = `// DecodeEndpointQueryPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveArrayStringValidate
// EndpointQueryPrimitiveArrayStringValidate endpoint.
func DecodeEndpointQueryPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]string, error) {
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

var PayloadQueryPrimitiveArrayBoolValidateDecodeCode = `// DecodeEndpointQueryPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveArrayBoolValidate
// EndpointQueryPrimitiveArrayBoolValidate endpoint.
func DecodeEndpointQueryPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]bool, error) {
		var (
			q   []bool
			err error
		)
		qRaw := r.URL.Query()["q"]
		if qRaw == nil {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make([]bool, len(qRaw))
		for i, rv := range qRaw {
			v, err := strconv.ParseBool(rv)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(qRaw, q, "array of booleans")
			}
			q[i] = v
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

var PayloadQueryPrimitiveMapStringArrayStringValidateDecodeCode = `// DecodeEndpointQueryPrimitiveMapStringArrayStringValidateRequest returns a
// decoder for requests sent to the
// ServiceQueryPrimitiveMapStringArrayStringValidate
// EndpointQueryPrimitiveMapStringArrayStringValidate endpoint.
func DecodeEndpointQueryPrimitiveMapStringArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (map[string][]string, error) {
		var (
			q   map[string][]string
			err error
		)
		q = r.URL.Query()
		if len(q) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
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

var PayloadQueryPrimitiveMapStringBoolValidateDecodeCode = `// DecodeEndpointQueryPrimitiveMapStringBoolValidateRequest returns a decoder
// for requests sent to the ServiceQueryPrimitiveMapStringBoolValidate
// EndpointQueryPrimitiveMapStringBoolValidate endpoint.
func DecodeEndpointQueryPrimitiveMapStringBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (map[string]bool, error) {
		var (
			q   map[string]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[string]bool, len(qRaw))
		for key, va := range qRaw {
			var val bool
			{
				valRaw := va[0]
				v, err := strconv.ParseBool(valRaw)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(valRaw, "query", "boolean")
				}
				val = v
			}
			q[key] = val
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

var PayloadQueryPrimitiveMapBoolArrayBoolValidateDecodeCode = `// DecodeEndpointQueryPrimitiveMapBoolArrayBoolValidateRequest returns a
// decoder for requests sent to the
// ServiceQueryPrimitiveMapBoolArrayBoolValidate
// EndpointQueryPrimitiveMapBoolArrayBoolValidate endpoint.
func DecodeEndpointQueryPrimitiveMapBoolArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (map[bool][]bool, error) {
		var (
			q   map[bool][]bool
			err error
		)
		qRaw := r.URL.Query()
		if len(qRaw) == 0 {
			return nil, goa.MissingFieldError("q", "query string")
		}
		q = make(map[bool][]bool, len(qRaw))
		for keyRaw, valRaw := range qRaw {
			var key bool
			{
				v, err := strconv.ParseBool(keyRaw)
				if err != nil {
					return nil, goa.InvalidFieldTypeError(keyRaw, "query", "boolean")
				}
				key = v
			}
			var val []bool
			{
				val = make([]bool, len(valRaw))
				for i, rv := range valRaw {
					v, err := strconv.ParseBool(rv)
					if err != nil {
						return nil, goa.InvalidFieldTypeError(valRaw, "query", "array of booleans")
					}
					val[i] = v
				}
			}
			q[key] = val
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

var PayloadQueryStringDefaultDecodeCode = `// DecodeEndpointQueryStringDefaultRequest returns a decoder for requests sent
// to the ServiceQueryStringDefault EndpointQueryStringDefault endpoint.
func DecodeEndpointQueryStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointQueryStringDefaultPayload, error) {
		var (
			q   string
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = qRaw
		} else {
			q = "def"
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointQueryStringDefaultPayload(q), nil
	}
}
`

var PayloadQueryPrimitiveStringDefaultDecodeCode = `// DecodeEndpointQueryPrimitiveStringDefaultRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveStringDefault
// EndpointQueryPrimitiveStringDefault endpoint.
func DecodeEndpointQueryPrimitiveStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
		var (
			q   string
			err error
		)
		qRaw := r.URL.Query().Get("q")
		if qRaw != "" {
			q = qRaw
		} else {
			q = "def"
		}

		if err != nil {
			return nil, err
		}
		return q, nil
	}
}
`

var PayloadPathStringDecodeCode = `// DecodeEndpointPathStringRequest returns a decoder for requests sent to the
// ServicePathString EndpointPathString endpoint.
func DecodeEndpointPathStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointPathStringPayload, error) {
		var (
			p   string
			err error

			params = rest.ContextParams(r.Context())
		)
		p = params["p"]

		if err != nil {
			return nil, err
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
			p   string
			err error

			params = rest.ContextParams(r.Context())
		)
		p = params["p"]
		if !(p == "val") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", p, []interface{}{"val"}))
		}

		if err != nil {
			return nil, err
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
			p   []string
			err error

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		pRawSlice := strings.Split(pRaw, ",")
		p = make([]string, len(pRawSlice))
		for i, rv := range pRawSlice {
			p[i] = rv
		}

		if err != nil {
			return nil, err
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
			p   []string
			err error

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		pRawSlice := strings.Split(pRaw, ",")
		p = make([]string, len(pRawSlice))
		for i, rv := range pRawSlice {
			p[i] = rv
		}
		if !(p == []string{"val"}) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("p", p, []interface{}{[]string{"val"}}))
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointPathArrayStringValidatePayload(p), nil
	}
}
`

var PayloadPathPrimitiveStringValidateDecodeCode = `// DecodeEndpointPathPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveStringValidate
// EndpointPathPrimitiveStringValidate endpoint.
func DecodeEndpointPathPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
		var (
			p   string
			err error

			params = rest.ContextParams(r.Context())
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

var PayloadPathPrimitiveBoolValidateDecodeCode = `// DecodeEndpointPathPrimitiveBoolValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveBoolValidate
// EndpointPathPrimitiveBoolValidate endpoint.
func DecodeEndpointPathPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (bool, error) {
		var (
			p   bool
			err error

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		v, err := strconv.ParseBool(pRaw)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(pRaw, p, "boolean")
		}
		p = v
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

var PayloadPathPrimitiveArrayStringValidateDecodeCode = `// DecodeEndpointPathPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveArrayStringValidate
// EndpointPathPrimitiveArrayStringValidate endpoint.
func DecodeEndpointPathPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]string, error) {
		var (
			p   []string
			err error

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		pRawSlice := strings.Split(pRaw, ",")
		p = make([]string, len(pRawSlice))
		for i, rv := range pRawSlice {
			p[i] = rv
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

var PayloadPathPrimitiveArrayBoolValidateDecodeCode = `// DecodeEndpointPathPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveArrayBoolValidate
// EndpointPathPrimitiveArrayBoolValidate endpoint.
func DecodeEndpointPathPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]bool, error) {
		var (
			p   []bool
			err error

			params = rest.ContextParams(r.Context())
		)
		pRaw := params["p"]
		pRawSlice := strings.Split(pRaw, ",")
		p = make([]bool, len(pRawSlice))
		for i, rv := range pRawSlice {
			v, err := strconv.ParseBool(rv)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(pRaw, p, "array of booleans")
			}
			p[i] = v
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

var PayloadHeaderStringDecodeCode = `// DecodeEndpointHeaderStringRequest returns a decoder for requests sent to the
// ServiceHeaderString EndpointHeaderString endpoint.
func DecodeEndpointHeaderStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointHeaderStringPayload, error) {
		var (
			h   *string
			err error
		)
		hRaw := r.Header.Get("h")
		if hRaw != "" {
			h = &hRaw
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointHeaderStringPayload(h), nil
	}
}
`

var PayloadHeaderStringValidateDecodeCode = `// DecodeEndpointHeaderStringValidateRequest returns a decoder for requests
// sent to the ServiceHeaderStringValidate EndpointHeaderStringValidate
// endpoint.
func DecodeEndpointHeaderStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointHeaderStringValidatePayload, error) {
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
		return NewEndpointHeaderStringValidatePayload(h), nil
	}
}
`

var PayloadHeaderArrayStringDecodeCode = `// DecodeEndpointHeaderArrayStringRequest returns a decoder for requests sent
// to the ServiceHeaderArrayString EndpointHeaderArrayString endpoint.
func DecodeEndpointHeaderArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointHeaderArrayStringPayload, error) {
		var (
			h   []string
			err error
		)
		h = r.Header["H"]

		if err != nil {
			return nil, err
		}
		return NewEndpointHeaderArrayStringPayload(h), nil
	}
}
`

var PayloadHeaderArrayStringValidateDecodeCode = `// DecodeEndpointHeaderArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceHeaderArrayStringValidate
// EndpointHeaderArrayStringValidate endpoint.
func DecodeEndpointHeaderArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointHeaderArrayStringValidatePayload, error) {
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
		return NewEndpointHeaderArrayStringValidatePayload(h), nil
	}
}
`

var PayloadHeaderPrimitiveStringValidateDecodeCode = `// DecodeEndpointHeaderPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveStringValidate
// EndpointHeaderPrimitiveStringValidate endpoint.
func DecodeEndpointHeaderPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
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

var PayloadHeaderPrimitiveBoolValidateDecodeCode = `// DecodeEndpointHeaderPrimitiveBoolValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveBoolValidate
// EndpointHeaderPrimitiveBoolValidate endpoint.
func DecodeEndpointHeaderPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (bool, error) {
		var (
			h   bool
			err error
		)
		hRaw := r.Header.Get("h")
		if hRaw == "" {
			return nil, goa.MissingFieldError("h", "header")
		}
		v, err := strconv.ParseBool(hRaw)
		if err != nil {
			return nil, goa.InvalidFieldTypeError(hRaw, h, "boolean")
		}
		h = v
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

var PayloadHeaderPrimitiveArrayStringValidateDecodeCode = `// DecodeEndpointHeaderPrimitiveArrayStringValidateRequest returns a decoder
// for requests sent to the ServiceHeaderPrimitiveArrayStringValidate
// EndpointHeaderPrimitiveArrayStringValidate endpoint.
func DecodeEndpointHeaderPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]string, error) {
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

var PayloadHeaderPrimitiveArrayBoolValidateDecodeCode = `// DecodeEndpointHeaderPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveArrayBoolValidate
// EndpointHeaderPrimitiveArrayBoolValidate endpoint.
func DecodeEndpointHeaderPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]bool, error) {
		var (
			h   []bool
			err error
		)
		hRaw := r.Header["H"]
		if hRaw == nil {
			return nil, goa.MissingFieldError("h", "header")
		}
		h = make([]bool, len(hRaw))
		for i, rv := range hRaw {
			v, err := strconv.ParseBool(rv)
			if err != nil {
				return nil, goa.InvalidFieldTypeError(hRaw, h, "array of booleans")
			}
			h[i] = v
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

var PayloadHeaderStringDefaultDecodeCode = `// DecodeEndpointHeaderStringDefaultRequest returns a decoder for requests sent
// to the ServiceHeaderStringDefault EndpointHeaderStringDefault endpoint.
func DecodeEndpointHeaderStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointHeaderStringDefaultPayload, error) {
		var (
			h   string
			err error
		)
		hRaw := r.Header.Get("h")
		if hRaw != "" {
			h = hRaw
		} else {
			h = "def"
		}

		if err != nil {
			return nil, err
		}
		return NewEndpointHeaderStringDefaultPayload(h), nil
	}
}
`

var PayloadHeaderPrimitiveStringDefaultDecodeCode = `// DecodeEndpointHeaderPrimitiveStringDefaultRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveStringDefault
// EndpointHeaderPrimitiveStringDefault endpoint.
func DecodeEndpointHeaderPrimitiveStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
		var (
			h   string
			err error
		)
		hRaw := r.Header.Get("h")
		if hRaw != "" {
			h = hRaw
		} else {
			h = "def"
		}

		if err != nil {
			return nil, err
		}
		return h, nil
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

		if err != nil {
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
		err = goa.MergeErrors(err, body.Validate())

		if err != nil {
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
		err = goa.MergeErrors(err, body.Validate())

		if err != nil {
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

		if err != nil {
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
		err = goa.MergeErrors(err, body.Validate())

		if err != nil {
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

		if err != nil {
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
		err = goa.MergeErrors(err, body.Validate())

		if err != nil {
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

		if err != nil {
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
		err = goa.MergeErrors(err, body.Validate())

		if err != nil {
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

		if err != nil {
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

		if err != nil {
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

		if err != nil {
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

		if err != nil {
			return nil, err
		}
		return &body, nil
	}
}
`

var PayloadBodyPrimitiveStringValidateDecodeCode = `// DecodeEndpointBodyPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveStringValidate
// EndpointBodyPrimitiveStringValidate endpoint.
func DecodeEndpointBodyPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
		var (
			body string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if !(body == "val") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body", body, []interface{}{"val"}))
		}

		if err != nil {
			return nil, err
		}
		return body, nil
	}
}
`

var PayloadBodyPrimitiveBoolValidateDecodeCode = `// DecodeEndpointBodyPrimitiveBoolValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveBoolValidate
// EndpointBodyPrimitiveBoolValidate endpoint.
func DecodeEndpointBodyPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (bool, error) {
		var (
			body bool
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if !(body == true) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body", body, []interface{}{true}))
		}

		if err != nil {
			return nil, err
		}
		return body, nil
	}
}
`

var PayloadBodyPrimitiveArrayStringValidateDecodeCode = `// DecodeEndpointBodyPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayStringValidate
// EndpointBodyPrimitiveArrayStringValidate endpoint.
func DecodeEndpointBodyPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]string, error) {
		var (
			body []string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
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

var PayloadBodyPrimitiveArrayBoolValidateDecodeCode = `// DecodeEndpointBodyPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayBoolValidate
// EndpointBodyPrimitiveArrayBoolValidate endpoint.
func DecodeEndpointBodyPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]bool, error) {
		var (
			body []bool
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
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

		if err != nil {
			return nil, err
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
		err = goa.MergeErrors(err, body.Validate())

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

		if err != nil {
			return nil, err
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
		err = goa.MergeErrors(err, body.Validate())

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

		if err != nil {
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
		err = goa.MergeErrors(err, body.Validate())

		if err != nil {
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

		if err != nil {
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
		err = goa.MergeErrors(err, body.Validate())

		if err != nil {
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
			c string
			b *string

			params = rest.ContextParams(r.Context())
		)
		c = params["c"]
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		if err != nil {
			return nil, err
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
		err = goa.MergeErrors(err, body.Validate())

		var (
			c string
			b string

			params = rest.ContextParams(r.Context())
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
			c string
			b *string

			params = rest.ContextParams(r.Context())
		)
		c = params["c"]
		bRaw := r.URL.Query().Get("b")
		if bRaw != "" {
			b = &bRaw
		}

		if err != nil {
			return nil, err
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
		err = goa.MergeErrors(err, body.Validate())

		var (
			c string
			b string

			params = rest.ContextParams(r.Context())
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
		return NewPayloadType(&body, b, c), nil
	}
}
`
