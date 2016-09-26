package codegen

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

var (
	arrayValT    *template.Template
	userValT     *template.Template
	enumValT     *template.Template
	formatValT   *template.Template
	patternValT  *template.Template
	minMaxValT   *template.Template
	lengthValT   *template.Template
	requiredValT *template.Template
)

//  init instantiates the templates.
func init() {
	var err error
	fm := template.FuncMap{
		"tabs":             Tabs,
		"slice":            toSlice,
		"oneof":            oneof,
		"constant":         constant,
		"goifyAtt":         GoifyAtt,
		"add":              Add,
		"recursiveChecker": RecursiveChecker,
	}
	if arrayValT, err = template.New("array").Funcs(fm).Parse(arrayValTmpl); err != nil {
		panic(err)
	}
	if userValT, err = template.New("user").Funcs(fm).Parse(userValTmpl); err != nil {
		panic(err)
	}
	if enumValT, err = template.New("enum").Funcs(fm).Parse(enumValTmpl); err != nil {
		panic(err)
	}
	if formatValT, err = template.New("format").Funcs(fm).Parse(formatValTmpl); err != nil {
		panic(err)
	}
	if patternValT, err = template.New("pattern").Funcs(fm).Parse(patternValTmpl); err != nil {
		panic(err)
	}
	if minMaxValT, err = template.New("minMax").Funcs(fm).Parse(minMaxValTmpl); err != nil {
		panic(err)
	}
	if lengthValT, err = template.New("length").Funcs(fm).Parse(lengthValTmpl); err != nil {
		panic(err)
	}
	if requiredValT, err = template.New("required").Funcs(fm).Parse(requiredValTmpl); err != nil {
		panic(err)
	}
}

// RecursiveChecker produces Go code that runs the validation checks recursively over the given
// attribute.
func RecursiveChecker(att *design.AttributeDefinition, nonzero, required, hasDefault bool, target, context string, depth int, private bool) string {
	var checks []string
	if o, ok := att.Type.(design.Object); ok {
		if ds, ok := att.Type.(design.DataStructure); ok {
			att = ds.Definition()
		}
		validation := ValidationChecker(att, nonzero, required, hasDefault, target, context, depth, private)
		if validation != "" {
			checks = append(checks, validation)
		}
		o.IterateAttributes(func(n string, catt *design.AttributeDefinition) error {
			var validation string
			if ds, ok := catt.Type.(design.DataStructure); ok {
				// We need to check empirically whether there are validations to be
				// generated, we can't just generate and check whether something was
				// generated to avoid infinite recursions.
				hasValidations := false
				done := errors.New("done")
				ds.Walk(func(a *design.AttributeDefinition) error {
					if a.Validation != nil {
						if private {
							hasValidations = true
							return done
						}
						// For public data structures there is a case where
						// there is validation but no actual validation
						// code: if the validation is a required validation
						// that applies to attributes that cannot be nil or
						// empty string i.e. primitive types other than
						// string.
						if !a.Validation.HasRequiredOnly() {
							hasValidations = true
							return done
						}
						for _, name := range a.Validation.Required {
							att := a.Type.(Object)[name]
							if att != nil && (!att.Type.IsPrimitive() || att.Type.Kind() == design.StringKind) {
								hasValidations = true
								return done
							}
						}
					}
					return nil
				})
				if hasValidations {
					validation = RunTemplate(
						userValT,
						map[string]interface{}{
							"depth":  depth,
							"target": fmt.Sprintf("%s.%s", target, GoifyAtt(catt, n, true)),
						},
					)
				}
			} else {
				dp := depth
				if catt.Type.IsObject() {
					dp++
				}
				validation = RecursiveChecker(
					catt,
					att.IsNonZero(n),
					att.IsRequired(n),
					att.HasDefaultValue(n),
					fmt.Sprintf("%s.%s", target, GoifyAtt(catt, n, true)),
					fmt.Sprintf("%s.%s", context, n),
					dp,
					private,
				)
			}
			if validation != "" {
				if catt.Type.IsObject() {
					validation = fmt.Sprintf("%sif %s.%s != nil {\n%s\n%s}",
						Tabs(depth), target, GoifyAtt(catt, n, true), validation, Tabs(depth))
				}
				checks = append(checks, validation)
			}
			return nil
		})
	} else if a := att.Type.ToArray(); a != nil {
		// Perform any validation on the array type such as MinLength, MaxLength, etc.
		validation := ValidationChecker(att, nonzero, required, hasDefault, target, context, depth, private)
		if validation != "" {
			checks = append(checks, validation)
		}
		data := map[string]interface{}{
			"elemType": a.ElemType,
			"context":  context,
			"target":   target,
			"depth":    1,
			"private":  private,
		}
		validation = RunTemplate(arrayValT, data)
		if validation != "" {
			checks = append(checks, validation)
		}
	} else {
		validation := ValidationChecker(att, nonzero, required, hasDefault, target, context, depth, private)
		if validation != "" {
			checks = append(checks, validation)
		}
	}
	return strings.Join(checks, "\n")
}

