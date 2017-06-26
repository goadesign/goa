// MarshalTypes codegen file
// Make sure to generate user type for each error even if declared with primitive
// Add metadata to specify error type attribute to be used for error message
// validate only one error per http status code
// Make sure primitive rename of []byte to bytes didn't break stuff

// DSL validation: make sure there's at least one response for all actions
// Make sure all routes define identical path params
// Remove required attributes from 'Required' slice that have default values

// Add response tags to account example
// Test response tags

// Initialize error responses headers and body from result (action.go)

package rest

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// ServerFiles returns all the server HTTP transport files.
func ServerFiles(root *rest.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.Resources))
	for i, r := range root.Resources {
		fw[i] = Server(r)
	}
	return fw
}

// Server returns the server HTTP transport file
func Server(r *rest.ResourceExpr) codegen.File {
	path := filepath.Join(codegen.KebabCase(r.Name()), "transport", "http", "server.go")
	data := Resources.Get(r.Name())
	sections := func(genPkg string) []*codegen.Section {
		title := fmt.Sprintf("%s server HTTP transport", r.Name())
		s := []*codegen.Section{
			codegen.Header(title, "http", []*codegen.ImportSpec{
				{Path: "fmt"},
				{Path: "io"},
				{Path: "net/http"},
				{Path: "strconv"},
				{Path: "strings"},
				{Path: "goa.design/goa.v2"},
				{Path: "goa.design/goa.v2/rest"},
				{Path: genPkg + "/endpoints"},
				{Path: genPkg + "/services"},
			}),
			{Template: serverStructTmpl(r), Data: data},
			{Template: serverConstructorTmpl(r), Data: data},
			{Template: serverMountTmpl(r), Data: data},
		}

		for _, a := range data.Actions {
			as := []*codegen.Section{
				{Template: serverHandlerTmpl(r), Data: a},
				{Template: serverHandlerConstructorTmpl(r), Data: a},
			}
			s = append(s, as...)

			if len(a.Responses) > 0 {
				s = append(s, &codegen.Section{Template: serverEncoderTmpl(r), Data: a})
			}

			if a.Payload != nil {
				s = append(s, &codegen.Section{Template: serverDecoderTmpl(r), Data: a})
			}

			if len(a.ErrorResponses) > 0 {
				s = append(s, &codegen.Section{Template: serverErrorEncoderTmpl(r), Data: a})
			}
		}
		return s
	}

	return codegen.NewSource(path, sections)
}

func serverStructTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("struct").Parse(serverStructT))
}

func serverConstructorTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("constructor").Parse(serverConstructorT))
}

func serverMountTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("mount").Parse(serverMountT))
}

func serverHandlerTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("handler").Parse(serverHandlerT))
}

func serverHandlerConstructorTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("handler_constructor").Parse(serverHandlerConstructorT))
}

func serverDecoderTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("decoder").Parse(serverDecoderT))
}

func serverEncoderTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("encoder").Parse(serverEncoderT))
}

func serverErrorEncoderTmpl(r *rest.ResourceExpr) *template.Template {
	return template.Must(serverTmpl(r).New("error_encoder").Parse(serverErrorEncoderT))
}

func serverTmpl(r *rest.ResourceExpr) *template.Template {
	scope := files.Services.Get(r.Name()).Scope
	return template.New("server").
		Funcs(template.FuncMap{
			"goTypeRef":            scope.GoTypeRef,
			"conversionData":       conversionData,
			"headerConversionData": headerConversionData,
			"printValue":           printValue,
		}).
		Funcs(codegen.TemplateFuncs())
}

// conversionData creates a template context suitable for executing the
// "type_conversion" template.
func conversionData(varName, name string, dt design.DataType) map[string]interface{} {
	return map[string]interface{}{
		"VarName": varName,
		"Name":    name,
		"Type":    dt,
	}
}

// headerConversionData produces the template data suitable for executing the
// "header_conversion" template.
func headerConversionData(dt design.DataType, varName string, required bool, target string) map[string]interface{} {
	return map[string]interface{}{
		"Type":     dt,
		"VarName":  varName,
		"Required": required,
		"Target":   target,
	}
}

