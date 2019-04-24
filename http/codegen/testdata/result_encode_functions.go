package testdata

var ResultHeaderBoolEncodeCode = `// EncodeMethodHeaderBoolResponse returns an encoder for responses returned by
// the ServiceHeaderBool MethodHeaderBool endpoint.
func EncodeMethodHeaderBoolResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderbool.MethodHeaderBoolResult)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatBool(*val)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderIntEncodeCode = `// EncodeMethodHeaderIntResponse returns an encoder for responses returned by
// the ServiceHeaderInt MethodHeaderInt endpoint.
func EncodeMethodHeaderIntResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderint.MethodHeaderIntResult)
		if res.H != nil {
			val := res.H
			hs := strconv.Itoa(*val)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderInt32EncodeCode = `// EncodeMethodHeaderInt32Response returns an encoder for responses returned by
// the ServiceHeaderInt32 MethodHeaderInt32 endpoint.
func EncodeMethodHeaderInt32Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderint32.MethodHeaderInt32Result)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatInt(int64(*val), 10)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderInt64EncodeCode = `// EncodeMethodHeaderInt64Response returns an encoder for responses returned by
// the ServiceHeaderInt64 MethodHeaderInt64 endpoint.
func EncodeMethodHeaderInt64Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderint64.MethodHeaderInt64Result)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatInt(*val, 10)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderUIntEncodeCode = `// EncodeMethodHeaderUIntResponse returns an encoder for responses returned by
// the ServiceHeaderUInt MethodHeaderUInt endpoint.
func EncodeMethodHeaderUIntResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderuint.MethodHeaderUIntResult)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatUint(uint64(*val), 10)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderUInt32EncodeCode = `// EncodeMethodHeaderUInt32Response returns an encoder for responses returned
// by the ServiceHeaderUInt32 MethodHeaderUInt32 endpoint.
func EncodeMethodHeaderUInt32Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderuint32.MethodHeaderUInt32Result)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatUint(uint64(*val), 10)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderUInt64EncodeCode = `// EncodeMethodHeaderUInt64Response returns an encoder for responses returned
// by the ServiceHeaderUInt64 MethodHeaderUInt64 endpoint.
func EncodeMethodHeaderUInt64Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderuint64.MethodHeaderUInt64Result)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatUint(*val, 10)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderFloat32EncodeCode = `// EncodeMethodHeaderFloat32Response returns an encoder for responses returned
// by the ServiceHeaderFloat32 MethodHeaderFloat32 endpoint.
func EncodeMethodHeaderFloat32Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderfloat32.MethodHeaderFloat32Result)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatFloat(float64(*val), 'f', -1, 32)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderFloat64EncodeCode = `// EncodeMethodHeaderFloat64Response returns an encoder for responses returned
// by the ServiceHeaderFloat64 MethodHeaderFloat64 endpoint.
func EncodeMethodHeaderFloat64Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderfloat64.MethodHeaderFloat64Result)
		if res.H != nil {
			val := res.H
			hs := strconv.FormatFloat(*val, 'f', -1, 64)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderStringEncodeCode = `// EncodeMethodHeaderStringResponse returns an encoder for responses returned
