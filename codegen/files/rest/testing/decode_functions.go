package testing

var PayloadQueryBoolDecodeCode = `// DecodeMethodQueryBoolRequest returns a decoder for requests sent to the
// ServiceQueryBool MethodQueryBool endpoint.
func DecodeMethodQueryBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryBoolPayload, error) {
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
		return NewMethodQueryBoolPayload(q), nil
	}
}
`

var PayloadQueryBoolValidateDecodeCode = `// DecodeMethodQueryBoolValidateRequest returns a decoder for requests sent to
// the ServiceQueryBoolValidate MethodQueryBoolValidate endpoint.
func DecodeMethodQueryBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryBoolValidatePayload, error) {
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
		return NewMethodQueryBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryIntDecodeCode = `// DecodeMethodQueryIntRequest returns a decoder for requests sent to the
// ServiceQueryInt MethodQueryInt endpoint.
func DecodeMethodQueryIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryIntPayload, error) {
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
		return NewMethodQueryIntPayload(q), nil
	}
}
`

var PayloadQueryIntValidateDecodeCode = `// DecodeMethodQueryIntValidateRequest returns a decoder for requests sent to
// the ServiceQueryIntValidate MethodQueryIntValidate endpoint.
func DecodeMethodQueryIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryIntValidatePayload, error) {
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
		return NewMethodQueryIntValidatePayload(q), nil
	}
}
`

var PayloadQueryInt32DecodeCode = `// DecodeMethodQueryInt32Request returns a decoder for requests sent to the
// ServiceQueryInt32 MethodQueryInt32 endpoint.
func DecodeMethodQueryInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryInt32Payload, error) {
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
		return NewMethodQueryInt32Payload(q), nil
	}
}
`

var PayloadQueryInt32ValidateDecodeCode = `// DecodeMethodQueryInt32ValidateRequest returns a decoder for requests sent to
// the ServiceQueryInt32Validate MethodQueryInt32Validate endpoint.
func DecodeMethodQueryInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryInt32ValidatePayload, error) {
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
		return NewMethodQueryInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryInt64DecodeCode = `// DecodeMethodQueryInt64Request returns a decoder for requests sent to the
// ServiceQueryInt64 MethodQueryInt64 endpoint.
func DecodeMethodQueryInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryInt64Payload, error) {
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
		return NewMethodQueryInt64Payload(q), nil
	}
}
`

var PayloadQueryInt64ValidateDecodeCode = `// DecodeMethodQueryInt64ValidateRequest returns a decoder for requests sent to
// the ServiceQueryInt64Validate MethodQueryInt64Validate endpoint.
func DecodeMethodQueryInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryInt64ValidatePayload, error) {
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
		return NewMethodQueryInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryUIntDecodeCode = `// DecodeMethodQueryUIntRequest returns a decoder for requests sent to the
// ServiceQueryUInt MethodQueryUInt endpoint.
func DecodeMethodQueryUIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryUIntPayload, error) {
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
		return NewMethodQueryUIntPayload(q), nil
	}
}
`

var PayloadQueryUIntValidateDecodeCode = `// DecodeMethodQueryUIntValidateRequest returns a decoder for requests sent to
// the ServiceQueryUIntValidate MethodQueryUIntValidate endpoint.
func DecodeMethodQueryUIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryUIntValidatePayload, error) {
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
		return NewMethodQueryUIntValidatePayload(q), nil
	}
}
`

var PayloadQueryUInt32DecodeCode = `// DecodeMethodQueryUInt32Request returns a decoder for requests sent to the
// ServiceQueryUInt32 MethodQueryUInt32 endpoint.
func DecodeMethodQueryUInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryUInt32Payload, error) {
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
		return NewMethodQueryUInt32Payload(q), nil
	}
}
`

var PayloadQueryUInt32ValidateDecodeCode = `// DecodeMethodQueryUInt32ValidateRequest returns a decoder for requests sent
// to the ServiceQueryUInt32Validate MethodQueryUInt32Validate endpoint.
func DecodeMethodQueryUInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryUInt32ValidatePayload, error) {
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
		return NewMethodQueryUInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryUInt64DecodeCode = `// DecodeMethodQueryUInt64Request returns a decoder for requests sent to the
// ServiceQueryUInt64 MethodQueryUInt64 endpoint.
func DecodeMethodQueryUInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryUInt64Payload, error) {
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
		return NewMethodQueryUInt64Payload(q), nil
	}
}
`

var PayloadQueryUInt64ValidateDecodeCode = `// DecodeMethodQueryUInt64ValidateRequest returns a decoder for requests sent
// to the ServiceQueryUInt64Validate MethodQueryUInt64Validate endpoint.
func DecodeMethodQueryUInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryUInt64ValidatePayload, error) {
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
		return NewMethodQueryUInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryFloat32DecodeCode = `// DecodeMethodQueryFloat32Request returns a decoder for requests sent to the
// ServiceQueryFloat32 MethodQueryFloat32 endpoint.
func DecodeMethodQueryFloat32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryFloat32Payload, error) {
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
		return NewMethodQueryFloat32Payload(q), nil
	}
}
`

var PayloadQueryFloat32ValidateDecodeCode = `// DecodeMethodQueryFloat32ValidateRequest returns a decoder for requests sent
// to the ServiceQueryFloat32Validate MethodQueryFloat32Validate endpoint.
func DecodeMethodQueryFloat32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryFloat32ValidatePayload, error) {
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
		return NewMethodQueryFloat32ValidatePayload(q), nil
	}
}
`

var PayloadQueryFloat64DecodeCode = `// DecodeMethodQueryFloat64Request returns a decoder for requests sent to the
// ServiceQueryFloat64 MethodQueryFloat64 endpoint.
func DecodeMethodQueryFloat64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryFloat64Payload, error) {
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
		return NewMethodQueryFloat64Payload(q), nil
	}
}
`

var PayloadQueryFloat64ValidateDecodeCode = `// DecodeMethodQueryFloat64ValidateRequest returns a decoder for requests sent
// to the ServiceQueryFloat64Validate MethodQueryFloat64Validate endpoint.
func DecodeMethodQueryFloat64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryFloat64ValidatePayload, error) {
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
		return NewMethodQueryFloat64ValidatePayload(q), nil
	}
}
`

var PayloadQueryStringDecodeCode = `// DecodeMethodQueryStringRequest returns a decoder for requests sent to the
// ServiceQueryString MethodQueryString endpoint.
func DecodeMethodQueryStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryStringPayload, error) {
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
		return NewMethodQueryStringPayload(q), nil
	}
}
`

var PayloadQueryStringValidateDecodeCode = `// DecodeMethodQueryStringValidateRequest returns a decoder for requests sent
// to the ServiceQueryStringValidate MethodQueryStringValidate endpoint.
func DecodeMethodQueryStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryStringValidatePayload, error) {
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
		return NewMethodQueryStringValidatePayload(q), nil
	}
}
`

var PayloadQueryBytesDecodeCode = `// DecodeMethodQueryBytesRequest returns a decoder for requests sent to the
// ServiceQueryBytes MethodQueryBytes endpoint.
func DecodeMethodQueryBytesRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryBytesPayload, error) {
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
		return NewMethodQueryBytesPayload(q), nil
	}
}
`

var PayloadQueryBytesValidateDecodeCode = `// DecodeMethodQueryBytesValidateRequest returns a decoder for requests sent to
// the ServiceQueryBytesValidate MethodQueryBytesValidate endpoint.
func DecodeMethodQueryBytesValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryBytesValidatePayload, error) {
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
		return NewMethodQueryBytesValidatePayload(q), nil
	}
}
`

var PayloadQueryAnyDecodeCode = `// DecodeMethodQueryAnyRequest returns a decoder for requests sent to the
// ServiceQueryAny MethodQueryAny endpoint.
func DecodeMethodQueryAnyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryAnyPayload, error) {
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
		return NewMethodQueryAnyPayload(q), nil
	}
}
`

var PayloadQueryAnyValidateDecodeCode = `// DecodeMethodQueryAnyValidateRequest returns a decoder for requests sent to
// the ServiceQueryAnyValidate MethodQueryAnyValidate endpoint.
func DecodeMethodQueryAnyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryAnyValidatePayload, error) {
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
		return NewMethodQueryAnyValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBoolDecodeCode = `// DecodeMethodQueryArrayBoolRequest returns a decoder for requests sent to the
// ServiceQueryArrayBool MethodQueryArrayBool endpoint.
func DecodeMethodQueryArrayBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayBoolPayload, error) {
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
		return NewMethodQueryArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryArrayBoolValidateDecodeCode = `// DecodeMethodQueryArrayBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayBoolValidate MethodQueryArrayBoolValidate
// endpoint.
func DecodeMethodQueryArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayBoolValidatePayload, error) {
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
		return NewMethodQueryArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayIntDecodeCode = `// DecodeMethodQueryArrayIntRequest returns a decoder for requests sent to the
// ServiceQueryArrayInt MethodQueryArrayInt endpoint.
func DecodeMethodQueryArrayIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayIntPayload, error) {
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
		return NewMethodQueryArrayIntPayload(q), nil
	}
}
`

var PayloadQueryArrayIntValidateDecodeCode = `// DecodeMethodQueryArrayIntValidateRequest returns a decoder for requests sent
// to the ServiceQueryArrayIntValidate MethodQueryArrayIntValidate endpoint.
func DecodeMethodQueryArrayIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayIntValidatePayload, error) {
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
		return NewMethodQueryArrayIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt32DecodeCode = `// DecodeMethodQueryArrayInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayInt32 MethodQueryArrayInt32 endpoint.
func DecodeMethodQueryArrayInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayInt32Payload, error) {
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
		return NewMethodQueryArrayInt32Payload(q), nil
	}
}
`

var PayloadQueryArrayInt32ValidateDecodeCode = `// DecodeMethodQueryArrayInt32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayInt32Validate MethodQueryArrayInt32Validate
// endpoint.
func DecodeMethodQueryArrayInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayInt32ValidatePayload, error) {
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
		return NewMethodQueryArrayInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayInt64DecodeCode = `// DecodeMethodQueryArrayInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayInt64 MethodQueryArrayInt64 endpoint.
func DecodeMethodQueryArrayInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayInt64Payload, error) {
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
		return NewMethodQueryArrayInt64Payload(q), nil
	}
}
`

var PayloadQueryArrayInt64ValidateDecodeCode = `// DecodeMethodQueryArrayInt64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayInt64Validate MethodQueryArrayInt64Validate
// endpoint.
func DecodeMethodQueryArrayInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayInt64ValidatePayload, error) {
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
		return NewMethodQueryArrayInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUIntDecodeCode = `// DecodeMethodQueryArrayUIntRequest returns a decoder for requests sent to the
// ServiceQueryArrayUInt MethodQueryArrayUInt endpoint.
func DecodeMethodQueryArrayUIntRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayUIntPayload, error) {
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
		return NewMethodQueryArrayUIntPayload(q), nil
	}
}
`

var PayloadQueryArrayUIntValidateDecodeCode = `// DecodeMethodQueryArrayUIntValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUIntValidate MethodQueryArrayUIntValidate
// endpoint.
func DecodeMethodQueryArrayUIntValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayUIntValidatePayload, error) {
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
		return NewMethodQueryArrayUIntValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt32DecodeCode = `// DecodeMethodQueryArrayUInt32Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt32 MethodQueryArrayUInt32 endpoint.
func DecodeMethodQueryArrayUInt32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayUInt32Payload, error) {
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
		return NewMethodQueryArrayUInt32Payload(q), nil
	}
}
`

var PayloadQueryArrayUInt32ValidateDecodeCode = `// DecodeMethodQueryArrayUInt32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUInt32Validate MethodQueryArrayUInt32Validate
// endpoint.
func DecodeMethodQueryArrayUInt32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayUInt32ValidatePayload, error) {
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
		return NewMethodQueryArrayUInt32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayUInt64DecodeCode = `// DecodeMethodQueryArrayUInt64Request returns a decoder for requests sent to
// the ServiceQueryArrayUInt64 MethodQueryArrayUInt64 endpoint.
func DecodeMethodQueryArrayUInt64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayUInt64Payload, error) {
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
		return NewMethodQueryArrayUInt64Payload(q), nil
	}
}
`

var PayloadQueryArrayUInt64ValidateDecodeCode = `// DecodeMethodQueryArrayUInt64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayUInt64Validate MethodQueryArrayUInt64Validate
// endpoint.
func DecodeMethodQueryArrayUInt64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayUInt64ValidatePayload, error) {
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
		return NewMethodQueryArrayUInt64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat32DecodeCode = `// DecodeMethodQueryArrayFloat32Request returns a decoder for requests sent to
// the ServiceQueryArrayFloat32 MethodQueryArrayFloat32 endpoint.
func DecodeMethodQueryArrayFloat32Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayFloat32Payload, error) {
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
		return NewMethodQueryArrayFloat32Payload(q), nil
	}
}
`

var PayloadQueryArrayFloat32ValidateDecodeCode = `// DecodeMethodQueryArrayFloat32ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayFloat32Validate MethodQueryArrayFloat32Validate
// endpoint.
func DecodeMethodQueryArrayFloat32ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayFloat32ValidatePayload, error) {
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
		return NewMethodQueryArrayFloat32ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayFloat64DecodeCode = `// DecodeMethodQueryArrayFloat64Request returns a decoder for requests sent to
// the ServiceQueryArrayFloat64 MethodQueryArrayFloat64 endpoint.
func DecodeMethodQueryArrayFloat64Request(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayFloat64Payload, error) {
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
		return NewMethodQueryArrayFloat64Payload(q), nil
	}
}
`

var PayloadQueryArrayFloat64ValidateDecodeCode = `// DecodeMethodQueryArrayFloat64ValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayFloat64Validate MethodQueryArrayFloat64Validate
// endpoint.
func DecodeMethodQueryArrayFloat64ValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayFloat64ValidatePayload, error) {
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
		return NewMethodQueryArrayFloat64ValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayStringDecodeCode = `// DecodeMethodQueryArrayStringRequest returns a decoder for requests sent to
// the ServiceQueryArrayString MethodQueryArrayString endpoint.
func DecodeMethodQueryArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayStringPayload, error) {
		var (
			q   []string
			err error
		)
		q = r.URL.Query()["q"]

		if err != nil {
			return nil, err
		}
		return NewMethodQueryArrayStringPayload(q), nil
	}
}
`

var PayloadQueryArrayStringValidateDecodeCode = `// DecodeMethodQueryArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayStringValidate MethodQueryArrayStringValidate
// endpoint.
func DecodeMethodQueryArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayStringValidatePayload, error) {
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
		return NewMethodQueryArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayBytesDecodeCode = `// DecodeMethodQueryArrayBytesRequest returns a decoder for requests sent to
// the ServiceQueryArrayBytes MethodQueryArrayBytes endpoint.
func DecodeMethodQueryArrayBytesRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayBytesPayload, error) {
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
		return NewMethodQueryArrayBytesPayload(q), nil
	}
}
`

var PayloadQueryArrayBytesValidateDecodeCode = `// DecodeMethodQueryArrayBytesValidateRequest returns a decoder for requests
// sent to the ServiceQueryArrayBytesValidate MethodQueryArrayBytesValidate
// endpoint.
func DecodeMethodQueryArrayBytesValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayBytesValidatePayload, error) {
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
		return NewMethodQueryArrayBytesValidatePayload(q), nil
	}
}
`

var PayloadQueryArrayAnyDecodeCode = `// DecodeMethodQueryArrayAnyRequest returns a decoder for requests sent to the
// ServiceQueryArrayAny MethodQueryArrayAny endpoint.
func DecodeMethodQueryArrayAnyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayAnyPayload, error) {
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
		return NewMethodQueryArrayAnyPayload(q), nil
	}
}
`

var PayloadQueryArrayAnyValidateDecodeCode = `// DecodeMethodQueryArrayAnyValidateRequest returns a decoder for requests sent
// to the ServiceQueryArrayAnyValidate MethodQueryArrayAnyValidate endpoint.
func DecodeMethodQueryArrayAnyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryArrayAnyValidatePayload, error) {
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
		return NewMethodQueryArrayAnyValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringStringDecodeCode = `// DecodeMethodQueryMapStringStringRequest returns a decoder for requests sent
// to the ServiceQueryMapStringString MethodQueryMapStringString endpoint.
func DecodeMethodQueryMapStringStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringStringPayload, error) {
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
		return NewMethodQueryMapStringStringPayload(q), nil
	}
}
`

var PayloadQueryMapStringStringValidateDecodeCode = `// DecodeMethodQueryMapStringStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringStringValidate
// MethodQueryMapStringStringValidate endpoint.
func DecodeMethodQueryMapStringStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringStringValidatePayload, error) {
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
		return NewMethodQueryMapStringStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringBoolDecodeCode = `// DecodeMethodQueryMapStringBoolRequest returns a decoder for requests sent to
// the ServiceQueryMapStringBool MethodQueryMapStringBool endpoint.
func DecodeMethodQueryMapStringBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringBoolPayload, error) {
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
		return NewMethodQueryMapStringBoolPayload(q), nil
	}
}
`

var PayloadQueryMapStringBoolValidateDecodeCode = `// DecodeMethodQueryMapStringBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryMapStringBoolValidate
// MethodQueryMapStringBoolValidate endpoint.
func DecodeMethodQueryMapStringBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringBoolValidatePayload, error) {
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
		return NewMethodQueryMapStringBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolStringDecodeCode = `// DecodeMethodQueryMapBoolStringRequest returns a decoder for requests sent to
// the ServiceQueryMapBoolString MethodQueryMapBoolString endpoint.
func DecodeMethodQueryMapBoolStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolStringPayload, error) {
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
		return NewMethodQueryMapBoolStringPayload(q), nil
	}
}
`

var PayloadQueryMapBoolStringValidateDecodeCode = `// DecodeMethodQueryMapBoolStringValidateRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolStringValidate
// MethodQueryMapBoolStringValidate endpoint.
func DecodeMethodQueryMapBoolStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolStringValidatePayload, error) {
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
		return NewMethodQueryMapBoolStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolBoolDecodeCode = `// DecodeMethodQueryMapBoolBoolRequest returns a decoder for requests sent to
// the ServiceQueryMapBoolBool MethodQueryMapBoolBool endpoint.
func DecodeMethodQueryMapBoolBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolBoolPayload, error) {
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
		return NewMethodQueryMapBoolBoolPayload(q), nil
	}
}
`

var PayloadQueryMapBoolBoolValidateDecodeCode = `// DecodeMethodQueryMapBoolBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolBoolValidate MethodQueryMapBoolBoolValidate
// endpoint.
func DecodeMethodQueryMapBoolBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolBoolValidatePayload, error) {
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
		return NewMethodQueryMapBoolBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayStringDecodeCode = `// DecodeMethodQueryMapStringArrayStringRequest returns a decoder for requests
// sent to the ServiceQueryMapStringArrayString MethodQueryMapStringArrayString
// endpoint.
func DecodeMethodQueryMapStringArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringArrayStringPayload, error) {
		var (
			q   map[string][]string
			err error
		)
		q = r.URL.Query()

		if err != nil {
			return nil, err
		}
		return NewMethodQueryMapStringArrayStringPayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayStringValidateDecodeCode = `// DecodeMethodQueryMapStringArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringArrayStringValidate
// MethodQueryMapStringArrayStringValidate endpoint.
func DecodeMethodQueryMapStringArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringArrayStringValidatePayload, error) {
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
		return NewMethodQueryMapStringArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayBoolDecodeCode = `// DecodeMethodQueryMapStringArrayBoolRequest returns a decoder for requests
// sent to the ServiceQueryMapStringArrayBool MethodQueryMapStringArrayBool
// endpoint.
func DecodeMethodQueryMapStringArrayBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringArrayBoolPayload, error) {
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
		return NewMethodQueryMapStringArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryMapStringArrayBoolValidateDecodeCode = `// DecodeMethodQueryMapStringArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapStringArrayBoolValidate
// MethodQueryMapStringArrayBoolValidate endpoint.
func DecodeMethodQueryMapStringArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapStringArrayBoolValidatePayload, error) {
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
		return NewMethodQueryMapStringArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayStringDecodeCode = `// DecodeMethodQueryMapBoolArrayStringRequest returns a decoder for requests
// sent to the ServiceQueryMapBoolArrayString MethodQueryMapBoolArrayString
// endpoint.
func DecodeMethodQueryMapBoolArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolArrayStringPayload, error) {
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
		return NewMethodQueryMapBoolArrayStringPayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayStringValidateDecodeCode = `// DecodeMethodQueryMapBoolArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapBoolArrayStringValidate
// MethodQueryMapBoolArrayStringValidate endpoint.
func DecodeMethodQueryMapBoolArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolArrayStringValidatePayload, error) {
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
		return NewMethodQueryMapBoolArrayStringValidatePayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayBoolDecodeCode = `// DecodeMethodQueryMapBoolArrayBoolRequest returns a decoder for requests sent
// to the ServiceQueryMapBoolArrayBool MethodQueryMapBoolArrayBool endpoint.
func DecodeMethodQueryMapBoolArrayBoolRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolArrayBoolPayload, error) {
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
		return NewMethodQueryMapBoolArrayBoolPayload(q), nil
	}
}
`

var PayloadQueryMapBoolArrayBoolValidateDecodeCode = `// DecodeMethodQueryMapBoolArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryMapBoolArrayBoolValidate
// MethodQueryMapBoolArrayBoolValidate endpoint.
func DecodeMethodQueryMapBoolArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryMapBoolArrayBoolValidatePayload, error) {
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
		return NewMethodQueryMapBoolArrayBoolValidatePayload(q), nil
	}
}
`

var PayloadQueryPrimitiveStringValidateDecodeCode = `// DecodeMethodQueryPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveStringValidate
// MethodQueryPrimitiveStringValidate endpoint.
func DecodeMethodQueryPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadQueryPrimitiveBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveBoolValidateRequest returns a decoder for requests
// sent to the ServiceQueryPrimitiveBoolValidate
// MethodQueryPrimitiveBoolValidate endpoint.
func DecodeMethodQueryPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadQueryPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodQueryPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveArrayStringValidate
// MethodQueryPrimitiveArrayStringValidate endpoint.
func DecodeMethodQueryPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadQueryPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveArrayBoolValidate
// MethodQueryPrimitiveArrayBoolValidate endpoint.
func DecodeMethodQueryPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadQueryPrimitiveMapStringArrayStringValidateDecodeCode = `// DecodeMethodQueryPrimitiveMapStringArrayStringValidateRequest returns a
// decoder for requests sent to the
// ServiceQueryPrimitiveMapStringArrayStringValidate
// MethodQueryPrimitiveMapStringArrayStringValidate endpoint.
func DecodeMethodQueryPrimitiveMapStringArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadQueryPrimitiveMapStringBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveMapStringBoolValidateRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveMapStringBoolValidate
// MethodQueryPrimitiveMapStringBoolValidate endpoint.
func DecodeMethodQueryPrimitiveMapStringBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadQueryPrimitiveMapBoolArrayBoolValidateDecodeCode = `// DecodeMethodQueryPrimitiveMapBoolArrayBoolValidateRequest returns a decoder
// for requests sent to the ServiceQueryPrimitiveMapBoolArrayBoolValidate
// MethodQueryPrimitiveMapBoolArrayBoolValidate endpoint.
func DecodeMethodQueryPrimitiveMapBoolArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadQueryStringDefaultDecodeCode = `// DecodeMethodQueryStringDefaultRequest returns a decoder for requests sent to
// the ServiceQueryStringDefault MethodQueryStringDefault endpoint.
func DecodeMethodQueryStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodQueryStringDefaultPayload, error) {
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
		return NewMethodQueryStringDefaultPayload(q), nil
	}
}
`

var PayloadQueryPrimitiveStringDefaultDecodeCode = `// DecodeMethodQueryPrimitiveStringDefaultRequest returns a decoder for
// requests sent to the ServiceQueryPrimitiveStringDefault
// MethodQueryPrimitiveStringDefault endpoint.
func DecodeMethodQueryPrimitiveStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadPathStringDecodeCode = `// DecodeMethodPathStringRequest returns a decoder for requests sent to the
// ServicePathString MethodPathString endpoint.
func DecodeMethodPathStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodPathStringPayload, error) {
		var (
			p   string
			err error

			params = rest.ContextParams(r.Context())
		)
		p = params["p"]

		if err != nil {
			return nil, err
		}
		return NewMethodPathStringPayload(p), nil
	}
}
`

