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
		VarName  string
		Type     design.DataType
		Required bool
	}

	serverHeaderData struct {
		Name string
		MapFrom string
		Value string
	}

	// serverFile
	serverFile struct {
		resource *rest.ResourceExpr
	}
)

var (
	serverTmpl = template.New("server").Funcs(template.FuncMap{
		"goTypeRef": codegen.GoTypeRef,
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
	return &serverFile{r}
}

func (e *serverFile) Sections(genPkg string) []*codegen.Section {
	d := buildServerData(e.resource)

	title := fmt.Sprintf("%s server HTTP transport", e.resource.Name)
	s := []*codegen.Section{
		codegen.Header(title, "http", []*codegen.ImportSpec{
			{Path: "fmt"},
			{Path: "io"},
			{Path: "net/http"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "github.com/dimfeld/httptreemux"},
			{Path: "goa.design/goa.v2"},
			{Path: "goa.design/goa.v2/rest"},
			{Path: genPkg + "/endpoints"},
			{Path: genPkg + "/services"},
		}),
		{Template: serverStructTmpl, Data: d},
		{Template: serverConstructorTmpl, Data: d},
		{Template: serverMountTmpl, Data: d},
	}

	for _, a := range d.ActionData {
		as := []*codegen.Section{
			{Template: serverHandlerTmpl, Data: a},
			{Template: serverHandlerConstructorTmpl, Data: a},
		}
		s = append(s, as...)

		if a.HasResponses() {
			s = append(s, &codegen.Section{Template: serverEncoderTmpl, Data: a})
		}

		if a.HasPayload() {
			s = append(s, &codegen.Section{Template: serverDecoderTmpl, Data: a})
		}

		if a.HasErrors() {
			s = append(s, &codegen.Section{Template: serverErrorEncoderTmpl, Data: a})
		}
	}
	return s
}

func (e *serverFile) OutputPath(reserved map[string]bool) string {
	name := fmt.Sprintf("transport/http/%s_server%%d.go", codegen.SnakeCase(e.resource.Name))
	return files.UniquePath(name, reserved)
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
				StatusCode: statusCodeToHTTPConst(v.StatusCode),
				HasBody:    hasBody,
			}
		}

		httpErrors := make([]*serverResponseData, len(a.HTTPErrors))
		for i, v := range a.HTTPErrors {
			httpErrors[i] = &serverResponseData{
				Name:       codegen.Goify(v.Name, true),
				StatusCode: statusCodeToHTTPConst(v.Response.StatusCode),
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
			VarName:  codegen.Goify(name, false),
			Type:     obj[name].Type,
			Required: true,
		}
	}

	return params
}

