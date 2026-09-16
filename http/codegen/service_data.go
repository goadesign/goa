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
		cliParsers map[string]*cli.ParserPlan
		// linkErr is the first conversion error found while building template data.
		// Plan.Link returns it before exposing any generated files.
		linkErr error
		// fileImports contains the exact design-derived imports collected for each
		// generated file before package names were frozen.
		fileImports map[string][]*codegen.ImportSpec
		// clientServicePackages contains the package name used by client files for
		// each generated service. Services with no client reference keep the
		// preferred name because no import alias was selected.
		clientServicePackages map[string]string
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
		// ServerHandlerWrappers lists the planned wrapper declarations copied into
		// every endpoint and file mount helper for this service.
		ServerHandlerWrappers []*codegen.NameDeclaration
		// ServerMounts lists functions that add routes after the routes defined in
		// the design.
		ServerMounts []*ServerMount
		// ServerStruct is the server type name kept for existing plugins.
		//
		// Deprecated: Use ServerStructDeclaration.Name() after planning so name collisions are handled.
		ServerStruct string
		// ServerStructDeclaration is the generated Go type name used by server definitions and calls.
		ServerStructDeclaration *codegen.NameDeclaration
		// MountPointStruct is the mount point type name kept for existing plugins.
		//
		// Deprecated: Use MountPointStructDeclaration.Name() after planning so name collisions are handled.
		MountPointStruct string
		// MountPointStructDeclaration is the generated Go type name used by the mount point type.
		MountPointStructDeclaration *codegen.NameDeclaration
		// ServerInit is the server constructor name kept for existing plugins.
		//
		// Deprecated: Use ServerInitDeclaration.Name() after planning so name collisions are handled.
		ServerInit string
		// ServerInitDeclaration is the generated Go function name used by the server constructor.
		ServerInitDeclaration *codegen.NameDeclaration
		// MountServer is the route mount function name kept for existing plugins.
		//
		// Deprecated: Use MountServerDeclaration.Name() after planning so name collisions are handled.
		MountServer string
		// MountServerDeclaration is the generated Go function name used to mount the service routes.
		MountServerDeclaration *codegen.NameDeclaration
		// ServerService is the name of service function.
		ServerService string
		// ClientStruct is the client type name kept for existing plugins.
		//
		// Deprecated: Use ClientStructDeclaration.Name() after planning so name collisions are handled.
		ClientStruct string
		// ClientStructDeclaration is the generated Go type name used by the client.
		ClientStructDeclaration *codegen.NameDeclaration
		// ClientInitDeclaration is the generated Go function name used by the client constructor.
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
		// Scope records unique Go names for all server and client types.
		Scope *codegen.NameScope
		// serverWireTypes stores declarations emitted in the server package.
		serverWireTypes *wireTypeCatalog
		// clientWireTypes stores declarations emitted in the client package.
		clientWireTypes *wireTypeCatalog
		// clientBodyConstructors contains planned constructors for unnamed
		// request and streamed request body values.
		clientBodyConstructors map[clientBodyConstructorKey]*codegen.NameDeclaration
		// transforms contains the exact conversions selected while this
		// service's HTTP shapes were planned.
		transforms plannedWireTransforms
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
		// IsJSONRPCNotification reports whether every call omits the request
		// ID and receives no JSON-RPC response.
		IsJSONRPCNotification bool
		// JSONRPCRequestID contains the complete request-ID plan. It is nil for
		// explicit notifications and non-JSON-RPC endpoints.
		JSONRPCRequestID *JSONRPCRequestIDData
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

		// MountHandler is the endpoint mount function name kept for existing plugins.
		//
		// Deprecated: Use MountHandlerDeclaration.Name() after planning so name collisions are handled.
		MountHandler string
		// MountHandlerDeclaration is the generated Go function name used to mount this endpoint.
		MountHandlerDeclaration *codegen.NameDeclaration
		// ServerHandlerWrappers lists functions that surround the handler before
		// this endpoint's mount function registers its routes.
		ServerHandlerWrappers []*codegen.NameDeclaration
		// HandlerInit is the handler constructor name kept for existing plugins.
		//
		// Deprecated: Use HandlerInitDeclaration.Name() after planning so name collisions are handled.
		HandlerInit string
		// HandlerInitDeclaration is the generated Go function name used to create
		// this endpoint's handler.
		HandlerInitDeclaration *codegen.NameDeclaration
		// RequestDecoder is the request decoder name kept for existing plugins.
		//
		// Deprecated: Use RequestDecoderDeclaration.Name() after planning so name collisions are handled.
		RequestDecoder string
		// RequestDecoderDeclaration is the generated Go function name used to
		// decode this endpoint's request.
		RequestDecoderDeclaration *codegen.NameDeclaration
		// ResponseEncoder is the response encoder name kept for existing plugins.
		//
		// Deprecated: Use ResponseEncoderDeclaration.Name() after planning so name collisions are handled.
		ResponseEncoder string
		// ResponseEncoderDeclaration is the generated Go function name used to
		// encode this endpoint's response.
		ResponseEncoderDeclaration *codegen.NameDeclaration
		// ErrorEncoder is the error encoder name kept for existing plugins.
		//
		// Deprecated: Use ErrorEncoderDeclaration.Name() after planning so name collisions are handled.
		ErrorEncoder string
		// ErrorEncoderDeclaration is the generated Go function name used to encode
		// this endpoint's errors.
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
		// HasMixedResults indicates that HTTP clients may request one normal result
		// or a stream of results.
		HasMixedResults bool

		// client

		// ClientStruct is the client type name kept for existing plugins.
		//
		// Deprecated: Use ClientStructDeclaration.Name() after planning so name collisions are handled.
		ClientStruct string
		// ClientStructDeclaration supplies the client type name used by endpoint methods.
		ClientStructDeclaration *codegen.NameDeclaration
		// EndpointInit is the name of the constructor function for the
		// client endpoint.
		EndpointInit string
		// RequestInit is the request builder function.
		RequestInit *InitData
		// RequestEncoder is the request encoder name kept for existing plugins.
		//
		// Deprecated: Use RequestEncoderDeclaration.Name() after planning so name collisions are handled.
		RequestEncoder string
		// RequestEncoderDeclaration is the generated Go function name used to
		// encode this endpoint's request.
		RequestEncoderDeclaration *codegen.NameDeclaration
		// ResponseDecoder is the response decoder name kept for existing plugins.
		//
		// Deprecated: Use ResponseDecoderDeclaration.Name() after planning so name collisions are handled.
		ResponseDecoder string
		// ResponseDecoderDeclaration is the generated Go function name used to
		// decode this endpoint's response.
		ResponseDecoderDeclaration *codegen.NameDeclaration
		// MultipartRequestEncoder indicates the request encoder for
		// multipart content type.
		MultipartRequestEncoder *MultipartData
		// ClientWebSocket holds the data to render the client struct which
		// implements the client stream interface.
		ClientWebSocket *WebSocketData
		// BuildStreamPayload is the streamed request helper name kept for existing plugins.
		//
		// Deprecated: Use BuildStreamPayloadDeclaration.Name() after planning so name collisions are handled.
		BuildStreamPayload string
		// BuildStreamPayloadDeclaration is the generated Go function name used to
		// build streamed requests.
		BuildStreamPayloadDeclaration *codegen.NameDeclaration
		// CLIPayloadDeclaration is the generated Go function name used to build command-line payloads.
		CLIPayloadDeclaration *codegen.NameDeclaration
	}

	// FileServerData lists the data needed to generate file servers.
	FileServerData struct {
		// MountHandler is the file server mount function name kept for existing plugins.
		//
		// Deprecated: Use MountHandlerDeclaration.Name() after planning so name collisions are handled.
		MountHandler string
		// MountHandlerDeclaration is the generated Go function name used to mount this file server.
		MountHandlerDeclaration *codegen.NameDeclaration
		// ServerHandlerWrappers lists functions that surround the handler before
		// this file server's mount function registers its routes.
		ServerHandlerWrappers []*codegen.NameDeclaration
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
		// CLIPlan describes how command-line text becomes the complete payload.
		CLIPlan *cli.FlagPlan
		// Request contains the data for the corresponding HTTP request.
		Request *RequestData
		// DecoderReturnValue is a reference to the decoder return value
		// if there is no payload constructor (i.e. if Init is nil).
		DecoderReturnValue string
		// IDAttribute is the name of the attribute where the ID of the
		// JSON-RPC request is stored.
		IDAttribute string
		// IDAttributeTypeRef is the named string type assigned from the
		// JSON-RPC request ID. It is empty for a plain string field.
		IDAttributeTypeRef string
		// IDAttributeRequired is true if the ID attribute is required.
		IDAttributeRequired bool
		// JSONRPCRequestID contains the typed payload field and generated-client
		// behavior for the JSON-RPC request ID.
		JSONRPCRequestID *JSONRPCRequestIDData
	}

	// JSONRPCRequestIDData contains the generation-time choices for one
	// JSON-RPC request ID.
	JSONRPCRequestIDData struct {
		// Attribute is the service payload field that receives the ID. It is
		// empty when the generated client creates every ID.
		Attribute string
		// Variable is the local constructor argument that carries the ID.
		Variable string
		// ValueTypeRef is the generated Go type for the ID value.
		ValueTypeRef string
		// Aliased is true when ValueTypeRef is a named string type.
		Aliased bool
		// Required is true when the design marks the payload field as required.
		Required bool
		// MustHave is true when the JSON-RPC request must contain an ID because
		// the service field has no default.
		MustHave bool
		// Pointer is true when the service payload stores the ID as a pointer.
		Pointer bool
		// HasDefault is true when an absent request ID uses the design default.
		HasDefault bool
		// DefaultValue is the string used when HasDefault is true.
		DefaultValue string
		// Generate is true when the client creates an ID if the payload does not
		// supply one.
		Generate bool
		// Validate contains the service-field checks emitted after decoding the
		// envelope ID.
		Validate string
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
		// IDAttribute is retained for generator compatibility.
		//
		// Deprecated: JSON-RPC results cannot define an ID field.
		IDAttribute string
		// IDAttributeRequired is retained for generator compatibility.
		//
		// Deprecated: JSON-RPC results cannot define an ID field.
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
		// ServerBody describes the request body decoded by server code.
		// Scalar fields use pointers when validation must distinguish an
		// omitted value from zero. Byte slices and Any fields remain values.
		ServerBody *TypeData
		// ClientBody describes the request body encoded by client code. Its
		// fields follow the client transport shape because no body decoding
		// validation runs there.
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
		// OptionalBody reports whether Body selects an optional payload field.
		OptionalBody bool
		// BodyIsUnion reports whether the selected optional field is a union whose
		// empty kind represents absence.
		BodyIsUnion bool
		// BodyFieldPointer reports whether the selected service payload field is
		// a pointer to a primitive value.
		BodyFieldPointer bool
		// BodyFieldCanBeAbsent reports whether the selected service payload field
		// can represent absence.
		BodyFieldCanBeAbsent bool
		// MustValidate is true if the request body or at least one
		// parameter or header requires validation.
		MustValidate bool
		// Multipart if true indicates the request is a multipart
		// request.
		Multipart bool
		// JSONRPCParams describes how the request body is carried in JSON-RPC
		// params. It is nil for ordinary HTTP requests and requests without a
		// body.
		JSONRPCParams *JSONRPCParamsData
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
		// ServerBody is the response body encoded by server code, or nil when
		// the body is empty. Its fields follow the server transport shape.
		// If the method result is a result
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
		// ClientBody is the response body decoded by client code, or nil when
		// the body is empty. Scalar fields use pointers when validation must
		// distinguish an omitted value from zero. Byte slices and Any fields
		// remain values. Service result fields keep their own pointer rules.
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
		// SelectClientBodyByView is true when a unary HTTP client must read the
		// goa-view response header before choosing the body type and constructor.
		SelectClientBodyByView bool
	}

	// ViewedRepresentationData describes the HTTP body used for one legal result
	// view. The server constructor converts the service result into ServerBody.
	// The client decodes ClientBody and ResultInit rebuilds the service result.
	ViewedRepresentationData struct {
		// View is the exact design view name carried on variable-view messages.
		View string
		// ResultAttr is the Go field selected by Body("name"). It is empty when
		// the response body uses the complete result containing the selected view's
		// fields.
		ResultAttr string
		// ServerBody is the body type encoded by the server for View.
		ServerBody *TypeData
		// ClientBody is the body type decoded by the client for View.
		ClientBody *TypeData
		// ClientDataPointer reports whether the SSE data line is assigned to
		// a pointer field in ClientBody.
		ClientDataPointer bool
		// ResultInit rebuilds the result containing the selected view's fields from
		// ClientBody.
		ResultInit *InitData
	}

	// InitData contains the data required to render a constructor.
	InitData struct {
		// Declaration is the generated Go function name used by this constructor.
		Declaration *codegen.NameDeclaration
		// ClientDeclaration is the generated Go function name written in the client package
		// when the same path constructor is also written in the server package.
		ClientDeclaration *codegen.NameDeclaration
		// Name is the constructor function name kept for existing plugins.
		//
		// Deprecated: Use Declaration.Name() after planning so name collisions are handled.
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
		// ReturnIsUnion reports whether the selected body field is a union value.
		ReturnIsUnion bool
		// OptionalBody reports whether body is the value of an optional payload
		// field and may be nil.
		OptionalBody bool
		// ClientBodyOptional reports whether the service field supplied by a
		// client can represent absence.
		ClientBodyOptional bool
		// BodyDefault is the specialized Go value assigned when an optional
		// request body is absent. It is nil when the field has no authored default.
		BodyDefault *codegen.GoValueCode
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
		// ValueTypeRef is TypeRef without the pointer used for optional fields.
		ValueTypeRef string
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
		// HasDefault reports whether DefaultValue was authored, including a zero value.
		HasDefault bool
		// Validate contains the validation code for the attribute value if any.
		Validate string
		// CLIPlan describes how command-line text becomes this attribute value and
		// how the generated payload builder validates it.
		CLIPlan *cli.FlagPlan
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
		// PreserveEmpty reports whether an empty header value is different from
		// an absent header.
		PreserveEmpty bool
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
		// Declaration is the generated Go type name.
		Declaration *codegen.NameDeclaration
		// VarName is the Go type spelling kept for existing plugins. When
		// Declaration is nonnil, it matches Declaration.Name(). Otherwise it is a
		// Go expression such as []string that does not declare a named type.
		//
		// Deprecated: Use Declaration.Name() after planning so name collisions are handled.
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
		// NestedValidateDef contains validation code whose error paths begin with
		// the path passed by another generated validator.
		NestedValidateDef string
		// ValidateRef contains inline validation code when no named validator is called.
		ValidateRef string
		// ValidationTarget is the value passed to ValidatorDeclaration. It is empty
		// when this body does not need a named validator call.
		ValidationTarget string
		// ValidatorDeclaration is the generated Go function name that runs ValidateDef.
		ValidatorDeclaration *codegen.NameDeclaration
		// ValidatorName is the validator name kept for existing plugins.
		//
		// Deprecated: Use ValidatorDeclaration.Name() after planning so name collisions are handled.
		ValidatorName string
		// NestedValidatorDeclaration is the private generated Go function name used
		// when this type appears inside another HTTP body value.
		NestedValidatorDeclaration *codegen.NameDeclaration
		// NestedValidatorName is the nested validator name kept for existing plugins.
		//
		// Deprecated: Use NestedValidatorDeclaration.Name() after planning so name collisions are handled.
		NestedValidatorName string
		// Example is an example value for the type.
		Example any
		// View is the view used to render the (result) type if any.
		View string
		// declaration points to the one request or response type record that the HTTP
		// plan uses for this generated type.
		declaration *wireTypeRecord
		// attribute is the copied HTTP type whose generated names produced Def and
		// Ref. Example generators use it to qualify nested request body types.
		attribute *expr.AttributeExpr
	}

	// MultipartData contains the data needed to render multipart
	// encoder/decoder.
	MultipartData struct {
		// FuncName is the multipart function type or helper name kept for existing plugins.
		//
		// Deprecated: Use FuncDeclaration.Name() after planning so name collisions are handled.
		FuncName string
		// FuncDeclaration is the generated Go name used by the multipart function type or helper.
		FuncDeclaration *codegen.NameDeclaration
		// InitName is the multipart constructor name kept for existing plugins.
		//
		// Deprecated: Use InitDeclaration.Name() after planning so name collisions are handled.
		InitName string
		// InitDeclaration is the generated Go function name used by the multipart constructor.
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

	// shapedBodies stores each HTTP body after it is copied from the design.
	// Requests, responses, and streamed results use their HTTP field names;
	// streamed requests keep their authored fields. Reusing each copy gives
	// every generated file the same body and the same example values.
	shapedBodies struct {
		requests      map[*expr.HTTPEndpointExpr]*expr.AttributeExpr
		streams       map[*expr.HTTPEndpointExpr]*expr.AttributeExpr
		streamResults map[*expr.HTTPEndpointExpr]*expr.AttributeExpr
		responses     map[*expr.HTTPResponseExpr]*expr.AttributeExpr
		errors        map[*expr.HTTPErrorExpr]*expr.AttributeExpr
	}

	// releasedWireTypePair stops recursive response types after their current
	// and released names have been paired once.
	releasedWireTypePair struct {
		current  expr.UserType
		released expr.UserType
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
	transportService.PkgName = sds.clientServicePackages[svc.Name]
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
	clientPkgName := strings.ToLower(codegen.Goify(svc.Name, false)) + "c"
	serverPkgName := strings.ToLower(codegen.Goify(svc.Name, false)) + "svr"
	clientOutputPackage := path.Join(sds.GenPkg(), sds.dir(), svc.PathName, "client")
	if sds.jsonrpc {
		serverPkgName = strings.ToLower(codegen.Goify(svc.Name, false)) + "jssvr"
	}
	if len(httpSvc.HTTPEndpoints) > 0 {
		for _, server := range sds.Root.API.Servers {
			if !slices.Contains(server.Services, svc.Name) || sds.cliParsers[server.Name] == nil {
				continue
			}
			serverName := codegen.SnakeCase(codegen.Goify(server.Name, true))
			cliOutputPackage := path.Join(sds.GenPkg(), sds.dir(), "cli", serverName)
			clientPkgName = sds.PackageImport(cliOutputPackage, clientOutputPackage).Name
			break
		}
	}
	sd := &ServiceData{
		Service:                             svc,
		ClientPkgName:                       clientPkgName,
		ServerPkgName:                       serverPkgName,
		ServerStruct:                        symbols.serverStruct.Name(),
		ServerStructDeclaration:             symbols.serverStruct,
		MountPointStruct:                    symbols.mountPoint.Name(),
		MountPointStructDeclaration:         symbols.mountPoint,
		ServerInit:                          symbols.serverInit.Name(),
		ServerInitDeclaration:               symbols.serverInit,
		MountServer:                         symbols.mountServer.Name(),
		MountServerDeclaration:              symbols.mountServer,
		ServerService:                       "Service",
		ClientStruct:                        symbols.clientStruct.Name(),
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
		clientBodyConstructors:              planned.clientBodyConstructors,
		transforms:                          planned.transforms,
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
			MountHandler:            symbols.fileServers[s].Name(),
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
			args       []*InitArgData
			payloadRef string
			pkg        string
		)
		{
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
						attribute := &expr.AttributeExpr{Type: ca.FieldType}
						ca.ServiceTypeRef = svcctx.Scope.Ref(attribute, svcctx.Pkg(attribute))
					}
					args = append(args, ca)
				}
			}
			pkg = svc.PkgName
			if len(routes[0].PathInit.ClientArgs) > 0 && httpEndpoint.MethodExpr.Payload.Type != expr.Empty {
				payloadPkg := ""
				if attributeUsesServiceType(httpEndpoint.MethodExpr.Payload, make(map[expr.UserType]struct{})) {
					payloadPkg = svcctx.Pkg(httpEndpoint.MethodExpr.Payload)
				}
				payloadRef = svcctx.Scope.Ref(httpEndpoint.MethodExpr.Payload, payloadPkg)
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
			Declaration: endpointSymbols.requestBuilder,
			Name:        endpointSymbols.requestBuilder.Name(),
			Description: fmt.Sprintf("%s instantiates a HTTP request object with method and path set to call the %q service %q endpoint", endpointSymbols.requestBuilder.Name(), svc.Name, method.Name),
			ClientCode:  buf.String(),
			ClientArgs:  clientArgs,
		}

		ed := &EndpointData{
			Method:                     method,
			IsJSONRPC:                  httpEndpoint.IsJSONRPC(),
			IsJSONRPCNotification:      httpEndpoint.IsJSONRPCNotification(),
			JSONRPCRequestID:           payload.JSONRPCRequestID,
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
			MountHandler:               endpointSymbols.mountHandler.Name(),
			MountHandlerDeclaration:    endpointSymbols.mountHandler,
			HandlerInit:                endpointSymbols.handlerInit.Name(),
			HandlerInitDeclaration:     endpointSymbols.handlerInit,
			RequestDecoderDeclaration:  endpointSymbols.requestDecoder,
			ResponseEncoderDeclaration: endpointSymbols.responseEncoder,
			ErrorEncoderDeclaration:    endpointSymbols.errorEncoder,
			DiscardStreamDeclaration:   endpointSymbols.discardStream,
			ClientStruct:               symbols.clientStruct.Name(),
			ClientStructDeclaration:    symbols.clientStruct,
			EndpointInit:               method.VarName,
			RequestInit:                requestInit,
			HasMixedResults:            httpEndpoint.MethodExpr.HasMixedResults(),
			RequestEncoderDeclaration:  endpointSymbols.requestEncoder,
			ResponseDecoder:            endpointSymbols.responseDecoder.Name(),
			ResponseDecoderDeclaration: endpointSymbols.responseDecoder,
			Requirements:               reqs,
		}
		if declaration := endpointSymbols.requestDecoder; declaration != nil {
			ed.RequestDecoder = declaration.Name()
		}
		if declaration := endpointSymbols.responseEncoder; declaration != nil {
			ed.ResponseEncoder = declaration.Name()
		}
		if declaration := endpointSymbols.errorEncoder; declaration != nil {
			ed.ErrorEncoder = declaration.Name()
		}
		if declaration := endpointSymbols.requestEncoder; declaration != nil {
			ed.RequestEncoder = declaration.Name()
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
				ed.SSE.StructName = endpointSymbols.serverStream.Name()
				ed.SSE.StructDeclaration = endpointSymbols.serverStream
				ed.SSE.ClientInterfaceDeclaration = endpointSymbols.sseClientInterface
				ed.SSE.ClientStructDeclaration = endpointSymbols.sseClientStruct
				ed.SSE.ClientInitDeclaration = endpointSymbols.sseClientInit
			}
		}

		if httpEndpoint.MultipartRequest {
			ed.MultipartRequestDecoder = &MultipartData{
				FuncName:        endpointSymbols.serverMultipart.functionType.Name(),
				FuncDeclaration: endpointSymbols.serverMultipart.functionType,
				InitName:        endpointSymbols.serverMultipart.constructor.Name(),
				InitDeclaration: endpointSymbols.serverMultipart.constructor,
				VarName:         fmt.Sprintf("%s%sDecoderFn", svc.VarName, method.VarName),
				ServiceName:     svc.Name,
				MethodName:      method.Name,
				Payload:         ed.Payload,
			}
			ed.MultipartRequestEncoder = &MultipartData{
				FuncName:        endpointSymbols.clientMultipart.functionType.Name(),
				FuncDeclaration: endpointSymbols.clientMultipart.functionType,
				InitName:        endpointSymbols.clientMultipart.constructor.Name(),
				InitDeclaration: endpointSymbols.clientMultipart.constructor,
				VarName:         fmt.Sprintf("%s%sEncoderFn", svc.VarName, method.VarName),
				ServiceName:     svc.Name,
				MethodName:      method.Name,
				Payload:         ed.Payload,
			}
		}

		if httpEndpoint.SkipRequestBodyEncodeDecode {
			ed.BuildStreamPayload = endpointSymbols.buildStreamPayload.Name()
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
		sds.buildRequestAttributeTypes(sd.bodies.request(a), wireRequestBody, sd)

		if a.MethodExpr.StreamingPayload.Type != expr.Empty {
			sds.buildRequestAttributeTypes(sd.bodies.streaming(a), wireStreamPayload, sd)
		}
	}

	return sd
}

