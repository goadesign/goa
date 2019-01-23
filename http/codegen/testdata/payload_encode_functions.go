package testdata

var PayloadQueryBoolValidateEncodeCode = `// EncodeMethodQueryBoolValidateRequest returns an encoder for requests sent to
// the ServiceQueryBoolValidate MethodQueryBoolValidate server.
func EncodeMethodQueryBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryboolvalidate.MethodQueryBoolValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryBoolValidate", "MethodQueryBoolValidate", "*servicequeryboolvalidate.MethodQueryBoolValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryIntValidateEncodeCode = `// EncodeMethodQueryIntValidateRequest returns an encoder for requests sent to
// the ServiceQueryIntValidate MethodQueryIntValidate server.
func EncodeMethodQueryIntValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryintvalidate.MethodQueryIntValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryIntValidate", "MethodQueryIntValidate", "*servicequeryintvalidate.MethodQueryIntValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryInt32ValidateEncodeCode = `// EncodeMethodQueryInt32ValidateRequest returns an encoder for requests sent
// to the ServiceQueryInt32Validate MethodQueryInt32Validate server.
func EncodeMethodQueryInt32ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryint32validate.MethodQueryInt32ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryInt32Validate", "MethodQueryInt32Validate", "*servicequeryint32validate.MethodQueryInt32ValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryInt64ValidateEncodeCode = `// EncodeMethodQueryInt64ValidateRequest returns an encoder for requests sent
// to the ServiceQueryInt64Validate MethodQueryInt64Validate server.
func EncodeMethodQueryInt64ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryint64validate.MethodQueryInt64ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryInt64Validate", "MethodQueryInt64Validate", "*servicequeryint64validate.MethodQueryInt64ValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryUIntValidateEncodeCode = `// EncodeMethodQueryUIntValidateRequest returns an encoder for requests sent to
// the ServiceQueryUIntValidate MethodQueryUIntValidate server.
func EncodeMethodQueryUIntValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryuintvalidate.MethodQueryUIntValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryUIntValidate", "MethodQueryUIntValidate", "*servicequeryuintvalidate.MethodQueryUIntValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryUInt32ValidateEncodeCode = `// EncodeMethodQueryUInt32ValidateRequest returns an encoder for requests sent
// to the ServiceQueryUInt32Validate MethodQueryUInt32Validate server.
func EncodeMethodQueryUInt32ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryuint32validate.MethodQueryUInt32ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryUInt32Validate", "MethodQueryUInt32Validate", "*servicequeryuint32validate.MethodQueryUInt32ValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryUInt64ValidateEncodeCode = `// EncodeMethodQueryUInt64ValidateRequest returns an encoder for requests sent
// to the ServiceQueryUInt64Validate MethodQueryUInt64Validate server.
func EncodeMethodQueryUInt64ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryuint64validate.MethodQueryUInt64ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryUInt64Validate", "MethodQueryUInt64Validate", "*servicequeryuint64validate.MethodQueryUInt64ValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryFloat32ValidateEncodeCode = `// EncodeMethodQueryFloat32ValidateRequest returns an encoder for requests sent
// to the ServiceQueryFloat32Validate MethodQueryFloat32Validate server.
func EncodeMethodQueryFloat32ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryfloat32validate.MethodQueryFloat32ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryFloat32Validate", "MethodQueryFloat32Validate", "*servicequeryfloat32validate.MethodQueryFloat32ValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryFloat64ValidateEncodeCode = `// EncodeMethodQueryFloat64ValidateRequest returns an encoder for requests sent
// to the ServiceQueryFloat64Validate MethodQueryFloat64Validate server.
func EncodeMethodQueryFloat64ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryfloat64validate.MethodQueryFloat64ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryFloat64Validate", "MethodQueryFloat64Validate", "*servicequeryfloat64validate.MethodQueryFloat64ValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryStringValidateEncodeCode = `// EncodeMethodQueryStringValidateRequest returns an encoder for requests sent
// to the ServiceQueryStringValidate MethodQueryStringValidate server.
func EncodeMethodQueryStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerystringvalidate.MethodQueryStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryStringValidate", "MethodQueryStringValidate", "*servicequerystringvalidate.MethodQueryStringValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", p.Q)
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryBytesValidateEncodeCode = `// EncodeMethodQueryBytesValidateRequest returns an encoder for requests sent
// to the ServiceQueryBytesValidate MethodQueryBytesValidate server.
func EncodeMethodQueryBytesValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerybytesvalidate.MethodQueryBytesValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryBytesValidate", "MethodQueryBytesValidate", "*servicequerybytesvalidate.MethodQueryBytesValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", string(p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryAnyValidateEncodeCode = `// EncodeMethodQueryAnyValidateRequest returns an encoder for requests sent to
// the ServiceQueryAnyValidate MethodQueryAnyValidate server.
func EncodeMethodQueryAnyValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryanyvalidate.MethodQueryAnyValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryAnyValidate", "MethodQueryAnyValidate", "*servicequeryanyvalidate.MethodQueryAnyValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("q", fmt.Sprintf("%v", p.Q))
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayBoolValidateEncodeCode = `// EncodeMethodQueryArrayBoolValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayBoolValidate MethodQueryArrayBoolValidate
// server.
func EncodeMethodQueryArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayBoolValidate", "MethodQueryArrayBoolValidate", "*servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatBool(value)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayIntValidateEncodeCode = `// EncodeMethodQueryArrayIntValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayIntValidate MethodQueryArrayIntValidate server.
func EncodeMethodQueryArrayIntValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayIntValidate", "MethodQueryArrayIntValidate", "*servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.Itoa(value)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayInt32ValidateEncodeCode = `// EncodeMethodQueryArrayInt32ValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayInt32Validate MethodQueryArrayInt32Validate
// server.
func EncodeMethodQueryArrayInt32ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayInt32Validate", "MethodQueryArrayInt32Validate", "*servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatInt(int64(value), 10)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayInt64ValidateEncodeCode = `// EncodeMethodQueryArrayInt64ValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayInt64Validate MethodQueryArrayInt64Validate
// server.
func EncodeMethodQueryArrayInt64ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayInt64Validate", "MethodQueryArrayInt64Validate", "*servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatInt(value, 10)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayUIntValidateEncodeCode = `// EncodeMethodQueryArrayUIntValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayUIntValidate MethodQueryArrayUIntValidate
// server.
func EncodeMethodQueryArrayUIntValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayUIntValidate", "MethodQueryArrayUIntValidate", "*servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatUint(uint64(value), 10)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayUInt32ValidateEncodeCode = `// EncodeMethodQueryArrayUInt32ValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayUInt32Validate MethodQueryArrayUInt32Validate
// server.
func EncodeMethodQueryArrayUInt32ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayUInt32Validate", "MethodQueryArrayUInt32Validate", "*servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatUint(uint64(value), 10)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayUInt64ValidateEncodeCode = `// EncodeMethodQueryArrayUInt64ValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayUInt64Validate MethodQueryArrayUInt64Validate
// server.
func EncodeMethodQueryArrayUInt64ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayUInt64Validate", "MethodQueryArrayUInt64Validate", "*servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatUint(value, 10)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayFloat32ValidateEncodeCode = `// EncodeMethodQueryArrayFloat32ValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayFloat32Validate MethodQueryArrayFloat32Validate
// server.
func EncodeMethodQueryArrayFloat32ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayFloat32Validate", "MethodQueryArrayFloat32Validate", "*servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatFloat(float64(value), 'f', -1, 32)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayFloat64ValidateEncodeCode = `// EncodeMethodQueryArrayFloat64ValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayFloat64Validate MethodQueryArrayFloat64Validate
// server.
func EncodeMethodQueryArrayFloat64ValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayFloat64Validate", "MethodQueryArrayFloat64Validate", "*servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := strconv.FormatFloat(value, 'f', -1, 64)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayStringValidateEncodeCode = `// EncodeMethodQueryArrayStringValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayStringValidate MethodQueryArrayStringValidate
// server.
func EncodeMethodQueryArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayStringValidate", "MethodQueryArrayStringValidate", "*servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			values.Add("q", value)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayBytesValidateEncodeCode = `// EncodeMethodQueryArrayBytesValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayBytesValidate MethodQueryArrayBytesValidate
// server.
func EncodeMethodQueryArrayBytesValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayBytesValidate", "MethodQueryArrayBytesValidate", "*servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := string(value)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryArrayAnyValidateEncodeCode = `// EncodeMethodQueryArrayAnyValidateRequest returns an encoder for requests
// sent to the ServiceQueryArrayAnyValidate MethodQueryArrayAnyValidate server.
func EncodeMethodQueryArrayAnyValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryArrayAnyValidate", "MethodQueryArrayAnyValidate", "*servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload", v)
		}
		values := req.URL.Query()
		for _, value := range p.Q {
			valueStr := fmt.Sprintf("%v", value)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapStringStringValidateEncodeCode = `// EncodeMethodQueryMapStringStringValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapStringStringValidate
// MethodQueryMapStringStringValidate server.
func EncodeMethodQueryMapStringStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapstringstringvalidate.MethodQueryMapStringStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapStringStringValidate", "MethodQueryMapStringStringValidate", "*servicequerymapstringstringvalidate.MethodQueryMapStringStringValidatePayload", v)
		}
		values := req.URL.Query()
		for k, value := range p.Q {
			key := fmt.Sprintf("q[%s]", k)
			values.Add(key, value)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapStringBoolValidateEncodeCode = `// EncodeMethodQueryMapStringBoolValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapStringBoolValidate
// MethodQueryMapStringBoolValidate server.
func EncodeMethodQueryMapStringBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapstringboolvalidate.MethodQueryMapStringBoolValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapStringBoolValidate", "MethodQueryMapStringBoolValidate", "*servicequerymapstringboolvalidate.MethodQueryMapStringBoolValidatePayload", v)
		}
		values := req.URL.Query()
		for k, value := range p.Q {
			key := fmt.Sprintf("q[%s]", k)
			valueStr := strconv.FormatBool(value)
			values.Add(key, valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapBoolStringValidateEncodeCode = `// EncodeMethodQueryMapBoolStringValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapBoolStringValidate
// MethodQueryMapBoolStringValidate server.
func EncodeMethodQueryMapBoolStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapboolstringvalidate.MethodQueryMapBoolStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapBoolStringValidate", "MethodQueryMapBoolStringValidate", "*servicequerymapboolstringvalidate.MethodQueryMapBoolStringValidatePayload", v)
		}
		values := req.URL.Query()
		for kRaw, value := range p.Q {
			k := strconv.FormatBool(kRaw)
			key := fmt.Sprintf("q[%s]", k)
			values.Add(key, value)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapBoolBoolValidateEncodeCode = `// EncodeMethodQueryMapBoolBoolValidateRequest returns an encoder for requests
// sent to the ServiceQueryMapBoolBoolValidate MethodQueryMapBoolBoolValidate
// server.
func EncodeMethodQueryMapBoolBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapboolboolvalidate.MethodQueryMapBoolBoolValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapBoolBoolValidate", "MethodQueryMapBoolBoolValidate", "*servicequerymapboolboolvalidate.MethodQueryMapBoolBoolValidatePayload", v)
		}
		values := req.URL.Query()
		for kRaw, value := range p.Q {
			k := strconv.FormatBool(kRaw)
			key := fmt.Sprintf("q[%s]", k)
			valueStr := strconv.FormatBool(value)
			values.Add(key, valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapStringArrayStringValidateEncodeCode = `// EncodeMethodQueryMapStringArrayStringValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapStringArrayStringValidate
// MethodQueryMapStringArrayStringValidate server.
func EncodeMethodQueryMapStringArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapstringarraystringvalidate.MethodQueryMapStringArrayStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapStringArrayStringValidate", "MethodQueryMapStringArrayStringValidate", "*servicequerymapstringarraystringvalidate.MethodQueryMapStringArrayStringValidatePayload", v)
		}
		values := req.URL.Query()
		for k, value := range p.Q {
			key := fmt.Sprintf("q[%s]", k)
			values[key] = value
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapStringArrayBoolValidateEncodeCode = `// EncodeMethodQueryMapStringArrayBoolValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapStringArrayBoolValidate
// MethodQueryMapStringArrayBoolValidate server.
func EncodeMethodQueryMapStringArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapstringarrayboolvalidate.MethodQueryMapStringArrayBoolValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapStringArrayBoolValidate", "MethodQueryMapStringArrayBoolValidate", "*servicequerymapstringarrayboolvalidate.MethodQueryMapStringArrayBoolValidatePayload", v)
		}
		values := req.URL.Query()
		for k, value := range p.Q {
			key := fmt.Sprintf("q[%s]", k)
			for _, val := range value {
				valStr := strconv.FormatBool(val)
				values.Add(key, valStr)
			}
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapBoolArrayStringValidateEncodeCode = `// EncodeMethodQueryMapBoolArrayStringValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapBoolArrayStringValidate
// MethodQueryMapBoolArrayStringValidate server.
func EncodeMethodQueryMapBoolArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapboolarraystringvalidate.MethodQueryMapBoolArrayStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapBoolArrayStringValidate", "MethodQueryMapBoolArrayStringValidate", "*servicequerymapboolarraystringvalidate.MethodQueryMapBoolArrayStringValidatePayload", v)
		}
		values := req.URL.Query()
		for kRaw, value := range p.Q {
			k := strconv.FormatBool(kRaw)
			key := fmt.Sprintf("q[%s]", k)
			values[key] = value
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapBoolArrayBoolValidateEncodeCode = `// EncodeMethodQueryMapBoolArrayBoolValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapBoolArrayBoolValidate
// MethodQueryMapBoolArrayBoolValidate server.
func EncodeMethodQueryMapBoolArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerymapboolarrayboolvalidate.MethodQueryMapBoolArrayBoolValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapBoolArrayBoolValidate", "MethodQueryMapBoolArrayBoolValidate", "*servicequerymapboolarrayboolvalidate.MethodQueryMapBoolArrayBoolValidatePayload", v)
		}
		values := req.URL.Query()
		for kRaw, value := range p.Q {
			k := strconv.FormatBool(kRaw)
			key := fmt.Sprintf("q[%s]", k)
			for _, val := range value {
				valStr := strconv.FormatBool(val)
				values.Add(key, valStr)
			}
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveStringValidateEncodeCode = `// EncodeMethodQueryPrimitiveStringValidateRequest returns an encoder for
// requests sent to the ServiceQueryPrimitiveStringValidate
// MethodQueryPrimitiveStringValidate server.
func EncodeMethodQueryPrimitiveStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveStringValidate", "MethodQueryPrimitiveStringValidate", "string", v)
		}
		values := req.URL.Query()
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveBoolValidateEncodeCode = `// EncodeMethodQueryPrimitiveBoolValidateRequest returns an encoder for
// requests sent to the ServiceQueryPrimitiveBoolValidate
// MethodQueryPrimitiveBoolValidate server.
func EncodeMethodQueryPrimitiveBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveBoolValidate", "MethodQueryPrimitiveBoolValidate", "bool", v)
		}
		values := req.URL.Query()
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveArrayStringValidateEncodeCode = `// EncodeMethodQueryPrimitiveArrayStringValidateRequest returns an encoder for
// requests sent to the ServiceQueryPrimitiveArrayStringValidate
// MethodQueryPrimitiveArrayStringValidate server.
func EncodeMethodQueryPrimitiveArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveArrayStringValidate", "MethodQueryPrimitiveArrayStringValidate", "[]string", v)
		}
		values := req.URL.Query()
		for _, value := range p {
			values.Add("q", value)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveArrayBoolValidateEncodeCode = `// EncodeMethodQueryPrimitiveArrayBoolValidateRequest returns an encoder for
// requests sent to the ServiceQueryPrimitiveArrayBoolValidate
// MethodQueryPrimitiveArrayBoolValidate server.
func EncodeMethodQueryPrimitiveArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveArrayBoolValidate", "MethodQueryPrimitiveArrayBoolValidate", "[]bool", v)
		}
		values := req.URL.Query()
		for _, value := range p {
			valueStr := strconv.FormatBool(value)
			values.Add("q", valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveMapStringArrayStringValidateEncodeCode = `// EncodeMethodQueryPrimitiveMapStringArrayStringValidateRequest returns an
// encoder for requests sent to the
// ServiceQueryPrimitiveMapStringArrayStringValidate
// MethodQueryPrimitiveMapStringArrayStringValidate server.
func EncodeMethodQueryPrimitiveMapStringArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[string][]string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveMapStringArrayStringValidate", "MethodQueryPrimitiveMapStringArrayStringValidate", "map[string][]string", v)
		}
		values := req.URL.Query()
		for k, value := range p {
			key := fmt.Sprintf("q[%s]", k)
			values[key] = value
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveMapStringBoolValidateEncodeCode = `// EncodeMethodQueryPrimitiveMapStringBoolValidateRequest returns an encoder
// for requests sent to the ServiceQueryPrimitiveMapStringBoolValidate
// MethodQueryPrimitiveMapStringBoolValidate server.
func EncodeMethodQueryPrimitiveMapStringBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[string]bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveMapStringBoolValidate", "MethodQueryPrimitiveMapStringBoolValidate", "map[string]bool", v)
		}
		values := req.URL.Query()
		for k, value := range p {
			key := fmt.Sprintf("q[%s]", k)
			valueStr := strconv.FormatBool(value)
			values.Add(key, valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveMapBoolArrayBoolValidateEncodeCode = `// EncodeMethodQueryPrimitiveMapBoolArrayBoolValidateRequest returns an encoder
// for requests sent to the ServiceQueryPrimitiveMapBoolArrayBoolValidate
// MethodQueryPrimitiveMapBoolArrayBoolValidate server.
func EncodeMethodQueryPrimitiveMapBoolArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[bool][]bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveMapBoolArrayBoolValidate", "MethodQueryPrimitiveMapBoolArrayBoolValidate", "map[bool][]bool", v)
		}
		values := req.URL.Query()
		for kRaw, value := range p {
			k := strconv.FormatBool(kRaw)
			key := fmt.Sprintf("q[%s]", k)
			for _, val := range value {
				valStr := strconv.FormatBool(val)
				values.Add(key, valStr)
			}
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapStringMapIntStringValidateEncodeCode = `// EncodeMethodQueryMapStringMapIntStringValidateRequest returns an encoder for
// requests sent to the ServiceQueryMapStringMapIntStringValidate
// MethodQueryMapStringMapIntStringValidate server.
func EncodeMethodQueryMapStringMapIntStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[string]map[int]string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapStringMapIntStringValidate", "MethodQueryMapStringMapIntStringValidate", "map[string]map[int]string", v)
		}
		values := req.URL.Query()
		for k, value := range p {
			key := fmt.Sprintf("q[%s]", k)
			for kRaw, value := range value {
				k := strconv.Itoa(kRaw)
				key = fmt.Sprintf("%s[%s]", key, k)
				values.Add(key, value)
			}
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryMapIntMapStringArrayIntValidateEncodeCode = `// EncodeMethodQueryMapIntMapStringArrayIntValidateRequest returns an encoder
// for requests sent to the ServiceQueryMapIntMapStringArrayIntValidate
// MethodQueryMapIntMapStringArrayIntValidate server.
func EncodeMethodQueryMapIntMapStringArrayIntValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[int]map[string][]int)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryMapIntMapStringArrayIntValidate", "MethodQueryMapIntMapStringArrayIntValidate", "map[int]map[string][]int", v)
		}
		values := req.URL.Query()
		for kRaw, value := range p {
			k := strconv.Itoa(kRaw)
			key := fmt.Sprintf("q[%s]", k)
			for k, value := range value {
				key = fmt.Sprintf("%s[%s]", key, k)
				for _, val := range value {
					valStr := strconv.Itoa(val)
					values.Add(key, valStr)
				}
			}
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryStringMappedEncodeCode = `// EncodeMethodQueryStringMappedRequest returns an encoder for requests sent to
// the ServiceQueryStringMapped MethodQueryStringMapped server.
func EncodeMethodQueryStringMappedRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerystringmapped.MethodQueryStringMappedPayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryStringMapped", "MethodQueryStringMapped", "*servicequerystringmapped.MethodQueryStringMappedPayload", v)
		}
		values := req.URL.Query()
		if p.Query != nil {
			values.Add("q", *p.Query)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryStringDefaultEncodeCode = `// EncodeMethodQueryStringDefaultRequest returns an encoder for requests sent
// to the ServiceQueryStringDefault MethodQueryStringDefault server.
func EncodeMethodQueryStringDefaultRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerystringdefault.MethodQueryStringDefaultPayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryStringDefault", "MethodQueryStringDefault", "*servicequerystringdefault.MethodQueryStringDefaultPayload", v)
		}
		values := req.URL.Query()
		values.Add("q", p.Q)
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryPrimitiveStringDefaultEncodeCode = `// EncodeMethodQueryPrimitiveStringDefaultRequest returns an encoder for
// requests sent to the ServiceQueryPrimitiveStringDefault
// MethodQueryPrimitiveStringDefault server.
func EncodeMethodQueryPrimitiveStringDefaultRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryPrimitiveStringDefault", "MethodQueryPrimitiveStringDefault", "string", v)
		}
		values := req.URL.Query()
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadHeaderStringValidateEncodeCode = `// EncodeMethodHeaderStringValidateRequest returns an encoder for requests sent
// to the ServiceHeaderStringValidate MethodHeaderStringValidate server.
func EncodeMethodHeaderStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*serviceheaderstringvalidate.MethodHeaderStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderStringValidate", "MethodHeaderStringValidate", "*serviceheaderstringvalidate.MethodHeaderStringValidatePayload", v)
		}
		if p.H != nil {
			req.Header.Set("h", *p.H)
		}
		return nil
	}
}
`

var PayloadHeaderArrayStringValidateEncodeCode = `// EncodeMethodHeaderArrayStringValidateRequest returns an encoder for requests
// sent to the ServiceHeaderArrayStringValidate MethodHeaderArrayStringValidate
// server.
func EncodeMethodHeaderArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderArrayStringValidate", "MethodHeaderArrayStringValidate", "*serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload", v)
		}
		req.Header.Set("h", p.H)
		return nil
	}
}
`

var PayloadHeaderPrimitiveStringValidateEncodeCode = `// EncodeMethodHeaderPrimitiveStringValidateRequest returns an encoder for
// requests sent to the ServiceHeaderPrimitiveStringValidate
// MethodHeaderPrimitiveStringValidate server.
func EncodeMethodHeaderPrimitiveStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderPrimitiveStringValidate", "MethodHeaderPrimitiveStringValidate", "string", v)
		}
		return nil
	}
}
`

var PayloadHeaderPrimitiveBoolValidateEncodeCode = `// EncodeMethodHeaderPrimitiveBoolValidateRequest returns an encoder for
// requests sent to the ServiceHeaderPrimitiveBoolValidate
// MethodHeaderPrimitiveBoolValidate server.
func EncodeMethodHeaderPrimitiveBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderPrimitiveBoolValidate", "MethodHeaderPrimitiveBoolValidate", "bool", v)
		}
		return nil
	}
}
`

var PayloadHeaderPrimitiveArrayStringValidateEncodeCode = `// EncodeMethodHeaderPrimitiveArrayStringValidateRequest returns an encoder for
// requests sent to the ServiceHeaderPrimitiveArrayStringValidate
// MethodHeaderPrimitiveArrayStringValidate server.
func EncodeMethodHeaderPrimitiveArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderPrimitiveArrayStringValidate", "MethodHeaderPrimitiveArrayStringValidate", "[]string", v)
		}
		return nil
	}
}
`

var PayloadHeaderPrimitiveArrayBoolValidateEncodeCode = `// EncodeMethodHeaderPrimitiveArrayBoolValidateRequest returns an encoder for
// requests sent to the ServiceHeaderPrimitiveArrayBoolValidate
// MethodHeaderPrimitiveArrayBoolValidate server.
func EncodeMethodHeaderPrimitiveArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderPrimitiveArrayBoolValidate", "MethodHeaderPrimitiveArrayBoolValidate", "[]bool", v)
		}
		return nil
	}
}
`

var PayloadHeaderStringDefaultEncodeCode = `// EncodeMethodHeaderStringDefaultRequest returns an encoder for requests sent
// to the ServiceHeaderStringDefault MethodHeaderStringDefault server.
func EncodeMethodHeaderStringDefaultRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*serviceheaderstringdefault.MethodHeaderStringDefaultPayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderStringDefault", "MethodHeaderStringDefault", "*serviceheaderstringdefault.MethodHeaderStringDefaultPayload", v)
		}
		req.Header.Set("h", p.H)
		return nil
	}
}
`

var PayloadHeaderPrimitiveStringDefaultEncodeCode = `// EncodeMethodHeaderPrimitiveStringDefaultRequest returns an encoder for
// requests sent to the ServiceHeaderPrimitiveStringDefault
// MethodHeaderPrimitiveStringDefault server.
func EncodeMethodHeaderPrimitiveStringDefaultRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceHeaderPrimitiveStringDefault", "MethodHeaderPrimitiveStringDefault", "string", v)
		}
		return nil
	}
}
`

var PayloadBodyStringValidateEncodeCode = `// EncodeMethodBodyStringValidateRequest returns an encoder for requests sent
// to the ServiceBodyStringValidate MethodBodyStringValidate server.
func EncodeMethodBodyStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodystringvalidate.MethodBodyStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyStringValidate", "MethodBodyStringValidate", "*servicebodystringvalidate.MethodBodyStringValidatePayload", v)
		}
		body := NewMethodBodyStringValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyStringValidate", "MethodBodyStringValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyUserValidateEncodeCode = `// EncodeMethodBodyUserValidateRequest returns an encoder for requests sent to
// the ServiceBodyUserValidate MethodBodyUserValidate server.
func EncodeMethodBodyUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyuservalidate.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyUserValidate", "MethodBodyUserValidate", "*servicebodyuservalidate.PayloadType", v)
		}
		body := NewMethodBodyUserValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyUserValidate", "MethodBodyUserValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyArrayStringValidateEncodeCode = `// EncodeMethodBodyArrayStringValidateRequest returns an encoder for requests
// sent to the ServiceBodyArrayStringValidate MethodBodyArrayStringValidate
// server.
func EncodeMethodBodyArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyarraystringvalidate.MethodBodyArrayStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyArrayStringValidate", "MethodBodyArrayStringValidate", "*servicebodyarraystringvalidate.MethodBodyArrayStringValidatePayload", v)
		}
		body := NewMethodBodyArrayStringValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyArrayStringValidate", "MethodBodyArrayStringValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyArrayUserValidateEncodeCode = `// EncodeMethodBodyArrayUserValidateRequest returns an encoder for requests
// sent to the ServiceBodyArrayUserValidate MethodBodyArrayUserValidate server.
func EncodeMethodBodyArrayUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyarrayuservalidate.MethodBodyArrayUserValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyArrayUserValidate", "MethodBodyArrayUserValidate", "*servicebodyarrayuservalidate.MethodBodyArrayUserValidatePayload", v)
		}
		body := NewMethodBodyArrayUserValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyArrayUserValidate", "MethodBodyArrayUserValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyMapStringValidateEncodeCode = `// EncodeMethodBodyMapStringValidateRequest returns an encoder for requests
// sent to the ServiceBodyMapStringValidate MethodBodyMapStringValidate server.
func EncodeMethodBodyMapStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodymapstringvalidate.MethodBodyMapStringValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyMapStringValidate", "MethodBodyMapStringValidate", "*servicebodymapstringvalidate.MethodBodyMapStringValidatePayload", v)
		}
		body := NewMethodBodyMapStringValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyMapStringValidate", "MethodBodyMapStringValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyMapUserValidateEncodeCode = `// EncodeMethodBodyMapUserValidateRequest returns an encoder for requests sent
// to the ServiceBodyMapUserValidate MethodBodyMapUserValidate server.
func EncodeMethodBodyMapUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodymapuservalidate.MethodBodyMapUserValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyMapUserValidate", "MethodBodyMapUserValidate", "*servicebodymapuservalidate.MethodBodyMapUserValidatePayload", v)
		}
		body := NewMethodBodyMapUserValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyMapUserValidate", "MethodBodyMapUserValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPrimitiveStringValidateEncodeCode = `// EncodeMethodBodyPrimitiveStringValidateRequest returns an encoder for
// requests sent to the ServiceBodyPrimitiveStringValidate
// MethodBodyPrimitiveStringValidate server.
func EncodeMethodBodyPrimitiveStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPrimitiveStringValidate", "MethodBodyPrimitiveStringValidate", "string", v)
		}
		body := p
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPrimitiveStringValidate", "MethodBodyPrimitiveStringValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPrimitiveBoolValidateEncodeCode = `// EncodeMethodBodyPrimitiveBoolValidateRequest returns an encoder for requests
// sent to the ServiceBodyPrimitiveBoolValidate MethodBodyPrimitiveBoolValidate
// server.
func EncodeMethodBodyPrimitiveBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPrimitiveBoolValidate", "MethodBodyPrimitiveBoolValidate", "bool", v)
		}
		body := p
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPrimitiveBoolValidate", "MethodBodyPrimitiveBoolValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPrimitiveArrayStringValidateEncodeCode = `// EncodeMethodBodyPrimitiveArrayStringValidateRequest returns an encoder for
// requests sent to the ServiceBodyPrimitiveArrayStringValidate
// MethodBodyPrimitiveArrayStringValidate server.
func EncodeMethodBodyPrimitiveArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayStringValidate", "MethodBodyPrimitiveArrayStringValidate", "[]string", v)
		}
		body := p
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayStringValidate", "MethodBodyPrimitiveArrayStringValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPrimitiveArrayBoolValidateEncodeCode = `// EncodeMethodBodyPrimitiveArrayBoolValidateRequest returns an encoder for
// requests sent to the ServiceBodyPrimitiveArrayBoolValidate
// MethodBodyPrimitiveArrayBoolValidate server.
func EncodeMethodBodyPrimitiveArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]bool)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayBoolValidate", "MethodBodyPrimitiveArrayBoolValidate", "[]bool", v)
		}
		body := p
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayBoolValidate", "MethodBodyPrimitiveArrayBoolValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPrimitiveArrayUserValidateEncodeCode = `// EncodeMethodBodyPrimitiveArrayUserValidateRequest returns an encoder for
// requests sent to the ServiceBodyPrimitiveArrayUserValidate
// MethodBodyPrimitiveArrayUserValidate server.
func EncodeMethodBodyPrimitiveArrayUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]*servicebodyprimitivearrayuservalidate.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", "[]*servicebodyprimitivearrayuservalidate.PayloadType", v)
		}
		body := NewPayloadTypeRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPrimitiveFieldArrayUserValidateEncodeCode = `// EncodeMethodBodyPrimitiveArrayUserValidateRequest returns an encoder for
