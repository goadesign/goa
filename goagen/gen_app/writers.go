package genapp

import (
	"fmt"
	"net/http"
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
		Finalizer   *codegen.Finalizer
		Validator   *codegen.Validator
	}

	// ControllersWriter generate code for a goa application handlers.
	// Handlers receive a HTTP request, create the action context, call the action code and send the
	// resulting HTTP response.
	ControllersWriter struct {
		*codegen.SourceFile
		CtrlTmpl    *template.Template
		MountTmpl   *template.Template
		handleCORST *template.Template
		Finalizer   *codegen.Finalizer
		Validator   *codegen.Validator
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
		Validator     *codegen.Validator
	}

	// UserTypesWriter generate code for a goa application user types.
	// User types are data structures defined in the DSL with "Type".
	UserTypesWriter struct {
		*codegen.SourceFile
		UserTypeTmpl *template.Template
		Finalizer    *codegen.Finalizer
		Validator    *codegen.Validator
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
		Actions        []map[string]interface{}       // Array of actions, each action has keys "Name", "DesignName", "Routes", "Context" and "Unmarshal"
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

// HasParamAndHeader returns true if the generated struct field name for the given header name
// matches the generated struct field name of a param in c.Params.
func (c *ContextTemplateData) HasParamAndHeader(name string) bool {
	if c.Params == nil || c.Headers == nil {
		return false
	}

	headerAtt := c.Headers.Type.ToObject()[name]
	headerName := codegen.GoifyAtt(headerAtt, name, true)
	for paramName, paramAtt := range c.Params.Type.ToObject() {
		paramName = codegen.GoifyAtt(paramAtt, paramName, true)
		if headerName == paramName {
			return true
		}
	}
	return false
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
	return &ContextsWriter{
		SourceFile: file,
		Finalizer:  codegen.NewFinalizer(),
		Validator:  codegen.NewValidator(),
	}, nil
}

// Execute writes the code for the context types to the writer.
func (w *ContextsWriter) Execute(data *ContextTemplateData) error {
	if err := w.ExecuteTemplate("context", ctxT, nil, data); err != nil {
		return err
	}
	fn := template.FuncMap{
		"newCoerceData":      newCoerceData,
		"arrayAttribute":     arrayAttribute,
		"printVal":           codegen.PrintVal,
		"canonicalHeaderKey": http.CanonicalHeaderKey,
		"isPathParam":        data.IsPathParam,
	}
	if err := w.ExecuteTemplate("new", ctxNewT, fn, data); err != nil {
		return err
	}
	if data.Payload != nil {
		found := false
		for _, t := range design.Design.Types {
			if t.TypeName == data.Payload.TypeName {
				found = true
				break
			}
		}
		if !found {
			fn := template.FuncMap{
				"finalizeCode":   w.Finalizer.Code,
				"validationCode": w.Validator.Code,
			}
			if err := w.ExecuteTemplate("payload", payloadT, fn, data); err != nil {
				return err
			}
		}
	}
	return data.IterateResponses(func(resp *design.ResponseDefinition) error {
		respData := map[string]interface{}{
			"Context":  data,
			"Response": resp,
		}
		var mt *design.MediaTypeDefinition
		if resp.Type != nil {
			var ok bool
			if mt, ok = resp.Type.(*design.MediaTypeDefinition); !ok {
				respData["Type"] = resp.Type
				respData["ContentType"] = resp.MediaType
				return w.ExecuteTemplate("response", ctxTRespT, nil, respData)
			}
		} else {
			mt = design.Design.MediaTypeWithIdentifier(resp.MediaType)
		}
		if mt != nil {
			var views []string
			if resp.ViewName != "" {
				views = []string{resp.ViewName}
			} else {
				views = make([]string, len(mt.Views))
				i := 0
				for name := range mt.Views {
					views[i] = name
					i++
				}
				sort.Strings(views)
			}
			for _, view := range views {
				projected, _, err := mt.Project(view)
				if err != nil {
					return err
				}
				respData["Projected"] = projected
				respData["ViewName"] = view
				respData["MediaType"] = mt
				respData["ContentType"] = mt.ContentType
				if view == "default" {
					respData["RespName"] = codegen.Goify(resp.Name, true)
				} else {
					base := fmt.Sprintf("%s%s", resp.Name, strings.Title(view))
					respData["RespName"] = codegen.Goify(base, true)
				}
				if err := w.ExecuteTemplate("response", ctxMTRespT, fn, respData); err != nil {
					return err
				}
			}
			return nil
		}
		return w.ExecuteTemplate("response", ctxNoMTRespT, nil, respData)
	})
}

// NewControllersWriter returns a handlers code writer.
// Handlers provide the glue between the underlying request data and the user controller.
func NewControllersWriter(filename string) (*ControllersWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &ControllersWriter{
		SourceFile: file,
		Finalizer:  codegen.NewFinalizer(),
		Validator:  codegen.NewValidator(),
	}, nil
}

// WriteInitService writes the initService function
func (w *ControllersWriter) WriteInitService(encoders, decoders []*EncoderTemplateData) error {
	ctx := map[string]interface{}{
		"API":      design.Design,
		"Encoders": encoders,
		"Decoders": decoders,
	}
	return w.ExecuteTemplate("service", serviceT, nil, ctx)
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
		fn := template.FuncMap{
			"finalizeCode":   w.Finalizer.Code,
			"validationCode": w.Validator.Code,
		}
		if err := w.ExecuteTemplate("unmarshal", unmarshalT, fn, d); err != nil {
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
	return &MediaTypesWriter{SourceFile: file, Validator: codegen.NewValidator()}, nil
}

// Execute writes the code for the context types to the writer.
func (w *MediaTypesWriter) Execute(mt *design.MediaTypeDefinition) error {
	var (
		mLinks *design.UserTypeDefinition
		fn     = template.FuncMap{"validationCode": w.Validator.Code}
	)
	err := mt.IterateViews(func(view *design.ViewDefinition) error {
		p, links, err := mt.Project(view.Name)
		if mLinks == nil {
			mLinks = links
		}
		if err != nil {
			return err
		}
		return w.ExecuteTemplate("mediatype", mediaTypeT, fn, p)
	})
	if err != nil {
		return err
	}
	if mLinks != nil {
		if err := w.ExecuteTemplate("mediatypelink", mediaTypeLinkT, fn, mLinks); err != nil {
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
	return &UserTypesWriter{
		SourceFile: file,
		Finalizer:  codegen.NewFinalizer(),
		Validator:  codegen.NewValidator(),
	}, nil
}

// Execute writes the code for the context types to the writer.
func (w *UserTypesWriter) Execute(t *design.UserTypeDefinition) error {
	fn := template.FuncMap{
		"finalizeCode":   w.Finalizer.Code,
		"validationCode": w.Validator.Code,
	}
	return w.ExecuteTemplate("types", userTypeT, fn, t)
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
{{ if .Headers }}{{ range $name, $att := .Headers.Type.ToObject }}{{ if not ($.HasParamAndHeader $name) }}{{/*
*/}}	{{ goifyatt $att $name true }} {{ if and $att.Type.IsPrimitive ($.Headers.IsPrimitivePointer $name) }}*{{ end }}{{ gotyperef .Type nil 0 false }}
{{ end }}{{ end }}{{ end }}{{ if .Params }}{{ range $name, $att := .Params.Type.ToObject }}{{/*
*/}}	{{ goifyatt $att $name true }} {{ if and $att.Type.IsPrimitive ($.Params.IsPrimitivePointer $name) }}*{{ end }}{{ gotyperef .Type nil 0 false }}
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
*/}}{{ tabs .Depth }}if {{ .VarName }}, err2 := time.Parse(time.RFC3339, raw{{ goify .Name true }}); err2 == nil {
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
{{ end }}{{ end }}`

	// ctxNewT generates the code for the context factory method.
	// template input: *ContextTemplateData
	ctxNewT = `{{ define "Coerce" }}` + coerceT + `{{ end }}` + `
// New{{ goify .Name true }} parses the incoming request URL and body, performs validations and creates the
// context used by the {{ .ResourceName }} controller {{ .ActionName }} action.
func New{{ .Name }}(ctx context.Context, r *http.Request, service *goa.Service) (*{{ .Name }}, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := {{ .Name }}{Context: ctx, ResponseData: resp, RequestData: req}{{/*
*/}}
{{ if .Headers }}{{ range $name, $att := .Headers.Type.ToObject }}	header{{ goify $name true }} := req.Header["{{ canonicalHeaderKey $name }}"]
{{ $mustValidate := $.Headers.IsRequired $name }}{{ if $mustValidate }}	if len(header{{ goify $name true }}) == 0 {
		err = goa.MergeErrors(err, goa.MissingHeaderError("{{ $name }}"))
	} else {
{{ else }}	if len(header{{ goify $name true }}) > 0 {
{{ end }}{{/* if $mustValidate */}}{{ if $att.Type.IsArray }}		req.Params["{{ $name }}"] = header{{ goify $name true }}
{{ if eq (arrayAttribute $att).Type.Kind 4 }}		headers := header{{ goify $name true }}
{{ else }}		headers := make({{ gotypedef $att 2 true false }}, len(header{{ goify $name true }}))
		for i, raw{{ goify $name true}} := range header{{ goify $name true}} {
{{ template "Coerce" (newCoerceData $name (arrayAttribute $att) ($.Headers.IsPrimitivePointer $name) "headers[i]" 3) }}{{/*
*/}}		}
{{ end }}		{{ printf "rctx.%s" (goifyatt $att $name true) }} = headers
{{ else }}		raw{{ goify $name true}} := header{{ goify $name true}}[0]
		req.Params["{{ $name }}"] = []string{raw{{ goify $name true }}}
{{ template "Coerce" (newCoerceData $name $att ($.Headers.IsPrimitivePointer $name) (printf "rctx.%s" (goifyatt $att $name true)) 2) }}{{ end }}{{/*
*/}}{{ $validation := validationChecker $att ($.Headers.IsNonZero $name) ($.Headers.IsRequired $name) ($.Headers.HasDefaultValue $name) (printf "rctx.%s" (goifyatt $att $name true)) $name 2 false }}{{/*
*/}}{{ if $validation }}{{ $validation }}
{{ end }}	}
{{ end }}{{ end }}{{/* if .Headers }}{{/*

*/}}{{ if .Params }}{{ range $name, $att := .Params.Type.ToObject }}{{/*
*/}}	param{{ goify $name true }} := req.Params["{{ $name }}"]
{{ $mustValidate := $.MustValidate $name }}{{ if $mustValidate }}	if len(param{{ goify $name true }}) == 0 {
		{{ if $.Params.HasDefaultValue $name }}{{printf "rctx.%s" (goifyatt $att $name true) }} = {{ printVal $att.Type $att.DefaultValue }}{{else}}{{/*
*/}}err = goa.MergeErrors(err, goa.MissingParamError("{{ $name }}")){{end}}
	} else {
{{ else }}{{ if $.Params.HasDefaultValue $name }}	if len(param{{ goify $name true }}) == 0 {
		{{printf "rctx.%s" (goifyatt $att $name true) }} = {{ printVal $att.Type $att.DefaultValue }}
	} else {
{{ else }}	if len(param{{ goify $name true }}) > 0 {
{{ end }}{{ end }}{{/* if $mustValidate */}}{{ if $att.Type.IsArray }}{{ if eq (arrayAttribute $att).Type.Kind 4 }}		params := param{{ goify $name true }}
{{ else }}		params := make({{ gotypedef $att 2 true false }}, len(param{{ goify $name true }}))
		for i, raw{{ goify $name true}} := range param{{ goify $name true}} {
{{ template "Coerce" (newCoerceData $name (arrayAttribute $att) ($.Params.IsPrimitivePointer $name) "params[i]" 3) }}{{/*
*/}}		}
{{ end }}		{{ printf "rctx.%s" (goifyatt $att $name true) }} = params
{{ else }}		raw{{ goify $name true}} := param{{ goify $name true}}[0]
{{ template "Coerce" (newCoerceData $name $att ($.Params.IsPrimitivePointer $name) (printf "rctx.%s" (goifyatt $att $name true)) 2) }}{{ end }}{{/*
*/}}{{ $validation := validationChecker $att ($.Params.IsNonZero $name) ($.Params.IsRequired $name) ($.Params.HasDefaultValue $name) (printf "rctx.%s" (goifyatt $att $name true)) $name 2 false }}{{/*
*/}}{{ if $validation }}{{ $validation }}
{{ end }}	}
{{ end }}{{ end }}{{/* if .Params */}}	return &rctx, err
}
`

	// ctxMTRespT generates the response helpers for responses with media types.
	// template input: map[string]interface{}
	ctxMTRespT = `// {{ goify .RespName true }} sends a HTTP response with status code {{ .Response.Status }}.
func (ctx *{{ .Context.Name }}) {{ goify .RespName true }}(r {{ gotyperef .Projected .Projected.AllRequired 0 false }}) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "{{ .ContentType }}")
	}
{{ if .Projected.Type.IsArray }}	if r == nil {
		r = {{ gotyperef .Projected .Projected.AllRequired 0 false }}{}
	}
{{ end }}	return ctx.ResponseData.Service.Send(ctx.Context, {{ .Response.Status }}, r)
}
`

	// ctxTRespT generates the response helpers for responses with overridden types.
	// template input: map[string]interface{}
	ctxTRespT = `// {{ goify .Response.Name true }} sends a HTTP response with status code {{ .Response.Status }}.
func (ctx *{{ .Context.Name }}) {{ goify .Response.Name true }}(r {{ gotyperef .Type nil 0 false }}) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "{{ .ContentType }}")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, {{ .Response.Status }}, r)
}
`

	// ctxNoMTRespT generates the response helpers for responses with no known media type.
	// template input: *ContextTemplateData
	ctxNoMTRespT = `
// {{ goify .Response.Name true }} sends a HTTP response with status code {{ .Response.Status }}.
func (ctx *{{ .Context.Name }}) {{ goify .Response.Name true }}({{ if .Response.MediaType }}resp []byte{{ end }}) error {
{{ if .Response.MediaType }}	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "{{ .Response.MediaType }}")
	}
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