// buildRequestAttributeTypes builds nested request declarations from separate
// tagged copies because server and client packages apply different pointer and
// default policies to the same authored body graph.
func (sds *ServicesData) buildRequestAttributeTypes(body *expr.AttributeExpr, role wireTypeRole, data *ServiceData) {
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
			declaration := sds.attributeTypeData(userType, true, side.pointer, side.server, wireUnionUse{role: role}, data)
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
// the generated client and server packages. NewPlans calls it before
// Generation.Freeze chooses Go names, and Link later uses these same copied
// values.
func collectPlannedWireTypes(api string, httpService *expr.HTTPServiceExpr, planned *plannedWireTypes, servicePlan *service.Plan) {
	bodies, server, client := &planned.bodies, planned.server, planned.client
	for _, endpoint := range httpService.HTTPEndpoints {
		request := expr.DupAtt(bodies.request(endpoint))
		addMarshalTags(request)
		serverRequestPolicy := jsonBodyPolicy(true, true, true, "")
		clientRequestPolicy := jsonBodyPolicy(true, false, false, "")
		server.collect(request, wireRequestBody, serverRequestPolicy, api)
		server.addValidationRoot(request, serverRequestPolicy)
		clientRequest := client.collect(request, wireRequestBody, clientRequestPolicy, api)
		if userType, named := request.Type.(expr.UserType); named && userType.Attribute().Validation != nil {
			client.addValidationRoot(request, clientRequestPolicy)
		}
		if clientRequest != nil && needInit(request.Type) {
			clientRequest.needsConstructor = true
		} else if needInit(request.Type) {
			key := clientBodyConstructorKey{endpoint: endpoint, role: wireRequestBody}
			planned.clientBodyConstructorNames[key] = anonymousClientBodyConstructorName(request, clientRequestPolicy)
		}
		requestUse := wireUnionUse{role: wireRequestBody}
		server.collectChildren(request, requestUse, jsonBodyPolicy(true, true, true, ""), api)
		client.collectChildren(request, requestUse, jsonBodyPolicy(true, false, true, ""), api)
		if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty {
			streaming := expr.DupAtt(bodies.streaming(endpoint))
			addMarshalTags(streaming)
			serverStreamPolicy := jsonBodyPolicy(true, true, true, "")
			clientStreamPolicy := jsonBodyPolicy(true, false, false, "")
			serverStream := server.collect(streaming, wireStreamPayload, serverStreamPolicy, api)
			server.addValidationRoot(streaming, serverStreamPolicy)
			if endpoint.UsesWebSocket() && needInit(endpoint.MethodExpr.StreamingPayload.Type) && serverStream != nil {
				serverStream.needsConstructor = true
				planned.streamPayloads[endpoint] = serverStream
			}
			clientStream := client.collect(streaming, wireStreamPayload, clientStreamPolicy, api)
			if userType, named := streaming.Type.(expr.UserType); !named || userType.Attribute().Validation != nil {
				client.addValidationRoot(streaming, clientStreamPolicy)
			}
			if clientStream != nil && needInit(streaming.Type) {
				clientStream.needsConstructor = true
			} else if needInit(streaming.Type) {
				key := clientBodyConstructorKey{endpoint: endpoint, role: wireStreamPayload}
				planned.clientBodyConstructorNames[key] = anonymousClientBodyConstructorName(streaming, clientStreamPolicy)
			}
			streamUse := wireUnionUse{role: wireStreamPayload}
			server.collectChildren(streaming, streamUse, jsonBodyPolicy(true, true, true, ""), api)
			client.collectChildren(streaming, streamUse, jsonBodyPolicy(true, false, true, ""), api)
		}
		if endpoint.UsesSSE() && endpoint.MethodExpr.HasMixedResults() {
			body := bodies.streamingResult(endpoint)
			collectResponseWireType(api, body, body, endpoint, server, true, nil, "")
			collectResponseWireType(api, body, body, endpoint, client, false, nil, "")
		}

		resultType, viewed := endpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr)
		for _, response := range endpoint.Responses {
			body := bodies.response(response)
			if !viewed {
				collectResponseWireType(api, body, body, endpoint, server, true, nil, "")
				collectResponseWireType(api, body, body, endpoint, client, false, nil, "")
				continue
			}
			origin := ""
			if value, ok := body.Meta["origin:attribute"]; ok {
				origin = value[0]
			}
			_, explicitBody := body.Meta["http:body"]
			emptyView := ""
			switch {
			case origin != "":
				collectResponseWireType(api, body, body, endpoint, server, true, &emptyView, "")
			case endpoint.MethodExpr.Result.Meta != nil:
				if view, ok := endpoint.MethodExpr.Result.Meta.Last(expr.ViewMetaKey); ok {
					collectResponseWireType(api, body, body, endpoint, server, true, &view, "")
				} else {
					for _, view := range resultType.Views {
						collectResponseWireType(api, body, body, endpoint, server, true, &view.Name, "")
					}
				}
			default:
				for _, view := range resultType.Views {
					collectResponseWireType(api, body, body, endpoint, server, true, &view.Name, "")
				}
			}
			clientView := clientResponseViewNameExpr(endpoint, resultType)
			if origin != "" {
				emptyView := ""
				collectResponseWireType(api, body, body, endpoint, client, false, &emptyView, "")
				continue
			}
			if clientView == "" && explicitBody {
				emptyView := ""
				collectResponseWireType(api, body, body, endpoint, client, false, &emptyView, "")
				continue
			}
			if clientView != "" {
				clientBody := effectiveClientResponseBodyForView(body, clientView, endpoint)
				collectResponseWireType(api, clientBody, body, endpoint, client, false, &clientView, "")
				continue
			}
			for _, view := range resultType.Views {
				clientBody := effectiveClientResponseBodyForView(body, view.Name, endpoint)
				collectResponseWireType(api, clientBody, body, endpoint, client, false, &view.Name, "")
			}
		}
		for _, transportError := range endpoint.HTTPErrors {
			body := bodies.errorResponse(transportError)
			collectResponseWireType(api, body, body, endpoint, server, true, nil, transportError.Name)
			collectResponseWireType(api, body, body, endpoint, client, false, nil, transportError.Name)
		}
		collectPlannedTransforms(endpoint, planned, servicePlan)
	}
}

