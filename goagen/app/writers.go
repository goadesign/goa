package app

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"bitbucket.org/pkg/inflect"
	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen"
)

// ParamsRegex is the regex used to capture path parameters.
var ParamsRegex = regexp.MustCompile("(?:[^/]*/:([^/]+))+")

type (
	// ContextsWriter generate codes for a goa application contexts.
	ContextsWriter struct {
		*goagen.GoGenerator
		CtxTmpl        *template.Template
		CtxNewTmpl     *template.Template
		CtxRespTmpl    *template.Template
		PayloadTmpl    *template.Template
		NewPayloadTmpl *template.Template
	}

	// HandlersWriter generate code for a goa application handlers.
	// Handlers receive a HTTP request, create the action context, call the action code and send the
	// resulting HTTP response.
	HandlersWriter struct {
		*goagen.GoGenerator
		InitTmpl    *template.Template
		HandlerTmpl *template.Template
	}

	// ResourcesWriter generate code for a goa application resources.
	// Resources are data structures initialized by the application handlers and passed to controller
	// actions.
	ResourcesWriter struct {
		*goagen.GoGenerator
		ResourceTmpl *template.Template
	}

	// ContextTemplateData contains all the information used by the template to render the context
	// code for an action.
	ContextTemplateData struct {
		Name         string // e.g. "ListBottleContext"
		ResourceName string // e.g. "bottles"
		ActionName   string // e.g. "list"
		Params       *design.AttributeDefinition
		Payload      *design.UserTypeDefinition
		Headers      *design.AttributeDefinition
		Responses    map[string]*design.ResponseDefinition
		MediaTypes   map[string]*design.MediaTypeDefinition
	}

	// HandlerTemplateData contains the information required to generate an action handler.
	HandlerTemplateData struct {
		Resource string // Lower case plural resource name, e.g. "bottles"
		Action   string // Lower case action name, e.g. "list"
		Verb     string // HTTP method, e.g. "GET"
		Path     string // Action request path, e.g. "/accounts/:accountID/bottles"
		Name     string // Handler function name, e.g. "listBottlesHandler"
		Context  string // Name of corresponding context data structure e.g. "ListBottleContext"
	}

	// ResourceTemplateData contains the information required to generate the resource GoGenerator
	ResourceTemplateData struct {
		Name              string                     // Name of resource
		Identifier        string                     // Identifier of resource media type
		Description       string                     // Description of resource
		Type              *design.UserTypeDefinition // Type of resource media type
		CanonicalTemplate string                     // CanonicalFormat represents the resource canonical path in the form of a fmt.Sprintf format.
		CanonicalParams   []string                   // CanonicalParams is the list of parameter names that appear in the resource canonical path in order.
	}
)

// NewContextsWriter returns a contexts code writer.
// Contexts provide the glue between the underlying request data and the user controller.
func NewContextsWriter(filename string) (*ContextsWriter, error) {
	cw := goagen.NewGoGenerator(filename)
	funcMap := cw.FuncMap
	funcMap["camelize"] = inflect.Camelize
	funcMap["gotyperef"] = goagen.GoTypeRef
	funcMap["gotypedef"] = goagen.GoTypeDef
	funcMap["goify"] = goagen.Goify
	funcMap["gotypename"] = goagen.GoTypeName
	funcMap["typeUnmarshaler"] = goagen.TypeUnmarshaler
	funcMap["validationChecker"] = goagen.ValidationChecker
	funcMap["tabs"] = goagen.Tabs
	funcMap["add"] = func(a, b int) int { return a + b }
	ctxTmpl, err := template.New("context").Funcs(funcMap).Parse(ctxT)
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
	w := ContextsWriter{
		GoGenerator:    cw,
		CtxTmpl:        ctxTmpl,
		CtxNewTmpl:     ctxNewTmpl,
		CtxRespTmpl:    ctxRespTmpl,
		PayloadTmpl:    payloadTmpl,
		NewPayloadTmpl: newPayloadTmpl,
	}
	return &w, nil
}

// Execute writes the code for the context types to the writer.
func (w *ContextsWriter) Execute(data *ContextTemplateData) error {
	if err := w.CtxTmpl.Execute(w, data); err != nil {
		return err
	}
	if err := w.CtxNewTmpl.Execute(w, data); err != nil {
		return err
	}
	if data.Payload != nil {
		if _, ok := data.Payload.Type.(design.Object); ok {
			if err := w.PayloadTmpl.Execute(w, data); err != nil {
				return err
			}
			if err := w.NewPayloadTmpl.Execute(w, data); err != nil {
				return err
			}
		}
	}
	if len(data.Responses) > 0 {
		if err := w.CtxRespTmpl.Execute(w, data); err != nil {
			return err
		}
	}
	return nil
}

