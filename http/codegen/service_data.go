// This file turns HTTP endpoint designs into the data used to write client,
// server, request, response, validation, and streaming code.
package codegen

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"slices"
	"sort"
	"strings"
	"text/template"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

var (
	// pathInitTmpl is the template used to render path constructors code.
	pathInitTmpl = template.Must(
		template.New("path-init").
			Funcs(template.FuncMap{"goify": codegen.Goify}).
			Parse(httpTemplates.Read(pathInitT, querySliceConversionP)),
	)

	// requestInitTmpl is the template used to render request constructors.
	requestInitTmpl = template.Must(
		template.New("request-init").
			Parse(httpTemplates.Read(requestInitT)),
	)
)

type (
	// ServicesData encapsulates the data computed from the design.
	ServicesData struct {
		*service.ServicesData
		Expressions *expr.HTTPExpr
		HTTPData    map[string]*ServiceData
		// jsonrpc is true when files are written under gen/jsonrpc and their
		// headings use the JSON-RPC name.
		jsonrpc bool
		// viewedResultConstructors contains every client result function name
		// chosen for the generated client package.
		viewedResultConstructors map[viewedConstructorKey]*codegen.NameDeclaration
		payloadConstructors      map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration
		streamConstructors       map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration
		errorConstructors        map[*expr.HTTPErrorExpr]*codegen.NameDeclaration
		// plannedWireTypes contains each copied request and response field with
		// the Go name used by both its definition and its references.
		plannedWireTypes map[*expr.HTTPServiceExpr]*plannedWireTypes
		// plannedSymbols contains the Go names used in each client and server package.
		plannedSymbols map[*expr.HTTPServiceExpr]*httpSymbols
		// cliParsers contains the function names for each command parser file.
		cliParsers map[*expr.ServerExpr]*cli.ParserPlan
	}

	// ServiceData contains the data used to render the code related to a
	// single service.
	ServiceData struct {
		// Service contains the related service data.
		Service *service.Data
		// ClientPkgName is the Go package name written before client types.
		ClientPkgName string
		// ServerPkgName is the Go package name written before server types.
		ServerPkgName string
		// Endpoints describes the endpoint data for this service.
		Endpoints []*EndpointData
		// FileServers lists the file servers for this service.
		FileServers []*FileServerData
		// ServerStructDeclaration is the package name used by server definitions and calls.
		ServerStructDeclaration *codegen.NameDeclaration
		// MountPointStructDeclaration is the package name used by the mount point type.
		MountPointStructDeclaration *codegen.NameDeclaration
		// ServerInitDeclaration is the package name used by the server constructor.
		ServerInitDeclaration *codegen.NameDeclaration
		// MountServerDeclaration is the package name used by the route mount function.
		MountServerDeclaration *codegen.NameDeclaration
		// ServerService is the name of service function.
		ServerService string
		// ClientStructDeclaration is the package name used by the client type.
		ClientStructDeclaration *codegen.NameDeclaration
		// ClientInitDeclaration is the package name used by the client constructor.
		ClientInitDeclaration *codegen.NameDeclaration
		// ServerConnConfigurerDeclaration names the server WebSocket configuration type.
		ServerConnConfigurerDeclaration *codegen.NameDeclaration
		// ServerConnConfigurerInitDeclaration names the server WebSocket configuration constructor.
		ServerConnConfigurerInitDeclaration *codegen.NameDeclaration
		// ClientConnConfigurerDeclaration names the client WebSocket configuration type.
		ClientConnConfigurerDeclaration *codegen.NameDeclaration
		// ClientConnConfigurerInitDeclaration names the client WebSocket configuration constructor.
		ClientConnConfigurerInitDeclaration *codegen.NameDeclaration
		// AppendFSDeclaration names the file system type used for mapped file paths.
		AppendFSDeclaration *codegen.NameDeclaration
		// AppendPrefixDeclaration names the function that adds a mapped file path prefix.
		AppendPrefixDeclaration *codegen.NameDeclaration
		// ServerBodyAttributeTypes is the list of user types used to
		// define the request, response and error response type
		// attributes in the server code.
		ServerBodyAttributeTypes []*TypeData
		// ClientBodyAttributeTypes is the list of user types used to
		// define the request, response and error response type
		// attributes in the client code.
		ClientBodyAttributeTypes []*TypeData
		// ServerTransformHelpers is the list of transform functions
		// required by the various server side constructors.
		ServerTransformHelpers []*codegen.TransformFunctionData
		// ClientTransformHelpers is the list of transform functions
		// required by the various client side constructors.
		ClientTransformHelpers []*codegen.TransformFunctionData
		// Scope initialized with all the server and client types.
		Scope *codegen.NameScope
		// serverWireTypes owns declarations emitted in the actual server
		// package.
		serverWireTypes *wireTypeCatalog
		// clientWireTypes owns declarations emitted in the actual client
		// package.
		clientWireTypes *wireTypeCatalog
		// bodies stores copied request and response fields after applying the HTTP
		// mappings. Building service data must never change the input design.
		bodies shapedBodies
	}

	// EndpointData contains the data used to render the code related to a
	// single service HTTP endpoint.
	EndpointData struct {
		// Method contains the related service method data.
		Method *service.MethodData
		// IsJSONRPC indicates whether this endpoint is a JSON-RPC
		// endpoint. Unlike Method.IsJSONRPC it is endpoint-scoped: a
		// method exposed over both plain HTTP and JSON-RPC yields two
		// endpoints with different values.
		IsJSONRPC bool
		// ServiceName is the name of the service exposing the endpoint.
		ServiceName string
		// ServiceVarName is the goified service name (first letter
		// lowercase).
		ServiceVarName string
		// ServicePkgName is the name of the service package.
		ServicePkgName string
		// Payload describes the method HTTP payload.
		Payload *PayloadData
		// Result describes the method HTTP result.
		Result *ResultData
		// Errors describes the method HTTP errors.
		Errors []*ErrorGroupData
		// Routes describes the possible routes for this endpoint.
		Routes []*RouteData
		// BasicScheme is the basic auth security scheme if any.
		BasicScheme *service.SchemeData
		// HeaderSchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request header.
		HeaderSchemes service.SchemesData
		// BodySchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request body.
		BodySchemes service.SchemesData
		// QuerySchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request query
		// string.
		QuerySchemes service.SchemesData
		// Requirements contains the security requirements for the
		// method.
		Requirements service.RequirementsData

		// server

		// MountHandlerDeclaration is the package name used by this endpoint's mount function.
		MountHandlerDeclaration *codegen.NameDeclaration
		// HandlerInitDeclaration is the package name used by this endpoint's handler constructor.
		HandlerInitDeclaration *codegen.NameDeclaration
		// RequestDecoderDeclaration is the package name used by this endpoint's request decoder.
		RequestDecoderDeclaration *codegen.NameDeclaration
		// ResponseEncoderDeclaration is the package name used by this endpoint's response encoder.
		ResponseEncoderDeclaration *codegen.NameDeclaration
		// ErrorEncoderDeclaration is the package name used by this endpoint's error encoder.
		ErrorEncoderDeclaration *codegen.NameDeclaration
		// DiscardStreamDeclaration names the no-output stream used by a mixed-result request.
		DiscardStreamDeclaration *codegen.NameDeclaration
		// MultipartRequestDecoder indicates the request decoder for
		// multipart content type.
		MultipartRequestDecoder *MultipartData
		// ServerWebSocket holds the data to render the server struct which
		// implements the server stream interface.
		ServerWebSocket *WebSocketData
		// SSE holds the data to render the server struct which implements the
		// server stream interface for SSE.
		SSE *SSEData
		// Redirect defines a redirect for the endpoint.
		Redirect *RedirectData
		// HasMixedResults indicates if the method has both Result and StreamingResult
		// defined with different types, enabling content negotiation.
		HasMixedResults bool

		// client

		// ClientStructDeclaration supplies the client type name used by endpoint methods.
		ClientStructDeclaration *codegen.NameDeclaration
		// EndpointInit is the name of the constructor function for the
		// client endpoint.
		EndpointInit string
		// RequestInit is the request builder function.
		RequestInit *InitData
		// RequestEncoderDeclaration is the package name used by this endpoint's request encoder.
		RequestEncoderDeclaration *codegen.NameDeclaration
		// ResponseDecoderDeclaration is the package name used by this endpoint's response decoder.
		ResponseDecoderDeclaration *codegen.NameDeclaration
		// MultipartRequestEncoder indicates the request encoder for
		// multipart content type.
		MultipartRequestEncoder *MultipartData
		// ClientWebSocket holds the data to render the client struct which
		// implements the client stream interface.
		ClientWebSocket *WebSocketData
		// BuildStreamPayloadDeclaration is the package name used by the streamed request helper.
		BuildStreamPayloadDeclaration *codegen.NameDeclaration
		// CLIPayloadDeclaration is the package name used by the command-line payload helper.
		CLIPayloadDeclaration *codegen.NameDeclaration
	}

	// FileServerData lists the data needed to generate file servers.
	FileServerData struct {
		// MountHandlerDeclaration is the package name used by this file server's mount function.
		MountHandlerDeclaration *codegen.NameDeclaration
		// RequestPaths is the set of HTTP paths to the server.
		RequestPaths []string
		// Root is the root server file path.
		FilePath string
		// Dir is true if the file server servers files under a
		// directory, false if it serves a single file.
		IsDir bool
		// PathParam is the name of the parameter used to capture the
		// path for file servers that serve files under a directory.
		PathParam string
		// Redirect defines a redirect for the endpoint.
		Redirect *RedirectData
		// VarName is the name of the variable that holds the file server.
		VarName string
		// ArgName is the name of the argument used to initialize the
		// file server.
		ArgName string
	}

	// RedirectData lists the data needed to generate a redirect.
	RedirectData struct {
		// URL is the URL that is being redirected to.
		URL string
		// StatusCode is the HTTP status code.
		StatusCode string
	}

	// PayloadData contains the payload information required to generate the
	// transport decode (server) and encode (client) code.
	PayloadData struct {
		// Name is the name of the payload type.
		Name string
		// Ref is the fully qualified reference to the payload type.
		Ref string
		// Request contains the data for the corresponding HTTP request.
		Request *RequestData
		// DecoderReturnValue is a reference to the decoder return value
		// if there is no payload constructor (i.e. if Init is nil).
		DecoderReturnValue string
		// IDAttribute is the name of the attribute where the ID of the
		// JSON-RPC request is stored.
		IDAttribute string
		// IDAttributeRequired is true if the ID attribute is required.
		IDAttributeRequired bool
	}

	// ResultData contains the result information required to generate the
	// transport decode (client) and encode (server) code.
	ResultData struct {
		// Name is the name of the result type.
		Name string
		// Ref is the reference to the result type.
		Ref string
		// IsStruct is true if the result type is a user type defining
		// an object.
		IsStruct bool
		// Inits contains the data required to render the result
		// constructors if any.
		Inits []*InitData
		// Responses contains the data for the corresponding HTTP
		// responses.
		Responses []*ResponseData
		// IDAttribute is the name of the attribute where the ID of the
		// JSON-RPC request is stored.
		IDAttribute string
		// IDAttributeRequired is true if the ID attribute is required.
		IDAttributeRequired bool
		// View is the view used to render the result.
		View string
		// MustInit indicates if a variable holding the result type must be
		// initialized. It is used by server response encoder to initialize
		// the result variable only if there are multiple responses, or the
		// response has a body, a header or a cookie.
		MustInit bool
	}

	// ErrorGroupData contains the error information required to generate
	// the transport decode (client) and encode (server) code for all errors
	// with responses using a given HTTP status code.
	ErrorGroupData struct {
		// StatusCode is the response HTTP status code.
		StatusCode string
		// Errors contains the information for each error.
		Errors []*ErrorData
	}

	// ErrorData contains the error information required to generate the
	// transport decode (client) and encode (server) code.
	ErrorData struct {
		// Name is the error name.
		Name string
		// Ref is a reference to the error type.
		Ref string
		// Response is the error response data.
		Response *ResponseData
	}

	// RequestData describes a request.
	RequestData struct {
		// PathParams describes the information about params that are
		// present in the request path.
		PathParams []*ParamData
		// QueryParams describes the information about the params that
		// are present in the request query string.
		QueryParams []*ParamData
		// Headers contains the HTTP request headers used to build the
		// method payload.
		Headers []*HeaderData
		// Cookies contains the HTTP request cookies used to build the
		// method payload.
		Cookies []*CookieData
		// ServerBody describes the request body type used by server
		// code. The type is generated using pointers for all fields so
		// that it can be validated.
		ServerBody *TypeData
		// ClientBody describes the request body type used by client
		// code. The type does NOT use pointers for every fields since
		// no validation is required.
		ClientBody *TypeData
		// PayloadInit contains the data required to render the
		// payload constructor used by server code if any.
		PayloadInit *InitData
		// PayloadType is the type of the payload.
		PayloadType expr.DataType
		// PayloadAttr sets the request body from the specified payload type
		// attribute. This field is set when the design uses Body("name") syntax
		// to set the request body and the payload type is an object.
		PayloadAttr string
		// MustHaveBody is true if the request body cannot be empty.
		MustHaveBody bool
		// MustValidate is true if the request body or at least one
		// parameter or header requires validation.
		MustValidate bool
		// Multipart if true indicates the request is a multipart
		// request.
		Multipart bool
	}

	// ResponseData describes a response.
	ResponseData struct {
		// StatusCode is the return code of the response.
		StatusCode string
		// Code is the return code of the response.
		Code int
		// Description is the response description.
		Description string
		// Headers provides information about the HTTP response headers.
		Headers []*HeaderData
		// Cookies provides information about the HTTP response cookies.
		Cookies []*CookieData
		// ContentType contains the value of the response
		// "Content-Type" header.
		ContentType string
		// ErrorHeader contains the value of the response "goa-error"
		// header if any.
		ErrorHeader string
		// ServerBody is the type of the response body used by server
		// code, nil if body should be empty. The type does NOT use
		// pointers for all fields. If the method result is a result
		// type and the response data describes a success response, then
		// this field contains a type for every view in the result type.
		// The type name is suffixed with the name of the view (except
		// for "default" view where no suffix is added). A constructor
		// is also generated server side for each view to transform the
		// result type to the corresponding response body type. If
		// method result is not a result type or if the response
		// describes an error response, then this field contains at most
		// one item.
		ServerBody []*TypeData
		// ClientBody is the type of the response body used by client
		// code, nil if body should be empty. The type uses pointers for
		// all fields so they can be validated.
		ClientBody *TypeData
		// Init contains the data required to render the result or error
		// constructor if any.
		ResultInit *InitData
		// TagName is the name of the attribute used to test whether the
		// response is the one to use.
		TagName string
		// TagValue is the value the result attribute named by TagName
		// must have for this response to be used.
		TagValue string
		// TagPointer is true if the tag attribute is a pointer.
		TagPointer bool
		// MustValidate is true if at least one header requires validation.
		MustValidate bool
		// ResultAttr sets the response body from the specified result
		// type attribute. This field is set when the design uses
		// Body("name") syntax to set the response body and the result
		// type is an object.
		ResultAttr string
		// ViewedResult indicates whether the response body type is a
		// result type.
		ViewedResult *service.ViewedResultTypeData
		// ViewedRepresentations lists the body type and constructor used for each
		// legal result view. A response that supports several views includes its
		// view name so the client can choose the matching entry.
		ViewedRepresentations []*ViewedRepresentationData
	}

	// ViewedRepresentationData describes the HTTP body used for one legal result
	// view. The server constructor converts the service result into ServerBody.
	// The client decodes ClientBody and ResultInit rebuilds the service result.
	ViewedRepresentationData struct {
		// View is the exact design view name carried on variable-view messages.
		View string
		// ResultAttr is the Go field selected by Body("name"). It is empty when
		// the response body uses the complete projected result.
		ResultAttr string
		// ServerBody is the body type encoded by the server for View.
		ServerBody *TypeData
		// ClientBody is the body type decoded by the client for View.
		ClientBody *TypeData
		// ResultInit rebuilds the projected result from ClientBody.
		ResultInit *InitData
	}

	// InitData contains the data required to render a constructor.
	InitData struct {
		// Declaration is the generated package function name used by this constructor.
		Declaration *codegen.NameDeclaration
		// ClientDeclaration is the client package name for a path function also emitted on the server.
		ClientDeclaration *codegen.NameDeclaration
		// Name is the constructor function name.
		Name string
		// Description is the function description.
		Description string
		// ServerArgs is the list of constructor arguments for server
		// side code.
		ServerArgs []*InitArgData
		// ClientArgs is the list of constructor arguments for client
		// side code.
		ClientArgs []*InitArgData
		// CLIArgs is the list of arguments that should be initialized
		// from CLI flags. This is used for implicit attributes which
		// as the time of writing is only used for the basic auth
		// username and password.
		CLIArgs []*InitArgData
		// ServerCode is the code that builds the payload from the
		// request on the server when it contains user types.
		ServerCode string
		// ClientCode is the code that builds the payload or result type
		// from the request or response state on the client when it
		// contains user types.
		ClientCode string
		// ReturnTypePkg is the package where the return type is present.
		ReturnTypePkg string
		// ReturnTypeName is the qualified (including the package name)
		// name of the payload, result or error type.
		ReturnTypeName string
		// ReturnTypeRef is the qualified (including the package name)
		// reference to the payload, result or error type.
		ReturnTypeRef string
		// ReturnTypeAttribute is the name of the attribute initialized by this
		// constructor when it only initializes one attribute (i.e. body was
		// defined with Body("name") syntax).
		ReturnTypeAttribute string
		// ReturnIsStruct is true if the payload, result or error type is a struct.
		ReturnIsStruct bool
		// ReturnIsPrimitivePointer indicates whether the payload, result or error
		// type is a primitive pointer.
		ReturnIsPrimitivePointer bool
	}

	// AttributeData contains the information needed to generate the code
	// related to a specific payload or result attribute.
	AttributeData struct {
		// Name is the name of the attribute.
		Name string
		// VarName is the name of the variable that holds the attribute value.
		VarName string
		// Pointer is true if the attribute value is a pointer.
		Pointer bool
		// Required is true if the attribute is required in the parent attribute.
		Required bool
		// Type is the attribute type.
		Type expr.DataType
		// TypeName is the generated attribute type name.
		TypeName string
		// TypeRef is the generated attribute type reference.
		TypeRef string
		// ElemTypeRef is the generated element type reference for an array.
		ElemTypeRef string
		// Description is the attribute description as defined in the design.
		Description string
		// FieldName is the name of the data structure field that should
		// be initialized with the value if any.
		FieldName string
		// FieldType is the type of the data structure field that should be
		// initialized with the attribute value or read into the attribute value.
		FieldType expr.DataType
		// FieldPointer if true indicates that the data structure field is a
		// pointer.
		FieldPointer bool
		// DefaultValue is the default value of the attribute if any.
		DefaultValue any
		// Validate contains the validation code for the attribute value if any.
		Validate string
		// Example is an example attribute value
		Example any
		// IsAliased is true when the field uses a user-defined type.
		IsAliased bool
		// ServiceTypeRef is the Go type used when the field comes from another service.
		ServiceTypeRef string
		// IsTextUnmarshaler is true if the attribute has a struct:field:type meta
		// whose underlying DSL type is string and the custom type is expected to
		// implement encoding.TextUnmarshaler for conversion from HTTP path/query params.
		IsTextUnmarshaler bool
	}

	// InitArgData represents a single constructor argument.
	InitArgData struct {
		*AttributeData
		// Reference to the argument, e.g. "&body".
		Ref string
	}

	// RouteData describes a route.
	RouteData struct {
		// Verb is the HTTP method.
		Verb string
		// Path is the fullpath including wildcards.
		Path string
		// PathInit contains the information needed to render and call
		// the path constructor for the route.
		PathInit *InitData
	}

	// Element defines the common fields needed to generate HTTP request and
	// response elements including headers, parameters and cookies.
	Element struct {
		*AttributeData
		// HTTPName is the name of the HTTP element (header name, query string name
		// or cookie name)
		HTTPName string
		// StringSlice is true if the attribute type is array of strings.
		StringSlice bool
		// Slice is true if the attribute type is an array.
		Slice bool
	}

	// ParamData describes a HTTP request parameter (query string or path
	// parameter).
	ParamData struct {
		*Element
		// MapStringSlice is true if the param type is a map of string
		// slice.
		MapStringSlice bool
		// Map is true if the param type is a map.
		Map bool
		// MapQueryParams indicates that the query params must be mapped
		// to the entire payload (empty string) or a payload attribute
		// (attribute name).
		MapQueryParams *string
	}

	// HeaderData describes a HTTP request or response header.
	HeaderData struct {
		*Element
		// CanonicalName is the standard HTTP header spelling.
		CanonicalName string
	}

	// CookieData describes a HTTP request or response cookie.
	CookieData struct {
		*Element
		// MaxAge is the cookie "max-age" attribute.
		MaxAge string
		// Path is the cookie "path" attribute.
		Path string
		// Domain is the cookie "domain" attribute.
		Domain string
		// Secure sets the cookie "secure" attribute to "Secure" if true.
		Secure bool
		// HTTPOnly sets the cookie "http-only" attribute to "HttpOnly" if true.
		HTTPOnly bool
		// SameSite sets the cookie "same-site" attribute to the given value.
		SameSite string
	}

	// TypeData contains the data needed to render a type definition.
	TypeData struct {
		// Name is the type name.
		Name string
		// VarName is the Go type name.
		VarName string
		// Description is the type human description.
		Description string
		// Init contains the data needed to render and call the type
		// constructor if any.
		Init *InitData
		// Def is the type definition Go code.
		Def string
		// Ref is the reference to the type.
		Ref string
		// ValidateDef contains the validation code.
		ValidateDef string
		// ValidateRef contains the call to the validation code.
		ValidateRef string
		// ValidatorName is the package-level function that runs ValidateDef.
		ValidatorName string
		// Example is an example value for the type.
		Example any
		// View is the view used to render the (result) type if any.
		View string
		// Declaration identifies the canonical declaration and validator owned
		// by the generated output package.
		declaration *wireTypeRecord
	}

	// MultipartData contains the data needed to render multipart
	// encoder/decoder.
	MultipartData struct {
		// FuncDeclaration is the package name used by the multipart function type or root helper.
		FuncDeclaration *codegen.NameDeclaration
		// InitDeclaration is the package name used by the multipart constructor.
		InitDeclaration *codegen.NameDeclaration
		// VarName is the name of the variable referring to the function.
		VarName string
		// ServiceName is the name of the service.
		ServiceName string
		// MethodName is the name of the method.
		MethodName string
		// Payload is the payload data required to generate
		// encoder/decoder.
		Payload *PayloadData
	}

	// httpElementKind identifies the kind of HTTP request or response
	// element extracted from a mapped attribute expression: path parameter,
	// query string parameter, header or cookie. Its value is used in bug
	// report messages.
	httpElementKind string

	// shapedBodies caches the detached body attributes computed from the
	// design expressions: request and response bodies are shaped with
	// makeHTTPType while streaming bodies are plain copies. Caching
	// guarantees the shaping runs once per expression and that all the
	// consumers share the same attribute instances, which keeps the
	// example generator call sequence stable.
	shapedBodies struct {
		requests  map[*expr.HTTPEndpointExpr]*expr.AttributeExpr
		streams   map[*expr.HTTPEndpointExpr]*expr.AttributeExpr
		responses map[*expr.HTTPResponseExpr]*expr.AttributeExpr
		errors    map[*expr.HTTPErrorExpr]*expr.AttributeExpr
	}
)

