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
	unionValT      *template.Template
	userValT       *template.Template
)

func init() {
	fm := template.FuncMap{
		"slice":    toSlice,
		"oneof":    oneof,
		"constant": constant,
		"add":      func(a, b int) int { return a + b },
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
	unionValT = template.Must(template.New("union").Funcs(fm).Parse(unionValTmpl))
	userValT = template.Must(template.New("user").Funcs(fm).Parse(userValTmpl))
}

// ValidationCode produces Go code that runs the validations defined in the
// given attribute and its children recursively against the value held by the
// variable named target.
//
// put is the parent UserType if any. It is used to compute proto oneof type names.
//
// attCtx is the attribute context used to generate attribute name and reference
// in the validation code.
//
// req indicates whether the attribute is required (true) or optional (false)
//
// alias indicates whether the attribute is an alias user type attribute.
//
// target is the variable name against which the validation code is generated
//
// context is used to produce helpful messages in case of error.
func ValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias bool, target string) string {
	seen := make(map[string]*bytes.Buffer)
	return recurseValidationCode(att, put, attCtx, req, alias, target, target, seen).String()
}

func recurseValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias bool, target, context string, seen map[string]*bytes.Buffer) *bytes.Buffer {
	var (
		buf      = new(bytes.Buffer)
		first    = true
		ut, isUT = att.Type.(expr.UserType)
	)

	// Break infinite recursions
	if isUT {
		if buf, ok := seen[ut.ID()]; ok {
			return buf
		}
		seen[ut.ID()] = buf
	}

	flattenValidations(att, make(map[string]struct{}))

	newline := func() {
		if !first {
			buf.WriteByte('\n')
		} else {
			first = false
		}
	}

	// Write validations on attribute if any.
	validation := validationCode(att, attCtx, req, alias, target, context)
	if validation != "" {
		buf.WriteString(validation)
		first = false
	}

	// Recurse down depending on attribute type.
	switch {
	case expr.IsObject(att.Type):
		if isUT {
			put = ut
		}
		for _, nat := range *(expr.AsObject(att.Type)) {
			tgt := fmt.Sprintf("%s.%s", target, attCtx.Scope.Field(nat.Attribute, nat.Name, true))
			ctx := fmt.Sprintf("%s.%s", context, nat.Name)
			val := validateAttribute(attCtx, nat.Attribute, put, tgt, ctx, att.IsRequired(nat.Name))
			if val != "" {
				newline()
				buf.WriteString(val)
			}
		}
	case expr.IsArray(att.Type):
		elem := expr.AsArray(att.Type).ElemType
		ctx := attCtx
		if ctx.Pointer && expr.IsPrimitive(elem.Type) {
			// Array elements of primitive type are never pointers
			ctx = attCtx.Dup()
			ctx.Pointer = false
		}
		val := validateAttribute(ctx, elem, put, "e", context+"[*]", true)
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
		keyVal := validateAttribute(ctx, m.KeyType, put, "k", context+".key", true)
		if keyVal != "" {
			keyVal = "\n" + keyVal
		}
		valueVal := validateAttribute(ctx, m.ElemType, put, "v", context+"[key]", true)
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
	case expr.IsUnion(att.Type):
		// NOTE: the only time we validate a union is when we are
		// validating a proto-generated type since the HTTP
		// serialization transforms unions into objects.
		u := expr.AsUnion(att.Type)
		tref := Goify(put.Name(), true)
		if attCtx.DefaultPkg != "" {
			tref = attCtx.DefaultPkg + "." + tref
		}
		tref = "*" + tref
		var vals []string
		var types []string
		for _, v := range u.Values {
			vatt := v.Attribute
			fieldName := attCtx.Scope.Field(vatt, v.Name, true)
			val := validateAttribute(attCtx, vatt, put, "v."+fieldName, context+".value", true)
			if val != "" {
				types = append(types, tref+"_"+fieldName)
				vals = append(vals, val)
			}
		}
		if len(vals) > 0 {
			newline()
			data := map[string]interface{}{
				"target": target,
				"types":  types,
				"values": vals,
			}
			if err := unionValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	}

	return buf
}

func validateAttribute(ctx *AttributeContext, att *expr.AttributeExpr, put expr.UserType, target, context string, req bool) string {
	ut, isUT := att.Type.(expr.UserType)
	if !isUT {
		return recurseValidationCode(att, put, ctx, req, false, target, context, nil).String()
	}
	if expr.IsAlias(ut) {
		return recurseValidationCode(ut.Attribute(), put, ctx, req, true, target, context, nil).String()
	}
	if !hasValidations(ctx, ut) {
		return ""
	}
	var buf bytes.Buffer
	name := ctx.Scope.Name(att, "", ctx.Pointer, ctx.UseDefault)
	data := map[string]interface{}{"name": Goify(name, true), "target": target}
	if err := userValT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return fmt.Sprintf("if %s != nil {\n\t%s\n}", target, buf.String())
}

// validationCode produces Go code that runs the validations defined in the
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
func validationCode(att *expr.AttributeExpr, attCtx *AttributeContext, req, alias bool, target, context string) string {
	validation := att.Validation
	if ut, ok := att.Type.(expr.UserType); ok {
		val := ut.Attribute().Validation
		if val != nil {
			if validation == nil {
				validation = val
			} else {
				validation.Merge(val)
			}
			att.Validation = validation
		}
	}
	if validation == nil {
		return ""
	}
	var (
		kind            = att.Type.Kind()
		isNativePointer = kind == expr.BytesKind || kind == expr.AnyKind
		isPointer       = attCtx.Pointer || (!req && (att.DefaultValue == nil || !attCtx.UseDefault))
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
		if val := runTemplate(exclMinMaxValT, data); val != "" {
			res = append(res, val)
		}
	}
	if min := validation.Minimum; min != nil {
		data["min"] = *min
		data["isMin"] = true
		if val := runTemplate(minMaxValT, data); val != "" {
			res = append(res, val)
		}
	}
	if exclMax := validation.ExclusiveMaximum; exclMax != nil {
		data["exclMax"] = *exclMax
		data["isExclMax"] = true
		if val := runTemplate(exclMinMaxValT, data); val != "" {
			res = append(res, val)
		}
	}
	if max := validation.Maximum; max != nil {
		data["max"] = *max
		data["isMin"] = false
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
	reqs := generatedRequiredValidation(att, attCtx)
	obj := expr.AsObject(att.Type)
	for _, r := range reqs {
		reqAtt := obj.Attribute(r)
		data["req"] = r
		data["reqAtt"] = reqAtt
		res = append(res, runTemplate(requiredValT, data))
	}
	return strings.Join(res, "\n")
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
		if attCtx.Pointer || !a.Validation.HasRequiredOnly() {
			res = true
			return done
		}
		res = len(generatedRequiredValidation(a, attCtx)) > 0
		if res {
			return done
		}
		return nil
	})
	return res
}

