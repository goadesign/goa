{{ printf "%s returns an encoder to encode the multipart request for the %q service %q endpoint." .InitName .ServiceName .MethodName | comment }}
func {{ .InitName }}(encoderFn {{ .FuncName }}) func(r *http.Request) goahttp.Encoder {
	return func(r *http.Request) goahttp.Encoder {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		return goahttp.EncodingFunc(func(v any) error {
			p := v.({{ .Payload.Ref }})
			if err := encoderFn(mw, p); err != nil {
				return err
			}
			r.Body = io.NopCloser(body)
			r.Header.Set("Content-Type", mw.FormDataContentType())
			return mw.Close()
		})
	}
}