const (
	// pathElement identifies path parameters.
	pathElement httpElementKind = "path"
	// queryElement identifies query string parameters.
	queryElement httpElementKind = "query"
	// headerElement identifies headers.
	headerElement httpElementKind = "header"
	// cookieElement identifies cookies.
	cookieElement httpElementKind = "cookie"
)

// newServicesData creates the HTTP service map that Plan.Link fills before it
// builds any generated file.
func newServicesData(services *service.ServicesData, expressions *expr.HTTPExpr) *ServicesData {
	return &ServicesData{
		ServicesData: services,
		Expressions:  expressions,
		HTTPData:     make(map[string]*ServiceData),
	}
}

// Get returns the generated HTTP information for the service with the given
// name. A missing entry means the design does not expose that service over the
// protocol handled by this plan.
func (sds *ServicesData) Get(name string) *ServiceData {
	return sds.HTTPData[name]
}

// Endpoint returns the service method transport data for the endpoint with the
// given name, nil if there isn't one.
func (svc *ServiceData) Endpoint(name string) *EndpointData {
	for _, e := range svc.Endpoints {
		if e.Method.Name == name {
			return e
		}
	}
	return nil
}

// dir returns the name of the transport directory under gen: "http" or
// "jsonrpc".
func (sds *ServicesData) dir() string {
	if sds.jsonrpc {
		return "jsonrpc"
	}
	return "http"
}

