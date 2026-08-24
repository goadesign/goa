// This file generates functions that check service values and values sent over
// HTTP, gRPC, and JSON-RPC. Each function uses the Go names already chosen for
// its package.
package codegen

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"goa.design/goa/v3/expr"
)

type (
	// unionValidationCase describes one possible union branch in generated
	// validation code.
	unionValidationCase struct {
		// Type is the generated Go type for the branch.
		Type string
		// Field is the field which stores the branch value.
		Field string
		// Name is the branch name shown in validation errors.
		Name string
		// PayloadRequiresPresence is true when selecting this branch also
		// requires a non-nil value.
		PayloadRequiresPresence bool
		// Validation checks the value stored by this branch.
		Validation string
	}

	// unionValidationData contains the information needed to write one union
	// check.
	unionValidationData struct {
		// Target is the generated union value being checked.
		Target string
		// Context identifies the union in validation errors.
		Context validationPath
		// Protobuf is true when each selected branch is stored in its own generated
		// protobuf struct.
		Protobuf bool
		// Cases lists every branch accepted by the union.
		Cases []unionValidationCase
		// Goa is the generated import name of Goa's error package.
		Goa string
	}

	// validationPath stores an error path while Goa writes validation source.
	// variable is true when root names a parameter in the generated function.
	validationPath struct {
		root     string
		suffix   string
		variable bool
	}
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
	unionSumValT   *template.Template
	userValT       *template.Template
)

func init() {
	fm := template.FuncMap{
		"slice":          toSlice,
		"oneof":          oneof,
		"constant":       constant,
		"validationPath": renderValidationPath,
		"isUnion": func(att *expr.AttributeExpr) bool {
			if att == nil {
				return false
			}
			return expr.IsUnion(att.Type)
		},
		"isSumType": func(scope Attributor) bool {
			if scope == nil {
				return false
			}
			return scope.IsSumType()
		},
		"isUnionPointer": func(ctx *AttributeContext, required bool) bool {
			return ctx.IsUnionPointer(required)
		},
		"add": func(a, b int) int { return a + b },
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
	unionSumValT = template.Must(template.New("union-sum").Funcs(fm).Parse(codegenTemplates.Read(validationUnionSumT)))
	userValT = template.Must(template.New("user").Funcs(fm).Parse(codegenTemplates.Read(validationUserT)))
}

// AttributeValidationCode produces Go code that runs the validations defined
// in the given attribute against the value held by the variable named target.
//
// See ValidationCode for a description of the arguments.
func AttributeValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias bool, target, attName string) string {
	return recurseValidationCode(att, put, attCtx, req, alias, false, target, literalValidationPath(attName), nil).String()
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
	return recurseValidationCode(att, put, attCtx, req, alias, view, target, literalValidationPath(target), nil).String()
}

// ValidationCodeWithPathParameter produces validation code whose error paths
// begin with the string held by pathParameter. target and pathParameter are Go
// expressions.
func ValidationCodeWithPathParameter(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias, view bool, target, pathParameter string) string {
	return recurseValidationCode(att, put, attCtx, req, alias, view, target, parameterValidationPath(pathParameter), nil).String()
}

func recurseValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias, view bool, target string, context validationPath, seen map[expr.UserType]*bytes.Buffer) *bytes.Buffer {
	return renderValidationCode(att, put, attCtx, req, alias, view, target, context, seen, true)
}