{{ $assignment := finalizeCode .Payload.AttributeDefinition "payload" 1 }}{{ if $assignment }}// Finalize sets the default values defined in the design.
func (payload {{ gotyperef .Payload .Payload.AllRequired 0 true }}) Finalize() {
{{ $assignment }}
}{{ end }}

{{ $validation := validationCode .Payload.AttributeDefinition false false false "payload" "raw" 1 true }}{{ if $validation }}// Validate runs the validation rules defined in the design.
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

{{ $validation := validationCode .Payload.AttributeDefinition false false false "payload" "raw" 1 false }}{{ if $validation }}// Validate runs the validation rules defined in the design.
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
{{ $res := .Resource }}{{ if .Origins }}{{ range .PreflightPaths }}{{/*
*/}}	service.Mux.Handle("OPTIONS", {{ printf "%q" . }}, ctrl.MuxHandler("preflight", handle{{ $res }}Origin(cors.HandlePreflight()), nil))
{{ end }}{{ end }}{{ range .Actions }}{{ $action := . }}
	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		// Check if there was an error loading the request
		if err := goa.ContextError(ctx); err != nil {
			return err
		}
		// Build the context
		rctx, err := New{{ .Context }}(ctx, req, service)
		if err != nil {
			return err
		}
{{ if .Payload }}		// Build the payload
		if rawPayload := goa.ContextRequest(ctx).Payload; rawPayload != nil {
			rctx.Payload = rawPayload.({{ gotyperef .Payload nil 1 false }})
{{ if not .PayloadOptional }}		} else {
			return goa.MissingPayloadError()
{{ end }}		}
{{ end }}		return ctrl.{{ .Name }}(rctx)
	}
{{ if .Security }}	h = handleSecurity({{ printf "%q" .Security.Scheme.SchemeName }}, h{{ range .Security.Scopes }}, {{ printf "%q" . }}{{ end }})
{{ end }}{{ if $.Origins }}	h = handle{{ $res }}Origin(h)
{{ end }}{{ range .Routes }}	service.Mux.Handle("{{ .Verb }}", {{ printf "%q" .FullPath }}, ctrl.MuxHandler({{ printf "%q" $action.DesignName }}, h, {{ if $action.Payload }}{{ $action.Unmarshal }}{{ else }}nil{{ end }}))
	service.LogInfo("mount", "ctrl", {{ printf "%q" $res }}, "action", {{ printf "%q" $action.Name }}, "route", {{ printf "%q" (printf "%s %s" .Verb .FullPath) }}{{ with $action.Security }}, "security", {{ printf "%q" .Scheme.SchemeName }}{{ end }})
{{ end }}{{ end }}{{ range .FileServers }}
	h = ctrl.FileHandler({{ printf "%q" .RequestPath }}, {{ printf "%q" .FilePath }})
{{ if .Security }}	h = handleSecurity({{ printf "%q" .Security.Scheme.SchemeName }}, h{{ range .Security.Scopes }}, {{ printf "%q" . }}{{ end }})
{{ end }}{{ if $.Origins }}	h = handle{{ $res }}Origin(h)
{{ end }}	service.Mux.Handle("GET", "{{ .RequestPath }}", ctrl.MuxHandler("serve", h, nil))
	service.LogInfo("mount", "ctrl", {{ printf "%q" $res }}, "files", {{ printf "%q" .FilePath }}, "route", {{ printf "%q" (printf "GET %s" .RequestPath) }}{{ with .Security }}, "security", {{ printf "%q" .Scheme.SchemeName }}{{ end }})
{{ end }}}
`

	// handleCORST generates the code that checks whether a CORS request is authorized
	// template input: *ControllerTemplateData
	handleCORST = `// handle{{ .Resource }}Origin applies the CORS response headers corresponding to the origin.
