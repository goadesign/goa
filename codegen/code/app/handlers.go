package app

import (
	"text/template"

	"github.com/raphael/goa/codegen/code"
)

// HandlersWriter generate code for a goa application handlers.
// Handlers receive a HTTP request, create the action context, call the action code and send the
// resulting HTTP response.
type HandlersWriter struct {
	*code.Writer
	InitTmpl    *template.Template
	HandlerTmpl *template.Template
}

// ActionHandlerTemplateData contains the information required to generate an action handler.
type ActionHandlerTemplateData struct {
	Resource string // Lower case plural resource name, e.g. "bottles"
	Action   string // Lower case action name, e.g. "list"
	Verb     string // HTTP method, e.g. "GET"
	Path     string // Action request path, e.g. "/accounts/:accountID/bottles"
	Name     string // Handler function name, e.g. "listBottlesHandler"
	Context  string // Name of corresponding context data structure e.g. "ListBottleContext"
}

// NewHandlersWriter returns a handlers code writer.
// Handlers provide the glue between the underlying request data and the user controller.
func NewHandlersWriter(filename string) (*HandlersWriter, error) {
	cw, err := code.NewWriter(filename)
	if err != nil {
		return nil, err
	}
	initTmpl, err := template.New("init").Funcs(cw.FuncMap).Parse(initT)
	if err != nil {
		return nil, err
	}
	handlerTmpl, err := template.New("handler").Funcs(cw.FuncMap).Parse(handlerT)
	if err != nil {
		return nil, err
	}
	w := HandlersWriter{
		Writer:      cw,
		HandlerTmpl: handlerTmpl,
		InitTmpl:    initTmpl,
	}
	return &w, nil
}

// Write writes the handlers code.
func (w *HandlersWriter) Write(data []*ActionHandlerTemplateData) error {
	if len(data) > 0 {
		if err := w.InitTmpl.Execute(w.Writer, data); err != nil {
			return err
		}
	}
	for _, h := range data {
		if err := w.HandlerTmpl.Execute(w.Writer, h); err != nil {
			return err
		}
	}
	return nil
}

const (
	initT = `func init() {
	goa.RegisterHandlers(
{{range .}}		&goa.HandlerFactory{"{{.Resource}}", "{{.Action}}", "{{.Verb}}", "{{.Path}}", {{.Name}}},
{{end}}	)
}
`

	handlerT = `
func {{.Name}}(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *{{.Context}}) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action {{.Action}} {{.Resource}}, expected 'func(c *{{.Context}}) error'")
	}
	return func(c goa.Context) error {
		ctx, err := New{{.Context}}(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`
)
