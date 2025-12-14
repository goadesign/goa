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
	enumValT = template.Must(template.New("enum").Funcs(fm).Parse(codegenTemplates.Read(validationEnumT)))
	formatValT = template.Must(template.New("format").Funcs(fm).Parse(codegenTemplates.Read(validationFormatT)))
	patternValT = template.Must(template.New("pattern").Funcs(fm).Parse(codegenTemplates.Read(validationPatternT)))
	exclMinMaxValT = template.Must(template.New("exclMinMax").Funcs(fm).Parse(codegenTemplates.Read(validationExclMinMaxT)))
	minMaxValT = template.Must(template.New("minMax").Funcs(fm).Parse(codegenTemplates.Read(validationMinMaxT)))
	lengthValT = template.Must(template.New("length").Funcs(fm).Parse(codegenTemplates.Read(validationLengthT)))
	requiredValT = template.Must(template.New("req").Funcs(fm).Parse(codegenTemplates.Read(validationRequiredT)))
	arrayValT = template.Must(template.New("array").Funcs(fm).Parse(codegenTemplates.Read(validationArrayT)))
	mapValT = template.Must(template.New("map").Funcs(fm).Parse(codegenTemplates.Read(validationMapT)))
	unionValT = template.Must(template.New("union").Funcs(fm).Parse(codegenTemplates.Read(validationUnionT)))
	userValT = template.Must(template.New("user").Funcs(fm).Parse(codegenTemplates.Read(validationUserT)))
}

// AttributeValidationCode produces Go code that runs the validations defined
// in the given attribute against the value held by the variable named target.
//
// See ValidationCode for a description of the arguments.
func AttributeValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias bool, target, attName string) string {
	return recurseValidationCode(att, put, attCtx, req, alias, false, target, attName, nil).String()
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
// view indicates whether the attribute is a view type attribute.
// This only matters for union types: generated Goa view union types have a
// different layout than proto generated union types.
//
// target is the variable name against which the validation code is generated
//
// context is used to produce helpful messages in case of error.
func ValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias, view bool, target string) string {
	return recurseValidationCode(att, put, attCtx, req, alias, view, target, target, nil).String()
}

func recurseValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias, view bool, target, context string, seen map[string]*bytes.Buffer) *bytes.Buffer {
	if seen == nil {
		seen = make(map[string]*bytes.Buffer)
	}
	var (
		buf      = new(bytes.Buffer)
		first    = true
		ut, isUT = att.Type.(expr.UserType)
	)

	// Break infinite recursions
	// Note: when alias=true, we're validating the underlying base type,
	// so alias types shouldn't use the recursion guard. Only non-alias user
	// types need cycle protection.
	if isUT && !alias {
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
			val := validateAttribute(attCtx, nat.Attribute, put, tgt, ctx, att.IsRequired(nat.Name), view, seen)
			if val != "" {
				newline()
				buf.WriteString(val)
			}
		}
	case expr.IsArray(att.Type):
		arr := expr.AsArray(att.Type)
		elem := arr.ElemType
		ctx := attCtx
		if ctx.Pointer && expr.IsPrimitive(elem.Type) {
			// Array elements of primitive type are never pointers
			ctx = attCtx.Dup()
			ctx.Pointer = false
		}
		val := validateAttribute(ctx, elem, put, "e", context+"[*]", true, view, seen)
		if val != "" || arr.NonNullableElems {
			newline()
			data := map[string]any{
				"target":           target,
				"validation":       val,
				"nonNullableElems": arr.NonNullableElems,
				"context":          context,
			}
			if err := arrayValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	case expr.IsMap(att.Type):
		m := expr.AsMap(att.Type)
		ctx := attCtx.Dup()
		ctx.Pointer = false
		keyVal := validateAttribute(ctx, m.KeyType, put, "k", context+".key", true, view, seen)
		if keyVal != "" {
			keyVal = "\n" + keyVal
		}
		valueVal := validateAttribute(ctx, m.ElemType, put, "v", context+"[key]", true, view, seen)
		if valueVal != "" {
			valueVal = "\n" + valueVal
		}
		if keyVal != "" || valueVal != "" {
			newline()
			data := map[string]any{"target": target, "keyValidation": keyVal, "valueValidation": valueVal}
			if err := mapValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	case expr.IsUnion(att.Type):
		// NOTE: the only time we validate a union is when we are
		// validating a proto-generated type or view types since the HTTP
		// serialization transforms unions into objects.
		u := expr.AsUnion(att.Type)
		var vals []string
		var types []string
		for _, v := range u.Values {
			vatt := v.Attribute
			if view {
				// Union values in views are never pointers - they are concrete typed values
				unionCtx := attCtx.Dup()
				unionCtx.Pointer = false
				val := validateAttribute(unionCtx, vatt, put, "v", context+".value", true, view, seen)
				if val != "" {
					types = append(types, attCtx.Scope.Ref(vatt, attCtx.DefaultPkg))
					vals = append(vals, val)
				}
			} else {
				fieldName := attCtx.Scope.Field(vatt, v.Name, true)
				val := validateAttribute(attCtx, vatt, put, "v."+fieldName, context+".value", true, view, seen)
				if val != "" {
					tref := attCtx.Scope.Ref(&expr.AttributeExpr{Type: put}, attCtx.DefaultPkg)
					types = append(types, tref+"_"+fieldName)
					vals = append(vals, val)
				}
			}
		}
		if len(vals) > 0 {
			newline()
			data := map[string]any{
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

func validateAttribute(ctx *AttributeContext, att *expr.AttributeExpr, put expr.UserType, target, context string, req, view bool, seen map[string]*bytes.Buffer) string {
	ut, isUT := att.Type.(expr.UserType)
	if !isUT {
		code := recurseValidationCode(att, put, ctx, req, false, view, target, context, seen).String()
		if code == "" {
			return ""
		}
		if expr.IsArray(att.Type) || expr.IsMap(att.Type) || expr.IsUnion(att.Type) {
			return code
		}
		if !ctx.Pointer && (req || (att.DefaultValue != nil && ctx.UseDefault)) {
			return code
		}
		cond := fmt.Sprintf("if %s != nil {\n", target)
		if strings.HasPrefix(code, cond) {
			return code
		}
		return fmt.Sprintf("%s%s\n}", cond, code)
	}
	// Alias user types: validate underlying attribute with alias flag so that
	// validation operates on the base value type while preserving pointer
	// semantics from the current attribute context.
	if expr.IsAlias(ut) {
		// Preserve field-level attributes (e.g., DefaultValue, Required) while
		// validating alias user types against their underlying base. Passing
		// the original attribute with alias=true ensures validations operate
		// on the correct value type without dropping field defaults.
		code := recurseValidationCode(att, put, ctx, req, true, view, target, context, seen).String()
		if code == "" {
			return ""
		}
		// For optional pointer fields, wrap validation code in nil check
		if !ctx.Pointer && (req || (att.DefaultValue != nil && ctx.UseDefault)) {
			return code
		}
		cond := fmt.Sprintf("if %s != nil {\n", target)
		if strings.HasPrefix(code, cond) {
			return code
		}
		return fmt.Sprintf("%s%s\n}", cond, code)
	}
	if !hasValidations(ctx, ut) {
		return ""
	}
	var buf bytes.Buffer
	name := ctx.Scope.Name(att, "", ctx.Pointer, ctx.UseDefault)
	// Use the scoped type name directly to preserve identifiers such as
	// protocol buffer-reserved names that include a trailing underscore
	// (e.g., Message_). Applying Goify here would drop underscores and
	// cause mismatches between function declarations and call sites.
	data := map[string]any{"name": name, "target": target}
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
// view indicates whether the attribute is a view type attribute.
// This only matters for union types: generated Goa view union types have a
// different layout than proto generated union types.
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
		unaliased       = unalias(att.Type)
		isNativePointer = unaliased.Kind() == expr.BytesKind || unaliased.Kind() == expr.AnyKind
		isPointer       = attCtx.Pointer || (!req && (att.DefaultValue == nil || !attCtx.UseDefault))
		tval            = target
	)
	if isPointer && expr.IsPrimitive(att.Type) && !isNativePointer {
		tval = "*" + tval
	}
	if alias {
		tval = fmt.Sprintf("%s(%s)", unaliased.Name(), tval)
		// When validating alias types, use the underlying type's kind
		// for string detection (needed for utf8.RuneCountInString usage)
		kind = unaliased.Kind()
	}
	data := map[string]any{
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
	runTemplate := func(tmpl *template.Template, data any) string {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		return buf.String()
	}
	res := make([]string, 0, 8) // preallocate with typical validation count
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
	if minVal := validation.Minimum; minVal != nil {
		data["min"] = *minVal
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
	if maxVal := validation.Maximum; maxVal != nil {
		data["max"] = *maxVal
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
	Walk(ut.Attribute(), func(a *expr.AttributeExpr) error { // nolint: errcheck
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
func toSlice(val []any) string {
	elems := make([]string, len(val))
	for i, v := range val {
		elems[i] = fmt.Sprintf("%#v", v)
	}
	return fmt.Sprintf("[]any{%s}", strings.Join(elems, ", "))
}

// oneof produces code that compares target with each element of vals and ORs
// the result, e.g. "target == 1 || target == 2".
func oneof(target string, vals []any) string {
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
