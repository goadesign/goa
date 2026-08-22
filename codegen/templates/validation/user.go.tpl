if err2 := {{ .name }}({{ .target }}); err2 != nil {
        err = goa.MergeErrors(err, err2)
}
{{- "" -}}
