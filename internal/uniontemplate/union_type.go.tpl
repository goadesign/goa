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
	{{ .StorageName }} {{ .FieldType }}
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
		kind: {{ .KindDeclaration.Name }},
		{{ .StorageName }}: v,
	}
}

// As{{ .FieldName }} returns the value when the {{ .Name }} branch is selected.
func (u {{ $.TypeDeclaration.Name }}) As{{ .FieldName }}() (_ {{ .FieldType }}, ok bool) {
	if u.kind != {{ .KindDeclaration.Name }} {
		return
	}
	return u.{{ .StorageName }}, true
}

// Set{{ .FieldName }} selects the {{ .Name }} branch and stores v.
func (u *{{ $.TypeDeclaration.Name }}) Set{{ .FieldName }}(v {{ .FieldType }}) {
	*u = {{ $.TypeDeclaration.Name }}{
		kind: {{ .KindDeclaration.Name }},
		{{ .StorageName }}: v,
	}
}
{{- end }}

// Validate ensures exactly one valid branch is selected.
func (u {{ .TypeDeclaration.Name }}) Validate() error {
	_, err := u.Value()
	return err
}

// Value returns the selected branch value, or the same selection error as Validate.
// Go templates can read this method directly; an invalid selection stops execution.
func (u {{ .TypeDeclaration.Name }}) Value() (any, error) {
	switch u.kind {
	case "":
		return nil, goa.InvalidEnumValueError({{ printf "%q" .TypeKey }}, "", []any{
			{{- range .Fields }}
			string({{ .KindDeclaration.Name }}),
			{{- end }}
		})
	{{- range .Fields }}
	case {{ .KindDeclaration.Name }}:
		{{- if .Nilable }}
		if u.{{ .StorageName }} == nil {
			return nil, goa.MissingFieldError({{ printf "%q" $.ValueKey }}, "{{ $.TypeDeclaration.Name }}")
		}
		{{- end }}
		return u.{{ .StorageName }}, nil
	{{- end }}
	default:
		return nil, goa.InvalidEnumValueError({{ printf "%q" $.TypeKey }}, u.kind, []any{
			{{- range .Fields }}
			string({{ .KindDeclaration.Name }}),
			{{- end }}
		})
	}
}

{{- if .Untagged }}
// MarshalJSON writes the selected branch value without a discriminator or envelope.
{{- else if .Flatten }}
// MarshalJSON writes the selected object branch beside its discriminator.
{{- else }}
// MarshalJSON marshals the union into the canonical {type,value} JSON shape.
{{- end }}
func (u {{ .TypeDeclaration.Name }}) MarshalJSON() ([]byte, error) {
	value, err := u.Value()
	if err != nil {
		return nil, err
	}
	{{- if .Untagged }}
	return json.Marshal(value)
	{{- else if .Flatten }}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, goa.InvalidFieldTypeError({{ printf "%q" .TypeKey }}, nil, "non-null JSON object")
	}
	if _, exists := fields[{{ printf "%q" .TypeKey }}]; exists {
		return nil, fmt.Errorf("{{ .TypeDeclaration.Name }} branch already contains discriminator %q", {{ printf "%q" .TypeKey }})
	}
	tag, err := json.Marshal(string(u.kind))
	if err != nil {
		return nil, err
	}
	fields[{{ printf "%q" .TypeKey }}] = tag
	return json.Marshal(fields)
	{{- else }}
	return json.Marshal(struct {
		Type  string {{ printf "`json:\"%s\"`" .TypeKey }}
		Value any    {{ printf "`json:\"%s\"`" .ValueKey }}
	}{
		Type:  string(u.kind),
		Value: value,
	})
	{{- end }}
}

{{- if .Untagged }}
// UnmarshalJSON selects the branch by JSON kind and decodes its typed value.
{{- else if .Flatten }}
// UnmarshalJSON reads the discriminator and decodes its declared object branch.
{{- else }}
// UnmarshalJSON unmarshals the union from the canonical {type,value} JSON shape.
{{- end }}
func (u *{{ .TypeDeclaration.Name }}) UnmarshalJSON(data []byte) error {
	{{- if .Untagged }}
	data = bytes.Trim(data, " \t\r\n")
	if len(data) == 0 {
		return fmt.Errorf("{{ .TypeDeclaration.Name }} requires a non-null JSON value")
	}
	switch data[0] {
	{{- range .Fields }}
	{{- if eq .JSONKind 48 }}
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
	{{- else if eq .JSONKind 116 }}
	case 't', 'f':
	{{- else }}
	case {{ printf "%q" .JSONKind }}:
	{{- end }}
		var value {{ .FieldType }}
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		selected := {{ .ConstructorDeclaration.Name }}(value)
		if err := selected.Validate(); err != nil {
			return err
		}
		*u = selected
		return nil
	{{- end }}
	default:
		return fmt.Errorf("{{ .TypeDeclaration.Name }} has no branch for this JSON value")
	}
	{{- else }}
	{{- if .Flatten }}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	tag, exists := fields[{{ printf "%q" .TypeKey }}]
	if !exists {
		return goa.MissingFieldError({{ printf "%q" .TypeKey }}, "{{ .TypeDeclaration.Name }}")
	}
	if bytes.Equal(bytes.TrimSpace(tag), []byte("null")) {
		return goa.InvalidFieldTypeError({{ printf "%q" .TypeKey }}, nil, "JSON string")
	}
	var raw struct {
		Type string
		Value json.RawMessage
	}
	if err := json.Unmarshal(tag, &raw.Type); err != nil {
		return err
	}
	delete(fields, {{ printf "%q" .TypeKey }})
	value, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	raw.Value = value
	{{- else }}
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
	{{- end }}
	switch raw.Type {
	{{- range .Fields }}
	case string({{ .KindDeclaration.Name }}):
		var v {{ .FieldType }}
		if err := json.Unmarshal(raw.Value, &v); err != nil {
			return err
		}
		u.Set{{ .FieldName }}(v)
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
	{{- end }}
}