func (d *serverActionData) HasResponses() bool {
	return len(d.Responses) >= 1
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
		{{- if .HasPayload }}
		decodeRequest  = {{ .Decoder }}(dec)
		{{- end }}
		{{- if .HasResponses }}
		encodeResponse = {{ .Encoder }}EncodeResponse(enc)
		{{- end }}
		encodeError    = {{ if .HasErrors }}{{ .ErrorEncoder }}{{ else }}EncodeError{{ end }}(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		{{- if .HasPayload }}
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
		{{- if .HasResponses }}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
		{{- else }}
		w.Write(http.StatusNoContent)
		{{- end }}
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
			{{ .VarName }} {{goTypeRef .Type }}
			{{- end }}
		)
{{ range .Payload.QueryParams }}
	{{- if eq .Type.Name "string" }}
		{{ .VarName }} = r.URL.Query().Get("{{ .Name }}")
	{{- else }}
		{{ .VarName }}Raw := r.URL.Query().Get("{{ .Name }}")
		{{- template "conversion" . }}
	{{- end }}
{{ end }}
{{- range .Payload.PathParams }}
	{{- if eq .Type.Name "string" }}
		{{ .VarName }} = params["{{ .Name }}"]
	{{- else }}
		{{ .VarName }}Raw := params["{{ .Name }}"]
		{{- template "conversion" . }}
	{{- end }}
{{ end }}
{{- end }}
		payload, err := {{ .Payload.Constructor }}(
			{{- if .Payload.HasBody }}&body{{ end -}}
			{{- range $i, $p := .Payload.AllParams }}
				{{- if or (ne $i 0) ($.Payload.HasBody) }}, {{ end -}}{{ .VarName }}
			{{- end }})
		return payload, err
	}
}
{{- define "conversion" }}
	{{- if eq .Type.Name "array" }}
		{{ .VarName }}RawSlice := strings.Split({{ .VarName }}Raw, ",")
		{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}RawSlice))
		for i, rv := range {{ .VarName }}RawSlice {
			{{- template "type_slice_conversion" . }}
		}
	{{- else }}
		{{- template "type_conversion" . }}
	{{- end }}
{{- end }}
{{- define "type_conversion" }}
	{{- if eq .Type.Name "string" }}
		{{ .VarName }} = url.QueryUnescape(v)
	{{- else if eq .Type.Name "int" }}
		if v, err := strconv.ParseInt({{ .VarName }}Raw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an integer, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = int(v)
		}
	{{- else if eq .Type.Name "int32" }}
		if v, err := strconv.ParseInt({{ .VarName }}Raw, 10, 32); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an integer, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = int32(v)
		}
	{{- else if eq .Type.Name "int64" }}
		if v, err := strconv.ParseInt({{ .VarName }}Raw, 10, 64); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an integer, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = v
		}
	{{- else if eq .Type.Name "uint" }}
		if v, err := strconv.ParseUint({{ .VarName }}Raw, 10, strconv.IntSize); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = uint(v)
		}
	{{- else if eq .Type.Name "uint32" }}
		if v, err := strconv.ParseUint({{ .VarName }}Raw, 10, 32); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = int32(v)
		}
	{{- else if eq .Type.Name "uint64" }}
		if v, err := strconv.ParseUint({{ .VarName }}Raw, 10, 64); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = v
		}
	{{- else if eq .Type.Name "float32" }}
		if v, err := strconv.ParseFloat({{ .VarName }}Raw, 32); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be a float, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = float32(v)
		}
	{{- else if eq .Type.Name "float64" }}
		if v, err := strconv.ParseFloat({{ .VarName }}Raw, 64); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be a float, got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = v
		}
	{{- else if eq .Type.Name "boolean" }}
		if v, err := strconv.ParseBool({{ .VarName }}Raw); err != nil {
			return nil, fmt.Errorf("{{ .Name }} must be a boolean (true or false), got '%s'", {{ .VarName }}Raw)
		} else {
			{{ .VarName }} = v
		}
	{{- else }}
		// unsupported type {{ .Type.Name }} for var {{ .VarName }}
	{{- end }}
{{- end }}
{{- define "type_slice_conversion" }}
		{{- if eq .Type.ElemType.Type.Name "string" }}
			{{ .VarName }}[i] = url.QueryUnescape(rv)
		{{- else if eq .Type.ElemType.Type.Name "int" }}
			if v, err := strconv.ParseInt(rv, 10, strconv.IntSize); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of integers, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = int(v)
			}
		{{- else if eq .Type.ElemType.Type.Name "int32" }}
			if v, err := strconv.ParseInt(rv, 10, 32); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of integers, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = int32(v)
			}
		{{- else if eq .Type.ElemType.Type.Name "int64" }}
			if v, err := strconv.ParseInt(rv, 10, 64); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of integers, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = v
			}
		{{- else if eq .Type.ElemType.Type.Name "uint" }}
			if v, err := strconv.ParseUint(rv, 10, strconv.IntSize); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of unsigned integers, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = uint(v)
			}
		{{- else if eq .Type.ElemType.Type.Name "uint32" }}
			if v, err := strconv.ParseUint(rv, 10, 32); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of unsigned integers, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = int32(v)
			}
		{{- else if eq .Type.ElemType.Type.Name "uint64" }}
			if v, err := strconv.ParseUint(rv, 10, 64); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of unsigned integers, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = v
			}
		{{- else if eq .Type.ElemType.Type.Name "float32" }}
			if v, err := strconv.ParseFloat(rv, 32); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of floats, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = float32(v)
			}
		{{- else if eq .Type.ElemType.Type.Name "float64" }}
			if v, err := strconv.ParseFloat(rv, 64); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of floats, got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = v
			}
		{{- else if eq .Type.ElemType.Type.Name "boolean" }}
			if v, err := strconv.ParseBool(rv); err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an set of booleans (true, false, 1 or 0), got value '%s' in set '%s'", rv, {{ .VarName }}Raw)
			} else {
				{{ .VarName }}[i] = v
			}
		{{- else }}
			// unsupported slice type {{ .Type.ElemType.Type.Name }} for var {{ .VarName }}
		{{- end }}
{{- end }}
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