// NewHandlersWriter returns a handlers code writer.
// Handlers provide the glue between the underlying request data and the user controller.
func NewHandlersWriter(filename string) (*HandlersWriter, error) {
	cw := goagen.NewGoGenerator(filename)
	initTmpl, err := template.New("init").Funcs(cw.FuncMap).Parse(initT)
	if err != nil {
		return nil, err
	}
	handlerTmpl, err := template.New("handler").Funcs(cw.FuncMap).Parse(handlerT)
	if err != nil {
		return nil, err
	}
	w := HandlersWriter{
		GoGenerator: cw,
		HandlerTmpl: handlerTmpl,
		InitTmpl:    initTmpl,
	}
	return &w, nil
}

// Execute writes the handlers GoGenerator
func (w *HandlersWriter) Execute(data []*HandlerTemplateData) error {
	if len(data) > 0 {
		if err := w.InitTmpl.Execute(w, data); err != nil {
			return err
		}
	}
	for _, h := range data {
		if err := w.HandlerTmpl.Execute(w, h); err != nil {
			return err
		}
	}
	return nil
}

// NewResourcesWriter returns a contexts code writer.
// Resources provide the glue between the underlying request data and the user controller.
func NewResourcesWriter(filename string) (*ResourcesWriter, error) {
	cw := goagen.NewGoGenerator(filename)
	funcMap := cw.FuncMap
	funcMap["join"] = strings.Join
	funcMap["gotypedef"] = goagen.GoTypeDef
	resourceTmpl, err := template.New("resource").Funcs(cw.FuncMap).Parse(resourceT)
	if err != nil {
		return nil, err
	}
	w := ResourcesWriter{
		GoGenerator:  cw,
		ResourceTmpl: resourceTmpl,
	}
	return &w, nil
}

// Execute writes the code for the context types to the writer.
func (w *ResourcesWriter) Execute(data *ResourceTemplateData) error {
	if data.Type == nil {
		return fmt.Errorf("missing resource type definition for %s", data.Name)
	}
	return w.ResourceTmpl.Execute(w, data)
}

// newCoerceData is a helper function that creates a map that can be given to the "Coerce"
// template.
func newCoerceData(name string, att *design.AttributeDefinition, target string, depth int) map[string]interface{} {
	return map[string]interface{}{
		"Name":      name,
		"VarName":   goagen.Goify(name, false),
		"Attribute": att,
		"Target":    target,
		"Depth":     depth,
	}
}

// arrayAttribute returns the array element attribute definition.
func arrayAttribute(a *design.AttributeDefinition) *design.AttributeDefinition {
	return a.Type.(*design.Array).ElemType
}

