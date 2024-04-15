{{- if or .PathParams .QueryParams .Headers .Cookies }}
{{- if .ServerBody }}{{/* we want a newline only if there was code before */}}
{{ end }}
		var (
		{{- range .PathParams }}
			{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- range .QueryParams }}
			{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- range .Headers }}
			{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- range .Cookies }}
			{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- if and .MustValidate (or (not .ServerBody) .Multipart) }}
			err error
		{{- end }}
		{{- if .Cookies }}
			c *http.Cookie
		{{- end }}
		{{- if .PathParams }}

			params = mux.Vars(r)
		{{- end }}
		)

{{- range .PathParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }} = params["{{ .HTTPName }}"]

	{{- else }}{{/* not string and not any */}}
		{
			{{ .VarName }}Raw := params["{{ .HTTPName }}"]
			{{- template "partial_path_conversion" . }}
		}

	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}

{{- $qpVar := "r.URL.Query()" }}
{{- if gt (len .QueryParams) 1 }}
{{- $qpVar = "qp" }}
qp := r.URL.Query()
{{- end }}
{{- range .QueryParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = {{$qpVar}}.Get("{{ .HTTPName }}")
		if {{ .VarName }} == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := {{$qpVar}}.Get("{{ .HTTPName }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}{{ printf "%q" .DefaultValue }}{{ else }}{{ printf "%#v" .DefaultValue }}{{ end }}
		}
		{{- end }}

	{{- else if .StringSlice }}
		{{ .VarName }} = {{$qpVar}}["{{ .HTTPName }}"]
		{{- if .Required }}
		if {{ .VarName }} == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }} == nil {
			{{ .VarName }} = []string{
                {{- range $i, $v := .DefaultValue }}
                    {{- if $i }}{{ print ", " }}{{ end }}
                    {{- printf "%q" $v -}}
                {{- end -}} }
		}
		{{- end }}

	{{- else if .Slice }}
	{
		{{ .VarName }}Raw := {{$qpVar}}["{{ .HTTPName }}"]
		{{- if .Required }}
		if {{ .VarName }}Raw == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != nil {
		{{- end }}
		{{- template "partial_element_slice_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	}

	{{- else if .Map }}
	{
		{{ .VarName }}Raw := {{$qpVar}}
		{{- if .Required }}
		if len({{ .VarName }}Raw) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- else if .DefaultValue }}
		if len({{ .VarName }}Raw) == 0 {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if len({{ .VarName }}Raw) != 0 {
		{{- end }}
		for keyRaw, valRaw := range {{ .VarName }}Raw {
			if strings.HasPrefix(keyRaw, "{{ .HTTPName }}[") {
				{{- template "partial_query_map_conversion" (mapQueryDecodeData .Type .VarName 0) }}
			}
		}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	}

	{{- else if .MapQueryParams }}
	{
		{{ .VarName }}Raw := {{$qpVar}}
		{{- if .Required }}
		if len({{ .VarName }}Raw) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- else if .DefaultValue }}
		if len({{ .VarName }}Raw) == 0 {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if len({{ .VarName }}Raw) != 0 {
		{{- end }}
		if {{ .VarName }} == nil {
			{{ .VarName }} = make({{ goTypeRef .Type }})
		}
		for keyRaw, valRaw := range {{ .VarName }}Raw {
			var key {{ goTypeRef .Type.KeyType.Type }}
			{{- if eq .Type.KeyType.Type.Name "string" }}
			key = keyRaw
			{{- else }}
				{{- template "partial_query_type_conversion" (conversionData "key" "query" .Type.KeyType.Type) }}
			{{- end }}
			{{- if eq .Type.ElemType.Type.Name "string" }}
				{{ .VarName }}[key] = valRaw[0]
			{{- else if eq .Type.ElemType.Type.Name "array" }}
				{{- if eq .Type.ElemType.Type.ElemType.Type.Name "string" }}
					{{ .VarName }}[key] = valRaw
				{{- else }}
					var val {{ goTypeRef .Type.ElemType.Type }}
					{
						{{- template "partial_element_slice_conversion" (conversionData "val" "query" .Type.ElemType.Type) }}
					}
					{{ .VarName }}[key] = val
				{{- end }}
			{{- else if eq .Type.ElemType.Type.Name "map" }}
				{{- template "partial_query_map_conversion" (mapQueryDecodeData .Type.ElemType.Type (printf "%s[key]" .VarName) 1) }}
			{{- else }}
				var val{{ .Loop }} {{ goTypeRef .Type.ElemType.Type }}
				{
					val{{ .Loop }}Raw := valRaw[0]
					{{- template "partial_query_type_conversion" (conversionData (printf "val%s" .Loop) "query" .Type.ElemType.Type) }}
				}
				{{ .VarName }}[key] = val{{ .Loop }}
			{{- end }}
		}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	}

	{{- else }}{{/* not string, not any, not slice and not map */}}
	{
		{{ .VarName }}Raw := {{$qpVar}}.Get("{{ .HTTPName }}")
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == "" {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != "" {
		{{- end }}
		{{- template "partial_query_type_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	}

	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}

{{- range .Headers }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = r.Header.Get("{{ .HTTPName }}")
		if {{ .VarName }} == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := r.Header.Get("{{ .HTTPName }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}{{ printf "%q" .DefaultValue }}{{ else }}{{ printf "%#v" .DefaultValue }}{{ end }}
		}
		{{- end }}

	{{- else if .StringSlice }}
		{{ .VarName }} = r.Header["{{ .CanonicalName }}"]
		{{- if .Required }}
		if {{ .VarName }} == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }} == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

	{{- else if .Slice }}
	{
		{{ .VarName }}Raw := r.Header["{{ .CanonicalName }}"]
		{{ if .Required }}if {{ .VarName }}Raw == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != nil {
		{{- end }}
		{{- template "partial_element_slice_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	}

	{{- else }}{{/* not string, not any and not slice */}}
	{
		{{ .VarName }}Raw := r.Header.Get("{{ .HTTPName }}")
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == "" {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != "" {
		{{- end }}
		{{- template "partial_query_type_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	}
	{{- end }}
	{{- if .Validate }}
		{{ .Validate }}
	{{- end }}
{{- end }}

{{- range .Cookies }}
	c, {{ if not .Required }}_{{ else }}err{{ end }} = r.Cookie("{{ .HTTPName }}")
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		if err == http.ErrNoCookie {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "cookie"))
		} else {
			{{ .VarName }} = c.Value
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		var {{ .VarName }}Raw string
		if c != nil {
			{{ .VarName }}Raw = c.Value
		}
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}{{ printf "%q" .DefaultValue }}{{ else }}{{ printf "%#v" .DefaultValue }}{{ end }}
		}
		{{- end }}

	{{- else }}{{/* not string and not any */}}
	{
		var {{ .VarName }}Raw string
		if c != nil {
			{{ .VarName }}Raw = c.Value
		}
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "cookie"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == "" {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != "" {
		{{- end }}
		{{- template "partial_query_type_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	}
	{{- end }}
	{{- if .Validate }}
		{{ .Validate }}
	{{- end }}
{{- end }}
{{- end -}}
