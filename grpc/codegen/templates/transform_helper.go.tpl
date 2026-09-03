{{ printf "%s builds a value of type %s from a value of type %s." .Declaration.Name .ResultTypeRef .ParamTypeRef | comment }}
func {{ .Declaration.Name }}(v {{ .ParamTypeRef }}) {{ .ResultTypeRef }} {
  {{ .Code }}
  return res
}
