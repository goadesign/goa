package main

import (
	"text/template"

	"github.com/raphael/goa/design"
)

type (
	// ContextsWriter generate codes for a goa application contexts.
	ContextsWriter struct {
		*CodeWriter
		CtxTmpl        *template.Template
		CtxNewTmpl     *template.Template
		CtxRespTmpl    *template.Template
		PayloadTmpl    *template.Template
		NewPayloadTmpl *template.Template
		MediaTypeTmpl  *template.Template
	}

	// ContextData contains all the information used by the template to render the context
	// code for an action.
	ContextData struct {
		Name         string // e.g. "ListBottleContext"
		ResourceName string // e.g. "bottles"
		ActionName   string // e.g. "list"
		Params       *design.AttributeDefinition
		Payload      *design.UserTypeDefinition
		Headers      *design.AttributeDefinition
		Responses    map[string]*design.ResponseDefinition
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
	newPayloadTmpl, err := template.New("newpayload").Funcs(cw.FuncMap).Parse(newPayloadT)
	if err != nil {
		return nil, err
	}
	mediaTypeTmpl, err := template.New("mediatype").Funcs(cw.FuncMap).Parse(mediaTypeT)
	if err != nil {
		return nil, err
	}
	w := ContextsWriter{
		CodeWriter:     cw,
		CtxTmpl:        ctxTmpl,
		CtxNewTmpl:     ctxNewTmpl,
		CtxRespTmpl:    ctxRespTmpl,
		PayloadTmpl:    payloadTmpl,
		NewPayloadTmpl: newPayloadTmpl,
		MediaTypeTmpl:  mediaTypeTmpl,
	}
	return &w, nil
}

// Write writes the code for the context types to the writer.
func (w *ContextsWriter) Write(data *ContextData) error {
	w.WriteHeader("main")
	if err := w.CtxTmpl.Execute(w.writer, data); err != nil {
		return err
	}
	if err := w.CtxNewTmpl.Execute(w.writer, data); err != nil {
		return err
	}
	if data.Payload != nil {
		if _, ok := data.Payload.Type.(design.Object); ok {
			if err := w.PayloadTmpl.Execute(w.writer, data); err != nil {
				return err
			}
			if err := w.NewPayloadTmpl.Execute(w.writer, data); err != nil {
				return err
			}
		}
	}
	if err := w.MediaTypeTmpl.Execute(w.writer, data); err != nil {
		return err
	}
	if err := w.CtxRespTmpl.Execute(w.writer, data); err != nil {
		return err
	}
	return w.FormatCode()
}

// newCoerceData is a helper function that creates a map that can be given to the "Coerce"
// template.
func newCoerceData(name string, att *design.AttributeDefinition, target string) map[string]interface{} {
	return map[string]interface{}{
		"Name":      name,
		"VarName":   design.Goify(name, false),
		"Attribute": att,
		"Target":    target,
	}
}

// arrayAttribute returns the array element attribute definition.
func arrayAttribute(a *design.AttributeDefinition) *design.AttributeDefinition {
	return a.Type.(*design.Array).ElemType
}

const (
	ctxT = `// {{.Name}} provides the {{.ResourceName}} {{.ActionName}} action context
type {{.Name}} struct {
	*goa.Context
	{{if .Params}}{{range $name, $att := object .Params.Type}}{{camelize $name}} {{gotyperef .Type 0}}
{{end}}{{end}}{{if .Payload}}	payload {{gotyperef .Payload 0}}
{{end}} }
`
	coerce = `	{{if eq .Attribute.Type.Kind 1}}{{/* BooleanType */}}if {{.VarName}}, err := strconv.ParseBool(raw{{camelize .Name}}); err == nil {
		{{.Target}} = {{.VarName}}
	} else {
		err = goa.InvalidParamValue("{{.Name}}", raw{{camelize .Name}}, "boolean", err)
	}
	{{end}}{{if eq .Attribute.Type.Kind 2}}{{/* IntegerType */}}if {{.VarName}}, err := strconv.Atoi(raw{{camelize .Name}}); err == nil {
		{{.Target}} = int({{.VarName}})
	} else {
		err = goa.InvalidParamValue("{{.Name}}", raw{{camelize .Name}}, "integer", err)
	}
	{{end}}{{if eq .Attribute.Type.Kind 3}}{{/* NumberType */}}if {{.VarName}}, err := strconv.ParseFloat(raw{{camelize .Name}}, 64); err == nil {
		{{.Target}} = {{.VarName}}
	} else {
		err = goa.InvalidParamValue("{{.Name}}", raw{{camelize .Name}}, "number", err)
	}
	{{end}}{{if eq .Attribute.Type.Kind 4}}{{/* StringType */}}{{.Target}} = raw{{camelize .Name}}
	{{end}}{{if eq .Attribute.Type.Kind 5}}{{/* ArrayType */}}elems{{camelize .Name}} := strings.Split(raw{{camelize .Name}}, ",")
	{{if eq (arrayAttribute .Attribute).Type.Kind 4}}{{.Target}} = elems{{camelize .Name}}
	{{else}}elems{{camelize .Name}}2 := make({{gotyperef .Attribute.Type 1}}, len(elems{{camelize .Name}}))
	for i, rawElem := range elems{{camelize .Name}} { 
		{{template "Coerce" (newCoerceData "elem" (arrayAttribute .Attribute) (printf "elems%s2[i]" (camelize .Name)))}}}
	{{.Target}} = elems{{camelize .Name}}2
{{end}}{{end}}`

	ctxNewT = `{{define "Coerce"}}` + coerce + `{{end}}` + `
// New{{camelize .Name}} parses the incoming request URL and body, performs validations and creates the
// context used by the controller action.
func New{{camelize .Name}}(c *goa.Context) (*{{.Name}}, error) {
	var err error
	ctx := {{.Name}}{Context: c}
	{{if.Params}}{{$params := .Params}}{{range $name, $att := object $params.Type}}raw{{camelize $name}}, {{if ($params.IsRequired $name)}}ok{{else}}_{{end}} := c.Get("{{$name}}")
	{{if ($params.IsRequired $name)}}if !ok {
		err = goa.MissingParam("{{$name}}", err)
	} else {
	{{end}}{{template "Coerce" (newCoerceData $name $att (printf "ctx.%s" (camelize (goify $name true))))}}{{if ($params.IsRequired $name)}}}
	{{end}}{{end}}{{end}}{{/* if .Params */}}{{if .Payload}}var p {{gotyperef .Payload 1}}
	if err := c.Bind(&p); err != nil {
		return nil, err
	}
	ctx.Payload = &p
	{{end}}return &ctx, err
}
`
	ctxRespT = `{{$ctx := .}}{{range .Responses}}// {{.Name}} sends a HTTP response with status code {{.Status}}.
func (c *{{$ctx.Name}}) {{.Name}}({{if .MediaType}}resp *{{.MediaType.TypeName}}{{end}}) error {
	{{if .MediaType}}return c.JSON({{.Status}}, resp){{else}}return c.Respond({{.Status}}, nil){{end}}
{{end}}}
`
	payloadT = `{{$payload := .Payload}}// {{gotypename .Payload 0}} is the {{.ResourceName}} {{.ActionName}} action payload.
type {{gotypename .Payload 1}} {{gotypedef .Payload 0 true false}}
`
	newPayloadT = `// New{{gotypename .Payload 0}} instantiates a {{gotypename .Payload 0}} from a raw request body.
// It validates each field and returns an error if any validation fails.
func New{{gotypename .Payload 0}}(raw interface{}) ({{gotyperef .Payload 0}}, error) {
	var err error
	p := {{gotypename .Payload 1}}{}
	//m, ok := raw.(map[string]interface{})
	//TBD
	
	return p, err
}`

	mediaTypeT = `// New{{.ActionName}}`
)
