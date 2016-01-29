package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/engine"
)

var (
	arrayValT    *template.Template
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
		"goify":            Goify,
		"add":              func(a, b int) int { return a + b },
		"recursiveChecker": RecursiveChecker,
	}
	if arrayValT, err = template.New("array").Funcs(fm).Parse(arrayValTmpl); err != nil {
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
func RecursiveChecker(att *design.AttributeDefinition, nonzero, required bool, target, context string, depth int) string {
	var checks []string
	validation := ValidationChecker(att, nonzero, required, target, context, depth)
	if validation != "" {
		checks = append(checks, validation)
	}
	if o := att.Type.ToObject(); o != nil {
		if mt, ok := att.Type.(*design.MediaTypeDefinition); ok {
			att = mt.AttributeDefinition
		} else if ut, ok := att.Type.(*design.UserTypeDefinition); ok {
			att = ut.AttributeDefinition
		}
		o.IterateAttributes(func(n string, catt *design.AttributeDefinition) error {
			actualDepth := depth
			if !att.IsRequired(n) && catt.Type.IsObject() {
				actualDepth = depth + 1
			}
			validation := RecursiveChecker(
				catt,
				att.IsNonZero(n),
				att.IsRequired(n),
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				fmt.Sprintf("%s.%s", context, n),
				actualDepth,
			)
			if validation != "" {
				if !att.IsRequired(n) && catt.Type.IsObject() {
					validation = fmt.Sprintf("%sif %s.%s != nil {\n%s\n%s}",
						Tabs(depth), target, Goify(n, true), validation, Tabs(depth))
				}
				checks = append(checks, validation)
			}
			return nil
		})
	} else if a := att.Type.ToArray(); a != nil {
		data := map[string]interface{}{
			"elemType": a.ElemType,
			"context":  context,
			"target":   target,
			"depth":    1,
		}
		validation := RunTemplate(arrayValT, data)
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
func ValidationChecker(att *design.AttributeDefinition, nonzero, required bool, target, context string, depth int) string {
	t := target
	isPointer := !required && !nonzero
	if isPointer && att.Type.IsPrimitive() {
		t = "*" + t
	}
	data := map[string]interface{}{
		"attribute": att,
		"isPointer": isPointer,
		"nonzero":   nonzero,
		"context":   context,
		"target":    target,
		"targetVal": t,
		"array":     att.Type.IsArray(),
		"depth":     depth,
	}
	res := validationsCode(att.Validations, data)
	return strings.Join(res, "\n")
}

func validationsCode(validations []engine.ValidationDefinition, data map[string]interface{}) (res []string) {
	for _, v := range validations {
		switch actual := v.(type) {
		case *engine.EnumValidationDefinition:
			data["values"] = actual.Values
			if val := RunTemplate(enumValT, data); val != "" {
				res = append(res, val)
			}
		case *engine.FormatValidationDefinition:
			data["format"] = actual.Format
			if val := RunTemplate(formatValT, data); val != "" {
				res = append(res, val)
			}
		case *engine.PatternValidationDefinition:
			data["pattern"] = actual.Pattern
			if val := RunTemplate(patternValT, data); val != "" {
				res = append(res, val)
			}
		case *engine.MinimumValidationDefinition:
			data["min"] = actual.Min
			data["isMin"] = true
			delete(data, "max")
			if val := RunTemplate(minMaxValT, data); val != "" {
				res = append(res, val)
			}
		case *engine.MaximumValidationDefinition:
			data["max"] = actual.Max
			data["isMin"] = false
			delete(data, "min")
			if val := RunTemplate(minMaxValT, data); val != "" {
				res = append(res, val)
			}
		case *engine.MinLengthValidationDefinition:
			data["minLength"] = actual.MinLength
			data["isMinLength"] = true
			delete(data, "maxLength")
			if val := RunTemplate(lengthValT, data); val != "" {
				res = append(res, val)
			}
		case *engine.MaxLengthValidationDefinition:
			data["maxLength"] = actual.MaxLength
			data["isMinLength"] = false
			delete(data, "minLength")
			if val := RunTemplate(lengthValT, data); val != "" {
				res = append(res, val)
			}
		case *engine.RequiredValidationDefinition:
			data["required"] = actual.Names
			if val := RunTemplate(requiredValT, data); val != "" {
				res = append(res, val)
			}
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
	arrayValTmpl = `{{$validation := recursiveChecker .elemType false false "e" (printf "%s[*]" .context) (add .depth 1)}}{{/*
*/}}{{if $validation}}{{tabs .depth}}for _, e := range {{.target}} {
{{$validation}}
{{tabs .depth}}}{{end}}`

	enumValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs $depth}}if !({{oneof .targetVal .values}}) {
{{tabs $depth}}	err = goa.InvalidEnumValueError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, {{slice .values}}, err)
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	patternValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs $depth}}if ok := goa.ValidatePattern(` + "`{{.pattern}}`" + `, {{.targetVal}}); !ok {
{{tabs $depth}}	err = goa.InvalidPatternError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, ` + "`{{.pattern}}`" + `, err)
{{tabs $depth}}}{{if .isPointer}}
{{tabs .depth}}}{{end}}`

	formatValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs $depth}}if err2 := goa.ValidateFormat({{constant .format}}, {{.targetVal}}); err2 != nil {
{{tabs $depth}}		err = goa.InvalidFormatError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, {{constant .format}}, err2, err)
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	minMaxValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs .depth}}	if {{.targetVal}} {{if .isMin}}<{{else}}>{{end}} {{if .isMin}}{{.min}}{{else}}{{.max}}{{end}} {
{{tabs $depth}}	err = goa.InvalidRangeError(` + "`" + `{{.context}}` + "`" + `, {{.targetVal}}, {{if .isMin}}{{.min}}, true{{else}}{{.max}}, false{{end}}, err)
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	lengthValTmpl = `{{$depth := or (and .isPointer (add .depth 1)) .depth}}{{/*
*/}}{{$target := or (and (or .array .nonzero) .target) .targetVal}}{{/*
*/}}{{if .isPointer}}{{tabs .depth}}if {{.target}} != nil {
{{end}}{{tabs .depth}}if len({{$target}}) {{if .isMinLength}}<{{else}}>{{end}} {{if .isMinLength}}{{.minLength}}{{else}}{{.maxLength}}{{end}} {
{{tabs $depth}}	err = goa.InvalidLengthError(` + "`" + `{{.context}}` + "`" + `, {{$target}}, len({{$target}}), {{if .isMinLength}}{{.minLength}}, true{{else}}{{.maxLength}}, false{{end}}, err)
{{if .isPointer}}{{tabs $depth}}}
{{end}}{{tabs .depth}}}`

	requiredValTmpl = `{{$ctx := .}}{{range $r := .required}}{{$catt := index $ctx.attribute.Type.ToObject $r}}{{if eq $catt.Type.Kind 4}}{{tabs $ctx.depth}}if {{$ctx.target}}.{{goify $r true}} == "" {
{{tabs $ctx.depth}}	err = goa.MissingAttributeError(` + "`" + `{{$ctx.context}}` + "`" + `, "{{$r}}", err)
{{tabs $ctx.depth}}}{{else if (not $catt.Type.IsPrimitive)}}{{tabs $ctx.depth}}if {{$ctx.target}}.{{goify $r true}} == nil {
{{tabs $ctx.depth}}	err = goa.MissingAttributeError(` + "`" + `{{$ctx.context}}` + "`" + `, "{{$r}}", err)
{{tabs $ctx.depth}}}{{end}}
{{end}}`
)