var PayloadPathStringValidateDecodeCode = `// DecodeMethodPathStringValidateRequest returns a decoder for requests sent to
// the ServicePathStringValidate MethodPathStringValidate endpoint.
func DecodeMethodPathStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodPathStringValidatePayload, error) {
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
		return NewMethodPathStringValidatePayload(p), nil
	}
}
`

var PayloadPathArrayStringDecodeCode = `// DecodeMethodPathArrayStringRequest returns a decoder for requests sent to
// the ServicePathArrayString MethodPathArrayString endpoint.
func DecodeMethodPathArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodPathArrayStringPayload, error) {
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
		return NewMethodPathArrayStringPayload(p), nil
	}
}
`

var PayloadPathArrayStringValidateDecodeCode = `// DecodeMethodPathArrayStringValidateRequest returns a decoder for requests
// sent to the ServicePathArrayStringValidate MethodPathArrayStringValidate
// endpoint.
func DecodeMethodPathArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodPathArrayStringValidatePayload, error) {
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
		return NewMethodPathArrayStringValidatePayload(p), nil
	}
}
`

var PayloadPathPrimitiveStringValidateDecodeCode = `// DecodeMethodPathPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveStringValidate
// MethodPathPrimitiveStringValidate endpoint.
func DecodeMethodPathPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadPathPrimitiveBoolValidateDecodeCode = `// DecodeMethodPathPrimitiveBoolValidateRequest returns a decoder for requests
// sent to the ServicePathPrimitiveBoolValidate MethodPathPrimitiveBoolValidate
// endpoint.
func DecodeMethodPathPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadPathPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodPathPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveArrayStringValidate
// MethodPathPrimitiveArrayStringValidate endpoint.
func DecodeMethodPathPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadPathPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodPathPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServicePathPrimitiveArrayBoolValidate
// MethodPathPrimitiveArrayBoolValidate endpoint.
func DecodeMethodPathPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadHeaderStringDecodeCode = `// DecodeMethodHeaderStringRequest returns a decoder for requests sent to the
// ServiceHeaderString MethodHeaderString endpoint.
func DecodeMethodHeaderStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodHeaderStringPayload, error) {
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
		return NewMethodHeaderStringPayload(h), nil
	}
}
`

