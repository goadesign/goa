// MarshalTypes codegen file
// add tags to generated type from restgen.GoTypeDef
// Make sure to generate user type for each error even if declared with primitive
// Add metadata to specify error type attribute to be used for error message
// validate only one error per http status code
// Make sure primitive rename of []byte to bytes didn't break stuff

// DSL validation: make sure there's at least one response for all actions
// Make sure all routes define identical path params
// Remove required attributes from 'Required' slice that have default values

// Add response tags to account example
// Test response tags
// Test default values and header decoding

package rest

import (
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"

	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	restgen "goa.design/goa.v2/codegen/rest"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

type (
	// ServerData describes a single endpoint/resource.
	ServerData struct {
		// ServiceName is the name of the service.
		ServiceName string
		// ServiceVarName is the goified name of the service.
		ServiceVarName string
		// HandlerStruct is the name of the main server handler structure.
		HandlersStruct string
		// Constructor is the name of the constructor of the handler struct function.
		Constructor string
		// MountHandlers is the name of the name of the mount function.
		MountHandlers string
		// Action describes the action data for this endpoint.
		ActionData []*ServerActionData
	}

	// ServerActionData describes a single action.
	ServerActionData struct {
		// EndpointName is the name of the endpoint.
		EndpointName string
		// EndpointVarName is the goified name of theendpoint/resource.
		EndpointVarName string
		// ServiceName is the name of the service.
		ServiceName string
		// ServiceVarName is the goified name of the service.
		ServiceVarName string
		// Routes describes the possible routes for this action.
		Routes []*ServerRouteData
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
		Payload *ServerPayloadData
		// ResultTypeRef is the service endpoint result type reference
		ResultTypeRef string
		// Responses describes the information about the different
		// responses. If there are more than one responses then the
		// tagless response must be last.
		Responses []*ServerResponseData
		// HTTPErrors describes the information about error responses.
		HTTPErrors []*ServerErrorData
	}

	// ServerPayloadData describes a payload.
	ServerPayloadData struct {
		// Ref is the reference to the payload type.
		Ref string
		// Constructor is the name of the payload constructor function.
		Constructor string
		// BodyTypeName is the name of the request body type if any.
		BodyTypeName string
		// BodyRef is a reference to the body variable as returned by the
		// decode function (i.e. "body" or "&body")
		BodyRef string
		// PathParams describes the information about params that are
		// present in the path.
		PathParams []*ServerParamData
		// QueryParams describes the information about the params that
		// are present in the query.
		QueryParams []*ServerParamData
		// AllParams describes the params, in path and query.
		AllParams []*ServerParamData
		// ConstructorParams is a string representing the Go code for the
		// section of the payload constructor signature that consists of
		// passing all the parameter and header values.
		ConstructorParams string
		// Headers contains the HTTP request headers used to build the
		// endpoint payload.
		Headers []*ServerHeaderData
		// ValidateBody contains the body validation code if any.
		ValidateBody string
	}

	// ServerRouteData describes a route.
	ServerRouteData struct {
		// Method is the HTTP method.
		Method string
		// Path is the full path.
		Path string
	}

	// ServerResponseData describes a response.
	ServerResponseData struct {
		// Body is the type of the response body, nil if body should be
		// empty.
		Body design.DataType
		// BodyUserTypeName is the name of the Body type if it is a user
		// or media type, the empty string otherwise.
		BodyUserTypeName string
		// StatusCode is the return code of the response.
		StatusCode string
		// Headers provides information about the headers in the response.
		Headers []*ServerHeaderData
		// BodyFields is the list of the response body type attributes
		// used to initialize the response body. Not needed if the
		// response type can be assigned to directly from the endpoint
		// result type (i.e. if the response body has all the result
		// attributes).
		BodyFields []string
		// TagName is the name of the attribute used to test whether the
		// response is the one to use.
		TagName string
		// TagValue is the value the result attribute named by TagName
		// must have for this response to be used.
		TagValue string
	}

	// ServerErrorData describes a error response.
	ServerErrorData struct {
		// TypeRef is a reference to the user type.
		TypeRef string
		// Response is the error response data.
		Response *ServerResponseData
	}

	// ServerParamData describes a parameter.
	ServerParamData struct {
		// Name is the name of the mapping to the actual variable name.
		Name string
		// FieldName is the name of the struct field that holds the
		// param value.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the param value.
		VarName string
		// Required is true if the param is required.
		Required bool
		// Type is the datatype of the variable.
		Type design.DataType
		// Pointer is true if and only the param variable is a pointer.
		Pointer bool
		// StringSlice is true if the param type is array of strings.
		StringSlice bool
		// Slice is true if the param type is an array.
		Slice bool
		// Validate contains the validation code if any.
		Validate string
		// DefaultValue contains the default value if any.
		DefaultValue interface{}
	}

	// ServerHeaderData describes a header.
	ServerHeaderData struct {
		// Name describes the name of the header key.
		Name string
		// FieldName is the name of the struct field that holds the
		// header value.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the header value.
		VarName string
		// Required is true if the header is required.
		Required bool
		// Type describes the datatype of the variable value. Mainly used for conversion.
		Type design.DataType
	}
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
	sections := func(genPkg string) []*codegen.Section {
		d := buildServiceData(r)

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
			{Template: serverStructTmpl(r), Data: d},
			{Template: serverConstructorTmpl(r), Data: d},
			{Template: serverMountTmpl(r), Data: d},
		}

		for _, a := range d.ActionData {
			as := []*codegen.Section{
				{Template: serverHandlerTmpl(r), Data: a},
				{Template: serverHandlerConstructorTmpl(r), Data: a},
			}
			s = append(s, as...)

			if a.HasResponses() {
				s = append(s, &codegen.Section{Template: serverEncoderTmpl(r), Data: a})
			}

			if a.HasPayload() {
				s = append(s, &codegen.Section{Template: serverDecoderTmpl(r), Data: a})
			}

			if a.HasErrors() {
				s = append(s, &codegen.Section{Template: serverErrorEncoderTmpl(r), Data: a})
			}
		}
		return s
	}

	return codegen.NewSource(path, sections)
}

