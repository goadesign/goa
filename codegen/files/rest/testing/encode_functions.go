package testing

var ResultHeaderBoolEncodeCode = `// EncodeMethodHeaderBoolResponse returns an encoder for responses returned by
// the ServiceHeaderBool MethodHeaderBool endpoint.
func EncodeMethodHeaderBoolResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderBoolResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatBool(*v)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderIntEncodeCode = `// EncodeMethodHeaderIntResponse returns an encoder for responses returned by
// the ServiceHeaderInt MethodHeaderInt endpoint.
func EncodeMethodHeaderIntResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderIntResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.Itoa(*v)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderInt32EncodeCode = `// EncodeMethodHeaderInt32Response returns an encoder for responses returned by
// the ServiceHeaderInt32 MethodHeaderInt32 endpoint.
func EncodeMethodHeaderInt32Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderInt32Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatInt(int64(*v), 10)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderInt64EncodeCode = `// EncodeMethodHeaderInt64Response returns an encoder for responses returned by
// the ServiceHeaderInt64 MethodHeaderInt64 endpoint.
func EncodeMethodHeaderInt64Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderInt64Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatInt(*v, 10)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderUIntEncodeCode = `// EncodeMethodHeaderUIntResponse returns an encoder for responses returned by
// the ServiceHeaderUInt MethodHeaderUInt endpoint.
func EncodeMethodHeaderUIntResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderUIntResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatUint(uint64(*v), 10)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderUInt32EncodeCode = `// EncodeMethodHeaderUInt32Response returns an encoder for responses returned
// by the ServiceHeaderUInt32 MethodHeaderUInt32 endpoint.
func EncodeMethodHeaderUInt32Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderUInt32Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatUint(uint64(*v), 10)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderUInt64EncodeCode = `// EncodeMethodHeaderUInt64Response returns an encoder for responses returned
// by the ServiceHeaderUInt64 MethodHeaderUInt64 endpoint.
func EncodeMethodHeaderUInt64Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderUInt64Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatUint(*v, 10)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderFloat32EncodeCode = `// EncodeMethodHeaderFloat32Response returns an encoder for responses returned
// by the ServiceHeaderFloat32 MethodHeaderFloat32 endpoint.
func EncodeMethodHeaderFloat32Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderFloat32Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatFloat(float64(*v), 'f', -1, 32)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderFloat64EncodeCode = `// EncodeMethodHeaderFloat64Response returns an encoder for responses returned
// by the ServiceHeaderFloat64 MethodHeaderFloat64 endpoint.
func EncodeMethodHeaderFloat64Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderFloat64Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatFloat(*v, 'f', -1, 64)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderStringEncodeCode = `// EncodeMethodHeaderStringResponse returns an encoder for responses returned
// by the ServiceHeaderString MethodHeaderString endpoint.
func EncodeMethodHeaderStringResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderStringResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			w.Header().Set("h", *res.H)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderBytesEncodeCode = `// EncodeMethodHeaderBytesResponse returns an encoder for responses returned by
// the ServiceHeaderBytes MethodHeaderBytes endpoint.
func EncodeMethodHeaderBytesResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderBytesResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := string(v)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderAnyEncodeCode = `// EncodeMethodHeaderAnyResponse returns an encoder for responses returned by
// the ServiceHeaderAny MethodHeaderAny endpoint.
func EncodeMethodHeaderAnyResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderAnyResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := fmt.Sprintf("%v", v)
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBoolEncodeCode = `// EncodeMethodHeaderArrayBoolResponse returns an encoder for responses
// returned by the ServiceHeaderArrayBool MethodHeaderArrayBool endpoint.
func EncodeMethodHeaderArrayBoolResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayBoolResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatBool(e)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayIntEncodeCode = `// EncodeMethodHeaderArrayIntResponse returns an encoder for responses returned
// by the ServiceHeaderArrayInt MethodHeaderArrayInt endpoint.
func EncodeMethodHeaderArrayIntResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayIntResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.Itoa(e)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayInt32EncodeCode = `// EncodeMethodHeaderArrayInt32Response returns an encoder for responses
// returned by the ServiceHeaderArrayInt32 MethodHeaderArrayInt32 endpoint.
func EncodeMethodHeaderArrayInt32Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayInt32Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatInt(int64(e), 10)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayInt64EncodeCode = `// EncodeMethodHeaderArrayInt64Response returns an encoder for responses
// returned by the ServiceHeaderArrayInt64 MethodHeaderArrayInt64 endpoint.
func EncodeMethodHeaderArrayInt64Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayInt64Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatInt(e, 10)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayUIntEncodeCode = `// EncodeMethodHeaderArrayUIntResponse returns an encoder for responses
// returned by the ServiceHeaderArrayUInt MethodHeaderArrayUInt endpoint.
func EncodeMethodHeaderArrayUIntResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayUIntResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatUint(uint64(e), 10)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayUInt32EncodeCode = `// EncodeMethodHeaderArrayUInt32Response returns an encoder for responses
// returned by the ServiceHeaderArrayUInt32 MethodHeaderArrayUInt32 endpoint.
func EncodeMethodHeaderArrayUInt32Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayUInt32Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatUint(uint64(e), 10)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayUInt64EncodeCode = `// EncodeMethodHeaderArrayUInt64Response returns an encoder for responses
// returned by the ServiceHeaderArrayUInt64 MethodHeaderArrayUInt64 endpoint.
func EncodeMethodHeaderArrayUInt64Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayUInt64Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatUint(e, 10)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayFloat32EncodeCode = `// EncodeMethodHeaderArrayFloat32Response returns an encoder for responses
// returned by the ServiceHeaderArrayFloat32 MethodHeaderArrayFloat32 endpoint.
func EncodeMethodHeaderArrayFloat32Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayFloat32Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatFloat(float64(e), 'f', -1, 32)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayFloat64EncodeCode = `// EncodeMethodHeaderArrayFloat64Response returns an encoder for responses
// returned by the ServiceHeaderArrayFloat64 MethodHeaderArrayFloat64 endpoint.
func EncodeMethodHeaderArrayFloat64Response(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayFloat64Result)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatFloat(e, 'f', -1, 64)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayStringEncodeCode = `// EncodeMethodHeaderArrayStringResponse returns an encoder for responses
// returned by the ServiceHeaderArrayString MethodHeaderArrayString endpoint.
func EncodeMethodHeaderArrayStringResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayStringResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strings.Join(v, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBytesEncodeCode = `// EncodeMethodHeaderArrayBytesResponse returns an encoder for responses
// returned by the ServiceHeaderArrayBytes MethodHeaderArrayBytes endpoint.
func EncodeMethodHeaderArrayBytesResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayBytesResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := string(e)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayAnyEncodeCode = `// EncodeMethodHeaderArrayAnyResponse returns an encoder for responses returned
// by the ServiceHeaderArrayAny MethodHeaderArrayAny endpoint.
func EncodeMethodHeaderArrayAnyResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayAnyResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := fmt.Sprintf("%v", e)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderBoolDefaultEncodeCode = `// EncodeMethodHeaderBoolDefaultResponse returns an encoder for responses
// returned by the ServiceHeaderBoolDefault MethodHeaderBoolDefault endpoint.
func EncodeMethodHeaderBoolDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderBoolDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatBool(*v)
			w.Header().Set("h", h)
		} else {
			w.Header().Set("h", "true")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderBoolRequiredDefaultEncodeCode = `// EncodeMethodHeaderBoolRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderBoolRequiredDefault
// MethodHeaderBoolRequiredDefault endpoint.
func EncodeMethodHeaderBoolRequiredDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderBoolRequiredDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strconv.FormatBool(*v)
			w.Header().Set("h", h)
		} else {
			w.Header().Set("h", "true")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderStringDefaultEncodeCode = `// EncodeMethodHeaderStringDefaultResponse returns an encoder for responses
// returned by the ServiceHeaderStringDefault MethodHeaderStringDefault
// endpoint.
func EncodeMethodHeaderStringDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderStringDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			w.Header().Set("h", *res.H)
		} else {
			w.Header().Set("h", "def")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderStringRequiredDefaultEncodeCode = `// EncodeMethodHeaderStringRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderStringRequiredDefault
// MethodHeaderStringRequiredDefault endpoint.
func EncodeMethodHeaderStringRequiredDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderStringRequiredDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			w.Header().Set("h", *res.H)
		} else {
			w.Header().Set("h", "def")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBoolDefaultEncodeCode = `// EncodeMethodHeaderArrayBoolDefaultResponse returns an encoder for responses
// returned by the ServiceHeaderArrayBoolDefault MethodHeaderArrayBoolDefault
// endpoint.
func EncodeMethodHeaderArrayBoolDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayBoolDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatBool(e)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		} else {
			w.Header().Set("h", "true, false")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayBoolRequiredDefaultEncodeCode = `// EncodeMethodHeaderArrayBoolRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderArrayBoolRequiredDefault
// MethodHeaderArrayBoolRequiredDefault endpoint.
func EncodeMethodHeaderArrayBoolRequiredDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayBoolRequiredDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			hSlice := make([]string, len(v))
			for i, e := range v {
				es := strconv.FormatBool(e)
				hSlice[i] = es
			}
			h := strings.Join(hSlice, ", ")
			w.Header().Set("h", h)
		} else {
			w.Header().Set("h", "true, false")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayStringDefaultEncodeCode = `// EncodeMethodHeaderArrayStringDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderArrayStringDefault
// MethodHeaderArrayStringDefault endpoint.
func EncodeMethodHeaderArrayStringDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayStringDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strings.Join(v, ", ")
			w.Header().Set("h", h)
		} else {
			w.Header().Set("h", "foo, bar")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultHeaderArrayStringRequiredDefaultEncodeCode = `// EncodeMethodHeaderArrayStringRequiredDefaultResponse returns an encoder for
// responses returned by the ServiceHeaderArrayStringRequiredDefault
// MethodHeaderArrayStringRequiredDefault endpoint.
func EncodeMethodHeaderArrayStringRequiredDefaultResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodHeaderArrayStringRequiredDefaultResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		if res.H != nil {
			v := res.H
			h := strings.Join(v, ", ")
			w.Header().Set("h", h)
		} else {
			w.Header().Set("h", "foo, bar")
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
}
`

var ResultBodyStringEncodeCode = `// EncodeMethodBodyStringResponse returns an encoder for responses returned by
// the ServiceBodyString MethodBodyString endpoint.
func EncodeMethodBodyStringResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodBodyStringResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyObjectEncodeCode = `// EncodeMethodBodyObjectResponse returns an encoder for responses returned by
// the ServiceBodyObject MethodBodyObject endpoint.
func EncodeMethodBodyObjectResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodBodyObjectResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyUserEncodeCode = `// EncodeMethodBodyUserResponse returns an encoder for responses returned by
// the ServiceBodyUser MethodBodyUser endpoint.
func EncodeMethodBodyUserResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*ResultType)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyArrayStringEncodeCode = `// EncodeMethodBodyArrayStringResponse returns an encoder for responses
// returned by the ServiceBodyArrayString MethodBodyArrayString endpoint.
func EncodeMethodBodyArrayStringResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodBodyArrayStringResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyArrayUserEncodeCode = `// EncodeMethodBodyArrayUserResponse returns an encoder for responses returned
// by the ServiceBodyArrayUser MethodBodyArrayUser endpoint.
func EncodeMethodBodyArrayUserResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodBodyArrayUserResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveStringEncodeCode = `// EncodeMethodBodyPrimitiveStringResponse returns an encoder for responses
// returned by the ServiceBodyPrimitiveString MethodBodyPrimitiveString
// endpoint.
func EncodeMethodBodyPrimitiveStringResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(string)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveBoolEncodeCode = `// EncodeMethodBodyPrimitiveBoolResponse returns an encoder for responses
// returned by the ServiceBodyPrimitiveBool MethodBodyPrimitiveBool endpoint.
func EncodeMethodBodyPrimitiveBoolResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(bool)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveArrayStringEncodeCode = `// EncodeMethodBodyPrimitiveArrayStringResponse returns an encoder for
// responses returned by the ServiceBodyPrimitiveArrayString
// MethodBodyPrimitiveArrayString endpoint.
func EncodeMethodBodyPrimitiveArrayStringResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.([]string)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyPrimitiveArrayBoolEncodeCode = `// EncodeMethodBodyPrimitiveArrayBoolResponse returns an encoder for responses
// returned by the ServiceBodyPrimitiveArrayBool MethodBodyPrimitiveArrayBool
// endpoint.
func EncodeMethodBodyPrimitiveArrayBoolResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.([]bool)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyHeaderObjectEncodeCode = `// EncodeMethodBodyHeaderObjectResponse returns an encoder for responses
// returned by the ServiceBodyHeaderObject MethodBodyHeaderObject endpoint.
func EncodeMethodBodyHeaderObjectResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodBodyHeaderObjectResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := &MethodBodyHeaderObjectResponseBody{
			A: res.A,
		}
		if res.B != nil {
			w.Header().Set("b", *res.B)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultBodyHeaderUserEncodeCode = `// EncodeMethodBodyHeaderUserResponse returns an encoder for responses returned
