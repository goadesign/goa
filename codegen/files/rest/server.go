package rest

import (
	"fmt"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
	"net/http"
	"sort"
	"strings"
)

//todo: in encode headers and mappings
//todo: in decode payload, query and paths

const serverStructT = `{{ printf "%s lists the %s service endpoint HTTP handlers." .HandlersStruct .ServiceName | comment }}
type {{ .HandlersStruct }} struct {
	{{- range .ActionData }}
	{{ .EndpointName }} http.Handler
	{{- end }}
}
`

const serverConstructorT = `{{ printf "%s instantiates HTTP handlers for all the %s service endpoints." .Constructor .ServiceName | comment }}
func {{ .Constructor }}(
	e *endpoints.{{ .ServiceName }},
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) *{{ .HandlersStruct }} {
	return &{{ .HandlersStruct }}{
		{{- range .ActionData }}
		{{ .EndpointName }}: {{ .Constructor }}(e.{{ .EndpointName }}, dec, enc, logger),
		{{- end }}
	}
}
`

const serverMountT = `{{ printf "%s configures the mux to serve the %s endpoints." .MountHandlers .ServiceName | comment }}
func {{ .MountHandlers }}(mux rest.ServeMux, h *{{ .HandlersStruct }}) {
	{{- range .ActionData }}
	{{ .MountHandler }}(mux, h.{{ .EndpointName }})
	{{- end }}
}
`

const serverHandlerT = `{{ printf "%s configures the mux to serve the \"%s\" service \"%s\" endpoint." .MountHandler .ServiceName .EndpointName | comment }}
func {{ .MountHandler }}(mux rest.ServeMux, h http.Handler) {
	{{- range .Routes }}
	mux.Handle("{{ .Method }}", "{{ .Path }}", h)
	{{- end }}
}
`

const serverHandlerConstructorT = `{{ printf "%s creates a HTTP handler which loads the HTTP request and calls the \"%s\" service \"%s\" endpoint." .Constructor .ServiceName .EndpointName | comment }}
{{ comment "The middleware is mounted so it executes after the request is loaded and thus may access the request state via the rest package ContextXXX functions."}}
func {{ .Constructor }}(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		decodeRequest  = {{ .Decoder }}(dec)
		encodeResponse = {{ .Encoder }}EncodeResponse(enc)
		encodeError    = {{ if .HasErrors }}{{ .ErrorEncoder }}{{ else }}EncodeError{{ end }}(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, goa.ErrInvalid("request invalid: %s", err))
			return
		}

		res, err := endpoint(r.Context(), payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}
`