var PayloadHeaderStringValidateDecodeCode = `// DecodeMethodHeaderStringValidateRequest returns a decoder for requests sent
// to the ServiceHeaderStringValidate MethodHeaderStringValidate endpoint.
func DecodeMethodHeaderStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodHeaderStringValidatePayload, error) {
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
		return NewMethodHeaderStringValidatePayload(h), nil
	}
}
`

var PayloadHeaderArrayStringDecodeCode = `// DecodeMethodHeaderArrayStringRequest returns a decoder for requests sent to
// the ServiceHeaderArrayString MethodHeaderArrayString endpoint.
func DecodeMethodHeaderArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodHeaderArrayStringPayload, error) {
		var (
			h   []string
			err error
		)
		h = r.Header["H"]

		if err != nil {
			return nil, err
		}
		return NewMethodHeaderArrayStringPayload(h), nil
	}
}
`

var PayloadHeaderArrayStringValidateDecodeCode = `// DecodeMethodHeaderArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceHeaderArrayStringValidate MethodHeaderArrayStringValidate
// endpoint.
func DecodeMethodHeaderArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodHeaderArrayStringValidatePayload, error) {
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
		return NewMethodHeaderArrayStringValidatePayload(h), nil
	}
}
`

var PayloadHeaderPrimitiveStringValidateDecodeCode = `// DecodeMethodHeaderPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveStringValidate
// MethodHeaderPrimitiveStringValidate endpoint.
func DecodeMethodHeaderPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadHeaderPrimitiveBoolValidateDecodeCode = `// DecodeMethodHeaderPrimitiveBoolValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveBoolValidate
// MethodHeaderPrimitiveBoolValidate endpoint.
func DecodeMethodHeaderPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadHeaderPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodHeaderPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveArrayStringValidate
// MethodHeaderPrimitiveArrayStringValidate endpoint.
func DecodeMethodHeaderPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadHeaderPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodHeaderPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveArrayBoolValidate
// MethodHeaderPrimitiveArrayBoolValidate endpoint.
func DecodeMethodHeaderPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadHeaderStringDefaultDecodeCode = `// DecodeMethodHeaderStringDefaultRequest returns a decoder for requests sent
// to the ServiceHeaderStringDefault MethodHeaderStringDefault endpoint.
func DecodeMethodHeaderStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodHeaderStringDefaultPayload, error) {
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
		return NewMethodHeaderStringDefaultPayload(h), nil
	}
}
`

