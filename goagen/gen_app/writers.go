package genapp

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"sort"

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
		CtrlTmpl    *template.Template
		MountTmpl   *template.Template
		handleCORST *template.Template
	}

	// SecurityWriter generate code for action-level security handlers.
	SecurityWriter struct {
		*codegen.SourceFile
		SecurityTmpl *template.Template
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
		DefaultPkg   string
		Security     *design.SecurityDefinition
	}

	// ControllerTemplateData contains the information required to generate an action handler.
	ControllerTemplateData struct {
		API            *design.APIDefinition          // API definition
		Resource       string                         // Lower case plural resource name, e.g. "bottles"
		Actions        []map[string]interface{}       // Array of actions, each action has keys "Name", "Routes", "Context" and "Unmarshal"
		FileServers    []*design.FileServerDefinition // File servers
		Encoders       []*EncoderTemplateData         // Encoder data
		Decoders       []*EncoderTemplateData         // Decoder data
		Origins        []*design.CORSDefinition       // CORS policies
		PreflightPaths []string
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

	// EncoderTemplateData contains the data needed to render the registration code for a single
	// encoder or decoder package.
	EncoderTemplateData struct {
		// PackagePath is the Go package path to the package implmenting the encoder/decoder.
		PackagePath string
		// PackageName is the name of the Go package implementing the encoder/decoder.
		PackageName string
		// Function is the name of the package function implementing the decoder/encoder factory.
		Function string
		// MIMETypes is the list of supported MIME types.
		MIMETypes []string
		// Default is true if this encoder/decoder should be set as the default.
		Default bool
	}
)

// IsPathParam returns true if the given parameter name corresponds to a path parameter for all
// the context action routes. Such parameter is required but does not need to be validated as
// httptreemux takes care of that.
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