// requests sent to the ServiceBodyPrimitiveArrayUserValidate
// MethodBodyPrimitiveArrayUserValidate server.
func EncodeMethodBodyPrimitiveArrayUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyprimitivearrayuservalidate.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", "*servicebodyprimitivearrayuservalidate.PayloadType", v)
		}
		body := p
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyQueryObjectValidateEncodeCode = `// EncodeMethodBodyQueryObjectValidateRequest returns an encoder for requests
// sent to the ServiceBodyQueryObjectValidate MethodBodyQueryObjectValidate
// server.
func EncodeMethodBodyQueryObjectValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyQueryObjectValidate", "MethodBodyQueryObjectValidate", "*servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("b", p.B)
		req.URL.RawQuery = values.Encode()
		body := NewMethodBodyQueryObjectValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyQueryObjectValidate", "MethodBodyQueryObjectValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyQueryUserValidateEncodeCode = `// EncodeMethodBodyQueryUserValidateRequest returns an encoder for requests
// sent to the ServiceBodyQueryUserValidate MethodBodyQueryUserValidate server.
func EncodeMethodBodyQueryUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyqueryuservalidate.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyQueryUserValidate", "MethodBodyQueryUserValidate", "*servicebodyqueryuservalidate.PayloadType", v)
		}
		values := req.URL.Query()
		values.Add("b", p.B)
		req.URL.RawQuery = values.Encode()
		body := NewMethodBodyQueryUserValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyQueryUserValidate", "MethodBodyQueryUserValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPathObjectValidateEncodeCode = `// EncodeMethodBodyPathObjectValidateRequest returns an encoder for requests
// sent to the ServiceBodyPathObjectValidate MethodBodyPathObjectValidate
// server.
func EncodeMethodBodyPathObjectValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPathObjectValidate", "MethodBodyPathObjectValidate", "*servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload", v)
		}
		body := NewMethodBodyPathObjectValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPathObjectValidate", "MethodBodyPathObjectValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyPathUserValidateEncodeCode = `// EncodeMethodUserBodyPathValidateRequest returns an encoder for requests sent
