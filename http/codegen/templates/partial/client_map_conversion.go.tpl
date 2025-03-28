for k{{ if not (eq .Type.KeyType.Type.Name "string") }}Raw{{ end }}, value := range {{ .SourceVar }}{{ if .SourceField }}.{{ .SourceField }}{{ end }} {
		{{- if not (eq .Type.KeyType.Type.Name "string") }}
			{{ template "partial_client_type_conversion" (typeConversionData .Type.KeyType.Type .FieldType.KeyType.Type "k" "kRaw") }}
		{{- end }}
		key {{ if .NewVar }}:={{ else }}={{ end }} fmt.Sprintf("{{ .VarName }}[%s]", {{ if not .NewVar }}key, {{ end }}k)
		{{- if eq .Type.ElemType.Type.Name "string" }}
			values.Add(key, {{ if (isAlias .FieldType.ElemType.Type) }}string({{ end }}value{{ if (isAlias .FieldType.ElemType.Type) }}){{ end }})
		{{- else if eq .Type.ElemType.Type.Name "map" }}
			{{- template "partial_client_map_conversion" (mapConversionData .Type.ElemType.Type .FieldType.ElemType.Type "%s" "value" "" false) }}
		{{- else if eq .Type.ElemType.Type.Name "array" }}
			{{- if and (eq .Type.ElemType.Type.ElemType.Type.Name "string") (not (isAlias .FieldType.ElemType.Type.ElemType.Type)) }}
				values[key] = value
			{{- else }}
				for _, val := range value {
					{{ template "partial_client_type_conversion" (typeConversionData .Type.ElemType.Type.ElemType.Type (aliasedType .FieldType.ElemType.Type).ElemType.Type "valStr" "val") }}
					values.Add(key, valStr)
				}
			{{- end }}
		{{- else }}
			{{ template "partial_client_type_conversion" (typeConversionData .Type.ElemType.Type .FieldType.ElemType.Type "valueStr" "value") }}
			values.Add(key, valueStr)
		{{- end }}
	}
