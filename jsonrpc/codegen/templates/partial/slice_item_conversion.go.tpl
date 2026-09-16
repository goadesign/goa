		{{- if eq .ElemTypeName "string" }}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(rv)
		{{- else if eq .ElemTypeName "bytes" }}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}([]byte(rv))
		{{- else if eq .ElemTypeName "int" }}
			v, err2 := strconv.ParseInt(rv, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of integers"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "int32" }}
			v, err2 := strconv.ParseInt(rv, 10, 32)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of integers"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "int64" }}
			v, err2 := strconv.ParseInt(rv, 10, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of integers"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "uint" }}
			v, err2 := strconv.ParseUint(rv, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of unsigned integers"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "uint32" }}
			v, err2 := strconv.ParseUint(rv, 10, 32)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of unsigned integers"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "uint64" }}
			v, err2 := strconv.ParseUint(rv, 10, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of unsigned integers"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "float32" }}
			v, err2 := strconv.ParseFloat(rv, 32)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of floats"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "float64" }}
			v, err2 := strconv.ParseFloat(rv, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of floats"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "boolean" }}
			v, err2 := strconv.ParseBool(rv)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .Name }}, {{ .VarName}}Raw, "array of booleans"))
			}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(v)
		{{- else if eq .ElemTypeName "any" }}
			{{ .VarName }}[i] = {{ .ElemTypeRef }}(rv)
		{{- else }}
			// The Goa design must use primitive array elements for this HTTP response value.
		{{- end }}