const serverDecoderT = `{{ printf "%s returns a decoder for requests sent to the create %s endpoint." .Decoder .ServiceName | comment }}
func {{ .Decoder }}(decoder rest.RequestDecoderFunc) DecodeRequestFunc {
	return func(r *http.Request) (*service.{{ .Payload.Name }}, error) {
{{- if .Payload.HasBody }}
		var (
			body {{ .Payload.Body }}
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = fmt.Errorf("empty body")
			}
			return nil, err
		}
{{ end }}
{{- if or .Payload.HasParams }}
		{{- if or .Payload.HasPathParams }}
		params := httptreemux.ContextParams(r.Context())
		{{- end }}
		var (
			{{- range .Payload.AllParams }}
			{{ goify .Name false }} {{goTypeRef .Type }}
			{{- end }}
		)
{{ range .Payload.QueryParams }}
	{{- if eq .Type.Name "string" }}
		{{ goify .Name false }} = r.URL.Query().Get("{{ .Name }}")
	{{- else }}
		{{ goify .Name false }}Raw := r.URL.Query().Get("{{ .Name }}")
		{{- template "type_conversion" . }}
	{{- end }}
{{- end }}
{{- range .Payload.PathParams }}
	{{- if eq .Type.Name "string" }}
		{{ goify .Name false }} = params["{{ .Name }}"]
	{{- else }}
		{{ goify .Name false }}Raw := params["{{ .Name }}"]
		{{- template "type_conversion" . }}
	{{- end }}
{{- end }}
{{- end }}
		payload, err := {{ .Payload.Constructor }}(
			{{- if .Payload.HasBody }}&body{{ end -}}
			{{- range $i, $p := .Payload.AllParams }}
				{{- if or (ne $i 0) ($.Payload.HasBody) }}, {{ end -}}{{ goify .Name false }}
			{{- end }})
		return payload, err
	}
}
{{ define "type_conversion"}}
	{{- if eq .Type.Name "int" }}
		if v, err := strconv.ParseInt({{ goify .Name false }}Raw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an integer, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = int(v)
		}
	{{- else if eq .Type.Name "int32" }}
		if v, err := strconv.ParseInt({{ goify .Name false }}Raw, 10, 32); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an integer, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = int32(v)
		}
	{{- else if eq .Type.Name "int64" }}
		if v, err := strconv.ParseInt({{ goify .Name false }}Raw, 10, 64); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an integer, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = v
		}
	{{- else if eq .Type.Name "uint" }}
		if v, err := strconv.ParseUint({{ goify .Name false }}Raw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = uint(v)
		}
	{{- else if eq .Type.Name "uint32" }}
		if v, err := strconv.ParseUint({{ goify .Name false }}Raw, 10, 32); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = int32(v)
		}
	{{- else if eq .Type.Name "uint64" }}
		if v, err := strconv.ParseUint({{ goify .Name false }}Raw, 10, 64); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = v
		}
	{{- else if eq .Type.Name "float32" }}
		if v, err := strconv.ParseFloat({{ goify .Name false }}Raw, 32); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be a float, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = float32(v)
		}
	{{- else if eq .Type.Name "float64" }}
		if v, err := strconv.ParseFloat({{ goify .Name false }}Raw, 64); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be a float, got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = v
		}
	{{- else if eq .Type.Name "boolean" }}
		if v, err := strconv.ParseBool({{ goify .Name false }}Raw); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be a boolean (true or false), got '%s'", {{ goify .Name false }}Raw)
		} else {
			{{ goify .Name false }} = v
		}
	{{- else }}
		// unsupported type YET!
	{{- end }}
{{end -}}
`

const serverEncoderT = `{{ printf "%s returns an encoder for responses returned by the %s %s endpoint." .Encoder .EndpointName .ServiceName | comment }}
func {{ .Encoder }}(encoder rest.ResponseEncoderFunc) EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	{{- if eq (len .Responses) 0 }}
		w.WriteHeader(http.StatusNoContent)
	{{- else if eq (len .Responses) 1 }}
		{{- with index .Responses 0}}
		{{- if .HasBody }}
		w.Header().Set("Content-Type", ResponseContentType(r)){{ end }}
		{{- /* TODO: need a way to fill and map header values e.g. w.Header().Set("Location", v.Href) */}}
		w.WriteHeader({{ .StatusCode }})
		{{- if .HasBody }}
		if v != nil {
			return encoder(w, r).Encode(v)
		}
		{{- end }}
		{{- end }}
	{{- else }}
		switch t := v.(type) {
		{{- range .Responses }}
		case *{{ .Name }}:
			{{- /* TODO: need a way to fill and map header values e.g. w.Header().Set("Location", v.Href) */}}
			w.WriteHeader({{ .StatusCode }})
			{{- if .HasBody }}
			return encoder(w, r).Encode(t)
			{{- end }}
		{{- end }}
		default:
			return fmt.Errorf("invalid response type")
		}
	{{- end }}
		return nil
	}
}
`

const serverErrorEncoderT = `{{ printf "%s returns an encoder for errors returned by the %s %s endpoint." .ErrorEncoder .EndpointName .ServiceName | comment }}
func {{ .ErrorEncoder }}(encoder rest.ResponseEncoderFunc, logger goa.Logger) EncodeErrorFunc {
	encodeError := EncodeError(encoder, logger)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		w.Header().Set("Content-Type", ResponseContentType(r))
		switch t := v.(type) {
		{{ range .HTTPErrors -}}
		case *service.{{ .Name }}:
			w.WriteHeader({{ .StatusCode }})
			encoder(w, r).Encode(t)
		{{- end }}
		default:
			encodeError(w, r, v)
		}
	}
}
`