// anonymousClientBodyConstructorName returns the preferred function name for
// a request body that uses a Go expression such as []T instead of declaring a
// package type.
func anonymousClientBodyConstructorName(body *expr.AttributeExpr, policy wireTypePolicy) string {
	scope := codegen.NewAttributeScope(codegen.NewNameScope())
	name := scope.Name(body, "", policy.pointer, policy.useDefault)
	return "New" + codegen.Goify(name, true)
}

// collectPlannedTransforms records each request, response, error, and stream
// conversion and stores its handle with the endpoint value that will write it.
// The generated package can then name every helper before Plan.Link.
func collectPlannedTransforms(
	endpoint *expr.HTTPEndpointExpr,
	planned *plannedWireTypes,
	servicePlan *service.Plan,
) {
	methodName := endpoint.MethodExpr.Name
	bodies := &planned.bodies
	server := planned.server
	client := planned.client
	servicePackage, viewsPackage, err := servicePlan.MethodPackageImports(endpoint.MethodExpr)
	if err != nil {
		panic(err)
	}
	request := expr.DupAtt(bodies.request(endpoint))
	addMarshalTags(request)
	payload := endpoint.MethodExpr.Payload
	if needInit(payload.Type) {
		if request.Type != expr.Empty {
			target := payload
			if origin, ok := request.Meta["origin:attribute"]; ok {
				target = expr.AsObject(payload.Type).Attribute(origin[0])
			}
			requestTransforms := planned.transforms.request(endpoint, wireRequestBody)
			if needInit(request.Type) {
				requestTransforms.clientEncode = client.collectTransform(target, request, "marshal", methodName+" request body", wireTransformLayout{
					wireSide:       wireTransformTarget,
					wirePolicy:     jsonBodyPolicy(true, false, false, ""),
					wireUse:        wireUnionUse{role: wireRequestBody},
					servicePackage: *servicePackage,
				})
			}
			requestTransforms.serverDecode = server.collectTransform(request, target, "unmarshal", methodName+" server payload", wireTransformLayout{
				wireSide:       wireTransformSource,
				wirePolicy:     jsonBodyPolicy(true, true, false, ""),
				wireUse:        wireUnionUse{role: wireRequestBody},
				servicePackage: *servicePackage,
			})
			requestTransforms.clientDecode = client.collectTransform(request, target, "marshal", methodName+" command payload", wireTransformLayout{
				wireSide:       wireTransformSource,
				wirePolicy:     jsonBodyPolicy(true, false, false, ""),
				wireUse:        wireUnionUse{role: wireRequestBody},
				servicePackage: *servicePackage,
			})
		} else if expr.IsArray(payload.Type) || expr.IsMap(payload.Type) {
			if params := expr.AsObject(endpoint.Params.Type); len(*params) > 0 {
				requestTransforms := planned.transforms.request(endpoint, wireRequestBody)
				requestTransforms.serverDecode = server.collectTransform((*params)[0].Attribute, payload, "unmarshal", methodName+" server parameters", wireTransformLayout{
					wireSide:       wireTransformSource,
					wirePolicy:     wireTypePolicy{request: true, pointer: true},
					wireUse:        wireUnionUse{role: wireRequestBody},
					servicePackage: *servicePackage,
				})
				requestTransforms.clientDecode = client.collectTransform((*params)[0].Attribute, payload, "marshal", methodName+" command parameters", wireTransformLayout{
					wireSide:       wireTransformSource,
					wirePolicy:     wireTypePolicy{request: true, useDefault: true},
					wireUse:        wireUnionUse{role: wireRequestBody},
					servicePackage: *servicePackage,
				})
			}
		}
	}

	result := endpoint.MethodExpr.Result
	resultType, viewed := result.Type.(*expr.ResultTypeExpr)
	resultPackage := *servicePackage
	if viewed {
		if viewsPackage == nil {
			panic(fmt.Sprintf("viewed method %q has no views package", methodName))
		}
		resultPackage = *viewsPackage
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
		_, explicitBody := body.Meta["http:body"]
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
			prepared, viewName := prepareResponseWireBody(body, view)
			if prepared.Type != expr.Empty && resultAttribute.Type != expr.Empty && needInit(prepared.Type) {
				responseTransforms := planned.transforms.response(endpoint, response, viewName)
				responseTransforms.serverEncode = server.collectTransform(resultAttribute, prepared, "marshal", transformResponseOwner(methodName, response, view, "server"), wireTransformLayout{
					wireSide:       wireTransformTarget,
					wirePolicy:     jsonBodyPolicy(false, true, false, viewName),
					wireUse:        wireUnionUse{role: wireResponseBody, view: viewName},
					servicePointer: view != nil,
					servicePackage: resultPackage,
				})
			}
		}

		if !needClientResponseInit(result.Type) {
			continue
		}
		clientViewCount := 1
		if viewed {
			clientViewCount += len(resultType.Views)
		}
		clientViews := make([]*string, 0, clientViewCount)
		if !viewed {
			clientViews = append(clientViews, nil)
		} else {
			selected := clientResponseViewNameExpr(endpoint, resultType)
			switch {
			case origin != "":
				empty := ""
				clientViews = append(clientViews, &empty)
			case selected != "":
				clientViews = append(clientViews, &selected)
			case explicitBody:
				empty := ""
				clientViews = append(clientViews, &empty)
			default:
				for index := range resultType.Views {
					clientViews = append(clientViews, &resultType.Views[index].Name)
				}
			}
		}
		for _, view := range clientViews {
			clientBody := body
			if view != nil && *view != "" {
				clientBody = effectiveClientResponseBodyForView(body, *view, endpoint)
			}
			prepared, viewName := prepareResponseWireBody(clientBody, view)
			if prepared.Type != expr.Empty {
				responseTransforms := planned.transforms.response(endpoint, response, viewName)
				responseTransforms.clientDecode = client.collectTransform(prepared, resultAttribute, "unmarshal", transformResponseOwner(methodName, response, view, "client"), wireTransformLayout{
					wireSide:       wireTransformSource,
					wirePolicy:     jsonBodyPolicy(false, false, true, viewName),
					wireUse:        wireUnionUse{role: wireResponseBody, view: viewName},
					servicePointer: viewed,
					servicePackage: resultPackage,
				})
			}
		}
		if body.Type == expr.Empty && (expr.IsArray(result.Type) || expr.IsMap(result.Type)) {
			if params := expr.AsObject(endpoint.QueryParams().Type); len(*params) > 0 {
				responseTransforms := planned.transforms.response(endpoint, response, "")
				responseTransforms.clientDecode = client.collectTransform((*params)[0].Attribute, result, "unmarshal", transformResponseOwner(methodName, response, nil, "client parameters"), wireTransformLayout{
					wireSide:       wireTransformSource,
					wirePolicy:     wireTypePolicy{pointer: true},
					wireUse:        wireUnionUse{role: wireResponseBody},
					servicePointer: viewed,
					servicePackage: resultPackage,
				})
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
			errorTransforms := planned.transforms.transportError(transportError)
			if needInit(body.Type) {
				errorTransforms.serverEncode = server.collectTransform(target, body, "marshal", methodName+" server error "+transportError.Name, wireTransformLayout{
					wireSide:       wireTransformTarget,
					wirePolicy:     jsonBodyPolicy(false, true, false, ""),
					wireUse:        wireUnionUse{role: wireResponseBody},
					servicePackage: *servicePackage,
				})
			}
			errorTransforms.clientDecode = client.collectTransform(body, target, "unmarshal", methodName+" client error "+transportError.Name, wireTransformLayout{
				wireSide:       wireTransformSource,
				wirePolicy:     jsonBodyPolicy(false, false, true, ""),
				wireUse:        wireUnionUse{role: wireResponseBody},
				servicePackage: *servicePackage,
			})
		} else if body.Type == expr.Empty && (expr.IsArray(transportError.Type) || expr.IsMap(transportError.Type)) {
			if params := expr.AsObject(endpoint.QueryParams().Type); len(*params) > 0 {
				errorTransforms := planned.transforms.transportError(transportError)
				errorTransforms.clientDecode = client.collectTransform((*params)[0].Attribute, endpoint.MethodExpr.Error(transportError.Name).AttributeExpr, "unmarshal", methodName+" client error parameters "+transportError.Name, wireTransformLayout{
					wireSide:       wireTransformSource,
					wirePolicy:     wireTypePolicy{pointer: true},
					wireUse:        wireUnionUse{role: wireResponseBody},
					servicePackage: *servicePackage,
				})
			}
		}
	}

	if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty && endpoint.UsesWebSocket() {
		body := expr.DupAtt(bodies.streaming(endpoint))
		addMarshalTags(body)
		if body.Type != expr.Empty && needInit(endpoint.MethodExpr.StreamingPayload.Type) {
			requestTransforms := planned.transforms.request(endpoint, wireStreamPayload)
			requestTransforms.serverDecode = server.collectTransform(body, endpoint.MethodExpr.StreamingPayload, "marshal", methodName+" server stream payload", wireTransformLayout{
				wireSide:       wireTransformSource,
				wirePolicy:     jsonBodyPolicy(true, true, false, ""),
				wireUse:        wireUnionUse{role: wireStreamPayload},
				servicePackage: *servicePackage,
			})
			requestTransforms.clientEncode = client.collectTransform(endpoint.MethodExpr.StreamingPayload, body, "marshal", methodName+" client stream body", wireTransformLayout{
				wireSide:       wireTransformTarget,
				wirePolicy:     jsonBodyPolicy(true, false, false, ""),
				wireUse:        wireUnionUse{role: wireStreamPayload},
				servicePackage: *servicePackage,
			})
		}
	}
	if endpoint.UsesSSE() && endpoint.MethodExpr.HasMixedResults() {
		body, _ := prepareResponseWireBody(bodies.streamingResult(endpoint), nil)
		result := endpoint.MethodExpr.StreamingResult
		streamTransforms := planned.transforms.streamingResult(endpoint)
		serviceLayout, err := servicePlan.StreamingResultLayout(endpoint.MethodExpr)
		if err != nil {
			panic(err)
		}
		direct, err := sameMixedSSERepresentation(body, serviceLayout)
		if err != nil {
			panic(err)
		}
		if body.Type != expr.Empty && direct {
			streamTransforms.clientDecodeDirect = true
		}
		if body.Type != expr.Empty && !streamTransforms.clientDecodeDirect {
			if needInit(body.Type) {
				streamTransforms.serverEncode = server.collectTransform(result, body, "marshal", methodName+" server streaming result", wireTransformLayout{
					wireSide:       wireTransformTarget,
					wirePolicy:     jsonBodyPolicy(false, true, false, ""),
					wireUse:        wireUnionUse{role: wireResponseBody},
					servicePackage: *servicePackage,
				})
			}
			streamTransforms.clientDecode = client.collectTransform(body, result, "unmarshal", methodName+" client streaming result", wireTransformLayout{
				wireSide:       wireTransformSource,
				wirePolicy:     jsonBodyPolicy(false, false, true, ""),
				wireUse:        wireUnionUse{role: wireResponseBody},
				servicePackage: *servicePackage,
			})
		}
	}
}

