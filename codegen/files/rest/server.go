package rest

import (
	"fmt"
	"text/template"

	"net/http"
	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

type (
	// serverActionData describes a single endpoint/resource.
	serverData struct {
		// ServiceName is the name of the service.
		ServiceName string
		// VarServiceName is the goified name of the service.
		VarServiceName string
		// HandlerStruct is the name of the main server handler structure.
		HandlersStruct string
		// Constructor is the name of the constructor of the handler struct function.
		Constructor string
		// MountHandlers is the name of the name of the mount function.
		MountHandlers string
		// Action describes the action data for this endpoint.
		ActionData []*serverActionData
	}

	// serverActionData describes a single action.
	serverActionData struct {
		// EndpointName is the name of the endpoint.
		EndpointName string
		// VarEndpointName is the goified name of theendpoint/resource.
		VarEndpointName string
		// ServiceName is the name of the service.
		ServiceName string
		// VarServiceName is the goified name of the service.
		VarServiceName string
		// Routes describes the possible routes for this action.
		Routes []*serverRouteData

		// MountHandler is the name of the mount handler function.
		MountHandler string
		// Constructor is the name of the constructor function for the http handler function.
		Constructor string
		// Decoder is the name of the decoder function.
		Decoder string
		// Encoder is the name of the encoder function.
		Encoder string
		// ErrorEncoder is the name of the error encoder function.
		ErrorEncoder string
		// Payload provides information about the payload.
		Payload *serverPayloadData
		// Responses describes the information about the different responses.
		Responses []*serverResponseData
		// HTTPErrors describes the information about error responses.
		HTTPErrors []*serverResponseData
	}

	// serverPayloadData describes a payload.
	serverPayloadData struct {
		// Name is the name of the payload structure.
		Name string
		// Constructor is the name of the payload constructor function.
		Constructor string
		// Body is the name of the body structure.
		Body string
		// HasBody indicate that the payload expects a body.
		HasBody bool
		// PathParams describes the information about params that are present in the path.
		PathParams []*serverParamData
		// QueryParams describes the information about the params that are present in the query.
		QueryParams []*serverParamData
		// AllParams describes the params, in path and query.
		AllParams []*serverParamData
	}

	// serverRouteData describes a route.
	serverRouteData struct {
		// Method is the HTTP method.
		Method string
		// Path is the full path.
		Path string
	}

	// serverResponseData describes a response.
	serverResponseData struct {
		// Name is the name of the response structure.
		Name string
		// StatusCode is the return code of the response.
		StatusCode string
		// HasBody indicates that the response will return data in the body.
		HasBody bool
		// Headers provides information about the headers in the response.
		Headers []*serverHeaderData
	}

	// serverParamData describes a parameter.
	serverParamData struct {
		// Name is the name of the mapping to the actual variable name.
		Name string
		// VarName is the name of the variable.
		VarName string
		// Type is the datatype of the variable.
		Type design.DataType
	}

	// serverHeaderData describes a header.
	serverHeaderData struct {
		// Name describes the name of the header key.
		Name string
		// VarName describes the variable name where the value for the header is obtained.
		VarName string
		// Type describes the datatype of the variable value. Mainly used for conversion.
		Type design.DataType
	}

	// serverFile is the codegenerator file for the given server.
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
	varServiceName := codegen.Goify(r.Name, true)
	sd := &serverData{
		ServiceName:    r.Name,
		VarServiceName: varServiceName,
		HandlersStruct: fmt.Sprintf("%sHandlers", varServiceName),
		Constructor:    fmt.Sprintf("New%sHandlers", varServiceName),
		MountHandlers:  fmt.Sprintf("Mount%sHandlers", varServiceName),
	}

	for _, a := range r.Actions {
		varEndpointName := codegen.Goify(a.Name, true)

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
					varServiceName,
					codegen.Goify(http.StatusText(v.StatusCode), true),
				),
				StatusCode: statusCodeToHTTPConst(v.StatusCode),
				HasBody:    hasBody,
				Headers:    extractHeaders(v.MappedHeaders()),
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
			EndpointName:    a.Name,
			VarEndpointName: varEndpointName,
			ServiceName:     r.Name,
			VarServiceName:  varServiceName,
			Routes:          routes,
			Responses:       responses,
			HTTPErrors:      httpErrors,
			MountHandler:    fmt.Sprintf("Mount%s%sHandler", varEndpointName, varServiceName),
			Constructor:     fmt.Sprintf("New%s%sHandler", varEndpointName, varServiceName),
			Decoder:         fmt.Sprintf("%s%sDecodeRequest", varEndpointName, varServiceName),
			Encoder:         fmt.Sprintf("%s%sEncodeResponse", varEndpointName, varServiceName),
			ErrorEncoder:    fmt.Sprintf("%s%sEncodeError", varEndpointName, varServiceName),
		}

		if a.Payload != nil && a.Payload != design.Empty {
			hasBody := a.Body != nil && a.Body.Type != design.Empty
			ad.Payload = &serverPayloadData{
				Name:        fmt.Sprintf("%s%sPayload", varEndpointName, varServiceName),
				Constructor: fmt.Sprintf("New%s%sPayload", varEndpointName, varServiceName),
				Body:        fmt.Sprintf("%s%sBody", varEndpointName, varServiceName),
				HasBody:     hasBody,
				PathParams:  extractParams(a.PathParams()),
				QueryParams: extractParams(a.QueryParams()),
				AllParams:   extractParams(a.AllParams()),
			}
		}
		sd.ActionData = append(sd.ActionData, ad)
	}
	return sd
}