// by the ServiceBodyHeaderUser MethodBodyHeaderUser endpoint.
func EncodeMethodBodyHeaderUserResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*ResultType)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := &MethodBodyHeaderUserResponseBody{
			A: res.A,
		}
		if res.B != nil {
			w.Header().Set("b", *res.B)
		}
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultTagStringEncodeCode = `// EncodeMethodTagStringResponse returns an encoder for responses returned by
// the ServiceTagString MethodTagString endpoint.
func EncodeMethodTagStringResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodTagStringResult)
		if res.H != nil && *res.H == "value" {
			enc, ct := encoder(w, r)
			rest.SetContentType(w, ct)
			w.Header().Set("h", *res.H)
			w.WriteHeader(http.StatusAccepted)
			return nil
		}
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`

var ResultTagStringRequiredEncodeCode = `// EncodeMethodTagStringRequiredResponse returns an encoder for responses
// returned by the ServiceTagStringRequired MethodTagStringRequired endpoint.
func EncodeMethodTagStringRequiredResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*MethodTagStringRequiredResult)
		if res.H == "value" {
			enc, ct := encoder(w, r)
			rest.SetContentType(w, ct)
			w.Header().Set("h", res.H)
			w.WriteHeader(http.StatusAccepted)
			return nil
		}
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}
`
