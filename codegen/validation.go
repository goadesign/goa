package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/expr"
)

var (
	enumValT     *template.Template
	formatValT   *template.Template
	patternValT  *template.Template
	minMaxValT   *template.Template
	lengthValT   *template.Template
	requiredValT *template.Template
	arrayValT    *template.Template
	mapValT      *template.Template
	userValT     *template.Template
)

func init() {
	fm := template.FuncMap{
		"slice":    toSlice,
		"oneof":    oneof,
		"constant": constant,
		"goifyAtt": GoifyAtt,
		"add":      func(a, b int) int { return a + b },
	}
	enumValT = template.Must(template.New("enum").Funcs(fm).Parse(enumValTmpl))
	formatValT = template.Must(template.New("format").Funcs(fm).Parse(formatValTmpl))
	patternValT = template.Must(template.New("pattern").Funcs(fm).Parse(patternValTmpl))
	minMaxValT = template.Must(template.New("minMax").Funcs(fm).Parse(minMaxValTmpl))
	lengthValT = template.Must(template.New("length").Funcs(fm).Parse(lengthValTmpl))
	requiredValT = template.Must(template.New("req").Funcs(fm).Parse(requiredValTmpl))
	arrayValT = template.Must(template.New("array").Funcs(fm).Parse(arrayValTmpl))
	mapValT = template.Must(template.New("map").Funcs(fm).Parse(mapValTmpl))
	userValT = template.Must(template.New("user").Funcs(fm).Parse(userValTmpl))
}

// ValidationCode produces Go code that runs the validations defined in the
// given attribute definition if any against the content of the variable named
// target. The generated code assumes that there is a pre-existing "err"
// variable of type error. It initializes that variable in case a validation
// fails.
//
// context is used to produce helpful messages in case of error.
//
func ValidationCode(ca *ContextualAttribute, target, context string) string {
	att := ca.Attribute.Expr()
	validation := att.Validation
	if validation == nil {
		return ""
	}
	var (
		kind            = att.Type.Kind()
		isNativePointer = kind == expr.BytesKind || kind == expr.AnyKind
		isPointer       = ca.IsPointer()
		tval            = target
	)
	if isPointer && expr.IsPrimitive(att.Type) && !isNativePointer {
		tval = "*" + tval
	}
	data := map[string]interface{}{
		"attribute": att,
		"isPointer": isPointer,
		"context":   context,
		"target":    target,
		"targetVal": tval,
		"string":    kind == expr.StringKind,
		"array":     expr.IsArray(att.Type),
		"map":       expr.IsMap(att.Type),
	}
	runTemplate := func(tmpl *template.Template, data interface{}) string {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		return buf.String()
	}
	var res []string
	if values := validation.Values; values != nil {
		data["values"] = values
		if val := runTemplate(enumValT, data); val != "" {
			res = append(res, val)
		}
	}
	if format := validation.Format; format != "" {
		data["format"] = string(format)
		if val := runTemplate(formatValT, data); val != "" {
			res = append(res, val)
		}
	}
	if pattern := validation.Pattern; pattern != "" {
		data["pattern"] = pattern
		if val := runTemplate(patternValT, data); val != "" {
			res = append(res, val)
		}
	}
	if min := validation.Minimum; min != nil {
		data["min"] = *min
		data["isMin"] = true
		delete(data, "max")
		if val := runTemplate(minMaxValT, data); val != "" {
			res = append(res, val)
		}
	}
	if max := validation.Maximum; max != nil {
		data["max"] = *max
		data["isMin"] = false
		delete(data, "min")
		if val := runTemplate(minMaxValT, data); val != "" {
			res = append(res, val)
		}
	}
	if minLength := validation.MinLength; minLength != nil {
		data["minLength"] = minLength
		data["isMinLength"] = true
		delete(data, "maxLength")
		if val := runTemplate(lengthValT, data); val != "" {
			res = append(res, val)
		}
	}
	if maxLength := validation.MaxLength; maxLength != nil {
		data["maxLength"] = maxLength
		data["isMinLength"] = false
		delete(data, "minLength")
		if val := runTemplate(lengthValT, data); val != "" {
			res = append(res, val)
		}
	}
	if req := validation.Required; len(req) > 0 {
		obj := expr.AsObject(att.Type)
		for _, r := range req {
			reqAtt := obj.Attribute(r)
			if reqAtt == nil {
				continue
			}
			if !ca.Pointer && expr.IsPrimitive(reqAtt.Type) &&
				reqAtt.Type.Kind() != expr.BytesKind &&
				reqAtt.Type.Kind() != expr.AnyKind {

				continue
			}
			data["req"] = r
			data["reqAtt"] = reqAtt
			res = append(res, runTemplate(requiredValT, data))
		}
	}
	return strings.Join(res, "\n")
}

