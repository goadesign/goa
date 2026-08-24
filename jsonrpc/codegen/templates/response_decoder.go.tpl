{{ printf "%s returns a decoder for responses returned by the %s service %s JSON-RPC method. restoreBody controls whether the response body should be restored after having been read." .ResponseDecoderDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .ResponseDecoderDeclaration.Name }}(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) (result any, decodeErr error) {
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

		if jresp.Error != nil {
			switch jresp.Error.Code {
{{- range .Errors }}
	{{- range .Errors }}
		{{- with .Response }}
			case {{ .StatusCode }}:
				resp.Body = io.NopCloser(bytes.NewBuffer(jresp.Error.Data))
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
{{- end }}
			default:
				return nil, goahttp.ErrInvalidResponse({{ printf "%q" .ServiceName }}, {{ printf "%q" .Method.Name }}, resp.StatusCode, string(jresp.Error.Data))
			}
		}

	{{- if .Method.ViewedResult }}
		return {{ viewedDecodeName .Method.Name }}(decoder, resp, jresp.Result)
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