const (
	// ctxT generates the code for the context data type.
	// template input: *ContextTemplateData
	ctxT = `// {{.Name}} provides the {{.ResourceName}} {{.ActionName}} action context.
type {{.Name}} struct {
	goa.Context
{{if .Params}}{{range $name, $att := .Params.Type.ToObject}}	{{camelize $name}} {{gotyperef .Type 0}}
{{end}}{{end}}{{if .Payload}}	payload {{gotyperef .Payload 0}}
{{end}}}
`
	// coerceT generates the code that coerces the generic deserialized
	// data to the actual type.
	// template input: map[string]interface{} as returned by newCoerceData
	coerceT = `{{if eq .Attribute.Type.Kind 1}}{{/* BooleanType */}}{{tabs .Depth}}if {{.VarName}}, err := strconv.ParseBool(raw{{camelize .Name}}); err == nil {
{{tabs .Depth}}	{{.Target}} = {{.VarName}}
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{camelize .Name}}, "boolean", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 2}}{{/* IntegerType */}}{{tabs .Depth}}if {{.VarName}}, err := strconv.Atoi(raw{{camelize .Name}}); err == nil {
{{tabs .Depth}}	{{.Target}} = int({{.VarName}})
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{camelize .Name}}, "integer", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 3}}{{/* NumberType */}}{{tabs .Depth}}if {{.VarName}}, err := strconv.ParseFloat(raw{{camelize .Name}}, 64); err == nil {
{{tabs .Depth}}	{{.Target}} = {{.VarName}}
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{camelize .Name}}, "number", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 4}}{{/* StringType */}}{{tabs .Depth}}{{.Target}} = raw{{camelize .Name}}
{{end}}{{if eq .Attribute.Type.Kind 5}}{{/* ArrayType */}}{{tabs .Depth}}elems{{camelize .Name}} := strings.Split(raw{{camelize .Name}}, ",")
{{if eq (arrayAttribute .Attribute).Type.Kind 4}}{{tabs .Depth}}{{.Target}} = elems{{camelize .Name}}
{{else}}{{tabs .Depth}}elems{{camelize .Name}}2 := make({{gotyperef .Attribute.Type .Depth}}, len(elems{{camelize .Name}}))
{{tabs .Depth}}for i, rawElem := range elems{{camelize .Name}} {
{{template "Coerce" (newCoerceData "elem" (arrayAttribute .Attribute) (printf "elems%s2[i]" (camelize .Name)) (add .Depth 1))}}{{tabs .Depth}}}
{{tabs .Depth}}{{.Target}} = elems{{camelize .Name}}
{{end}}{{end}}`

	// ctxNewT generates the code for the context factory method.
	// template input: *ContextTemplateData
	ctxNewT = `{{define "Coerce"}}` + coerceT + `{{end}}` + `
// New{{camelize .Name}} parses the incoming request URL and body, performs validations and creates the
// context used by the {{.ResourceName}} controller {{.ActionName}} action.
func New{{camelize .Name}}(c goa.Context) (*{{.Name}}, error) {
	var err error
	ctx := {{.Name}}{Context: c}
	{{if .Headers}}{{$headers := .Headers}}{{range $name, $_ := $headers.Type.ToObject}}{{if ($headers.IsRequired $name)}}if c.Header.Get("{{$name}}") == "" {
		err = goa.MissingHeaderError("{{$name}}", err)
	}{{end}}{{end}}
{{end}}{{if.Params}}{{$params := .Params}}{{range $name, $att := $params.Type.ToObject}}	raw{{camelize $name}}, {{if ($params.IsRequired $name)}}ok{{else}}_{{end}} := c.Get("{{$name}}")
{{if ($params.IsRequired $name)}}	if !ok {
		err = goa.MissingParamError("{{$name}}", err)
	} else {
{{end}}{{$depth := or (and ($params.IsRequired $name) 2) 1}}{{template "Coerce" (newCoerceData $name $att (printf "ctx.%s" (camelize (goify $name true))) $depth)}}{{if ($params.IsRequired $name)}}	}
{{end}}{{validationChecker $att $name}}{{end}}{{end}}{{/* if .Params */}}{{if .Payload}}	if payload := c.Payload(); payload != nil {
		p, err := New{{gotypename .Payload 0}}(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
{{end}}	return &ctx, err
}
`
	// ctxRespT generates response helper methods GoGenerator
	// template input: *ContextTemplateData
	ctxRespT = `{{$ctx := .}}{{range .Responses}}// {{.FormatName false }} sends a HTTP response with status code {{.Status}}.
func (c *{{$ctx.Name}}) {{.FormatName false}}({{$mt := (index $ctx.MediaTypes .MediaType)}}{{if $mt}}resp *{{$mt.TypeName}}{{end}}) error {
	{{if $mt}}return c.JSON({{.Status}}, resp){{else}}return c.Respond({{.Status}}, nil){{end}}
}
{{end}}`
	// payloadT generates the payload type definition GoGenerator
	// template input: *ContextTemplateData
	payloadT = `{{$payload := .Payload}}// {{gotypename .Payload 0}} is the {{.ResourceName}} {{.ActionName}} action payload.
type {{gotypename .Payload 1}} {{gotypedef .Payload 0 true false}}
`
	// newPayloadT generates the code for the payload factory method.
	// template input: *ContextTemplateData
	newPayloadT = `// New{{gotypename .Payload 0}} instantiates a {{gotypename .Payload 0}} from a raw request body.
// It validates each field and returns an error if any validation fails.
func New{{gotypename .Payload 0}}(raw interface{}) ({{gotyperef .Payload 0}}, error) {
	var err error
	var p {{gotyperef .Payload 1}}
{{typeUnmarshaler .Payload "" "raw" "p"}}
{{validationChecker .Payload.AttributeDefinition "p"}}
	return p, err
}
`

	// initT generates the package init function which registers all
	// handlers with goa.
	// template input: *HandlerTemplateData
	initT = `
func init() {
	goa.RegisterHandlers(
{{range .}}		&goa.HandlerFactory{"{{.Resource}}", "{{.Action}}", "{{.Verb}}", "{{.Path}}", {{.Name}}},
{{end}}	)
}
`
	// handlerT generates the code for an action handler.
	// template input: *HandlerTemplateData
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
	// resourceT generates the code for a resource.
	// template input: *ResourceTemplateData
	resourceT = `// {{.Description}}
// Media type: {{.Identifier}}
type {{.Name}} {{gotypedef .Type 0 true false}}
{{if .CanonicalTemplate}}
// {{.Name}}Href returns the resource href.
func {{.Name}}Href({{join .CanonicalParams ", "}} string) string {
	return fmt.Sprintf("{{.CanonicalTemplate}}", {{join .CanonicalParams ", "}})
}
{{end}}`
)