type (
	serverData struct {
		ServiceName string

		HandlersStruct string
		Constructor    string
		MountHandlers  string

		ActionData []*serverActionData
	}

	serverActionData struct {
		EndpointName string
		ServiceName  string

		Routes []*serverRouteData

		MountHandler string
		Constructor  string
		Decoder      string
		Encoder      string
		ErrorEncoder string

		Payload *serverPayloadData

		Responses  []*serverResponseData
		HTTPErrors []*serverResponseData
	}

	serverPayloadData struct {
		Name        string
		Constructor string
		Body        string
		hasBody     bool

		PathParams  []*serverParamData
		QueryParams []*serverParamData
		AllParams   []*serverParamData
	}

	serverRouteData struct {
		Method string
		Path   string
	}

	serverResponseData struct {
		Name       string
		StatusCode string
		HasBody    bool
	}

	serverParamData struct {
		Name     string
		Type     design.DataType
		Required bool
	}

	// serverWriter
	serverWriter struct {
		sections   []*codegen.Section
		outputPath string
	}
)

var (
	serverTmpl = template.New("server").Funcs(template.FuncMap{
		"add":       codegen.Add,
		"goTypeRef": codegen.GoTypeRef,
		"goify":     codegen.Goify,
	}).Funcs(codegen.TemplateFuncs())
	serverStructTmpl             = template.Must(serverTmpl.New("struct").Parse(serverStructT))
	serverConstructorTmpl        = template.Must(serverTmpl.New("constructor").Parse(serverConstructorT))
	serverMountTmpl              = template.Must(serverTmpl.New("mount").Parse(serverMountT))
	serverHandlerTmpl            = template.Must(serverTmpl.New("handler").Parse(serverHandlerT))
	serverHandlerConstructorTmpl = template.Must(serverTmpl.New("handler_constructor").Parse(serverHandlerConstructorT))
	serverDecoderTmpl            = template.Must(serverTmpl.New("decoder").Parse(serverDecoderT))
	serverEncoderTmpl            = template.Must(serverTmpl.New("encoder").Parse(serverEncoderT))
	serverErrorEncoderTmpl       = template.Must(serverTmpl.New("error_encoder").Parse(serverErrorEncoderT))
)

// ServerWriters returns the server HTTP transport writers.
func ServerWriters(api *design.APIExpr, root *rest.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.Resources))
	for i, r := range root.Resources {
		title := fmt.Sprintf("%s server HTTP transport", api.Name)
		sections := []*codegen.Section{
			codegen.Header(title, "http", []*codegen.ImportSpec{
				{Path: "fmt"},
				{Path: "io"},
				{Path: "net/http"},
				{Path: "strconv"},
				{Path: "github.com/dimfeld/httptreemux"},
				{Path: "goa.design/goa.v2"},
				{Path: "goa.design/goa.v2/examples/account/gen/endpoints"},
				{Path: "goa.design/goa.v2/examples/account/gen/services"},
				{Path: "goa.design/goa.v2/rest"},
			}),
		}

		fw[i] = &serverWriter{
			sections:   sections,
			outputPath: fmt.Sprintf("gen/transport/http/%s_server%%d.go", codegen.SnakeCase(r.Name)),
		}
	}
	return fw
}

func Server(r *rest.ResourceExpr) []*codegen.Section {

	d := buildServerData(r)

	s := []*codegen.Section{
		{Template: serverStructTmpl, Data: d},
		{Template: serverConstructorTmpl, Data: d},
		{Template: serverMountTmpl, Data: d},
	}

	for _, a := range d.ActionData {
		as := []*codegen.Section{
			{Template: serverHandlerTmpl, Data: a},
			{Template: serverHandlerConstructorTmpl, Data: a},
			{Template: serverEncoderTmpl, Data: a},
		}
		s = append(s, as...)

		if a.HasPayload() {
			s = append(s, &codegen.Section{Template: serverDecoderTmpl, Data: a})
		}

		if a.HasErrors() {
			s = append(s, &codegen.Section{Template: serverErrorEncoderTmpl, Data: a})
		}
	}
	return s
}

func (e *serverWriter) Sections() []*codegen.Section {
	return e.sections
}

