{{- /* Union sum-type definition and helpers. */ -}}
{{- range .Fields }}
{{- if .EmitPrimitiveAlias }}
type {{ .FieldType }} {{ .PrimitiveAliasType }}

{{- end }}
{{- end }}
// {{ .Name }} is a sum-type union.
type {{ .Name }} struct {
	kind {{ .KindName }}
	{{- range .Fields }}
	{{ .FieldName }} {{ .FieldType }}
	{{- end }}
}

// {{ .KindName }} enumerates the union variants for {{ .Name }}.
type {{ .KindName }} string

const (
	{{- range .Fields }}
	// {{ .KindConst }} identifies the {{ .Name }} branch of the union.
	{{ .KindConst }} {{ $.KindName }} = "{{ .TypeTag }}"
	{{- end }}
)

// Kind returns the discriminator value of the union.
func (u {{ .Name }}) Kind() {{ .KindName }} {
	return u.kind
}

{{- range .Fields }}
// New{{ $.Name }}{{ .FieldName }} constructs a {{ $.Name }} with the {{ .Name }} branch set.
func New{{ $.Name }}{{ .FieldName }}(v {{ .FieldType }}) {{ $.Name }} {
	return {{ $.Name }}{
		kind:      {{ .KindConst }},
		{{ .FieldName }}: v,
	}
}

// As{{ .FieldName }} returns the value of the {{ .Name }} branch if set.
func (u {{ $.Name }}) As{{ .FieldName }}() (_ {{ .FieldType }}, ok bool) {
	if u.kind != {{ .KindConst }} {
		return
	}
	return u.{{ .FieldName }}, true
}

// Set{{ .FieldName }} sets the {{ .Name }} branch of the union.
func (u *{{ $.Name }}) Set{{ .FieldName }}(v {{ .FieldType }}) {
	u.kind = {{ .KindConst }}
	u.{{ .FieldName }} = v
}
{{- end }}

// Validate ensures the union discriminant is valid.
func (u {{ .Name }}) Validate() error {
	switch u.kind {
	case "":
		return goa.InvalidEnumValueError({{ printf "%q" .TypeKey }}, "", []any{
			{{- range .Fields }}
			string({{ .KindConst }}),
			{{- end }}
		})
	{{- range .Fields }}
	case {{ .KindConst }}:
		{{- if .Nilable }}
		if u.{{ .FieldName }} == nil {
			return goa.MissingFieldError({{ printf "%q" $.ValueKey }}, "{{ $.Name }}")
		}
		{{- end }}
		return nil
	{{- end }}
	default:
		return goa.InvalidEnumValueError({{ printf "%q" $.TypeKey }}, u.kind, []any{
			{{- range .Fields }}
			string({{ .KindConst }}),
			{{- end }}
		})
	}
}

// MarshalJSON marshals the union into the canonical {type,value} JSON shape.
func (u {{ .Name }}) MarshalJSON() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	var (
		value any
	)
	switch u.kind {
	{{- range .Fields }}
	case {{ .KindConst }}:
		value = u.{{ .FieldName }}
	{{- end }}
	default:
		return nil, fmt.Errorf("unexpected {{ .Name }} discriminant %q", u.kind)
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
func (u *{{ .Name }}) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  string          {{ printf "`json:\"%s\"`" .TypeKey }}
		Value json.RawMessage {{ printf "`json:\"%s\"`" .ValueKey }}
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Value) == 0 {
		return goa.MissingFieldError({{ printf "%q" .ValueKey }}, "{{ .Name }}")
	}
	if bytes.Equal(bytes.TrimSpace(raw.Value), []byte("null")) {
		return goa.InvalidFieldTypeError({{ printf "%q" .ValueKey }}, nil, "non-null JSON value")
	}
	switch raw.Type {
	{{- range .Fields }}
	case string({{ .KindConst }}):
		var v {{ .FieldType }}
		if err := json.Unmarshal(raw.Value, &v); err != nil {
			return err
		}
		u.kind = {{ .KindConst }}
		u.{{ .FieldName }} = v
	{{- end }}
	default:
		if raw.Type == "" {
			return goa.MissingFieldError({{ printf "%q" .TypeKey }}, "{{ .Name }}")
		}
		return goa.InvalidEnumValueError({{ printf "%q" .TypeKey }}, raw.Type, []any{
			{{- range .Fields }}
			string({{ .KindConst }}),
			{{- end }}
		})
	}
	return nil
}