// label returns the transport label used in generated file headers: "HTTP" or
// "JSON-RPC".
func (sds *ServicesData) label() string {
	if sds.jsonrpc {
		return "JSON-RPC"
	}
	return "HTTP"
}

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (sds *ServicesData) analyze(httpSvc *expr.HTTPServiceExpr) *ServiceData {
	svc := sds.ServicesData.Get(httpSvc.ServiceExpr.Name)
	transportService := *svc
	transportService.PkgName = sds.ServiceImport(svc.Name).Name
	svc = &transportService
	scope := codegen.NewNameScope()
	scope.Unique("c") // 'c' is reserved as the client's receiver name.
	scope.Unique("v") // 'v' is reserved as the request builder payload argument name.
	// Reserve 'websocket' to avoid collision with gorilla/websocket
	scope.Unique("websocket")
	// Reserve the service package alias to avoid collision with parameter names in generated code.
	scope.Unique(svc.PkgName)
	planned := sds.plannedWireTypes[httpSvc]
	if planned == nil {
		panic(fmt.Sprintf("HTTP service %q has no planned generated types", httpSvc.Name()))
	}
	planned.server.Link()
	planned.client.Link()
	symbols := sds.plannedSymbols[httpSvc]
	if symbols == nil {
		panic(fmt.Sprintf("HTTP service %q has no package names", httpSvc.Name()))
	}
	sd := &ServiceData{
		Service:                             svc,
		ClientPkgName:                       sds.PackageImport(path.Join(sds.GenPkg(), sds.dir(), svc.PathName, "client")).Name,
		ServerPkgName:                       sds.PackageImport(path.Join(sds.GenPkg(), sds.dir(), svc.PathName, "server")).Name,
		ServerStructDeclaration:             symbols.serverStruct,
		MountPointStructDeclaration:         symbols.mountPoint,
		ServerInitDeclaration:               symbols.serverInit,
		MountServerDeclaration:              symbols.mountServer,
		ServerService:                       "Service",
		ClientStructDeclaration:             symbols.clientStruct,
		ClientInitDeclaration:               symbols.clientInit,
		ServerConnConfigurerDeclaration:     symbols.serverConfigurer,
		ServerConnConfigurerInitDeclaration: symbols.serverConfigurerInit,
		ClientConnConfigurerDeclaration:     symbols.clientConfigurer,
		ClientConnConfigurerInitDeclaration: symbols.clientConfigurerInit,
		AppendFSDeclaration:                 symbols.appendFS,
		AppendPrefixDeclaration:             symbols.appendPrefix,
		Scope:                               scope,
		serverWireTypes:                     planned.server,
		clientWireTypes:                     planned.client,
		bodies:                              planned.bodies,
	}

	for _, s := range httpSvc.FileServers {
		paths := make([]string, len(s.RequestPaths))
		for i, p := range s.RequestPaths {
			idx := strings.LastIndex(p, "/{")
			switch {
			case idx == 0:
				paths[i] = "/"
			case idx > 0:
				paths[i] = p[:idx]
			default:
				paths[i] = p
			}
		}
		var pp string
		if s.IsDir() {
			pp = expr.ExtractHTTPWildcards(s.RequestPaths[0])[0]
		}
		var redirect *RedirectData
		if s.Redirect != nil {
			redirect = &RedirectData{
				URL:        s.Redirect.URL,
				StatusCode: statusCodeToHTTPConst(s.Redirect.StatusCode),
			}
		}
		data := &FileServerData{
			MountHandlerDeclaration: symbols.fileServers[s],
			RequestPaths:            paths,
			FilePath:                s.FilePath,
			IsDir:                   s.IsDir(),
			PathParam:               pp,
			Redirect:                redirect,
			VarName:                 scope.Unique(codegen.Goify(s.FilePath, true)),
			ArgName:                 scope.Unique(fmt.Sprintf("fileSystem%s", codegen.Goify(s.FilePath, true))),
		}
		sd.FileServers = append(sd.FileServers, data)
	}

	for _, httpEndpoint := range httpSvc.HTTPEndpoints {
		method := svc.Method(httpEndpoint.MethodExpr.Name)

		routesCap := 0
		for _, r := range httpEndpoint.Routes {
			routesCap += len(r.FullPaths())
		}
		routes := make([]*RouteData, 0, routesCap)
		endpointSymbols := symbols.endpoints[httpEndpoint]
		if endpointSymbols == nil {
			panic(fmt.Sprintf("HTTP endpoint %q has no package names", httpEndpoint.Name()))
		}
		pathCount := 0
		for _, r := range httpEndpoint.Routes {
			for _, rpath := range r.FullPaths() {
				params := expr.ExtractHTTPWildcards(rpath)
				var (
					init *InitData
				)
				{
					initArgs := make([]*InitArgData, len(params))
					pathParamsObj := expr.AsObject(httpEndpoint.PathParams().Type)
					declaration := endpointSymbols.serverPaths[pathCount]
					name := declaration.Name()
					for j, arg := range params {
						patt := pathParamsObj.Attribute(arg)
						att := makeHTTPType(patt)
						pointer := httpEndpoint.Params.IsPrimitivePointer(arg, true)
						if expr.IsObject(httpEndpoint.MethodExpr.Payload.Type) {
							// Path params may override requiredness, need to check payload.
							pointer = httpEndpoint.MethodExpr.Payload.IsPrimitivePointer(arg, true)
						}
						name := sd.Scope.Name(codegen.Goify(arg, false))
						var vcode string
						if att.Validation != nil {
							ctx := httpContext(sd.Scope, true, false)
							vcode = codegen.AttributeValidationCode(att, nil, ctx, true, expr.IsAlias(att.Type), name, arg)
						}
						initArgs[j] = &InitArgData{
							Ref: name,
							AttributeData: &AttributeData{
								Name:        arg,
								VarName:     name,
								Description: att.Description,
								FieldName:   codegen.Goify(arg, true),
								FieldType:   patt.Type,
								TypeName:    sd.Scope.GoTypeName(att),
								TypeRef:     sd.Scope.GoTypeRef(att),
								Type:        att.Type,
								Pointer:     pointer,
								Required:    true,
								Example:     sds.FieldExample(att, httpEndpoint.MethodExpr.Payload, arg, expr.MethodPayloadExampleIdentity(httpEndpoint.MethodExpr)),
								Validate:    vcode,
							},
						}
					}

					var buffer bytes.Buffer
					pf := expr.HTTPWildcardRegex.ReplaceAllString(rpath, "/%v")
					err := pathInitTmpl.Execute(&buffer, map[string]any{
						"Args":       initArgs,
						"PathParams": pathParamsObj,
						"PathFormat": pf,
					})
					if err != nil {
						panic(err)
					}
					// The request builder construction below renames the
					// client-side arguments in place (VarName, Ref,
					// IsAliased and ServiceTypeRef) so the client args
					// must not alias the server args.
					clientArgs := make([]*InitArgData, len(initArgs))
					for j, arg := range initArgs {
						attCopy := *arg.AttributeData
						clientArgs[j] = &InitArgData{
							AttributeData: &attCopy,
							Ref:           arg.Ref,
						}
					}
					init = &InitData{
						Declaration:       declaration,
						ClientDeclaration: endpointSymbols.clientPaths[pathCount],
						Name:              name,
						Description:       fmt.Sprintf("%s returns the URL path to the %s service %s HTTP endpoint. ", name, svc.Name, method.Name),
						ServerArgs:        initArgs,
						ClientArgs:        clientArgs,
						ReturnTypeName:    "string",
						ReturnTypeRef:     "string",
						ServerCode:        buffer.String(),
						ClientCode:        buffer.String(),
					}
				}

				routes = append(routes, &RouteData{
					Verb:     strings.ToUpper(r.Method),
					Path:     rpath,
					PathInit: init,
				})
				pathCount++
			}
		}

		payload := sds.buildPayloadData(httpEndpoint, sd)

		var (
			reqs  = make(service.RequirementsData, 0, len(httpEndpoint.Requirements))
			hsch  service.SchemesData
			bosch service.SchemesData
			qsch  service.SchemesData
			basch *service.SchemeData
		)
		for _, req := range httpEndpoint.Requirements {
			rs := make(service.SchemesData, 0, len(req.Schemes))
			for _, sch := range req.Schemes {
				s := service.BuildSchemeData(sch, httpEndpoint.MethodExpr)
				rs = rs.Append(s)
				switch s.Type {
				case "Basic":
					basch = s
				default:
					switch s.In {
					case "query":
						qsch = qsch.Append(s)
					case "header":
						hsch = hsch.Append(s)
					default:
						bosch = bosch.Append(s)
					}
				}
			}
			reqs = append(reqs, &service.RequirementData{Schemes: rs, Scopes: req.Scopes})
		}

		var requestInit *InitData
		var (
			name       string
			args       []*InitArgData
			payloadRef string
			pkg        string
		)
		{
			name = fmt.Sprintf("Build%sRequest", method.VarName)
			svcctx := sds.serviceTypeContext(sd, "client").Enter(httpEndpoint.MethodExpr.Payload)
			s := codegen.NewNameScope()
			s.Unique("c") // 'c' is reserved as the client's receiver name.
			for _, ca := range routes[0].PathInit.ClientArgs {
				if ca.FieldName != "" {
					ca.VarName = s.Unique(ca.VarName)
					ca.Ref = ca.VarName
					// Populate service-aware type resolution fields
					_, ca.IsAliased = ca.FieldType.(expr.UserType)
					if ca.IsAliased {
						attribute := &expr.AttributeExpr{Type: ca.Type}
						ca.ServiceTypeRef = svcctx.Scope.Ref(attribute, svcctx.Pkg(attribute))
					}
					args = append(args, ca)
				}
			}
			pkg = svc.PkgName
			if len(routes[0].PathInit.ClientArgs) > 0 && httpEndpoint.MethodExpr.Payload.Type != expr.Empty {
				payloadRef = svcctx.Scope.Ref(httpEndpoint.MethodExpr.Payload, svcctx.Pkg(httpEndpoint.MethodExpr.Payload))
			}
		}
		data := map[string]any{
			"PayloadRef":   payloadRef,
			"HasFields":    expr.IsObject(httpEndpoint.MethodExpr.Payload.Type),
			"ServiceName":  svc.Name,
			"EndpointName": method.Name,
			"Args":         args,
			"PathInit":     routes[0].PathInit,
			"Verb":         routes[0].Verb,
			"IsWebSocket":  httpEndpoint.UsesWebSocket(),
		}
		if httpEndpoint.SkipRequestBodyEncodeDecode {
			data["RequestStruct"] = pkg + "." + method.RequestStruct
		}
		var buf bytes.Buffer
		if err := requestInitTmpl.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		clientArgs := []*InitArgData{{Ref: "v", AttributeData: &AttributeData{Name: "payload", VarName: "v", TypeRef: "any"}}}
		requestInit = &InitData{
			Name:        name,
			Description: fmt.Sprintf("%s instantiates a HTTP request object with method and path set to call the %q service %q endpoint", name, svc.Name, method.Name),
			ClientCode:  buf.String(),
			ClientArgs:  clientArgs,
		}

		ed := &EndpointData{
			Method:                     method,
			IsJSONRPC:                  httpEndpoint.IsJSONRPC(),
			ServiceName:                svc.Name,
			ServiceVarName:             svc.VarName,
			ServicePkgName:             svc.PkgName,
			Payload:                    payload,
			Result:                     sds.buildResultData(httpEndpoint, sd),
			Errors:                     sds.buildErrorsData(httpEndpoint, sd),
			HeaderSchemes:              hsch,
			BodySchemes:                bosch,
			QuerySchemes:               qsch,
			BasicScheme:                basch,
			Routes:                     routes,
			MountHandlerDeclaration:    endpointSymbols.mountHandler,
			HandlerInitDeclaration:     endpointSymbols.handlerInit,
			RequestDecoderDeclaration:  endpointSymbols.requestDecoder,
			ResponseEncoderDeclaration: endpointSymbols.responseEncoder,
			ErrorEncoderDeclaration:    endpointSymbols.errorEncoder,
			DiscardStreamDeclaration:   endpointSymbols.discardStream,
			ClientStructDeclaration:    symbols.clientStruct,
			EndpointInit:               method.VarName,
			RequestInit:                requestInit,
			HasMixedResults:            httpEndpoint.MethodExpr.HasMixedResults(),
			RequestEncoderDeclaration:  endpointSymbols.requestEncoder,
			ResponseDecoderDeclaration: endpointSymbols.responseDecoder,
			Requirements:               reqs,
		}
		if httpEndpoint.MethodExpr.IsStreaming() {
			sds.initWebSocketData(ed, httpEndpoint, sd)
			sds.initSSEData(ed, httpEndpoint, sd)
			if ed.ServerWebSocket != nil {
				ed.ServerWebSocket.VarDeclaration = endpointSymbols.serverStream
				ed.ServerWebSocket.VarName = endpointSymbols.serverStream.Name()
			}
			if ed.ClientWebSocket != nil {
				ed.ClientWebSocket.VarDeclaration = endpointSymbols.clientStream
				ed.ClientWebSocket.VarName = endpointSymbols.clientStream.Name()
			}
			if ed.SSE != nil {
				ed.SSE.StructDeclaration = endpointSymbols.serverStream
				ed.SSE.ClientInterfaceDeclaration = endpointSymbols.sseClientInterface
				ed.SSE.ClientStructDeclaration = endpointSymbols.sseClientStruct
				ed.SSE.ClientInitDeclaration = endpointSymbols.sseClientInit
			}
		}

		if httpEndpoint.MultipartRequest {
			ed.MultipartRequestDecoder = &MultipartData{
				FuncDeclaration: endpointSymbols.serverMultipart.functionType,
				InitDeclaration: endpointSymbols.serverMultipart.constructor,
				VarName:         fmt.Sprintf("%s%sDecoderFn", svc.VarName, method.VarName),
				ServiceName:     svc.Name,
				MethodName:      method.Name,
				Payload:         ed.Payload,
			}
			ed.MultipartRequestEncoder = &MultipartData{
				FuncDeclaration: endpointSymbols.clientMultipart.functionType,
				InitDeclaration: endpointSymbols.clientMultipart.constructor,
				VarName:         fmt.Sprintf("%s%sEncoderFn", svc.VarName, method.VarName),
				ServiceName:     svc.Name,
				MethodName:      method.Name,
				Payload:         ed.Payload,
			}
		}

		if httpEndpoint.SkipRequestBodyEncodeDecode {
			ed.BuildStreamPayloadDeclaration = endpointSymbols.buildStreamPayload
		}
		ed.CLIPayloadDeclaration = endpointSymbols.cliPayload

		if httpEndpoint.Redirect != nil {
			ed.Redirect = &RedirectData{
				URL:        httpEndpoint.Redirect.URL,
				StatusCode: statusCodeToHTTPConst(httpEndpoint.Redirect.StatusCode),
			}
		}

		sd.Endpoints = append(sd.Endpoints, ed)
	}

	for _, a := range httpSvc.HTTPEndpoints {
		sds.buildRequestAttributeTypes(sd.bodies.request(a), sd)

		if a.MethodExpr.StreamingPayload.Type != expr.Empty {
			sds.buildRequestAttributeTypes(sd.bodies.streaming(a), sd)
		}

	}

	return sd
}

// buildRequestAttributeTypes builds nested request declarations from separate
// tagged copies because server and client packages apply different pointer and
// default policies to the same authored body graph.
func (sds *ServicesData) buildRequestAttributeTypes(body *expr.AttributeExpr, data *ServiceData) {
	for _, side := range []struct {
		server  bool
		pointer bool
	}{
		{server: true, pointer: true},
		{server: false, pointer: false},
	} {
		body := expr.DupAtt(body)
		addMarshalTags(body)
		top, _ := body.Type.(expr.UserType)
		collectUserTypes(body.Type, func(userType expr.UserType) {
			if top != nil && userType.Origin() == top.Origin() {
				return
			}
			declaration := sds.attributeTypeData(userType, true, side.pointer, side.server, data)
			if declaration == nil {
				return
			}
			if side.server {
				data.ServerBodyAttributeTypes = append(data.ServerBodyAttributeTypes, declaration)
			} else {
				data.ClientBodyAttributeTypes = append(data.ClientBodyAttributeTypes, declaration)
			}
		})
	}
}

// collectPlannedWireTypes records every request and response type written by
// the generated client and server packages. NewPlans calls it before Goa
// assigns package names, and Link later uses these same copied values.
func collectPlannedWireTypes(httpService *expr.HTTPServiceExpr, planned *plannedWireTypes, servicePlan *service.Plan) {
	bodies, server, client := &planned.bodies, planned.server, planned.client
	for _, endpoint := range httpService.HTTPEndpoints {
		request := expr.DupAtt(bodies.request(endpoint))
		addMarshalTags(request)
		server.collect(request, wireRequestBody, wireTypePolicy{request: true, pointer: true, validate: true}, "")
		clientRequest := client.collect(request, wireRequestBody, wireTypePolicy{request: true, useDefault: true, validate: true}, "")
		if clientRequest != nil && needInit(request.Type) {
			clientRequest.needsConstructor = true
		}
		server.collectChildren(request, wireAttribute, wireTypePolicy{request: true, pointer: true, validate: true})
		client.collectChildren(request, wireAttribute, wireTypePolicy{request: true, useDefault: true, validate: true})
		if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty {
			streaming := expr.DupAtt(bodies.streaming(endpoint))
			addMarshalTags(streaming)
			serverStream := server.collect(streaming, wireStreamPayload, wireTypePolicy{request: true, pointer: true, validate: true}, "")
			if endpoint.UsesWebSocket() && needInit(endpoint.MethodExpr.StreamingPayload.Type) && serverStream != nil {
				serverStream.needsConstructor = true
				planned.streamPayloads[endpoint] = serverStream
			}
			clientStream := client.collect(streaming, wireStreamPayload, wireTypePolicy{request: true, useDefault: true, validate: true}, "")
			if clientStream != nil && needInit(streaming.Type) {
				clientStream.needsConstructor = true
			}
			server.collectChildren(streaming, wireAttribute, wireTypePolicy{request: true, pointer: true, validate: true})
			client.collectChildren(streaming, wireAttribute, wireTypePolicy{request: true, useDefault: true, validate: true})
		}

		resultType, viewed := endpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr)
		for _, response := range endpoint.Responses {
			body := bodies.response(response)
			if !viewed {
				collectResponseWireType(body, endpoint, server, true, nil)
				collectResponseWireType(body, endpoint, client, false, nil)
				continue
			}
			origin := ""
			if value, ok := body.Meta["origin:attribute"]; ok {
				origin = value[0]
			}
			emptyView := ""
			switch {
			case origin != "":
				collectResponseWireType(body, endpoint, server, true, &emptyView)
			case endpoint.MethodExpr.Result.Meta != nil:
				if view, ok := endpoint.MethodExpr.Result.Meta.Last(expr.ViewMetaKey); ok {
					collectResponseWireType(body, endpoint, server, true, &view)
				} else {
					for _, view := range resultType.Views {
						collectResponseWireType(body, endpoint, server, true, &view.Name)
					}
				}
			default:
				for _, view := range resultType.Views {
					collectResponseWireType(body, endpoint, server, true, &view.Name)
				}
			}
			clientView := clientResponseViewNameExpr(endpoint, resultType)
			if origin != "" {
				emptyView := ""
				collectResponseWireType(body, endpoint, client, false, &emptyView)
				continue
			}
			if clientView == "" && !endpoint.UsesSSE() && !endpoint.IsJSONRPC() {
				emptyView := ""
				collectResponseWireType(body, endpoint, client, false, &emptyView)
				continue
			}
			if clientView != "" {
				clientBody := effectiveClientResponseBodyForView(body, clientView)
				collectResponseWireType(clientBody, endpoint, client, false, &clientView)
				continue
			}
			for _, view := range resultType.Views {
				clientBody := effectiveClientResponseBodyForView(body, view.Name)
				collectResponseWireType(clientBody, endpoint, client, false, &view.Name)
			}
		}
		for _, transportError := range endpoint.HTTPErrors {
			body := bodies.errorResponse(transportError)
			collectResponseWireType(body, endpoint, server, true, nil)
			collectResponseWireType(body, endpoint, client, false, nil)
		}
		collectPlannedTransforms(endpoint, bodies, servicePlan, server, client)
	}
}

// collectPlannedTransforms records each HTTP body conversion in the same order
// that Link writes it. This lets the generated package name every extra
// conversion function before Plan.Link.
func collectPlannedTransforms(endpoint *expr.HTTPEndpointExpr, bodies *shapedBodies, servicePlan *service.Plan, server, client *wireTypeCatalog) {
	methodName := endpoint.MethodExpr.Name
	request := expr.DupAtt(bodies.request(endpoint))
	addMarshalTags(request)
	payload := endpoint.MethodExpr.Payload
	if needInit(payload.Type) {
		if request.Type != expr.Empty {
			target := payload
			if origin, ok := request.Meta["origin:attribute"]; ok {
				target = expr.AsObject(payload.Type).Attribute(origin[0])
			}
			client.collectTransform(target, request, "marshal", methodName+" request body")
			server.collectTransform(request, target, "unmarshal", methodName+" server payload")
			client.collectTransform(request, target, "marshal", methodName+" command payload")
		} else if expr.IsArray(payload.Type) || expr.IsMap(payload.Type) {
			if params := expr.AsObject(endpoint.Params.Type); len(*params) > 0 {
				server.collectTransform((*params)[0].Attribute, payload, "unmarshal", methodName+" server parameters")
				client.collectTransform((*params)[0].Attribute, payload, "marshal", methodName+" command parameters")
			}
		}
	}

	result := endpoint.MethodExpr.Result
	resultType, viewed := result.Type.(*expr.ResultTypeExpr)
	if viewed {
		var err error
		result, err = servicePlan.ProjectedResult(endpoint.MethodExpr)
		if err != nil {
			panic(err)
		}
	}
	for _, response := range endpoint.Responses {
		body := bodies.response(response)
		origin := ""
		if value, ok := body.Meta["origin:attribute"]; ok {
			origin = value[0]
		}
		resultAttribute := result
		if origin != "" {
			resultAttribute = expr.AsObject(result.Type).Attribute(origin)
		}
		var serverViews []*string
		switch {
		case !viewed:
			serverViews = []*string{nil}
		case origin != "":
			empty := ""
			serverViews = []*string{&empty}
		case endpoint.MethodExpr.Result.Meta != nil:
			if view, ok := endpoint.MethodExpr.Result.Meta.Last(expr.ViewMetaKey); ok {
				serverViews = []*string{&view}
			} else {
				for index := range resultType.Views {
					serverViews = append(serverViews, &resultType.Views[index].Name)
				}
			}
		default:
			for index := range resultType.Views {
				serverViews = append(serverViews, &resultType.Views[index].Name)
			}
		}
		for _, view := range serverViews {
			prepared, _ := prepareResponseWireBody(body, view)
			if prepared.Type != expr.Empty && resultAttribute.Type != expr.Empty && needInit(prepared.Type) {
				server.collectTransform(resultAttribute, prepared, "marshal", transformResponseOwner(methodName, response, view, "server"))
			}
		}

		if !needInit(result.Type) {
			continue
		}
		var clientViews []*string
		if !viewed {
			clientViews = []*string{nil}
		} else {
			selected := clientResponseViewNameExpr(endpoint, resultType)
			switch {
			case origin != "":
				empty := ""
				clientViews = []*string{&empty}
			case selected != "":
				clientViews = []*string{&selected}
			case !endpoint.UsesSSE() && !endpoint.IsJSONRPC():
				empty := ""
				clientViews = []*string{&empty}
			default:
				for index := range resultType.Views {
					clientViews = append(clientViews, &resultType.Views[index].Name)
				}
			}
		}
		for _, view := range clientViews {
			clientBody := body
			if view != nil && *view != "" {
				clientBody = effectiveClientResponseBodyForView(body, *view)
			}
			prepared, _ := prepareResponseWireBody(clientBody, view)
			if prepared.Type != expr.Empty {
				client.collectTransform(prepared, resultAttribute, "unmarshal", transformResponseOwner(methodName, response, view, "client"))
			}
		}
		if body.Type == expr.Empty && (expr.IsArray(result.Type) || expr.IsMap(result.Type)) {
			if params := expr.AsObject(endpoint.QueryParams().Type); len(*params) > 0 {
				client.collectTransform((*params)[0].Attribute, result, "unmarshal", transformResponseOwner(methodName, response, nil, "client parameters"))
			}
		}
	}

	for _, transportError := range endpoint.HTTPErrors {
		body, _ := prepareResponseWireBody(bodies.errorResponse(transportError), nil)
		target := endpoint.MethodExpr.Error(transportError.Name).AttributeExpr
		if origin, ok := body.Meta["origin:attribute"]; ok {
			target = expr.AsObject(target.Type).Attribute(origin[0])
		}
		if body.Type != expr.Empty && needInit(transportError.Type) {
			server.collectTransform(target, body, "marshal", methodName+" server error "+transportError.Name)
			client.collectTransform(body, target, "unmarshal", methodName+" client error "+transportError.Name)
		} else if body.Type == expr.Empty && (expr.IsArray(transportError.Type) || expr.IsMap(transportError.Type)) {
			if params := expr.AsObject(endpoint.QueryParams().Type); len(*params) > 0 {
				client.collectTransform((*params)[0].Attribute, endpoint.MethodExpr.Error(transportError.Name).AttributeExpr, "unmarshal", methodName+" client error parameters "+transportError.Name)
			}
		}
	}

	if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty && endpoint.UsesWebSocket() {
		body := expr.DupAtt(bodies.streaming(endpoint))
		addMarshalTags(body)
		if body.Type != expr.Empty && needInit(endpoint.MethodExpr.StreamingPayload.Type) {
			server.collectTransform(body, endpoint.MethodExpr.StreamingPayload, "marshal", methodName+" server stream payload")
			client.collectTransform(endpoint.MethodExpr.StreamingPayload, body, "marshal", methodName+" client stream body")
		}
	}
}

