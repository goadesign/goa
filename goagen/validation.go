package goagen

import (
	"fmt"
	"text/template"

	"github.com/raphael/goa/design"
)

var (
	enumValT   *template.Template
	formatValT *template.Template
	minMaxValT *template.Template
	lengthValT *template.Template
)

//  init instantiates the templates.
func init() {
	var err error
	fm := template.FuncMap{
		"tabs":    Tabs,
		"tempvar": tempvar,
	}
	if enumValT, err = template.New("user").Funcs(fm).Parse(enumValTmpl); err != nil {
		panic(err)
	}
	if formatValT, err = template.New("user").Funcs(fm).Parse(formatValTmpl); err != nil {
		panic(err)
	}
	if minMaxValT, err = template.New("user").Funcs(fm).Parse(minMaxValTmpl); err != nil {
		panic(err)
	}
	if lengthValT, err = template.New("user").Funcs(fm).Parse(lengthValTmpl); err != nil {
		panic(err)
	}
}

// ValidationChecker produces Go code that runs the validation defined in the given attribute
// definition against the content of the variable named target recursively.
// context is used to keep track of recursion to produce helpful error messages in case of type
// validation error.
// The generated code assumes that there is a pre-existing "err" variable of type
// error. It initializes that variable in case a validation fails.
func ValidationChecker(att *design.AttributeDefinition, context, target string) string {
	return validationCheckerR(att, context, target, 1)
}
func validationCheckerR(att *design.AttributeDefinition, context, target string, depth int) string {
	data := map[string]interface{}{
		"target":  target,
		"context": context,
		"depth":   depth,
	}
	var res string
	for _, v := range att.Validations {
		switch actual := v.(type) {
		case *design.EnumValidationDefinition:
			data["values"] = actual.Values
			res += runTemplate(enumValT, data)
		case *design.FormatValidationDefinition:
			data["format"] = actual.Format
			res += runTemplate(formatValT, data)
		case *design.MinimumValidationDefinition:
			data["min"] = actual.Min
			delete(data, "max")
			res += runTemplate(minMaxValT, data)
		case *design.MaximumValidationDefinition:
			data["max"] = actual.Max
			delete(data, "min")
			res += runTemplate(minMaxValT, data)
		case *design.MinLengthValidationDefinition:
			data["minLength"] = actual.MinLength
			delete(data, "maxLength")
			res += runTemplate(lengthValT, data)
		case *design.MaxLengthValidationDefinition:
			data["maxLength"] = actual.MaxLength
			delete(data, "minLength")
			res += runTemplate(lengthValT, data)
		}
	}
	for name, catt := range att.Type.ToObject() {
		cctx := fmt.Sprintf("%s.%s", context, name)
		ctgt := fmt.Sprintf("%s.%s", target, name)
		res += validationCheckerR(catt, cctx, ctgt, depth+1)
	}
	return res
}

const (
	enumValTmpl = `{{tabs .depth}}if err == nil {
{{tabs .depth}}{{$depth := .depth}}{{$target := .target}}{{$ok := tempvar}}	{{$ok}} := false
{{$goto_marker := tempvar}}{{range .values}}{{tabs $depth}}	if {{$target}} == {{printf "%#v" .}} {
{{tabs $depth}}		{{$ok}} = true
{{tabs $depth}}		goto {{$goto_marker}}
{{tabs $depth}}	}
{{end}}	{{$goto_marker}}:
{{tabs .depth}}	if !{{$ok}} {
{{tabs .depth}}		err = goa.InvalidEnumValueError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{.values}})
{{tabs .depth}}	}
{{tabs .depth}}}
`

	formatValTmpl = `{{tabs .depth}}if err == nil {
{{tabs .depth}}	if err2 := goa.ValidateFormat({{.format}}, {{.target}}); err2 != nil {
{{tabs .depth}}		err = goa.InvalidFormatError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{.format}}, err2.Error())
{{tabs .depth}}	}
{{tabs .depth}}}
`

	minMaxValTmpl = `{{tabs .depth}}if err == nil {
{{tabs .depth}} if {{.target}} {{if .min}}<{{else}}>{{end}} {{if .min}}{{.min}}{{else}}{{.max}}{{end}} {
{{tabs .depth}}		err = goa.InvalidRangeError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{if .min}}{{.min}}, true{{else}}{{.max}}, false{{end}})
{{tabs .depth}}	}
{{tabs .depth}}}
`

	lengthValTmpl = `{{tabs .depth}}if err == nil {
{{tabs .depth}} if len({{.target}}) {{if .minLength}}<{{else}}>{{end}} {{if .minLength}}{{.minLength}}{{else}}{{.maxLength}}{{end}} {
{{tabs .depth}}		err = goa.InvalidLengthError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{if .minLength}}{{.minLength}}, true{{else}}{{.maxLength}}, false{{end}})
{{tabs .depth}}	}
{{tabs .depth}}}
`
)
