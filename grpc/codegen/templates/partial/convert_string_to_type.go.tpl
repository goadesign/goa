{{- define "slice_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for i, rv := range {{ .VarName }}Raw {
		{{- template "slice_item_conversion" . }}
	}
{{- end }}

{{- define "slice_item_conversion" }}
	{{- if eq .Type.ElemType.Type.Name "string" }}
		{{ .VarName }}[i] = rv
	{{- else if eq .Type.ElemType.Type.Name "bytes" }}
		{{ .VarName }}[i] = []byte(rv)
	{{- else if eq .Type.ElemType.Type.Name "int" }}
		v, err2 := strconv.ParseInt(rv, 10, strconv.IntSize)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of integers"))
		}
		{{ .VarName }}[i] = int(v)
	{{- else if eq .Type.ElemType.Type.Name "int32" }}
		v, err2 := strconv.ParseInt(rv, 10, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of integers"))
		}
		{{ .VarName }}[i] = int32(v)
	{{- else if eq .Type.ElemType.Type.Name "int64" }}
		v, err2 := strconv.ParseInt(rv, 10, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of integers"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "uint" }}
		v, err2 := strconv.ParseUint(rv, 10, strconv.IntSize)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of unsigned integers"))
		}
		{{ .VarName }}[i] = uint(v)
	{{- else if eq .Type.ElemType.Type.Name "uint32" }}
		v, err2 := strconv.ParseUint(rv, 10, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of unsigned integers"))
		}
		{{ .VarName }}[i] = int32(v)
	{{- else if eq .Type.ElemType.Type.Name "uint64" }}
		v, err2 := strconv.ParseUint(rv, 10, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of unsigned integers"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "float32" }}
		v, err2 := strconv.ParseFloat(rv, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of floats"))
		}
		{{ .VarName }}[i] = float32(v)
	{{- else if eq .Type.ElemType.Type.Name "float64" }}
		v, err2 := strconv.ParseFloat(rv, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of floats"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "boolean" }}
		v, err2 := strconv.ParseBool(rv)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of booleans"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "any" }}
		{{ .VarName }}[i] = rv
	{{- else }}
		// unsupported slice type {{ .Type.ElemType.Type.Name }} for var {{ .VarName }}
	{{- end }}
{{- end }}

{{- define "type_conversion" }}
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
{{- end }}