if err2 := {{ .call }}; err2 != nil {
		err = {{ .goa }}.MergeErrors(err, err2)
}
{{- "" -}}