// printValue generates the Go code for a literal string containing the given
// value. printValue panics if the data type is not a primitive or an array.
func printValue(dt design.DataType, v interface{}) string {
	switch actual := dt.(type) {
	case *design.Array:
		val := reflect.ValueOf(v)
		elems := make([]string, val.Len())
		for i := 0; i < val.Len(); i++ {
			elems[i] = printValue(actual.ElemType.Type, val.Index(i).Interface())
		}
		return strings.Join(elems, ", ")
	case design.Primitive:
		return fmt.Sprintf("%v", v)
	default:
		panic("unsupported type value " + dt.Name()) // bug
	}
}

const serverStructT = `{{ printf "%s lists the %s service endpoint HTTP handlers." .HandlersStruct .Service.Name | comment }}
type {{ .HandlersStruct }} struct {
	{{- range .Actions }}
	{{ .Method.VarName }} http.Handler
	{{- end }}
}
`

const serverConstructorT = `{{ printf "%s instantiates HTTP handlers for all the %s service endpoints." .Constructor .Service.Name | comment }}
func {{ .Constructor }}(
	e *endpoints.{{ .Service.VarName }},
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) *{{ .HandlersStruct }} {
	return &{{ .HandlersStruct }}{
		{{- range .Actions }}
		{{ .Method.VarName }}: {{ .Constructor }}(e.{{ .Method.VarName }}, dec, enc, logger),
		{{- end }}
	}
}
`

const serverMountT = `{{ printf "%s configures the mux to serve the %s endpoints." .MountHandlers .Service.Name | comment }}
func {{ .MountHandlers }}(mux rest.Muxer, h *{{ .HandlersStruct }}) {
	{{- range .Actions }}
	{{ .MountHandler }}(mux, h.{{ .Method.VarName }})
	{{- end }}
}
`

const serverHandlerT = `{{ printf "%s configures the mux to serve the \"%s\" service \"%s\" endpoint." .MountHandler .ServiceName .Method.Name | comment }}
func {{ .MountHandler }}(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	{{- range .Routes }}
	mux.Handle("{{ .Method }}", "{{ .Path }}", f)
	{{- end }}
}
`

const serverHandlerConstructorT = `{{ printf "%s creates a HTTP handler which loads the HTTP request and calls the \"%s\" service \"%s\" endpoint." .Constructor .ServiceName .Method.Name | comment }}
{{ comment "The middleware is mounted so it executes after the request is loaded and thus may access the request state via the rest package ContextXXX functions."}}
func {{ .Constructor }}(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		{{- if .Payload }}
		decodeRequest  = {{ .Decoder }}(dec)
		{{- end }}
		{{- if .Responses }}
		encodeResponse = {{ .Encoder }}EncodeResponse(enc)
		{{- end }}
		encodeError    = {{ if .ErrorResponses }}{{ .ErrorEncoder }}{{ else }}rest.EncodeError{{ end }}(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		{{- if .Payload }}
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, goa.ErrInvalid("request invalid: %s", err))
			return
		}

		res, err := endpoint(r.Context(), payload)
		{{- else }}
		res, err := endpoint(r.Context())
		{{- end }}

		if err != nil {
			encodeError(w, r, err)
			return
		}
		{{- if .Responses }}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
		{{- else }}
		w.Write(http.StatusNoContent)
		{{- end }}
	})
}
`

