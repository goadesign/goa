package genapp

import (
	"regexp"
	"strings"
	"text/template"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
)

// WildcardRegex is the regex used to capture path parameters.
var WildcardRegex = regexp.MustCompile("(?:[^/]*/:([^/]+))+")

type (
	// ContextsWriter generate codes for a goa application contexts.
	ContextsWriter struct {
		*codegen.GoGenerator
		CtxTmpl        *template.Template
		CtxNewTmpl     *template.Template
		CtxRespTmpl    *template.Template
		PayloadTmpl    *template.Template
		NewPayloadTmpl *template.Template
	}

	// ControllersWriter generate code for a goa application handlers.
	// Handlers receive a HTTP request, create the action context, call the action code and send the
	// resulting HTTP response.
	ControllersWriter struct {
		*codegen.GoGenerator
		CtrlTmpl  *template.Template
		MountTmpl *template.Template
	}

	// ResourcesWriter generate code for a goa application resources.
	// Resources are data structures initialized by the application handlers and passed to controller
	// actions.
	ResourcesWriter struct {
		*codegen.GoGenerator
		ResourceTmpl *template.Template
	}

	// MediaTypesWriter generate code for a goa application media types.
	// Media types are data structures used to render the response bodies.
	MediaTypesWriter struct {
		*codegen.GoGenerator
		MediaTypeTmpl *template.Template
	}

	// UserTypesWriter generate code for a goa application user types.
	// User types are data structures defined in the DSL with "Type".
	UserTypesWriter struct {
		*codegen.GoGenerator
		UserTypeTmpl *template.Template
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
		Routes       []*design.RouteDefinition
		Responses    map[string]*design.ResponseDefinition
		API          *design.APIDefinition
		Version      *design.APIVersionDefinition
	}

	// ControllerTemplateData contains the information required to generate an action handler.
	ControllerTemplateData struct {
		Resource string                   // Lower case plural resource name, e.g. "bottles"
		Actions  []map[string]interface{} // Array of actions, each action has keys "Name", "Routes" and "Context"
		Version  string                   // Controller API version
	}

	// ResourceData contains the information required to generate the resource GoGenerator
	ResourceData struct {
		Name              string                      // Name of resource
		Identifier        string                      // Identifier of resource media type
		Description       string                      // Description of resource
		Type              *design.MediaTypeDefinition // Type of resource media type
		CanonicalTemplate string                      // CanonicalFormat represents the resource canonical path in the form of a fmt.Sprintf format.
		CanonicalParams   []string                    // CanonicalParams is the list of parameter names that appear in the resource canonical path in order.
	}
)

// IsPathParam returns true if the given parameter name corresponds to a path parameter for all
// the context action routes. Such parameter is required but does not need to be validated as
// httprouter takes care of that.
func (c *ContextTemplateData) IsPathParam(param string) bool {
	params := c.Params
	pp := false
	if params.Type.IsObject() {
		for _, r := range c.Routes {
			pp = false
			for _, p := range r.Params() {
				if p == param {
					pp = true
					break
				}
			}
			if !pp {
				break
			}
		}
	}
	return pp
}

// MustValidate returns true if code that checks for the presence of the given param must be
// generated.
func (c *ContextTemplateData) MustValidate(name string) bool {
	return c.Params.IsRequired(name) && !c.IsPathParam(name)
}

// MustSetHas returns true if the "Has" context field for the given parameter must be generated.
func (c *ContextTemplateData) MustSetHas(name string) bool {
	return !c.Params.IsRequired(name) && !c.IsPathParam(name)
}

