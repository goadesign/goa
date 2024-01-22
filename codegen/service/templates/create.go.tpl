{{ printf "%s initializes t from the fields of v" .Name | comment }}
func (t {{ .ReceiverTypeRef }}) {{ .Name }}(v {{ .TypeRef }}) {
	{{ .Code }}
	*t = *temp
}