var PayloadHeaderPrimitiveStringDefaultDecodeCode = `// DecodeMethodHeaderPrimitiveStringDefaultRequest returns a decoder for
// requests sent to the ServiceHeaderPrimitiveStringDefault
// MethodHeaderPrimitiveStringDefault endpoint.
func DecodeMethodHeaderPrimitiveStringDefaultRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadBodyStringDecodeCode = `// DecodeMethodBodyStringRequest returns a decoder for requests sent to the
// ServiceBodyString MethodBodyString endpoint.
func DecodeMethodBodyStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyStringPayload, error) {
		var (
			body MethodBodyStringPayload
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

var PayloadBodyStringValidateDecodeCode = `// DecodeMethodBodyStringValidateRequest returns a decoder for requests sent to
// the ServiceBodyStringValidate MethodBodyStringValidate endpoint.
func DecodeMethodBodyStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyStringValidatePayload, error) {
		var (
			body MethodBodyStringValidatePayload
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

var PayloadBodyObjectDecodeCode = `// DecodeMethodBodyObjectRequest returns a decoder for requests sent to the
// ServiceObjectBody MethodBodyObject endpoint.
func DecodeMethodBodyObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyObjectPayload, error) {
		var (
			body MethodBodyObjectPayload
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

var PayloadObjectBodyValidateDecodeCode = `// DecodeMethodBodyObjectValidateRequest returns a decoder for requests sent
// to the ServiceObjectBodyValidate MethodBodyObjectValidate endpoint.
func DecodeMethodBodyObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyObjectValidatePayload, error) {
		var (
			body MethodBodyObjectValidatePayload
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

var PayloadBodyUserDecodeCode = `// DecodeMethodBodyUserRequest returns a decoder for requests sent to the
// ServiceBodyUser MethodBodyUser endpoint.
func DecodeMethodBodyUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadUserBodyValidateDecodeCode = `// DecodeMethodBodyUserValidateRequest returns a decoder for requests sent to
// the ServiceBodyUserValidate MethodBodyUserValidate endpoint.
func DecodeMethodBodyUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadBodyArrayStringDecodeCode = `// DecodeMethodBodyArrayStringRequest returns a decoder for requests sent to
// the ServiceBodyArrayString MethodBodyArrayString endpoint.
func DecodeMethodBodyArrayStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyArrayStringPayload, error) {
		var (
			body MethodBodyArrayStringPayload
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

var PayloadBodyArrayStringValidateDecodeCode = `// DecodeMethodBodyArrayStringValidateRequest returns a decoder for requests
// sent to the ServiceBodyArrayStringValidate MethodBodyArrayStringValidate
// endpoint.
func DecodeMethodBodyArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyArrayStringValidatePayload, error) {
		var (
			body MethodBodyArrayStringValidatePayload
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

var PayloadBodyArrayUserDecodeCode = `// DecodeMethodBodyArrayUserRequest returns a decoder for requests sent to the
// ServiceBodyArrayUser MethodBodyArrayUser endpoint.
func DecodeMethodBodyArrayUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyArrayUserPayload, error) {
		var (
			body MethodBodyArrayUserPayload
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

var PayloadBodyArrayUserValidateDecodeCode = `// DecodeMethodBodyArrayUserValidateRequest returns a decoder for requests sent
// to the ServiceBodyArrayUserValidate MethodBodyArrayUserValidate endpoint.
func DecodeMethodBodyArrayUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyArrayUserValidatePayload, error) {
		var (
			body MethodBodyArrayUserValidatePayload
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

var PayloadBodyMapStringDecodeCode = `// DecodeMethodBodyMapStringRequest returns a decoder for requests sent to the
// ServiceBodyMapString MethodBodyMapString endpoint.
func DecodeMethodBodyMapStringRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyMapStringPayload, error) {
		var (
			body MethodBodyMapStringPayload
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

var PayloadBodyMapStringValidateDecodeCode = `// DecodeMethodBodyMapStringValidateRequest returns a decoder for requests sent
// to the ServiceBodyMapStringValidate MethodBodyMapStringValidate endpoint.
func DecodeMethodBodyMapStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyMapStringValidatePayload, error) {
		var (
			body MethodBodyMapStringValidatePayload
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

var PayloadBodyMapUserDecodeCode = `// DecodeMethodBodyMapUserRequest returns a decoder for requests sent to the
// ServiceBodyMapUser MethodBodyMapUser endpoint.
func DecodeMethodBodyMapUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyMapUserPayload, error) {
		var (
			body MethodBodyMapUserPayload
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

var PayloadBodyMapUserValidateDecodeCode = `// DecodeMethodBodyMapUserValidateRequest returns a decoder for requests sent
// to the ServiceBodyMapUserValidate MethodBodyMapUserValidate endpoint.
func DecodeMethodBodyMapUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyMapUserValidatePayload, error) {
		var (
			body MethodBodyMapUserValidatePayload
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

var PayloadBodyPrimitiveStringValidateDecodeCode = `// DecodeMethodBodyPrimitiveStringValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveStringValidate
// MethodBodyPrimitiveStringValidate endpoint.
func DecodeMethodBodyPrimitiveStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadBodyPrimitiveBoolValidateDecodeCode = `// DecodeMethodBodyPrimitiveBoolValidateRequest returns a decoder for requests
// sent to the ServiceBodyPrimitiveBoolValidate MethodBodyPrimitiveBoolValidate
// endpoint.
func DecodeMethodBodyPrimitiveBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadBodyPrimitiveArrayStringValidateDecodeCode = `// DecodeMethodBodyPrimitiveArrayStringValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayStringValidate
// MethodBodyPrimitiveArrayStringValidate endpoint.
func DecodeMethodBodyPrimitiveArrayStringValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadBodyPrimitiveArrayBoolValidateDecodeCode = `// DecodeMethodBodyPrimitiveArrayBoolValidateRequest returns a decoder for
// requests sent to the ServiceBodyPrimitiveArrayBoolValidate
// MethodBodyPrimitiveArrayBoolValidate endpoint.
func DecodeMethodBodyPrimitiveArrayBoolValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

var PayloadBodyQueryObjectDecodeCode = `// DecodeMethodBodyQueryObjectRequest returns a decoder for requests sent to
// the ServiceBodyQueryObject MethodBodyQueryObject endpoint.
func DecodeMethodBodyQueryObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyQueryObjectPayload, error) {
		var (
			body MethodBodyQueryObjectServerRequestBody
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
		return NewMethodBodyQueryObjectPayload(&body, b), nil
	}
}
`

var PayloadBodyQueryObjectValidateDecodeCode = `// DecodeMethodBodyQueryObjectValidateRequest returns a decoder for requests
// sent to the ServiceBodyQueryObjectValidate MethodBodyQueryObjectValidate
// endpoint.
func DecodeMethodBodyQueryObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyQueryObjectValidatePayload, error) {
		var (
			body MethodBodyQueryObjectValidateServerRequestBody
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
		return NewMethodBodyQueryObjectValidatePayload(&body, b), nil
	}
}
`

var PayloadBodyQueryUserDecodeCode = `// DecodeMethodBodyQueryUserRequest returns a decoder for requests sent to the
// ServiceBodyQueryUser MethodBodyQueryUser endpoint.
func DecodeMethodBodyQueryUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body MethodBodyQueryUserServerRequestBody
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

var PayloadBodyQueryUserValidateDecodeCode = `// DecodeMethodBodyQueryUserValidateRequest returns a decoder for requests sent
// to the ServiceBodyQueryUserValidate MethodBodyQueryUserValidate endpoint.
func DecodeMethodBodyQueryUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body MethodBodyQueryUserValidateServerRequestBody
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

var PayloadBodyPathObjectDecodeCode = `// DecodeMethodBodyPathObjectRequest returns a decoder for requests sent to the
// ServiceBodyPathObject MethodBodyPathObject endpoint.
func DecodeMethodBodyPathObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyPathObjectPayload, error) {
		var (
			body MethodBodyPathObjectServerRequestBody
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
			b string

			params = rest.ContextParams(r.Context())
		)
		b = params["b"]

		if err != nil {
			return nil, err
		}
		return NewMethodBodyPathObjectPayload(&body, b), nil
	}
}
`

var PayloadBodyPathObjectValidateDecodeCode = `// DecodeMethodBodyPathObjectValidateRequest returns a decoder for requests
// sent to the ServiceBodyPathObjectValidate MethodBodyPathObjectValidate
// endpoint.
func DecodeMethodBodyPathObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyPathObjectValidatePayload, error) {
		var (
			body MethodBodyPathObjectValidateServerRequestBody
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

			params = rest.ContextParams(r.Context())
		)
		b = params["b"]
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))

		if err != nil {
			return nil, err
		}
		return NewMethodBodyPathObjectValidatePayload(&body, b), nil
	}
}
`

var PayloadBodyPathUserDecodeCode = `// DecodeMethodBodyPathUserRequest returns a decoder for requests sent to the
// ServiceBodyPathUser MethodBodyPathUser endpoint.
func DecodeMethodBodyPathUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body MethodBodyPathUserServerRequestBody
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
			b string

			params = rest.ContextParams(r.Context())
		)
		b = params["b"]

		if err != nil {
			return nil, err
		}
		return NewPayloadType(&body, b), nil
	}
}
`

var PayloadBodyPathUserValidateDecodeCode = `// DecodeMethodUserBodyPathValidateRequest returns a decoder for requests sent
// to the ServiceBodyPathUserValidate MethodUserBodyPathValidate endpoint.
func DecodeMethodUserBodyPathValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body MethodUserBodyPathValidateServerRequestBody
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

			params = rest.ContextParams(r.Context())
		)
		b = params["b"]
		err = goa.MergeErrors(err, goa.ValidatePattern("b", b, "patternb"))

		if err != nil {
			return nil, err
		}
		return NewPayloadType(&body, b), nil
	}
}
`

var PayloadBodyQueryPathObjectDecodeCode = `// DecodeMethodBodyQueryPathObjectRequest returns a decoder for requests sent
// to the ServiceBodyQueryPathObject MethodBodyQueryPathObject endpoint.
func DecodeMethodBodyQueryPathObjectRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyQueryPathObjectPayload, error) {
		var (
			body MethodBodyQueryPathObjectServerRequestBody
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
		return NewMethodBodyQueryPathObjectPayload(&body, b, c), nil
	}
}
`

var PayloadBodyQueryPathObjectValidateDecodeCode = `// DecodeMethodBodyQueryPathObjectValidateRequest returns a decoder for
// requests sent to the ServiceBodyQueryPathObjectValidate
// MethodBodyQueryPathObjectValidate endpoint.
func DecodeMethodBodyQueryPathObjectValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*MethodBodyQueryPathObjectValidatePayload, error) {
		var (
			body MethodBodyQueryPathObjectValidateServerRequestBody
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
		return NewMethodBodyQueryPathObjectValidatePayload(&body, b, c), nil
	}
}
`

var PayloadBodyQueryPathUserDecodeCode = `// DecodeMethodBodyQueryPathUserRequest returns a decoder for requests sent to
// the ServiceBodyQueryPathUser MethodBodyQueryPathUser endpoint.
func DecodeMethodBodyQueryPathUserRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body MethodBodyQueryPathUserServerRequestBody
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

var PayloadBodyQueryPathUserValidateDecodeCode = `// DecodeMethodBodyQueryPathUserValidateRequest returns a decoder for requests
// sent to the ServiceBodyQueryPathUserValidate MethodBodyQueryPathUserValidate
// endpoint.
func DecodeMethodBodyQueryPathUserValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body MethodBodyQueryPathUserValidateServerRequestBody
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