func handle{{ .Resource }}Origin(h goa.Handler) goa.Handler {
{{ range $i, $policy := .Origins }}{{ if $policy.Regexp }}	spec{{$i}} := regexp.MustCompile({{ printf "%q" $policy.Origin }})
{{ end }}{{ end }}
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		origin := req.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			return h(ctx, rw, req)
		}
{{ range $i, $policy := .Origins }}		{{ if $policy.Regexp }}if cors.MatchOriginRegexp(origin, spec{{$i}}){{else}}if cors.MatchOrigin(origin, {{ printf "%q" $policy.Origin }}){{end}} {
			ctx = goa.WithLogContext(ctx, "origin", origin)
			rw.Header().Set("Access-Control-Allow-Origin", origin)
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
	}{{ $assignment := finalizeCode .Payload.AttributeDefinition "payload" 1 }}{{ if $assignment }}
	payload.Finalize(){{ end }}{{ else }}var payload {{ gotypename .Payload nil 1 false }}
	if err := service.DecodeRequest(req, &payload); err != nil {
		return err
	}{{ end }}{{ $validation := validationCode .Payload.AttributeDefinition false false false "payload" "raw" 1 false }}{{ if $validation }}
	if err := payload.Validate(); err != nil {
		// Initialize payload with private data structure so it can be logged
		goa.ContextRequest(ctx).Payload = payload
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
{{ range $param := .CanonicalParams }}	param{{$param}} := strings.TrimLeftFunc(fmt.Sprintf("%v", {{$param}}), func(r rune) bool { return r == '/' })
{{ end }}{{ if .CanonicalParams }}	return fmt.Sprintf("{{ .CanonicalTemplate }}", param{{ join .CanonicalParams ", param" }})
{{ else }}	return "{{ .CanonicalTemplate }}"
{{ end }}}
{{ end }}`

	// mediaTypeT generates the code for a media type.
	// template input: MediaTypeTemplateData
	mediaTypeT = `// {{ gotypedesc . true }}