// renderValidationCode writes one validation tree. localGuards reports whether
// local rule templates must check a pointer before reading it. Nested fields
// disable those checks when validateAttribute wraps the whole field once.
func renderValidationCode(att *expr.AttributeExpr, put expr.UserType, attCtx *AttributeContext, req, alias, view bool, target string, context validationPath, seen map[expr.UserType]*bytes.Buffer, localGuards bool) *bytes.Buffer {
	if seen == nil {
		seen = make(map[expr.UserType]*bytes.Buffer)
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
		origin := ut.Origin()
		if buf, ok := seen[origin]; ok {
			return buf
		}
		seen[origin] = buf
	}

	newline := func() {
		if !first {
			buf.WriteByte('\n')
		} else {
			first = false
		}
	}

	// Write validations on attribute if any.
	validation := validationCode(att, attCtx, req, alias, target, context, localGuards)
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
			ctx := context.child("." + nat.Name)
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
		if expr.IsPrimitive(elem.Type) {
			ctx = attCtx.Dup()
			ctx.Pointer = attCtx.IsArrayElementPointer(arr)
		}
		val := validateAttribute(ctx, elem, put, "e", context.child("[*]"), true, view, seen)
		nonNullableElems := arr.NonNullableElems &&
			(IsNilable(elem.Type) || attCtx.IsArrayElementPointer(arr))
		if val != "" || nonNullableElems {
			newline()
			data := map[string]any{
				"target":           target,
				"validation":       val,
				"checkNilElements": nonNullableElems,
				"context":          context,
				"goa":              "goa",
			}
			if err := arrayValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	case expr.IsMap(att.Type):
		m := expr.AsMap(att.Type)
		ctx := attCtx.Dup()
		ctx.Pointer = false
		keyVal := validateAttribute(ctx, m.KeyType, put, "k", context.child(".key"), true, view, seen)
		if keyVal != "" {
			keyVal = "\n" + keyVal
		}
		valueVal := validateAttribute(ctx, m.ElemType, put, "v", context.child("[key]"), true, view, seen)
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
		u := expr.AsUnion(att.Type)
		if attCtx.Scope.IsSumType() {
			cases := make([]map[string]any, 0, len(u.Values))
			for _, v := range u.Values {
				// Sum-type unions (struct-based, with Kind/AsX accessors) store each
				// branch as either a value (primitives, arrays, maps) or a pointer
				// (object user types). Request-body validation may already use value
				// semantics for nested objects, so preserve the enclosing context and
				// only keep pointer semantics when both layers use pointers.
				unionCtx := attCtx.Dup()
				unionCtx.Pointer = unionCtx.Pointer && expr.IsObject(v.Attribute.Type)
				val := validateAttribute(unionCtx, v.Attribute, put, "actual", context.child(".value"), true, view, seen)
				if val == "" {
					continue
				}
				cases = append(cases, map[string]any{
					"typeTag":    v.Name,
					"fieldName":  Goify(v.Name, true),
					"validation": val,
				})
			}
			if len(cases) > 0 {
				newline()
				data := map[string]any{
					"target": target,
					"cases":  cases,
					"goa":    "goa",
				}
				if err := unionSumValT.Execute(buf, data); err != nil {
					panic(err) // bug
				}
			}
			break
		}

		// Validate unions represented as interfaces (e.g., protobuf oneof wrappers).
		var cases []unionValidationCase
		for _, v := range u.Values {
			vatt := v.Attribute
			if view {
				// Union values in views are never pointers - they are concrete typed values
				unionCtx := attCtx.Dup()
				unionCtx.Pointer = false
				val := validateAttribute(unionCtx, vatt, put, "v", context.child(".value"), true, view, seen)
				if val != "" {
					cases = append(cases, unionValidationCase{
						Type:       attCtx.Scope.Ref(vatt, attCtx.Pkg(vatt)),
						Validation: val,
					})
				}
			} else {
				fieldName := attCtx.Scope.Field(vatt, v.Name, true)
				val := validateAttribute(attCtx, vatt, put, "v."+fieldName, context.child(".value"), true, view, seen)
				parent := &expr.AttributeExpr{Type: put}
				tref := attCtx.Scope.Ref(parent, attCtx.Pkg(parent))
				cases = append(cases, unionValidationCase{
					Type:                    tref + "_" + fieldName,
					Field:                   fieldName,
					Name:                    v.Name,
					PayloadRequiresPresence: protobufUnionPayloadRequiresPresence(vatt),
					Validation:              val,
				})
			}
		}
		if len(cases) > 0 {
			newline()
			data := unionValidationData{
				Target:   target,
				Context:  context,
				Protobuf: !view,
				Cases:    cases,
				Goa:      "goa",
			}
			if err := unionValT.Execute(buf, data); err != nil {
				panic(err) // bug
			}
		}
	}

	return buf
}

// protobufUnionPayloadRequiresPresence reports whether selecting a protobuf
// union branch requires a non-nil value. Messages, byte slices, and Any values
// may be nil in Go, so their generated checks must reject nil explicitly.
func protobufUnionPayloadRequiresPresence(att *expr.AttributeExpr) bool {
	kind := unalias(att.Type).Kind()
	return !expr.IsPrimitive(att.Type) || kind == expr.BytesKind || kind == expr.AnyKind
}

func validateAttribute(ctx *AttributeContext, att *expr.AttributeExpr, put expr.UserType, target string, context validationPath, req, view bool, seen map[expr.UserType]*bytes.Buffer) string {
	ut, isUT := att.Type.(expr.UserType)
	if !isUT {
		guard := validationAttributeNeedsNilGuard(att, ctx, req)
		code := renderValidationCode(att, put, ctx, req, false, view, target, context, seen, !guard).String()
		if code == "" {
			return ""
		}
		if expr.IsArray(att.Type) || expr.IsMap(att.Type) {
			return code
		}
		if !guard {
			return code
		}
		cond := fmt.Sprintf("if %s != nil {\n", target)
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
		guard := validationAttributeNeedsNilGuard(att, ctx, req)
		code := renderValidationCode(att, put, ctx, req, true, view, target, context, seen, !guard).String()
		if code == "" {
			return ""
		}
		if !guard {
			return code
		}
		cond := fmt.Sprintf("if %s != nil {\n", target)
		return fmt.Sprintf("%s%s\n}", cond, code)
	}
	if !hasValidations(ctx, ut) {
		return ""
	}
	var buf bytes.Buffer
	call := ctx.Scope.ValidatorCall(att, "", target, renderValidationPath(context))
	data := map[string]any{"call": call, "goa": "goa"}
	if err := userValT.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return fmt.Sprintf("if %s != nil {\n\t%s\n}", target, buf.String())
}