// transformResponseOwner returns the design values that distinguish helper
// functions for two responses with the same generated Go types.
func transformResponseOwner(method string, response *expr.HTTPResponseExpr, view *string, side string) string {
	viewName := ""
	if view != nil {
		viewName = *view
	}
	return fmt.Sprintf("%s %s response %d %s %s %s", method, side, response.StatusCode, response.Tag[0], response.Tag[1], viewName)
}

// collectResponseWireType applies the selected view and records response body
// declarations using the same policy later consumed by buildResponseBodyType.
func collectResponseWireType(body *expr.AttributeExpr, endpoint *expr.HTTPEndpointExpr, catalog *wireTypeCatalog, server bool, view *string) {
	body, viewName := prepareResponseWireBody(body, view)
	policy := wireTypePolicy{pointer: !server, useDefault: server, validate: !server && view == nil, view: viewName}
	preferred := ""
	if server && !expr.IsPrimitive(body.Type) && needInit(body.Type) {
		if _, userType := body.Type.(expr.UserType); !userType {
			preferred = codegen.Goify(endpoint.Name(), true) + "ResponseBody"
		}
	}
	record := catalog.collect(body, wireResponseBody, policy, preferred)
	if server && record != nil && needInit(body.Type) {
		record.needsConstructor = true
	}
	attributePolicy := wireTypePolicy{pointer: !server, useDefault: server, validate: !server}
	catalog.collectChildren(body, wireAttribute, attributePolicy)
}

// prepareResponseWireBody returns the detached, projected, and tagged shape
// consumed by collection, declarations, and client response transforms.
func prepareResponseWireBody(body *expr.AttributeExpr, view *string) (*expr.AttributeExpr, string) {
	body = expr.DupAtt(body)
	viewName := ""
	if view != nil && *view != "" {
		viewName = *view
		if resultType, ok := body.Type.(*expr.ResultTypeExpr); ok {
			projected, err := expr.Project(resultType, *view)
			if err != nil {
				panic(err)
			}
			body.Type = projected
		}
	}
	addMarshalTags(body)
	return body, viewName
}

// makeHTTPType traverses the attribute recursively and performs these actions:
//
// * removes aliased user type by replacing them with the underlying type.
// * changes unions into structs with Type and Value fields.
func makeHTTPType(att *expr.AttributeExpr) *expr.AttributeExpr {
	att = expr.DupAtt(att)
	return makeHTTPTypeRecursive(att, make(map[expr.UserType]struct{}))
}

func makeHTTPTypeRecursive(att *expr.AttributeExpr, seen map[expr.UserType]struct{}) *expr.AttributeExpr {
	delete(att.Meta, "struct:pkg:path")
	switch dt := att.Type.(type) {
	case expr.UserType:
		if dt == expr.Empty {
			// Empty is a shared sentinel that expr.Dup deliberately never
			// duplicates: rewriting its attribute would mutate global design
			// state. There is nothing to flatten in it anyway.
			return att
		}
		if _, ok := dt.(*expr.ResultTypeExpr); !ok && !expr.IsObject(dt) {
			// Aliased user type. Use the underlying aliased type instead of
			// generating new types in the client and server packages
			att.Type = dt.Attribute().Type
			if v := dt.Attribute().Validation; v != nil {
				if att.Validation == nil {
					att.Validation = v
				} else {
					att.Validation.Merge(v)
				}
			}
			att.DefaultValue = dt.Attribute().DefaultValue
			att.UserExamples = dt.Attribute().UserExamples
		}
		origin := dt.Origin()
		if _, ok := seen[origin]; ok {
			return att
		}
		seen[origin] = struct{}{}
		dt.SetAttribute(makeHTTPTypeRecursive(dt.Attribute(), seen))
	case *expr.Array:
		dt.ElemType = makeHTTPTypeRecursive(dt.ElemType, seen)
	case *expr.Map:
		dt.KeyType = makeHTTPTypeRecursive(dt.KeyType, seen)
		dt.ElemType = makeHTTPTypeRecursive(dt.ElemType, seen)
	case *expr.Object:
		obj := make(expr.Object, len(*dt))
		for i, nat := range *dt {
			obj[i] = &expr.NamedAttributeExpr{Name: nat.Name, Attribute: makeHTTPTypeRecursive(nat.Attribute, seen)}
		}
		att.Type = &obj
	case *expr.Union:
		// Unions are represented as first-class sum types; HTTP uses the same
		// type for request and response bodies.
	}
	return att
}

// request returns the shaped HTTP request body for the given endpoint. The
// returned attribute is a detached copy of the design body: aliased user
// types are flattened and marshal tag meta may be added to it without
// affecting the design expression tree.
func (b *shapedBodies) request(e *expr.HTTPEndpointExpr) *expr.AttributeExpr {
	if att, ok := b.requests[e]; ok {
		return att
	}
	if b.requests == nil {
		b.requests = make(map[*expr.HTTPEndpointExpr]*expr.AttributeExpr)
	}
	att := makeHTTPType(e.Body)
	b.requests[e] = att
	return att
}

// streaming returns the streaming request body for the given endpoint. The
// returned attribute is a detached copy of the design body so that marshal
// tag meta may be added to it without affecting the design expression tree.
// Streaming bodies are not shaped with makeHTTPType: aliased user types have
// never been flattened in streaming bodies.
func (b *shapedBodies) streaming(e *expr.HTTPEndpointExpr) *expr.AttributeExpr {
	if att, ok := b.streams[e]; ok {
		return att
	}
	if b.streams == nil {
		b.streams = make(map[*expr.HTTPEndpointExpr]*expr.AttributeExpr)
	}
	att := expr.DupAtt(e.StreamingBody)
	expr.RemovePkgPath(att)
	b.streams[e] = att
	return att
}

// response returns the shaped HTTP body for the given success response, see
// request.
func (b *shapedBodies) response(resp *expr.HTTPResponseExpr) *expr.AttributeExpr {
	if att, ok := b.responses[resp]; ok {
		return att
	}
	if b.responses == nil {
		b.responses = make(map[*expr.HTTPResponseExpr]*expr.AttributeExpr)
	}
	att := makeHTTPType(resp.Body)
	b.responses[resp] = att
	return att
}

// errorResponse returns the shaped HTTP body for the given error response,
// see request.
func (b *shapedBodies) errorResponse(v *expr.HTTPErrorExpr) *expr.AttributeExpr {
	if att, ok := b.errors[v]; ok {
		return att
	}
	if b.errors == nil {
		b.errors = make(map[*expr.HTTPErrorExpr]*expr.AttributeExpr)
	}
	att := makeHTTPType(v.Response.Body)
	b.errors[v] = att
	return att
}