func extractHeaders(a *rest.MappedAttributeExpr) []*serverHeaderData {
	var headers []*serverHeaderData
	WalkMappedAttr(a, func(name, elem string, required bool, a *design.AttributeExpr) error {
		headers = append(headers, &serverHeaderData{
			Name:    elem,
			VarName: codegen.Goify(name, true),
			Type:    a.Type,
		})
		return nil
	})

	return headers
}

func extractParams(a *rest.MappedAttributeExpr) []*serverParamData {
	var params []*serverParamData
	WalkMappedAttr(a, func(name, elem string, required bool, a *design.AttributeExpr) error {
		params = append(params, &serverParamData{
			Name:    elem,
			VarName: codegen.Goify(name, true),
			Type:    a.Type,
		})
		return nil
	})

	return params
}

// HasResponses indicates if an action has responses.
func (d *serverActionData) HasResponses() bool {
	return len(d.Responses) >= 1
}

// HasPayload indicates if an action has a payload.
func (d *serverActionData) HasPayload() bool {
	return d.Payload != nil
}

// HasErrors indicates if an action has errors defined.
func (d *serverActionData) HasErrors() bool {
	return len(d.HTTPErrors) > 0
}

// HasPathParams indicates if a payload has path parameters.
func (d *serverPayloadData) HasPathParams() bool {
	return len(d.PathParams) > 0
}

// HasQueryParams indicates if a payload has query parameters.
func (d *serverPayloadData) HasQueryParams() bool {
	return len(d.QueryParams) > 0
}