func serverTmpl(r *rest.ResourceExpr) *template.Template {
	scope := files.Services.Get(r.Name()).Scope
	return template.New("server").
		Funcs(template.FuncMap{"goTypeRef": scope.GoTypeRef}).
		Funcs(codegen.TemplateFuncs())
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

func buildServiceData(r *rest.ResourceExpr) *ServerData {
	svc := files.Services.Get(r.ServiceExpr.Name)

	sd := &ServerData{
		ServiceName:    svc.Name,
		ServiceVarName: svc.VarName,
		HandlersStruct: fmt.Sprintf("%sHandlers", svc.VarName),
		Constructor:    fmt.Sprintf("New%sHandlers", svc.VarName),
		MountHandlers:  fmt.Sprintf("Mount%sHandlers", svc.VarName),
	}

	for _, a := range r.Actions {
		routes := make([]*ServerRouteData, len(a.Routes))
		for i, r := range a.Routes {
			routes[i] = &ServerRouteData{
				Method: strings.ToUpper(r.Method),
				Path:   r.FullPath(),
			}
		}

		var responses []*ServerResponseData
		notag := -1
		for i, v := range a.Responses {
			if v.Tag[0] == "" {
				if notag > -1 {
					continue // we don't want more than one response with no tag
				}
				notag = i
			}
			responses = append(responses, buildResponseData(r, a, v))
		}
		count := len(responses)
		if notag >= 0 && notag < count-1 {
			// Make sure tagless response is last
			responses[notag], responses[count-1] = responses[count-1], responses[notag]
		}

		httperrs := make([]*ServerErrorData, len(a.HTTPErrors))
		for i, v := range a.HTTPErrors {
			httperrs[i] = buildErrorData(r, a, v)
		}

		ep := svc.Method(a.EndpointExpr.Name)

		ad := &ServerActionData{
			EndpointName:    ep.Name,
			EndpointVarName: ep.VarName,
			ServiceName:     svc.Name,
			ServiceVarName:  svc.VarName,
			Routes:          routes,
			ResultTypeRef:   ep.ResultRef,
			Responses:       responses,
			HTTPErrors:      httperrs,
			MountHandler:    fmt.Sprintf("Mount%sHandler", ep.VarName),
			Constructor:     fmt.Sprintf("New%sHandler", ep.VarName),
			Decoder:         fmt.Sprintf("Decode%sRequest", ep.VarName),
			Encoder:         fmt.Sprintf("Encode%sResponse", ep.VarName),
			ErrorEncoder:    fmt.Sprintf("Encode%sError", ep.VarName),
		}

		if ep.Payload != "" {
			var (
				constructor  string
				validateBody string
				bodyTypeName string
				bodyRef      string
				params       []string
			)
			{
				body := restgen.RequestBodyType(r, a, "ServerRequestBody")
				if ut, ok := a.EndpointExpr.Payload.Type.(design.UserType); ok {
					if ut != body {
						constructor = fmt.Sprintf("New%s", ep.Payload)
					}
				}
				if body != design.Empty {
					{
						if ut, ok := body.(design.UserType); ok {
							if codegen.HasValidations(ut.Attribute(), false) {
								validateBody = `
		if err := body.Validate(); err != nil {
			return nil, err
		}`
							}
						} else {
							code := codegen.RecursiveValidationCode(a.EndpointExpr.Payload, true, true, "body")
							if code != "" {
								validateBody = "\n" + code + `
		if err != nil {
			return nil, err
		}`
							}
						}
						bodyTypeName = svc.Scope.GoTypeName(body)
						bodyRef = "body"
						if design.IsObject(body) {
							bodyRef = "&" + bodyRef
						}
					}
					params = []string{bodyRef}
				}
				restgen.WalkMappedAttr(a.AllParams(), func(_, elem string, _ bool, _ *design.AttributeExpr) error {
					params = append(params, codegen.Goify(elem, false))
					return nil
				})
				restgen.WalkMappedAttr(a.MappedHeaders(), func(_, elem string, _ bool, _ *design.AttributeExpr) error {
					params = append(params, codegen.Goify(elem, false))
					return nil
				})
			}
			ad.Payload = &ServerPayloadData{
				Ref:               ep.PayloadRef,
				Constructor:       constructor,
				BodyTypeName:      bodyTypeName,
				BodyRef:           bodyRef,
				PathParams:        extractParams(a.PathParams()),
				QueryParams:       extractParams(a.QueryParams()),
				AllParams:         extractParams(a.AllParams()),
				Headers:           extractHeaders(a.MappedHeaders()),
				ConstructorParams: strings.Join(params, ", "),
				ValidateBody:      validateBody,
			}
		}
		sd.ActionData = append(sd.ActionData, ad)
	}
	return sd
}

func buildResponseData(r *rest.ResourceExpr, a *rest.ActionExpr, v *rest.HTTPResponseExpr) *ServerResponseData {
	var suffix, bodyTypeName string
	if len(a.Responses) > 1 {
		suffix = http.StatusText(v.StatusCode)
	}
	body := restgen.ResponseBodyType(r, v, a.EndpointExpr.Result, suffix)
	if body != nil {
		if ut, ok := body.(design.UserType); ok {
			bodyTypeName = ut.Name()
		}
	}
	var bodyFields []string
	if design.IsObject(a.EndpointExpr.Result.Type) && design.IsObject(body) {
		if body != a.EndpointExpr.Result.Type {
			codegen.WalkAttributes(design.AsObject(body), func(name string, att *design.AttributeExpr) error {
				bodyFields = append(bodyFields, codegen.GoifyAtt(att, name, true))
				return nil
			})
		}
	}
	return &ServerResponseData{
		Body:             body,
		BodyUserTypeName: bodyTypeName,
		StatusCode:       restgen.StatusCodeToHTTPConst(v.StatusCode),
		Headers:          extractHeaders(v.MappedHeaders()),
		BodyFields:       bodyFields,
		TagName:          v.Tag[0],
		TagValue:         v.Tag[1],
	}
}

func buildErrorData(r *rest.ResourceExpr, a *rest.ActionExpr, v *rest.HTTPErrorExpr) *ServerErrorData {
	var bodyTypeName string
	body := restgen.ResponseBodyType(r, v.Response, v.ErrorExpr.AttributeExpr, http.StatusText(v.Response.StatusCode))
	if body != nil {
		if ut, ok := body.(design.UserType); ok {
			bodyTypeName = ut.Name()
		}
	}
	var bodyFields []string
	if design.IsObject(v.ErrorExpr.Type) && design.IsObject(body) {
		if body != v.ErrorExpr.Type {
			codegen.WalkAttributes(design.AsObject(body), func(name string, att *design.AttributeExpr) error {
				bodyFields = append(bodyFields, codegen.GoifyAtt(att, name, true))
				return nil
			})
		}
	}
	response := ServerResponseData{
		Body:             body,
		BodyUserTypeName: bodyTypeName,
		StatusCode:       restgen.StatusCodeToHTTPConst(v.Response.StatusCode),
		Headers:          extractHeaders(v.Response.MappedHeaders()),
		BodyFields:       bodyFields,
	}
	return &ServerErrorData{
		TypeRef:  files.Services.Get(r.Name()).Scope.GoTypeRef(v.ErrorExpr.Type),
		Response: &response,
	}
}

func extractHeaders(a *rest.MappedAttributeExpr) []*ServerHeaderData {
	var headers []*ServerHeaderData
	restgen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		isNativePointer := c.Type.Kind() == design.BytesKind || c.Type.Kind() == design.AnyKind
		headers = append(headers, &ServerHeaderData{
			Name:      elem,
			FieldName: codegen.Goify(name, true),
			VarName:   codegen.Goify(name, false),
			Required:  required || isNativePointer,
			Type:      c.Type,
		})
		return nil
	})

	return headers
}

