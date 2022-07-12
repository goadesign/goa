package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/v3/expr"
)

var (
	enumValT       *template.Template
	formatValT     *template.Template
	patternValT    *template.Template
	exclMinMaxValT *template.Template
	minMaxValT     *template.Template
	lengthValT     *template.Template
	requiredValT   *template.Template
	arrayValT      *template.Template
	mapValT        *template.Template
	userValT       *template.Template
)

func init() {
	fm := template.FuncMap{
		"slice":    toSlice,
		"oneof":    oneof,
		"constant": constant,
		"add":      func(a, b int) int { return a + b },
		"isset":    func(i interface{}) bool { return i != nil },
	}
	enumValT = template.Must(template.New("enum").Funcs(fm).Parse(enumValTmpl))
	formatValT = template.Must(template.New("format").Funcs(fm).Parse(formatValTmpl))
	patternValT = template.Must(template.New("pattern").Funcs(fm).Parse(patternValTmpl))
	exclMinMaxValT = template.Must(template.New("exclMinMax").Funcs(fm).Parse(exclMinMaxValTmpl))
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
// attCtx is the attribute context
//
// req indicates whether the attribute is required (true) or optional (false)
//
// alias indicates whether the attribute is an alias user type attribute.
//
// target is the variable name against which the validation code is generated
//
// context is used to produce helpful messages in case of error.
//
func ValidationCode(att *expr.AttributeExpr, attCtx *AttributeContext, req, alias bool, target, context string) string {
	validation := att.Validation
	ut, isut := att.Type.(expr.UserType)
	if isut && validation == nil {
		validation = ut.Attribute().Validation
	}
	if validation == nil {
		return ""
	}
	var (
		kind            = att.Type.Kind()
		isNativePointer = kind == expr.BytesKind || kind == expr.AnyKind
		isPointer       = attCtx.Pointer || !attCtx.IgnoreRequired && (!req && (att.DefaultValue == nil || !attCtx.UseDefault))
		tval            = target
	)
	if isPointer && expr.IsPrimitive(att.Type) && !isNativePointer {
		tval = "*" + tval
	}
	if alias {
		tval = fmt.Sprintf("%s(%s)", att.Type.Name(), tval)
	}
	data := map[string]interface{}{
		"attribute": att,
		"attCtx":    attCtx,
		"isPointer": isPointer,
		"context":   context,
		"target":    target,
		"targetVal": tval,
		"string":    kind == expr.StringKind,
		"array":     expr.IsArray(att.Type),
		"map":       expr.IsMap(att.Type),
		"zeroVal":   att.ZeroValue,
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
	if exclMin := validation.ExclusiveMinimum; exclMin != nil {
		data["exclMin"] = *exclMin
		data["isExclMin"] = true
		delete(data, "exclMax")
		if val := runTemplate(exclMinMaxValT, data); val != "" {
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
	if exclMax := validation.ExclusiveMaximum; exclMax != nil {
		data["exclMax"] = *exclMax
		data["isExclMax"] = true
		delete(data, "exclMax")
		if val := runTemplate(exclMinMaxValT, data); val != "" {
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
			if !attCtx.Pointer && expr.IsPrimitive(reqAtt.Type) &&
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
// the variable named target. See ValidationCode for a description of the
// arguments and their effects.
//
// attCtx is the Attributor for the given attribute which is used to generate
// attribute name and reference in the validation code.
//
func RecursiveValidationCode(att *expr.AttributeExpr, attCtx *AttributeContext, req, alias bool, target string) string {
	seen := make(map[string]*bytes.Buffer)
	return recurseValidationCode(att, attCtx, req, alias, target, target, seen).String()
}

func recurseValidationCode(att *expr.AttributeExpr, attCtx *AttributeContext, req, alias bool, target, context string, seen map[string]*bytes.Buffer) *bytes.Buffer {
	var (
		buf   = new(bytes.Buffer)
		first = true
	)

	// Break infinite recursions
	if ut, ok := att.Type.(expr.UserType); ok {
		if buf, ok := seen[ut.ID()]; ok {
			return buf
		}
		seen[ut.ID()] = buf
	}

	code := func(ctx *AttributeContext, att *expr.AttributeExpr, tgt, suf string) string {
		ut, ok := att.Type.(expr.UserType)
		if !ok {
			return recurseValidationCode(att, ctx, true, false, tgt, context+suf, seen).String()
		}
		if expr.IsAlias(ut) {
			return recurseValidationCode(ut.Attribute(), ctx, true, true, tgt, context+suf, seen).String()
		}
		if !hasValidations(attCtx, ut) {
			return ""
		}
		var buf bytes.Buffer
		name := attCtx.Scope.Name(att, ctx.Pkg(att), ctx.Pointer, ctx.UseDefault)
		data := map[string]interface{}{"name": Goify(name, true), "target": tgt}
		if err := userValT.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		return fmt.Sprintf("if %s != nil {\n\t%s\n}", tgt, buf.String())
	}

	newline := func() {
		if !first {
			buf.WriteByte('\n')
		} else {
			first = false
		}
	}

	// Write validations on attribute if any.
	validation := ValidationCode(att, attCtx, req, alias, target, context)
	if validation != "" {
		buf.WriteString(validation)
		first = false
	}

	// Recurse down depending on attribute type.
	switch {
	case expr.IsObject(att.Type):
		for _, nat := range *(expr.AsObject(att.Type)) {
			validation := recurseAttribute(att, attCtx, nat, target, context, seen)
			if validation != "" {
				newline()
				buf.WriteString(validation)
			}
		}
	case expr.IsArray(att.Type):
		elem := expr.AsArray(att.Type).ElemType
		ctx := attCtx
		if ctx.Pointer && expr.IsPrimitive(elem.Type) {
			// Array elements of primtive type are never pointers
			ctx = attCtx.Dup()
			ctx.Pointer = false
		}
		val := code(ctx, elem, "e", "[*]")
		if val != "" {
			newline()
			data := map[string]interface{}{"target": target, "validation": val}
			if err := arrayValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	case expr.IsMap(att.Type):
		m := expr.AsMap(att.Type)
		ctx := attCtx.Dup()
		ctx.Pointer = false
		keyVal := code(ctx, m.KeyType, "k", ".key")
		if keyVal != "" {
			keyVal = "\n" + keyVal
		}
		valueVal := code(ctx, m.ElemType, "v", "[key]")
		if valueVal != "" {
			valueVal = "\n" + valueVal
		}
		if keyVal != "" || valueVal != "" {
			newline()
			data := map[string]interface{}{"target": target, "keyValidation": keyVal, "valueValidation": valueVal}
			if err := mapValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	}

	return buf
}

func recurseAttribute(att *expr.AttributeExpr, attCtx *AttributeContext, nat *expr.NamedAttributeExpr, target, context string, seen map[string]*bytes.Buffer) string {
	var validation string
	ut, isut := nat.Attribute.Type.(expr.UserType)
	if !isut {
		target := fmt.Sprintf("%s.%s", target, attCtx.Scope.Field(nat.Attribute, nat.Name, true))
		context := fmt.Sprintf("%s.%s", context, nat.Name)
		code := recurseValidationCode(nat.Attribute, attCtx, att.IsRequired(nat.Name), false, target, context, seen)
		validation = code.String()
	} else if hasValidations(attCtx, ut) {
		var buf bytes.Buffer
		tgt := fmt.Sprintf("%s.%s", target, attCtx.Scope.Field(nat.Attribute, nat.Name, true))
		if expr.IsPrimitive(nat.Attribute.Type) {
			buf.Write(recurseValidationCode(ut.Attribute(), attCtx, att.IsRequired(nat.Name), true, tgt, context, seen).Bytes())
		} else {
			name := attCtx.Scope.Name(nat.Attribute, attCtx.Pkg(nat.Attribute), attCtx.Pointer, attCtx.UseDefault)
			if err := userValT.Execute(&buf, map[string]interface{}{"name": Goify(name, true), "target": tgt}); err != nil {
				panic(err) // bug
			}
		}
		validation = buf.String()
	}
	if validation != "" {
		if expr.IsObject(nat.Attribute.Type) {
			validation = fmt.Sprintf("if %s.%s != nil {\n%s\n}",
				target, attCtx.Scope.Field(nat.Attribute, nat.Name, true), validation)
		}
	}
	return validation
}

// hasValidations returns true if a UserType contains validations.
func hasValidations(attCtx *AttributeContext, ut expr.UserType) bool {
	// We need to check empirically whether there are validations to be
	// generated, we can't just generate and check whether something was
	// generated to avoid infinite recursions.
	res := false
	done := errors.New("done")
	Walk(ut.Attribute(), func(a *expr.AttributeExpr) error {
		if a.Validation == nil {
			return nil
		}
		if attCtx.Pointer && !attCtx.IgnoreRequired || !a.Validation.HasRequiredOnly() {
			res = true
			return done
		}
		// For public data structures there is a case where there is
		// validation but no actual validation code: if the validation
		// is a required validation that applies to attributes that
		// cannot be nil i.e. primitive types.
		obj := expr.AsObject(a.Type)
		for _, name := range a.Validation.Required {
			if att := obj.Attribute(name); att != nil && !expr.IsPrimitive(att.Type) {
				res = !attCtx.UseDefault || att.DefaultValue == nil
				if res {
					return done
				}
			}
		}
		return nil
	})
	return res
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

	enumValTmpl = `{{ if isset .zeroVal -}}
if {{ .target }} != {{ if and (not .zeroVal) .string }}""{{ else }}{{ .zeroVal }}{{ end }} {
{{ else if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
if !({{ oneof .targetVal .values }}) {
        err = goa.MergeErrors(err, goa.InvalidEnumValueError({{ printf "%q" .context }}, {{ .targetVal }}, {{ slice .values }}))
{{ if or (isset .zeroVal) .isPointer -}}
}
{{ end -}}
}`

	patternValTmpl = `{{ if isset .zeroVal -}}
if {{ .target }} != {{ if and (not .zeroVal) .string }}""{{ else }}{{ .zeroVal }}{{ end }} {
{{ else if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
        err = goa.MergeErrors(err, goa.ValidatePattern({{ printf "%q" .context }}, {{ .targetVal }}, {{ printf "%q" .pattern }}))
{{- if or (isset .zeroVal) .isPointer }}
}
{{- end }}`

	formatValTmpl = `{{ if isset .zeroVal -}}
if {{ .target }} != {{ if and (not .zeroVal) .string }}""{{ else }}{{ .zeroVal }}{{ end }} {
{{ else if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
        err = goa.MergeErrors(err, goa.ValidateFormat({{ printf "%q" .context }}, {{ .targetVal}}, {{ constant .format }}))
{{ if or (isset .zeroVal) .isPointer -}}
}
{{- end }}`

	exclMinMaxValTmpl = `{{ if isset .zeroVal -}}
if {{ .target }} != {{ if and (not .zeroVal) .string }}""{{ else }}{{ .zeroVal }}{{ end }} {
{{ else if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
        if {{ .targetVal }} {{ if .isExclMin }}<{{ else }}>{{ end }} {{ if .isExclMin }}{{ .exclMin }}{{ else }}{{ .exclMax }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidRangeError({{ printf "%q" .context }}, {{ .targetVal }}, {{ if .isExclMin }}{{ .exclMin }}, true{{ else }}{{ .exclMax }}, false{{ end }}))
{{ if or (isset .zeroVal) .isPointer -}}
}
{{ end -}}
}`

	minMaxValTmpl = `{{ if isset .zeroVal -}}
if {{ .target }} != {{ if and (not .zeroVal) .string }}""{{ else }}{{ .zeroVal }}{{ end }} {
{{ else if .isPointer -}}
if {{ .target }} != nil {
{{ end -}}
        if {{ .targetVal }} {{ if .isMin }}<{{ else }}>{{ end }} {{ if .isMin }}{{ .min }}{{ else }}{{ .max }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidRangeError({{ printf "%q" .context }}, {{ .targetVal }}, {{ if .isMin }}{{ .min }}, true{{ else }}{{ .max }}, false{{ end }}))
{{ if or (isset .zeroVal) .isPointer -}}
}
{{ end -}}
}`

	lengthValTmpl = `{{ $target := or (and (or (or .array .map) .nonzero) .target) .targetVal -}}
{{ if and (isset .zeroVal) .string -}}
if {{ .target }} != {{ if and (not .zeroVal) .string }}""{{ else }}{{ .zeroVal }}{{ end }} {
{{ else if and .isPointer .string -}}
if {{ .target }} != nil {
{{ end -}}
if {{ if .string }}utf8.RuneCountInString({{ $target }}){{ else }}len({{ $target }}){{ end }} {{ if .isMinLength }}<{{ else }}>{{ end }} {{ if .isMinLength }}{{ .minLength }}{{ else }}{{ .maxLength }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidLengthError({{ printf "%q" .context }}, {{ $target }}, {{ if .string }}utf8.RuneCountInString({{ $target }}){{ else }}len({{ $target }}){{ end }}, {{ if .isMinLength }}{{ .minLength }}, true{{ else }}{{ .maxLength }}, false{{ end }}))
}{{- if and (or (isset .zeroVal) .isPointer) .string }}
}
{{- end }}`

	requiredValTmpl = `if {{ $.target }}.{{ .attCtx.Scope.Field $.reqAtt .req true }} == nil {
        err = goa.MergeErrors(err, goa.MissingFieldError("{{ .req }}", {{ printf "%q" $.context }}))
}`
)