//
// Identifier: {{ .Identifier }}{{ $typeName := gotypename . .AllRequired 0 false }}
type {{ $typeName }} {{ gotypedef . 0 true false }}

{{ $validation := validationCode .AttributeDefinition false false false "mt" "response" 1 false }}{{ if $validation }}// Validate validates the {{$typeName}} media type instance.
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
{{ $validation := validationCode .AttributeDefinition false false false "ut" "response" 1 false }}{{ if $validation }}// Validate validates the {{$typeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 false }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
`

	// userTypeT generates the code for a user type.
	// template input: UserTypeTemplateData
	userTypeT = `// {{ gotypedesc . false }}{{ $privateTypeName := gotypename . .AllRequired 0 true }}
type {{ $privateTypeName }} {{ gotypedef . 0 true true }}
{{ $assignment := finalizeCode .AttributeDefinition "ut" 1 }}{{ if $assignment }}// Finalize sets the default values for {{$privateTypeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 true }}) Finalize() {
{{ $assignment }}
}{{ end }}
{{ $validation := validationCode .AttributeDefinition false false false "ut" "request" 1 true }}{{ if $validation }}// Validate validates the {{$privateTypeName}} type instance.
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
{{ $validation := validationCode .AttributeDefinition false false false "ut" "type" 1 false }}{{ if $validation }}// Validate validates the {{$typeName}} type instance.
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
{{ $funcName := printf "Use%sMiddleware" (goify .SchemeName true) }}// {{ $funcName }} mounts the {{ .SchemeName }} auth middleware onto the service.
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