func extractParams(a *rest.MappedAttributeExpr) []*ServerParamData {
	var params []*ServerParamData
	restgen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			field           = codegen.Goify(name, true)
			varn            = codegen.Goify(name, false)
			arr             = design.AsArray(c.Type)
			isNativePointer = c.Type.Kind() == design.BytesKind || c.Type.Kind() == design.AnyKind
		)
		params = append(params, &ServerParamData{
			Name:         elem,
			FieldName:    field,
			VarName:      varn,
			Required:     required,
			Type:         c.Type,
			Pointer:      !required && design.IsPrimitive(c.Type) && !isNativePointer,
			StringSlice:  arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Slice:        arr != nil,
			Validate:     codegen.RecursiveValidationCode(c, a.IsRequired(name) || isNativePointer, false, varn),
			DefaultValue: c.DefaultValue,
		})
		return nil
	})

	return params
}

// HasResponses indicates if an action has responses.
func (d *ServerActionData) HasResponses() bool {
	return len(d.Responses) >= 1
}

// HasPayload indicates if an action has a payload.
func (d *ServerActionData) HasPayload() bool {
	return d.Payload != nil
}

// HasErrors indicates if an action has errors defined.
func (d *ServerActionData) HasErrors() bool {
	return len(d.HTTPErrors) > 0
}