// There is a case where there is validation but no actual validation code: if
// the validation is a required validation that applies to attributes that
// cannot be nil i.e. primitive types.
func generatedRequiredValidation(att *expr.AttributeExpr, attCtx *AttributeContext) (res []string) {
	if att.Validation == nil {
		return
	}
	obj := expr.AsObject(att.Type)
	for _, req := range att.Validation.Required {
		reqAtt := obj.Attribute(req)
		if reqAtt == nil {
			continue
		}
		if !attCtx.Pointer && expr.IsPrimitive(reqAtt.Type) &&
			reqAtt.Type.Kind() != expr.BytesKind &&
			reqAtt.Type.Kind() != expr.AnyKind {
			continue
		}
		if attCtx.IgnoreRequired && expr.IsPrimitive(reqAtt.Type) {
			continue
		}
		res = append(res, req)
	}
	return
}

func flattenValidations(att *expr.AttributeExpr, seen map[string]struct{}) {
	switch actual := att.Type.(type) {
	case *expr.Array:
		flattenValidations(actual.ElemType, seen)
	case *expr.Map:
		flattenValidations(actual.KeyType, seen)
		flattenValidations(actual.ElemType, seen)
	case *expr.Object:
		for _, nat := range *actual {
			flattenValidations(nat.Attribute, seen)
		}
	case *expr.Union:
		for _, nat := range actual.Values {
			flattenValidations(nat.Attribute, seen)
		}
	case expr.UserType:
		if _, ok := seen[actual.ID()]; ok {
			return
		}
		seen[actual.ID()] = struct{}{}
		v := att.Validation
		ut, ok := actual.Attribute().Type.(expr.UserType)
		for ok {
			if val := ut.Attribute().Validation; val != nil {
				if v == nil {
					v = val
				} else {
					v.Merge(val)
				}
			}
			ut, ok = ut.Attribute().Type.(expr.UserType)
		}
		att.Validation = v
		flattenValidations(actual.Attribute(), seen)
	}
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

	unionValTmpl = `switch v := {{ .target }}.(type) {
{{- range $i, $val := .values }}
	case {{ index $.types $i }}:
		{{ $val }}
{{ end -}}
}`

	userValTmpl = `if err2 := Validate{{ .name }}({{ .target }}); err2 != nil {
        err = goa.MergeErrors(err, err2)
}`

	enumValTmpl = `{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
if !({{ oneof .targetVal .values }}) {
        err = goa.MergeErrors(err, goa.InvalidEnumValueError({{ printf "%q" .context }}, {{ .targetVal }}, {{ slice .values }}))
{{ if .isPointer -}}
}
{{ end -}}
}`

	patternValTmpl = `{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
        err = goa.MergeErrors(err, goa.ValidatePattern({{ printf "%q" .context }}, {{ .targetVal }}, {{ printf "%q" .pattern }}))
{{- if .isPointer }}
}
{{- end }}`

	formatValTmpl = `{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
        err = goa.MergeErrors(err, goa.ValidateFormat({{ printf "%q" .context }}, {{ .targetVal}}, {{ constant .format }}))
{{- if .isPointer }}
}
{{- end }}`

	exclMinMaxValTmpl = `{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
        if {{ .targetVal }} {{ if .isExclMin }}<={{ else }}>={{ end }} {{ if .isExclMin }}{{ .exclMin }}{{ else }}{{ .exclMax }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidRangeError({{ printf "%q" .context }}, {{ .targetVal }}, {{ if .isExclMin }}{{ .exclMin }}, true{{ else }}{{ .exclMax }}, false{{ end }}))
{{ if .isPointer -}}
}
{{ end -}}
}`

	minMaxValTmpl = `{{ if .isPointer -}}if {{ .target }} != nil {
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

	requiredValTmpl = `if {{ $.target }}.{{ .attCtx.Scope.Field $.reqAtt .req true }} == nil {
        err = goa.MergeErrors(err, goa.MissingFieldError("{{ .req }}", {{ printf "%q" $.context }}))
}`
)