// ValidationChecker produces Go code that runs the validation defined in the given attribute
// definition against the content of the variable named target recursively.
// context is used to keep track of recursion to produce helpful error messages in case of type
// validation error.
// The generated code assumes that there is a pre-existing "err" variable of type
// error. It initializes that variable in case a validation fails.
// Note: we do not want to recurse here, recursion is done by the marshaler/unmarshaler code.
func ValidationChecker(att *design.AttributeDefinition, nonzero, required, hasDefault bool, target, context string, depth int, private bool) string {
	t := target
	isPointer := private || (!required && !hasDefault && !nonzero)
	if isPointer && att.Type.IsPrimitive() {
		t = "*" + t
	}
	data := map[string]interface{}{
		"attribute": att,
		"isPointer": private || isPointer,
		"nonzero":   nonzero,
		"context":   context,
		"target":    target,
		"targetVal": t,
		"string":    att.Type.Name() == "string",
		"array":     att.Type.IsArray(),
		"hash":      att.Type.IsHash(),
		"depth":     depth,
		"private":   private,
	}
	res := validationsCode(att.Validation, data)
	return strings.Join(res, "\n")
}

func validationsCode(validation *dslengine.ValidationDefinition, data map[string]interface{}) (res []string) {
	if validation == nil {
		return nil
	}
	if values := validation.Values; values != nil {
		data["values"] = values
		if val := RunTemplate(enumValT, data); val != "" {
			res = append(res, val)
		}
	}
	if format := validation.Format; format != "" {
		data["format"] = format
		if val := RunTemplate(formatValT, data); val != "" {
			res = append(res, val)
		}
	}
	if pattern := validation.Pattern; pattern != "" {
		data["pattern"] = pattern
		if val := RunTemplate(patternValT, data); val != "" {
			res = append(res, val)
		}
	}
	if min := validation.Minimum; min != nil {
		data["min"] = *min
		data["isMin"] = true
		delete(data, "max")
		if val := RunTemplate(minMaxValT, data); val != "" {
			res = append(res, val)
		}
	}
	if max := validation.Maximum; max != nil {
		data["max"] = *max
		data["isMin"] = false
		delete(data, "min")
		if val := RunTemplate(minMaxValT, data); val != "" {
			res = append(res, val)
		}
	}
	if minLength := validation.MinLength; minLength != nil {
		data["minLength"] = minLength
		data["isMinLength"] = true
		delete(data, "maxLength")
		if val := RunTemplate(lengthValT, data); val != "" {
			res = append(res, val)
		}
	}
	if maxLength := validation.MaxLength; maxLength != nil {
		data["maxLength"] = maxLength
		data["isMinLength"] = false
		delete(data, "minLength")
		if val := RunTemplate(lengthValT, data); val != "" {
			res = append(res, val)
		}
	}
	if required := validation.Required; len(required) > 0 {
		data["required"] = required
		if val := RunTemplate(requiredValT, data); val != "" {
			res = append(res, val)
		}
	}
	return
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
	case "date-time":
		return "goa.FormatDateTime"
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
	}
	panic("unknown format") // bug
}