// buildPayloadData returns the data structure used to describe the endpoint
// payload including the HTTP request details. It also returns the user types
// used by the request body type recursively if any.
func (sds *ServicesData) buildPayloadData(e *expr.HTTPEndpointExpr, sd *ServiceData) *PayloadData {
	httpBody := sd.bodies.request(e)
	serverHTTPBody := expr.DupAtt(httpBody)
	clientHTTPBody := expr.DupAtt(httpBody)
	if httpBody.Type != expr.Empty {
		addMarshalTags(serverHTTPBody)
		addMarshalTags(clientHTTPBody)
		serverPolicy := wireTypePolicy{request: true, pointer: true, validate: true}
		clientPolicy := wireTypePolicy{request: true, useDefault: true, validate: true}
		sd.serverWireTypes.applyNames(serverHTTPBody, wireRequestBody, serverPolicy)
		sd.clientWireTypes.applyNames(clientHTTPBody, wireRequestBody, clientPolicy)
	}
	var (
		payload      = e.MethodExpr.Payload
		svc          = sd.Service
		body         = httpBody.Type
		ep           = svc.Method(e.MethodExpr.Name)
		httpsvrctx   = wireHTTPContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
		httpclictx   = wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, true, false)
		svcsvrctx    = sds.serviceTypeContext(sd, "server").Enter(payload)
		svcclictx    = sds.serviceTypeContext(sd, "client").Enter(payload)
		payloadOwner = expr.MethodPayloadExampleIdentity(e.MethodExpr)
		bodyOwner    = expr.RequestBodyExampleIdentity(e)

		request       *RequestData
		mapQueryParam *ParamData
	)
	{
		var (
			serverBodyData = sds.buildRequestBodyType(httpBody, payload, e, true, sd, payloadOwner, bodyOwner)
			clientBodyData = sds.buildRequestBodyType(httpBody, payload, e, false, sd, payloadOwner, bodyOwner)
			paramsData     = sds.extractPathParams(e.PathParams(), payload, sd, payloadOwner)
			queryData      = sds.extractQueryParams(e.QueryParams(), payload, sd, payloadOwner)
			headersData    = sds.extractHeaders(e.Headers, payload, svcsvrctx, sd.Scope, payloadOwner)
			cookiesData    = sds.extractCookies(e.Cookies, payload, svcsvrctx, sd.Scope, payloadOwner)
			origin         string

			mustValidate bool
			mustHaveBody = true
		)
		if e.MapQueryParams != nil {
			var (
				fieldName string
				name      = "query"
				required  = true
				pAtt      = payload
			)
			if n := *e.MapQueryParams; n != "" {
				pAtt = expr.AsObject(payload.Type).Attribute(n)
				required = payload.IsRequired(n)
				name = n
				fieldName = codegen.Goify(name, true)
			}
			varn := codegen.Goify(name, false)
			mapQueryParam = &ParamData{
				MapQueryParams: e.MapQueryParams,
				Map:            expr.AsMap(payload.Type) != nil,
				Element: &Element{
					HTTPName: name,
					AttributeData: &AttributeData{
						Name:         name,
						VarName:      varn,
						FieldName:    fieldName,
						FieldType:    pAtt.Type,
						Required:     required,
						Type:         pAtt.Type,
						TypeName:     sd.Scope.GoTypeName(pAtt),
						TypeRef:      sd.Scope.GoTypeRef(pAtt),
						Validate:     codegen.AttributeValidationCode(pAtt, nil, httpsvrctx, required, expr.IsAlias(pAtt.Type), varn, name),
						DefaultValue: pAtt.DefaultValue,
						Example:      sds.FieldExample(pAtt, e.MethodExpr.Payload, name, payloadOwner),
					},
				},
			}
			queryData = append(queryData, mapQueryParam)
		}
		for _, p := range cookiesData {
			if p.Required || p.Validate != "" || needConversion(p.Type) {
				mustValidate = true
				break
			}
		}
		if !mustValidate {
			for _, p := range paramsData {
				if p.Validate != "" || needConversion(p.Type) || p.IsTextUnmarshaler {
					mustValidate = true
					break
				}
			}
		}
		if !mustValidate {
			for _, q := range queryData {
				if q.Map || q.Validate != "" || q.Required || needConversion(q.Type) || q.IsTextUnmarshaler {
					mustValidate = true
					break
				}
			}
		}
		if !mustValidate {
			for _, h := range headersData {
				if h.Validate != "" || h.Required || needConversion(h.Type) {
					mustValidate = true
					break
				}
			}
		}
		if body != expr.Empty {
			// If design uses Body("name") syntax we need to use the
			// corresponding attribute in the result type for body
			// transformation.
			if o, ok := serverHTTPBody.Meta["origin:attribute"]; ok {
				origin = o[0]
				if !payload.IsRequired(o[0]) {
					mustHaveBody = false
				}
			}
		}
		request = &RequestData{
			PathParams:   paramsData,
			QueryParams:  queryData,
			Headers:      headersData,
			Cookies:      cookiesData,
			ServerBody:   serverBodyData,
			ClientBody:   clientBodyData,
			PayloadAttr:  codegen.Goify(origin, true),
			PayloadType:  e.MethodExpr.Payload.Type,
			MustHaveBody: mustHaveBody,
			MustValidate: mustValidate,
			Multipart:    e.MultipartRequest,
		}
	}

	var init *InitData
	if needInit(payload.Type) {
		// generate constructor function to transform request body,
		// params, headers and cookies into the method payload type
		var (
			name       string
			desc       string
			isObject   bool
			clientArgs []*InitArgData
			serverArgs []*InitArgData
		)
		argsCap := len(request.PathParams) + len(request.QueryParams) + len(request.Headers) + len(request.Cookies)
		declaration := sds.payloadConstructors[e]
		if declaration == nil {
			panic(fmt.Sprintf("payload constructor for %s.%s was not submitted", svc.Name, e.Name()))
		}
		name = declaration.Name()
		desc = fmt.Sprintf("%s builds a %s service %s endpoint payload.",
			name, svc.Name, e.Name())
		isObject = expr.IsObject(payload.Type)
		serverArgs = make([]*InitArgData, 0, argsCap+1)
		clientArgs = make([]*InitArgData, 0, argsCap+1)
		if body != expr.Empty {
			var (
				svcode         string
				cvcode         string
				serverTypeName string
				serverTypeRef  string
				clientTypeName string
				clientTypeRef  string
			)
			if record := sd.serverWireTypes.lookupUser(serverHTTPBody, wireRequestBody, wireTypePolicy{request: true, pointer: true, validate: true}); record != nil {
				serverTypeName = record.name
				serverTypeRef = record.ref
			} else {
				serverTypeName = httpsvrctx.Scope.Name(serverHTTPBody, "", httpsvrctx.Pointer, httpsvrctx.UseDefault)
				serverTypeRef = httpsvrctx.Scope.Ref(serverHTTPBody, "")
			}
			if record := sd.clientWireTypes.lookupUser(clientHTTPBody, wireRequestBody, wireTypePolicy{request: true, useDefault: true, validate: true}); record != nil {
				clientTypeName = record.name
				clientTypeRef = record.ref
			} else {
				clientTypeName = httpclictx.Scope.Name(clientHTTPBody, "", httpclictx.Pointer, httpclictx.UseDefault)
				clientTypeRef = httpclictx.Scope.Ref(clientHTTPBody, "")
			}
			if ut, ok := serverHTTPBody.Type.(expr.UserType); ok {
				if val := ut.Attribute().Validation; val != nil {
					svcode = codegen.ValidationCode(ut.Attribute(), ut, httpsvrctx, true, expr.IsAlias(ut), false, "body")
				}
			}
			if ut, ok := clientHTTPBody.Type.(expr.UserType); ok {
				if val := ut.Attribute().Validation; val != nil {
					cvcode = codegen.ValidationCode(ut.Attribute(), ut, httpclictx, true, expr.IsAlias(ut), false, "body")
				}
			}
			serverArgs = append(serverArgs, &InitArgData{
				Ref: sd.serverWireTypes.scope.GoVar("body", serverHTTPBody.Type),
				AttributeData: &AttributeData{
					Name:     "body",
					VarName:  "body",
					TypeName: serverTypeName,
					TypeRef:  serverTypeRef,
					Type:     serverHTTPBody.Type,
					Required: true,
					Example:  sds.Example(httpBody, bodyOwner),
					Validate: svcode,
				},
			})
			clientArgs = append(clientArgs, &InitArgData{
				Ref: sd.clientWireTypes.scope.GoVar("body", clientHTTPBody.Type),
				AttributeData: &AttributeData{
					Name:     "body",
					VarName:  "body",
					TypeName: clientTypeName,
					TypeRef:  clientTypeRef,
					Type:     clientHTTPBody.Type,
					Required: true,
					Example:  sds.Example(httpBody, bodyOwner),
					Validate: cvcode,
				},
			})
		}
		args := make([]*InitArgData, 0, argsCap)
		for _, p := range request.PathParams {
			arg := elementInitArg(p.Element)
			// Path parameter flags never carry a default value in the
			// generated CLI code.
			arg.DefaultValue = nil
			args = append(args, arg)
		}
		// Query string, header and cookie flags never carry a description in
		// the generated CLI code (only path parameter flags do).
		for _, p := range request.QueryParams {
			arg := elementInitArg(p.Element)
			arg.Description = ""
			args = append(args, arg)
		}
		for _, h := range request.Headers {
			arg := elementInitArg(h.Element)
			arg.Description = ""
			args = append(args, arg)
		}
		for _, c := range request.Cookies {
			arg := elementInitArg(c.Element)
			arg.Description = ""
			args = append(args, arg)
		}
		serverArgs = append(serverArgs, args...)
		clientArgs = append(clientArgs, args...)

		var (
			cliArgs []*InitArgData
		)
		for _, r := range ep.Requirements {
			done := false
			for _, sc := range r.Schemes {
				if sc.Type == "Basic" {
					uatt := e.MethodExpr.Payload.Find(sc.UsernameAttr)
					uctx := svcclictx.Enter(uatt)
					uref := uctx.Scope.Ref(uatt, uctx.Pkg(uatt))
					if sc.UsernamePointer {
						uref = "*" + uref
					}
					uarg := &InitArgData{
						Ref: sc.UsernameAttr,
						AttributeData: &AttributeData{
							Name:         sc.UsernameAttr,
							VarName:      sc.UsernameAttr,
							FieldName:    sc.UsernameField,
							FieldPointer: sc.UsernamePointer,
							FieldType:    uatt.Type,
							Description:  uatt.Description,
							Required:     sc.UsernameRequired,
							TypeName:     uctx.Scope.Name(uatt, uctx.Pkg(uatt), false, true),
							TypeRef:      uref,
							Type:         uatt.Type,
							Pointer:      sc.UsernamePointer,
							Validate:     codegen.ValidationCode(uatt, nil, httpsvrctx, sc.UsernameRequired, expr.IsAlias(uatt.Type), false, sc.UsernameAttr),
							Example:      sds.FieldExample(uatt, e.MethodExpr.Payload, sc.UsernameAttr, payloadOwner),
						},
					}
					patt := e.MethodExpr.Payload.Find(sc.PasswordAttr)
					pctx := svcclictx.Enter(patt)
					pref := pctx.Scope.Ref(patt, pctx.Pkg(patt))
					if sc.PasswordPointer {
						pref = "*" + pref
					}
					parg := &InitArgData{
						Ref: sc.PasswordAttr,
						AttributeData: &AttributeData{
							Name:         sc.PasswordAttr,
							VarName:      sc.PasswordAttr,
							FieldName:    sc.PasswordField,
							FieldPointer: sc.PasswordPointer,
							FieldType:    patt.Type,
							Description:  patt.Description,
							Required:     sc.PasswordRequired,
							TypeName:     pctx.Scope.Name(patt, pctx.Pkg(patt), false, true),
							TypeRef:      pref,
							Type:         patt.Type,
							Pointer:      sc.PasswordPointer,
							Validate:     codegen.ValidationCode(patt, nil, httpsvrctx, sc.PasswordRequired, expr.IsAlias(patt.Type), false, sc.PasswordAttr),
							Example:      sds.FieldExample(patt, e.MethodExpr.Payload, sc.PasswordAttr, payloadOwner),
						},
					}
					cliArgs = []*InitArgData{uarg, parg}
					done = true
					break
				}
			}
			if done {
				break
			}
		}

		var (
			serverCode string
			clientCode string
			err        error
			origin     string
			pointer    bool

			pAtt = payload
		)
		if body != expr.Empty {
			// If design uses Body("name") syntax then need to use payload
			// attribute to transform.
			if o, ok := httpBody.Meta["origin:attribute"]; ok {
				origin = o[0]
				pAtt = expr.AsObject(payload.Type).Attribute(origin)
				pointer = !payload.IsRequired(o[0]) && expr.IsPrimitive(pAtt.Type)
			}

			var (
				helpers []*codegen.TransformFunctionData
			)
			transformctx := wireHTTPContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
			serverCode, helpers, err = sd.serverWireTypes.renderTransform(serverHTTPBody, pAtt, "body", "v", "unmarshal", transformctx, svcsvrctx)
			if err == nil {
				sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
			}
			// The client code for building the method payload from a request
			// body is used by the CLI tool to build the payload given to the
			// client endpoint. It differs because the body type there does not
			// use pointers for all fields (no need to validate).
			transformctx = wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, true, false)
			clientCode, helpers, err = sd.clientWireTypes.renderTransform(clientHTTPBody, pAtt, "body", "v", "marshal", transformctx, svcclictx)
			if err == nil {
				sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
			}
		} else if expr.IsArray(payload.Type) || expr.IsMap(payload.Type) {
			if params := expr.AsObject(e.Params.Type); len(*params) > 0 {
				var helpers []*codegen.TransformFunctionData
				transformctx := wireHTTPContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
				serverCode, helpers, err = sd.serverWireTypes.renderTransform((*params)[0].Attribute, payload, codegen.Goify((*params)[0].Name, false), "v", "unmarshal", transformctx, svcsvrctx)
				if err == nil {
					sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
				}
				transformctx = wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, true, false)
				clientCode, helpers, err = sd.clientWireTypes.renderTransform((*params)[0].Attribute, payload, codegen.Goify((*params)[0].Name, false), "v", "marshal", transformctx, svcclictx)
				if err == nil {
					sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
				}
			}
		}
		if err != nil {
			panic(err) // bug
		}
		init = &InitData{
			Declaration:              declaration,
			Name:                     name,
			Description:              desc,
			ServerArgs:               serverArgs,
			ClientArgs:               clientArgs,
			CLIArgs:                  cliArgs,
			ReturnTypeName:           svcsvrctx.Scope.Name(payload, svcsvrctx.Pkg(payload), false, true),
			ReturnTypeRef:            svcsvrctx.Scope.Ref(payload, svcsvrctx.Pkg(payload)),
			ReturnIsStruct:           isObject,
			ReturnTypeAttribute:      codegen.Goify(origin, true),
			ReturnTypePkg:            svcsvrctx.Pkg(payload),
			ServerCode:               serverCode,
			ClientCode:               clientCode,
			ReturnIsPrimitivePointer: pointer,
		}
	}
	request.PayloadInit = init

	var (
		returnValue string
		name        string
		ref         string
	)
	if payload.Type != expr.Empty {
		name = svcsvrctx.Scope.Name(payload, svcsvrctx.Pkg(payload), false, true)
		ref = svcsvrctx.Scope.Ref(payload, svcsvrctx.Pkg(payload))
	}
	if init == nil {
		if o := expr.AsObject(e.Params.Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if o := expr.AsObject(e.Headers.Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if o := expr.AsObject(e.Cookies.Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if e.MapQueryParams != nil && *e.MapQueryParams == "" {
			returnValue = mapQueryParam.VarName
		}
	}
	data := &PayloadData{
		Name:               name,
		Ref:                ref,
		Request:            request,
		DecoderReturnValue: returnValue,
	}
	if e.IsJSONRPC() {
		obj := expr.AsObject(e.MethodExpr.Payload.Type)
		if obj != nil {
			for _, att := range *obj {
				if _, ok := att.Attribute.Meta["jsonrpc:id"]; ok {
					data.IDAttribute = codegen.Goify(att.Name, true)
					data.IDAttributeRequired = e.MethodExpr.Payload.IsRequired(att.Name)
					break
				}
			}
		}
	}
	return data
}

// buildResultData builds the result data for the given service endpoint.
func (sds *ServicesData) buildResultData(e *expr.HTTPEndpointExpr, sd *ServiceData) *ResultData {
	var (
		result = e.MethodExpr.Result
		method = sd.Service.Method(e.MethodExpr.Name)
		svcctx = sds.serviceTypeContext(sd, "server").Enter(result)

		name string
		ref  string
		view string
	)

	view = expr.DefaultView
	if v, ok := result.Meta.Last(expr.ViewMetaKey); ok {
		view = v
	}
	if result.Type != expr.Empty {
		name = svcctx.Scope.Name(result, svcctx.Pkg(result), false, true)
		ref = svcctx.Scope.Ref(result, svcctx.Pkg(result))
	}

	var (
		mustInit  bool
		responses []*ResponseData
	)
	{
		viewed := false
		if method.ViewedResult != nil {
			result = expr.AsObject(method.ViewedResult.Type).Attribute("projected")
			viewed = true
		}
		responses = sds.buildResponses(e, result, viewed, sd)
		for _, r := range responses {
			// response has a body, headers, cookies or tag
			if len(r.ServerBody) > 0 || len(r.Headers) > 0 || len(r.Cookies) > 0 || r.TagName != "" {
				mustInit = true
			}
		}
	}
	idAtt := ""
	idAttRequired := false
	if e.IsJSONRPC() && result.Type != expr.Empty {
		obj := expr.AsObject(result.Type)
		if obj != nil {
			for _, att := range *obj {
				if _, ok := att.Attribute.Meta["jsonrpc:id"]; ok {
					idAtt = codegen.Goify(att.Name, true)
					idAttRequired = result.IsRequired(att.Name)
					break
				}
			}
		}
	}
	return &ResultData{
		IsStruct:            expr.IsObject(result.Type),
		Name:                name,
		Ref:                 ref,
		IDAttribute:         idAtt,
		IDAttributeRequired: idAttRequired,
		Responses:           responses,
		View:                view,
		MustInit:            mustInit,
	}
}

// buildResponses builds the response data for all the responses in the endpoint
// expression. The response headers, cookies and body for each response are
// inferred from the method's result expression if not specified explicitly.
//
// viewed parameter indicates if the method result uses views.
func (sds *ServicesData) buildResponses(e *expr.HTTPEndpointExpr, result *expr.AttributeExpr, viewed bool, sd *ServiceData) []*ResponseData {
	var (
		responses []*ResponseData

		svc    = sd.Service
		md     = svc.Method(e.Name())
		scope  = svc.Scope
		svcctx = sds.serviceTypeContext(sd, "client").Enter(result)
	)
	{
		if viewed {
			scope = svc.ViewScope
			svcctx = sds.viewTypeContext(sd, "client").Enter(result)
		}
		notag := -1
		for i, resp := range e.Responses {
			respBody := sd.bodies.response(resp)
			resultOwner := expr.MethodResultExampleIdentity(e.MethodExpr)
			bodyOwner := expr.ResponseBodyExampleIdentity(e, resp)
			if resp.Tag[0] == "" {
				if notag > -1 {
					continue // we don't want more than one response with no tag
				}
				notag = i
			}
			var (
				headersData    []*HeaderData
				cookiesData    []*CookieData
				serverBodyData []*TypeData
				clientBodyData *TypeData
				init           *InitData
				origin         string
				mustValidate   bool
				clientRespBody = respBody
				clientBodyView *string

				resAttr = result
			)
			{
				headersData = sds.extractHeaders(resp.Headers, result, svcctx, scope, resultOwner)
				cookiesData = sds.extractCookies(resp.Cookies, result, svcctx, scope, resultOwner)
				if respBody.Type != expr.Empty {
					// If design uses Body("name") syntax we need to use the
					// corresponding attribute in the result type for body
					// transformation.
					if o, ok := respBody.Meta["origin:attribute"]; ok {
						origin = o[0]
						resAttr = expr.AsObject(resAttr.Type).Attribute(origin)
					}
				}
				if viewed {
					vname := ""
					clientView := clientResponseViewName(e, md)
					if origin != "" {
						// Response body is explicitly set to an attribute in the method
						// result type. No need to do any view-based projections server side.
						if sbd := sds.buildResponseBodyType(respBody, result, e, true, &vname, sd, resultOwner, bodyOwner); sbd != nil {
							serverBodyData = append(serverBodyData, sbd)
						}
					} else if v, ok := e.MethodExpr.Result.Meta.Last(expr.ViewMetaKey); ok {
						// Design explicitly sets the view to render the result.
						// We generate only one server body type which will be rendered
						// using the specified view.
						if sbd := sds.buildResponseBodyType(respBody, result, e, true, &v, sd, resultOwner, bodyOwner); sbd != nil {
							serverBodyData = append(serverBodyData, sbd)
						}
					} else {
						// If a method result uses views (i.e., a result type), we generate
						// one response body type per view defined in the result type. The
						// generated body type names are suffixed with the name of the view
						// (except for the "default" view). Constructors are also generated
						// to create a view-specific body type from the method result.
						// This makes it possible for the server side to return only the
						// attributes defined in the view in the response (NOTE: a required
						// attribute in the result type may not be present in all its views)
						for _, view := range md.ViewedResult.Views {
							if sbd := sds.buildResponseBodyType(respBody, result, e, true, &view.Name, sd, resultOwner, bodyOwner); sbd != nil {
								serverBodyData = append(serverBodyData, sbd)
							}
						}
					}
					if clientView != "" {
						clientRespBody = effectiveClientResponseBodyForView(respBody, clientView)
						clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, &clientView, sd, resultOwner, bodyOwner)
						clientBodyView = &clientView
					} else if origin != "" || !e.UsesSSE() && !e.IsJSONRPC() {
						clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, &vname, sd, resultOwner, bodyOwner)
						clientBodyView = &vname
					} else {
						clientRespBody = &expr.AttributeExpr{Type: expr.Empty}
					}
				} else {
					if sbd := sds.buildResponseBodyType(respBody, result, e, true, nil, sd, resultOwner, bodyOwner); sbd != nil {
						serverBodyData = append(serverBodyData, sbd)
					}
					clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, nil, sd, resultOwner, bodyOwner)
				}
				if clientBodyData != nil && clientRespBody.Type != expr.Empty {
					var viewName string
					clientRespBody, viewName = prepareResponseWireBody(clientRespBody, clientBodyView)
					policy := wireTypePolicy{pointer: true, validate: clientBodyView == nil, view: viewName}
					sd.clientWireTypes.applyNames(clientRespBody, wireResponseBody, policy)
				}
				for _, h := range headersData {
					if h.Validate != "" || h.Required || needConversion(h.Type) {
						mustValidate = true
						break
					}
				}
				for _, c := range cookiesData {
					if c.Validate != "" || c.Required || needConversion(c.Type) {
						mustValidate = true
						break
					}
				}
				variableWire := viewed && origin == "" && clientResponseViewName(e, md) == "" && (e.UsesSSE() || e.IsJSONRPC())
				if needInit(result.Type) && !variableWire {
					init = sds.buildResponseResultInit(
						e, resp, result, resAttr, clientRespBody, origin,
						headersData, cookiesData, sd, "", clientBodyData,
					)
				}

				var representations []*ViewedRepresentationData
				if viewed && (e.UsesSSE() || e.IsJSONRPC()) {
					clientView := clientResponseViewName(e, md)
					if origin != "" {
						views := md.ViewedResult.Views
						if clientView != "" {
							views = []*service.ViewData{{Name: clientView}}
						}
						for _, view := range views {
							representation := &ViewedRepresentationData{
								View:       view.Name,
								ResultAttr: codegen.Goify(origin, true),
								ClientBody: clientBodyData,
								ResultInit: init,
							}
							if len(serverBodyData) > 0 {
								representation.ServerBody = serverBodyData[0]
							}
							representations = append(representations, representation)
						}
					} else {
						if clientView != "" {
							representation := &ViewedRepresentationData{
								View:       clientView,
								ResultAttr: codegen.Goify(origin, true),
								ClientBody: clientBodyData,
								ResultInit: init,
							}
							if len(serverBodyData) > 0 {
								representation.ServerBody = viewedServerBody(serverBodyData, clientView)
							}
							representations = append(representations, representation)
						}
						for _, view := range md.ViewedResult.Views {
							if clientView != "" {
								break
							}
							viewName := view.Name
							body := effectiveClientResponseBodyForView(respBody, viewName)
							clientBody := sds.buildResponseBodyType(
								respBody, result, e, false, &viewName, sd, resultOwner, bodyOwner,
							)
							if body.Type != expr.Empty {
								policy := wireTypePolicy{pointer: true, view: viewName}
								sd.clientWireTypes.applyNames(body, wireResponseBody, policy)
							}
							resultInit := sds.buildResponseResultInit(
								e, resp, result, resAttr, body, origin,
								headersData, cookiesData, sd, viewName, clientBody,
							)
							representation := &ViewedRepresentationData{
								View:       viewName,
								ResultAttr: codegen.Goify(origin, true),
								ClientBody: clientBody,
								ResultInit: resultInit,
							}
							if len(serverBodyData) > 0 {
								representation.ServerBody = viewedServerBody(serverBodyData, viewName)
							}
							representations = append(representations, representation)
						}
					}
				}

				var (
					tagName string
					tagVal  string
					tagPtr  bool
				)
				if resp.Tag[0] != "" {
					tagName = codegen.Goify(resp.Tag[0], true)
					tagVal = resp.Tag[1]
					tagPtr = viewed || result.IsPrimitivePointer(resp.Tag[0], true)
				}
				responses = append(responses, &ResponseData{
					StatusCode:            statusCodeToHTTPConst(resp.StatusCode),
					Description:           resp.Description,
					Headers:               headersData,
					Cookies:               cookiesData,
					ContentType:           resp.ContentType,
					ServerBody:            serverBodyData,
					ClientBody:            clientBodyData,
					ResultInit:            init,
					TagName:               tagName,
					TagValue:              tagVal,
					TagPointer:            tagPtr,
					MustValidate:          mustValidate,
					ResultAttr:            codegen.Goify(origin, true),
					ViewedResult:          md.ViewedResult,
					ViewedRepresentations: representations,
				})
			}
		}
		count := len(responses)
		if notag >= 0 && notag < count-1 {
			// Make sure tagless response is last
			responses[notag], responses[count-1] = responses[count-1], responses[notag]
		}
	}
	return responses
}