// HasPathParams indicates if a payload has path parameters.
func (d *ServerPayloadData) HasPathParams() bool {
	return len(d.PathParams) > 0
}

const serverStructT = `{{ printf "%s lists the %s service endpoint HTTP handlers." .HandlersStruct .ServiceName | comment }}
type {{ .HandlersStruct }} struct {
	{{- range .ActionData }}
	{{ .EndpointVarName }} http.Handler
	{{- end }}
}
`

const serverConstructorT = `{{ printf "%s instantiates HTTP handlers for all the %s service endpoints." .Constructor .ServiceName | comment }}
func {{ .Constructor }}(
	e *endpoints.{{ .ServiceVarName }},
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) *{{ .HandlersStruct }} {
	return &{{ .HandlersStruct }}{
		{{- range .ActionData }}
		{{ .EndpointVarName }}: {{ .Constructor }}(e.{{ .EndpointVarName }}, dec, enc, logger),
		{{- end }}
	}
}
`

const serverMountT = `{{ printf "%s configures the mux to serve the %s endpoints." .MountHandlers .ServiceName | comment }}
func {{ .MountHandlers }}(mux rest.Muxer, h *{{ .HandlersStruct }}) {
	{{- range .ActionData }}
	{{ .MountHandler }}(mux, h.{{ .EndpointVarName }})
	{{- end }}
}
`