const serverDecoderT = `{{ printf "%s returns a decoder for requests sent to the %s %s endpoint." .Decoder .ServiceName .Method.Name | comment }}
func {{ .Decoder }}(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ({{ .Payload.Ref }}, error) {

{{- if .Payload.RequestBody }}
		var (
			body {{ .Payload.RequestBody.VarName }}
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		{{- if .Payload.RequestBody.ValidateRef }}
		{{ .Payload.RequestBody.ValidateRef }}
		{{- end }}
{{ end }}

{{- if or .Payload.PathParams .Payload.QueryParams .Payload.Headers }}
		var (
		{{- range .Payload.PathParams }}
			{{ .VarName }} {{ if .Pointer }}*{{ end }}{{goTypeRef .Type }}
		{{- end }}
		{{- range .Payload.QueryParams }}
			{{ .VarName }} {{ if .Pointer }}*{{ end }}{{goTypeRef .Type }}
		{{- end }}
		{{- range .Payload.Headers }}
			{{ .VarName }} {{ if .Pointer }}*{{ end }}{{goTypeRef .Type }}
		{{- end }}
		{{- if not .Payload.RequestBody }}
			err error
		{{- end }}
		{{- if .Payload.PathParams }}

			params = rest.ContextParams(r.Context())
		{{- end }}
		)

{{- range .Payload.PathParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }} = params["{{ .Name }}"]

	{{- else }}{{/* not string */}}
		{{ .VarName }}Raw := params["{{ .Name }}"]
		{{- template "path_conversion" . }}

	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}

{{- range .Payload.QueryParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = r.URL.Query().Get("{{ .Name }}")
		if {{ .VarName }} == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := r.URL.Query().Get("{{ .Name }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}{{ printf "%q" .DefaultValue }}{{ else }}{{ printf "%#v" .DefaultValue }}{{ end }}
		}
		{{- end }}

	{{- else if .StringSlice }}
		{{ .VarName }} = r.URL.Query()["{{ .Name }}"]
		{{- if .Required }}
		if {{ .VarName }} == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }} == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

	{{- else if .Slice }}
		{{ .VarName }}Raw := r.URL.Query()["{{ .Name }}"]
		{{- if .Required }}
		if {{ .VarName }}Raw == nil {
			return nil, goa.MissingFieldError("{{ .Name }}", "query string")
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != nil {
		{{- end }}
		{{- template "slice_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}

	{{- else if .MapStringSlice }}
		{{ .VarName }} = r.URL.Query()
		{{- if .Required }}
		if len({{ .VarName }}) == 0 {
			return nil, goa.MissingFieldError("{{ .Name }}", "query string")
		}
		{{- else if .DefaultValue }}
		if len({{ .VarName }}Raw) == 0 {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

	{{- else if .Map }}
		{{ .VarName }}Raw := r.URL.Query()
		{{- if .Required }}
		if len({{ .VarName }}Raw) == 0 {
			return nil, goa.MissingFieldError("{{ .Name }}", "query string")
		}
		{{- else if .DefaultValue }}
		if len({{ .VarName }}Raw) == 0 {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if len({{ .VarName }}Raw) != 0 {
		{{- end }}
		{{- if eq .Type.ElemType.Type.Name "array" }}
			{{- if eq .Type.ElemType.Type.ElemType.Type.Name "string" }}
			{{- template "map_key_conversion" . }}
			{{- else }}
			{{- template "map_slice_conversion" . }}
			{{- end }}
		{{- else }}
			{{- template "map_conversion" . }}
		{{- end }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}

	{{- else }}{{/* not string, not any, not slice and not map */}}
		{{ .VarName }}Raw := r.URL.Query().Get("{{ .Name }}")
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			return nil, goa.MissingFieldError("{{ .Name }}", "query string")
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == "" {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}
		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != "" {
		{{- end }}
		{{- template "type_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	
	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}

{{- range .Payload.Headers }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = r.Header.Get("{{ .Name }}")
		if {{ .VarName }} == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := r.Header.Get("{{ .Name }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}{{ printf "%q" .DefaultValue }}{{ else }}{{ printf "%#v" .DefaultValue }}{{ end }}
		}
		{{- end }}

	{{- else if .StringSlice }}
		{{ .VarName }} = r.Header["{{ .CanonicalName }}"]
		{{ if .Required }}
		if {{ .VarName }} == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }} == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

	{{- else if .Slice }}
		{{ .VarName }}Raw := r.Header["{{ .CanonicalName }}"]
		{{ if .Required }}if {{ .VarName }}Raw == nil {
			return nil, goa.MissingFieldError("{{ .Name }}", "header")
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != nil {
		{{- end }}
		{{- template "slice_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}

	{{- else }}{{/* not string, not any and not slice */}}
		{{ .VarName }}Raw := r.Header.Get("{{ .Name }}")
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			return nil, goa.MissingFieldError("{{ .Name }}", "header")
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }}Raw == "" {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

		{{- if .DefaultValue }}else {
		{{- else if not .Required }}
		if {{ .VarName }}Raw != "" {
		{{- end }}
		{{- template "type_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}
{{ end }}
		if err != nil {
			return nil, err
		}
		{{- if .Payload.Constructor }}
		return {{ .Payload.Constructor }}({{ .Payload.ConstructorParams }}), nil
		{{- else if .Payload.DecoderReturnValue }}
		return {{ .Payload.DecoderReturnValue }}, nil
		{{- end }}
	}
}

{{- define "path_conversion" }}
	{{- if eq .Type.Name "array" }}
		{{ .VarName }}RawSlice := strings.Split({{ .VarName }}Raw, ",")
		{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}RawSlice))
		for i, rv := range {{ .VarName }}RawSlice {
			{{- template "slice_item_conversion" . }}
		}
	{{- else }}
		{{- template "type_conversion" . }}
	{{- end }}
{{- end }}

{{- define "slice_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for i, rv := range {{ .VarName }}Raw {
		{{- template "slice_item_conversion" . }}
	}
{{- end }}

{{- define "map_key_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for keyRaw, val := range {{ .VarName }}Raw {
		var key {{ goTypeRef .Type.KeyType.Type }}
		{
		{{- template "type_conversion" (conversionData "key" (printf "%q" "query") .Type.KeyType.Type) }}
		}
		{{ .VarName }}[key] = val
	}
{{- end }}

{{- define "map_slice_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for key{{ if not (eq .Type.KeyType.Type.Name "string") }}Raw{{ end }}, valRaw := range {{ .VarName }}Raw {

		{{- if not (eq .Type.KeyType.Type.Name "string") }}
		var key {{ goTypeRef .Type.KeyType.Type }}
		{
			{{- template "type_conversion" (conversionData "key" (printf "%q" "query") .Type.KeyType.Type) }}
		}
		{{- end }}
		var val {{ goTypeRef .Type.ElemType.Type }}
		{
		{{- template "slice_conversion" (conversionData "val" (printf "%q" "query") .Type.ElemType.Type) }}
		}
		{{ .VarName }}[key] = val
	}
{{- end }}

{{- define "map_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for key{{ if not (eq .Type.KeyType.Type.Name "string") }}Raw{{ end }}, va := range {{ .VarName }}Raw {

		{{- if not (eq .Type.KeyType.Type.Name "string") }}
		var key {{ goTypeRef .Type.KeyType.Type }}
		{
			{{- if eq .Type.KeyType.Type.Name "string" }}
			key = keyRaw
			{{- else }}
			{{- template "type_conversion" (conversionData "key" (printf "%q" "query") .Type.KeyType.Type) }}
			{{- end }}
		}
		{{- end }}
		var val {{ goTypeRef .Type.ElemType.Type }}
		{
			{{- if eq .Type.ElemType.Type.Name "string" }}
			val = va[0]
			{{- else }}
			valRaw := va[0]
			{{- template "type_conversion" (conversionData "val" (printf "%q" "query") .Type.ElemType.Type) }}
			{{- end }}
		}
		{{ .VarName }}[key] = val
	}
{{- end }}

{{- define "type_conversion" }}
	{{- if eq .Type.Name "bytes" }}
		{{ .VarName }} = []byte({{.VarName}}Raw)
	{{- else if eq .Type.Name "int" }}
		v, err := strconv.ParseInt({{ .VarName }}Raw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "integer")
		}
		{{- if .Pointer }}
		pv := int(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = int(v)
		{{- end }}
	{{- else if eq .Type.Name "int32" }}
		v, err := strconv.ParseInt({{ .VarName }}Raw, 10, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "integer")
		}
		{{- if .Pointer }}
		pv := int32(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = int32(v)
		{{- end }}
	{{- else if eq .Type.Name "int64" }}
		v, err := strconv.ParseInt({{ .VarName }}Raw, 10, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "integer")
		}
		{{ .VarName }} = {{ if .Pointer}}&{{ end }}v
	{{- else if eq .Type.Name "uint" }}
		v, err := strconv.ParseUint({{ .VarName }}Raw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "unsigned integer")
		}
		{{- if .Pointer }}
		pv := uint(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = uint(v)
		{{- end }}
	{{- else if eq .Type.Name "uint32" }}
		v, err := strconv.ParseUint({{ .VarName }}Raw, 10, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "unsigned integer")
		}
		{{- if .Pointer }}
		pv := uint32(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = uint32(v)
		{{- end }}
	{{- else if eq .Type.Name "uint64" }}
		v, err := strconv.ParseUint({{ .VarName }}Raw, 10, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "unsigned integer")
		}
		{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
	{{- else if eq .Type.Name "float32" }}
		v, err := strconv.ParseFloat({{ .VarName }}Raw, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "float")
		}
		{{- if .Pointer }}
		pv := float32(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = float32(v)
		{{- end }}
	{{- else if eq .Type.Name "float64" }}
		v, err := strconv.ParseFloat({{ .VarName }}Raw, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "float")
		}
		{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
	{{- else if eq .Type.Name "boolean" }}
		v, err := strconv.ParseBool({{ .VarName }}Raw)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "boolean")
		}
		{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
	{{- else }}
		// unsupported type {{ .Type.Name }} for var {{ .VarName }}
	{{- end }}
{{- end }}
{{- define "slice_item_conversion" }}
		{{- if eq .Type.ElemType.Type.Name "string" }}
			{{ .VarName }}[i] = rv
		{{- else if eq .Type.ElemType.Type.Name "bytes" }}
			{{ .VarName }}[i] = []byte(rv)
		{{- else if eq .Type.ElemType.Type.Name "int" }}
			v, err := strconv.ParseInt(rv, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of integers")
			}
			{{ .VarName }}[i] = int(v)
		{{- else if eq .Type.ElemType.Type.Name "int32" }}
			v, err := strconv.ParseInt(rv, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of integers")
			}
			{{ .VarName }}[i] = int32(v)
		{{- else if eq .Type.ElemType.Type.Name "int64" }}
			v, err := strconv.ParseInt(rv, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of integers")
			}
			{{ .VarName }}[i] = v
		{{- else if eq .Type.ElemType.Type.Name "uint" }}
			v, err := strconv.ParseUint(rv, 10, strconv.IntSize)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of unsigned integers")
			}
			{{ .VarName }}[i] = uint(v)
		{{- else if eq .Type.ElemType.Type.Name "uint32" }}
			v, err := strconv.ParseUint(rv, 10, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of unsigned integers")
			}
			{{ .VarName }}[i] = int32(v)
		{{- else if eq .Type.ElemType.Type.Name "uint64" }}
			v, err := strconv.ParseUint(rv, 10, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of unsigned integers")
			}
			{{ .VarName }}[i] = v
		{{- else if eq .Type.ElemType.Type.Name "float32" }}
			v, err := strconv.ParseFloat(rv, 32)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of floats")
			}
			{{ .VarName }}[i] = float32(v)
		{{- else if eq .Type.ElemType.Type.Name "float64" }}
			v, err := strconv.ParseFloat(rv, 64)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of floats")
			}
			{{ .VarName }}[i] = v
		{{- else if eq .Type.ElemType.Type.Name "boolean" }}
			v, err := strconv.ParseBool(rv)
			if err != nil {
				return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "array of booleans")
			}
			{{ .VarName }}[i] = v
		{{- else if eq .Type.ElemType.Type.Name "any" }}
			{{ .VarName }}[i] = rv
		{{- else }}
			// unsupported slice type {{ .Type.ElemType.Type.Name }} for var {{ .VarName }}
		{{- end }}
{{- end }}
`

