{{ comment .Description }}
type {{ .VarName }} {{ .Def }}
{{- if .RawJSON }}
// MarshalJSON writes the stored JSON value without base64 encoding it.
func (v {{ .VarName }}) MarshalJSON() ([]byte, error) {
 return ({{ .Def }})(v).MarshalJSON()
}

// UnmarshalJSON copies the JSON value into the declared raw JSON representation.
func (v *{{ .VarName }}) UnmarshalJSON(data []byte) error {
 var raw {{ .Def }}
 if err := raw.UnmarshalJSON(data); err != nil {
  return err
 }
 *v = {{ .VarName }}(raw)
 return nil
}
{{- end }}