const serverHandlerT = `{{ printf "%s configures the mux to serve the \"%s\" service \"%s\" endpoint." .MountHandler .ServiceName .EndpointName | comment }}
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

const serverHandlerConstructorT = `{{ printf "%s creates a HTTP handler which loads the HTTP request and calls the \"%s\" service \"%s\" endpoint." .Constructor .ServiceName .EndpointName | comment }}
{{ comment "The middleware is mounted so it executes after the request is loaded and thus may access the request state via the rest package ContextXXX functions."}}
func {{ .Constructor }}(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) rest.Encoder,
	logger goa.LogAdapter,
) http.Handler {
	var (
		{{- if .HasPayload }}
		decodeRequest  = {{ .Decoder }}(dec)
		{{- end }}
		{{- if .HasResponses }}
		encodeResponse = {{ .Encoder }}EncodeResponse(enc)
		{{- end }}
		encodeError    = {{ if .HasErrors }}{{ .ErrorEncoder }}{{ else }}rest.EncodeError{{ end }}(enc, logger)
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

const serverDecoderT = `{{ printf "%s returns a decoder for requests sent to the %s %s endpoint." .Decoder .ServiceName .EndpointName | comment }}
func {{ .Decoder }}(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ({{ .Payload.Ref }}, error) {

{{- if .Payload.BodyTypeName }}
		var (
			body {{ .Payload.BodyTypeName }}
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		{{- .Payload.ValidateBody }}
{{ end }}

{{- if or .Payload.AllParams .Payload.Headers }}
		var (
		{{- range .Payload.AllParams }}
			{{ .VarName }} {{ if .Pointer }}*{{ end }}{{goTypeRef .Type }}
		{{- end }}
		{{- range .Payload.Headers }}
			{{ .VarName }} {{ if .Pointer }}*{{ end }}{{goTypeRef .Type }}
		{{- end }}
		{{- if .Payload.HasPathParams }}

			params = rest.ContextParams(r.Context())
		{{- end }}
		)

{{- range .Payload.PathParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = params["{{ .Name }}"]
		if {{ .VarName }} == "" {
			return nil, goa.MissingFieldError("{{ .Name }}", "path")
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := params["{{ .Name }}"]
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
		{{ .VarName }}Def := {{ print "%q" .DefaultValue }}
		{{ .VarName }} = {{ if eq .Type.Name "string" }}&{{ end }}{{ .VarName }}Def
	}
		{{- end }}

	{{- else }}{{/* not string */}}
		{{ .VarName }}Raw := params["{{ .Name }}"]
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			return nil, goa.MissingFieldError("{{ .Name }}", "path")
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
		{{- template "path_conversion" . }}
		{{- if or .DefaultValue (not .Required) }}
		}
		{{- end }}
	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}

{{- range .Payload.QueryParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = r.URL.Query().Get("{{ .Name }}")
		if {{ .VarName }} == "" {
			return nil, goa.MissingFieldError("{{ .Name }}", "query string")
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := r.URL.Query().Get("{{ .Name }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
			{{ .VarName }}Def := {{ printf "%q" .DefaultValue }}
			{{ .VarName }} = {{ if eq .Type.Name "string" }}&{{ end }}{{ .VarName }}Def
		}
		{{- end }}

	{{- else if .StringSlice }}
		{{ .VarName }} = r.URL.Query()["{{ .Name }}"]
		{{- if .Required }}
		if {{ .VarName }} == nil {
			return nil, goa.MissingFieldError("{{ .Name }}", "query string")
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

	{{- else }}{{/* not string, not any and not slice */}}
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
			return nil, goa.MissingFieldError("{{ .Name }}", "header")
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := r.Header.Get("{{ .Name }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if eq .Type.Name "string" }}&{{ end }}{{ .VarName }}Raw
		}
		{{- if .DefaultValue }} else {
			{{ .VarName }}Def := {{ print "%q" .DefaultValue }}
			{{ .VarName }} = {{ if eq .Type.Name "string" }}&{{ end }}{{ .VarName }}Def
		}
		{{- end }}

	{{- else if .StringSlice }}
		{{ .VarName }} = r.Header["{{ .Name }}"]
		{{ if .Required }}
		if {{ .VarName }} == nil {
			return nil, goa.MissingFieldError("{{ .Name }}", "header")
		}
		{{- else if .DefaultValue }}
		if {{ .VarName }} == nil {
			{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
		}
		{{- end }}

	{{- else if .Slice }}
		{{ .VarName }}Raw := r.Header["{{ .Name }}"]
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
		{{- if .Payload.Constructor }}
		return {{ .Payload.Constructor }}({{ .Payload.ConstructorParams }}), nil
		{{- else if .Payload.BodyRef }}
		return {{ .Payload.BodyRef }}, nil
		{{- else }}
		return {{ (index .Payload.AllParams 0).VarName }}, nil
		{{- end }}
	}
}

{{- define "path_conversion" }}
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

{{- define "slice_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for i, rv := range {{ .VarName }}Raw {
		{{- template "type_slice_conversion" . }}
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
		{{- if .Required }}
		{{ .VarName }} = int(v)
		{{- else }}
		pv := int(v)
		{{ .VarName }} = &pv
		{{- end }}
	{{- else if eq .Type.Name "int32" }}
		v, err := strconv.ParseInt({{ .VarName }}Raw, 10, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "integer")
		}
		{{- if .Required }}
		{{ .VarName }} = int32(v)
		{{- else }}
		pv := int32(v)
		{{ .VarName }} = &pv
		{{- end }}
	{{- else if eq .Type.Name "int64" }}
		v, err := strconv.ParseInt({{ .VarName }}Raw, 10, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "integer")
		}
		{{ .VarName }} = {{ if not .Required}}&{{ end }}v
	{{- else if eq .Type.Name "uint" }}
		v, err := strconv.ParseUint({{ .VarName }}Raw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "unsigned integer")
		}
		{{- if  .Required }}
		{{ .VarName }} = uint(v)
		{{- else }}
		pv := uint(v)
		{{ .VarName }} = &pv
		{{- end }}
	{{- else if eq .Type.Name "uint32" }}
		v, err := strconv.ParseUint({{ .VarName }}Raw, 10, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "unsigned integer")
		}
		{{- if  .Required }}
		{{ .VarName }} = uint32(v)
		{{- else }}
		pv := uint32(v)
		{{ .VarName }} = &pv
		{{- end }}
	{{- else if eq .Type.Name "uint64" }}
		v, err := strconv.ParseUint({{ .VarName }}Raw, 10, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "unsigned integer")
		}
		{{ .VarName }} = {{ if not .Required }}&{{ end }}v
	{{- else if eq .Type.Name "float32" }}
		v, err := strconv.ParseFloat({{ .VarName }}Raw, 32)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "float")
		}
		{{- if  .Required }}
		{{ .VarName }} = float32(v)
		{{- else }}
		pv := float32(v)
		{{ .VarName }} = &pv
		{{- end }}
	{{- else if eq .Type.Name "float64" }}
		v, err := strconv.ParseFloat({{ .VarName }}Raw, 64)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "float")
		}
		{{ .VarName }} = {{ if not .Required }}&{{ end }}v
	{{- else if eq .Type.Name "boolean" }}
		v, err := strconv.ParseBool({{ .VarName }}Raw)
		if err != nil {
			return nil, goa.InvalidFieldTypeError({{ .VarName}}Raw, {{ .Name }}, "boolean")
		}
		{{ .VarName }} = {{ if not .Required }}&{{ end }}v
	{{- else }}
		// unsupported type {{ .Type.Name }} for var {{ .VarName }}
	{{- end }}
{{- end }}
{{- define "type_slice_conversion" }}
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

const serverEncoderT = `{{ printf "%s returns an encoder for responses returned by the %s %s endpoint." .Encoder .EndpointName .ServiceName | comment }}
func {{ .Encoder }}(encoder func(http.ResponseWriter, *http.Request) rest.Encoder) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {

	{{- if .ResultTypeRef }}
		t := v.({{ .ResultTypeRef }})

		{{- range .Responses }}

			{{- if .TagName }}
		if t.{{ .TagName }} == {{ printf "%q" .TagValue }} {
			{{- end }}
			{{ template "response" . }}

			{{- if .Body }}
			return enc.Encode(&body)
			{{- else }}
			return nil
			{{- end }}

			{{- if .TagName }}
		}
			{{- end }}

		{{- end }}

	{{- else }}

		{{- with (index .Responses 0) }}

		{{- if .Body }}
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		w.WriteHeader({{ .StatusCode }})
		return enc.Encode(v)

		{{- else }}
		w.WriteHeader({{ .StatusCode }})
		return nil

		{{- end }}

		{{- end }}

	{{- end }}
	}
}
` + responseT

const serverErrorEncoderT = `{{ printf "%s returns an encoder for errors returned by the %s %s endpoint." .ErrorEncoder .EndpointName .ServiceName | comment }}
func {{ .ErrorEncoder }}(encoder func(http.ResponseWriter, *http.Request) rest.Encoder, logger goa.LogAdapter) func(http.ResponseWriter, *http.Request, error) {
	encodeError := rest.EncodeError(encoder, logger)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch t := v.(type) {

		{{- range .HTTPErrors }}
		case {{ .TypeRef }}:

			{{- template "response" .Response }}
			{{- if .Response.Body }}
			if err := enc.Encode(&body); err != nil {
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
		{{- if .BodyUserTypeName }}
			{{- if .BodyFields }}
		body := {{ .BodyUserTypeName }}{
				{{- range .BodyFields}}	
			{{ . }}: v.{{ . }},
				{{- end }}
		}
			{{- else }}
			body := {{ .BodyUserTypeName }}(*v)
			{{- end }}
		{{- else }}
		body := v
		{{- end }}
	{{- end }}
	{{- range .Headers }}
		{{- if not .Required }}
	if t.{{ .FieldName }} != nil {
		{{- end }}

		{{- if eq .Type.Name "string" }}
	w.Header().Set("{{ .Name }}", {{ if not .Required }}*{{ end }}t.{{ .FieldName }})
		{{- else }}
	v := t.{{ .FieldName }}
	{{ template "header_conversion" . }}
	w.Header().Set("{{ .Name }}", {{ .VarName }})
		{{- end }}

		{{- if not .Required }}
	}
		{{- end }}
	{{- end }}
	w.WriteHeader({{ .StatusCode }})
{{- end }}

{{- define "header_conversion" }}
	{{- if eq .Type.Name "boolean" }}
		{{ .VarName }} := strconv.FormatBool({{ if not .Required }}*{{ end }}v)
	{{- else if eq .Type.Name "int" }}
		{{ .VarName }} := strconv.Itoa({{ if not .Required }}*{{ end }}v)
	{{- else if eq .Type.Name "int32" }}
		{{ .VarName }} := strconv.FormatInt(int64({{ if not .Required }}*{{ end }}v), 10)
	{{- else if eq .Type.Name "int64" }}
		{{ .VarName }} := strconv.FormatInt({{ if not .Required }}*{{ end }}v, 10)
	{{- else if eq .Type.Name "uint" }}
		{{ .VarName }} := strconv.FormatUint(uint64({{ if not .Required }}*{{ end }}v), 10)
	{{- else if eq .Type.Name "uint32" }}
		{{ .VarName }} := strconv.FormatUint(uint64({{ if not .Required }}*{{ end }}v), 10)
	{{- else if eq .Type.Name "uint64" }}
		{{ .VarName }} := strconv.FormatUint({{ if not .Required }}*{{ end }}v, 10)
	{{- else if eq .Type.Name "float32" }}
		{{ .VarName }} := strconv.FormatFloat(float64({{ if not .Required }}*{{ end }}v), 'f', -1, 32)
	{{- else if eq .Type.Name "float64" }}
		{{ .VarName }} := strconv.FormatFloat({{ if not .Required }}*{{ end }}v, 'f', -1, 64)
	{{- else if eq .Type.Name "string" }}
		{{ .VarName }} := v
	{{- else if eq .Type.Name "bytes" }}
		{{ .VarName }} := []byte({{ if not .Required }}*{{ end }}v)
	{{- else if eq .Type.Name "any" }}
		{{ .VarName }} := {{ if not .Required }}*{{ end }}v
	{{- else }}
		// unsupported type {{ .Type.Name }} for header field {{ .FieldName }}
	{{- end }}
{{- end -}}
`