// by the ServiceHeaderString MethodHeaderString endpoint.
func EncodeMethodHeaderStringResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderstring.MethodHeaderStringResult)
		if res.H != nil {
			w.Header().Set("H", *res.H)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderBytesEncodeCode = `// EncodeMethodHeaderBytesResponse returns an encoder for responses returned by
// the ServiceHeaderBytes MethodHeaderBytes endpoint.
func EncodeMethodHeaderBytesResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderbytes.MethodHeaderBytesResult)
		if res.H != nil {
			val := res.H
			hs := string(val)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderAnyEncodeCode = `// EncodeMethodHeaderAnyResponse returns an encoder for responses returned by
// the ServiceHeaderAny MethodHeaderAny endpoint.
func EncodeMethodHeaderAnyResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderany.MethodHeaderAnyResult)
		if res.H != nil {
			val := res.H
			hs := fmt.Sprintf("%v", val)
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBoolEncodeCode = `// EncodeMethodHeaderArrayBoolResponse returns an encoder for responses
// returned by the ServiceHeaderArrayBool MethodHeaderArrayBool endpoint.
func EncodeMethodHeaderArrayBoolResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarraybool.MethodHeaderArrayBoolResult)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatBool(e)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayIntEncodeCode = `// EncodeMethodHeaderArrayIntResponse returns an encoder for responses returned
// by the ServiceHeaderArrayInt MethodHeaderArrayInt endpoint.
func EncodeMethodHeaderArrayIntResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayint.MethodHeaderArrayIntResult)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.Itoa(e)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayInt32EncodeCode = `// EncodeMethodHeaderArrayInt32Response returns an encoder for responses
// returned by the ServiceHeaderArrayInt32 MethodHeaderArrayInt32 endpoint.
func EncodeMethodHeaderArrayInt32Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayint32.MethodHeaderArrayInt32Result)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatInt(int64(e), 10)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayInt64EncodeCode = `// EncodeMethodHeaderArrayInt64Response returns an encoder for responses
// returned by the ServiceHeaderArrayInt64 MethodHeaderArrayInt64 endpoint.
func EncodeMethodHeaderArrayInt64Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayint64.MethodHeaderArrayInt64Result)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatInt(e, 10)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayUIntEncodeCode = `// EncodeMethodHeaderArrayUIntResponse returns an encoder for responses
// returned by the ServiceHeaderArrayUInt MethodHeaderArrayUInt endpoint.
func EncodeMethodHeaderArrayUIntResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayuint.MethodHeaderArrayUIntResult)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatUint(uint64(e), 10)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayUInt32EncodeCode = `// EncodeMethodHeaderArrayUInt32Response returns an encoder for responses
// returned by the ServiceHeaderArrayUInt32 MethodHeaderArrayUInt32 endpoint.
func EncodeMethodHeaderArrayUInt32Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayuint32.MethodHeaderArrayUInt32Result)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatUint(uint64(e), 10)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayUInt64EncodeCode = `// EncodeMethodHeaderArrayUInt64Response returns an encoder for responses
// returned by the ServiceHeaderArrayUInt64 MethodHeaderArrayUInt64 endpoint.
func EncodeMethodHeaderArrayUInt64Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayuint64.MethodHeaderArrayUInt64Result)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatUint(e, 10)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayFloat32EncodeCode = `// EncodeMethodHeaderArrayFloat32Response returns an encoder for responses
// returned by the ServiceHeaderArrayFloat32 MethodHeaderArrayFloat32 endpoint.
func EncodeMethodHeaderArrayFloat32Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayfloat32.MethodHeaderArrayFloat32Result)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatFloat(float64(e), 'f', -1, 32)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayFloat64EncodeCode = `// EncodeMethodHeaderArrayFloat64Response returns an encoder for responses
// returned by the ServiceHeaderArrayFloat64 MethodHeaderArrayFloat64 endpoint.
func EncodeMethodHeaderArrayFloat64Response(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayfloat64.MethodHeaderArrayFloat64Result)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatFloat(e, 'f', -1, 64)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayStringEncodeCode = `// EncodeMethodHeaderArrayStringResponse returns an encoder for responses
// returned by the ServiceHeaderArrayString MethodHeaderArrayString endpoint.
func EncodeMethodHeaderArrayStringResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarraystring.MethodHeaderArrayStringResult)
		if res.H != nil {
			val := res.H
			hs := strings.Join(val, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBytesEncodeCode = `// EncodeMethodHeaderArrayBytesResponse returns an encoder for responses
// returned by the ServiceHeaderArrayBytes MethodHeaderArrayBytes endpoint.
func EncodeMethodHeaderArrayBytesResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarraybytes.MethodHeaderArrayBytesResult)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := string(e)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayAnyEncodeCode = `// EncodeMethodHeaderArrayAnyResponse returns an encoder for responses returned
// by the ServiceHeaderArrayAny MethodHeaderArrayAny endpoint.
func EncodeMethodHeaderArrayAnyResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayany.MethodHeaderArrayAnyResult)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := fmt.Sprintf("%v", e)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderBoolDefaultEncodeCode = `// EncodeMethodHeaderBoolDefaultResponse returns an encoder for responses
// returned by the ServiceHeaderBoolDefault MethodHeaderBoolDefault endpoint.
func EncodeMethodHeaderBoolDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderbooldefault.MethodHeaderBoolDefaultResult)
		val := res.H
		hs := strconv.FormatBool(val)
		w.Header().Set("H", hs)
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderBoolRequiredDefaultEncodeCode = `// EncodeMethodHeaderBoolRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderBoolRequiredDefault
// MethodHeaderBoolRequiredDefault endpoint.
func EncodeMethodHeaderBoolRequiredDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderboolrequireddefault.MethodHeaderBoolRequiredDefaultResult)
		val := res.H
		hs := strconv.FormatBool(val)
		w.Header().Set("H", hs)
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderStringDefaultEncodeCode = `// EncodeMethodHeaderStringDefaultResponse returns an encoder for responses
// returned by the ServiceHeaderStringDefault MethodHeaderStringDefault
// endpoint.
func EncodeMethodHeaderStringDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderstringdefault.MethodHeaderStringDefaultResult)
		w.Header().Set("H", res.H)
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderStringRequiredDefaultEncodeCode = `// EncodeMethodHeaderStringRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderStringRequiredDefault
// MethodHeaderStringRequiredDefault endpoint.
func EncodeMethodHeaderStringRequiredDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderstringrequireddefault.MethodHeaderStringRequiredDefaultResult)
		w.Header().Set("H", res.H)
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBoolDefaultEncodeCode = `// EncodeMethodHeaderArrayBoolDefaultResponse returns an encoder for responses
// returned by the ServiceHeaderArrayBoolDefault MethodHeaderArrayBoolDefault
// endpoint.
func EncodeMethodHeaderArrayBoolDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarraybooldefault.MethodHeaderArrayBoolDefaultResult)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatBool(e)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		} else {
			w.Header().Set("H", "true, false")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBoolRequiredDefaultEncodeCode = `// EncodeMethodHeaderArrayBoolRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderArrayBoolRequiredDefault
// MethodHeaderArrayBoolRequiredDefault endpoint.
func EncodeMethodHeaderArrayBoolRequiredDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarrayboolrequireddefault.MethodHeaderArrayBoolRequiredDefaultResult)
		if res.H != nil {
			val := res.H
			hsSlice := make([]string, len(val))
			for i, e := range val {
				es := strconv.FormatBool(e)
				hsSlice[i] = es
			}
			hs := strings.Join(hsSlice, ", ")
			w.Header().Set("H", hs)
		} else {
			w.Header().Set("H", "true, false")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayStringDefaultEncodeCode = `// EncodeMethodHeaderArrayStringDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderArrayStringDefault
// MethodHeaderArrayStringDefault endpoint.
func EncodeMethodHeaderArrayStringDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarraystringdefault.MethodHeaderArrayStringDefaultResult)
		if res.H != nil {
			val := res.H
			hs := strings.Join(val, ", ")
			w.Header().Set("H", hs)
		} else {
			w.Header().Set("H", "foo, bar")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayStringRequiredDefaultEncodeCode = `// EncodeMethodHeaderArrayStringRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderArrayStringRequiredDefault
// MethodHeaderArrayStringRequiredDefault endpoint.
func EncodeMethodHeaderArrayStringRequiredDefaultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceheaderarraystringrequireddefault.MethodHeaderArrayStringRequiredDefaultResult)
		if res.H != nil {
			val := res.H
			hs := strings.Join(val, ", ")
			w.Header().Set("H", hs)
		} else {
			w.Header().Set("H", "foo, bar")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultBodyStringEncodeCode = `// EncodeMethodBodyStringResponse returns an encoder for responses returned by
// the ServiceBodyString MethodBodyString endpoint.
func EncodeMethodBodyStringResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodystring.MethodBodyStringResult)
		enc := encoder(ctx, w)
		body := NewMethodBodyStringResponseBody(res)
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyObjectEncodeCode = `// EncodeMethodBodyObjectResponse returns an encoder for responses returned by
// the ServiceBodyObject MethodBodyObject endpoint.
func EncodeMethodBodyObjectResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodyobject.MethodBodyObjectResult)
		enc := encoder(ctx, w)
		body := NewMethodBodyObjectResponseBody(res)
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyUserEncodeCode = `// EncodeMethodBodyUserResponse returns an encoder for responses returned by
// the ServiceBodyUser MethodBodyUser endpoint.
func EncodeMethodBodyUserResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodyuser.ResultType)
		enc := encoder(ctx, w)
		body := NewMethodBodyUserResponseBody(res)
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyMultipleViewsEncodeCode = `// EncodeMethodBodyMultipleViewResponse returns an encoder for responses
// returned by the ServiceBodyMultipleView MethodBodyMultipleView endpoint.
func EncodeMethodBodyMultipleViewResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodymultipleviewviews.Resulttypemultipleviews)
		w.Header().Set("goa-view", res.View)
		enc := encoder(ctx, w)
		var body interface{}
		switch res.View {
		case "default", "":
			body = NewMethodBodyMultipleViewResponseBody(res.Projected)
		case "tiny":
			body = NewMethodBodyMultipleViewResponseBodyTiny(res.Projected)
		}
		if res.Projected.C != nil {
			w.Header().Set("Location", *res.Projected.C)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyCollectionMultipleViewsEncodeCode = `// EncodeMethodBodyCollectionResponse returns an encoder for responses returned
// by the ServiceBodyCollection MethodBodyCollection endpoint.
func EncodeMethodBodyCollectionResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(servicebodycollectionviews.ResulttypecollectionCollection)
		w.Header().Set("goa-view", res.View)
		enc := encoder(ctx, w)
		var body interface{}
		switch res.View {
		case "default", "":
			body = NewResulttypecollectionResponseCollection(res.Projected)
		case "tiny":
			body = NewResulttypecollectionResponseTinyCollection(res.Projected)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyCollectionExplicitViewEncodeCode = `// EncodeMethodBodyCollectionExplicitViewResponse returns an encoder for
// responses returned by the ServiceBodyCollectionExplicitView
// MethodBodyCollectionExplicitView endpoint.
func EncodeMethodBodyCollectionExplicitViewResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(servicebodycollectionexplicitviewviews.ResulttypecollectionCollection)
		enc := encoder(ctx, w)
		body := NewResulttypecollectionResponseTinyCollection(res.Projected)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var EmptyBodyResultMultipleViewsEncodeCode = `// EncodeMethodEmptyBodyResultMultipleViewResponse returns an encoder for
// responses returned by the ServiceEmptyBodyResultMultipleView
// MethodEmptyBodyResultMultipleView endpoint.
func EncodeMethodEmptyBodyResultMultipleViewResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceemptybodyresultmultipleviewviews.Resulttypemultipleviews)
		w.Header().Set("goa-view", res.View)
		if res.Projected.C != nil {
			w.Header().Set("Location", *res.Projected.C)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultBodyArrayStringEncodeCode = `// EncodeMethodBodyArrayStringResponse returns an encoder for responses
// returned by the ServiceBodyArrayString MethodBodyArrayString endpoint.
func EncodeMethodBodyArrayStringResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodyarraystring.MethodBodyArrayStringResult)
		enc := encoder(ctx, w)
		body := NewMethodBodyArrayStringResponseBody(res)
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyArrayUserEncodeCode = `// EncodeMethodBodyArrayUserResponse returns an encoder for responses returned
// by the ServiceBodyArrayUser MethodBodyArrayUser endpoint.
func EncodeMethodBodyArrayUserResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodyarrayuser.MethodBodyArrayUserResult)
		enc := encoder(ctx, w)
		body := NewMethodBodyArrayUserResponseBody(res)
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ExplicitBodyPrimitiveResultMultipleViewsEncodeCode = `// EncodeMethodExplicitBodyPrimitiveResultMultipleViewResponse returns an
// encoder for responses returned by the
// ServiceExplicitBodyPrimitiveResultMultipleView
// MethodExplicitBodyPrimitiveResultMultipleView endpoint.
func EncodeMethodExplicitBodyPrimitiveResultMultipleViewResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceexplicitbodyprimitiveresultmultipleviewviews.Resulttypemultipleviews)
		w.Header().Set("goa-view", res.View)
		enc := encoder(ctx, w)
		body := res.Projected.A
		if res.Projected.C != nil {
			w.Header().Set("Location", *res.Projected.C)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ExplicitBodyUserResultMultipleViewsEncodeCode = `// EncodeMethodExplicitBodyUserResultMultipleViewResponse returns an encoder
// for responses returned by the ServiceExplicitBodyUserResultMultipleView
// MethodExplicitBodyUserResultMultipleView endpoint.
func EncodeMethodExplicitBodyUserResultMultipleViewResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceexplicitbodyuserresultmultipleviewviews.Resulttypemultipleviews)
		w.Header().Set("goa-view", res.View)
		enc := encoder(ctx, w)
		body := NewUserType(res.Projected)
		if res.Projected.C != nil {
			w.Header().Set("Location", *res.Projected.C)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ExplicitBodyResultCollectionEncodeCode = `// EncodeMethodExplicitBodyResultCollectionResponse returns an encoder for
// responses returned by the ServiceExplicitBodyResultCollection
// MethodExplicitBodyResultCollection endpoint.
func EncodeMethodExplicitBodyResultCollectionResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceexplicitbodyresultcollection.MethodExplicitBodyResultCollectionResult)
		enc := encoder(ctx, w)
		body := NewResulttypeCollection(res)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ExplicitContentTypeResultEncodeCode = `// EncodeMethodExplicitContentTypeResultResponse returns an encoder for
// responses returned by the ServiceExplicitContentTypeResult
// MethodExplicitContentTypeResult endpoint.
func EncodeMethodExplicitContentTypeResultResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceexplicitcontenttyperesultviews.Resulttype)
		ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "application/custom+json")
		enc := encoder(ctx, w)
		body := NewMethodExplicitContentTypeResultResponseBody(res.Projected)
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ExplicitContentTypeResponseEncodeCode = `// EncodeMethodExplicitContentTypeResponseResponse returns an encoder for
// responses returned by the ServiceExplicitContentTypeResponse
// MethodExplicitContentTypeResponse endpoint.
func EncodeMethodExplicitContentTypeResponseResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceexplicitcontenttyperesponseviews.Resulttype)
		ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "application/custom+json")
		enc := encoder(ctx, w)
		body := NewMethodExplicitContentTypeResponseResponseBody(res.Projected)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveStringEncodeCode = `// EncodeMethodBodyPrimitiveStringResponse returns an encoder for responses
// returned by the ServiceBodyPrimitiveString MethodBodyPrimitiveString
// endpoint.
func EncodeMethodBodyPrimitiveStringResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(string)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveBoolEncodeCode = `// EncodeMethodBodyPrimitiveBoolResponse returns an encoder for responses
// returned by the ServiceBodyPrimitiveBool MethodBodyPrimitiveBool endpoint.
func EncodeMethodBodyPrimitiveBoolResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(bool)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveArrayStringEncodeCode = `// EncodeMethodBodyPrimitiveArrayStringResponse returns an encoder for
// responses returned by the ServiceBodyPrimitiveArrayString
// MethodBodyPrimitiveArrayString endpoint.
func EncodeMethodBodyPrimitiveArrayStringResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.([]string)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveArrayBoolEncodeCode = `// EncodeMethodBodyPrimitiveArrayBoolResponse returns an encoder for responses
// returned by the ServiceBodyPrimitiveArrayBool MethodBodyPrimitiveArrayBool
// endpoint.
func EncodeMethodBodyPrimitiveArrayBoolResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.([]bool)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveArrayUserEncodeCode = `// EncodeMethodBodyPrimitiveArrayUserResponse returns an encoder for responses
// returned by the ServiceBodyPrimitiveArrayUser MethodBodyPrimitiveArrayUser
// endpoint.
func EncodeMethodBodyPrimitiveArrayUserResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.([]*servicebodyprimitivearrayuser.ResultType)
		enc := encoder(ctx, w)
		body := NewResultTypeResponse(res)
		w.WriteHeader(http.StatusNoContent)
		return enc.Encode(body)
	}
}
`

