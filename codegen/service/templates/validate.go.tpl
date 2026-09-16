{{ comment .Description }}
func {{ .Declaration.Name }}(result {{ .Ref }}) (err error) {
	{{ .Validate }}
  return
}
