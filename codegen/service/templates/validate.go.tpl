{{ comment .Description }}
func {{ .Name }}(result {{ .Ref }}) (err error) {
	{{ .Validate }}
  return
}