// IterateResponses iterates through the responses sorted by status code.
func (c *ContextTemplateData) IterateResponses(it func(*design.ResponseDefinition) error) error {
	m := make(map[int]*design.ResponseDefinition, len(c.Responses))
	var s []int
	for _, resp := range c.Responses {
		status := resp.Status
		m[status] = resp
		s = append(s, status)
	}
	sort.Ints(s)
	for _, status := range s {
		if err := it(m[status]); err != nil {
			return err
		}
	}
	return nil
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
	fn = template.FuncMap{
		"project": func(mt *design.MediaTypeDefinition, v string) *design.MediaTypeDefinition {
			p, _, _ := mt.Project(v)
			return p
		},
	}
	data.IterateResponses(func(resp *design.ResponseDefinition) error {
		respData := map[string]interface{}{
			"Context":  data,
			"Response": resp,
		}
		if resp.Type != nil {
			respData["Type"] = resp.Type
			if err := w.ExecuteTemplate("response", ctxTRespT, fn, respData); err != nil {
				return err
			}
		} else if mt := design.Design.MediaTypeWithIdentifier(resp.MediaType); mt != nil {
			respData["MediaType"] = mt
			fn["respName"] = func(resp *design.ResponseDefinition, view string) string {
				if view == "default" {
					return codegen.Goify(resp.Name, true)
				}
				base := fmt.Sprintf("%s%s", resp.Name, strings.Title(view))
				return codegen.Goify(base, true)
			}
			if err := w.ExecuteTemplate("response", ctxMTRespT, fn, respData); err != nil {
				return err
			}
		} else {
			if err := w.ExecuteTemplate("response", ctxNoMTRespT, fn, respData); err != nil {
				return err
			}
		}
		return nil
	})
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

// WriteInitService writes the initService function
func (w *ControllersWriter) WriteInitService(encoders, decoders []*EncoderTemplateData) error {
	ctx := map[string]interface{}{
		"API":      design.Design,
		"Encoders": encoders,
		"Decoders": decoders,
	}
	if err := w.ExecuteTemplate("service", serviceT, nil, ctx); err != nil {
		return err
	}
	return nil
}

// Execute writes the handlers GoGenerator
func (w *ControllersWriter) Execute(data []*ControllerTemplateData) error {
	if len(data) == 0 {
		return nil
	}
	for _, d := range data {
		if err := w.ExecuteTemplate("controller", ctrlT, nil, d); err != nil {
			return err
		}
		if err := w.ExecuteTemplate("mount", mountT, nil, d); err != nil {
			return err
		}
		if len(d.Origins) > 0 {
			if err := w.ExecuteTemplate("handleCORS", handleCORST, nil, d); err != nil {
				return err
			}
		}
		if err := w.ExecuteTemplate("unmarshal", unmarshalT, nil, d); err != nil {
			return err
		}
	}
	return nil
}

// NewSecurityWriter returns a security functionality code writer.
// Those functionalities are there to support action-middleware related to security.
func NewSecurityWriter(filename string) (*SecurityWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &SecurityWriter{SourceFile: file}, nil
}

// Execute adds the different security schemes and middleware supporting functions.
func (w *SecurityWriter) Execute(schemes []*design.SecuritySchemeDefinition) error {
	return w.ExecuteTemplate("security_schemes", securitySchemesT, nil, schemes)
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
func (w *MediaTypesWriter) Execute(mt *design.MediaTypeDefinition) error {
	var mLinks *design.UserTypeDefinition
	viewMT := mt
	err := mt.IterateViews(func(view *design.ViewDefinition) error {
		p, links, err := mt.Project(view.Name)
		if mLinks == nil {
			mLinks = links
		}
		if err != nil {
			return err
		}
		viewMT = p
		if err := w.ExecuteTemplate("mediatype", mediaTypeT, nil, viewMT); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if mLinks != nil {
		if err := w.ExecuteTemplate("mediatypelink", mediaTypeLinkT, nil, mLinks); err != nil {
			return err
		}
	}
	return nil
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
func (w *UserTypesWriter) Execute(t *design.UserTypeDefinition) error {
	return w.ExecuteTemplate("types", userTypeT, nil, t)
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

// arrayAttribute returns the array element attribute definition.
func arrayAttribute(a *design.AttributeDefinition) *design.AttributeDefinition {
	return a.Type.(*design.Array).ElemType
}

const (
	// ctxT generates the code for the context data type.
	// template input: *ContextTemplateData
	ctxT = `// {{ .Name }} provides the {{ .ResourceName }} {{ .ActionName }} action context.
type {{ .Name }} struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
{{ if .Params }}{{ range $name, $att := .Params.Type.ToObject }}{{/*
*/}}	{{ goify $name true }} {{ if and $att.Type.IsPrimitive ($.Params.IsPrimitivePointer $name) }}*{{ end }}{{ gotyperef .Type nil 0 false }}
{{ end }}{{ end }}{{ if .Payload }}	Payload {{ gotyperef .Payload nil 0 false }}
{{ end }}}
`
	// coerceT generates the code that coerces the generic deserialized
	// data to the actual type.
	// template input: map[string]interface{} as returned by newCoerceData
	coerceT = `{{ if eq .Attribute.Type.Kind 1 }}{{/*

*/}}{{/* BooleanType */}}{{/*
*/}}{{ $varName := or (and (not .Pointer) .VarName) tempvar }}{{/*
*/}}{{ tabs .Depth }}if {{ .VarName }}, err2 := strconv.ParseBool(raw{{ goify .Name true }}); err2 == nil {
{{ if .Pointer }}{{ tabs .Depth }}	{{ $varName }} := &{{ .VarName }}
{{ end }}{{ tabs .Depth }}	{{ .Pkg }} = {{ $varName }}
{{ tabs .Depth }}} else {
{{ tabs .Depth }}	err = goa.MergeErrors(err, goa.InvalidParamTypeError("{{ .Name }}", raw{{ goify .Name true }}, "boolean"))
{{ tabs .Depth }}}
{{ end }}{{ if eq .Attribute.Type.Kind 2 }}{{/*

*/}}{{/* IntegerType */}}{{/*
*/}}{{ $tmp := tempvar }}{{/*
*/}}{{ tabs .Depth }}if {{ .VarName }}, err2 := strconv.Atoi(raw{{ goify .Name true }}); err2 == nil {
{{ if .Pointer }}{{ $tmp2 := tempvar }}{{ tabs .Depth }}	{{ $tmp2 }} := {{ .VarName }}
{{ tabs .Depth }}	{{ $tmp }} := &{{ $tmp2 }}
{{ tabs .Depth }}	{{ .Pkg }} = {{ $tmp }}
{{ else }}{{ tabs .Depth }}	{{ .Pkg }} = {{ .VarName }}
{{ end }}{{ tabs .Depth }}} else {
{{ tabs .Depth }}	err = goa.MergeErrors(err, goa.InvalidParamTypeError("{{ .Name }}", raw{{ goify .Name true }}, "integer"))
{{ tabs .Depth }}}
{{ end }}{{ if eq .Attribute.Type.Kind 3 }}{{/*

*/}}{{/* NumberType */}}{{/*
*/}}{{ $varName := or (and (not .Pointer) .VarName) tempvar }}{{/*
*/}}{{ tabs .Depth }}if {{ .VarName }}, err2 := strconv.ParseFloat(raw{{ goify .Name true }}, 64); err2 == nil {
{{ if .Pointer }}{{ tabs .Depth }}	{{ $varName }} := &{{ .VarName }}
{{ end }}{{ tabs .Depth }}	{{ .Pkg }} = {{ $varName }}
{{ tabs .Depth }}} else {
{{ tabs .Depth }}	err = goa.MergeErrors(err, goa.InvalidParamTypeError("{{ .Name }}", raw{{ goify .Name true }}, "number"))
{{ tabs .Depth }}}
{{ end }}{{ if eq .Attribute.Type.Kind 4 }}{{/*

*/}}{{/* StringType */}}{{/*
*/}}{{ tabs .Depth }}{{ .Pkg }} = {{ if .Pointer }}&{{ end }}raw{{ goify .Name true }}
{{ end }}{{ if eq .Attribute.Type.Kind 5 }}{{/*

*/}}{{/* DateTimeType */}}{{/*
*/}}{{ $varName := or (and (not .Pointer) .VarName) tempvar }}{{/*
*/}}{{ tabs .Depth }}if {{ .VarName }}, err2 := time.Parse("RFC3339", raw{{ goify .Name true }}); err2 == nil {
{{ if .Pointer }}{{ tabs .Depth }}	{{ $varName }} := &{{ .VarName }}
{{ end }}{{ tabs .Depth }}	{{ .Pkg }} = {{ $varName }}
{{ tabs .Depth }}} else {
{{ tabs .Depth }}	err = goa.MergeErrors(err, goa.InvalidParamTypeError("{{ .Name }}", raw{{ goify .Name true }}, "datetime"))
{{ tabs .Depth }}}
{{ end }}{{ if eq .Attribute.Type.Kind 6 }}{{/*

*/}}{{/* UUIDType */}}{{/*
*/}}{{ $varName := or (and (not .Pointer) .VarName) tempvar }}{{/*
*/}}{{ tabs .Depth }}if {{ .VarName }}, err2 := uuid.FromString(raw{{ goify .Name true }}); err2 == nil {
{{ if .Pointer }}{{ tabs .Depth }}	{{ $varName }} := &{{ .VarName }}
{{ end }}{{ tabs .Depth }}	{{ .Pkg }} = {{ $varName }}
{{ tabs .Depth }}} else {
{{ tabs .Depth }}	err = goa.MergeErrors(err, goa.InvalidParamTypeError("{{ .Name }}", raw{{ goify .Name true }}, "uuid"))
{{ tabs .Depth }}}
{{ end }}{{ if eq .Attribute.Type.Kind 7 }}{{/*

*/}}{{/* AnyType */}}{{/*
*/}}{{ if .Pointer }}{{ $tmp := tempvar }}{{ tabs .Depth }}{{ $tmp }} := interface{}(raw{{ goify .Name true }})
{{ tabs .Depth }}{{ .Pkg }} = &{{ $tmp }}
{{ else }}{{ tabs .Depth }}{{ .Pkg }} = raw{{ goify .Name true }}
{{ end }}{{ end }}{{ if eq .Attribute.Type.Kind 8 }}{{/*

*/}}{{/* ArrayType */}}{{/*
*/}}{{ tabs .Depth }}elems{{ goify .Name true }} := strings.Split(raw{{ goify .Name true }}, ",")
{{ if eq (arrayAttribute .Attribute).Type.Kind 4 }}{{ tabs .Depth }}{{ .Pkg }} = elems{{ goify .Name true }}
{{ else }}{{ tabs .Depth }}elems{{ goify .Name true }}2 := make({{ gotyperef .Attribute.Type nil .Depth false }}, len(elems{{ goify .Name true }}))
{{ tabs .Depth }}for i, rawElem := range elems{{ goify .Name true }} {
{{ template "Coerce" (newCoerceData "elem" (arrayAttribute .Attribute) false (printf "elems%s2[i]" (goify .Name true)) (add .Depth 1)) }}{{ tabs .Depth }}}
{{ tabs .Depth }}{{ .Pkg }} = elems{{ goify .Name true }}2
{{ end }}{{ end }}`

	// ctxNewT generates the code for the context factory method.
	// template input: *ContextTemplateData
	ctxNewT = `{{ define "Coerce" }}` + coerceT + `{{ end }}` + `
// New{{ goify .Name true }} parses the incoming request URL and body, performs validations and creates the
// context used by the {{ .ResourceName }} controller {{ .ActionName }} action.
func New{{ .Name }}(ctx context.Context, service *goa.Service) (*{{ .Name }}, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := {{ .Name }}{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
{{ if .Headers }}{{ $headers := .Headers }}{{ range $name, $att := $headers.Type.ToObject }}	raw{{ goify $name true }} := req.Header.Get("{{ $name }}")
{{ if $headers.IsRequired $name }}	if raw{{ goify $name true }} == "" {
		err = goa.MergeErrors(err, goa.MissingHeaderError("{{ $name }}"))
	} else {
{{ else }}	if raw{{ goify $name true }} != "" {
{{ end }}{{ $validation := validationChecker $att ($headers.IsNonZero $name) ($headers.IsRequired $name) ($headers.HasDefaultValue $name) (printf "raw%s" (goify $name true)) $name 2 false }}{{/*
*/}}{{ if $validation }}{{ $validation }}
{{ end }}	}
{{ end }}{{ end }}{{/*
*/}}{{ if.Params }}{{ range $name, $att := .Params.Type.ToObject }}	param{{ goify $name true }} := req.Params["{{ $name }}"]
{{ $mustValidate := $.MustValidate $name }}{{ if $mustValidate }}	if len(param{{ goify $name true }}) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("{{ $name }}"))
	} else {
{{ else }}	if len(param{{ goify $name true }}) > 0 {
{{ end }}{{/* if $mustValidate */}}{{ if $att.Type.IsArray }}		var params {{ gotypedef $att 2 true false }}
		for _, raw{{ goify $name true}} := range param{{ goify $name true}} {
{{ template "Coerce" (newCoerceData $name $att ($.Params.IsPrimitivePointer $name) "params" 3) }}{{/*
*/}}			{{ printf "rctx.%s" (goify $name true) }} = append({{ printf "rctx.%s" (goify $name true) }}, params...)
		}
{{ else }}		raw{{ goify $name true}} := param{{ goify $name true}}[0]
{{ template "Coerce" (newCoerceData $name $att ($.Params.IsPrimitivePointer $name) (printf "rctx.%s" (goify $name true)) 2) }}{{ end }}{{/*
*/}}{{ $validation := validationChecker $att ($.Params.IsNonZero $name) ($.Params.IsRequired $name) ($.Params.HasDefaultValue $name) (printf "rctx.%s" (goify $name true)) $name 2 false }}{{/*
*/}}{{ if $validation }}{{ $validation }}
{{ end }}	}
{{ end }}{{ end }}{{/* if .Params */}}	return &rctx, err
}
`

	// ctxMTRespT generates the response helpers for responses with media types.
	// template input: map[string]interface{}
	ctxMTRespT = `{{ $ctx := .Context }}{{ $resp := .Response }}{{ $mt := .MediaType }}{{/*
*/}}{{ range $name, $view := $mt.Views }}{{ if not (eq $name "link") }}{{ $projected := project $mt $name }}
// {{ respName $resp $name }} sends a HTTP response with status code {{ $resp.Status }}.
func (ctx *{{ $ctx.Name }}) {{ respName $resp $name }}(r {{ gotyperef $projected $projected.AllRequired 0 false }}) error {
	ctx.ResponseData.Header().Set("Content-Type", "{{ $resp.MediaType }}")
	return ctx.Service.Send(ctx.Context, {{ $resp.Status }}, r)
}
{{ end }}{{ end }}
`

	// ctxTRespT generates the response helpers for responses with overridden types.
	// template input: map[string]interface{}
	ctxTRespT = `// {{ goify .Response.Name true }} sends a HTTP response with status code {{ .Response.Status }}.
func (ctx *{{ .Context.Name }}) {{ goify .Response.Name true }}(r {{ gotyperef .Type nil 0 false }}) error {
	ctx.ResponseData.Header().Set("Content-Type", "{{ .Response.MediaType }}")
	return ctx.Service.Send(ctx.Context, {{ .Response.Status }}, r)
}
`

	// ctxNoMTRespT generates the response helpers for responses with no known media type.
	// template input: *ContextTemplateData
	ctxNoMTRespT = `
// {{ goify .Response.Name true }} sends a HTTP response with status code {{ .Response.Status }}.
func (ctx *{{ .Context.Name }}) {{ goify .Response.Name true }}({{ if .Response.MediaType }}resp []byte{{ end }}) error {
{{ if .Response.MediaType }}	ctx.ResponseData.Header().Set("Content-Type", "{{ .Response.MediaType }}")
{{ end }}	ctx.ResponseData.WriteHeader({{ .Response.Status }}){{ if .Response.MediaType }}
	_, err := ctx.ResponseData.Write(resp)
	return err{{ else }}
	return nil{{ end }}
}
`

	// payloadT generates the payload type definition GoGenerator
	// template input: *ContextTemplateData
	payloadT = `{{ $payload := .Payload }}{{ if .Payload.IsObject }}// {{ gotypename .Payload nil 0 true }} is the {{ .ResourceName }} {{ .ActionName }} action payload.{{/*
*/}}{{ $privateTypeName := gotypename .Payload nil 1 true }}
type {{ $privateTypeName }} {{ gotypedef .Payload 0 true true }}

{{ $assignment := recursiveFinalizer .Payload.AttributeDefinition "payload" 1 }}{{ if $assignment }}// Finalize sets the default values defined in the design.
func (payload {{ gotyperef .Payload .Payload.AllRequired 0 true }}) Finalize() {
{{ $assignment }}
}{{ end }}

{{ $validation := recursiveValidate .Payload.AttributeDefinition false false false "payload" "raw" 1 true }}{{ if $validation }}// Validate runs the validation rules defined in the design.
func (payload {{ gotyperef .Payload .Payload.AllRequired 0 true }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
{{ $typeName := gotypename .Payload .Payload.AllRequired 1 false }}
// Publicize creates {{ $typeName }} from {{ $privateTypeName }}
func (payload {{ gotyperef .Payload .Payload.AllRequired 0 true }}) Publicize() {{ gotyperef .Payload .Payload.AllRequired 0 false }} {
	var pub {{ $typeName }}
	{{ recursivePublicizer .Payload.AttributeDefinition "payload" "pub" 1 }}
	return &pub
}{{ end }}

// {{ gotypename .Payload nil 0 false }} is the {{ .ResourceName }} {{ .ActionName }} action payload.
type {{ gotypename .Payload nil 1 false }} {{ gotypedef .Payload 0 true false }}

{{ $validation := recursiveValidate .Payload.AttributeDefinition false false false "payload" "raw" 1 false }}{{ if $validation }}// Validate runs the validation rules defined in the design.
func (payload {{ gotyperef .Payload .Payload.AllRequired 0 false }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
`
	// ctrlT generates the controller interface for a given resource.
	// template input: *ControllerTemplateData
	ctrlT = `// {{ .Resource }}Controller is the controller interface for the {{ .Resource }} actions.
type {{ .Resource }}Controller interface {
	goa.Muxer
{{ if .FileServers }}	goa.FileServer
{{ end }}{{ range .Actions }}	{{ .Name }}(*{{ .Context }}) error
{{ end }}}
`

	// serviceT generates the service initialization code.
	// template input: *ControllerTemplateData
	serviceT = `
// initService sets up the service encoders, decoders and mux.
func initService(service *goa.Service) {
	// Setup encoders and decoders
{{ range .Encoders }}{{/*
*/}}	service.Encoder.Register({{ .PackageName }}.{{ .Function }}, "{{ join .MIMETypes "\", \"" }}")
{{ end }}{{ range .Decoders }}{{/*
*/}}	service.Decoder.Register({{ .PackageName }}.{{ .Function }}, "{{ join .MIMETypes "\", \"" }}")
{{ end }}

	// Setup default encoder and decoder
{{ range .Encoders }}{{ if .Default }}{{/*
*/}}	service.Encoder.Register({{ .PackageName }}.{{ .Function }}, "*/*")
{{ end }}{{ end }}{{ range .Decoders }}{{ if .Default }}{{/*
*/}}	service.Decoder.Register({{ .PackageName }}.{{ .Function }}, "*/*")
{{ end }}{{ end }}}
`

	// mountT generates the code for a resource "Mount" function.
	// template input: *ControllerTemplateData
	mountT = `
// Mount{{ .Resource }}Controller "mounts" a {{ .Resource }} resource controller on the given service.
func Mount{{ .Resource }}Controller(service *goa.Service, ctrl {{ .Resource }}Controller) {
	initService(service)
	var h goa.Handler
{{ $res := .Resource }}{{ if .Origins }}{{ range .PreflightPaths }}	service.Mux.Handle("OPTIONS", "{{ . }}", cors.HandlePreflight(service.Context, handle{{ $res }}Origin))
{{ end }}{{ end }}{{ range .Actions }}{{ $action := . }}
	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rctx, err := New{{ .Context }}(ctx, service)
		if err != nil {
			return err
		}
{{ if .Payload }}if rawPayload := goa.ContextRequest(ctx).Payload; rawPayload != nil {
			rctx.Payload = rawPayload.({{ gotyperef .Payload nil 1 false }})
		}
		{{ end }}		return ctrl.{{ .Name }}(rctx)
	}
{{ if $.Origins }}	h = handle{{ $res }}Origin(h)
{{ end }}{{ if .Security }}	h = handleSecurity({{ printf "%q" .Security.Scheme.SchemeName }}, h{{ range .Security.Scopes }}, {{ printf "%q" . }}{{ end }})
{{ end }}{{ range .Routes }}	service.Mux.Handle("{{ .Verb }}", {{ printf "%q" .FullPath }}, ctrl.MuxHandler({{ printf "%q" $action.Name }}, h, {{ if $action.Payload }}{{ $action.Unmarshal }}{{ else }}nil{{ end }}))
	service.LogInfo("mount", "ctrl", {{ printf "%q" $res }}, "action", {{ printf "%q" $action.Name }}, "route", {{ printf "%q" (printf "%s %s" .Verb .FullPath) }}{{ with $action.Security }}, "security", {{ printf "%q" .Scheme.SchemeName }}{{ end }})
{{ end }}{{ end }}{{ range .FileServers }}
	h = ctrl.FileHandler("{{ .RequestPath }}", "{{ .FilePath }}")
{{ if $.Origins }}	h = handle{{ $res }}Origin(h)
{{ end }}{{ if .Security }}	h = handleSecurity({{ printf "%q" .Security.Scheme.SchemeName }}, h{{ range .Security.Scopes }}, {{ printf "%q" . }}{{ end }})
{{ end }}	service.Mux.Handle("GET", "{{ .RequestPath }}", ctrl.MuxHandler("serve", h, nil))
	service.LogInfo("mount", "ctrl", {{ printf "%q" $res }}, "files", {{ printf "%q" .FilePath }}, "route", {{ printf "%q" (printf "GET %s" .RequestPath) }}{{ with .Security }}, "security", {{ printf "%q" .Scheme.SchemeName }}{{ end }})
{{ end }}}
`

	// handleCORST generates the code that checks whether a CORS request is authorized
	// template input: *ControllerTemplateData
	handleCORST = `// handle{{ .Resource }}Origin applies the CORS response headers corresponding to the origin.
func handle{{ .Resource }}Origin(h goa.Handler) goa.Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		origin := req.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			return h(ctx, rw, req)
		}
{{ range $policy := .Origins }}		if cors.MatchOrigin(origin, {{ printf "%q" $policy.Origin }}) {
			ctx = goa.WithLogContext(ctx, "origin", origin)
			rw.Header().Set("Access-Control-Allow-Origin", "{{ $policy.Origin }}")
{{ if not (eq $policy.Origin "*") }}			rw.Header().Set("Vary", "Origin")
{{ end }}{{ if $policy.Exposed }}			rw.Header().Set("Access-Control-Expose-Headers", "{{ join $policy.Exposed ", " }}")
{{ end }}{{ if gt $policy.MaxAge 0 }}			rw.Header().Set("Access-Control-Max-Age", "{{ $policy.MaxAge }}")
{{ end }}			rw.Header().Set("Access-Control-Allow-Credentials", "{{ $policy.Credentials }}")
			if acrm := req.Header.Get("Access-Control-Request-Method"); acrm != "" {
				// We are handling a preflight request
{{ if $policy.Methods }}				rw.Header().Set("Access-Control-Allow-Methods", "{{ join $policy.Methods ", " }}")
{{ end }}{{ if $policy.Headers }}				rw.Header().Set("Access-Control-Allow-Headers", "{{ join $policy.Headers ", " }}")
{{ end }}			}
			return h(ctx, rw, req)
		}
{{ end }}
		return h(ctx, rw, req)
	}
}
`

	// unmarshalT generates the code for an action payload unmarshal function.
	// template input: *ControllerTemplateData
	unmarshalT = `{{ range .Actions }}{{ if .Payload }}
// {{ .Unmarshal }} unmarshals the request body into the context request data Payload field.
func {{ .Unmarshal }}(ctx context.Context, service *goa.Service, req *http.Request) error {
	{{ if .Payload.IsObject }}payload := &{{ gotypename .Payload nil 1 true }}{}
	if err := service.DecodeRequest(req, payload); err != nil {
		return err
	}{{ $assignment := recursiveFinalizer .Payload.AttributeDefinition "payload" 1 }}{{ if $assignment }}
	payload.Finalize(){{ end }}{{ else }}var payload {{ gotypename .Payload nil 1 false }}
	if err := service.DecodeRequest(req, &payload); err != nil {
		return err
	}{{ end }}{{ $validation := recursiveValidate .Payload.AttributeDefinition false false false "payload" "raw" 1 false }}{{ if $validation }}
	if err := payload.Validate(); err != nil {
		return err
	}{{ end }}
	goa.ContextRequest(ctx).Payload = payload{{ if .Payload.IsObject }}.Publicize(){{ end }}
	return nil
}
{{ end }}
{{ end }}`

	// resourceT generates the code for a resource.
	// template input: *ResourceData
	resourceT = `{{ if .CanonicalTemplate }}// {{ .Name }}Href returns the resource href.
func {{ .Name }}Href({{ if .CanonicalParams }}{{ join .CanonicalParams ", " }} interface{}{{ end }}) string {
	return fmt.Sprintf("{{ .CanonicalTemplate }}", {{ join .CanonicalParams ", " }})
}
{{ end }}`

	// mediaTypeT generates the code for a media type.
	// template input: MediaTypeTemplateData
	mediaTypeT = `// {{ gotypedesc . true }}
//
// Identifier: {{ .Identifier }}{{ $typeName := gotypename . .AllRequired 0 false }}
type {{ $typeName }} {{ gotypedef . 0 true false }}

{{ $validation := recursiveValidate .AttributeDefinition false false false "mt" "response" 1 false }}{{ if $validation }}// Validate validates the {{$typeName}} media type instance.
func (mt {{ gotyperef . .AllRequired 0 false }}) Validate() (err error) {
{{ $validation }}
	return
}
{{ end }}
`

	// mediaTypeLinkT generates the code for a media type link.
	// template input: MediaTypeLinkTemplateData
	mediaTypeLinkT = `// {{ gotypedesc . true }}{{ $typeName := gotypename . .AllRequired 0 false }}
type {{ $typeName }} {{ gotypedef . 0 true false }}
{{ $validation := recursiveValidate .AttributeDefinition false false false "ut" "response" 1 false }}{{ if $validation }}// Validate validates the {{$typeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 false }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
`

	// userTypeT generates the code for a user type.
	// template input: UserTypeTemplateData
	userTypeT = `// {{ gotypedesc . false }}{{ $privateTypeName := gotypename . .AllRequired 0 true }}
type {{ $privateTypeName }} {{ gotypedef . 0 true true }}
{{ $assignment := recursiveFinalizer .AttributeDefinition "ut" 1 }}{{ if $assignment }}// Finalize sets the default values for {{$privateTypeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 true }}) Finalize() {
{{ $assignment }}
}{{ end }}
{{ $validation := recursiveValidate .AttributeDefinition false false false "ut" "response" 1 true }}{{ if $validation }}// Validate validates the {{$privateTypeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 true }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
{{ $typeName := gotypename . .AllRequired 0 false }}
// Publicize creates {{ $typeName }} from {{ $privateTypeName }}
func (ut {{ gotyperef . .AllRequired 0 true }}) Publicize() {{ gotyperef . .AllRequired 0 false }} {
	var pub {{ gotypename . .AllRequired 0 false }}
	{{ recursivePublicizer .AttributeDefinition "ut" "pub" 1 }}
	return &pub
}

// {{ gotypedesc . true }}
type {{ $typeName }} {{ gotypedef . 0 true false }}
{{ $validation := recursiveValidate .AttributeDefinition false false false "ut" "response" 1 false }}{{ if $validation }}// Validate validates the {{$typeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 false }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
`

	// securitySchemesT generates the code for the security module.
	// template input: []*design.SecuritySchemeDefinition
	securitySchemesT = `
type (
	// Private type used to store auth handler info in request context
	authMiddlewareKey string
)

{{ range . }}
{{ $funcName := printf "Use%s" (goify .SchemeName true) }}// {{ $funcName }} mounts the {{ .SchemeName }} auth middleware onto the service.
func {{ $funcName }}(service *goa.Service, middleware goa.Middleware) {
	service.Context = context.WithValue(service.Context, authMiddlewareKey({{ printf "%q" .SchemeName }}), middleware)
}

{{ $funcName := printf "New%sSecurity" (goify .SchemeName true) }}// {{ $funcName }} creates a {{ .SchemeName }} security definition.
func {{ $funcName }}() *goa.{{ .Context }} {
	def := goa.{{ .Context }}{
{{ if eq .Context "APIKeySecurity" }}{{/*
*/}}		In:   {{ if eq .In "header" }}goa.LocHeader{{ else }}goa.LocQuery{{ end }},
		Name: {{ printf "%q" .Name }},
{{ else if eq .Context "OAuth2Security" }}{{/*
*/}}		Flow:             {{ printf "%q" .Flow }},
		TokenURL:         {{ printf "%q" .TokenURL }},
		AuthorizationURL: {{ printf "%q" .AuthorizationURL }},{{ with .Scopes }}
		Scopes: map[string]string{
{{ range $k, $v := . }}			{{ printf "%q" $k }}: {{ printf "%q" $v }},
{{ end }}{{/*
*/}}		},{{ end }}{{/*
*/}}{{ else if eq .Context "BasicAuthSecurity" }}{{/*
*/}}{{ else if eq .Context "JWTSecurity" }}{{/*
*/}}		In:   {{ if eq .In "header" }}goa.LocHeader{{ else }}goa.LocQuery{{ end }},
		Name:             {{ printf "%q" .Name }},
		TokenURL:         {{ printf "%q" .TokenURL }},{{ with .Scopes }}
		Scopes: map[string]string{
{{ range $k, $v := . }}			{{ printf "%q" $k }}: {{ printf "%q" $v }},
{{ end }}{{/*
*/}}		},{{ end }}
{{ end }}{{/*
*/}}	}
{{ if .Description }} def.Description = {{ printf "%q" .Description }}
{{ end }}	return &def
}

{{ end }}// handleSecurity creates a handler that runs the auth middleware for the security scheme.
func handleSecurity(schemeName string, h goa.Handler, scopes ...string) goa.Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		scheme := ctx.Value(authMiddlewareKey(schemeName))
		am, ok := scheme.(goa.Middleware)
		if !ok {
			return goa.NoAuthMiddleware(schemeName)
		}
		ctx = goa.WithRequiredScopes(ctx, scopes)
		return am(h)(ctx, rw, req)
	}
}
`
)
