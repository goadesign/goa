{{ printf "%s returns a decoder for responses returned by the %s service %s JSON-RPC method. The decoder rejects responses that do not repeat requestID. restoreBody controls whether the response body should be restored after having been read." .ResponseDecoderDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .ResponseDecoderDeclaration.Name }}(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response, string) (any, error) {
	return func(resp *http.Response, requestID string) (result any, decodeErr error) {
		responseBody := resp.Body
		if restoreBody {
			b, readErr := io.ReadAll(responseBody)
			closeErr := responseBody.Close()
			if err := errors.Join(readErr, closeErr); err != nil {
				return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer func() {
				if err := responseBody.Close(); err != nil {
					decodeErr = errors.Join(decodeErr, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err))
				}
			}()
		}

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
			}
			return nil, goahttp.ErrInvalidResponse("{{ .ServiceName }}", "{{ .Method.Name }}", resp.StatusCode, string(body))
		}

		var jresp jsonrpc.RawResponse
		if err := decoder(resp).Decode(&jresp); err != nil {
			return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		if err := jresp.Validate(requestID); err != nil {
			return nil, goahttp.ErrInvalidResponse("{{ .ServiceName }}", "{{ .Method.Name }}", resp.StatusCode, err.Error())
		}

		if jresp.Error != nil {
		{{- if .Errors }}
			serviceErrorName, serviceErrorBody, ok := jsonrpc.DecodeServiceErrorData(jresp.Error.Data)
			if !ok {
				return nil, jresp.Error
			}
			switch jresp.Error.Code {
{{- range .Errors }}
			case {{ .StatusCode }}:
				switch serviceErrorName {
	{{- range $mapped := .Errors }}
		{{- with $mapped.Response }}
				case {{ printf "%q" $mapped.Name }}:
					{{- if or (eq .Code -32700) (eq .Code -32600) }}
					if jresp.ID == nil {
						return nil, goahttp.ErrInvalidResponse("{{ $.ServiceName }}", "{{ $.Method.Name }}", resp.StatusCode, "response id is null")
					}
					{{- end }}
					resp.Body = io.NopCloser(bytes.NewReader(serviceErrorBody))
			{{- if not .ClientBody }}
					if !bytes.Equal(bytes.TrimSpace(serviceErrorBody), []byte("null")) {
						return nil, jresp.Error
					}
			{{- end }}
				{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
			{{- if .ResultInit }}
					return nil, {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
			{{- else if .ClientBody }}
					return nil, body
			{{- else }}
					return nil, nil
			{{- end }}
		{{- end }}
	{{- end }}
				default:
					return nil, jresp.Error
				}
{{- end }}
			default:
				return nil, jresp.Error
			}
		{{- else }}
			return nil, jresp.Error
		{{- end }}
		}

	{{- if .Method.ViewedResult }}
		return {{ viewedDecodeName .Method.Name }}(decoder, resp, jresp.Result{{ if .SSE }}{{ if .SSE.IDField }}, "", false{{ end }}{{ if .SSE.EventField }}, "", false{{ end }}{{ if .SSE.RetryField }}, "", false{{ end }}{{ end }})
	{{- else }}
{{- with index .Result.Responses 0 }}
		resp.Body = io.NopCloser(bytes.NewBuffer(jresp.Result))
		{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
{{- if .ResultInit }}
		res := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
		return res, nil
{{- else if .ClientBody }}
		return body, nil
{{- else if .Headers }}
		return {{ (index .Headers 0).VarName }}, nil
{{- else if .Cookies }}
		return {{ (index .Cookies 0).VarName }}, nil
{{- else }}
		return nil, nil
{{- end }}
{{- end }}
	{{- end }}
	}
}

{{- define "viewed_sse_outer_fields" }}{{ end }}
