{{- /* Definition and helpers for a value that holds exactly one branch. */ -}}
{{- range .Fields }}
{{- if .EmitPrimitiveAlias }}
type {{ .FieldType }} {{ .PrimitiveAliasType }}

{{- end }}
{{- end }}
// {{ .TypeDeclaration.Name }} holds exactly one of its branch values.
type {{ .TypeDeclaration.Name }} struct {
	kind {{ .KindDeclaration.Name }}
	{{- range .Fields }}
	{{ .FieldName }} {{ .FieldType }}
	{{- end }}
}

// {{ .KindDeclaration.Name }} records which {{ .TypeDeclaration.Name }} branch is selected.
type {{ .KindDeclaration.Name }} string

const (
	{{- range .Fields }}
	// {{ .KindDeclaration.Name }} identifies the {{ .Name }} branch.
	{{ .KindDeclaration.Name }} {{ $.KindDeclaration.Name }} = "{{ .TypeTag }}"
	{{- end }}
)

// Kind returns the selected branch.
func (u {{ .TypeDeclaration.Name }}) Kind() {{ .KindDeclaration.Name }} {
	return u.kind
}

{{- range .Fields }}
// {{ .ConstructorDeclaration.Name }} constructs {{ $.TypeDeclaration.Name }} with the {{ .Name }} branch set.
func {{ .ConstructorDeclaration.Name }}(v {{ .FieldType }}) {{ $.TypeDeclaration.Name }} {
	return {{ $.TypeDeclaration.Name }}{
		kind:      {{ .KindDeclaration.Name }},
		{{ .FieldName }}: v,
	}
}

// As{{ .FieldName }} returns the value when the {{ .Name }} branch is selected.
func (u {{ $.TypeDeclaration.Name }}) As{{ .FieldName }}() (_ {{ .FieldType }}, ok bool) {
	if u.kind != {{ .KindDeclaration.Name }} {
		return
	}
	return u.{{ .FieldName }}, true
}

// Set{{ .FieldName }} selects the {{ .Name }} branch and stores v.
func (u *{{ $.TypeDeclaration.Name }}) Set{{ .FieldName }}(v {{ .FieldType }}) {
	u.kind = {{ .KindDeclaration.Name }}
	u.{{ .FieldName }} = v
}
{{- end }}

// Validate ensures exactly one valid branch is selected.
func (u {{ .TypeDeclaration.Name }}) Validate() error {
	switch u.kind {
	case "":
		return goa.InvalidEnumValueError({{ printf "%q" .TypeKey }}, "", []any{
			{{- range .Fields }}
			string({{ .KindDeclaration.Name }}),
			{{- end }}
		})
	{{- range .Fields }}
	case {{ .KindDeclaration.Name }}:
		{{- if .Nilable }}
		if u.{{ .FieldName }} == nil {
			return goa.MissingFieldError({{ printf "%q" $.ValueKey }}, "{{ $.TypeDeclaration.Name }}")
		}
		{{- end }}
		return nil
	{{- end }}
	default:
		return goa.InvalidEnumValueError({{ printf "%q" $.TypeKey }}, u.kind, []any{
			{{- range .Fields }}
			string({{ .KindDeclaration.Name }}),
			{{- end }}
		})
	}
}

// MarshalJSON marshals the union into the canonical {type,value} JSON shape.
func (u {{ .TypeDeclaration.Name }}) MarshalJSON() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	var (
		value any
	)
	switch u.kind {
	{{- range .Fields }}
	case {{ .KindDeclaration.Name }}:
		value = u.{{ .FieldName }}
	{{- end }}
	default:
		return nil, fmt.Errorf("unexpected {{ .TypeDeclaration.Name }} kind %q", u.kind)
	}
	return json.Marshal(struct {
		Type  string {{ printf "`json:\"%s\"`" .TypeKey }}
		Value any    {{ printf "`json:\"%s\"`" .ValueKey }}
	}{
		Type:  string(u.kind),
		Value: value,
	})
}

// UnmarshalJSON unmarshals the union from the canonical {type,value} JSON shape.
func (u *{{ .TypeDeclaration.Name }}) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  string          {{ printf "`json:\"%s\"`" .TypeKey }}
		Value json.RawMessage {{ printf "`json:\"%s\"`" .ValueKey }}
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Value) == 0 {
		return goa.MissingFieldError({{ printf "%q" .ValueKey }}, "{{ .TypeDeclaration.Name }}")
	}
	if bytes.Equal(bytes.TrimSpace(raw.Value), []byte("null")) {
		return goa.InvalidFieldTypeError({{ printf "%q" .ValueKey }}, nil, "non-null JSON value")
	}
	switch raw.Type {
	{{- range .Fields }}
	case string({{ .KindDeclaration.Name }}):
		var v {{ .FieldType }}
		if err := json.Unmarshal(raw.Value, &v); err != nil {
			return err
		}
		u.kind = {{ .KindDeclaration.Name }}
		u.{{ .FieldName }} = v
	{{- end }}
	default:
		if raw.Type == "" {
			return goa.MissingFieldError({{ printf "%q" .TypeKey }}, "{{ .TypeDeclaration.Name }}")
		}
		return goa.InvalidEnumValueError({{ printf "%q" .TypeKey }}, raw.Type, []any{
			{{- range .Fields }}
			string({{ .KindDeclaration.Name }}),
			{{- end }}
		})
	}
	return nil
}