// RecursiveValidationCode produces Go code that runs the validations defined in
// the given attribute and its children recursively against the value held by
// the variable named target.
func RecursiveValidationCode(a *ContextualAttribute, target string) string {
	seen := make(map[string]*bytes.Buffer)
	return recurseValidationCode(a, target, target, seen).String()
}

func recurseValidationCode(ca *ContextualAttribute, target, context string, seen map[string]*bytes.Buffer) *bytes.Buffer {
	var (
		buf   = new(bytes.Buffer)
		first = true
		att   = ca.Attribute.Expr()
	)

	// Break infinite recursions
	if ut, ok := att.Type.(expr.UserType); ok {
		if buf, ok := seen[ut.ID()]; ok {
			return buf
		}
		seen[ut.ID()] = buf
	}

	validation := ValidationCode(ca, target, context)
	if validation != "" {
		buf.WriteString(validation)
		first = false
	}

	runUserValT := func(name, target string) string {
		var buf bytes.Buffer
		data := map[string]interface{}{
			"name":   Goify(name, true),
			"target": target,
		}
		if err := userValT.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		return fmt.Sprintf("if %s != nil {\n\t%s\n}", target, buf.String())
	}

	if o := expr.AsObject(att.Type); o != nil {
		for _, nat := range *o {
			validation := recurseAttribute(ca, nat, target, context, seen)
			if validation != "" {
				if !first {
					buf.WriteByte('\n')
				} else {
					first = false
				}
				buf.WriteString(validation)
			}
		}
	} else if a := expr.AsArray(att.Type); a != nil {
		elem := ca.Dup(a.ElemType, true)
		elem.Pointer = false
		val := recurseValidationCode(elem, "e", context+"[*]", seen).String()
		if val != "" {
			switch a.ElemType.Type.(type) {
			case expr.UserType:
				// For user and result types, call the Validate method
				val = runUserValT(elem.Attribute.Name(), "e")
			}
			data := map[string]interface{}{
				"target":     target,
				"validation": val,
			}
			if !first {
				buf.WriteByte('\n')
			} else {
				first = false
			}
			if err := arrayValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	} else if m := expr.AsMap(att.Type); m != nil {
		key := ca.Dup(m.KeyType, true)
		key.Pointer = false
		keyVal := recurseValidationCode(key, "k", context+".key", seen).String()
		elem := ca.Dup(m.ElemType, true)
		elem.Pointer = false
		valueVal := recurseValidationCode(elem, "v", context+"[key]", seen).String()
		if keyVal != "" || valueVal != "" {
			if keyVal != "" {
				if _, ok := m.KeyType.Type.(expr.UserType); ok {
					keyVal = runUserValT(key.Attribute.Name(), "k")
				} else {
					keyVal = "\n" + keyVal
				}
			}
			if valueVal != "" {
				if _, ok := m.ElemType.Type.(expr.UserType); ok {
					valueVal = runUserValT(elem.Attribute.Name(), "v")
				} else {
					valueVal = "\n" + valueVal
				}
			}
			data := map[string]interface{}{
				"target":          target,
				"keyValidation":   keyVal,
				"valueValidation": valueVal,
			}
			if !first {
				buf.WriteByte('\n')
			} else {
				first = false
			}
			if err := mapValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	}
	return buf
}

func recurseAttribute(ca *ContextualAttribute, nat *expr.NamedAttributeExpr, target, context string, seen map[string]*bytes.Buffer) string {
	var (
		validation string

		att = ca.Attribute.Expr()
	)
	ca = ca.Dup(nat.Attribute, att.IsRequired(nat.Name))
	if ut, ok := nat.Attribute.Type.(expr.UserType); ok {
		// We need to check empirically whether there are validations to be
		// generated, we can't just generate and check whether something was
		// generated to avoid infinite recursions.
		hasValidations := false
		done := errors.New("done")
		Walk(ut.Attribute(), func(a *expr.AttributeExpr) error {
			if a.Validation != nil {
				if ca.Pointer {
					hasValidations = true
					return done
				}
				// For public data structures there is a case
				// where there is validation but no actual
				// validation code: if the validation is a
				// required validation that applies to
				// attributes that cannot be nil i.e. primitive
				// types.
				if !a.Validation.HasRequiredOnly() {
					hasValidations = true
					return done
				}
				obj := expr.AsObject(a.Type)
				for _, name := range a.Validation.Required {
					if att := obj.Attribute(name); att != nil && !expr.IsPrimitive(att.Type) {
						hasValidations = true
						return done
					}
				}
			}
			return nil
		})
		if hasValidations {
			var buf bytes.Buffer
			tgt := fmt.Sprintf("%s.%s", target, ca.Attribute.Field(nat.Name, true))
			if expr.IsArray(nat.Attribute.Type) {
				buf.Write(recurseValidationCode(ca, tgt, context, seen).Bytes())
			} else {
				if err := userValT.Execute(&buf, map[string]interface{}{"name": Goify(ut.Name(), true), "target": tgt}); err != nil {
					panic(err) // bug
				}
			}
			validation = buf.String()
		}
	} else {
		validation = recurseValidationCode(
			ca,
			fmt.Sprintf("%s.%s", target, ca.Attribute.Field(nat.Name, true)),
			fmt.Sprintf("%s.%s", context, nat.Name),
			seen,
		).String()
	}
	if validation != "" {
		if expr.IsObject(nat.Attribute.Type) {
			validation = fmt.Sprintf("if %s.%s != nil {\n%s\n}",
				target, ca.Attribute.Field(nat.Name, true), validation)
		}
	}
	return validation
}

// toSlice returns Go code that represents the given slice.
func toSlice(val []interface{}) string {
	elems := make([]string, len(val))
	for i, v := range val {
		elems[i] = fmt.Sprintf("%#v", v)
	}
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elems, ", "))
}

