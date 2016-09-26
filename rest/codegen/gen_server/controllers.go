package genserver

import (
	"regexp"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

// WildcardRegex is the regex used to capture path parameters.
var WildcardRegex = regexp.MustCompile("(?:[^/]*/:([^/]+))+")

type (
	// ControllerWriter generate code for a controller.
	ControllerWriter struct {
		*codegen.SourceFile
		CtrlTmpl  *template.Template
		MountTmpl *template.Template
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
)

// NewControllerWriter returns a writer that writes the code for a controller and its data types.
func NewControllerWriter(filename string) (*ControllerWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &ControllerWriter{SourceFile: file}, nil
}

// Execute generates the controller code for the given resource data.
func (w *ControllerWriter) Execute(data []*ControllerTemplateData) error {
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

const (
	// ctxT generates the code for the context data type.
	// template input: *ContextTemplateData
	ctxT = `// {{ .Name }} provides the {{ .ResourceName }} {{ .ActionName }} action context.
type {{ .Name }} struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
{{ if .Headers }}{{ range $name, $att := ToObject(.Headers.Type) }}{{ if not ($.HasParamAndHeader $name) }}{{/*
*/}}	{{ goifyatt $att $name true }} {{ if and $att.Type.IsPrimitive ($.Headers.IsPrimitivePointer $name) }}*{{ end }}{{ gotyperef .Type nil 0 false }}
{{ end }}{{ end }}{{ end }}{{ if .Params }}{{ range $name, $att := ToObject(.Params.Type) }}{{/*
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
func New{{ .Name }}(ctx context.Context, service *goa.Service) (*{{ .Name }}, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	rctx := {{ .Name }}{Context: ctx, ResponseData: resp, RequestData: req}{{/*
*/}}
{{ if .Headers }}{{ range $name, $att := ToObject(.Headers.Type) }}	header{{ goify $name true }} := req.Header["{{ canonicalHeaderKey $name }}"]
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

*/}}{{ if.Params }}{{ range $name, $att := ToObject(.Params.Type) }}	param{{ goify $name true }} := req.Params["{{ $name }}"]
{{ $mustValidate := $.MustValidate $name }}{{ if $mustValidate }}	if len(param{{ goify $name true }}) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("{{ $name }}"))
	} else {
{{ else }}	if len(param{{ goify $name true }}) > 0 {
{{ end }}{{/* if $mustValidate */}}{{ if $att.Type.IsArray }}{{ if eq (arrayAttribute $att).Type.Kind 4 }}		params := param{{ goify $name true }}
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
	ctx.ResponseData.Header().Set("Content-Type", "{{ .ContentType }}")
	return ctx.ResponseData.Service.Send(ctx.Context, {{ .Response.Status }}, r)
}
`

	// ctxTRespT generates the response helpers for responses with overridden types.
	// template input: map[string]interface{}
	ctxTRespT = `// {{ goify .Response.Name true }} sends a HTTP response with status code {{ .Response.Status }}.
func (ctx *{{ .Context.Name }}) {{ goify .Response.Name true }}(r {{ gotyperef .Type nil 0 false }}) error {
	ctx.ResponseData.Header().Set("Content-Type", "{{ .ContentType }}")
	return ctx.ResponseData.Service.Send(ctx.Context, {{ .Response.Status }}, r)
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

	// mountT generates the code for a resource "Mount" function.
	// template input: *ControllerTemplateData
	mountT = `
// Mount{{ .Resource }}Controller "mounts" a {{ .Resource }} resource controller on the given service.
func Mount{{ .Resource }}Controller(service *goa.Service, ctrl {{ .Resource }}Controller) {
	initService(service)
	var h goa.Handler
{{ $res := .Resource }}{{ if .Origins }}{{ range .PreflightPaths }}{{/*
*/}}	service.Mux.Handle("OPTIONS", "{{ . }}", ctrl.MuxHandler("preflight", handle{{ $res }}Origin(cors.HandlePreflight()), nil))
{{ end }}{{ end }}{{ range .Actions }}{{ $action := . }}
	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		// Check if there was an error loading the request
		if err := goa.ContextError(ctx); err != nil {
			return err
		}
		// Build the context
		rctx, err := New{{ .Context }}(ctx, service)
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
{{ if $.Origins }}	h = handle{{ $res }}Origin(h)
{{ end }}{{ if .Security }}	h = handleSecurity({{ printf "%q" .Security.Scheme.SchemeName }}, h{{ range .Security.Scopes }}, {{ printf "%q" . }}{{ end }})
{{ end }}{{ range .Routes }}	service.Mux.Handle("{{ .Verb }}", {{ printf "%q" .FullPath }}, ctrl.MuxHandler({{ printf "%q" $action.Name }}, h, {{ if $action.Payload }}{{ $action.Unmarshal }}{{ else }}nil{{ end }}))
	service.LogInfo("mount", "ctrl", {{ printf "%q" $res }}, "action", {{ printf "%q" $action.Name }}, "route", {{ printf "%q" (printf "%s %s" .Verb .FullPath) }}{{ with $action.Security }}, "security", {{ printf "%q" .Scheme.SchemeName }}{{ end }})
{{ end }}{{ end }}{{ range .FileServers }}
	h = ctrl.FileHandler({{ printf "%q" .RequestPath }}, {{ printf "%q" .FilePath }})
{{ if $.Origins }}	h = handle{{ $res }}Origin(h)
{{ end }}{{ if .Security }}	h = handleSecurity({{ printf "%q" .Security.Scheme.SchemeName }}, h{{ range .Security.Scopes }}, {{ printf "%q" . }}{{ end }})
{{ end }}	service.Mux.Handle("GET", "{{ .RequestPath }}", ctrl.MuxHandler("serve", h, nil))
	service.LogInfo("mount", "ctrl", {{ printf "%q" $res }}, "files", {{ printf "%q" .FilePath }}, "route", {{ printf "%q" (printf "GET %s" .RequestPath) }}{{ with .Security }}, "security", {{ printf "%q" .Scheme.SchemeName }}{{ end }})
{{ end }}}
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
)