// to the ServiceBodyPathUserValidate MethodUserBodyPathValidate server.
func EncodeMethodUserBodyPathValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodypathuservalidate.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyPathUserValidate", "MethodUserBodyPathValidate", "*servicebodypathuservalidate.PayloadType", v)
		}
		body := NewMethodUserBodyPathValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyPathUserValidate", "MethodUserBodyPathValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyQueryPathObjectValidateEncodeCode = `// EncodeMethodBodyQueryPathObjectValidateRequest returns an encoder for
// requests sent to the ServiceBodyQueryPathObjectValidate
// MethodBodyQueryPathObjectValidate server.
func EncodeMethodBodyQueryPathObjectValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyQueryPathObjectValidate", "MethodBodyQueryPathObjectValidate", "*servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload", v)
		}
		values := req.URL.Query()
		values.Add("b", p.B)
		req.URL.RawQuery = values.Encode()
		body := NewMethodBodyQueryPathObjectValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyQueryPathObjectValidate", "MethodBodyQueryPathObjectValidate", err)
		}
		return nil
	}
}
`

var PayloadBodyQueryPathUserValidateEncodeCode = `// EncodeMethodBodyQueryPathUserValidateRequest returns an encoder for requests
// sent to the ServiceBodyQueryPathUserValidate MethodBodyQueryPathUserValidate
// server.
func EncodeMethodBodyQueryPathUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyquerypathuservalidate.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyQueryPathUserValidate", "MethodBodyQueryPathUserValidate", "*servicebodyquerypathuservalidate.PayloadType", v)
		}
		values := req.URL.Query()
		values.Add("b", p.B)
		req.URL.RawQuery = values.Encode()
		body := NewMethodBodyQueryPathUserValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyQueryPathUserValidate", "MethodBodyQueryPathUserValidate", err)
		}
		return nil
	}
}
`