var ResultBodyHeaderObjectEncodeCode = `// EncodeMethodBodyHeaderObjectResponse returns an encoder for responses
// returned by the ServiceBodyHeaderObject MethodBodyHeaderObject endpoint.
func EncodeMethodBodyHeaderObjectResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodyheaderobject.MethodBodyHeaderObjectResult)
		enc := encoder(ctx, w)
		body := NewMethodBodyHeaderObjectResponseBody(res)
		if res.B != nil {
			w.Header().Set("B", *res.B)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyHeaderUserEncodeCode = `// EncodeMethodBodyHeaderUserResponse returns an encoder for responses returned
// by the ServiceBodyHeaderUser MethodBodyHeaderUser endpoint.
func EncodeMethodBodyHeaderUserResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicebodyheaderuser.ResultType)
		enc := encoder(ctx, w)
		body := NewMethodBodyHeaderUserResponseBody(res)
		if res.B != nil {
			w.Header().Set("B", *res.B)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultTagStringEncodeCode = `// EncodeMethodTagStringResponse returns an encoder for responses returned by
// the ServiceTagString MethodTagString endpoint.
func EncodeMethodTagStringResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicetagstring.MethodTagStringResult)
		if res.H != nil && *res.H == "value" {
			w.Header().Set("H", *res.H)
			w.WriteHeader(http.StatusAccepted)
			return nil
		}
		enc := encoder(ctx, w)
		body := NewMethodTagStringOKResponseBody(res)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultTagStringRequiredEncodeCode = `// EncodeMethodTagStringRequiredResponse returns an encoder for responses
// returned by the ServiceTagStringRequired MethodTagStringRequired endpoint.
func EncodeMethodTagStringRequiredResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicetagstringrequired.MethodTagStringRequiredResult)
		if res.H == "value" {
			w.Header().Set("H", res.H)
			w.WriteHeader(http.StatusAccepted)
			return nil
		}
		enc := encoder(ctx, w)
		body := NewMethodTagStringRequiredOKResponseBody(res)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultMultipleViewsTagEncodeCode = `// EncodeMethodTagMultipleViewsResponse returns an encoder for responses
// returned by the ServiceTagMultipleViews MethodTagMultipleViews endpoint.
func EncodeMethodTagMultipleViewsResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*servicetagmultipleviewsviews.Resulttypemultipleviews)
		w.Header().Set("goa-view", res.View)
		if res.Projected.B != nil && *res.Projected.B == "value" {
			enc := encoder(ctx, w)
			var body interface{}
			switch res.View {
			case "default", "":
				body = NewMethodTagMultipleViewsAcceptedResponseBody(res.Projected)
			case "tiny":
				body = NewMethodTagMultipleViewsAcceptedResponseBodyTiny(res.Projected)
			}
			w.Header().Set("C", *res.Projected.C)
			w.WriteHeader(http.StatusAccepted)
			return enc.Encode(body)
		}
		enc := encoder(ctx, w)
		var body interface{}
		switch res.View {
		case "default", "":
			body = NewMethodTagMultipleViewsOKResponseBody(res.Projected)
		case "tiny":
			body = NewMethodTagMultipleViewsOKResponseBodyTiny(res.Projected)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var EmptyServerResponseEncodeCode = `// EncodeMethodEmptyServerResponseResponse returns an encoder for responses
// returned by the ServiceEmptyServerResponse MethodEmptyServerResponse
// endpoint.
func EncodeMethodEmptyServerResponseResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var EmptyServerResponseWithTagsEncodeCode = `// EncodeMethodEmptyServerResponseWithTagsResponse returns an encoder for
// responses returned by the ServiceEmptyServerResponseWithTags
// MethodEmptyServerResponseWithTags endpoint.
func EncodeMethodEmptyServerResponseWithTagsResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*serviceemptyserverresponsewithtags.MethodEmptyServerResponseWithTagsResult)
		if res.H == "true" {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
`