// NewContextsWriter returns a contexts code writer.
// Contexts provide the glue between the underlying request data and the user controller.
func NewContextsWriter(filename string) (*ContextsWriter, error) {
	cw := codegen.NewGoGenerator(filename)
	funcMap := cw.FuncMap
	funcMap["gotyperef"] = codegen.GoTypeRef
	funcMap["gotypedef"] = codegen.GoTypeDef
	funcMap["goify"] = codegen.Goify
	funcMap["gotypename"] = codegen.GoTypeName
	funcMap["typeUnmarshaler"] = codegen.TypeUnmarshaler
	funcMap["userTypeUnmarshalerImpl"] = codegen.UserTypeUnmarshalerImpl
	funcMap["validationChecker"] = codegen.ValidationChecker
	funcMap["tabs"] = codegen.Tabs
	funcMap["add"] = func(a, b int) int { return a + b }
	ctxTmpl, err := template.New("context").Funcs(funcMap).Parse(ctxT)
	if err != nil {
		return nil, err
	}
	ctxNewTmpl, err := template.New("new").
		Funcs(cw.FuncMap).
		Funcs(template.FuncMap{
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
	payloadTmpl, err := template.New("payload").Funcs(funcMap).Parse(payloadT)
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
		if err := w.PayloadTmpl.Execute(w, data); err != nil {
			return err
		}
		if err := w.NewPayloadTmpl.Execute(w, data); err != nil {
			return err
		}
	}
	if len(data.Responses) > 0 {
		if err := w.CtxRespTmpl.Execute(w, data); err != nil {
			return err
		}
	}
	return nil
}

// NewControllersWriter returns a handlers code writer.
// Handlers provide the glue between the underlying request data and the user controller.
func NewControllersWriter(filename string) (*ControllersWriter, error) {
	cw := codegen.NewGoGenerator(filename)
	funcMap := cw.FuncMap
	funcMap["add"] = func(a, b int) int { return a + b }
	ctrlTmpl, err := template.New("controller").Funcs(funcMap).Parse(ctrlT)
	if err != nil {
		return nil, err
	}
	mountTmpl, err := template.New("mount").Funcs(funcMap).Parse(mountT)
	if err != nil {
		return nil, err
	}
	w := ControllersWriter{
		GoGenerator: cw,
		CtrlTmpl:    ctrlTmpl,
		MountTmpl:   mountTmpl,
	}
	return &w, nil
}

// Execute writes the handlers GoGenerator
func (w *ControllersWriter) Execute(data []*ControllerTemplateData) error {
	for _, d := range data {
		if err := w.CtrlTmpl.Execute(w, d); err != nil {
			return err
		}
		if err := w.MountTmpl.Execute(w, d); err != nil {
			return err
		}
	}
	return nil
}

// NewResourcesWriter returns a contexts code writer.
// Resources provide the glue between the underlying request data and the user controller.
func NewResourcesWriter(filename string) (*ResourcesWriter, error) {
	cw := codegen.NewGoGenerator(filename)
	funcMap := cw.FuncMap
	funcMap["join"] = strings.Join
	funcMap["goresdef"] = codegen.GoResDef
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
func (w *ResourcesWriter) Execute(data *ResourceData) error {
	return w.ResourceTmpl.Execute(w, data)
}

// NewMediaTypesWriter returns a contexts code writer.
// Media types contain the data used to render response bodies.
func NewMediaTypesWriter(filename string) (*MediaTypesWriter, error) {
	cw := codegen.NewGoGenerator(filename)
	funcMap := cw.FuncMap
	funcMap["gotypedef"] = codegen.GoTypeDef
	funcMap["gotyperef"] = codegen.GoTypeRef
	funcMap["goify"] = codegen.Goify
	funcMap["gotypename"] = codegen.GoTypeName
	funcMap["gonative"] = codegen.GoNativeType
	funcMap["typeUnmarshaler"] = codegen.TypeUnmarshaler
	funcMap["typeMarshaler"] = codegen.MediaTypeMarshaler
	funcMap["recursiveValidate"] = codegen.RecursiveChecker
	funcMap["tempvar"] = codegen.Tempvar
	funcMap["newDumpData"] = newDumpData
	funcMap["userTypeUnmarshalerImpl"] = codegen.UserTypeUnmarshalerImpl
	funcMap["mediaTypeMarshalerImpl"] = codegen.MediaTypeMarshalerImpl
	mediaTypeTmpl, err := template.New("media type").Funcs(funcMap).Parse(mediaTypeT)
	if err != nil {
		return nil, err
	}
	w := MediaTypesWriter{
		GoGenerator:   cw,
		MediaTypeTmpl: mediaTypeTmpl,
	}
	return &w, nil
}

// Execute writes the code for the context types to the writer.
func (w *MediaTypesWriter) Execute(mt *design.MediaTypeDefinition) error {
	return w.MediaTypeTmpl.Execute(w, mt)
}

// NewUserTypesWriter returns a contexts code writer.
// User types contain custom data structured defined in the DSL with "Type".
func NewUserTypesWriter(filename string) (*UserTypesWriter, error) {
	cw := codegen.NewGoGenerator(filename)
	funcMap := cw.FuncMap
	funcMap["gotypedef"] = codegen.GoTypeDef
	funcMap["gotyperef"] = codegen.GoTypeRef
	funcMap["goify"] = codegen.Goify
	funcMap["gotypename"] = codegen.GoTypeName
	funcMap["recursiveValidate"] = codegen.RecursiveChecker
	funcMap["userTypeUnmarshalerImpl"] = codegen.UserTypeUnmarshalerImpl
	funcMap["userTypeMarshalerImpl"] = codegen.UserTypeMarshalerImpl
	userTypeTmpl, err := template.New("user type").Funcs(funcMap).Parse(userTypeT)
	if err != nil {
		return nil, err
	}
	w := UserTypesWriter{
		GoGenerator:  cw,
		UserTypeTmpl: userTypeTmpl,
	}
	return &w, nil
}

// Execute writes the code for the context types to the writer.
func (w *UserTypesWriter) Execute(ut *design.UserTypeDefinition) error {
	return w.UserTypeTmpl.Execute(w, ut)
}

// newCoerceData is a helper function that creates a map that can be given to the "Coerce" template.
func newCoerceData(name string, att *design.AttributeDefinition, pkg string, depth int) map[string]interface{} {
	return map[string]interface{}{
		"Name":      name,
		"VarName":   codegen.Goify(name, false),
		"Attribute": att,
		"Pkg":       pkg,
		"Depth":     depth,
	}
}

// newDumpData is a helper function that creates a map that can be given to the "Dump" template.
func newDumpData(mt *design.MediaTypeDefinition, context, source, target, view string) map[string]interface{} {
	return map[string]interface{}{
		"MediaType": mt,
		"Context":   context,
		"Source":    source,
		"Target":    target,
		"View":      view,
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
	*goa.Context
{{if .Params}}{{$ctx := .}}{{range $name, $att := .Params.Type.ToObject}}	{{goify $name true}} {{gotyperef .Type 0}}
{{if $ctx.MustSetHas $name}}
	Has{{goify $name true}} bool
{{end}}{{end}}{{end}}{{if .Payload}}	Payload {{gotyperef .Payload 0}}
{{end}}}
`
	// coerceT generates the code that coerces the generic deserialized
	// data to the actual type.
	// template input: map[string]interface{} as returned by newCoerceData
	coerceT = `{{if eq .Attribute.Type.Kind 1}}{{/* BooleanType */}}{{tabs .Depth}}if {{.VarName}}, err2 := strconv.ParseBool(raw{{goify .Name true}}); err2 == nil {
{{tabs .Depth}}	{{.Pkg}} = {{.VarName}}
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{goify .Name true}}, "boolean", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 2}}{{/* IntegerType */}}{{tabs .Depth}}if {{.VarName}}, err2 := strconv.Atoi(raw{{goify .Name true}}); err2 == nil {
{{tabs .Depth}}	{{.Pkg}} = int({{.VarName}})
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{goify .Name true}}, "integer", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 3}}{{/* NumberType */}}{{tabs .Depth}}if {{.VarName}}, err2 := strconv.ParseFloat(raw{{goify .Name true}}, 64); err2 == nil {
{{tabs .Depth}}	{{.Pkg}} = {{.VarName}}
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{goify .Name true}}, "number", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 4}}{{/* StringType */}}{{tabs .Depth}}{{.Pkg}} = raw{{goify .Name true}}
{{end}}{{if eq .Attribute.Type.Kind 5}}{{/* AnyType */}}{{tabs .Depth}}{{.Pkg}} = raw{{goify .Name true}}
{{end}}{{if eq .Attribute.Type.Kind 6}}{{/* ArrayType */}}{{tabs .Depth}}elems{{goify .Name true}} := strings.Split(raw{{goify .Name true}}, ",")
{{if eq (arrayAttribute .Attribute).Type.Kind 4}}{{tabs .Depth}}{{.Pkg}} = elems{{goify .Name true}}
{{else}}{{tabs .Depth}}elems{{goify .Name true}}2 := make({{gotyperef .Attribute.Type .Depth}}, len(elems{{goify .Name true}}))
{{tabs .Depth}}for i, rawElem := range elems{{goify .Name true}} {
{{template "Coerce" (newCoerceData "elem" (arrayAttribute .Attribute) (printf "elems%s2[i]" (goify .Name true)) (add .Depth 1))}}{{tabs .Depth}}}
{{tabs .Depth}}{{.Pkg}} = elems{{goify .Name true}}2
{{end}}{{end}}`

	// ctxNewT generates the code for the context factory method.
	// template input: *ContextTemplateData
	ctxNewT = `{{define "Coerce"}}` + coerceT + `{{end}}` + `