// buildResponseResultInit builds the data used to write one client result
// function. It uses the name chosen by NewPlans and converts the decoded HTTP
// body, headers, and cookies into the method result.
func (sds *ServicesData) buildResponseResultInit(e *expr.HTTPEndpointExpr, resp *expr.HTTPResponseExpr, result, resAttr, clientBody *expr.AttributeExpr, origin string, headers []*HeaderData, cookies []*CookieData, sd *ServiceData, view string, bodyType *TypeData) *InitData {
	var (
		svc        = sd.Service
		md         = svc.Method(e.Name())
		httpclictx = wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
		svcctx     = sds.serviceTypeContext(sd, "client").Enter(result)
	)
	if md.ViewedResult != nil {
		svcctx = sds.viewTypeContext(sd, "client").Enter(result)
	}
	tname := svcctx.Scope.Name(result, svcctx.Pkg(result), false, true)
	tref := svcctx.Scope.Ref(result, svcctx.Pkg(result))
	status := codegen.Goify(http.StatusText(resp.StatusCode), true)
	declaration := sds.viewedResultConstructors[viewedConstructorKey{endpoint: e, response: resp, view: view}]
	if declaration == nil {
		panic(fmt.Sprintf("result constructor for %s.%s view %q was not submitted", svc.Name, e.Name(), view))
	}
	name := declaration.Name()
	desc := fmt.Sprintf("%s builds a %q service %q endpoint result from a HTTP %q response.", name, svc.Name, e.Name(), status)

	var (
		code       string
		pointer    bool
		clientArgs []*InitArgData
	)
	if clientBody.Type != expr.Empty {
		if origin != "" {
			pointer = svcctx.IsPrimitivePointer(origin, result)
		}
		ref := "body"
		if expr.IsObject(clientBody.Type) {
			ref = "&body"
			pointer = false
		}
		var validate string
		if ut, ok := clientBody.Type.(expr.UserType); ok && ut.Attribute().Validation != nil {
			validate = codegen.ValidationCode(ut.Attribute(), ut, httpclictx, true, expr.IsAlias(ut), false, "body")
		}
		bodyTypeRef := bodyType.Ref
		if bodyTypeRef == "" {
			bodyTypeRef = bodyType.VarName
		}
		clientArgs = []*InitArgData{{
			Ref: ref,
			AttributeData: &AttributeData{
				Name:     "body",
				VarName:  "body",
				TypeRef:  bodyTypeRef,
				Validate: validate,
			},
		}}
		transformctx := wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
		transformctx.Scope = sd.clientWireTypes.resolver(sd.clientWireTypes.scope, wireTypePolicy{
			pointer: transformctx.Pointer,
			view:    bodyType.View,
		})
		converted, helpers, err := sd.clientWireTypes.renderTransform(clientBody, resAttr, "body", "v", "unmarshal", transformctx, svcctx)
		if err != nil {
			panic(err) // bug
		}
		code = converted
		sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
	} else if expr.IsArray(result.Type) || expr.IsMap(result.Type) {
		if params := expr.AsObject(e.QueryParams().Type); len(*params) > 0 {
			converted, helpers, err := sd.clientWireTypes.renderTransform((*params)[0].Attribute, result, codegen.Goify((*params)[0].Name, false), "v", "unmarshal", httpclictx, svcctx)
			if err != nil {
				panic(err) // bug
			}
			code = converted
			sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
		}
	}
	for _, header := range headers {
		clientArgs = append(clientArgs, resultInitArg(header.Element))
	}
	for _, cookie := range cookies {
		clientArgs = append(clientArgs, resultInitArg(cookie.Element))
	}
	return &InitData{
		Declaration:              declaration,
		Name:                     name,
		Description:              desc,
		ClientArgs:               clientArgs,
		ReturnTypeName:           tname,
		ReturnTypeRef:            tref,
		ReturnIsStruct:           expr.IsObject(result.Type),
		ReturnTypeAttribute:      codegen.Goify(origin, true),
		ReturnTypePkg:            svcctx.Pkg(result),
		ReturnIsPrimitivePointer: pointer,
		ClientCode:               code,
	}
}

// buildErrorsData builds the error data for all the error responses in the
// endpoint expression. The response headers, cookies and body for each response
// are inferred from the method's error expression if not specified explicitly.
func (sds *ServicesData) buildErrorsData(e *expr.HTTPEndpointExpr, sd *ServiceData) []*ErrorGroupData {
	var (
		svc        = sd.Service
		httpclictx = wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
	)

	data := make(map[string][]*ErrorData)
	for _, v := range e.HTTPErrors {
		respBody := expr.DupAtt(sd.bodies.errorResponse(v))
		addMarshalTags(respBody)
		errorAttribute := e.MethodExpr.Error(v.Name).AttributeExpr
		errorOwner := expr.MethodErrorExampleIdentity(e.MethodExpr, v.ErrorExpr)
		bodyOwner := expr.ErrorResponseBodyExampleIdentity(e, v)
		var (
			init *InitData
			body = respBody.Type
		)

		errctx := sds.serviceTypeContext(sd, "client").Enter(errorAttribute)

		if needInit(v.Type) {
			var (
				name     string
				desc     string
				isObject bool
				args     []*InitArgData
			)
			declaration := sds.errorConstructors[v]
			if declaration == nil {
				panic(fmt.Sprintf("error constructor for %s.%s error %q was not submitted", svc.Name, e.Name(), v.Name))
			}
			name = declaration.Name()
			desc = fmt.Sprintf("%s builds a %s service %s endpoint %s error.",
				name, svc.Name, e.Name(), v.ErrorExpr.Name)
			headers := sds.extractHeaders(v.Response.Headers, errorAttribute, errctx, sd.Scope, errorOwner)
			cookies := sds.extractCookies(v.Response.Cookies, errorAttribute, errctx, sd.Scope, errorOwner)
			argsCap := len(headers) + len(cookies)
			if body != expr.Empty {
				argsCap++
			}
			args = make([]*InitArgData, 0, argsCap)
			if body != expr.Empty {
				isObject = expr.IsObject(body)
				ref := "body"
				if isObject {
					ref = "&body"
				}
				policy := wireTypePolicy{pointer: true, validate: true}
				bodyRecord := sd.clientWireTypes.lookupUser(respBody, wireResponseBody, policy)
				sd.clientWireTypes.applyNames(respBody, wireResponseBody, policy)
				var bodyTypeRef string
				if bodyRecord != nil {
					bodyTypeRef = bodyRecord.ref
				} else {
					bodyTypeRef = httpclictx.Scope.Ref(respBody, "")
				}
				args = append(args, &InitArgData{
					Ref:           ref,
					AttributeData: &AttributeData{Name: "body", VarName: "body", TypeRef: bodyTypeRef},
				})
			}
			for _, h := range headers {
				args = append(args, errorInitArg(h.Element))
			}
			for _, c := range cookies {
				args = append(args, errorInitArg(c.Element))
			}

			var (
				code   string
				origin string
				err    error
			)
			if body != expr.Empty {
				eAtt := errorAttribute
				// If design uses Body("name") syntax then need to use payload
				// attribute to transform.
				if o, ok := respBody.Meta["origin:attribute"]; ok {
					origin = o[0]
					eAtt = expr.AsObject(v.ErrorExpr.Type).Attribute(origin)
				}

				var helpers []*codegen.TransformFunctionData
				transformctx := wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
				code, helpers, err = sd.clientWireTypes.renderTransform(respBody, eAtt, "body", "v", "unmarshal", transformctx, errctx)
				if err == nil {
					sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
				}
			} else if expr.IsArray(v.Type) || expr.IsMap(v.Type) {
				if params := expr.AsObject(e.QueryParams().Type); len(*params) > 0 {
					var helpers []*codegen.TransformFunctionData
					code, helpers, err = sd.clientWireTypes.renderTransform((*params)[0].Attribute, errorAttribute, codegen.Goify((*params)[0].Name, false), "v", "unmarshal", httpclictx, errctx)
					if err == nil {
						sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
					}
				}
			}
			if err != nil {
				panic(err) // bug
			}

			init = &InitData{
				Declaration:         declaration,
				Name:                name,
				Description:         desc,
				ClientArgs:          args,
				ReturnTypeName:      errctx.Scope.Name(errorAttribute, errctx.Pkg(errorAttribute), false, true),
				ReturnTypeRef:       errctx.Scope.Ref(errorAttribute, errctx.Pkg(errorAttribute)),
				ReturnIsStruct:      expr.IsObject(v.Type),
				ReturnTypeAttribute: codegen.Goify(origin, true),
				ReturnTypePkg:       errctx.Pkg(errorAttribute),
				ClientCode:          code,
			}
		}

		var (
			responseData *ResponseData
		)
		{
			var (
				serverBodyData []*TypeData
				clientBodyData *TypeData
			)
			{
				if sbd := sds.buildResponseBodyType(respBody, errorAttribute, e, true, nil, sd, errorOwner, bodyOwner); sbd != nil {
					serverBodyData = append(serverBodyData, sbd)
				}
				clientBodyData = sds.buildResponseBodyType(respBody, errorAttribute, e, false, nil, sd, errorOwner, bodyOwner)
				if clientBodyData != nil {
					clientBodyData.Description = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP response body for the %q error.",
						clientBodyData.VarName, svc.Name, e.Name(), v.Name)
					serverBodyData[0].Description = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP response body for the %q error.",
						serverBodyData[0].VarName, svc.Name, e.Name(), v.Name)
				}
			}

			headers := sds.extractHeaders(v.Response.Headers, errorAttribute, errctx, sd.Scope, errorOwner)
			cookies := sds.extractCookies(v.Response.Cookies, errorAttribute, errctx, sd.Scope, errorOwner)
			var mustValidate bool
			for _, h := range headers {
				if h.Validate != "" || h.Required || needConversion(h.Type) {
					mustValidate = true
					break
				}
			}
			for _, c := range cookies {
				if c.Validate != "" || c.Required || needConversion(c.Type) {
					mustValidate = true
					break
				}
			}
			var contentType string
			if v.Response.ContentType != expr.ErrorResultIdentifier {
				contentType = v.Response.ContentType
			}
			responseData = &ResponseData{
				StatusCode:   statusCodeToHTTPConst(v.Response.StatusCode),
				Code:         v.Response.StatusCode,
				Headers:      headers,
				ContentType:  contentType,
				Cookies:      cookies,
				ErrorHeader:  v.Name,
				ServerBody:   serverBodyData,
				ClientBody:   clientBodyData,
				ResultInit:   init,
				MustValidate: mustValidate,
			}
		}

		ref := errctx.Scope.Ref(errorAttribute, errctx.Pkg(errorAttribute))
		data[ref] = append(data[ref], &ErrorData{
			Name:     v.Name,
			Response: responseData,
			Ref:      ref,
		})
	}
	keys := make([]string, len(data))
	i := 0
	for k := range data {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var vals []*ErrorGroupData
	for _, k := range keys {
		es := data[k]
		for _, e := range es {
			found := false
			for _, eg := range vals {
				if eg.StatusCode == e.Response.StatusCode {
					eg.Errors = append(eg.Errors, e)
					found = true
					break
				}
			}
			if !found {
				vals = append(vals,
					&ErrorGroupData{
						StatusCode: e.Response.StatusCode,
						Errors:     []*ErrorData{e},
					})
			}
		}
	}
	return vals
}

