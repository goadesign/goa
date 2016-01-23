package genapp

import (
	"regexp"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

// WildcardRegex is the regex used to capture path parameters.
var WildcardRegex = regexp.MustCompile("(?:[^/]*/:([^/]+))+")

type (
	// ContextsWriter generate codes for a goa application contexts.
	ContextsWriter struct {
		*codegen.SourceFile
		CtxTmpl     *template.Template
		CtxNewTmpl  *template.Template
		CtxRespTmpl *template.Template
		PayloadTmpl *template.Template
	}

	// ControllersWriter generate code for a goa application handlers.
	// Handlers receive a HTTP request, create the action context, call the action code and send the
	// resulting HTTP response.
	ControllersWriter struct {
		*codegen.SourceFile
		CtrlTmpl      *template.Template
		MountTmpl     *template.Template
		UnmarshalTmpl *template.Template
	}

	// ResourcesWriter generate code for a goa application resources.
	// Resources are data structures initialized by the application handlers and passed to controller
	// actions.
	ResourcesWriter struct {
		*codegen.SourceFile
		ResourceTmpl *template.Template
	}

	// MediaTypesWriter generate code for a goa application media types.
	// Media types are data structures used to render the response bodies.
	MediaTypesWriter struct {
		*codegen.SourceFile
		MediaTypeTmpl *template.Template
	}

	// UserTypesWriter generate code for a goa application user types.
	// User types are data structures defined in the DSL with "Type".
	UserTypesWriter struct {
		*codegen.SourceFile
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
		DefaultPkg   string
	}

	// MediaTypeTemplateData contains all the information used by the template to redner the
	// media types code.
	MediaTypeTemplateData struct {
		MediaType  *design.MediaTypeDefinition
		Versioned  bool
		DefaultPkg string
	}

	// UserTypeTemplateData contains all the information used by the template to redner the
	// media types code.
	UserTypeTemplateData struct {
		UserType   *design.UserTypeDefinition
		Versioned  bool
		DefaultPkg string
	}

	// ControllerTemplateData contains the information required to generate an action handler.
	ControllerTemplateData struct {
		Resource   string                          // Lower case plural resource name, e.g. "bottles"
		Actions    []map[string]interface{}        // Array of actions, each action has keys "Name", "Routes", "Context" and "Unmarshal"
		Version    *design.APIVersionDefinition    // Controller API version
		EncoderMap map[string]*EncoderTemplateData // Encoder data indexed by package path
		DecoderMap map[string]*EncoderTemplateData // Decoder data indexed by package path
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

	// EncoderTemplateData containes the data needed to render the registration code for a single
	// encoder or decoder package.
	EncoderTemplateData struct {
		// PackagePath is the Go package path to the package implmenting the encoder / decoder.
		PackagePath string
		// PackageName is the name of the Go package implementing the encoder / decoder.
		PackageName string
		// Factory is the name of the package variable implementing the decoder / encoder factory.
		Factory string
		// MIMETypes is the list of supported MIME types.
		MIMETypes []string
		// Default is true if this encoder / decoder should be set as the default.
		Default bool
	}
)

// Versioned returns true if the context was built from an API version.
func (c *ContextTemplateData) Versioned() bool {
	return !c.Version.IsDefault()
}

// IsPathParam returns true if the given parameter name corresponds to a path parameter for all
// the context action routes. Such parameter is required but does not need to be validated as
// httprouter takes care of that.
func (c *ContextTemplateData) IsPathParam(param string) bool {
	params := c.Params
	pp := false
	if params.Type.IsObject() {
		for _, r := range c.Routes {
			pp = false
			for _, p := range r.Params(c.Version) {
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

// NewContextsWriter returns a contexts code writer.
// Contexts provide the glue between the underlying request data and the user controller.
func NewContextsWriter(filename string) (*ContextsWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &ContextsWriter{SourceFile: file}, nil
}

// Execute writes the code for the context types to the writer.
func (w *ContextsWriter) Execute(data *ContextTemplateData) error {
	if err := w.ExecuteTemplate("context", ctxT, nil, data); err != nil {
		return err
	}
	fn := template.FuncMap{
		"newCoerceData":  newCoerceData,
		"arrayAttribute": arrayAttribute,
	}
	if err := w.ExecuteTemplate("new", ctxNewT, fn, data); err != nil {
		return err
	}
	if data.Payload != nil {
		if err := w.ExecuteTemplate("payload", payloadT, nil, data); err != nil {
			return err
		}
	}
	if len(data.Responses) > 0 {
		if err := w.ExecuteTemplate("response", ctxRespT, nil, data); err != nil {
			return err
		}
	}
	return nil
}

// NewControllersWriter returns a handlers code writer.
// Handlers provide the glue between the underlying request data and the user controller.
func NewControllersWriter(filename string) (*ControllersWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &ControllersWriter{SourceFile: file}, nil
}

// Execute writes the handlers GoGenerator
func (w *ControllersWriter) Execute(data []*ControllerTemplateData) error {
	for _, d := range data {
		if err := w.ExecuteTemplate("controller", ctrlT, nil, d); err != nil {
			return err
		}
		if err := w.ExecuteTemplate("mount", mountT, nil, d); err != nil {
			return err
		}
		if err := w.ExecuteTemplate("unmarshal", unmarshalT, nil, d); err != nil {
			return err
		}
	}
	return nil
}

// NewResourcesWriter returns a contexts code writer.
// Resources provide the glue between the underlying request data and the user controller.
func NewResourcesWriter(filename string) (*ResourcesWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &ResourcesWriter{SourceFile: file}, nil
}

// Execute writes the code for the context types to the writer.
func (w *ResourcesWriter) Execute(data *ResourceData) error {
	return w.ExecuteTemplate("resource", resourceT, nil, data)
}

// NewMediaTypesWriter returns a contexts code writer.
// Media types contain the data used to render response bodies.
func NewMediaTypesWriter(filename string) (*MediaTypesWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &MediaTypesWriter{SourceFile: file}, nil
}

// Execute writes the code for the context types to the writer.
func (w *MediaTypesWriter) Execute(data *MediaTypeTemplateData) error {
	fn := template.FuncMap{"newDumpData": newDumpData}
	return w.ExecuteTemplate("new", mediaTypeT, fn, data)
}

// NewUserTypesWriter returns a contexts code writer.
// User types contain custom data structured defined in the DSL with "Type".
func NewUserTypesWriter(filename string) (*UserTypesWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &UserTypesWriter{SourceFile: file}, nil
}

// Execute writes the code for the context types to the writer.
func (w *UserTypesWriter) Execute(data *UserTypeTemplateData) error {
	return w.ExecuteTemplate("types", userTypeT, nil, data)
}

// newCoerceData is a helper function that creates a map that can be given to the "Coerce" template.
func newCoerceData(name string, att *design.AttributeDefinition, pointer bool, pkg string, depth int) map[string]interface{} {
	return map[string]interface{}{
		"Name":      name,
		"VarName":   codegen.Goify(name, false),
		"Pointer":   pointer,
		"Attribute": att,
		"Pkg":       pkg,
		"Depth":     depth,
	}
}

// newDumpData is a helper function that creates a map that can be given to the "Dump" template.
func newDumpData(mt *design.MediaTypeDefinition, versioned bool, defaultPkg, context, source, target, view string) map[string]interface{} {
	return map[string]interface{}{
		"MediaType":  mt,
		"Context":    context,
		"Source":     source,
		"Target":     target,
		"View":       view,
		"Versioned":  versioned,
		"DefaultPkg": defaultPkg,
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
{{if .Params}}{{$ctx := .}}{{range $name, $att := .Params.Type.ToObject}}{{/*
*/}}	{{goify $name true}} {{if and $att.Type.IsPrimitive ($ctx.Params.IsPrimitivePointer $name)}}*{{end}}{{gotyperef .Type nil 0}}
{{end}}{{end}}{{if .Payload}}	Payload {{gotyperef .Payload nil 0}}
{{end}}}
`
	// coerceT generates the code that coerces the generic deserialized
	// data to the actual type.
	// template input: map[string]interface{} as returned by newCoerceData
	coerceT = `{{if eq .Attribute.Type.Kind 1}}{{/*

*/}}{{/* BooleanType */}}{{/*
*/}}{{$varName := or (and (not .Pointer) .VarName) tempvar}}{{/*
*/}}{{tabs .Depth}}if {{.VarName}}, err2 := strconv.ParseBool(raw{{goify .Name true}}); err2 == nil {
{{if .Pointer}}{{tabs .Depth}}	{{$varName}} := &{{.VarName}}
{{end}}{{tabs .Depth}}	{{.Pkg}} = {{$varName}}
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{goify .Name true}}, "boolean", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 2}}{{/*

*/}}{{/* IntegerType */}}{{/*
*/}}{{$tmp := tempvar}}{{/*
*/}}{{tabs .Depth}}if {{.VarName}}, err2 := strconv.Atoi(raw{{goify .Name true}}); err2 == nil {
{{if .Pointer}}{{$tmp2 := tempvar}}{{tabs .Depth}}	{{$tmp2}} := int({{.VarName}})
{{tabs .Depth}}	{{$tmp}} := &{{$tmp2}}
{{tabs .Depth}}	{{.Pkg}} = {{$tmp}}
{{else}}{{tabs .Depth}}	{{.Pkg}} = int({{.VarName}})
{{end}}{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{goify .Name true}}, "integer", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 3}}{{/*

*/}}{{/* NumberType */}}{{/*
*/}}{{$varName := or (and (not .Pointer) .VarName) tempvar}}{{/*
*/}}{{tabs .Depth}}if {{.VarName}}, err2 := strconv.ParseFloat(raw{{goify .Name true}}, 64); err2 == nil {
{{if .Pointer}}{{tabs .Depth}}	{{$varName}} := &{{.VarName}}
{{end}}{{tabs .Depth}}	{{.Pkg}} = {{$varName}}
{{tabs .Depth}}} else {
{{tabs .Depth}}	err = goa.InvalidParamTypeError("{{.Name}}", raw{{goify .Name true}}, "number", err)
{{tabs .Depth}}}
{{end}}{{if eq .Attribute.Type.Kind 4}}{{/*

*/}}{{/* StringType */}}{{/*
*/}}{{tabs .Depth}}{{.Pkg}} = {{if .Pointer}}&{{end}}raw{{goify .Name true}}
{{end}}{{if eq .Attribute.Type.Kind 5}}{{/*

*/}}{{/* AnyType */}}{{/*
*/}}{{tabs .Depth}}{{.Pkg}} = {{if .Pointer}}&{{end}}raw{{goify .Name true}}
{{end}}{{if eq .Attribute.Type.Kind 6}}{{/*

*/}}{{/* ArrayType */}}{{/*
*/}}{{tabs .Depth}}elems{{goify .Name true}} := strings.Split(raw{{goify .Name true}}, ",")
{{if eq (arrayAttribute .Attribute).Type.Kind 4}}{{tabs .Depth}}{{.Pkg}} = elems{{goify .Name true}}
{{else}}{{tabs .Depth}}elems{{goify .Name true}}2 := make({{gotyperef .Attribute.Type nil .Depth}}, len(elems{{goify .Name true}}))
{{tabs .Depth}}for i, rawElem := range elems{{goify .Name true}} {
{{template "Coerce" (newCoerceData "elem" (arrayAttribute .Attribute) false (printf "elems%s2[i]" (goify .Name true)) (add .Depth 1))}}{{tabs .Depth}}}
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
{{end}}{{if.Params}}{{$ctx := .}}{{range $name, $att := .Params.Type.ToObject}}	raw{{goify $name true}} := c.Get("{{$name}}")
{{$mustValidate := $ctx.MustValidate $name}}{{if $mustValidate}}	if raw{{goify $name true}} == "" {
		err = goa.MissingParamError("{{$name}}", err)
	} else {
{{else}}	if raw{{goify $name true}} != "" {
{{end}}{{template "Coerce" (newCoerceData $name $att ($ctx.Params.IsPrimitivePointer $name) (printf "ctx.%s" (goify $name true)) 2)}}{{/*
*/}}{{$validation := validationChecker $att ($ctx.Params.IsNonZero $name) ($ctx.Params.IsRequired $name) (printf "ctx.%s" (goify $name true)) $name 2}}{{/*
*/}}{{if $validation}}{{$validation}}
{{end}}	}
{{end}}{{end}}{{/* if .Params */}}	return &ctx, err
}
`
	// ctxRespT generates response helper methods GoGenerator
	// template input: *ContextTemplateData
	ctxRespT = `{{$ctx := .}}{{range .Responses}}{{$mt := $ctx.API.MediaTypeWithIdentifier .MediaType}}{{/*
*/}}// {{goify .Name true}} sends a HTTP response with status code {{.Status}}.
func (ctx *{{$ctx.Name}}) {{goify .Name true}}({{/*
*/}}{{if $mt}}resp {{gopkgtyperef $mt $mt.AllRequired $ctx.Versioned $ctx.DefaultPkg 0}}{{if gt (len $mt.ComputeViews) 1}}, view {{gopkgtypename $mt $mt.AllRequired $ctx.Versioned $ctx.DefaultPkg 0}}ViewEnum{{end}}{{/*
*/}}{{else if .MediaType}}resp []byte{{end}}) error {
{{if $mt}}	r, err := resp.Dump({{if gt (len $mt.ComputeViews) 1}}view{{end}})
	if err != nil {
		return fmt.Errorf("invalid response: %s", err)
	}
	ctx.Header().Set("Content-Type", "{{$mt.Identifier}}; charset=utf-8")
	return ctx.Respond({{.Status}}, r){{else}}	return ctx.RespondBytes({{.Status}}, {{if and (not $mt) .MediaType}}resp{{else}}nil{{end}}){{end}}
}

{{end}}`

	// payloadT generates the payload type definition GoGenerator
	// template input: *ContextTemplateData
	payloadT = `{{$payload := .Payload}}// {{gotypename .Payload nil 0}} is the {{.ResourceName}} {{.ActionName}} action payload.
type {{gotypename .Payload nil 1}} {{gotypedef .Payload .Versioned .DefaultPkg 0 true}}

{{$validation := recursiveValidate .Payload.AttributeDefinition false false "payload" "raw" 1}}{{if $validation}}// Validate runs the validation rules defined in the design.
func (payload {{gotyperef .Payload .Payload.AllRequired 0}}) Validate() (err error) {
{{$validation}}
       return
}{{end}}
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
func Mount{{.Resource}}Controller(service goa.Service, ctrl {{.Resource}}Controller) {
	// Setup encoders and decoders. This is idempotent and is done by each MountXXX function.
{{$ctx := .}}{{range .EncoderMap}}{{$tmp := tempvar}}{{/*
*/}}	{{$tmp}} := {{.PackageName}}.{{.Factory}}()
	service.{{if not $ctx.Version.IsDefault}}Version("{{$ctx.Version.Version}}").{{end}}SetEncoder({{$tmp}}, {{.Default}}, "{{join .MIMETypes "\", \""}}")
{{end}}{{range .DecoderMap}}{{$tmp := tempvar}}{{/*
*/}}	{{$tmp}} := {{.PackageName}}.{{.Factory}}()
	service.{{if not $ctx.Version.IsDefault}}Version("{{$ctx.Version.Version}}").{{end}}SetDecoder({{$tmp}}, {{.Default}}, "{{join .MIMETypes "\", \""}}")
{{end}}
	// Setup endpoint handler
	var h goa.Handler
	mux := service.{{if not .Version.IsDefault}}Version("{{.Version.Version}}").{{end}}ServeMux()
{{$res := .Resource}}{{$ver := .Version}}{{range .Actions}}{{$action := .}}	h = func(c *goa.Context) error {
		ctx, err := New{{.Context}}(c)
{{if not $ver.IsDefault}}		ctx.Version = service.Version("{{$ver.Version}}")
{{end}}{{if .Payload}}		ctx.Payload = ctx.RawPayload().(*{{gotypename .Payload nil 1}})
{{end}}		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.{{.Name}}(ctx)
	}
{{range .Routes}}	mux.Handle("{{.Verb}}", "{{.FullPath $ver}}", ctrl.HandleFunc("{{$action.Name}}", h, {{if $action.Payload}}{{$action.Unmarshal}}{{else}}nil{{end}}))
	service.Info("mount", "ctrl", "{{$res}}",{{if not $ver.IsDefault}} "version", "{{$ver.Version}}",{{end}} "action", "{{$action.Name}}", "route", "{{.Verb}} {{.FullPath $ver}}")
{{end}}{{end}}}
`

	// unmarshalT generates the code for an action payload unmarshal function.
	// template input: *ControllerTemplateData
	unmarshalT = `{{range .Actions}}{{if .Payload}}
// {{.Unmarshal}} unmarshals the request body.
func {{.Unmarshal}}(ctx *goa.Context) error {
	payload := &{{gotypename .Payload nil 1}}{}
	if err := ctx.Service().DecodeRequest(ctx, payload); err != nil {
		return err
	}{{$validation := recursiveValidate .Payload.AttributeDefinition false false "payload" "raw" 1}}{{if $validation}}
	if err := payload.Validate(); err != nil {
		return err
	}{{end}}
	ctx.SetPayload(payload)
	return nil
}
{{end}}
{{end}}`

	// resourceT generates the code for a resource.
	// template input: *ResourceData
	resourceT = `{{if .CanonicalTemplate}}// {{.Name}}Href returns the resource href.
func {{.Name}}Href({{if .CanonicalParams}}{{join .CanonicalParams ", "}} interface{}{{end}}) string {
	return fmt.Sprintf("{{.CanonicalTemplate}}", {{join .CanonicalParams ", "}})
}
{{end}}`

	// mediaTypeT generates the code for a media type.
	// template input: MediaTypeTemplateData
	mediaTypeT = `{{define "Dump"}}` + dumpT + `{{end}}` + `// {{if .MediaType.Description}}{{.MediaType.Description}}{{else}}{{gotypename .MediaType .MediaType.AllRequired 0}} media type{{end}}
// Identifier: {{.MediaType.Identifier}}{{$typeName := gotypename .MediaType .MediaType.AllRequired 0}}
type {{$typeName}} {{gotypedef .MediaType .Versioned .DefaultPkg 0 true}}{{$computedViews := .MediaType.ComputeViews}}{{if gt (len $computedViews) 1}}

// {{$typeName}} views
type {{$typeName}}ViewEnum string

const (
{{range $name, $view := $computedViews}}// {{if .Description}}{{.Description}}{{else}}{{$typeName}} {{.Name}} view{{end}}
	{{$typeName}}{{goify .Name true}}View {{$typeName}}ViewEnum = "{{.Name}}"
{{end}}){{end}}

// Dump produces raw data from an instance of {{$typeName}} running all the
// validations. See Load{{$typeName}} for the definition of raw data.
func (mt {{gotyperef .MediaType .MediaType.AllRequired 0}}) Dump({{if gt (len $computedViews) 1}}view {{$typeName}}ViewEnum{{end}}) (res {{gonative .MediaType}}, err error) {
{{$mt := .MediaType}}{{$ctx := .}}{{if gt (len $computedViews) 1}}{{range $computedViews}}	if view == {{gotypename $mt $mt.AllRequired 0}}{{goify .Name true}}View {
		{{template "Dump" (newDumpData $mt $ctx.Versioned $ctx.DefaultPkg (printf "%s view" .Name) "mt" "res" .Name)}}
	}
{{end}}{{else}}{{range $mt.ComputeViews}}{{template "Dump" (newDumpData $mt $ctx.Versioned $ctx.DefaultPkg (printf "%s view" .Name) "mt" "res" .Name)}}{{/* ranges over the one element */}}
{{end}}{{end}}	return
}

{{$validation := recursiveValidate .MediaType.AttributeDefinition false false "mt" "response" 1}}{{if $validation}}// Validate validates the media type instance.
func (mt {{gotyperef .MediaType .MediaType.AllRequired 0}}) Validate() (err error) {
{{$validation}}
	return
}
{{end}}{{range $computedViews}}
{{mediaTypeMarshalerImpl $mt $ctx.Versioned $ctx.DefaultPkg .Name}}
{{end}}
`

	// dumpT generates the code for dumping a media type or media type collection element.
	dumpT = `{{if .MediaType.IsArray}}	{{.Target}} = make({{gonative .MediaType}}, len({{.Source}}))
{{$tmp := tempvar}}	for i, {{$tmp}} := range {{.Source}} {
{{$tmpel := tempvar}}		var {{$tmpel}} {{gonative .MediaType.ToArray.ElemType.Type}}
		{{template "Dump" (newDumpData .MediaType.ToArray.ElemType.Type .Versioned .DefaultPkg (printf "%s[*]" .Context) $tmp $tmpel .View)}}
		{{.Target}}[i] = {{$tmpel}}
	}{{else}}{{typeMarshaler .MediaType .Versioned .DefaultPkg .Context .Source .Target .View}}{{end}}`

	// userTypeT generates the code for a user type.
	// template input: UserTypeTemplateData
	userTypeT = `// {{if .UserType.Description}}{{.UserType.Description}}{{else}}{{gotypename .UserType .UserType.AllRequired 0}} type{{end}}
type {{gotypename .UserType .UserType.AllRequired 0}} {{gotypedef .UserType .Versioned .DefaultPkg 0 true}}

{{$validation := recursiveValidate .UserType.AttributeDefinition false false "ut" "response" 1}}{{if $validation}}// Validate validates the type instance.
func (ut {{gotyperef .UserType .UserType.AllRequired 0}}) Validate() (err error) {
{{$validation}}
	return
}

{{end}}{{userTypeMarshalerImpl .UserType .Versioned .DefaultPkg}}
`
)