// HasParams indicates if a payload has any parameters.
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
	e *endpoints.{{ .VarServiceName }},
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) *{{ .HandlersStruct }} {
	return &{{ .HandlersStruct }}{
		{{- range .ActionData }}
		{{ .VarEndpointName }}: {{ .Constructor }}(e.{{ .VarEndpointName }}, dec, enc, logger),
		{{- end }}
	}
}
`

const serverMountT = `{{ printf "%s configures the mux to serve the %s endpoints." .MountHandlers .ServiceName | comment }}
func {{ .MountHandlers }}(mux rest.ServeMux, h *{{ .HandlersStruct }}) {
	{{- range .ActionData }}
	{{ .MountHandler }}(mux, h.{{ .VarEndpointName }})
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
	{{- else if eq .Type.Name "[]byte" }}
		{{ .VarName }} = url.QueryUnescape(string(v))
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
		{{- else if eq .Type.ElemType.Type.Name "[]byte" }}
			{{ .VarName }}[i] = url.QueryUnescape(string(rv))
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
	{{- if eq (len .Responses) 1 }}
		{{- with index .Responses 0}}{{/* TBD: support multiple responses */}}
		{{- if or .HasBody .Headers }}
		t := v.(*{{ .Name }})
		{{- end }}

		{{- if .HasBody }}
		w.Header().Set("Content-Type", ResponseContentType(r))
		{{- end }}

		{{- range .Headers }}

		{{- if not .Required }}
		if t.{{ .VarName }} != nil {
		{{- end }}

		{{- if eq .Type.Name "string" }}
		w.Header().Set("{{ .Name }}", {{ if not .Required }}*{{ end }}t.{{ .VarName }})
		{{- else }}
		v := t.{{ .VarName }}
		{{ template "conversion" . }}
		w.Header().Set("{{ .Name }}", tmp{{ .VarName }})
		{{- end }}

		{{- if not .Required }}
		}
		{{- end }}

		{{- end }}

		w.WriteHeader({{ .StatusCode }})
		{{- if .HasBody }}
		if t != nil {
			return encoder(w, r).Encode(t)
		}
		{{- end }}
		{{- end }}
	{{- else }}
		switch t := v.(type) {
		{{- range .Responses }}
		case *{{ .Name }}:

			{{- range .Headers }}

			{{- if eq .Type.Name "string" }}
			w.Header().Set("{{ .Name }}", {{ if not .Required }}*{{ end }}t.{{ .VarName }})
			{{- else }}
			v := t.{{ .VarName }}
			{{ template "conversion" . }}
			w.Header().Set("{{ .Name }}", tmp{{ .VarName }})
			{{- end }}

			{{- end }}

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
{{- define "conversion" }}
	{{- if eq .Type.Name "boolean" }}
		tmp{{ .VarName }} := strconv.FormatBool({{ if not .Required }}*{{ end }}v)
	{{- else if eq .Type.Name "int" }}
		tmp{{ .VarName }} := strconv.Itoa({{ if not .Required }}*{{ end }}v)
	{{- else if eq .Type.Name "int32" }}
		tmp{{ .VarName }} := strconv.FormatInt(int64({{ if not .Required }}*{{ end }}v), 10)
	{{- else if eq .Type.Name "int64" }}
		tmp{{ .VarName }} := strconv.FormatInt({{ if not .Required }}*{{ end }}v, 10)
	{{- else if eq .Type.Name "uint" }}
		tmp{{ .VarName }} := strconv.FormatUint(uint64({{ if not .Required }}*{{ end }}v), 10)
	{{- else if eq .Type.Name "uint32" }}
		tmp{{ .VarName }} := strconv.FormatUint(uint64({{ if not .Required }}*{{ end }}v), 10)
	{{- else if eq .Type.Name "uint64" }}
		tmp{{ .VarName }} := strconv.FormatUint({{ if not .Required }}*{{ end }}v, 10)
	{{- else if eq .Type.Name "float32" }}
		tmp{{ .VarName }} := strconv.FormatFloat(float64({{ if not .Required }}*{{ end }}v), 'f', -1, 32)
	{{- else if eq .Type.Name "float64" }}
		tmp{{ .VarName }} := strconv.FormatFloat({{ if not .Required }}*{{ end }}v, 'f', -1, 64)
	{{- else if eq .Type.Name "string" }}
		tmp{{ .VarName }} := v
	{{- else if eq .Type.Name "[]byte" }}
		tmp{{ .VarName }} := string({{ if not .Required }}*{{ end }}v)
	{{- else }}
		// unsupported type {{ .Type.Name }} for var {{ .VarName }}
	{{- end }}
{{- end }}`

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