var PayloadMapQueryPrimitivePrimitiveEncodeCode = `// EncodeMapQueryPrimitivePrimitiveRequest returns an encoder for requests sent
// to the ServiceMapQueryPrimitivePrimitive MapQueryPrimitivePrimitive server.
func EncodeMapQueryPrimitivePrimitiveRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[string]string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceMapQueryPrimitivePrimitive", "MapQueryPrimitivePrimitive", "map[string]string", v)
		}
		values := req.URL.Query()
		for key, value := range p {
			keyStr := key
			valueStr := value
			values.Add(keyStr, valueStr)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadMapQueryPrimitiveArrayEncodeCode = `// EncodeMapQueryPrimitiveArrayRequest returns an encoder for requests sent to
// the ServiceMapQueryPrimitiveArray MapQueryPrimitiveArray server.
func EncodeMapQueryPrimitiveArrayRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[string][]uint)
		if !ok {
			return goahttp.ErrInvalidType("ServiceMapQueryPrimitiveArray", "MapQueryPrimitiveArray", "map[string][]uint", v)
		}
		values := req.URL.Query()
		for key, value := range p {
			keyStr := key
			for _, val := range value {
				valStr := strconv.FormatUint(uint64(val), 10)
				values.Add(keyStr, valStr)
			}
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadMapQueryObjectEncodeCode = `// EncodeMethodMapQueryObjectRequest returns an encoder for requests sent to
// the ServiceMapQueryObject MethodMapQueryObject server.
func EncodeMethodMapQueryObjectRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicemapqueryobject.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceMapQueryObject", "MethodMapQueryObject", "*servicemapqueryobject.PayloadType", v)
		}
		values := req.URL.Query()
		for key, value := range p.C {
			keyStr := strconv.Itoa(key)
			for _, val := range value {
				valStr := val
				values.Add(keyStr, valStr)
			}
		}
		req.URL.RawQuery = values.Encode()
		body := NewMethodMapQueryObjectRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceMapQueryObject", "MethodMapQueryObject", err)
		}
		return nil
	}
}
`

var PayloadMultipartBodyPrimitiveEncodeCode = `// EncodeMethodMultipartPrimitiveRequest returns an encoder for requests sent
// to the ServiceMultipartPrimitive MethodMultipartPrimitive server.
func EncodeMethodMultipartPrimitiveRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(string)
		if !ok {
			return goahttp.ErrInvalidType("ServiceMultipartPrimitive", "MethodMultipartPrimitive", "string", v)
		}
		if err := encoder(req).Encode(p); err != nil {
			return goahttp.ErrEncodingError("ServiceMultipartPrimitive", "MethodMultipartPrimitive", err)
		}
		return nil
	}
}
`

var PayloadMultipartBodyUserTypeEncodeCode = `// EncodeMethodMultipartUserTypeRequest returns an encoder for requests sent to
// the ServiceMultipartUserType MethodMultipartUserType server.
func EncodeMethodMultipartUserTypeRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicemultipartusertype.MethodMultipartUserTypePayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceMultipartUserType", "MethodMultipartUserType", "*servicemultipartusertype.MethodMultipartUserTypePayload", v)
		}
		if err := encoder(req).Encode(p); err != nil {
			return goahttp.ErrEncodingError("ServiceMultipartUserType", "MethodMultipartUserType", err)
		}
		return nil
	}
}
`

var PayloadMultipartBodyArrayTypeEncodeCode = `// EncodeMethodMultipartArrayTypeRequest returns an encoder for requests sent
// to the ServiceMultipartArrayType MethodMultipartArrayType server.
func EncodeMethodMultipartArrayTypeRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.([]*servicemultipartarraytype.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceMultipartArrayType", "MethodMultipartArrayType", "[]*servicemultipartarraytype.PayloadType", v)
		}
		if err := encoder(req).Encode(p); err != nil {
			return goahttp.ErrEncodingError("ServiceMultipartArrayType", "MethodMultipartArrayType", err)
		}
		return nil
	}
}
`

var PayloadMultipartBodyMapTypeEncodeCode = `// EncodeMethodMultipartMapTypeRequest returns an encoder for requests sent to
// the ServiceMultipartMapType MethodMultipartMapType server.
func EncodeMethodMultipartMapTypeRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(map[string]int)
		if !ok {
			return goahttp.ErrInvalidType("ServiceMultipartMapType", "MethodMultipartMapType", "map[string]int", v)
		}
		if err := encoder(req).Encode(p); err != nil {
			return goahttp.ErrEncodingError("ServiceMultipartMapType", "MethodMultipartMapType", err)
		}
		return nil
	}
}
`