// validationAttributeNeedsNilGuard reports whether a nested value may be nil
// in the generated Go layout and must be checked before any validation uses it.
func validationAttributeNeedsNilGuard(att *expr.AttributeExpr, ctx *AttributeContext, required bool) bool {
	if expr.IsArray(att.Type) || expr.IsMap(att.Type) {
		return false
	}
	if expr.IsUnion(att.Type) {
		if ctx.Scope.IsSumType() {
			return ctx.IsUnionPointer(required)
		}
		return !required
	}
	return ctx.Pointer || !required && (att.DefaultValue == nil || !ctx.UseDefault)
}

// validationCode produces Go code that runs the validations that effectively
// apply to the given attribute - see expr.EffectiveValidation - if any
// against the content of the variable named target. The generated code
// assumes that there is a pre-existing "err" variable of type error. It
// initializes that variable in case a validation fails. validationCode is
// pure: it never mutates att or any expression reachable from it.
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
func validationCode(att *expr.AttributeExpr, attCtx *AttributeContext, req, alias bool, target string, context validationPath, localGuards bool) string {
	validation := expr.EffectiveValidation(att)
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
		"isPointer": isPointer && localGuards,
		"context":   context,
		"target":    target,
		"targetVal": tval,
		"goa":       "goa",
		"utf8":      "utf8",
		"string":    kind == expr.StringKind,
		"array":     expr.IsArray(att.Type),
		"map":       expr.IsMap(att.Type),
	}
	runTemplate := func(tmpl *template.Template, data any) string {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		return strings.Trim(buf.String(), "\n")
	}
	res := make([]string, 0, 8) // preallocate with typical validation count
	if values := validation.Values; values != nil {
		data["values"] = values
		if val := runTemplate(enumValT, data); val != "" {
			res = append(res, val)
		}
	}
	if format := validation.Format; format != "" {
		// Skip format validation when struct:field:type overrides a string attribute
		// with a custom type — the custom type's own parsing (e.g. UnmarshalText)
		// already validates the format, and ValidateFormat expects a plain string.
		typeName, _ := GetMetaType(att)
		if typeName == "" {
			data["format"] = string(format)
			if val := runTemplate(formatValT, data); val != "" {
				res = append(res, val)
			}
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
		data["isExclMin"] = false
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
	reqs := generatedRequiredValidation(att, validation, attCtx)
	obj := expr.AsObject(att.Type)
	for _, r := range reqs {
		reqAtt := obj.Attribute(r)
		data["req"] = r
		data["reqAtt"] = reqAtt
		res = append(res, runTemplate(requiredValT, data))
	}
	return strings.Join(res, "\n")
}

// literalValidationPath folds a complete error path into a quoted Go string
// while Goa is generating source.
func literalValidationPath(root string) validationPath {
	return validationPath{root: root}
}

// parameterValidationPath writes an error path relative to the string held by
// a generated validator parameter.
func parameterValidationPath(parameter string) validationPath {
	return validationPath{root: parameter, variable: true}
}

// child returns the context used for a field or collection value below c.
func (p validationPath) child(prefix string) validationPath {
	p.suffix += prefix
	return p
}

// renderValidationPath returns the Go expression passed to a generated
// validation error.
func renderValidationPath(path validationPath) string {
	if !path.variable {
		return strconv.Quote(path.root + path.suffix)
	}
	if path.suffix == "" {
		return path.root
	}
	return path.root + " + " + strconv.Quote(path.suffix)
}

// hasValidations reports whether validating ut can write any code with the Go
// layout described by attCtx.
func hasValidations(attCtx *AttributeContext, ut expr.UserType) bool {
	policy := GoLayoutPolicy{
		Pointer:             attCtx.Pointer,
		IgnoreRequired:      attCtx.IgnoreRequired,
		UseDefault:          attCtx.UseDefault,
		UnionPointer:        attCtx.UnionPointer,
		ArrayElementPointer: attCtx.ArrayElementPointer,
		SumType:             attCtx.Scope.IsSumType(),
	}
	return NeedsValidation(ut.Attribute(), policy)
}

// There is a case where there is validation but no actual validation code: if
// the validation is a required validation that applies to attributes that
// cannot be nil i.e. primitive types. val is the validation that effectively
// applies to att as computed by expr.EffectiveValidation.
func generatedRequiredValidation(att *expr.AttributeExpr, val *expr.ValidationExpr, attCtx *AttributeContext) (res []string) {
	obj := expr.AsObject(att.Type)
	for _, req := range val.Required {
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
