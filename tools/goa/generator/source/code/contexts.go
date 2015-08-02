package main

import (
	"text/template"

	"github.com/raphael/goa/design"
)

type (
	ContextsWriter struct {
		*CodeWriter
		CtxTmpl     *template.Template
		CtxNewTmpl  *template.Template
		CtxRespTmpl *template.Template
		PayloadTmpl *template.Template
	}

	ContextData struct {
		Name          string // e.g. "ListBottleContext"
		TargetPackage string // e.g. "github.com/goauser/repo"
		Params        *design.AttributeDefinition
		Payload       *design.AttributeDefinition
		Headers       *design.AttributeDefinition
		Responses     []*design.ResponseDefinition
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
	ctxNewTmpl, err := template.New("new").Funcs(cw.FuncMap).Parse(ctxNewT)
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
	if err := w.WriteHeader(data.TargetPackage); err != nil {
		return err
	}
	return nil
}

const (
	ctxT = `// {{.Name}} provides the {{.ResourceName}} {{.ActionName}} action context
type {{.Name}} struct {
	*goa.Context
	{{range .Params}}{{camelize .Name}} {{.Type.Name}}
{{end}} }
`

	ctxNewT = `
// New{{.Name}} parses the incoming request URL and body, performs validations and creates the
// context used by the controller action.
func New{{.Name}}(c *goa.Context) (*{{.Name}}, error) {
	var err error
	ctx := {{.Name}}{Context: c}
	{{range .Params}}{{initContext .}}
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