// buildRequestBodyType builds the TypeData for a request body. The data makes
// it possible to generate a function on the client side that creates the body
// from the service method payload.
//
// body is the HTTP request body
//
// att is the payload attribute
//
// e is the HTTP endpoint expression
//
// svr is true if the function is generated for server side code.
//
// sd is the service data
func (sds *ServicesData) buildRequestBodyType(body, att *expr.AttributeExpr, e *expr.HTTPEndpointExpr, svr bool, sd *ServiceData, sourceOwner, bodyOwner expr.ExampleIdentity) *TypeData {
	if body.Type == expr.Empty {
		return nil
	}
	body = expr.DupAtt(body)
	var (
		name        string
		varname     string
		desc        string
		def         string
		ref         string
		validateDef string
		validateRef string

		svc     = sd.Service
		catalog = sd.wireTypes(svr)
		policy  = wireTypePolicy{request: true, pointer: svr, useDefault: !svr, validate: true}
		httpctx = wireHTTPContext(catalog, catalog.scope, true, svr)
		side    = "client"
	)
	if svr {
		side = "server"
	}
	svcctx := sds.serviceTypeContext(sd, side).Enter(att)
	addMarshalTags(body)
	record := catalog.lookupUser(body, wireRequestBody, policy)
	catalog.applyNames(body, wireRequestBody, policy)
	name = body.Type.Name()
	if record != nil {
		name = record.name
		ref = record.ref
	} else {
		ref = httpctx.Scope.Ref(body, "")
	}

	if ut, ok := body.Type.(expr.UserType); ok {
		varname = record.name
		def = goTypeDefForContext(ut.Attribute(), httpctx)
		desc = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP request body.",
			varname, svc.Name, e.Name())
		if svr {
			// generate validation code for unmarshaled type (server-side).
			validateDef = codegen.ValidationCode(ut.Attribute(), ut, httpctx, true, expr.IsAlias(ut), false, "body")
			if validateDef != "" {
				validateRef = fmt.Sprintf("err = Validate%s(&body)", varname)
			}
		}
	} else {
		// Generate validation code first because inline struct validation is removed.
		ctx := wireHTTPContext(catalog, catalog.scope, true, svr)
		ctx.Pointer = !expr.IsPrimitive(body.Type)
		ctx.UseDefault = !svr
		validateRef = codegen.ValidationCode(body, nil, ctx, true, expr.IsAlias(body.Type), false, "body")
		if svr && expr.IsObject(body.Type) {
			// Body is an explicit object described in the design and in
			// this case the GoTypeRef is an inline struct definition. We
			// want to force all attributes to be pointers because we are
			// generating the server body type pre-validation.
			body.Validation = nil
		}
		varname = httpctx.Scope.Ref(body, "")
		desc = body.Description
	}
	var init *InitData
	if !svr && att.Type != expr.Empty && needInit(body.Type) {
		var (
			name    string
			desc    string
			code    string
			origin  string
			err     error
			helpers []*codegen.TransformFunctionData

			sourceVar = "p"
			svc       = sd.Service
		)
		{
			if record != nil {
				name = fmt.Sprintf("New%s", record.name)
			} else {
				name = fmt.Sprintf("New%s", codegen.Goify(httpctx.Scope.Name(body, "", httpctx.Pointer, httpctx.UseDefault), true))
			}
			desc = fmt.Sprintf("%s builds the HTTP request body from the payload of the %q endpoint of the %q service.",
				name, e.Name(), svc.Name)
			src := sourceVar
			srcAtt := att
			// If design uses Body("name") syntax then need to use payload attribute
			// to transform.
			if o, ok := body.Meta["origin:attribute"]; ok {
				srcObj := expr.AsObject(att.Type)
				origin = o[0]
				srcAtt = srcObj.Attribute(origin)
				src += "." + codegen.Goify(origin, true)
			}
			transformctx := wireHTTPContext(catalog, catalog.scope, true, svr)
			code, helpers, err = catalog.renderTransform(srcAtt, body, src, "body", "marshal", svcctx, transformctx)
			if err != nil {
				panic(err) // bug
			}
			sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
		}
		arg := InitArgData{
			Ref: sourceVar,
			AttributeData: &AttributeData{
				Name:     "payload",
				VarName:  sourceVar,
				TypeRef:  svcctx.Scope.Ref(att, svcctx.Pkg(att)),
				Type:     att.Type,
				Validate: validateDef,
				Example:  sds.Example(att, sourceOwner),
			},
		}
		init = &InitData{
			Name:                name,
			Description:         desc,
			ReturnTypeRef:       ref,
			ReturnTypeAttribute: codegen.Goify(origin, true),
			ClientCode:          code,
			ClientArgs:          []*InitArgData{&arg},
		}
	}
	data := &TypeData{
		Name:        name,
		VarName:     varname,
		Description: desc,
		Def:         def,
		Ref:         ref,
		Init:        init,
		ValidateDef: validateDef,
		ValidateRef: validateRef,
		Example:     sds.Example(body, bodyOwner),
	}
	if record == nil || data.Def == "" && data.ValidateDef == "" {
		return data
	}
	return catalog.bind(record, data)
}

// buildResponseBodyType builds the TypeData for a response body. The data
// makes it possible to generate a function that creates the server response
// body from the service method result/projected result or error.
//
// body is the response (success or error) HTTP body.
//
// att is the result/projected attribute.
//
// svr is true if the function is generated for server side code
//
// view is the view name to add as a suffix to the type name.
func (sds *ServicesData) buildResponseBodyType(body, att *expr.AttributeExpr, e *expr.HTTPEndpointExpr, svr bool, view *string, sd *ServiceData, sourceOwner, bodyOwner expr.ExampleIdentity) *TypeData {
	if body.Type == expr.Empty {
		return nil
	}
	body, viewName := prepareResponseWireBody(body, view)
	var (
		name        string
		varname     string
		desc        string
		def         string
		ref         string
		validateDef string
		validateRef string
		mustInit    bool

		svc  = sd.Service
		side = "client"
	)
	if svr {
		side = "server"
	}
	svcctx := sds.serviceTypeContext(sd, side).Enter(att)
	catalog := sd.wireTypes(svr)
	policy := wireTypePolicy{pointer: !svr, useDefault: svr, validate: !svr && view == nil, view: viewName}
	// Add each nested named field before body receives its chosen Go names. This
	// keeps each copied request or response field tied to its own definition.
	topLevel, _ := body.Type.(expr.UserType)
	collectUserTypes(body.Type, func(ut expr.UserType) {
		if topLevel != nil && ut == topLevel {
			return
		}
		if d := sds.attributeTypeData(ut, false, !svr, svr, sd); d != nil {
			if svr {
				sd.ServerBodyAttributeTypes = append(sd.ServerBodyAttributeTypes, d)
			} else {
				sd.ClientBodyAttributeTypes = append(sd.ClientBodyAttributeTypes, d)
			}
		}
	})
	record := catalog.lookupUser(body, wireResponseBody, policy)
	catalog.applyNames(body, wireResponseBody, policy)
	httpctx := wireHTTPContext(catalog, catalog.scope, false, svr)
	name = body.Type.Name()
	if record != nil {
		name = record.name
		ref = record.ref
	} else {
		ref = httpctx.Scope.Ref(body, "")
	}
	mustInit = att.Type != expr.Empty && needInit(body.Type)

	if ut, ok := body.Type.(expr.UserType); ok {
		// response body is a user type.
		varname = record.name
		def = goTypeDefForContext(ut.Attribute(), httpctx)
		desc = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP response body.",
			varname, svc.Name, e.Name())
		if !svr && view == nil {
			// generate validation code for unmarshaled type (client-side).
			validateDef = codegen.ValidationCode(body, ut, httpctx, true, expr.IsAlias(body.Type), false, "body")
			if validateDef != "" {
				target := "&body"
				if expr.IsArray(ut) {
					// result type collection
					target = "body"
				}
				validateRef = fmt.Sprintf("err = Validate%s(%s)", varname, target)
			}
		}
	} else if !expr.IsPrimitive(body.Type) && mustInit {
		// Response body is an array or map type.
		//
		// Server-side code needs a named wrapper (scoped to the endpoint) so the
		// generator can produce stable constructor identifiers (e.g.
		// New<Endpoint>ResponseBody) for element-wise transforms and projections.
		//
		// Client-side code decodes directly into the concrete composite type (e.g.
		// []T, map[K]V) and validates/transforms the value in-place. This avoids
		// generating endpoint-named alias types that are structurally identical and
		// may be deduplicated away in client/types.go.
		if svr {
			name = codegen.Goify(e.Name(), true) + "ResponseBody"
			record = catalog.lookup(body, wireResponseBody, policy, name)
			varname = record.name
			name = record.name
			desc = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP response body.",
				varname, svc.Name, e.Name())
			def = goTypeDefForContext(body, httpctx)
		} else {
			varname = httpctx.Scope.Ref(body, "")
			desc = body.Description
			def = ""
		}
		if !svr {
			validateRef = codegen.ValidationCode(body, nil, httpctx, true, expr.IsAlias(body.Type), false, "body")
		}
	} else {
		// response body is a primitive type. They are used as non-pointers when
		// encoding/decoding responses.
		httpctx = wireHTTPContext(catalog, catalog.scope, false, true)
		if !svr {
			validateRef = codegen.ValidationCode(body, nil, httpctx, true, expr.IsAlias(body.Type), false, "body")
		}
		varname = httpctx.Scope.Ref(body, "")
		desc = body.Description
	}
	var init *InitData
	if svr && mustInit {
		var (
			name    string
			desc    string
			rtref   string
			code    string
			origin  string
			err     error
			helpers []*codegen.TransformFunctionData

			sourceVar = "res"
			svc       = sd.Service
		)
		{
			var rtname string
			if _, ok := body.Type.(expr.UserType); !ok && !expr.IsPrimitive(body.Type) {
				rtname = codegen.Goify(e.Name(), true) + "ResponseBody"
				rtref = rtname
			} else {
				rtname = record.name
				rtref = ref
			}
			name = fmt.Sprintf("New%s", rtname)
			desc = fmt.Sprintf("%s builds the HTTP response body from the result of the %q endpoint of the %q service.",
				name, e.Name(), svc.Name)
			if view != nil {
				svcctx = sds.viewTypeContext(sd, "server").Enter(att)
			}
			src := sourceVar
			srcAtt := att
			// If design uses Body("name") syntax then need to use result attribute
			// to transform.
			if o, ok := body.Meta["origin:attribute"]; ok {
				srcObj := expr.AsObject(att.Type)
				origin = o[0]
				srcAtt = srcObj.Attribute(origin)
				src += "." + codegen.Goify(origin, true)
			}
			transformctx := wireHTTPContext(catalog, catalog.scope, false, svr)
			transformctx.Scope = catalog.resolver(catalog.scope, policy)
			code, helpers, err = catalog.renderTransform(srcAtt, body, src, "body", "marshal", svcctx, transformctx)
			if err != nil {
				panic(err) // bug
			}
			sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
		}
		ref := sourceVar
		if view != nil {
			ref += ".Projected"
		}
		tref := svcctx.Scope.Ref(att, svcctx.Pkg(att))
		arg := InitArgData{
			Ref: ref,
			AttributeData: &AttributeData{
				Name:     "result",
				VarName:  sourceVar,
				TypeRef:  tref,
				Type:     att.Type,
				Validate: validateDef,
				Example:  sds.Example(att, sourceOwner),
			},
		}
		init = &InitData{
			Name:                name,
			Description:         desc,
			ReturnTypeRef:       rtref,
			ReturnTypeAttribute: codegen.Goify(origin, true),
			ServerCode:          code,
			ServerArgs:          []*InitArgData{&arg},
		}
	}
	td := &TypeData{
		Name:        name,
		VarName:     varname,
		Description: desc,
		Def:         def,
		Ref:         ref,
		Init:        init,
		ValidateDef: validateDef,
		ValidateRef: validateRef,
		Example:     sds.Example(body, bodyOwner),
		View:        viewName,
	}
	if record == nil || td.Def == "" && td.ValidateDef == "" {
		return td
	}
	return catalog.bind(record, td)
}

func (sds *ServicesData) extractPathParams(a *expr.MappedAttributeExpr, service *expr.AttributeExpr, sd *ServiceData, owner expr.ExampleIdentity) []*ParamData {
	var params []*ParamData
	svcctx := sds.serviceTypeContext(sd, "server").Enter(service)
	sds.extractElements(pathElement, a, service, svcctx, sd.Scope, owner, func(el *Element, _ *expr.AttributeExpr) {
		params = append(params, &ParamData{
			Map:            false,
			MapStringSlice: false,
			Element:        el,
		})
	})
	return params
}

func (sds *ServicesData) extractQueryParams(a *expr.MappedAttributeExpr, service *expr.AttributeExpr, sd *ServiceData, owner expr.ExampleIdentity) []*ParamData {
	var params []*ParamData
	svcctx := sds.serviceTypeContext(sd, "server").Enter(service)
	sds.extractElements(queryElement, a, service, svcctx, sd.Scope, owner, func(el *Element, att *expr.AttributeExpr) {
		mp := expr.AsMap(att.Type)
		params = append(params, &ParamData{
			Map: mp != nil,
			MapStringSlice: mp != nil &&
				mp.KeyType.Type.Kind() == expr.StringKind &&
				mp.ElemType.Type.Kind() == expr.ArrayKind &&
				expr.AsArray(mp.ElemType.Type).ElemType.Type.Kind() == expr.StringKind,
			Element: el,
		})
	})
	return params
}

func (sds *ServicesData) extractHeaders(a *expr.MappedAttributeExpr, svcAtt *expr.AttributeExpr, svcCtx *codegen.AttributeContext, scope *codegen.NameScope, owner expr.ExampleIdentity) []*HeaderData {
	var headers []*HeaderData
	sds.extractElements(headerElement, a, svcAtt, svcCtx, scope, owner, func(el *Element, _ *expr.AttributeExpr) {
		headers = append(headers, &HeaderData{
			CanonicalName: http.CanonicalHeaderKey(el.HTTPName),
			Element:       el,
		})
	})
	return headers
}

func (sds *ServicesData) extractCookies(a *expr.MappedAttributeExpr, svcAtt *expr.AttributeExpr, svcCtx *codegen.AttributeContext, scope *codegen.NameScope, owner expr.ExampleIdentity) []*CookieData {
	var cookies []*CookieData
	sds.extractElements(cookieElement, a, svcAtt, svcCtx, scope, owner, func(el *Element, _ *expr.AttributeExpr) {
		c := &CookieData{Element: el}
		for n, v := range a.Meta {
			switch n {
			case "cookie:max-age":
				c.MaxAge = v[0]
			case "cookie:path":
				c.Path = v[0]
			case "cookie:domain":
				c.Domain = v[0]
			case "cookie:secure":
				c.Secure = v[0] == "Secure"
			case "cookie:http-only":
				c.HTTPOnly = v[0] == "HttpOnly"
			case "cookie:same-site":
				switch v[0] {
				case string(expr.CookieSameSiteLax):
					c.SameSite = "http.SameSiteLaxMode"
				case string(expr.CookieSameSiteStrict):
					c.SameSite = "http.SameSiteStrictMode"
				case string(expr.CookieSameSiteNone):
					c.SameSite = "http.SameSiteNoneMode"
				case string(expr.CookieSameSiteDefault):
					c.SameSite = "http.SameSiteDefaultMode"
				}
			}
		}
		cookies = append(cookies, c)
	})
	return cookies
}

// extractElements walks the mapped attribute and builds the Element data
// shared by path parameters, query string parameters, headers and cookies.
//
// a is the mapped attribute expression listing the HTTP elements.
//
// svcAtt is the service-level attribute (payload, result or error) the
// elements map to.
//
// svcCtx is the attribute context used to compute validation code and field
// pointer semantics for headers and cookies.
//
// add is called for each element with the built Element and the HTTP version
// of the element attribute so callers can derive kind-specific data.
//
// The element kinds differ as follows:
//
//   - path parameters are always required and never pointers,
//
//   - path and query string parameters are built from the mapped attribute
//     while headers and cookies are resolved against the service attribute,
//
//   - path and query string parameters may use text unmarshalers and rely on
//     the service expression to compute field pointer semantics,
//
//   - cookies do not track slice information (cookie values are scalars).
func (sds *ServicesData) extractElements(kind httpElementKind, a *expr.MappedAttributeExpr, svcAtt *expr.AttributeExpr, svcCtx *codegen.AttributeContext, scope *codegen.NameScope, owner expr.ExampleIdentity, add func(el *Element, att *expr.AttributeExpr)) {
	codegen.WalkMappedAttr(a, func(name, elem string, required bool, c *expr.AttributeExpr) error { // nolint: errcheck
		if kind == pathElement {
			required = true
		}
		attr := c
		if kind == headerElement || kind == cookieElement {
			attr = svcAtt.Find(name)
			if attr == nil {
				// Primitive payloads map the whole payload to a single element in
				// which case the mapped name has no corresponding attribute.
				if expr.IsObject(svcAtt.Type) {
					panic(fmt.Sprintf("%s %q does not map to a payload attribute", kind, name)) // bug
				}
				attr = svcAtt
			}
		}
		// The StringSlice field must be false for aliased primitive types
		var stringSlice bool
		if kind != cookieElement {
			if arr := expr.AsArray(attr.Type); arr != nil {
				stringSlice = arr.ElemType.Type.Kind() == expr.StringKind
			}
		}
		att := makeHTTPType(attr)
		var (
			varn        = scope.Name(codegen.Goify(name, false))
			typeRef     = scope.GoTypeRef(att)
			elemTypeRef string
			ft          = svcAtt.Type

			slice   bool
			pointer bool
			fptr    bool
		)
		if arr := expr.AsArray(att.Type); arr != nil {
			elemCtx := svcCtx.Enter(arr.ElemType)
			elemTypeRef = elemCtx.Scope.Ref(arr.ElemType, elemCtx.Pkg(arr.ElemType))
		}
		if kind != cookieElement {
			slice = expr.AsArray(att.Type) != nil
		}
		if kind != pathElement {
			pointer = a.IsPrimitivePointer(name, true)
		}
		if pointer {
			typeRef = "*" + typeRef
		}
		fieldName := codegen.GoifyAtt(att, name, true)
		if !expr.IsObject(svcAtt.Type) {
			fieldName = ""
		} else {
			ft = svcAtt.Find(name).Type
			if kind == pathElement || kind == queryElement {
				fptr = svcAtt.IsPrimitivePointer(name, true)
			} else {
				fptr = svcCtx.IsPrimitivePointer(name, svcAtt)
			}
		}
		validate := codegen.AttributeValidationCode(att, nil, svcCtx, required, expr.IsAlias(att.Type), varn, name)
		isText := (kind == pathElement || kind == queryElement) && isStringMetaType(att)
		if isText {
			// Build a copy of the attribute with Format cleared so the shared
			// validation code does not emit a format check (UnmarshalText covers it).
			attNoFmt := *att
			if att.Validation != nil {
				v := *att.Validation
				v.Format = ""
				attNoFmt.Validation = &v
			}
			validate = codegen.AttributeValidationCode(&attNoFmt, nil, svcCtx, required, expr.IsAlias(att.Type), varn+"Raw", name)
		}
		add(&Element{
			HTTPName:    elem,
			Slice:       slice,
			StringSlice: stringSlice,
			AttributeData: &AttributeData{
				Name:              name,
				Description:       att.Description,
				FieldName:         fieldName,
				FieldPointer:      fptr,
				FieldType:         ft,
				VarName:           varn,
				Required:          required,
				Type:              att.Type,
				TypeName:          scope.GoTypeName(att),
				TypeRef:           typeRef,
				ElemTypeRef:       elemTypeRef,
				Pointer:           pointer,
				Validate:          validate,
				IsTextUnmarshaler: isText,
				DefaultValue:      att.DefaultValue,
				Example:           sds.FieldExample(att, svcAtt, name, owner),
			},
		}, att)
		return nil
	})
}

