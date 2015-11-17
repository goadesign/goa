package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/raphael/goa/design"
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
func RecursiveChecker(att *design.AttributeDefinition, required bool, target, context string, depth int) string {
	var checks []string
	validation := ValidationChecker(att, required, target, context, depth)
	if validation != "" {
		checks = append(checks, validation)
	}
	if o := att.Type.ToObject(); o != nil {
		o.IterateAttributes(func(n string, catt *design.AttributeDefinition) error {
			validation := RecursiveChecker(
				catt,
				att.IsRequired(n),
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				fmt.Sprintf("%s.%s", context, n),
				depth+1,
			)
			if validation != "" {
				checks = append(checks, validation)
			}
			return nil
		})
	} else if a := att.Type.ToArray(); a != nil {
		data := map[string]interface{}{
			"attribute": att,
			"context":   context,
			"target":    target,
			"depth":     1,
		}
		validation := runTemplate(arrayValT, data)
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
func ValidationChecker(att *design.AttributeDefinition, required bool, target, context string, depth int) string {
	data := map[string]interface{}{
		"attribute": att,
		"required":  required,
		"context":   context,
		"target":    target,
		"depth":     depth,
	}
	var res []string
	for _, v := range att.Validations {
		switch actual := v.(type) {
		case *design.EnumValidationDefinition:
			data["values"] = actual.Values
			if val := runTemplate(enumValT, data); val != "" {
				res = append(res, val)
			}
		case *design.FormatValidationDefinition:
			data["format"] = actual.Format
			if val := runTemplate(formatValT, data); val != "" {
				res = append(res, val)
			}
		case *design.PatternValidationDefinition:
			data["pattern"] = actual.Pattern
			if val := runTemplate(patternValT, data); val != "" {
				res = append(res, val)
			}
		case *design.MinimumValidationDefinition:
			data["min"] = actual.Min
			delete(data, "max")
			if val := runTemplate(minMaxValT, data); val != "" {
				res = append(res, val)
			}
		case *design.MaximumValidationDefinition:
			data["max"] = actual.Max
			delete(data, "min")
			if val := runTemplate(minMaxValT, data); val != "" {
				res = append(res, val)
			}
		case *design.MinLengthValidationDefinition:
			data["minLength"] = actual.MinLength
			delete(data, "maxLength")
			if val := runTemplate(lengthValT, data); val != "" {
				res = append(res, val)
			}
		case *design.MaxLengthValidationDefinition:
			data["maxLength"] = actual.MaxLength
			delete(data, "minLength")
			if val := runTemplate(lengthValT, data); val != "" {
				res = append(res, val)
			}
		case *design.RequiredValidationDefinition:
			data["required"] = actual.Names
			if val := runTemplate(requiredValT, data); val != "" {
				res = append(res, val)
			}
		}
	}
	return strings.Join(res, "\n")
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
	arrayValTmpl = `{{$validation := recursiveChecker .attribute.Type.ToArray.ElemType false "e" (printf "%s[*]" .context) .depth}}{{if $validation}}{{tabs .depth}}for _, e := range {{.target}} {
{{$validation}}
{{tabs .depth}}}{{end}}`

	enumValTmpl = `{{$depth := or (and (and (not .required) (eq .attribute.Type.Kind 4)) (add .depth 1)) .depth}}{{if not .required}}{{if eq .attribute.Type.Kind 4}}{{tabs .depth}}if {{.target}} != "" {
{{else if gt .attribute.Type.Kind 4}}{{tabs $depth}}if {{.target}} != nil {
{{end}}{{end}}{{tabs $depth}}if !({{oneof .target .values}}) {
{{tabs $depth}}	err = goa.InvalidEnumValueError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{slice .values}}, err)
{{if and (not .required) (gt .attribute.Type.Kind 3)}}{{tabs $depth}}	}
{{end}}{{tabs .depth}}}`

	patternValTmpl = `{{$depth := or (and (not .required) (add .depth 1)) .depth}}{{if not .required}}{{tabs .depth}}if {{.target}} != "" {
{{end}}{{tabs $depth}}if ok := goa.ValidatePattern(` + "`{{.pattern}}`" + `, {{.target}}); !ok {
{{tabs $depth}}	err = goa.InvalidPatternError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, ` + "`{{.pattern}}`" + `, err)
{{tabs $depth}}}{{if not .required}}
{{tabs .depth}}}{{end}}`

	formatValTmpl = `{{$depth := or (and (not .required) (add .depth 1)) .depth}}{{ if not .required}}{{tabs .depth}}if {{.target}} != "" {
{{end}}{{tabs $depth}}if err2 := goa.ValidateFormat({{constant .format}}, {{.target}}); err2 != nil {
{{tabs $depth}}		err = goa.InvalidFormatError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{constant .format}}, err2, err)
{{if not .required}}{{tabs $depth}}	}
{{end}}{{tabs .depth}}}`

	minMaxValTmpl = `{{$depth := or (and (not .required) (add .depth 1)) .depth}}{{tabs .depth}}if {{.target}} {{if .min}}<{{else}}>{{end}} {{if .min}}{{.min}}{{else}}{{.max}}{{end}} {
{{tabs $depth}}	err = goa.InvalidRangeError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{if .min}}{{.min}}, true{{else}}{{.max}}, false{{end}}, err)
{{tabs .depth}}}`

	lengthValTmpl = `{{$depth := or (and (not .required) (add .depth 1)) .depth}}{{tabs .depth}}if len({{.target}}) {{if .minLength}}<{{else}}>{{end}} {{if .minLength}}{{.minLength}}{{else}}{{.maxLength}}{{end}} {
{{tabs $depth}}	err = goa.InvalidLengthError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{if .minLength}}{{.minLength}}, true{{else}}{{.maxLength}}, false{{end}}, err)
{{tabs .depth}}}`

	requiredValTmpl = `{{$ctx := .}}{{range $r := .required}}{{$catt := index $ctx.attribute.Type.ToObject $r}}{{if eq $catt.Type.Kind 4}}{{tabs $ctx.depth}}if {{$ctx.target}}.{{goify $r true}} == "" {
{{tabs $ctx.depth}}	err = goa.MissingAttributeError(` + "`" + `{{$ctx.context}}` + "`" + `, "{{$r}}", err)
{{tabs $ctx.depth}}}{{else if gt $catt.Type.Kind 4}}{{tabs $ctx.depth}}if {{$ctx.target}}.{{goify $r true}} == nil {
{{tabs $ctx.depth}}	err = goa.MissingAttributeError(` + "`" + `{{$ctx.context}}` + "`" + `, "{{$r}}", err)
{{tabs $ctx.depth}}}{{end}}
{{end}}`
)