func (e *serverWriter) OutputPath(reserved map[string]bool) string {
	return files.UniquePath(e.outputPath, reserved)
}

func buildServerData(r *rest.ResourceExpr) *serverData {

	serviceName := codegen.Goify(r.Name, true)
	sd := &serverData{
		ServiceName: serviceName,

		HandlersStruct: fmt.Sprintf("%sHandlers", serviceName),
		Constructor:    fmt.Sprintf("New%sHandlers", serviceName),
		MountHandlers:  fmt.Sprintf("Mount%sHandlers", serviceName),
	}

	for _, a := range r.Actions {
		endpointName := codegen.Goify(a.Name, true)

		routes := make([]*serverRouteData, len(a.Routes))
		for i, v := range a.Routes {
			routes[i] = &serverRouteData{
				Method: strings.ToUpper(v.Method),
				Path:   v.FullPath(),
			}
		}

		responses := make([]*serverResponseData, len(a.Responses))
		for i, v := range a.Responses {
			hasBody := v.Body != nil && v.Body.Type != design.Empty
			responses[i] = &serverResponseData{
				Name: fmt.Sprintf("%s%s",
					codegen.Goify(serviceName, true),
					codegen.Goify(http.StatusText(v.StatusCode), true),
				),
				StatusCode: statusCodeToHttpConst(v.StatusCode),
				HasBody:    hasBody,
			}
		}

		httpErrors := make([]*serverResponseData, len(a.HTTPErrors))
		for i, v := range a.HTTPErrors {
			httpErrors[i] = &serverResponseData{
				Name:       codegen.Goify(v.Name, true),
				StatusCode: statusCodeToHttpConst(v.Response.StatusCode),
			}
		}

		ad := &serverActionData{
			EndpointName: endpointName,
			ServiceName:  serviceName,
			Routes:       routes,
			Responses:    responses,
			HTTPErrors:   httpErrors,

			MountHandler: fmt.Sprintf("Mount%s%sHandler", endpointName, serviceName),
			Constructor:  fmt.Sprintf("New%s%sHandler", endpointName, serviceName),
			Decoder:      fmt.Sprintf("%s%sDecodeRequest", endpointName, serviceName),
			Encoder:      fmt.Sprintf("%s%sEncodeResponse", endpointName, serviceName),
			ErrorEncoder: fmt.Sprintf("%s%sEncodeError", endpointName, serviceName),
		}

		if a.Payload != nil && a.Payload != design.Empty {
			hasBody := a.Body != nil && a.Body.Type != design.Empty
			ad.Payload = &serverPayloadData{
				Name:        fmt.Sprintf("%s%sPayload", endpointName, serviceName),
				Constructor: fmt.Sprintf("New%s%sPayload", endpointName, serviceName),
				Body:        fmt.Sprintf("%s%sBody", endpointName, serviceName),
				hasBody:     hasBody,
				PathParams:  extractParams(a.PathParams()),
				QueryParams: extractParams(a.QueryParams()),
				AllParams:   extractParams(a.AllParams()),
			}
		}

		sd.ActionData = append(sd.ActionData, ad)
	}

	return sd
}

func extractParams(a *design.AttributeExpr) []*serverParamData {
	obj := design.AsObject(a.Type)
	keys := make([]string, len(obj))
	i := 0
	for n := range obj {
		keys[i] = n
		i++
	}
	sort.Strings(keys)
	params := make([]*serverParamData, len(obj))
	for i, name := range keys {
		params[i] = &serverParamData{
			Name:     name,
			Type:     obj[name].Type,
			Required: true,
		}
	}

	return params
}

func (d *serverActionData) HasPayload() bool {
	return d.Payload != nil
}

func (d *serverActionData) HasErrors() bool {
	return len(d.HTTPErrors) > 0
}

func (d *serverPayloadData) HasBody() bool {
	return d.hasBody
}

func (d *serverPayloadData) HasPathParams() bool {
	return len(d.PathParams) > 0
}

func (d *serverPayloadData) HasQueryParams() bool {
	return len(d.QueryParams) > 0
}

func (d *serverPayloadData) HasParams() bool {
	return d.HasPathParams() || d.HasQueryParams()
}
