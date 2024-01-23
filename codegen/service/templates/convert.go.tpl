{{ printf "%s creates an instance of %s initialized from t." .Name .TypeName | comment }}
func (t {{ .ReceiverTypeRef }}) {{ .Name }}() {{ .TypeRef }} {
    {{ .Code }}
    return v
}