const (
	arrayValTmpl = `{{$validation := recursiveChecker .elemType false false false "e" (printf "%s[*]" .context) (add .depth 1) .private}}{{/*
*/}}{{if $validation}}{{tabs .depth}}for _, e := range {{.target}} {
{{$validation}}
{{tabs .depth}}}{{end}}`

	userValTmpl = `{{tabs .depth}}if err2 := {{.target}}.Validate(); err2 != nil {
{{tabs .depth}}	err = goa.MergeErrors(err, err2)
{{tabs .depth}}}`

	enumValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs $depth}}if !({{oneof .targetVal .values}}) {
{{tabs $depth}}	err = goa.MergeErrors(err, goa.InvalidEnumValueError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, {{slice .values}}))
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	patternValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs $depth}}if ok := goa.ValidatePattern(` + "`{{.pattern}}`" + `, {{.targetVal}}); !ok {
{{tabs $depth}}	err = goa.MergeErrors(err, goa.InvalidPatternError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, ` + "`{{.pattern}}`" + `))
{{tabs $depth}}}{{if .isPointer}}
{{tabs .depth}}}{{end}}`

	formatValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs $depth}}if err2 := goa.ValidateFormat({{constant .format}}, {{.targetVal}}); err2 != nil {
{{tabs $depth}}		err = goa.MergeErrors(err, goa.InvalidFormatError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, {{constant .format}}, err2))
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	minMaxValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs .depth}}	if {{.targetVal}} {{if .isMin}}<{{else}}>{{end}} {{if .isMin}}{{.min}}{{else}}{{.max}}{{end}} {
{{tabs $depth}}	err = goa.MergeErrors(err, goa.InvalidRangeError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, {{if .isMin}}{{.min}}, true{{else}}{{.max}}, false{{end}}))
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	lengthValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{$target := or (and (or (or .array .hash) .nonzero) .target) .targetVal}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs .depth}}	if {{if .string}}utf8.RuneCountInString({{$target}}){{else}}len({{$target}}){{end}} {{if .isMinLength}}<{{else}}>{{end}} {{if .isMinLength}}{{.minLength}}{{else}}{{.maxLength}}{{end}} {
{{tabs $depth}}	err = goa.MergeErrors(err, goa.InvalidLengthError(` + "`" + `{{.context}}` + "`" + `, {{$target}}, {{if .string}}utf8.RuneCountInString({{$target}}){{else}}len({{$target}}){{end}}, {{if .isMinLength}}{{.minLength}}, true{{else}}{{.maxLength}}, false{{end}}))
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	requiredValTmpl = `{{range $r := .required}}{{$catt := index ToObject($.attribute.Type) $r}}{{/*
*/}}{{if and (not $.private) (eq $catt.Type.Kind 4)}}{{tabs $.depth}}if {{$.target}}.{{goifyAtt $catt $r true}} == "" {
{{tabs $.depth}}	err = goa.MergeErrors(err, goa.MissingAttributeError(` + "`" + `{{$.context}}` + "`" + `, "{{$r}}"))
{{tabs $.depth}}}
{{else if or $.private (not $catt.Type.IsPrimitive)}}{{tabs $.depth}}if {{$.target}}.{{goifyAtt $catt $r true}} == nil {
{{tabs $.depth}}	err = goa.MergeErrors(err, goa.MissingAttributeError(` + "`" + `{{$.context}}` + "`" + `, "{{$r}}"))
{{tabs $.depth}}}
{{end}}{{end}}`
)