// sameMixedSSERepresentation compares the retained service layout with the
// wire layout decoded by the client. Named values, unions, and structs always
// use a planned conversion; primitive values and their collections are direct
// only when every retained Go type detail matches.
func sameMixedSSERepresentation(wire *expr.AttributeExpr, serviceLayout *codegen.GoTypePlan) (bool, error) {
	if !mixedSSEDirectLayout(serviceLayout) {
		return false, nil
	}
	wireLayout, err := codegen.PlanGoType(wire, codegen.GoTypePlanOptions{
		Owner:  serviceLayout.Owner(),
		Policy: serviceLayout.Policy(),
	})
	if err != nil {
		return false, err
	}
	return serviceLayout.Equivalent(wireLayout), nil
}

// mixedSSEDirectLayout reports whether a layout can be assigned without any
// generated declaration or field-by-field conversion.
func mixedSSEDirectLayout(layout *codegen.GoTypePlan) bool {
	switch layout.Kind() {
	case codegen.GoPrimitive:
		return true
	case codegen.GoArray:
		return mixedSSEDirectLayout(layout.Elem())
	case codegen.GoMap:
		return mixedSSEDirectLayout(layout.Key()) && mixedSSEDirectLayout(layout.Elem())
	default:
		return false
	}
}

// request returns the retained conversions for one ordinary or streamed
// request body, creating the record during collection when needed.
func (p *plannedWireTransforms) request(
	endpoint *expr.HTTPEndpointExpr,
	role wireTypeRole,
) *plannedRequestTransforms {
	key := clientBodyConstructorKey{endpoint: endpoint, role: role}
	transforms := p.requests[key]
	if transforms == nil {
		transforms = &plannedRequestTransforms{}
		p.requests[key] = transforms
	}
	return transforms
}

// response returns the retained conversions for one status, tag, and view
// representation, creating the record during collection when needed.
func (p *plannedWireTransforms) response(
	endpoint *expr.HTTPEndpointExpr,
	response *expr.HTTPResponseExpr,
	view string,
) *plannedResponseTransforms {
	key := viewedConstructorKey{endpoint: endpoint, response: response, view: view}
	transforms := p.responses[key]
	if transforms == nil {
		transforms = &plannedResponseTransforms{}
		p.responses[key] = transforms
	}
	return transforms
}

// transportError returns the retained conversions for one designed error,
// creating the record during collection when needed.
func (p *plannedWireTransforms) transportError(
	transportError *expr.HTTPErrorExpr,
) *plannedResponseTransforms {
	transforms := p.errors[transportError]
	if transforms == nil {
		transforms = &plannedResponseTransforms{}
		p.errors[transportError] = transforms
	}
	return transforms
}

// streamingResult returns the retained conversions for a mixed method's
// streamed result, creating the record during collection when needed.
func (p *plannedWireTransforms) streamingResult(
	endpoint *expr.HTTPEndpointExpr,
) *plannedResponseTransforms {
	transforms := p.streamingResults[endpoint]
	if transforms == nil {
		transforms = &plannedResponseTransforms{}
		p.streamingResults[endpoint] = transforms
	}
	return transforms
}

// transformResponseOwner returns the method, transport side, status, tag, and
// view values that distinguish helper functions for two responses with the
// same generated Go types.
func transformResponseOwner(method string, response *expr.HTTPResponseExpr, view *string, side string) string {
	viewName := ""
	if view != nil {
		viewName = *view
	}
	return fmt.Sprintf("%s %s response %d %s %s %s", method, side, response.StatusCode, response.Tag[0], response.Tag[1], viewName)
}

// collectResponseWireType applies the selected view and records response body
// declarations using the same policy later consumed by buildResponseBodyType.
func collectResponseWireType(
	api string,
	body *expr.AttributeExpr,
	releasedBody *expr.AttributeExpr,
	endpoint *expr.HTTPEndpointExpr,
	catalog *wireTypeCatalog,
	server bool,
	view *string,
	errorName string,
) {
	body, viewName := prepareResponseWireBody(body, view)
	releasedNames := releasedResponseWireNames(releasedBody, body, view, endpoint)
	policy := jsonBodyPolicy(false, server, !server, viewName)
	preferred := ""
	if server && !expr.IsPrimitive(body.Type) && needInit(body.Type) {
		if _, userType := body.Type.(expr.UserType); !userType {
			preferred = codegen.Goify(endpoint.Name(), true) + "ResponseBody"
		}
	}
	record := catalog.collectWithReleasedNames(body, wireResponseBody, policy, preferred, releasedNames, api)
	if record != nil && errorName != "" {
		record.addErrorUse(wireErrorUse{
			service: endpoint.Service.Name(),
			method:  endpoint.Name(),
			name:    errorName,
		})
	}
	if policy.validate {
		catalog.addValidationRoot(body, policy)
	}
	if server && record != nil && needInit(body.Type) {
		record.needsConstructor = true
	}
	attributePolicy := jsonBodyPolicy(false, server, !server, "")
	responseUse := wireUnionUse{role: wireResponseBody, view: viewName}
	catalog.collectChildrenWithReleasedNames(body, responseUse, attributePolicy, releasedNames)
}

// prepareResponseWireBody copies the response body, selects the requested view
// fields, and adds JSON tags. Collection, declaration generation, and client
// response conversion all use the returned shape.
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

// releasedResponseWireNames returns the response type names produced when Goa
// added transport suffixes before selecting a result view.
func releasedResponseWireNames(original, prepared *expr.AttributeExpr, view *string, endpoint *expr.HTTPEndpointExpr) map[expr.UserType]string {
	released := expr.DupAtt(original)
	suffix := releasedWireTypeSuffix(released, wireResponseBody)
	if userType, ok := released.Type.(expr.UserType); ok {
		appendReleasedWireSuffix(userType.Attribute().Type, suffix, make(map[expr.UserType]struct{}))
	} else {
		appendReleasedWireSuffix(released.Type, suffix, make(map[expr.UserType]struct{}))
	}
	released, _ = prepareResponseWireBody(released, view)
	addJSONRPCSSEViewFields(released.Type, original.Type, endpoint)
	names := make(map[expr.UserType]string)
	collectReleasedWireNames(prepared.Type, released.Type, names, make(map[releasedWireTypePair]struct{}))
	if collection, ok := prepared.Type.(*expr.ResultTypeExpr); ok {
		if array := expr.AsArray(collection.Attribute().Type); array != nil {
			if element, ok := array.ElemType.Type.(expr.UserType); ok {
				names[collection] = names[element] + "Collection"
			}
		}
	}
	return names
}

// appendReleasedWireSuffix reproduces the order used by released Goa versions
// on a private copy of the response type.
func appendReleasedWireSuffix(dataType expr.DataType, suffix string, seen map[expr.UserType]struct{}) {
	switch actual := dataType.(type) {
	case expr.UserType:
		if _, ok := seen[actual]; ok {
			return
		}
		seen[actual] = struct{}{}
		actual.Rename(actual.Name() + suffix)
		appendReleasedWireSuffix(actual.Attribute().Type, suffix, seen)
	case *expr.Object:
		for _, named := range *actual {
			appendReleasedWireSuffix(named.Attribute.Type, suffix, seen)
		}
	case *expr.Array:
		appendReleasedWireSuffix(actual.ElemType.Type, suffix, seen)
	case *expr.Map:
		appendReleasedWireSuffix(actual.KeyType.Type, suffix, seen)
		appendReleasedWireSuffix(actual.ElemType.Type, suffix, seen)
	case *expr.Union:
		for _, named := range actual.Values {
			appendReleasedWireSuffix(named.Attribute.Type, suffix, seen)
		}
	}
}

// collectReleasedWireNames pairs the current response types with the names
// from the earlier suffix-before-view order.
func collectReleasedWireNames(current, released expr.DataType, names map[expr.UserType]string, seen map[releasedWireTypePair]struct{}) {
	currentUser, currentNamed := current.(expr.UserType)
	releasedUser, releasedNamed := released.(expr.UserType)
	if currentNamed || releasedNamed {
		if !currentNamed || !releasedNamed {
			panic("response view changed whether a generated type is named")
		}
		pair := releasedWireTypePair{current: currentUser, released: releasedUser}
		if _, ok := seen[pair]; ok {
			return
		}
		seen[pair] = struct{}{}
		names[currentUser] = codegen.Goify(releasedUser.Name(), true)
		collectReleasedWireNames(currentUser.Attribute().Type, releasedUser.Attribute().Type, names, seen)
		return
	}

	switch currentType := current.(type) {
	case *expr.Object:
		releasedType, ok := released.(*expr.Object)
		if !ok {
			panic("response view changed the generated object shape")
		}
		for _, named := range *currentType {
			other := releasedType.Attribute(named.Name)
			if other == nil {
				panic("response view changed a generated field name")
			}
			collectReleasedWireNames(named.Attribute.Type, other.Type, names, seen)
		}
	case *expr.Array:
		releasedType, ok := released.(*expr.Array)
		if !ok {
			panic("response view changed the generated array shape")
		}
		collectReleasedWireNames(currentType.ElemType.Type, releasedType.ElemType.Type, names, seen)
	case *expr.Map:
		releasedType, ok := released.(*expr.Map)
		if !ok {
			panic("response view changed the generated map shape")
		}
		collectReleasedWireNames(currentType.KeyType.Type, releasedType.KeyType.Type, names, seen)
		collectReleasedWireNames(currentType.ElemType.Type, releasedType.ElemType.Type, names, seen)
	case *expr.Union:
		releasedType, ok := released.(*expr.Union)
		if !ok || len(currentType.Values) != len(releasedType.Values) {
			panic("response view changed the generated union shape")
		}
		for index, named := range currentType.Values {
			collectReleasedWireNames(named.Attribute.Type, releasedType.Values[index].Attribute.Type, names, seen)
		}
	}
}

// makeHTTPType traverses the attribute recursively and performs these actions:
//
// * removes aliased user type by replacing them with the underlying type.
// * keeps unions as generated values with one selected branch.
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
		// Prepare every branch before the HTTP package catalog assigns the union's
		// request, streaming request, response, or response-view name.
		for _, branch := range dt.Values {
			branch.Attribute = makeHTTPTypeRecursive(branch.Attribute, seen)
		}
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

// streaming returns a copy of the endpoint's streaming request body. Generated
// JSON field information may be added to the copy without changing the authored
// design. Streaming requests keep named user types instead of replacing them
// with their fields.
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

