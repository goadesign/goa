package codegen

import (
	"fmt"
	"strings"
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
		"json":    toJSON,
		"slice":   toSlice,
		"oneof":   oneof,
	}
	if enumValT, err = template.New("enum").Funcs(fm).Parse(enumValTmpl); err != nil {
		panic(err)
	}
	if formatValT, err = template.New("format").Funcs(fm).Parse(formatValTmpl); err != nil {
		panic(err)
	}
	if minMaxValT, err = template.New("minMax").Funcs(fm).Parse(minMaxValTmpl); err != nil {
		panic(err)
	}
	if lengthValT, err = template.New("length").Funcs(fm).Parse(lengthValTmpl); err != nil {
		panic(err)
	}
}

// ValidationChecker produces Go code that runs the validation defined in the given attribute
// definition against the content of the variable named target recursively.
// context is used to keep track of recursion to produce helpful error messages in case of type
// validation error.
// The generated code assumes that there is a pre-existing "err" variable of type
// error. It initializes that variable in case a validation fails.
// TBD: decide whether context is something given to the checker or something
// that is not needed because the error message is "in context". Apply a consistent
// behavior between this and the code generation functions in types.go.
func ValidationChecker(att *design.AttributeDefinition, target string) string {
	return validationCheckerR(att, "", target, 1)
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

// oneof produces code that compares target with each element of vals and ORs
// the result, e.g. "target == 1 || target == 2".
func oneof(target string, vals []interface{}) string {
	elems := make([]string, len(vals))
	for i, v := range vals {
		elems[i] = fmt.Sprintf("%s == %#v", target, v)
	}
	return strings.Join(elems, " || ")
}

const (
	enumValTmpl = `{{tabs .depth}}if !({{oneof .target .values}}) {
{{tabs .depth}}	err = goa.InvalidEnumValueError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{slice .values}}, err)
{{tabs .depth}}}
`

	formatValTmpl = `{{tabs .depth}}if err2 := goa.ValidateFormat({{.format}}, {{.target}}); err2 != nil {
{{tabs .depth}}		err = goa.InvalidFormatError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{.format}}, err2.Error(), err)
{{tabs .depth}}}
`

	minMaxValTmpl = `{{tabs .depth}}if {{.target}} {{if .min}}<{{else}}>{{end}} {{if .min}}{{.min}}{{else}}{{.max}}{{end}} {
{{tabs .depth}}	err = goa.InvalidRangeError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{if .min}}{{.min}}, true{{else}}{{.max}}, false{{end}}, err)
{{tabs .depth}}}
`

	lengthValTmpl = `{{tabs .depth}}if len({{.target}}) {{if .minLength}}<{{else}}>{{end}} {{if .minLength}}{{.minLength}}{{else}}{{.maxLength}}{{end}} {
{{tabs .depth}}	err = goa.InvalidLengthError(` + "`" + `{{.context}}` + "`" + `, {{.target}}, {{if .minLength}}{{.minLength}}, true{{else}}{{.maxLength}}, false{{end}}, err)
{{tabs .depth}}}
`
)
