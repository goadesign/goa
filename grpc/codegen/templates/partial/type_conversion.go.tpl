{{- if eq .Type.Name "bytes" }}
	{{ .VarName }} = []byte({{ .VarName }}Raw)
{{- else if eq .Type.Name "int" }}
	v, err2 := strconv.ParseInt({{ .VarName }}Raw, 10, strconv.IntSize)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "integer"))
	}
	{{- if .Pointer }}
	pv := int(v)
	{{ .VarName }} = &pv
	{{- else }}
	{{ .VarName }} = int(v)
	{{- end }}
{{- else if eq .Type.Name "int32" }}
	v, err2 := strconv.ParseInt({{ .VarName }}Raw, 10, 32)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "integer"))
	}
	{{- if .Pointer }}
	pv := int32(v)
	{{ .VarName }} = &pv
	{{- else }}
	{{ .VarName }} = int32(v)
	{{- end }}
{{- else if eq .Type.Name "int64" }}
	v, err2 := strconv.ParseInt({{ .VarName }}Raw, 10, 64)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "integer"))
	}
	{{ .VarName }} = {{ if .Pointer}}&{{ end }}v
{{- else if eq .Type.Name "uint" }}
	v, err2 := strconv.ParseUint({{ .VarName }}Raw, 10, strconv.IntSize)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "unsigned integer"))
	}
	{{- if .Pointer }}
	pv := uint(v)
	{{ .VarName }} = &pv
	{{- else }}
	{{ .VarName }} = uint(v)
	{{- end }}
{{- else if eq .Type.Name "uint32" }}
	v, err2 := strconv.ParseUint({{ .VarName }}Raw, 10, 32)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "unsigned integer"))
	}
	{{- if .Pointer }}
	pv := uint32(v)
	{{ .VarName }} = &pv
	{{- else }}
	{{ .VarName }} = uint32(v)
	{{- end }}
{{- else if eq .Type.Name "uint64" }}
	v, err2 := strconv.ParseUint({{ .VarName }}Raw, 10, 64)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "unsigned integer"))
	}
	{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
{{- else if eq .Type.Name "float32" }}
	v, err2 := strconv.ParseFloat({{ .VarName }}Raw, 32)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "float"))
	}
	{{- if .Pointer }}
	pv := float32(v)
	{{ .VarName }} = &pv
	{{- else }}
	{{ .VarName }} = float32(v)
	{{- end }}
{{- else if eq .Type.Name "float64" }}
	v, err2 := strconv.ParseFloat({{ .VarName }}Raw, 64)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "float"))
	}
	{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
{{- else if eq .Type.Name "boolean" }}
	v, err2 := strconv.ParseBool({{ .VarName }}Raw)
	if err2 != nil {
		err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "boolean"))
	}
	{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
{{- else }}
	// unsupported type {{ .Type.Name }} for var {{ .VarName }}
{{- end }}
