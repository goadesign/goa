package main

import (
	"text/template"
	"unicode"

	"github.com/raphael/goa/design"
)

type (
	// ContextsWriter generate codes for a goa application contexts.
	ContextsWriter struct {
		*CodeWriter
		CtxTmpl     *template.Template
		CtxNewTmpl  *template.Template
		CtxRespTmpl *template.Template
		PayloadTmpl *template.Template
	}

	// ContextData contains all the information used by the template to render the contexts
	// code.
	ContextData struct {
		Name         string // e.g. "ListBottleContext"
		ResourceName string // e.g. "bottles"
		ActionName   string // e.g. "list"
		Params       *design.AttributeDefinition
		Payload      *design.AttributeDefinition
		Headers      *design.AttributeDefinition
		Responses    []*design.ResponseDefinition
	}
)

// NewContextsWriter returns a contexts code writer.
// Contexts provide the glue between the underlying request data and the user controller.
func NewContextsWriter(filename string) (*ContextsWriter, error) {
	cw, err := NewCodeWriter(filename)
	if err != nil {
		return nil, err
	}
	ctxTmpl, err := template.New("context").Funcs(cw.FuncMap).Parse(ctxT)
	if err != nil {
		return nil, err
	}
	ctxNewTmpl, err := template.New("new").Funcs(
		cw.FuncMap).Funcs(template.FuncMap{
		"newCoerceData":  newCoerceData,
		"elemType":       elemType,
		"arrayAttribute": arrayAttribute,
	}).Parse(ctxNewT)
	if err != nil {
		return nil, err
	}
	ctxRespTmpl, err := template.New("response").Funcs(cw.FuncMap).Parse(ctxRespT)
	if err != nil {
		return nil, err
	}
	payloadTmpl, err := template.New("payload").Funcs(cw.FuncMap).Parse(payloadT)
	if err != nil {
		return nil, err
	}
	w := ContextsWriter{
		CodeWriter:  cw,
		CtxTmpl:     ctxTmpl,
		CtxNewTmpl:  ctxNewTmpl,
		CtxRespTmpl: ctxRespTmpl,
		PayloadTmpl: payloadTmpl,
	}
	return &w, nil
}

// Write writes the code for the context types to outdir.
func (w *ContextsWriter) Write(data *ContextData) error {
	if err := w.CtxTmpl.Execute(w.writer, data); err != nil {
		return err
	}
	if err := w.CtxNewTmpl.Execute(w.writer, data); err != nil {
		return err
	}
	return nil
}

// newCoerceData is a helper function that creates a map that can be given to the "Coerce"
// template.
func newCoerceData(name string, att *design.AttributeDefinition, target string) map[string]interface{} {
	return map[string]interface{}{
		"Name":      goify(name),
		"Attribute": att,
		"Target":    target,
	}
}

// goify makes a valid go identifier out of any string.
// It does that by replacing any non letter and non digit character with "_" and by making sure
// the first character is a letter or "_".
func goify(str string) string {
	if str == "" {
		return "_"
	}
	if !unicode.IsLetter(str[0]) && str[0] != '_' {
		res = "_" + str[0:1]
	} else {
		res = str[0:1]
	}
	i := 1
	for i < len(str) {
		if !unicode.IsLetter(str[i]) && !unicode.IsDigit(str[i]) {
			res += "_"
		} else {
			res += str[i : i+1]
		}
		i++
	}
	return res
}

// elemType returns the go type name of the array elements.
func elemType(a *design.AttributeDefinition) string {
	return a.Type.(*design.Array).ElemType.Type.Name()
}

// arrayAttribute returns the array element attribute definition.
func arrayAttribute(a *design.AttributeDefinition) string {
	return a.Type.(*design.Array).ElemType
}

const (
	ctxT = `// {{.Name}} provides the {{.ResourceName}} {{.ActionName}} action context
type {{.Name}} struct {
	*goa.Context
	{{range .Params.Type.(Object)}}{{camelize .Name}} {{.Type.Name}}
{{end}} }
`

	ctxNewT = `
{{define "Coerce"}}{{if eq .Attribute.Type.Kind 1}}{{/* BooleanType */}}	if {{.Name}}, err := strconv.ParseBool(raw{{camelize .Name}}); err == nil {
		{{.Target}} = {{.Name}}
	} else {
		err = goa.InvalidParamValue("{{.Name}}", raw{{camelize .Name}}, "boolean", err)
	}
	{{end}}{{if eq .Attribute.Type.Kind 2}}{{/* IntegerType */}}if {{.Name}}, err := strconv.Atoi(raw{{camelize .Name}}); err == nil {
		{{.Target}} = int({{.Name}})
	} else {
		err = goa.InvalidParamValue("{{.Name}}", raw{{camelize .Name}}, "integer", err)
	}
	{{end}}{{if eq .Attribute.Type.Kind 3}}{{/* NumberType */}}if {{.Name}}, err := strconv.ParseFloat(raw{{camelize .Name}}, 64); err == nil {
		{{.Target}} = {{.Name}}
	} else {
		err = goa.InvalidParamValue("{{.Name}}", raw{{camelize .Name}}, "number", err)
	}
	{{end}}{{if eq .Attribute.Type.Kind 4}}{{/* StringType */}}{{.Target}} = raw{{camelize .Name}}
	{{end}}{{if eq .Attribute.Type.Kind 5}}{{/* ArrayType */}}elems{{camelize .Name}} := strings.Split(raw{{camelize .Name}}, ",")
	elems{{camelize .Name}}2 := make([]{{elemType .Attribute}}, len(elems{{camelize .Name}}))
	for i, rawElem := range elems{{camelize .Name}} {
		{{template "Coerce" (newCoerceData "elem" (arrayAttribute .Attribute) (printf "elems%s2[i]" (camelize .Name)))}}
	}
	{{.Target}} = elems{{camelize .Name}}2
{{end}}// New{{.Name}} parses the incoming request URL and body, performs validations and creates the
// context used by the controller action.
func New{{.Name}}(c *goa.Context) (*{{.Name}}, error) {
	var err error
	ctx := {{.Name}}{Context: c}
	{{$params := .Params}}{{range $name, $att := $params.Type.(Object)}}raw{{camelize $name}}, {{if $params.IsRequired $name}}ok{{else}}_{{end}} := c.Get("{{$name}}")
	{{if $params.IsRequired $name}}if !ok {
		err = goa.MissingParam("$name", err)
	} else {
	{{end}}{{template "Coerce" (newCoerceData $name $att ctx.(camelize (goify $name)))}}
	{{if $params.IsRequired $name}} }
	{{end}}
	{{end}}{{/* range $params */}}{{if .Payload}}var p {{.PayloadTypeName}}
	if err := c.Bind(&p); err != nil {
		return nil, err
	}
	ctx.Payload = &p
	{{end}}return &ctx, err
}
`
	ctxRespT = `// {.Name}} builds a HTTP response with status code {{.Code}}.
func (c *{{.Context}}) {{.Name}}({{.Resource}} {{.Type}}) error {
	return c.JSON({{.Code}}, {{.Resource}})
}
`
	payloadT = `// {{.Name}} is the {{.ResourceName}} {{.ActionName}} action payload.
type {{.Name}} struct {
	{{$name, $val := range .Type.Object}}{{camelize $name}} {{$val.Type.Name}} ` + "`" + `json:"{{.Name}}{{if not .Required}},omitempty{{end}}"` + "`" + `
{{end}} }
`
)