// elementInitArg returns a payload constructor argument backed by a copy of
// the element attribute data. The text unmarshaler marker is dropped because
// it only drives the request decoding code, not constructors or CLI flags.
func elementInitArg(el *Element) *InitArgData {
	att := *el.AttributeData
	att.IsTextUnmarshaler = false
	return &InitArgData{Ref: att.VarName, AttributeData: &att}
}

// resultInitArg returns a result constructor argument backed by a copy of the
// element attribute data. Result constructor arguments carry no description,
// type name or default value: the constructor templates do not read them.
func resultInitArg(el *Element) *InitArgData {
	arg := elementInitArg(el)
	arg.Description = ""
	arg.TypeName = ""
	arg.DefaultValue = nil
	return arg
}

// errorInitArg returns an error constructor argument backed by a copy of the
// element attribute data. On top of the fields dropped by resultInitArg,
// error constructor arguments are never required, never pointers and never
// initialize pointer fields: the error constructor reads values directly.
func errorInitArg(el *Element) *InitArgData {
	arg := resultInitArg(el)
	arg.Required = false
	arg.Pointer = false
	arg.FieldPointer = false
	return arg
}

// collectUserTypes traverses the given data type recursively and calls back the
// given function for each attribute using a user type.
func collectUserTypes(dt expr.DataType, cb func(expr.UserType)) {
	collectUserTypesRecursive(dt, cb, make(map[expr.UserType]struct{}))
}

// collectUserTypesRecursive follows nested declarations once per authored
// origin so recursive copies terminate without hiding unrelated declarations.
func collectUserTypesRecursive(dt expr.DataType, cb func(expr.UserType), seen map[expr.UserType]struct{}) {
	if dt == expr.Empty {
		return
	}
	switch actual := dt.(type) {
	case *expr.Object:
		for _, nat := range *actual {
			collectUserTypesRecursive(nat.Attribute.Type, cb, seen)
		}
	case *expr.Union:
		for _, nat := range actual.Values {
			collectUserTypesRecursive(nat.Attribute.Type, cb, seen)
		}
	case *expr.Array:
		collectUserTypesRecursive(actual.ElemType.Type, cb, seen)
	case *expr.Map:
		collectUserTypesRecursive(actual.KeyType.Type, cb, seen)
		collectUserTypesRecursive(actual.ElemType.Type, cb, seen)
	case expr.UserType:
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		cb(actual)
		collectUserTypesRecursive(actual.Attribute().Type, cb, seen)
	}
}

// effectiveClientResponseBodyForView returns a copied response body containing
// the fields visible in one selected view. Type naming and client decoding both
// use this copy so they cannot disagree about its fields.
func effectiveClientResponseBodyForView(body *expr.AttributeExpr, view string) *expr.AttributeExpr {
	body = expr.DupAtt(body)
	rt, ok := body.Type.(*expr.ResultTypeExpr)
	if !ok {
		return body
	}
	projected, err := expr.Project(rt, view)
	if err != nil {
		panic(err) // bug
	}
	body.Type = projected
	return body
}

// clientResponseViewName returns the response view used by client code
// generation when the design fixes the response to a single view. An empty
// string means the client must keep the unprojected transport body because the
// server may render multiple views.
func clientResponseViewName(e *expr.HTTPEndpointExpr, md *service.MethodData) string {
	if md.ViewedResult == nil {
		return ""
	}
	if v, ok := e.MethodExpr.Result.Meta.Last(expr.ViewMetaKey); ok {
		return v
	}
	if len(md.ViewedResult.Views) == 1 {
		return md.ViewedResult.Views[0].Name
	}
	return ""
}

// clientResponseViewNameExpr returns the one view selected by the HTTP design.
// An empty result means each streamed response may name any allowed view.
func clientResponseViewNameExpr(e *expr.HTTPEndpointExpr, result *expr.ResultTypeExpr) string {
	if view, ok := e.MethodExpr.Result.Meta.Last(expr.ViewMetaKey); ok {
		return view
	}
	if len(result.Views) == 1 {
		return result.Views[0].Name
	}
	return ""
}

func buildHTTPUnionTypeData(u *expr.Union, scope codegen.Attributor, record *wireUnionRecord) *service.UnionTypeData {
	fields := make([]*service.UnionFieldData, len(u.Values))
	for i, nat := range u.Values {
		fieldName := codegen.Goify(nat.Name, true)
		fieldType := scope.Ref(nat.Attribute, scope.Package(nat.Attribute))
		fields[i] = &service.UnionFieldData{
			Name:        nat.Name,
			KindConst:   record.kindConsts[i],
			Constructor: record.constructors[i],
			FieldName:   fieldName,
			FieldType:   fieldType,
			Nilable:     codegen.IsNilable(nat.Attribute.Type),
			TypeTag:     nat.Name,
		}
	}

	return &service.UnionTypeData{
		Name:     record.name,
		KindName: record.kindName,
		Fields:   fields,
		TypeKey:  u.GetTypeKey(),
		ValueKey: u.GetValueKey(),
	}
}

func (sds *ServicesData) attributeTypeData(ut expr.UserType, req, ptr, server bool, rd *ServiceData) *TypeData {
	return sds.attributeTypeDataView(ut, req, ptr, server, "", rd)
}

// attributeTypeDataView builds a nested declaration using the view policy
// that selected its enclosing response shape.
func (sds *ServicesData) attributeTypeDataView(ut expr.UserType, req, ptr, server bool, view string, rd *ServiceData) *TypeData {
	if ut == expr.Empty {
		return nil
	}

	var (
		name        string
		desc        string
		validate    string
		validateRef string

		att     = expr.DupAtt(&expr.AttributeExpr{Type: ut})
		catalog = rd.wireTypes(server)
		policy  = wireTypePolicy{request: req, pointer: ptr, useDefault: hctxUseDefault(req, server), validate: req || !server, view: view}
	)
	ut = att.Type.(expr.UserType)
	record := catalog.lookupUser(att, wireAttribute, policy)
	catalog.applyNames(att, wireAttribute, policy)
	hctx := wireHTTPContext(catalog, catalog.scope, req, server)
	name = record.name
	ctx := "request"
	if !req {
		ctx = "response"
	}
	desc = name + " is used to define fields on " + ctx + " body types."
	if (req || !req && !server) && !expr.IsAlias(ut) {
		// Generate validations for responses client-side and for
		// requests server-side and CLI.
		// Alias types are validated inline in the parent type
		validate = codegen.ValidationCode(ut.Attribute(), ut, hctx, true, expr.IsAlias(ut), false, "body")
	}
	if validate != "" {
		validateRef = fmt.Sprintf("err = Validate%s(v)", name)
	}
	return catalog.bind(record, &TypeData{
		Name:        ut.Name(),
		VarName:     name,
		Description: desc,
		Def:         goTypeDefForContext(ut.Attribute(), hctx),
		Ref:         record.ref,
		ValidateDef: validate,
		ValidateRef: validateRef,
		Example:     sds.Example(att, expr.UserTypeExampleIdentity(ut)),
	})
}

// wireTypes returns the request and response types for the server or client package.
func (sd *ServiceData) wireTypes(server bool) *wireTypeCatalog {
	if server {
		return sd.serverWireTypes
	}
	return sd.clientWireTypes
}

// hctxUseDefault reports whether missing HTTP values receive their design
// defaults for the selected request or response side.
func hctxUseDefault(request, server bool) bool {
	return !request && server || request && !server
}

// httpContext returns a context for attributes of types used to marshal and
// unmarshal HTTP requests and responses.
//
// pkg is the package name where the body type exists
//
// scope is the named scope
//
// request if true indicates that the type is a request type, else response
// type
//
// svr if true indicates that the type is a server type, else client type
func httpContext(scope *codegen.NameScope, request, svr bool) *codegen.AttributeContext {
	marshal := !request && svr || request && !svr
	ctx := codegen.NewAttributeContext(!marshal, false, marshal, "", scope)
	ctx.UnionPointer = true
	return ctx
}

// wireHTTPContext returns the pointer and default-value rules for one generated
// HTTP package. It maps each copied field to the Go type name chosen for
// that particular request or response.
func wireHTTPContext(catalog *wireTypeCatalog, scope *codegen.NameScope, request, server bool) *codegen.AttributeContext {
	context := httpContext(scope, request, server)
	context.Scope = catalog.resolver(scope, wireTypePolicy{
		request:    request,
		pointer:    context.Pointer,
		useDefault: context.UseDefault,
	})
	return context
}

// serviceTypeContext returns the service type names as referenced from the
// generated client or server package named by side.
func (sds *ServicesData) serviceTypeContext(sd *ServiceData, side string) *codegen.AttributeContext {
	outputPackage := path.Join(sds.GenPkg(), sds.dir(), sd.Service.PathName, side)
	return &codegen.AttributeContext{
		UseDefault: true,
		Scope:      sds.ServiceAttributor(sd.Service.Name, outputPackage),
	}
}

// viewTypeContext returns the result-view type names as referenced from the
// generated client or server package named by side.
func (sds *ServicesData) viewTypeContext(sd *ServiceData, side string) *codegen.AttributeContext {
	outputPackage := path.Join(sds.GenPkg(), sds.dir(), sd.Service.PathName, side)
	return &codegen.AttributeContext{
		Pointer:    true,
		UseDefault: true,
		Scope:      sds.ViewAttributor(sd.Service.Name, outputPackage),
	}
}

// unmarshal initializes a data structure defined by target type from a data
// structure defined by source type. The attributes in the source data
// structure are pointers and the attributes in the target data structure that
// have default values are non-pointers. Fields in target type are initialized
// with their default values (if any).
//
// source, target are the attributes used in the transformation
//
// sourceVar, targetVar are the variable names for source and target used in
// the transformation code
//
// sourceCtx, targetCtx are the source and target attribute contexts
func unmarshal(source, target *expr.AttributeExpr, sourceVar string, sourceCtx, targetCtx *codegen.AttributeContext) (string, []*codegen.TransformFunctionData, error) {
	return codegen.GoTransform(source, target, sourceVar, "v", sourceCtx, targetCtx, "unmarshal", true)
}

// marshal initializes a data structure defined by target type from a data
// structure defined by source type. The fields in the source and target
// data structure use non-pointers for attributes with default values.
//
// source, target are the attributes used in the transformation
//
// sourceVar, targetVar are the variable names for source and target used in
// the transformation code
//
// sourceCtx, targetCtx are the source and target attribute contexts
func marshal(source, target *expr.AttributeExpr, sourceVar, targetVar string, sourceCtx, targetCtx *codegen.AttributeContext) (string, []*codegen.TransformFunctionData, error) {
	return codegen.GoTransform(source, target, sourceVar, targetVar, sourceCtx, targetCtx, "marshal", true)
}

// needConversion returns true if the type needs to be converted from a string.
func needConversion(dt expr.DataType) bool {
	if dt == expr.Empty {
		return false
	}
	switch actual := dt.(type) {
	case expr.Primitive:
		if actual.Kind() == expr.StringKind ||
			actual.Kind() == expr.AnyKind ||
			actual.Kind() == expr.BytesKind {
			return false
		}
		return true
	case *expr.Array:
		return needConversion(actual.ElemType.Type)
	case *expr.Map:
		return needConversion(actual.KeyType.Type) ||
			needConversion(actual.ElemType.Type)
	default:
		return true
	}
}

// isStringMetaType returns true if the attribute has a struct:field:type meta
// whose underlying DSL type is string, indicating the custom type should
// implement encoding.TextUnmarshaler for HTTP parameter conversion.
func isStringMetaType(c *expr.AttributeExpr) bool {
	typeName, _ := codegen.GetMetaType(c)
	if typeName == "" {
		return false
	}
	return c.Type.Kind() == expr.StringKind
}

// addMarshalTags adds JSON, XML and Form tags to all inline object attributes recursively.
func addMarshalTags(att *expr.AttributeExpr) {
	addMarshalTagsRecursive(att, make(map[expr.UserType]struct{}))
}

// addMarshalTagsRecursive annotates every inline object reachable through one
// declaration origin and stops when recursive copies return to that origin.
func addMarshalTagsRecursive(att *expr.AttributeExpr, seen map[expr.UserType]struct{}) {
	if ut, ok := att.Type.(expr.UserType); ok {
		origin := ut.Origin()
		if _, ok := seen[origin]; ok {
			return // avoid infinite recursions
		}
		seen[origin] = struct{}{}
		if expr.IsObject(ut.Attribute().Type) {
			for _, att := range *(expr.AsObject(att.Type)) {
				addMarshalTagsRecursive(att.Attribute, seen)
			}
		}
		return
	}
	if expr.IsArray(att.Type) {
		addMarshalTagsRecursive(expr.AsArray(att.Type).ElemType, seen)
		return
	}
	if expr.IsMap(att.Type) {
		addMarshalTagsRecursive(expr.AsMap(att.Type).KeyType, seen)
		addMarshalTagsRecursive(expr.AsMap(att.Type).ElemType, seen)
		return
	}
	if !expr.IsObject(att.Type) {
		return
	}
	// inline object
	for _, natt := range *(expr.AsObject(att.Type)) {
		if natt.Attribute.Meta == nil {
			natt.Attribute.Meta = expr.MetaExpr{}
		}
		ns := []string{natt.Name}
		natt.Attribute.Meta["struct:tag:form"] = ns
		natt.Attribute.Meta["struct:tag:json"] = ns
		natt.Attribute.Meta["struct:tag:xml"] = ns
	}
}

// needInit returns true if and only if the given type is or makes use of user
// types.
func needInit(dt expr.DataType) bool {
	if dt == expr.Empty {
		return false
	}
	switch actual := dt.(type) {
	case expr.Primitive:
		return false
	case *expr.Array:
		return needInit(actual.ElemType.Type)
	case *expr.Map:
		return needInit(actual.KeyType.Type) ||
			needInit(actual.ElemType.Type)
	case *expr.Object:
		for _, nat := range *actual {
			if needInit(nat.Attribute.Type) {
				return true
			}
		}
		return false
	case expr.UserType:
		return true
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// upgradeParams returns the data required to render the websocket_upgrade
// template.
func upgradeParams(e *EndpointData, fn string) map[string]any {
	return map[string]any{
		"ViewedResult": e.Method.ViewedResult,
		"Function":     fn,
	}
}

// serviceHasViewedResult reports whether the selected endpoint sections
// reference a projected result from the service views package.
func serviceHasViewedResult(service *ServiceData, selected func(*EndpointData) bool) bool {
	for _, endpoint := range service.Endpoints {
		if selected != nil && !selected(endpoint) {
			continue
		}
		if endpoint.Method.ViewedResult != nil {
			return true
		}
	}
	return false
}

// NeedDialer returns true if at least one method in the defined services
// uses WebSocket for sending payload or result.
func NeedDialer(data []*ServiceData) bool {
	return slices.ContainsFunc(data, HasWebSocket)
}