// streamingResult returns the JSON body written for each result in a mixed SSE
// stream. It copies the service result before applying HTTP field names so
// generation never changes the authored service type.
func (b *shapedBodies) streamingResult(e *expr.HTTPEndpointExpr) *expr.AttributeExpr {
	if att, ok := b.streamResults[e]; ok {
		return att
	}
	if b.streamResults == nil {
		b.streamResults = make(map[*expr.HTTPEndpointExpr]*expr.AttributeExpr)
	}
	att := makeHTTPType(e.MethodExpr.StreamingResult)
	b.streamResults[e] = att
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
		serverPolicy := jsonBodyPolicy(true, true, true, "")
		clientPolicy := jsonBodyPolicy(true, false, true, "")
		sd.serverWireTypes.applyNames(serverHTTPBody, wireRequestBody, serverPolicy)
		sd.clientWireTypes.applyNames(clientHTTPBody, wireRequestBody, clientPolicy)
	}
	var (
		payload      = e.MethodExpr.Payload
		svc          = sd.Service
		body         = httpBody.Type
		ep           = svc.Method(e.MethodExpr.Name)
		httpsvrctx   = jsonBodyContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
		httpclictx   = jsonBodyContext(sd.clientWireTypes, sd.clientWireTypes.scope, true, false)
		svcsvrctx    = sds.serviceTypeContext(sd, "server").Enter(payload)
		svcclictx    = sds.serviceTypeContext(sd, "client").Enter(payload)
		payloadOwner = expr.MethodPayloadExampleIdentity(e.MethodExpr)
		bodyOwner    = expr.RequestBodyExampleIdentity(e)

		request         *RequestData
		mapQueryParam   *ParamData
		origin          string
		originAttribute *expr.AttributeExpr
		bodyDefault     any
	)
	httpsvrctx.Scope = sd.serverWireTypes.resolverForUse(
		sd.serverWireTypes.scope,
		jsonBodyPolicy(true, true, true, ""),
		wireUnionUse{role: wireRequestBody},
	)
	httpclictx.Scope = sd.clientWireTypes.resolverForUse(
		sd.clientWireTypes.scope,
		jsonBodyPolicy(true, false, true, ""),
		wireUnionUse{role: wireRequestBody},
	)
	idPolicy := jsonRPCRequestIDPolicyFor(e)
	var idData *JSONRPCRequestIDData
	if idPolicy != nil {
		idData = &JSONRPCRequestIDData{Generate: idPolicy.generates()}
		if idPolicy.attribute != nil {
			field := idPolicy.attribute
			attribute := field.Attribute
			variable := sd.Scope.Name("requestID")
			idData.Attribute = codegen.Goify(field.Name, true)
			idData.Variable = variable
			idData.ValueTypeRef = svcsvrctx.Scope.Ref(attribute, svcsvrctx.Pkg(attribute))
			idData.Aliased = expr.IsAlias(attribute.Type)
			idData.Required = idPolicy.required
			idData.HasDefault = idPolicy.defaultValue != nil
			idData.MustHave = idPolicy.required && !idData.HasDefault
			idData.Pointer = idPolicy.pointer
			if idData.HasDefault {
				idData.DefaultValue = idPolicy.defaultValue.(string)
			}
			idData.Validate = codegen.AttributeValidationCode(attribute, nil, svcsvrctx, idPolicy.required, idData.Aliased, variable, field.Name)
		}
	}
	{
		var (
			serverBodyData = sds.buildRequestBodyType(httpBody, payload, e, wireRequestBody, true, sd, payloadOwner, bodyOwner)
			clientBodyData = sds.buildRequestBodyType(httpBody, payload, e, wireRequestBody, false, sd, payloadOwner, bodyOwner)
			paramsData     = sds.extractPathParams(e.PathParams(), payload, sd, payloadOwner)
			queryData      = sds.extractQueryParams(e.QueryParams(), payload, sd, payloadOwner)
			headersData    = sds.extractHeaders(e.Headers, payload, svcsvrctx, sd.Scope, payloadOwner)
			cookiesData    = sds.extractCookies(e.Cookies, payload, svcsvrctx, sd.Scope, payloadOwner)
			mustValidate   bool
			mustHaveBody   = true
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
			typeName := sd.Scope.GoTypeName(pAtt)
			typeRef := sd.Scope.GoTypeRef(pAtt)
			validate := codegen.AttributeValidationCode(pAtt, nil, httpsvrctx, required, expr.IsAlias(pAtt.Type), varn, name)
			mapQueryParam = &ParamData{
				MapQueryParams: e.MapQueryParams,
				Element: &Element{
					HTTPName: name,
					AttributeData: &AttributeData{
						Name:      name,
						VarName:   varn,
						FieldName: fieldName,
						FieldType: pAtt.Type,
						Required:  required,
						Type:      pAtt.Type,
						TypeName:  typeName,
						TypeRef:   typeRef,
						Validate:  validate,
						CLIPlan: cli.NewFlagPlan(
							pAtt,
							typeName,
							typeRef,
							cliValidationRenderer(validate != "", pAtt, httpsvrctx, name),
						),
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
				if h.PreserveEmpty || h.Validate != "" || h.Required || needConversion(h.Type) {
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
				originAttribute = expr.AsObject(payload.Type).Attribute(origin)
				bodyDefault = payload.GetDefault(origin)
				if !payload.IsRequired(o[0]) {
					mustHaveBody = false
				}
			}
		}
		bodyIsUnion := originAttribute != nil && expr.AsUnion(originAttribute.Type) != nil
		bodyFieldPointer := origin != "" && payload.IsPrimitivePointer(origin, true)
		bodyFieldCanBeAbsent := originAttribute != nil && !payload.IsRequired(origin) &&
			(bodyFieldPointer || codegen.IsNilable(originAttribute.Type) || bodyIsUnion)
		request = &RequestData{
			PathParams:           paramsData,
			QueryParams:          queryData,
			Headers:              headersData,
			Cookies:              cookiesData,
			ServerBody:           serverBodyData,
			ClientBody:           clientBodyData,
			PayloadAttr:          codegen.Goify(origin, true),
			PayloadType:          e.MethodExpr.Payload.Type,
			MustHaveBody:         mustHaveBody,
			OptionalBody:         origin != "" && !payload.IsRequired(origin),
			BodyIsUnion:          bodyIsUnion,
			BodyFieldPointer:     bodyFieldPointer,
			BodyFieldCanBeAbsent: bodyFieldCanBeAbsent,
			MustValidate:         mustValidate,
			Multipart:            e.MultipartRequest,
		}
		if e.IsJSONRPC() && clientBodyData != nil {
			paramsTypeRef := clientBodyData.Ref
			if clientBodyData.Init == nil {
				paramsAttribute := payload
				if origin != "" {
					paramsAttribute = expr.AsObject(payload.Type).Attribute(origin)
				}
				paramsContext := svcclictx.Enter(paramsAttribute)
				paramsLayout, layoutErr := paramsContext.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(paramsAttribute, paramsContext.LayoutPolicy())
				if layoutErr != nil {
					sds.recordLinkError(layoutErr)
					return nil
				}
				if bodyFieldPointer {
					paramsTypeRef = paramsLayout.RefWithPointer(true)
				} else {
					paramsTypeRef = paramsLayout.Ref()
				}
			}
			request.JSONRPCParams = jsonRPCParams(httpBody.Type, paramsTypeRef, mustHaveBody, !mustHaveBody)
		}
	}

	var init *InitData
	if needInit(payload.Type) {
		// generate constructor function to transform request body,
		// params, headers and cookies into the method payload type
		var (
			name                  string
			desc                  string
			isObject              bool
			clientArgs            []*InitArgData
			serverArgs            []*InitArgData
			serverBodyDereference bool
			clientBodyDereference bool
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
				svcode            string
				cvcode            string
				serverTypeName    string
				serverTypeRef     string
				clientTypeName    string
				clientTypeRef     string
				clientBodyPointer bool
			)
			serverLayout, layoutErr := httpsvrctx.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(serverHTTPBody, httpsvrctx.LayoutPolicy())
			if layoutErr != nil {
				sds.recordLinkError(layoutErr)
			} else {
				serverTypeName = serverLayout.Name()
				serverTypeRef, _, serverBodyDereference = optionalWireTypeRef(serverLayout, request.OptionalBody)
			}
			clientLayout, layoutErr := httpclictx.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(clientHTTPBody, httpclictx.LayoutPolicy())
			if layoutErr != nil {
				sds.recordLinkError(layoutErr)
			} else {
				clientTypeName = clientLayout.Name()
				clientTypeRef, clientBodyPointer, clientBodyDereference = optionalWireTypeRef(clientLayout, request.BodyFieldCanBeAbsent)
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
			cliValidation := cliValidationRenderer(cvcode != "", clientHTTPBody, httpclictx, "body")
			clientDefault := clientBodyDefault(httpBody, bodyDefault)
			serverBodyRef := sd.serverWireTypes.scope.GoVar("body", serverHTTPBody.Type)
			if request.OptionalBody {
				serverBodyRef = "body"
			}
			serverArgs = append(serverArgs, &InitArgData{
				Ref: serverBodyRef,
				AttributeData: &AttributeData{
					Name:         "body",
					VarName:      "body",
					TypeName:     serverTypeName,
					TypeRef:      serverTypeRef,
					Type:         serverHTTPBody.Type,
					Required:     !request.OptionalBody,
					DefaultValue: bodyDefault,
					Example:      sds.Example(httpBody, bodyOwner),
					Validate:     svcode,
				},
			})
			clientBodyArgRef := sd.clientWireTypes.scope.GoVar("body", clientHTTPBody.Type)
			if request.OptionalBody {
				clientBodyArgRef = "body"
			}
			clientArgs = append(clientArgs, &InitArgData{
				Ref: clientBodyArgRef,
				AttributeData: &AttributeData{
					Name:         "body",
					VarName:      "body",
					TypeName:     clientTypeName,
					TypeRef:      clientTypeRef,
					Type:         clientHTTPBody.Type,
					Required:     !request.OptionalBody,
					Pointer:      clientBodyPointer,
					DefaultValue: clientDefault,
					Example:      sds.Example(httpBody, bodyOwner),
					Validate:     cvcode,
					CLIPlan: cli.NewFlagPlan(
						clientHTTPBody,
						clientTypeName,
						clientTypeName,
						cliValidation,
					),
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
		if idPolicy != nil && idPolicy.attribute != nil {
			serverArgs = append(serverArgs, sds.jsonRPCRequestIDInitArg(idPolicy, idData.Variable, payload, svcsvrctx, payloadOwner, false))
			clientArgs = append(clientArgs, sds.jsonRPCRequestIDInitArg(idPolicy, idData.Variable, payload, svcclictx, payloadOwner, true))
		}

		var (
			cliArgs []*InitArgData
		)
		for _, r := range ep.Requirements {
			done := false
			for _, sc := range r.Schemes {
				if sc.Type == "Basic" {
					uatt := e.MethodExpr.Payload.Find(sc.UsernameAttr)
					uctx := svcclictx.Enter(uatt)
					ulayout, layoutErr := uctx.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(uatt, uctx.LayoutPolicy())
					if layoutErr != nil {
						sds.recordLinkError(layoutErr)
						return nil
					}
					uref := ulayout.Ref()
					uvalueRef := uref
					uvalidate := codegen.ValidationCode(uatt, nil, httpsvrctx, sc.UsernameRequired, expr.IsAlias(uatt.Type), false, sc.UsernameAttr)
					if sc.UsernamePointer {
						uref = ulayout.RefWithPointer(true)
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
							Validate:     uvalidate,
							CLIPlan: cli.NewFlagPlan(
								uatt,
								uctx.Scope.Name(uatt, uctx.Pkg(uatt), false, true),
								uvalueRef,
								cliValidationRenderer(uvalidate != "", uatt, uctx, sc.UsernameAttr),
							),
							Example: sds.FieldExample(uatt, e.MethodExpr.Payload, sc.UsernameAttr, payloadOwner),
						},
					}
					patt := e.MethodExpr.Payload.Find(sc.PasswordAttr)
					pctx := svcclictx.Enter(patt)
					playout, layoutErr := pctx.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(patt, pctx.LayoutPolicy())
					if layoutErr != nil {
						sds.recordLinkError(layoutErr)
						return nil
					}
					pref := playout.Ref()
					pvalueRef := pref
					pvalidate := codegen.ValidationCode(patt, nil, httpsvrctx, sc.PasswordRequired, expr.IsAlias(patt.Type), false, sc.PasswordAttr)
					if sc.PasswordPointer {
						pref = playout.RefWithPointer(true)
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
							Validate:     pvalidate,
							CLIPlan: cli.NewFlagPlan(
								patt,
								pctx.Scope.Name(patt, pctx.Pkg(patt), false, true),
								pvalueRef,
								cliValidationRenderer(pvalidate != "", patt, pctx, sc.PasswordAttr),
							),
							Example: sds.FieldExample(patt, e.MethodExpr.Payload, sc.PasswordAttr, payloadOwner),
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
		if len(cliArgs) > 0 {
			for index, arg := range cliArgs {
				field := e.MethodExpr.Payload.Find(arg.Name)
				ctx := svcsvrctx.Enter(field)
				layout, layoutErr := ctx.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(field, ctx.LayoutPolicy())
				if layoutErr != nil {
					sds.recordLinkError(layoutErr)
					return nil
				}
				ref := "user"
				if index == 1 {
					ref = "pass"
				}
				if arg.FieldPointer {
					ref += "Ptr"
				}
				typeRef := "string"
				if arg.FieldPointer {
					typeRef = "*string"
				}
				serverArgs = append(serverArgs, &InitArgData{
					Ref: ref,
					AttributeData: &AttributeData{
						Name:           arg.Name,
						VarName:        arg.VarName,
						TypeRef:        typeRef,
						Type:           expr.String,
						Pointer:        arg.FieldPointer,
						FieldName:      arg.FieldName,
						FieldType:      arg.FieldType,
						FieldPointer:   arg.FieldPointer,
						ServiceTypeRef: layout.Ref(),
					},
				})
			}
		}

		var (
			serverCode string
			clientCode string
			err        error
			pointer    bool
		)
		requestTransforms := sd.transforms.requests[clientBodyConstructorKey{endpoint: e, role: wireRequestBody}]
		if body != expr.Empty {
			if origin != "" {
				pointer = payload.IsPrimitivePointer(origin, true)
			}

			var (
				helpers []*codegen.TransformFunctionData
			)
			transformctx := jsonBodyContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
			serverBodyRef := "body"
			if serverBodyDereference {
				serverBodyRef = "*body"
			}
			serverCode, helpers, err = sd.serverWireTypes.renderTransform(requestTransforms.serverDecode, serverHTTPBody, serverBodyRef, "v", transformctx, svcsvrctx)
			if err == nil {
				sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
			} else {
				sds.recordLinkError(err)
			}
			// The client code for building the method payload from a request
			// body is used by the CLI tool to build the payload given to the
			// client endpoint. It follows the encoded request shape because this
			// path does not validate a decoded body.
			transformctx = jsonBodyContext(sd.clientWireTypes, sd.clientWireTypes.scope, true, false)
			clientBodyRef := "body"
			if clientBodyDereference {
				clientBodyRef = "*body"
			}
			clientCode, helpers, err = sd.clientWireTypes.renderTransform(requestTransforms.clientDecode, clientHTTPBody, clientBodyRef, "v", transformctx, svcclictx)
			if err == nil {
				sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
			} else {
				sds.recordLinkError(err)
			}
		} else if expr.IsArray(payload.Type) || expr.IsMap(payload.Type) {
			if params := expr.AsObject(e.Params.Type); len(*params) > 0 {
				var helpers []*codegen.TransformFunctionData
				transformctx := wireHTTPContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
				serverCode, helpers, err = sd.serverWireTypes.renderTransform(requestTransforms.serverDecode, (*params)[0].Attribute, codegen.Goify((*params)[0].Name, false), "v", transformctx, svcsvrctx)
				if err == nil {
					sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
				} else {
					sds.recordLinkError(err)
				}
				transformctx = wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, true, false)
				clientCode, helpers, err = sd.clientWireTypes.renderTransform(requestTransforms.clientDecode, (*params)[0].Attribute, codegen.Goify((*params)[0].Name, false), "v", transformctx, svcclictx)
				if err == nil {
					sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
				} else {
					sds.recordLinkError(err)
				}
			}
		}
		if err != nil {
			sds.recordLinkError(err)
		}
		var renderedBodyDefault *codegen.GoValueCode
		if bodyDefault != nil {
			var resolveUnion codegen.UnionConstructorResolver
			if resolver, ok := svcsvrctx.Scope.(interface {
				UnionConstructor(*expr.AttributeExpr, string) (string, error)
			}); ok {
				resolveUnion = resolver.UnionConstructor
			}
			fieldPointer := origin != "" && svcsvrctx.IsFieldPointer(origin, payload)
			layoutResolver, ok := svcsvrctx.Scope.(codegen.GoTypeLayoutResolver)
			if !ok {
				sds.recordLinkError(fmt.Errorf("service package cannot resolve the Go layout for %s", originAttribute.Type.Name()))
			} else if layout, layoutErr := layoutResolver.GoTypeLayout(originAttribute, svcsvrctx.LayoutPolicy()); layoutErr != nil {
				sds.recordLinkError(layoutErr)
			} else {
				rendered, renderErr := codegen.RenderGoValue(
					originAttribute,
					bodyDefault,
					layout,
					fieldPointer,
					resolveUnion,
					"bodyDefault",
				)
				if renderErr != nil {
					sds.recordLinkError(renderErr)
				} else {
					renderedBodyDefault = &rendered
				}
			}
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
			ReturnIsUnion:            request.BodyIsUnion,
			OptionalBody:             request.OptionalBody,
			ClientBodyOptional:       request.BodyFieldCanBeAbsent,
			BodyDefault:              renderedBodyDefault,
		}
	}
	request.PayloadInit = init

	var (
		returnValue string
		name        string
		ref         string
		cliPlan     *cli.FlagPlan
	)
	if payload.Type != expr.Empty {
		name = svcsvrctx.Scope.Name(payload, svcsvrctx.Pkg(payload), false, true)
		ref = svcsvrctx.Scope.Ref(payload, svcsvrctx.Pkg(payload))
		cliPlan = cli.NewFlagPlan(payload, name, ref, nil)
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
		CLIPlan:            cliPlan,
		Request:            request,
		DecoderReturnValue: returnValue,
		JSONRPCRequestID:   idData,
	}
	if e.IsJSONRPC() {
		obj := expr.AsObject(e.MethodExpr.Payload.Type)
		if obj != nil {
			for _, att := range *obj {
				if _, ok := att.Attribute.Meta["jsonrpc:id"]; ok {
					data.IDAttribute = codegen.Goify(att.Name, true)
					if expr.IsAlias(att.Attribute.Type) {
						data.IDAttributeTypeRef = svcsvrctx.Scope.Ref(att.Attribute, svcsvrctx.Pkg(att.Attribute))
					}
					data.IDAttributeRequired = e.MethodExpr.Payload.IsRequired(att.Name)
					break
				}
			}
		}
	}
	return data
}

// jsonRPCParams returns the exact JSON-RPC params shape for one service value.
// Primitive values use one positional item. Nil optional objects, arrays, and
// maps omit params because JSON-RPC does not allow params to be null.
func jsonRPCParams(dataType expr.DataType, typeRef string, rejectNull, allowAbsent bool) *JSONRPCParamsData {
	union := expr.AsUnion(dataType) != nil
	return &JSONRPCParamsData{
		Positional:  expr.IsPrimitive(dataType),
		TypeRef:     typeRef,
		RejectNull:  rejectNull,
		AllowAbsent: allowAbsent,
		OmitAbsent:  allowAbsent && !expr.IsPrimitive(dataType) && (codegen.IsNilable(dataType) || union),
	}
}

// jsonRPCRequestIDInitArg builds the payload-constructor argument that carries
// the envelope ID separately from params. Client arguments also describe the
// command-line flag used to supply a caller-chosen ID.
func (sds *ServicesData) jsonRPCRequestIDInitArg(policy *jsonRPCRequestIDPolicy, variable string, payload *expr.AttributeExpr, context *codegen.AttributeContext, owner expr.ExampleIdentity, client bool) *InitArgData {
	field := policy.attribute
	attribute := field.Attribute
	required := policy.required
	pointer := policy.pointer
	layout, err := context.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(attribute, context.LayoutPolicy())
	if err != nil {
		sds.recordLinkError(err)
		return nil
	}
	valueTypeRef := layout.Ref()
	typeRef := valueTypeRef
	if pointer {
		typeRef = layout.RefWithPointer(true)
	}
	data := &AttributeData{
		Name:           field.Name,
		VarName:        variable,
		Pointer:        pointer,
		Required:       required,
		Type:           attribute.Type,
		TypeName:       context.Scope.Name(attribute, context.Pkg(attribute), false, true),
		TypeRef:        typeRef,
		Description:    attribute.Description,
		FieldName:      codegen.Goify(field.Name, true),
		FieldType:      attribute.Type,
		ServiceTypeRef: valueTypeRef,
		FieldPointer:   pointer,
		DefaultValue:   policy.defaultValue,
		Example:        sds.FieldExample(attribute, payload, field.Name, owner),
	}
	if client {
		validate := codegen.AttributeValidationCode(attribute, nil, context, required, expr.IsAlias(attribute.Type), variable, field.Name)
		data.CLIPlan = cli.NewFlagPlan(
			attribute,
			data.TypeName,
			valueTypeRef,
			cliValidationRenderer(validate != "", attribute, context, field.Name),
		)
	}
	return &InitArgData{AttributeData: data, Ref: variable}
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
	return &ResultData{
		IsStruct:  expr.IsObject(result.Type),
		Name:      name,
		Ref:       ref,
		Responses: responses,
		View:      view,
		MustInit:  mustInit,
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
		defaultResponseIndex := -1
		for _, resp := range e.Responses {
			respBody := sd.bodies.response(resp)
			resultOwner := expr.MethodResultExampleIdentity(e.MethodExpr)
			bodyOwner := expr.ResponseBodyExampleIdentity(e, resp)
			if resp.Tag[0] == "" {
				defaultResponseIndex = len(responses)
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
				_, explicitBody := respBody.Meta["http:body"]
				if viewed {
					vname := ""
					clientView := clientResponseViewName(e, md)
					if origin != "" {
						// Response body is explicitly set to an attribute in the method
						// result type. No need to do any view-based projections server side.
						transforms := sd.transforms.responses[viewedConstructorKey{endpoint: e, response: resp, view: vname}]
						if sbd := sds.buildResponseBodyType(respBody, result, e, true, &vname, sd, transforms, resultOwner, bodyOwner); sbd != nil {
							serverBodyData = append(serverBodyData, sbd)
						}
					} else if v, ok := e.MethodExpr.Result.Meta.Last(expr.ViewMetaKey); ok {
						// Design explicitly sets the view to render the result.
						// We generate only one server body type which will be rendered
						// using the specified view.
						transforms := sd.transforms.responses[viewedConstructorKey{endpoint: e, response: resp, view: v}]
						if sbd := sds.buildResponseBodyType(respBody, result, e, true, &v, sd, transforms, resultOwner, bodyOwner); sbd != nil {
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
							transforms := sd.transforms.responses[viewedConstructorKey{endpoint: e, response: resp, view: view.Name}]
							if sbd := sds.buildResponseBodyType(respBody, result, e, true, &view.Name, sd, transforms, resultOwner, bodyOwner); sbd != nil {
								serverBodyData = append(serverBodyData, sbd)
							}
						}
					}
					switch {
					case origin != "":
						clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, &vname, sd, nil, resultOwner, bodyOwner)
						clientBodyView = &vname
					case clientView != "":
						clientRespBody = effectiveClientResponseBodyForView(respBody, clientView, e)
						clientBodyData = sds.buildResponseBodyType(clientRespBody, result, e, false, &clientView, sd, nil, resultOwner, bodyOwner)
						clientBodyView = &clientView
					case explicitBody:
						clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, &vname, sd, nil, resultOwner, bodyOwner)
						clientBodyView = &vname
					default:
						clientRespBody = &expr.AttributeExpr{Type: expr.Empty}
					}
				} else {
					transforms := sd.transforms.responses[viewedConstructorKey{endpoint: e, response: resp}]
					if sbd := sds.buildResponseBodyType(respBody, result, e, true, nil, sd, transforms, resultOwner, bodyOwner); sbd != nil {
						serverBodyData = append(serverBodyData, sbd)
					}
					clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, nil, sd, nil, resultOwner, bodyOwner)
				}
				if clientBodyData != nil && clientRespBody.Type != expr.Empty {
					var viewName string
					clientRespBody, viewName = prepareResponseWireBody(clientRespBody, clientBodyView)
					policy := jsonBodyPolicy(false, false, true, viewName)
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
				variableView := viewed && origin == "" && clientResponseViewName(e, md) == ""
				variableWire := variableView && !explicitBody
				selectClientBodyByView := variableWire &&
					!e.IsJSONRPC() &&
					(!e.MethodExpr.IsStreaming() || e.MethodExpr.HasMixedResults())
				if needClientResponseInit(result.Type) && !variableWire {
					init = sds.buildResponseResultInit(
						e, resp, result, clientRespBody, origin,
						headersData, cookiesData, sd, "", clientBodyData,
					)
				}

				var representations []*ViewedRepresentationData
				if viewed && (e.UsesSSE() || e.IsJSONRPC() || (e.UsesWebSocket() && variableWire) || selectClientBodyByView) {
					clientView := clientResponseViewName(e, md)
					if explicitBody {
						views := md.ViewedResult.Views
						if clientView != "" {
							views = []*service.ViewData{{Name: clientView}}
						}
						for _, view := range views {
							representation := &ViewedRepresentationData{
								View:              view.Name,
								ResultAttr:        codegen.Goify(origin, true),
								ClientBody:        clientBodyData,
								ClientDataPointer: clientSSEDataPointer(e, clientRespBody),
								ResultInit:        init,
							}
							if len(serverBodyData) > 0 {
								representation.ServerBody = serverBodyData[0]
							}
							representations = append(representations, representation)
						}
					} else {
						if clientView != "" {
							representation := &ViewedRepresentationData{
								View:              clientView,
								ResultAttr:        codegen.Goify(origin, true),
								ClientBody:        clientBodyData,
								ClientDataPointer: clientSSEDataPointer(e, clientRespBody),
								ResultInit:        init,
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
							body := effectiveClientResponseBodyForView(respBody, viewName, e)
							clientBody := sds.buildResponseBodyType(
								body, result, e, false, &viewName, sd, nil, resultOwner, bodyOwner,
							)
							if body.Type != expr.Empty {
								policy := jsonBodyPolicy(false, false, true, viewName)
								sd.clientWireTypes.applyNames(body, wireResponseBody, policy)
							}
							resultInit := sds.buildResponseResultInit(
								e, resp, result, body, origin,
								headersData, cookiesData, sd, viewName, clientBody,
							)
							representation := &ViewedRepresentationData{
								View:              viewName,
								ResultAttr:        codegen.Goify(origin, true),
								ClientBody:        clientBody,
								ClientDataPointer: clientSSEDataPointer(e, body),
								ResultInit:        resultInit,
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
					StatusCode:             statusCodeToHTTPConst(resp.StatusCode),
					Description:            resp.Description,
					Headers:                headersData,
					Cookies:                cookiesData,
					ContentType:            resp.ContentType,
					ServerBody:             serverBodyData,
					ClientBody:             clientBodyData,
					ResultInit:             init,
					TagName:                tagName,
					TagValue:               tagVal,
					TagPointer:             tagPtr,
					MustValidate:           mustValidate,
					ResultAttr:             codegen.Goify(origin, true),
					ViewedResult:           md.ViewedResult,
					ViewedRepresentations:  representations,
					SelectClientBodyByView: selectClientBodyByView,
				})
			}
		}
		count := len(responses)
		if defaultResponseIndex < count-1 {
			// Put the default response last so generated encoders test tagged
			// responses in the order written in the design.
			defaultResponse := responses[defaultResponseIndex]
			copy(responses[defaultResponseIndex:], responses[defaultResponseIndex+1:])
			responses[count-1] = defaultResponse
		}
	}
	return responses
}

// buildResponseResultInit builds the data used to write one client result
// function. It uses the name chosen by NewPlans and converts the decoded HTTP
// body, headers, and cookies into the method result.
func (sds *ServicesData) buildResponseResultInit(e *expr.HTTPEndpointExpr, resp *expr.HTTPResponseExpr, result, clientBody *expr.AttributeExpr, origin string, headers []*HeaderData, cookies []*CookieData, sd *ServiceData, view string, bodyType *TypeData) *InitData {
	var (
		svc    = sd.Service
		md     = svc.Method(e.Name())
		svcctx = sds.serviceTypeContext(sd, "client").Enter(result)
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
		clientArgs = make([]*InitArgData, 0, len(headers)+len(cookies)+1)
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
		bodyTypeRef := bodyType.Ref
		if bodyTypeRef == "" {
			bodyTypeRef = bodyType.VarName
		}
		clientArgs = append(clientArgs, &InitArgData{
			Ref: ref,
			AttributeData: &AttributeData{
				Name:    "body",
				VarName: "body",
				TypeRef: bodyTypeRef,
			},
		})
		transformctx := jsonBodyContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
		bodyPolicy := wireTypePolicy{
			pointer:             transformctx.Pointer,
			arrayElementPointer: transformctx.ArrayElementPointer,
			view:                bodyType.View,
		}
		responseUse := wireUnionUse{role: wireResponseBody, view: bodyType.View}
		transformctx.Scope = sd.clientWireTypes.resolverForUse(sd.clientWireTypes.scope, bodyPolicy, responseUse)
		if bodyPolicy.view != "" {
			transformctx.Scope = sd.clientWireTypes.rootResolver(sd.clientWireTypes.scope, bodyPolicy, responseUse, bodyType.declaration)
		}
		transforms := sd.transforms.responses[viewedConstructorKey{endpoint: e, response: resp, view: bodyType.View}]
		converted, helpers, err := sd.clientWireTypes.renderTransform(transforms.clientDecode, clientBody, "body", "v", transformctx, svcctx)
		if err != nil {
			sds.recordLinkError(err)
		}
		code = converted
		sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
	} else if expr.IsArray(result.Type) || expr.IsMap(result.Type) {
		if params := expr.AsObject(e.QueryParams().Type); len(*params) > 0 {
			queryctx := wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
			transforms := sd.transforms.responses[viewedConstructorKey{endpoint: e, response: resp, view: view}]
			converted, helpers, err := sd.clientWireTypes.renderTransform(transforms.clientDecode, (*params)[0].Attribute, codegen.Goify((*params)[0].Name, false), "v", queryctx, svcctx)
			if err != nil {
				sds.recordLinkError(err)
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
		httpclictx = jsonBodyContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
	)

	data := make(map[string][]*ErrorData)
	for _, v := range e.HTTPErrors {
		respBody := expr.DupAtt(sd.bodies.errorResponse(v))
		addMarshalTags(respBody)
		errorAttribute := e.MethodExpr.Error(v.Name).AttributeExpr
		errorOwner := expr.MethodErrorExampleIdentity(e.MethodExpr, v.ErrorExpr)
		bodyOwner := expr.ErrorResponseBodyExampleIdentity(e, v)
		var (
			init   *InitData
			body   = respBody.Type
			origin string
		)
		if values, ok := respBody.Meta["origin:attribute"]; ok {
			origin = values[0]
		}

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
				policy := jsonBodyPolicy(false, false, true, "")
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
				code string
				err  error
			)
			if body != expr.Empty {
				var helpers []*codegen.TransformFunctionData
				transformctx := jsonBodyContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
				transforms := sd.transforms.errors[v]
				code, helpers, err = sd.clientWireTypes.renderTransform(transforms.clientDecode, respBody, "body", "v", transformctx, errctx)
				if err == nil {
					sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
				}
			} else if expr.IsArray(v.Type) || expr.IsMap(v.Type) {
				if params := expr.AsObject(e.QueryParams().Type); len(*params) > 0 {
					var helpers []*codegen.TransformFunctionData
					queryctx := wireHTTPContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
					transforms := sd.transforms.errors[v]
					code, helpers, err = sd.clientWireTypes.renderTransform(transforms.clientDecode, (*params)[0].Attribute, codegen.Goify((*params)[0].Name, false), "v", queryctx, errctx)
					if err == nil {
						sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
					}
				}
			}
			if err != nil {
				sds.recordLinkError(err)
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
				transforms := sd.transforms.errors[v]
				if sbd := sds.buildResponseBodyType(respBody, errorAttribute, e, true, nil, sd, transforms, errorOwner, bodyOwner); sbd != nil {
					serverBodyData = append(serverBodyData, sbd)
				}
				clientBodyData = sds.buildResponseBodyType(respBody, errorAttribute, e, false, nil, sd, nil, errorOwner, bodyOwner)
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
				ResultAttr:   codegen.Goify(origin, true),
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
func (sds *ServicesData) buildRequestBodyType(body, att *expr.AttributeExpr, e *expr.HTTPEndpointExpr, role wireTypeRole, svr bool, sd *ServiceData, sourceOwner, bodyOwner expr.ExampleIdentity) *TypeData {
	if body.Type == expr.Empty {
		return nil
	}
	body = expr.DupAtt(body)
	var (
		name              string
		varname           string
		desc              string
		def               string
		ref               string
		validateDef       string
		nestedValidateDef string
		validateRef       string
		validationTarget  string

		svc     = sd.Service
		catalog = sd.wireTypes(svr)
		policy  = jsonBodyPolicy(true, svr, true, "")
		httpctx = jsonBodyContext(catalog, catalog.scope, true, svr)
		side    = "client"
	)
	if svr {
		side = "server"
	}
	svcctx := sds.serviceTypeContext(sd, side).Enter(att)
	addMarshalTags(body)
	record := catalog.lookupUser(body, role, policy)
	catalog.applyNames(body, role, policy)
	httpctx.Scope = catalog.resolverForUse(catalog.scope, policy, wireUnionUse{role: role})
	layoutResolver := httpctx.Scope.(codegen.GoTypeLayoutResolver)
	bodyLayout, layoutErr := layoutResolver.GoTypeLayout(body, httpctx.LayoutPolicy())
	if layoutErr != nil {
		sds.recordLinkError(layoutErr)
		return nil
	}
	name = body.Type.Name()
	if record != nil {
		name = record.name
	}
	ref = bodyLayout.Ref()

	if ut, ok := body.Type.(expr.UserType); ok {
		varname = record.name
		def = goTypeDefForContext(ut.Attribute(), httpctx)
		desc = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP request body.",
			varname, svc.Name, e.Name())
		if svr {
			// generate validation code for unmarshaled type (server-side).
			validateDef = codegen.ValidationCode(ut.Attribute(), ut, httpctx, true, expr.IsAlias(ut), false, "body")
			if record.needsNestedCall {
				nestedValidateDef = codegen.ValidationCodeWithPathParameter(ut.Attribute(), ut, httpctx, true, expr.IsAlias(ut), false, "body", "path")
			}
			if validateDef != "" {
				validationTarget = "&body"
			}
		}
	} else {
		if svr {
			// Generate validation code first because inline struct validation is removed.
			ctx := jsonBodyContext(catalog, catalog.scope, true, true)
			ctx.Pointer = !expr.IsPrimitive(body.Type)
			ctx.Scope = catalog.resolverForUse(catalog.scope, policy, wireUnionUse{role: role})
			validationSource := "body"
			if origin, ok := body.Meta["origin:attribute"]; ok && !att.IsRequired(origin[0]) && !bodyLayout.ReferenceCanBeNil() {
				validationSource = "*body"
			}
			validateRef = codegen.AttributeValidationCode(body, nil, ctx, true, expr.IsAlias(body.Type), validationSource, "body")
		}
		if svr && expr.IsObject(body.Type) {
			// Body is an explicit object described in the design and in
			// this case the GoTypeRef is an inline struct definition. Keep
			// scalar fields pointer-backed until server validation checks
			// whether each value was present.
			body.Validation = nil
		}
		varname = httpctx.Scope.Ref(body, "")
		desc = body.Description
	}
	var init *InitData
	if !svr && att.Type != expr.Empty && needInit(body.Type) {
		var (
			name        string
			desc        string
			code        string
			origin      string
			err         error
			helpers     []*codegen.TransformFunctionData
			declaration *codegen.NameDeclaration

			sourceVar = "p"
			svc       = sd.Service
		)
		{
			if record != nil {
				declaration = record.constructor
			} else {
				declaration = sd.clientBodyConstructors[clientBodyConstructorKey{endpoint: e, role: role}]
			}
			if declaration == nil {
				panic(fmt.Sprintf("client body constructor for %s.%s was not submitted", svc.Name, e.Name()))
			}
			name = declaration.Name()
			desc = fmt.Sprintf("%s builds the HTTP request body from the payload of the %q endpoint of the %q service.",
				name, e.Name(), svc.Name)
			src := sourceVar
			// If design uses Body("name") syntax then need to use payload attribute
			// to transform.
			if o, ok := body.Meta["origin:attribute"]; ok {
				origin = o[0]
				src += "." + codegen.Goify(origin, true)
			}
			transformctx := jsonBodyContext(catalog, catalog.scope, true, svr)
			transformctx.Scope = catalog.resolverForUse(catalog.scope, policy, wireUnionUse{role: role})
			transforms := sd.transforms.requests[clientBodyConstructorKey{endpoint: e, role: role}]
			code, helpers, err = catalog.renderTransform(transforms.clientEncode, body, src, "body", svcctx, transformctx)
			if err != nil {
				sds.recordLinkError(err)
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
			Declaration:         declaration,
			Name:                name,
			Description:         desc,
			ReturnTypeRef:       ref,
			ReturnTypeAttribute: codegen.Goify(origin, true),
			ClientCode:          code,
			ClientArgs:          []*InitArgData{&arg},
		}
	}
	data := &TypeData{
		Name:              name,
		VarName:           varname,
		Description:       desc,
		Def:               def,
		Ref:               ref,
		Init:              init,
		ValidateDef:       validateDef,
		NestedValidateDef: nestedValidateDef,
		ValidateRef:       validateRef,
		ValidationTarget:  validationTarget,
		Example:           sds.Example(body, bodyOwner),
		attribute:         body,
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
func (sds *ServicesData) buildResponseBodyType(
	body, att *expr.AttributeExpr,
	e *expr.HTTPEndpointExpr,
	svr bool,
	view *string,
	sd *ServiceData,
	transforms *plannedResponseTransforms,
	sourceOwner, bodyOwner expr.ExampleIdentity,
) *TypeData {
	if body.Type == expr.Empty {
		return nil
	}
	body, viewName := prepareResponseWireBody(body, view)
	var (
		name              string
		varname           string
		desc              string
		def               string
		ref               string
		validateDef       string
		nestedValidateDef string
		validateRef       string
		validationTarget  string
		mustInit          bool

		svc  = sd.Service
		side = "client"
	)
	if svr {
		side = "server"
	}
	svcctx := sds.serviceTypeContext(sd, side).Enter(att)
	catalog := sd.wireTypes(svr)
	policy := jsonBodyPolicy(false, svr, !svr, viewName)
	// Add each nested named field before body receives its chosen Go names. This
	// keeps each copied request or response field tied to its own definition.
	topLevel, _ := body.Type.(expr.UserType)
	collectUserTypes(body.Type, func(ut expr.UserType) {
		if topLevel != nil && ut == topLevel {
			return
		}
		if d := sds.attributeTypeData(ut, false, !svr, svr, wireUnionUse{role: wireResponseBody, view: viewName}, sd); d != nil {
			if svr {
				sd.ServerBodyAttributeTypes = append(sd.ServerBodyAttributeTypes, d)
			} else {
				sd.ClientBodyAttributeTypes = append(sd.ClientBodyAttributeTypes, d)
			}
		}
	})
	record := catalog.lookupUser(body, wireResponseBody, policy)
	catalog.applyNames(body, wireResponseBody, policy)
	httpctx := jsonBodyContext(catalog, catalog.scope, false, svr)
	responseUse := wireUnionUse{role: wireResponseBody, view: viewName}
	httpctxPolicy := policy
	httpctxPolicy.view = ""
	httpctx.Scope = catalog.resolverForUse(catalog.scope, httpctxPolicy, responseUse)
	name = body.Type.Name()
	if record != nil {
		name = record.name
		ref = record.ref
	} else {
		ref = httpctx.Scope.Ref(body, "")
	}
	mustInit = att.Type != expr.Empty && (needInit(body.Type) || !svr && needClientResponseInit(att.Type))

	if ut, ok := body.Type.(expr.UserType); ok {
		// response body is a user type.
		varname = record.name
		def = goTypeDefForContext(ut.Attribute(), httpctx)
		desc = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP response body.",
			varname, svc.Name, e.Name())
		if !svr {
			// generate validation code for unmarshaled type (client-side).
			validateDef = codegen.ValidationCode(body, ut, httpctx, true, expr.IsAlias(body.Type), false, "body")
			if record.needsNestedCall {
				nestedValidateDef = codegen.ValidationCodeWithPathParameter(body, ut, httpctx, true, expr.IsAlias(body.Type), false, "body", "path")
			}
			if validateDef != "" {
				target := "&body"
				if expr.IsArray(ut) {
					// result type collection
					target = "body"
				}
				validationTarget = target
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
		httpctx = jsonBodyContext(catalog, catalog.scope, false, true)
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
			// If design uses Body("name") syntax then need to use result attribute
			// to transform.
			if o, ok := body.Meta["origin:attribute"]; ok {
				origin = o[0]
				src += "." + codegen.Goify(origin, true)
			}
			transformctx := jsonBodyContext(catalog, catalog.scope, false, svr)
			transformctx.Scope = catalog.resolverForUse(catalog.scope, policy, responseUse)
			if policy.view != "" {
				transformctx.Scope = catalog.rootResolver(catalog.scope, policy, responseUse, record)
			}
			code, helpers, err = catalog.renderTransform(transforms.serverEncode, body, src, "body", svcctx, transformctx)
			if err != nil {
				sds.recordLinkError(err)
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
		Name:              name,
		VarName:           varname,
		Description:       desc,
		Def:               def,
		Ref:               ref,
		Init:              init,
		ValidateDef:       validateDef,
		NestedValidateDef: nestedValidateDef,
		ValidateRef:       validateRef,
		ValidationTarget:  validationTarget,
		Example:           sds.Example(body, bodyOwner),
		View:              viewName,
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
		_, preserveEmpty := a.Find(el.Name).Meta["sse:last-event-id"]
		headers = append(headers, &HeaderData{
			CanonicalName: http.CanonicalHeaderKey(el.HTTPName),
			PreserveEmpty: preserveEmpty,
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
		layout, err := svcCtx.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(att, svcCtx.LayoutPolicy())
		if err != nil {
			sds.recordLinkError(err)
			return err
		}
		var (
			varn        = scope.Name(codegen.Goify(name, false))
			typeRef     = layout.Ref()
			elemTypeRef string
			ft          = svcAtt.Type
			fieldAtt    = svcAtt

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
		valueTypeRef := typeRef
		if pointer {
			typeRef = layout.RefWithPointer(true)
		}
		fieldName := codegen.GoifyAtt(att, name, true)
		if !expr.IsObject(svcAtt.Type) {
			fieldName = ""
		} else {
			fieldAtt = svcAtt.Find(name)
			ft = fieldAtt.Type
			if kind == pathElement || kind == queryElement {
				fptr = svcAtt.IsPrimitivePointer(name, true)
			} else {
				fptr = svcCtx.IsPrimitivePointer(name, svcAtt)
			}
		}
		fieldPackage := ""
		if _, userType := fieldAtt.Type.(expr.UserType); userType {
			fieldPackage = svcCtx.Pkg(fieldAtt)
		}
		fieldTypeRef := svcCtx.Scope.Ref(fieldAtt, fieldPackage)
		validationAttribute := att
		defaultValue := requestElementDefault(svcAtt, name, att)
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
			validationAttribute = &attNoFmt
			validate = codegen.AttributeValidationCode(validationAttribute, nil, svcCtx, required, expr.IsAlias(att.Type), varn+"Raw", name)
		}
		add(&Element{
			HTTPName:    elem,
			Slice:       slice,
			StringSlice: stringSlice,
			AttributeData: &AttributeData{
				Name:           name,
				Description:    att.Description,
				FieldName:      fieldName,
				FieldPointer:   fptr,
				FieldType:      ft,
				ServiceTypeRef: fieldTypeRef,
				VarName:        varn,
				Required:       required,
				Type:           att.Type,
				TypeName:       scope.GoTypeName(att),
				TypeRef:        typeRef,
				ValueTypeRef:   valueTypeRef,
				ElemTypeRef:    elemTypeRef,
				Pointer:        pointer,
				Validate:       validate,
				CLIPlan: cli.NewFlagPlan(
					validationAttribute,
					scope.GoTypeName(validationAttribute),
					valueTypeRef,
					cliValidationRenderer(validate != "", validationAttribute, svcCtx, name),
				),
				IsTextUnmarshaler: isText,
				DefaultValue:      defaultValue,
				HasDefault:        defaultValue != nil,
				IsAliased:         expr.IsAlias(att.Type),
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

// cliValidationRenderer returns nil when the transport plan has no checks. A
// non-nil function writes checks for the concrete value parsed from a flag.
func cliValidationRenderer(enabled bool, attribute *expr.AttributeExpr, context *codegen.AttributeContext, name string) func(string) string {
	if !enabled {
		return nil
	}
	valueContext := context.Dup()
	valueContext.Pointer = false
	return func(target string) string {
		return codegen.AttributeValidationCode(attribute, nil, valueContext, true, expr.IsAlias(attribute.Type), target, name)
	}
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
func effectiveClientResponseBodyForView(body *expr.AttributeExpr, view string, endpoint *expr.HTTPEndpointExpr) *expr.AttributeExpr {
	body = expr.DupAtt(body)
	rt, ok := body.Type.(*expr.ResultTypeExpr)
	if !ok {
		return body
	}
	projected, err := expr.Project(rt, view)
	if err != nil {
		panic(err) // bug
	}
	addJSONRPCSSEViewFields(projected, rt, endpoint)
	body.Type = projected
	return body
}

// addJSONRPCSSEViewFields adds values carried by outer SSE lines to a client-only view body.
func addJSONRPCSSEViewFields(projected, result expr.DataType, endpoint *expr.HTTPEndpointExpr) {
	if !endpoint.IsJSONRPC() || endpoint.SSE == nil {
		return
	}
	if endpoint.SSE.IDField == "" && endpoint.SSE.EventField == "" && endpoint.SSE.RetryField == "" {
		return
	}
	resultObject := expr.AsObject(result)
	projectedResult, ok := projected.(*expr.ResultTypeExpr)
	resultType, resultOK := result.(expr.UserType)
	if !ok || !resultOK || resultObject == nil {
		return
	}
	addFields := func(attribute *expr.AttributeExpr) {
		object := expr.AsObject(attribute.Type)
		if object == nil {
			return
		}
		for _, name := range []string{endpoint.SSE.IDField, endpoint.SSE.EventField, endpoint.SSE.RetryField} {
			if name == "" || object.Attribute(name) != nil {
				continue
			}
			object.Set(name, expr.DupAtt(resultObject.Attribute(name)))
			if resultType.Attribute().IsRequired(name) {
				if attribute.Validation == nil {
					attribute.Validation = &expr.ValidationExpr{}
				}
				attribute.Validation.AddRequired(name)
			}
		}
	}
	addFields(projectedResult.Attribute())
	for _, view := range projectedResult.Views {
		addFields(view.AttributeExpr)
	}
}

// clientSSEDataPointer reports whether a configured SSE data field uses the
// pointer layout required by client response validation. Complete primitive
// response bodies remain values.
func clientSSEDataPointer(endpoint *expr.HTTPEndpointExpr, body *expr.AttributeExpr) bool {
	if endpoint.SSE == nil || endpoint.SSE.DataField == "" {
		return false
	}
	object := expr.AsObject(body.Type)
	if object == nil {
		return false
	}
	attribute := object.Attribute(endpoint.SSE.DataField)
	if attribute == nil {
		panic(fmt.Sprintf("SSE data field %q is missing from the client response body", endpoint.SSE.DataField))
	}
	return sseBodyFieldPointer(attribute)
}

// clientResponseViewName returns the response view used by client code
// generation when the design fixes a viewed result to one view. It also
// returns an empty string for unviewed methods and for responses that choose
// among the authored views at runtime.
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
			Name:                   nat.Name,
			KindConst:              record.kindConsts[i],
			Constructor:            record.constructors[i],
			KindDeclaration:        record.kindDecls[i],
			ConstructorDeclaration: record.ctorDecls[i],
			FieldName:              fieldName,
			StorageName:            record.storageNames[i],
			FieldType:              fieldType,
			Nilable:                codegen.IsNilable(nat.Attribute.Type),
			TypeTag:                nat.Name,
		}
	}

	return &service.UnionTypeData{
		Name:            record.name,
		KindName:        record.kindName,
		TypeDeclaration: record.declaration,
		KindDeclaration: record.kind,
		Fields:          fields,
		TypeKey:         u.GetTypeKey(),
		ValueKey:        u.GetValueKey(),
	}
}

// attributeTypeData builds a nested declaration and keeps the HTTP body use
// that owns any union below it.
func (sds *ServicesData) attributeTypeData(ut expr.UserType, req, ptr, server bool, use wireUnionUse, rd *ServiceData) *TypeData {
	if ut == expr.Empty {
		return nil
	}

	var (
		name           string
		desc           string
		validate       string
		nestedValidate string
		validateRef    string

		att     = expr.DupAtt(&expr.AttributeExpr{Type: ut})
		catalog = rd.wireTypes(server)
		policy  = wireTypePolicy{request: req, pointer: ptr, useDefault: hctxUseDefault(req, server), validate: req || !server, arrayElementPointer: req == server}
	)
	ut = att.Type.(expr.UserType)
	record := catalog.lookupUser(att, wireAttribute, policy)
	catalog.applyNamesRecursive(att, wireAttribute, policy, use, make(map[expr.UserType]struct{}))
	hctx := jsonBodyContext(catalog, catalog.scope, req, server)
	hctx.Scope = catalog.resolverForUse(catalog.scope, policy, use)
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
		if record.needsNestedCall {
			nestedValidate = codegen.ValidationCodeWithPathParameter(ut.Attribute(), ut, hctx, true, expr.IsAlias(ut), false, "body", "path")
		}
	}
	validationTarget := ""
	if validate != "" {
		validationTarget = "v"
	}
	return catalog.bind(record, &TypeData{
		Name:              ut.Name(),
		VarName:           name,
		Description:       desc,
		Def:               goTypeDefForContext(ut.Attribute(), hctx),
		Ref:               record.ref,
		ValidateDef:       validate,
		NestedValidateDef: nestedValidate,
		ValidateRef:       validateRef,
		ValidationTarget:  validationTarget,
		Example:           sds.Example(att, expr.UserTypeExampleIdentity(ut)),
	})
}

// recordLinkError keeps the first failed conversion so Plan.Link can return it
// before callers receive files built from incomplete template data.
func (sds *ServicesData) recordLinkError(err error) {
	if sds.linkErr == nil {
		sds.linkErr = err
	}
}

// wireTypes returns the request and response types for the server or client package.
func (sd *ServiceData) wireTypes(server bool) *wireTypeCatalog {
	if server {
		return sd.serverWireTypes
	}
	return sd.clientWireTypes
}

// jsonBodyPolicy describes one generated JSON body. Bodies being decoded keep
// required primitive array elements as pointers until validation rejects null.
func jsonBodyPolicy(request, server, validate bool, view string) wireTypePolicy {
	// A server decodes a request, and a client decodes a response.
	decode := request == server
	return wireTypePolicy{
		request:             request,
		pointer:             decode,
		useDefault:          !decode,
		validate:            validate,
		arrayElementPointer: decode,
		view:                view,
	}
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

// jsonBodyContext uses pointer elements only while decoding a JSON body. This
// lets generated validation reject null before conversion to service values.
func jsonBodyContext(catalog *wireTypeCatalog, scope *codegen.NameScope, request, server bool) *codegen.AttributeContext {
	context := wireHTTPContext(catalog, scope, request, server)
	decode := request == server
	context.ArrayElementPointer = decode
	context.Scope = catalog.resolver(scope, wireTypePolicy{
		request:             request,
		pointer:             context.Pointer,
		useDefault:          context.UseDefault,
		arrayElementPointer: context.ArrayElementPointer,
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
	case *expr.Union:
		return true
	case expr.UserType:
		return true
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// needClientResponseInit reports whether decoding a response requires a
// separate transport value before Goa can return the service result. In
// addition to named types, required primitive array elements need pointers so
// validation can distinguish JSON null from the primitive zero value.
func needClientResponseInit(dt expr.DataType) bool {
	if needInit(dt) {
		return true
	}
	switch actual := dt.(type) {
	case *expr.Array:
		if actual.NonNullableElems && expr.IsPrimitive(actual.ElemType.Type) &&
			!codegen.IsNilable(actual.ElemType.Type) {
			return true
		}
		return needClientResponseInit(actual.ElemType.Type)
	case *expr.Map:
		return needClientResponseInit(actual.KeyType.Type) ||
			needClientResponseInit(actual.ElemType.Type)
	default:
		return false
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

// NeedDialer returns true if at least one method in the defined services
// uses WebSocket for sending payload or result.
func NeedDialer(data []*ServiceData) bool {
	return slices.ContainsFunc(data, HasWebSocket)
}
