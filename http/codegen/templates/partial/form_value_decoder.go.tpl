					if !utf8.ValidString({{ .VarName }}Raw) {
						return payload, goa.DecodePayloadError({{ printf "%q" (printf "form field %s must contain UTF-8 text" .Name) }})
					}
					{{- if eq .Type.Name "string" }}
					{{ .VarName }} := {{ .VarName }}Raw
					{{- else if eq .Type.Name "bytes" }}
					{{ .VarName }}, valueErr := base64.StdEncoding.DecodeString({{ .VarName }}Raw)
					if valueErr != nil {
						return payload, goa.DecodePayloadError({{ printf "%q" (printf "form field %s must contain base64" .Name) }})
					}
					{{- else }}
					var {{ .VarName }} {{ .TypeRef }}
					{{- template "partial_query_type_conversion" (conversionData .VarName .Name .Type) }}
					if err != nil {
						return payload, err
					}
					{{- end }}