const serverEncoderT = `{{ printf "%s returns an encoder for responses returned by the %s %s endpoint." .Encoder .ServiceName .Method.Name | comment }}
func {{ .Encoder }}(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {

	{{- if .Method.ResultRef }}
		res := v.({{ .Method.ResultRef }})

		{{- range .Responses }}

			{{- if .TagName }}
		if res.{{ .TagName }} == {{ printf "%q" .TagValue }} {
			{{- end }}
			{{ template "response" . }}

			{{- if .Body }}
			return enc.Encode(body)
			{{- else }}
			return nil
			{{- end }}

			{{- if .TagName }}
		}
			{{- end }}

		{{- end }}

	{{- else }}

		{{- with (index .Responses 0) }}

		w.WriteHeader({{ .StatusCode }})
		return nil

		{{- end }}

	{{- end }}
	}
}
` + responseT

const serverErrorEncoderT = `{{ printf "%s returns an encoder for errors returned by the %s %s endpoint." .ErrorEncoder .Method.Name .ServiceName | comment }}
func {{ .ErrorEncoder }}(encoder func(http.ResponseWriter, *http.Request) rest.Encoder, logger goa.LogAdapter) func(http.ResponseWriter, *http.Request, error) {
	encodeError := rest.EncodeError(encoder, logger)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch res := v.(type) {

		{{- range .ErrorResponses }}
		case {{ .TypeRef }}:

			{{- template "response" .Response }}
			{{- if .Response.Body }}
			if err := enc.Encode(body); err != nil {
				encodeError(w, r, err)
			}
			{{- end }}

		{{- end }}
		default:
			encodeError(w, r, v)
		}
	}
}
` + responseT