// oneof produces code that compares target with each element of vals and ORs
// the result, e.g. "target == 1 || target == 2".
func oneof(target string, vals []interface{}) string {
	elems := make([]string, len(vals))
	for i, v := range vals {
		elems[i] = fmt.Sprintf("%s == %#v", target, v)
	}
	return strings.Join(elems, " || ")
}

// constant returns the Go constant name of the format with the given value.
func constant(formatName string) string {
	switch formatName {
	case "date":
		return "goa.FormatDate"
	case "date-time":
		return "goa.FormatDateTime"
	case "uuid":
		return "goa.FormatUUID"
	case "email":
		return "goa.FormatEmail"
	case "hostname":
		return "goa.FormatHostname"
	case "ipv4":
		return "goa.FormatIPv4"
	case "ipv6":
		return "goa.FormatIPv6"
	case "ip":
		return "goa.FormatIP"
	case "uri":
		return "goa.FormatURI"
	case "mac":
		return "goa.FormatMAC"
	case "cidr":
		return "goa.FormatCIDR"
	case "regexp":
		return "goa.FormatRegexp"
	case "json":
		return "goa.FormatJSON"
	case "rfc1123":
		return "goa.FormatRFC1123"
	}
	panic("unknown format") // bug
}

const (
	arrayValTmpl = `for _, e := range {{ .target }} {
{{ .validation }}
}`

	mapValTmpl = `for {{if .keyValidation }}k{{ else }}_{{ end }}, {{ if .valueValidation }}v{{ else }}_{{ end }} := range {{ .target }} {
{{- .keyValidation }}
{{- .valueValidation }}
}`

	userValTmpl = `if err2 := Validate{{ .name }}({{ .target }}); err2 != nil {
        err = goa.MergeErrors(err, err2)
}`

	enumValTmpl = `{{ if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
if !({{ oneof .targetVal .values }}) {
        err = goa.MergeErrors(err, goa.InvalidEnumValueError({{ printf "%q" .context }}, {{ .targetVal }}, {{ slice .values }}))
{{ if .isPointer -}}
}
{{ end -}}
}`

	patternValTmpl = `{{ if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
        err = goa.MergeErrors(err, goa.ValidatePattern({{ printf "%q" .context }}, {{ .targetVal }}, {{ printf "%q" .pattern }}))
{{- if .isPointer }}
}
{{- end }}`

	formatValTmpl = `{{ if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
        err = goa.MergeErrors(err, goa.ValidateFormat({{ printf "%q" .context }}, {{ .targetVal}}, {{ constant .format }}))
{{ if .isPointer -}}
}
{{- end }}`

	minMaxValTmpl = `{{ if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
        if {{ .targetVal }} {{ if .isMin }}<{{ else }}>{{ end }} {{ if .isMin }}{{ .min }}{{ else }}{{ .max }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidRangeError({{ printf "%q" .context }}, {{ .targetVal }}, {{ if .isMin }}{{ .min }}, true{{ else }}{{ .max }}, false{{ end }}))
{{ if .isPointer -}}
}
{{ end -}}
}`

	lengthValTmpl = `{{ $target := or (and (or (or .array .map) .nonzero) .target) .targetVal -}}
{{ if and .isPointer .string -}}
if {{ .target }} != nil {
{{ end -}}
if {{ if .string }}utf8.RuneCountInString({{ $target }}){{ else }}len({{ $target }}){{ end }} {{ if .isMinLength }}<{{ else }}>{{ end }} {{ if .isMinLength }}{{ .minLength }}{{ else }}{{ .maxLength }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidLengthError({{ printf "%q" .context }}, {{ $target }}, {{ if .string }}utf8.RuneCountInString({{ $target }}){{ else }}len({{ $target }}){{ end }}, {{ if .isMinLength }}{{ .minLength }}, true{{ else }}{{ .maxLength }}, false{{ end }}))
}{{- if and .isPointer .string }}
}
{{- end }}`

	requiredValTmpl = `if {{ $.target }}.{{ goifyAtt $.reqAtt .req true }} == nil {
        err = goa.MergeErrors(err, goa.MissingFieldError("{{ .req }}", {{ printf "%q" $.context }}))
}`
)
