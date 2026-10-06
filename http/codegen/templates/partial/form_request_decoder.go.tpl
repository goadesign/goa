		// Read only the body so query parameters cannot supply missing form fields.
		formBytes, readErr := io.ReadAll(r.Body)
		err = readErr
		if err == nil {
			if len(formBytes) == 0 {
				err = io.EOF
			} else {
				mediaType, _, mediaErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if mediaErr != nil || mediaType != "application/x-www-form-urlencoded" {
					return payload, goa.DecodePayloadError("expected application/x-www-form-urlencoded request body")
				}
				form, parseErr := url.ParseQuery(string(formBytes))
				if parseErr != nil {
					return payload, goa.DecodePayloadError("invalid URL-encoded request body")
				}
				{{- if .OptionalBody }}
				body = new({{ if .ServerBody.Declaration }}{{ .ServerBody.Declaration.Name }}{{ else }}{{ .ServerBody.VarName }}{{ end }})
				{{- end }}
				{{- range .ServerBody.FormFields }}
				if values, present := form[{{ printf "%q" .Name }}]; present {
					{{- if .Item }}
					body.{{ .FieldName }} = make({{ .TypeRef }}, len(values))
					for index, formItemRaw := range values {
						{{- template "partial_form_value_decoder" .Item }}
						body.{{ .FieldName }}[index] = {{ if .Item.Pointer }}&{{ end }}formItem
					}
					{{- else }}
					if len(values) != 1 {
						return payload, goa.DecodePayloadError({{ printf "%q" (printf "form field %s must occur once" .Name) }})
					}
					formValueRaw := values[0]
					{{- template "partial_form_value_decoder" .AttributeData }}
					body.{{ .FieldName }} = {{ if .Pointer }}&{{ end }}formValue
					{{- end }}
				}
				{{- end }}
			}
		}