// New{{goify .Name true}} parses the incoming request URL and body, performs validations and creates the
// context used by the {{.ResourceName}} controller {{.ActionName}} action.
func New{{.Name}}(c *goa.Context) (*{{.Name}}, error) {
	var err error
	ctx := {{.Name}}{Context: c}
{{if .Headers}}{{$headers := .Headers}}{{range $name, $_ := $headers.Type.ToObject}}{{if ($headers.IsRequired $name)}}	if c.Request().Header.Get("{{$name}}") == "" {
		err = goa.MissingHeaderError("{{$name}}", err)
	}{{end}}{{end}}
{{end}}{{if.Params}}{{$ctx := .}}{{range $name, $att := .Params.Type.ToObject}}	raw{{goify $name true}}, ok := c.Get("{{$name}}")
{{if ($ctx.MustValidate $name)}}	if !ok {
		err = goa.MissingParamError("{{$name}}", err)
	} else {
{{else}}	if ok {
{{end}}{{template "Coerce" (newCoerceData $name $att (printf "ctx.%s" (goify $name true)) 2)}}{{if $ctx.MustSetHas $name}}		ctx.Has{{goify $name true}} = true
{{end}}{{$validation := validationChecker $att ($ctx.Params.IsRequired $name) (printf "ctx.%s" (goify $name true)) $name 1}}{{if $validation}}{{$validation}}
{{end}}	}
{{end}}{{end}}{{/* if .Params */}}{{if .Payload}}	p, err := New{{gotypename .Payload 0}}(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
{{end}}	return &ctx, err
}

`
	// ctxRespT generates response helper methods GoGenerator
	// template input: *ContextTemplateData
	ctxRespT = `{{$ctx := .}}{{range .Responses}}// {{goify .Name true}} sends a HTTP response with status code {{.Status}}.
	func (ctx *{{$ctx.Name}}) {{goify .Name true}}({{$mt := ($ctx.API.MediaTypeWithIdentifier .MediaType)}}{{if $mt}}resp {{gotyperef $mt 0}}{{if gt (len $mt.ComputeViews) 1}}, view {{gotypename $mt 0}}ViewEnum{{end}}{{else if .MediaType}}resp []byte{{end}}) error {
{{if $mt}}	r, err := resp.Dump({{if gt (len $mt.ComputeViews) 1}}view{{end}})
	if err != nil {
		return fmt.Errorf("invalid response: %s", err)
	}
	ctx.Header().Set("Content-Type", "{{$mt.Identifier}}; charset=utf-8")
	return ctx.JSON({{.Status}}, r){{else}}return ctx.Respond({{.Status}}, {{if and (not $mt) .MediaType}}resp{{else}}nil{{end}}){{end}}
}
{{end}}`

	// payloadT generates the payload type definition GoGenerator
	// template input: *ContextTemplateData
	payloadT = `{{$payload := .Payload}}// {{gotypename .Payload 0}} is the {{.ResourceName}} {{.ActionName}} action payload.
type {{gotypename .Payload 1}} {{gotypedef .Payload 0 false false}}
`
	// newPayloadT generates the code for the payload factory method.
	// template input: *ContextTemplateData
	newPayloadT = `// New{{gotypename .Payload 0}} instantiates a {{gotypename .Payload 0}} from a raw request body.
// It validates each field and returns an error if any validation fails.
func New{{gotypename .Payload 0}}(raw interface{}) (p {{gotyperef .Payload 0}}, err error) {
{{typeUnmarshaler .Payload "payload" "raw" "p"}}
	return
}{{if (not .Payload.IsPrimitive)}}

{{userTypeUnmarshalerImpl .Payload "payload"}}{{end}}
`

	// ctrlT generates the controller interface for a given resource.
	// template input: *ControllerTemplateData
	ctrlT = `// {{.Resource}}Controller is the controller interface for the {{.Resource}} actions.
type {{.Resource}}Controller interface {
	goa.Controller
{{range .Actions}}	{{.Name}}(*{{.Context}}) error
{{end}}}
`

	// mountT generates the code for a resource "Mount" function.
	// template input: *ControllerTemplateData
	mountT = `
// Mount{{.Resource}}Controller "mounts" a {{.Resource}} resource controller on the given service.
func Mount{{.Resource}}Controller(service goa.Service, mux goa.VersionMux, ctrl {{.Resource}}Controller) {
	var h goa.Handler
{{$res := .Resource}}{{$ver := .Resource.Version}}{{range .Actions}}{{$action := .}}	h = func(c *goa.Context) error {
		ctx, err := New{{.Context}}(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.{{.Name}}(ctx)
	}
{{range .Routes}}	mux.Handle("{{.Verb}}", "{{.FullPath}}", ctrl.HandleFunc("{{$action.Name}}", h))
	service.Info("mount", "ctrl", "{{$res}}",{{if $ver}} "version", "{{$ver}}",{{end}} "action", "{{$action.Name}}", "route", "{{.Verb}} {{.FullPath}}")
{{end}}{{end}}}
`

	// resourceT generates the code for a resource.
	// template input: *ResourceData
	resourceT = `{{if .CanonicalTemplate}}// {{.Name}}Href returns the resource href.
func {{.Name}}Href({{if .CanonicalParams}}{{join .CanonicalParams ", "}} interface{}{{end}}) string {
	return fmt.Sprintf("{{.CanonicalTemplate}}", {{join .CanonicalParams ", "}})
}
{{end}}`

	// mediaTypeT generates the code for a media type.
	// template input: *design.MediaTypeDefinition
	mediaTypeT = `{{define "Dump"}}` + dumpT + `{{end}}` + `// {{if .Description}}{{.Description}}{{else}}{{gotypename . 0}} media type{{end}}
// Identifier: {{.Identifier}}{{$typeName := gotypename . 0}}
type {{$typeName}} {{gotypedef . 0 false false}}{{$computedViews := .ComputeViews}}{{if gt (len $computedViews) 1}}

// {{$typeName}} views
type {{$typeName}}ViewEnum string

const (
{{range $name, $view := $computedViews}}// {{if .Description}}{{.Description}}{{else}}{{$typeName}} {{.Name}} view{{end}}
	{{$typeName}}{{goify .Name true}}View {{$typeName}}ViewEnum = "{{.Name}}"
{{end}}){{end}}
// Load{{$typeName}} loads raw data into an instance of {{$typeName}} running all the
// validations. Raw data is defined by data that the JSON unmarshaler would create when unmarshaling
// into a variable of type interface{}. See https://golang.org/pkg/encoding/json/#Unmarshal for the
// complete list of supported data types.
func Load{{$typeName}}(raw interface{}) (res {{gotyperef . 1}}, err error) {
	{{typeUnmarshaler . "" "raw" "res"}}
	return
}

// Dump produces raw data from an instance of {{$typeName}} running all the
// validations. See Load{{$typeName}} for the definition of raw data.
func (mt {{gotyperef . 0}}) Dump({{if gt (len $computedViews) 1}}view {{$typeName}}ViewEnum{{end}}) (res {{gonative .}}, err error) {
{{$mt := .}}{{if gt (len $computedViews) 1}}{{range $computedViews}}	if view == {{gotypename $mt 0}}{{goify .Name true}}View {
		{{template "Dump" (newDumpData $mt (printf "%s view" .Name) "mt" "res" .Name)}}
	}
{{end}}{{else}}{{range $mt.ComputeViews}}{{template "Dump" (newDumpData $mt (printf "%s view" .Name) "mt" "res" .Name)}}{{/* ranges over the one element */}}
{{end}}{{end}}	return
}

{{$validation := recursiveValidate .AttributeDefinition false "mt" "response" 1}}{{if $validation}}// Validate validates the media type instance.
func (mt {{gotyperef . 0}}) Validate() (err error) {
{{$validation}}
	return
}
{{end}}{{range $computedViews}}
{{mediaTypeMarshalerImpl $mt .Name}}
{{end}}
{{userTypeUnmarshalerImpl .UserTypeDefinition "load"}}
`

	// dumpT generates the code for dumping a media type or media type collection element.
	dumpT = `{{if .MediaType.IsArray}}	{{.Target}} = make({{gonative .MediaType}}, len({{.Source}}))
{{$tmp := tempvar}}	for i, {{$tmp}} := range {{.Source}} {
{{$tmpel := tempvar}}		var {{$tmpel}} {{gonative .MediaType.ToArray.ElemType.Type}}
		{{template "Dump" (newDumpData .MediaType.ToArray.ElemType.Type (printf "%s[*]" .Context) $tmp $tmpel .View)}}
		{{.Target}}[i] = {{$tmpel}}
	}{{else}}{{typeMarshaler .MediaType .Context .Source .Target .View}}{{end}}`

	// userTypeT generates the code for a user type.
	// template input: *design.UserTypeDefinition
	userTypeT = `// {{if .Description}}{{.Description}}{{else}}{{gotypename . 0}} type{{end}}
type {{gotypename . 0}} {{gotypedef . 0 false false}}

{{$validation := recursiveValidate .AttributeDefinition false "ut" "response" 1}}{{if $validation}}// Validate validates the type instance.
func (ut {{gotyperef . 0}}) Validate() (err error) {
{{$validation}}
	return
}

{{end}}{{userTypeMarshalerImpl .}}

{{userTypeUnmarshalerImpl . "load"}}
`
)