const responseT = `{{ define "response" -}}
	enc, ct := encoder(w, r)
	rest.SetContentType(w, ct)

	{{- if .Body }}
		{{- if .Body.IsUser }}
			{{- if .Body.Fields }}
		body := &{{ .Body.VarName }}{
				{{- range .Body.Fields}}	
			{{ .FieldName }}: res.{{ .VarName }},
				{{- end }}
		}
			{{- else }}
			val := {{ .Body.VarName }}(*res)
			body := &val
			{{- end }}
		{{- else }}
		body := res 
		{{- end }}
	{{- end }}
	{{- range .Headers }}
		{{- if not .Required }}
	if res.{{ .FieldName }} != nil {
		{{- end }}

		{{- if eq .Type.Name "string" }}
	w.Header().Set("{{ .Name }}", {{ if not .Required }}*{{ end }}res.{{ .FieldName }})
		{{- else }}
	v := res.{{ .FieldName }}
	{{ template "header_conversion" (headerConversionData .Type .VarName .Required "v") }}
	w.Header().Set("{{ .Name }}", {{ .VarName }})
		{{- end }}

		{{- if not .Required }}
			{{- if .DefaultValue }}
	} else {
		w.Header().Set("{{ .Name }}", "{{ printValue .Type .DefaultValue }}")
			{{- end }}
	}
		{{- end }}
	{{- end }}
	w.WriteHeader({{ .StatusCode }})
{{- end }}

{{- define "header_conversion" }}
	{{- if eq .Type.Name "boolean" -}}
		{{ .VarName }} := strconv.FormatBool({{ if not .Required }}*{{ end }}{{ .Target }})
	{{- else if eq .Type.Name "int" -}}
		{{ .VarName }} := strconv.Itoa({{ if not .Required }}*{{ end }}{{ .Target }})
	{{- else if eq .Type.Name "int32" -}}
		{{ .VarName }} := strconv.FormatInt(int64({{ if not .Required }}*{{ end }}{{ .Target }}), 10)
	{{- else if eq .Type.Name "int64" -}}
		{{ .VarName }} := strconv.FormatInt({{ if not .Required }}*{{ end }}{{ .Target }}, 10)
	{{- else if eq .Type.Name "uint" -}}
		{{ .VarName }} := strconv.FormatUint(uint64({{ if not .Required }}*{{ end }}{{ .Target }}), 10)
	{{- else if eq .Type.Name "uint32" -}}
		{{ .VarName }} := strconv.FormatUint(uint64({{ if not .Required }}*{{ end }}{{ .Target }}), 10)
	{{- else if eq .Type.Name "uint64" -}}
		{{ .VarName }} := strconv.FormatUint({{ if not .Required }}*{{ end }}{{ .Target }}, 10)
	{{- else if eq .Type.Name "float32" -}}
		{{ .VarName }} := strconv.FormatFloat(float64({{ if not .Required }}*{{ end }}{{ .Target }}), 'f', -1, 32)
	{{- else if eq .Type.Name "float64" -}}
		{{ .VarName }} := strconv.FormatFloat({{ if not .Required }}*{{ end }}{{ .Target }}, 'f', -1, 64)
	{{- else if eq .Type.Name "string" -}}
		{{ .VarName }} := {{ .Target }} 
	{{- else if eq .Type.Name "bytes" -}}
		{{ .VarName }} := string({{ .Target }})
	{{- else if eq .Type.Name "any" -}}
		{{ .VarName }} := fmt.Sprintf("%v", {{ .Target }})
	{{- else if eq .Type.Name "array" -}}
		{{- if eq .Type.ElemType.Type.Name "string" -}}
		{{ .VarName }} := strings.Join({{ .Target }}, ", ")
		{{- else -}}
		{{ .VarName }}Slice := make([]string, len({{ .Target }}))
		for i, e := range {{ .Target }}  {
			{{ template "header_conversion" (headerConversionData .Type.ElemType.Type "es" true "e") }}
			{{ .VarName }}Slice[i] = es	
		}
		{{ .VarName }} := strings.Join({{ .VarName }}Slice, ", ")
		{{- end }}
	{{- else }}
		// unsupported type {{ .Type.Name }} for header field {{ .FieldName }}
	{{- end }}
{{- end -}}
`
