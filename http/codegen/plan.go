// This file builds HTTP output in two steps. NewPlans requests every Go package
// name that the output files need. Plan.Link then builds the HTTP and JSON-RPC
// data after the service names are known.
package codegen

import (
	"cmp"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// PlanInput pairs one design with the generated service names chosen for it.
	PlanInput struct {
		// Root contains the HTTP services that Goa will generate.
		Root *expr.RootExpr
		// Service is the service plan created for Root.
		Service *service.Plan
	}

	// Plan records the generated Go declarations for one design and later builds
	// its HTTP files.
	Plan struct {
		root           *expr.RootExpr
		servicePlan    *service.Plan
		generation     *codegen.Generation
		transport      transportKind
		serverPackages map[*expr.HTTPServiceExpr]*codegen.GeneratedPackage
		extensions     map[*expr.HTTPServiceExpr]*serverExtensions
		constructors   map[viewedConstructorKey]*codegen.NameDeclaration
		payloads       map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration
		streams        map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration
		errors         map[*expr.HTTPErrorExpr]*codegen.NameDeclaration
		wireTypes      map[*expr.HTTPServiceExpr]*plannedWireTypes
		symbols        map[*expr.HTTPServiceExpr]*httpSymbols
		servicePaths   map[*expr.HTTPServiceExpr]string
		cliParsers     map[string]*cli.ParserPlan
		fileImports    map[string]*codegen.GeneratedImportPlan
		services       *ServicesData
		viewed         map[viewedMethodKey]*viewedResultPlan
		jsonServices   map[string]*jsonRPCServicePlan
		server         []*codegen.File
		client         []*codegen.File
		serverTypes    []*codegen.File
		clientTypes    []*codegen.File
		paths          []*codegen.File
		clientCLI      []*codegen.File
	}

	// ServerMountPoint describes one route added by a declared server mount.
	// Goa includes it in Server.Mounts so logs and startup output list the added
	// route with the routes defined in the design.
	ServerMountPoint struct {
		// Method is the operation name shown for the route.
		Method string
		// Verb is the HTTP method accepted by the route.
		Verb string
		// Pattern is the path pattern accepted by the route.
		Pattern string
	}

	// ServerMount gives server templates the chosen mount function name and
	// the routes that function adds.
	ServerMount struct {
		// Declaration supplies the generated mount function name.
		Declaration *codegen.NameDeclaration
		// MountPoints lists the routes added by Declaration.
		MountPoints []ServerMountPoint
	}

	// ExamplePlan builds runnable HTTP programs from server data and generated
	// services that came from the same design.
	ExamplePlan struct {
		root      *example.Root
		transport *Plan
	}

	// jsonRPCServicePlan stores the HTTP data copied for the JSON-RPC file writer.
	jsonRPCServicePlan struct {
		data        *ServiceData
		fileImports map[string][]*codegen.ImportSpec
		clientCodec *codegen.File
		serverCodec *codegen.File
	}

	// viewedResultPlan stores the HTTP response data copied for the JSON-RPC file
	// writer.
	viewedResultPlan struct {
		variable        bool
		fixedView       string
		service         *service.ViewedResultTypeData
		representations []viewedRepresentationPlan
	}

	// viewedRepresentationPlan associates one body conversion with the headers
	// and cookies written by the same successful response.
	viewedRepresentationPlan struct {
		data    *ViewedRepresentationData
		headers []*HeaderData
		cookies []*CookieData
	}

	// serverExtensions stores declarations submitted for one HTTP service before
	// Generation.Freeze chooses their Go names. Link copies these values into
	// template data.
	serverExtensions struct {
		handlerWrappers         []*codegen.NameDeclaration
		endpointHandlerWrappers map[*expr.HTTPEndpointExpr][]*codegen.NameDeclaration
		mounts                  []*ServerMount
	}

	// JSONRPCServiceSnapshot holds a separate copy of the HTTP service data used to write
	// JSON-RPC client and server files. Callers may change it without changing
	// the HTTP plan or a later copy.
	JSONRPCServiceSnapshot struct {
		// Service is a copy of the generated Goa service description.
		Service JSONRPCServiceData
		// Endpoints contains the JSON-RPC method data in design order.
		Endpoints []JSONRPCEndpointSnapshot
		// ClientStruct is the client type name kept for existing plugins. Goa
		// copies it after choosing all names. Changing it does not rename generated code.
		//
		// Deprecated: Use ClientStructDeclaration.Name() after planning.
		ClientStruct string
		// ClientStructDeclaration supplies the client type name written in HTTP files.
		ClientStructDeclaration *codegen.NameDeclaration
		// ClientInitDeclaration supplies the client constructor name.
		ClientInitDeclaration *codegen.NameDeclaration
		// ServerStruct is the server type name kept for existing plugins. Goa
		// copies it after choosing all names. Changing it does not rename generated code.
		//
		// Deprecated: Use ServerStructDeclaration.Name() after planning.
		ServerStruct string
		// ServerStructDeclaration supplies the server type name written in HTTP files.
		ServerStructDeclaration *codegen.NameDeclaration
		// ServerInit is the server constructor name kept for existing plugins. Goa
		// copies it after choosing all names. Changing it does not rename generated code.
		//
		// Deprecated: Use ServerInitDeclaration.Name() after planning.
		ServerInit string
		// ServerInitDeclaration supplies the server constructor name written in HTTP files.
		ServerInitDeclaration *codegen.NameDeclaration
		// MountServer is the route mount function name kept for existing plugins.
		// Goa copies it after choosing all names. Changing it does not rename generated code.
		//
		// Deprecated: Use MountServerDeclaration.Name() after planning.
		MountServer string
		// MountServerDeclaration supplies the route mounting function name written in HTTP files.
		MountServerDeclaration *codegen.NameDeclaration
		// ServerService is the generated function that returns the service implementation.
		ServerService string
		fileImports   map[string][]*codegen.ImportSpec
		clientCodec   *codegen.File
		serverCodec   *codegen.File
	}

	// JSONRPCServiceData contains the service names written in JSON-RPC files.
	JSONRPCServiceData struct {
		// Name is the design service name.
		Name string
		// StructName is the exported Go spelling derived from Name.
		StructName string
		// EndpointsDeclaration supplies the service endpoint collection name.
		EndpointsDeclaration *codegen.NameDeclaration
		// MethodNamesDeclaration supplies the service method name list.
		MethodNamesDeclaration *codegen.NameDeclaration
		// PkgName is the import name of the generated service package.
		PkgName string
		// PathName is the generated service directory name.
		PathName string
	}

	// JSONRPCEndpointSnapshot holds a separate copy of the HTTP values that
	// JSON-RPC files read for one service method.
	JSONRPCEndpointSnapshot struct {
		// IsJSONRPC is true because this value describes a JSON-RPC method.
		IsJSONRPC bool
		// IsJSONRPCNotification reports whether every call omits the request
		// ID and receives no JSON-RPC response.
		IsJSONRPCNotification bool
		// JSONRPCRequestID contains the complete request-ID plan. It is nil for
		// explicit notifications.
		JSONRPCRequestID *JSONRPCRequestIDData
		// Method contains the service method names and stream methods used in JSON-RPC files.
		Method JSONRPCMethodData
		// ServiceName is the design service name written in generated errors.
		ServiceName string
		// ServicePkgName is the import name used for service types.
		ServicePkgName string
		// Payload describes the JSON-RPC request. It is nil when the method has no payload.
		Payload *JSONRPCPayloadData
		// Result describes the JSON-RPC result. It is nil when the method has no result.
		Result *JSONRPCResultData
		// Errors lists the designed errors returned by the method.
		Errors []JSONRPCErrorGroupData
		// Routes lists the HTTP paths and verbs accepted by the JSON-RPC server.
		Routes []JSONRPCRouteData
		// RequestInit builds the HTTP request used for a JSON-RPC call.
		RequestInit *InitData
		// EndpointInit is the client method that builds the Goa endpoint.
		EndpointInit string
		// HandlerInit is the handler constructor name kept for existing plugins.
		// Goa copies it after choosing all names. Changing it does not rename generated code.
		//
		// Deprecated: Use HandlerInitDeclaration.Name() after planning.
		HandlerInit string
		// HandlerInitDeclaration supplies the server handler constructor name.
		HandlerInitDeclaration *codegen.NameDeclaration
		// ClientStruct is the client type name kept for existing plugins. Goa
		// copies it after choosing all names. Changing it does not rename generated code.
		//
		// Deprecated: Use ClientStructDeclaration.Name() after planning.
		ClientStruct string
		// ClientStructDeclaration supplies the client type name used by request builders.
		ClientStructDeclaration *codegen.NameDeclaration
		// RequestEncoder is the request encoder name kept for existing plugins.
		// Goa copies it after choosing all names. Changing it does not rename
		// generated code. It is empty when Goa does not generate a request encoder.
		//
		// Deprecated: Use RequestEncoderDeclaration.Name() after planning.
		RequestEncoder string
		// RequestEncoderDeclaration supplies the request encoder name written in HTTP files.
		RequestEncoderDeclaration *codegen.NameDeclaration
		// RequestDecoder is the request decoder name kept for existing plugins.
		// Goa copies it after choosing all names. Changing it does not rename
		// generated code. It is empty when Goa does not generate a request decoder.
		//
		// Deprecated: Use RequestDecoderDeclaration.Name() after planning.
		RequestDecoder string
		// RequestDecoderDeclaration supplies the request decoder name written in HTTP files.
		RequestDecoderDeclaration *codegen.NameDeclaration
		// ResponseDecoder is the response decoder name kept for existing plugins.
		// Goa copies it after choosing all names. Changing it does not rename generated code.
		//
		// Deprecated: Use ResponseDecoderDeclaration.Name() after planning.
		ResponseDecoder string
		// ResponseDecoderDeclaration supplies the response decoder name written in HTTP files.
		ResponseDecoderDeclaration *codegen.NameDeclaration
		// SSE contains event-stream values when the method uses server-sent events.
		SSE *JSONRPCSSEData
	}

	// JSONRPCMethodData contains the service method values written in JSON-RPC files.
	JSONRPCMethodData struct {
		// Name is the design method name.
		Name string
		// VarName is the exported Go method name.
		VarName string
		// Result is the generated service result type name.
		Result string
		// HasMixedResults reports whether the method returns one synchronous type
		// and streams another type.
		HasMixedResults bool
		// Idempotent reports whether the client may retry the same call.
		Idempotent bool
		// Errors lists the retry properties of the method errors.
		Errors []JSONRPCMethodErrorData
		// ViewedResult contains result-view names when the method returns a viewed result.
		ViewedResult *JSONRPCMethodViewedResultData
		// ServerStream contains server stream method names when the method streams.
		ServerStream *JSONRPCStreamData
		// ClientStream contains client stream method names when the method streams.
		ClientStream *JSONRPCStreamData
		// StreamKind identifies which side sends stream values.
		StreamKind expr.StreamKind
		// SkipRequestBodyEncodeDecode reports whether the service reads the raw request body.
		SkipRequestBodyEncodeDecode bool
		// RequestStruct is the service type that carries a raw request body.
		RequestStruct string
	}

	// JSONRPCMethodErrorData contains the two error values used by client retry code.
	JSONRPCMethodErrorData struct {
		// ErrName is the service error name.
		ErrName string
		// Temporary reports whether retrying the call may succeed.
		Temporary bool
	}

	// JSONRPCMethodViewedResultData contains the result-view fields written in method files.
	JSONRPCMethodViewedResultData struct {
		JSONRPCViewedResultData
		// ViewName is the fixed view name. It is empty when each response selects a view.
		ViewName string
	}

	// JSONRPCStreamData contains the stream method names written by JSON-RPC files.
	JSONRPCStreamData struct {
		// Interface is the service stream interface implemented by the generated stream.
		Interface string
		// VarName is the generated stream implementation type name.
		VarName string
		// SendName is the method that sends one value.
		SendName string
		// SendDesc documents SendName.
		SendDesc string
		// SendWithContextName is the send method that accepts a context.
		SendWithContextName string
		// SendWithContextDesc documents SendWithContextName.
		SendWithContextDesc string
		// SendTypeName is the sent service type name.
		SendTypeName string
		// SendTypeRef is the sent service type reference.
		SendTypeRef string
		// RecvName is the method that receives one value.
		RecvName string
		// RecvDesc documents RecvName.
		RecvDesc string
		// RecvWithContextName is the receive method that accepts a context.
		RecvWithContextName string
		// RecvWithContextDesc documents RecvWithContextName.
		RecvWithContextDesc string
		// RecvTypeName is the received service type name.
		RecvTypeName string
		// RecvTypeRef is the received service type reference.
		RecvTypeRef string
		// EndpointStruct is the service type passed to a streaming endpoint.
		EndpointStruct string
		// Kind identifies which side sends stream values.
		Kind expr.StreamKind
	}

	// JSONRPCPayloadData contains the request values read by JSON-RPC files.
	JSONRPCPayloadData struct {
		// Ref is the service payload type reference.
		Ref string
		// Request describes the request body when the payload has one.
		Request *JSONRPCRequestData
		// IDAttribute is the payload field that receives the JSON-RPC request ID.
		IDAttribute string
		// IDAttributeTypeRef is the named string type assigned from the
		// JSON-RPC request ID. It is empty for a plain string field.
		IDAttributeTypeRef string
		// IDAttributeRequired reports whether IDAttribute is a value instead of a pointer.
		IDAttributeRequired bool
		// JSONRPCRequestID contains the typed payload field and generated-client
		// behavior for the request ID.
		JSONRPCRequestID *JSONRPCRequestIDData
		// DecoderReturnValue is returned directly when the server needs no payload constructor.
		DecoderReturnValue string
	}

	// JSONRPCRequestData contains the request body values read by JSON-RPC files.
	JSONRPCRequestData struct {
		// ClientBody describes the request body encoded by the client.
		ClientBody *JSONRPCBodyData
		// ServerBody describes the request body decoded by the server.
		ServerBody *JSONRPCBodyData
		// PayloadInit builds the service payload from decoded request values.
		PayloadInit *InitData
		// Headers contains the HTTP request headers read by shared JSON code.
		Headers []JSONRPCHeaderData
		// Cookies contains the HTTP request cookies read by shared JSON code.
		Cookies []JSONRPCCookieData
		// QueryParams is empty because JSON-RPC parameters are carried in the JSON request.
		QueryParams []any
		// PathParams is empty because every JSON-RPC method uses the service route.
		PathParams []any
		// PayloadAttr is the payload field encoded as the JSON request body.
		PayloadAttr string
		// MustHaveBody reports whether an empty JSON request is invalid.
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
		// MustValidate reports whether decoded request values require validation.
		MustValidate bool
		// Params describes how the request body is placed in JSON-RPC params.
		Params *JSONRPCParamsData
	}

	// JSONRPCParamsData describes how one service value is carried in the
	// JSON-RPC params member.
	JSONRPCParamsData struct {
		// Positional reports whether params is a one-element array containing the
		// service value. Structured values remain objects or arrays directly.
		Positional bool
		// TypeRef is the generated Go type placed in the positional array.
		TypeRef string
		// RejectNull reports whether a null positional value is missing required
		// input.
		RejectNull bool
		// AllowAbsent reports whether the params member may be omitted because
		// the mapped service field is optional.
		AllowAbsent bool
		// OmitAbsent reports whether an absent direct value must omit params.
		// JSON-RPC does not allow params to be null.
		OmitAbsent bool
	}

	// JSONRPCResultData contains the response values read by JSON-RPC files.
	JSONRPCResultData struct {
		// Ref is the service result type reference.
		Ref string
		// Responses contains the successful HTTP responses in design order.
		Responses []JSONRPCResponseData
		// IDAttribute is retained for generator compatibility.
		//
		// Deprecated: JSON-RPC results cannot define an ID field.
		IDAttribute string
		// IDAttributeRequired is retained for generator compatibility.
		//
		// Deprecated: JSON-RPC results cannot define an ID field.
		IDAttributeRequired bool
		// View is the default result view selected by the design.
		View string
	}

	// JSONRPCResponseData contains one HTTP response read by JSON-RPC files.
	JSONRPCResponseData struct {
		// StatusCode is the JSON-RPC code used for a designed error.
		StatusCode string
		// Code is the numeric JSON-RPC code used for a designed error.
		Code int
		// Headers contains fields decoded from HTTP response headers.
		Headers []JSONRPCHeaderData
		// Cookies contains fields decoded from HTTP response cookies.
		Cookies []JSONRPCCookieData
		// ServerBody contains the response bodies written by the server.
		ServerBody []JSONRPCBodyData
		// ClientBody describes the response body read by the client.
		ClientBody *JSONRPCBodyData
		// ResultInit builds the service result or error from decoded response values.
		ResultInit *InitData
		// ResultAttr is the Go field selected by Body("name"). It is empty when
		// the response body uses the complete service result or error.
		ResultAttr string
		// MustValidate reports whether decoded header or cookie values require validation.
		MustValidate bool
	}

	// JSONRPCErrorGroupData contains errors that use the same JSON-RPC code.
	JSONRPCErrorGroupData struct {
		// StatusCode is the JSON-RPC code shared by Errors.
		StatusCode string
		// Errors contains the designed errors for StatusCode.
		Errors []JSONRPCErrorData
	}

	// JSONRPCErrorData contains one designed error and its response conversion.
	JSONRPCErrorData struct {
		// Name is the design error name.
		Name string
		// Ref is the generated service error type reference.
		Ref string
		// Response describes the encoded error data.
		Response JSONRPCResponseData
	}

	// JSONRPCRouteData contains one HTTP path and verb used for JSON-RPC calls.
	JSONRPCRouteData struct {
		// Verb is the uppercase HTTP method.
		Verb string
		// Path is the full request path.
		Path string
	}

	// JSONRPCSSEData contains the event fields read by JSON-RPC stream files.
	JSONRPCSSEData struct {
		// StructDeclaration supplies the server stream type name.
		StructDeclaration *codegen.NameDeclaration
		// ClientInterfaceDeclaration supplies the client stream interface name.
		ClientInterfaceDeclaration *codegen.NameDeclaration
		// ClientStructDeclaration supplies the client stream implementation name.
		ClientStructDeclaration *codegen.NameDeclaration
		// ClientInitDeclaration supplies the client stream constructor name.
		ClientInitDeclaration *codegen.NameDeclaration
		// EventTypeRef is the service result type carried by each event.
		EventTypeRef string
		// EventTypeName is the service result type allocated by the client.
		EventTypeName string
		// EventIsStruct reports whether the service result is an object.
		EventIsStruct bool
		// DataField is the service result field carried by notification params.
		// It is empty when params carry the complete result body.
		DataField string
		// Data describes the value carried by notification params.
		Data SSEValueData
		// IDField is the service result field carried by the SSE id line.
		IDField string
		// ID describes the value carried by the SSE id line.
		ID *SSEValueData
		// ClientIDPointer reports whether the decoded response body stores IDField as a pointer.
		ClientIDPointer bool
		// EventField is the service result field carried by the SSE event line.
		EventField string
		// Event describes the value carried by the SSE event line.
		Event *SSEValueData
		// ClientEventPointer reports whether the decoded response body stores EventField as a pointer.
		ClientEventPointer bool
		// RetryField is the service result field carried by the SSE retry line.
		RetryField string
		// Retry describes the value carried by the SSE retry line.
		Retry *SSEValueData
		// HasResponseBody reports whether Response converts the service result to JSON.
		HasResponseBody bool
		// Response is the successful response used to encode stream events.
		Response *JSONRPCResponseData
		// RequestIDField is the payload field that receives Last-Event-ID.
		RequestIDField string
		// RequestIDPointer reports whether RequestIDField stores a pointer.
		RequestIDPointer bool
		// Params describes how each streamed result is placed in notification
		// params.
		Params *JSONRPCParamsData
		// ClientEventCode converts the decoded response body into the streamed
		// service result when the ordinary and streamed result types differ.
		ClientEventCode string
	}

	// JSONRPCBodyData contains only the JSON body fields read by JSON-RPC files.
	JSONRPCBodyData struct {
		// Declaration supplies the generated body type name. It is nil when the
		// body uses a Go type expression that does not declare a named type.
		Declaration *codegen.NameDeclaration
		// VarName is the generated body type name.
		VarName string
		// Ref is the generated body type reference.
		Ref string
		// ValidateRef is inline validation code run after decoding.
		ValidateRef string
		// ValidatorDeclaration supplies the named validator called after decoding.
		ValidatorDeclaration *codegen.NameDeclaration
		// ValidationTarget is the decoded value passed to ValidatorDeclaration. It
		// is empty when this body does not need a named validator call.
		ValidationTarget string
		// Init converts between the body and the service value.
		Init *InitData
	}

	// JSONRPCElementData contains one header or cookie value read from a response.
	JSONRPCElementData struct {
		// Name is the service attribute name used in errors.
		Name string
		// VarName is the local Go variable name.
		VarName string
		// TypeName is the Goa primitive or array name.
		TypeName string
		// Type is the generated transport type used by shared request templates.
		Type expr.DataType
		// FieldType is the service field type used by shared request templates.
		FieldType expr.DataType
		// ElemTypeName is the Goa name of an array element. It is empty for non-arrays.
		ElemTypeName string
		// ElemTypeRef is the generated Go reference for an array element. It is empty for non-arrays.
		ElemTypeRef string
		// TypeRef is the generated Go type reference.
		TypeRef string
		// ValueTypeRef is TypeRef without the optional pointer.
		ValueTypeRef string
		// Pointer reports whether TypeRef is a pointer.
		Pointer bool
		// FieldName is the service result field that supplies the value on the server.
		FieldName string
		// FieldPointer reports whether FieldName holds a pointer.
		FieldPointer bool
		// IsAliased reports whether FieldName uses a user-defined primitive type.
		IsAliased bool
		// Required reports whether the response must contain the value.
		Required bool
		// DefaultValue is written when the response omits an optional value.
		DefaultValue any
		// HasDefault reports whether DefaultValue was authored, including zero.
		HasDefault bool
		// Validate contains the validation code run after conversion.
		Validate string
		// HTTPName is the header or cookie name sent over HTTP.
		HTTPName string
		// StringSlice reports whether the value is an array of strings.
		StringSlice bool
		// Slice reports whether the value is an array.
		Slice bool
	}

	// JSONRPCHeaderData contains one response header read by JSON-RPC clients.
	JSONRPCHeaderData struct {
		JSONRPCElementData
		// CanonicalName is the standard HTTP spelling, such as "Content-Type".
		CanonicalName string
		// PreserveEmpty reports whether an empty header value is different from
		// an absent header.
		PreserveEmpty bool
	}

	// JSONRPCCookieData contains one response cookie read by JSON-RPC clients.
	JSONRPCCookieData struct {
		JSONRPCElementData
		// MaxAge is the cookie max-age text written to generated code.
		MaxAge string
		// Path is the cookie path written to generated code.
		Path string
		// Domain is the cookie domain written to generated code.
		Domain string
		// Secure reports whether the Secure cookie flag is set.
		Secure bool
		// HTTPOnly reports whether the HttpOnly cookie flag is set.
		HTTPOnly bool
		// SameSite is the cookie SameSite text written to generated code.
		SameSite string
	}

	// ViewedResultSnapshot holds a separate copy of the result views and HTTP response bodies
	// used by one JSON-RPC method.
	ViewedResultSnapshot struct {
		// Variable reports whether each response carries its selected view.
		Variable bool
		// FixedView is the view selected in the generated method when it cannot vary.
		FixedView string
		// Service contains the viewed-result names written in JSON-RPC files.
		Service JSONRPCViewedResultData
		// Representations contains one copied response conversion for each legal view.
		Representations []ViewedRepresentationSnapshot
	}

	// JSONRPCViewedResultData contains the service package names and functions
	// needed to validate and convert one viewed result.
	JSONRPCViewedResultData struct {
		// FullRef is the complete Go reference to the viewed-result type.
		FullRef string
		// VarName is the viewed-result type name without its package.
		VarName string
		// ViewsPkg is the import name of the generated views package.
		ViewsPkg string
		// Validate is the function that validates the viewed result.
		Validate *codegen.NameDeclaration
		// ResultInit converts a viewed result into the service result.
		ResultInit *codegen.NameDeclaration
		// Init converts the service result into a viewed result.
		Init *codegen.NameDeclaration
		// IsCollection reports whether the viewed result is a collection.
		IsCollection bool
	}

	// ViewedRepresentationSnapshot holds copied client and server body data
	// for one legal result view.
	ViewedRepresentationSnapshot struct {
		// View is the result view carried by the response.
		View string
		// ResultAttr is the Go field selected by Body("name"). It is empty when
		// the server converts the complete projected result.
		ResultAttr string
		// ServerBody describes the value encoded by the server. It is nil when a
		// successful response carries only headers or cookies.
		ServerBody *JSONRPCBodyData
		// ClientBody describes the value decoded by the client. It is nil when a
		// successful response carries only headers or cookies.
		ClientBody *JSONRPCBodyData
		// ResultInit describes how the decoded body rebuilds the service result.
		ResultInit InitData
		// Headers contains copied response header mappings.
		Headers []JSONRPCHeaderData
		// Cookies contains copied response cookie mappings.
		Cookies []JSONRPCCookieData
	}

	// viewedMethodKey identifies one service method without joining its names.
	viewedMethodKey struct {
		service string
		method  string
	}

	// viewedConstructorKey identifies one view-specific client result function
	// in the input design.
	viewedConstructorKey struct {
		endpoint *expr.HTTPEndpointExpr
		response *expr.HTTPResponseExpr
		view     string
	}

	// viewedConstructorOrder provides a stable total order for colliding
	// constructor preferences in one generated client package.
	viewedConstructorOrder struct {
		transport string
		api       string
		service   string
		method    string
		status    int
		tagName   string
		tagValue  string
		view      string
		role      string
	}

	// plannedWireTypes stores each copied request and response field with the
	// client or server package that defines it. Plan.Link uses the same copies
	// after Generation.Freeze chooses every generated Go name.
	plannedWireTypes struct {
		bodies                     shapedBodies
		server                     *wireTypeCatalog
		client                     *wireTypeCatalog
		transforms                 plannedWireTransforms
		streamPayloads             map[*expr.HTTPEndpointExpr]*wireTypeRecord
		clientBodyConstructors     map[clientBodyConstructorKey]*codegen.NameDeclaration
		clientBodyConstructorNames map[clientBodyConstructorKey]string
	}

	// plannedWireTransforms retains the exact conversion selected while HTTP
	// request and response shapes are collected.
	plannedWireTransforms struct {
		requests         map[clientBodyConstructorKey]*plannedRequestTransforms
		responses        map[viewedConstructorKey]*plannedResponseTransforms
		errors           map[*expr.HTTPErrorExpr]*plannedResponseTransforms
		streamingResults map[*expr.HTTPEndpointExpr]*plannedResponseTransforms
	}

	// plannedRequestTransforms contains each direction used by request body and
	// streaming payload code.
	plannedRequestTransforms struct {
		clientEncode wireTransformHandle
		serverDecode wireTransformHandle
		clientDecode wireTransformHandle
	}

	// plannedResponseTransforms contains the server encoder and client decoder
	// for one response representation.
	plannedResponseTransforms struct {
		serverEncode       wireTransformHandle
		clientDecode       wireTransformHandle
		clientDecodeDirect bool
	}

	// clientBodyConstructorKey identifies an unnamed request body constructor.
	clientBodyConstructorKey struct {
		endpoint *expr.HTTPEndpointExpr
		role     wireTypeRole
	}

	// transportKind records whether a plan writes HTTP or JSON-RPC files.
	transportKind uint8
)

const (
	httpTransport transportKind = iota + 1
	jsonrpcTransport
)

// NewPlans submits the Go names used by every ordinary HTTP design in inputs.
// All inputs are required so two designs that write the same package resolve
// name conflicts together.
func NewPlans(generation *codegen.Generation, inputs ...PlanInput) ([]*Plan, error) {
	return newPlans(generation, httpTransport, inputs)
}

// NewJSONRPCPlans requests the HTTP body, encoder, and decoder names used by
// every JSON-RPC design in inputs. JSON-RPC writes its files from these plans.
func NewJSONRPCPlans(generation *codegen.Generation, inputs ...PlanInput) ([]*Plan, error) {
	return newPlans(generation, jsonrpcTransport, inputs)
}

// NewExamplePlan returns an example renderer only when examples contains the
// server data copied from transport's service design.
func NewExamplePlan(transport *Plan, examples *example.Plan) (*ExamplePlan, error) {
	root, ok := examples.Root(transport.servicePlan)
	if !ok {
		return nil, fmt.Errorf("HTTP examples require server data created from the same service design")
	}
	if err := planExampleFileImports(transport, nil, root); err != nil {
		return nil, err
	}
	return &ExamplePlan{root: root, transport: transport}, nil
}

// NewCombinedExamplePlan records the imports used by runnable servers that
// mount JSON-RPC and ordinary HTTP services from the same design.
func NewCombinedExamplePlan(transport, application *Plan, examples *example.Plan) (*ExamplePlan, error) {
	root, ok := examples.Root(transport.servicePlan)
	if !ok {
		return nil, fmt.Errorf("HTTP examples require server data created from the same service design")
	}
	if application != nil && application.servicePlan != transport.servicePlan {
		return nil, fmt.Errorf("combined HTTP examples require transports created from the same service design")
	}
	if err := planExampleFileImports(transport, application, root); err != nil {
		return nil, err
	}
	return &ExamplePlan{root: root, transport: transport}, nil
}

// planExampleFileImports records the packages named by runnable transport
// files. application adds ordinary HTTP services to a JSON-RPC server file.
func planExampleFileImports(transport, application *Plan, root *example.Root) error {
	rootPath := path.Dir(transport.generation.GenPkg())
	for _, server := range root.Servers {
		serverOutput, err := transport.generation.ClaimOutputPackage(
			path.Join(rootPath, "cmd", server.Dir),
			path.Join("cmd", server.Dir),
		)
		if err != nil {
			return err
		}
		fixed := []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("net/http"),
			codegen.SimpleImport("net/url"),
			codegen.SimpleImport("os"),
			codegen.SimpleImport("sync"),
			codegen.SimpleImport("time"),
			codegen.GoaNamedImport("http", "goahttp"),
			codegen.SimpleImport("goa.design/clue/debug"),
			codegen.SimpleImport("goa.design/clue/log"),
		}
		var generated []*codegen.ImportSpec
		ordinary := application
		if transport.transport == httpTransport {
			ordinary = transport
		}
		if ordinary != nil {
			services := configuredTransportServices(ordinary, server.Services)
			if slices.ContainsFunc(services, func(service *expr.HTTPServiceExpr) bool {
				return len(httpWebSocketEndpoints(service)) > 0
			}) {
				fixed = append(fixed, codegen.SimpleImport("github.com/gorilla/websocket"))
			}
			imports, err := exampleServerGeneratedImports(ordinary, services)
			if err != nil {
				return err
			}
			generated = append(generated, imports...)
			if slices.ContainsFunc(services, serviceHasMultipartRequest) {
				generated = append(generated, codegen.NewImport(examplePackageImportName(ordinary.root), rootPath))
			}
		}
		if transport != ordinary {
			imports, err := exampleServerGeneratedImports(transport, configuredTransportServices(transport, server.Services))
			if err != nil {
				return err
			}
			generated = append(generated, imports...)
		}
		if err := retainPlannedFileImports(
			transport,
			serverOutput,
			fixed,
			generated,
			nil,
			nil,
			path.Join("cmd", server.Dir, "http.go"),
		); err != nil {
			return err
		}

		services := configuredTransportServices(transport, server.Services)
		if len(services) == 0 {
			continue
		}
		clientOutput, err := transport.generation.ClaimOutputPackage(
			path.Join(rootPath, "cmd", server.Dir+"-cli"),
			path.Join("cmd", server.Dir+"-cli"),
		)
		if err != nil {
			return err
		}
		fixed = []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("io"),
			codegen.SimpleImport("net/http"),
			codegen.SimpleImport("time"),
			codegen.GoaNamedImport("http", "goahttp"),
		}
		if exampleServicesHaveInputStreams(services) {
			fixed = append(fixed, codegen.SimpleImport("errors"))
		}
		if slices.ContainsFunc(services, func(service *expr.HTTPServiceExpr) bool {
			return len(httpWebSocketEndpoints(service)) > 0
		}) {
			fixed = append(fixed, codegen.SimpleImport("github.com/gorilla/websocket"))
		}
		generated = []*codegen.ImportSpec{codegen.NewImport(
			"cli",
			path.Join(transport.generation.GenPkg(), transportDirectory(transport.transport), "cli", server.Dir),
		)}
		hasInterceptors := false
		for _, service := range services {
			if len(service.ServiceExpr.ClientInterceptors) > 0 || exampleServiceHasOutputStream(service) {
				servicePackage, _, err := servicePackagePreferences(transport.servicePlan, service)
				if err != nil {
					return err
				}
				generated = append(generated, servicePackage)
			}
			hasInterceptors = hasInterceptors || len(service.ServiceExpr.ClientInterceptors) > 0
		}
		if hasInterceptors {
			generated = append(generated, codegen.NewImport("interceptors", path.Join(rootPath, "interceptors")))
		}
		if slices.ContainsFunc(services, serviceHasMultipartRequest) {
			generated = append(generated, codegen.NewImport(examplePackageImportName(transport.root), rootPath))
		}
		if err := retainPlannedFileImports(
			transport,
			clientOutput,
			fixed,
			generated,
			nil,
			nil,
			path.Join("cmd", server.Dir+"-cli", transportDirectory(transport.transport)+".go"),
		); err != nil {
			return err
		}
	}
	return nil
}

// configuredTransportServices returns only services mounted by one server and
// implemented by the selected transport.
func configuredTransportServices(plan *Plan, names []string) []*expr.HTTPServiceExpr {
	expressions := transportExpressions(plan.root, plan.transport)
	services := make([]*expr.HTTPServiceExpr, 0, len(names))
	for _, name := range names {
		if service := expressions.Service(name); service != nil {
			services = append(services, service)
		}
	}
	return services
}

// exampleServerGeneratedImports returns service and transport server packages
// named by one runnable server file.
func exampleServerGeneratedImports(plan *Plan, services []*expr.HTTPServiceExpr) ([]*codegen.ImportSpec, error) {
	imports := make([]*codegen.ImportSpec, 0, 2*len(services))
	for _, service := range services {
		servicePackage, _, err := servicePackagePreferences(plan.servicePlan, service)
		if err != nil {
			return nil, err
		}
		preferred := servicePackage.Name + "svr"
		if plan.transport == jsonrpcTransport {
			preferred = servicePackage.Name + "jssvr"
		}
		imports = append(imports,
			servicePackage,
			codegen.NewImport(preferred, path.Join(
				plan.generation.GenPkg(),
				transportDirectory(plan.transport),
				plan.servicePaths[service],
				"server",
			)),
		)
	}
	return imports, nil
}

// exampleServicesHaveInputStreams reports whether the example rejects any
// configured method because it sends values after the call starts.
func exampleServicesHaveInputStreams(services []*expr.HTTPServiceExpr) bool {
	for _, service := range services {
		for _, endpoint := range service.HTTPEndpoints {
			kind := endpoint.MethodExpr.Stream
			if kind == expr.ClientStreamKind || kind == expr.BidirectionalStreamKind {
				return true
			}
		}
	}
	return false
}

// exampleServiceHasOutputStream reports whether the example asserts a direct
// server-stream result to its generated service interface.
func exampleServiceHasOutputStream(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.MethodExpr.Stream == expr.ServerStreamKind && !endpoint.MethodExpr.HasMixedResults() {
			return true
		}
	}
	return false
}

// MatchesHTTP reports whether NewPlans created p for root and servicePlan.
func (p *Plan) MatchesHTTP(root *expr.RootExpr, servicePlan *service.Plan) bool {
	return p.transport == httpTransport && p.root == root && p.servicePlan == servicePlan
}

// MatchesJSONRPC reports whether NewJSONRPCPlans created p for root and
// servicePlan.
func (p *Plan) MatchesJSONRPC(root *expr.RootExpr, servicePlan *service.Plan) bool {
	return p.transport == jsonrpcTransport && p.root == root && p.servicePlan == servicePlan
}

// DeclareServerHandlerWrapper records an exported func(http.Handler)
// http.Handler that wraps each designed endpoint handler, file handler, and
// redirect mounted for service. Wrappers are applied in registration order,
// with the first registered function surrounding the others. Routes added by
// DeclareServerMount are not wrapped automatically.
func (p *Plan) DeclareServerHandlerWrapper(service *expr.HTTPServiceExpr, preferred string, order codegen.PackageNameOrder) (*codegen.NameDeclaration, error) {
	pkg, err := p.serverExtensionPackage(service)
	if err != nil {
		return nil, err
	}
	declaration := codegen.NewPreferredName(codegen.NameFunction, preferred, codegen.ExportedName, order)
	if err := pkg.DeclareName(declaration); err != nil {
		return nil, err
	}
	p.extensions[service].handlerWrappers = append(p.extensions[service].handlerWrappers, declaration)
	return declaration, nil
}

// DeclareServerEndpointHandlerWrapper records an unexported func(http.Handler)
// http.Handler that wraps the designed routes for endpoint. Service wrappers
// surround endpoint wrappers. File handlers and routes added by plugins are not
// affected.
func (p *Plan) DeclareServerEndpointHandlerWrapper(endpoint *expr.HTTPEndpointExpr, preferred string, order codegen.PackageNameOrder) (*codegen.NameDeclaration, error) {
	pkg, service, err := p.serverEndpointExtensionPackage(endpoint)
	if err != nil {
		return nil, err
	}
	declaration := codegen.NewPreferredName(codegen.NameFunction, preferred, codegen.UnexportedName, order)
	if err := pkg.DeclareName(declaration); err != nil {
		return nil, err
	}
	extensions := p.extensions[service]
	extensions.endpointHandlerWrappers[endpoint] = append(extensions.endpointHandlerWrappers[endpoint], declaration)
	return declaration, nil
}

// DeclareServerMount records an exported func(goahttp.Muxer) that adds routes
// to the HTTP server mux. Goa calls the function after mounting routes from the
// design and includes mountPoints in the server's route list.
func (p *Plan) DeclareServerMount(service *expr.HTTPServiceExpr, preferred string, order codegen.PackageNameOrder, mountPoints []ServerMountPoint) (*codegen.NameDeclaration, error) {
	pkg, err := p.serverExtensionPackage(service)
	if err != nil {
		return nil, err
	}
	if len(mountPoints) == 0 {
		return nil, fmt.Errorf("HTTP server mount requires at least one mount point")
	}
	for index, mount := range mountPoints {
		switch {
		case mount.Method == "":
			return nil, fmt.Errorf("HTTP server mount point %d has an empty method", index)
		case mount.Verb == "":
			return nil, fmt.Errorf("HTTP server mount point %d has an empty verb", index)
		case mount.Pattern == "":
			return nil, fmt.Errorf("HTTP server mount point %d has an empty pattern", index)
		}
	}
	declaration := codegen.NewPreferredName(codegen.NameFunction, preferred, codegen.ExportedName, order)
	if err := pkg.DeclareName(declaration); err != nil {
		return nil, err
	}
	p.extensions[service].mounts = append(p.extensions[service].mounts, &ServerMount{
		Declaration: declaration,
		MountPoints: append([]ServerMountPoint(nil), mountPoints...),
	})
	return declaration, nil
}

// Link reads the chosen Go declaration and import names, builds template data
// for each HTTP service once, and builds every file returned by this plan.
func (p *Plan) Link() error {
	if !p.generation.Frozen() {
		return fmt.Errorf("HTTP plan cannot link before generation freeze")
	}
	if p.services != nil {
		return fmt.Errorf("HTTP plan is already linked")
	}
	if err := p.link(); err != nil {
		return err
	}
	return nil
}

// ServerFiles returns the HTTP server files built by Link.
func (p *Plan) ServerFiles() []*codegen.File {
	p.requireLinked()
	return p.server
}

// Service returns the template data built by Link for the supplied HTTP service
// expression. Callers must call Link before reading the service data.
func (p *Plan) Service(service *expr.HTTPServiceExpr) (*ServiceData, bool) {
	p.requireLinked()
	if _, ok := p.extensions[service]; !ok {
		return nil, false
	}
	data := p.services.Get(service.Name())
	return data, data != nil
}

// ClientFiles returns the HTTP client files built by Link.
func (p *Plan) ClientFiles() []*codegen.File {
	p.requireLinked()
	return p.client
}

// ServerTypeFiles returns the server request and response type files built by Link.
func (p *Plan) ServerTypeFiles() []*codegen.File {
	p.requireLinked()
	return p.serverTypes
}

// ClientTypeFiles returns the client request and response type files built by Link.
func (p *Plan) ClientTypeFiles() []*codegen.File {
	p.requireLinked()
	return p.clientTypes
}

// PathFiles returns the URL path helper files built by Link.
func (p *Plan) PathFiles() []*codegen.File {
	p.requireLinked()
	return p.paths
}

// ClientCLIFiles returns the command-line client files built by Link.
func (p *Plan) ClientCLIFiles() []*codegen.File {
	p.requireLinked()
	return p.clientCLI
}

// ServerFiles builds runnable HTTP servers from the copied server data.
func (p *ExamplePlan) ServerFiles() []*codegen.File {
	p.transport.requireLinked()
	return exampleServerFiles(p.root, p.transport.services)
}

// CLIFiles builds runnable HTTP clients from the copied server data.
func (p *ExamplePlan) CLIFiles() []*codegen.File {
	p.transport.requireLinked()
	return exampleCLIFiles(p.root, p.transport.services)
}

// CombinedServerFiles returns new runnable server files containing this plan's
// JSON-RPC services and application's ordinary HTTP services. Pass nil when
// the design has no ordinary HTTP services.
func (p *ExamplePlan) CombinedServerFiles(application *Plan) []*codegen.File {
	p.transport.requireLinked()
	if p.transport.transport != jsonrpcTransport {
		panic("combined example servers require a JSON-RPC HTTP plan")
	}
	var applicationServices *ServicesData
	if application != nil {
		application.requireLinked()
		if application.transport != httpTransport || application.root != p.transport.root || application.servicePlan != p.transport.servicePlan {
			panic("ordinary HTTP and JSON-RPC plans must use the same design root and service plan")
		}
		applicationServices = application.services
	}
	return combinedExampleServerFiles(p.root, p.transport.services, applicationServices)
}

// ViewedResult returns copied HTTP response data for the named method's result
// views. The second result is false when the method does not use result views.
func (p *Plan) ViewedResult(serviceName, methodName string) (ViewedResultSnapshot, bool) {
	p.requireLinked()
	viewed, ok := p.viewed[viewedMethodKey{service: serviceName, method: methodName}]
	if !ok {
		return ViewedResultSnapshot{}, false
	}
	representations := make([]ViewedRepresentationSnapshot, len(viewed.representations))
	for index, planned := range viewed.representations {
		representation := planned.data
		if representation.ResultInit == nil {
			panic("viewed result representation is missing its result constructor")
		}
		representations[index] = ViewedRepresentationSnapshot{
			View:       representation.View,
			ResultAttr: representation.ResultAttr,
			ServerBody: copyJSONRPCBody(representation.ServerBody),
			ClientBody: copyJSONRPCBody(representation.ClientBody),
			ResultInit: *copyInitData(representation.ResultInit),
			Headers:    copyJSONRPCHeaders(planned.headers),
			Cookies:    copyJSONRPCCookies(planned.cookies),
		}
	}
	return ViewedResultSnapshot{
		Variable:        viewed.variable,
		FixedView:       viewed.fixedView,
		Service:         copyJSONRPCViewedResult(viewed.service),
		Representations: representations,
	}, true
}

// JSONRPCService returns copied HTTP information used to write one JSON-RPC
// service. The second result is false when the plan has no service with name.
func (p *Plan) JSONRPCService(name string) (JSONRPCServiceSnapshot, bool) {
	p.requireLinked()
	planned, ok := p.jsonServices[name]
	if !ok {
		return JSONRPCServiceSnapshot{}, false
	}
	endpoints := make([]JSONRPCEndpointSnapshot, len(planned.data.Endpoints))
	for index, endpoint := range planned.data.Endpoints {
		endpoints[index] = copyJSONRPCEndpoint(endpoint)
	}
	fileImports := make(map[string][]*codegen.ImportSpec, len(planned.fileImports))
	for filePath, imports := range planned.fileImports {
		fileImports[filePath] = cloneImportSpecs(imports)
	}
	return JSONRPCServiceSnapshot{
		Service: JSONRPCServiceData{
			Name:                   planned.data.Service.Name,
			StructName:             planned.data.Service.StructName,
			EndpointsDeclaration:   planned.data.Service.EndpointsDeclaration,
			MethodNamesDeclaration: planned.data.Service.MethodNamesDeclaration,
			PkgName:                planned.data.Service.PkgName,
			PathName:               planned.data.Service.PathName,
		},
		Endpoints:               endpoints,
		ClientStruct:            planned.data.ClientStructDeclaration.Name(),
		ClientStructDeclaration: planned.data.ClientStructDeclaration,
		ClientInitDeclaration:   planned.data.ClientInitDeclaration,
		ServerStruct:            planned.data.ServerStructDeclaration.Name(),
		ServerStructDeclaration: planned.data.ServerStructDeclaration,
		ServerInit:              planned.data.ServerInitDeclaration.Name(),
		ServerInitDeclaration:   planned.data.ServerInitDeclaration,
		MountServer:             planned.data.MountServerDeclaration.Name(),
		MountServerDeclaration:  planned.data.MountServerDeclaration,
		ServerService:           planned.data.ServerService,
		fileImports:             fileImports,
		clientCodec:             planned.clientCodec,
		serverCodec:             planned.serverCodec,
	}, true
}

// FileImports returns a new copy of the service-type imports needed by one
// JSON-RPC output file. It rejects paths that this service does not generate.
func (p JSONRPCServiceSnapshot) FileImports(filePath string) []*codegen.ImportSpec {
	imports, ok := p.fileImports[strings.ReplaceAll(filePath, "\\", "/")]
	if !ok {
		panic("JSON-RPC file is not part of this HTTP service plan")
	}
	return cloneImportSpecs(imports)
}

// ClientCodecFile returns a new client encoder and decoder file for this
// service. The JSON-RPC file writer may change the returned file. It returns
// nil when the service needs neither function.
func (p JSONRPCServiceSnapshot) ClientCodecFile() *codegen.File {
	return cloneJSONRPCCodecFile(p.clientCodec)
}

// ServerCodecFile returns a new server encoder and decoder file for this
// service. The JSON-RPC file writer may change the returned file. It returns
// nil when the service needs neither function.
func (p JSONRPCServiceSnapshot) ServerCodecFile() *codegen.File {
	return cloneJSONRPCCodecFile(p.serverCodec)
}

// serverExtensionPackage returns the generated server package that will contain
// service's extension functions. It first checks that this ordinary HTTP plan
// still accepts new declarations.
func (p *Plan) serverExtensionPackage(service *expr.HTTPServiceExpr) (*codegen.GeneratedPackage, error) {
	if err := p.validateServerExtensionLifecycle(); err != nil {
		return nil, err
	}
	if service == nil {
		return nil, fmt.Errorf("HTTP server extension requires a service from this plan")
	}
	pkg, ok := p.serverPackages[service]
	if !ok {
		return nil, fmt.Errorf("HTTP service does not belong to this plan")
	}
	return pkg, nil
}

// serverEndpointExtensionPackage returns the generated server package and
// service that contain endpoint. It first checks that this ordinary HTTP plan
// still accepts new declarations.
func (p *Plan) serverEndpointExtensionPackage(endpoint *expr.HTTPEndpointExpr) (*codegen.GeneratedPackage, *expr.HTTPServiceExpr, error) {
	if err := p.validateServerExtensionLifecycle(); err != nil {
		return nil, nil, err
	}
	if endpoint == nil {
		return nil, nil, fmt.Errorf("HTTP server endpoint wrapper requires an endpoint from this plan")
	}
	for service, symbols := range p.symbols {
		if _, ok := symbols.endpoints[endpoint]; ok {
			return p.serverPackages[service], service, nil
		}
	}
	return nil, nil, fmt.Errorf("HTTP endpoint does not belong to this plan")
}

// validateServerExtensionLifecycle checks that the plan still accepts new Go
// declarations and that it writes ordinary HTTP files.
func (p *Plan) validateServerExtensionLifecycle() error {
	if p.transport != httpTransport {
		return fmt.Errorf("JSON-RPC HTTP plans do not support server extensions")
	}
	if p.services != nil {
		return fmt.Errorf("HTTP server extension cannot be declared after plan linking")
	}
	if p.generation.Frozen() {
		return fmt.Errorf("HTTP server extension cannot be declared after generation freeze")
	}
	return nil
}

// planImports records each import on the generated package that writes the
// reference. NewPlans calls it after those packages have been claimed.
func planImports(generation *codegen.Generation, transport transportKind, plans []*Plan) error {
	for _, candidate := range generation.Roots() {
		design, ok := candidate.(*expr.RootExpr)
		if !ok {
			continue
		}
		expressions := transportExpressions(design, transport)
		if len(expressions.Services) == 0 {
			continue
		}
		plan := planForRoot(plans, design)
		if plan == nil {
			return fmt.Errorf("%s design has no HTTP plan", transportLabel(transport))
		}
		dir := transportDirectory(transport)
		for _, transportService := range expressions.Services {
			pathName := plan.servicePaths[transportService]
			clientPath := path.Join(generation.GenPkg(), dir, pathName, "client")
			serverPath := path.Join(generation.GenPkg(), dir, pathName, "server")
			servicePackage, viewsPackage, err := servicePackagePreferences(plan.servicePlan, transportService)
			if err != nil {
				return err
			}
			for index, outputPackage := range []*codegen.GeneratedPackage{
				generation.Package(clientPath),
				generation.Package(serverPath),
			} {
				side := "server"
				if index == 0 {
					side = "client"
				}
				fileKinds := []httpGeneratedFile{
					{kind: httpCodecFile, name: "encode_decode.go"},
					{kind: httpTypesFile, name: "types.go"},
				}
				if transport == httpTransport {
					fileKinds = append(fileKinds,
						httpGeneratedFile{kind: httpTransportFile, name: side + ".go"},
						httpGeneratedFile{kind: httpPathsFile, name: "paths.go"},
					)
				}
				if index == 0 && httpServiceHasPayloadBuilder(transportService) {
					fileKinds = append(fileKinds, httpGeneratedFile{kind: httpPayloadBuilderFile, name: "cli.go"})
				}
				for _, file := range fileKinds {
					var definitions, references []*expr.AttributeExpr
					fixedImports := httpFixedFileImports(transportService, index == 0, file.kind)
					switch file.kind {
					case httpTypesFile:
						catalog := plan.wireTypes[transportService].server
						if index == 0 {
							catalog = plan.wireTypes[transportService].client
						}
						definitions, references = wireCatalogImportAttributes(catalog)
						fixedImports = append(fixedImports, wireCatalogValidationImports(catalog)...)
						references = append(references, serviceReferenceAttributes(transportService.HTTPEndpoints...)...)
					case httpCodecFile:
						fixedImports = append(fixedImports, httpCodecValidationImports(transportService, index == 0)...)
						references = serviceReferenceAttributes(transportService.HTTPEndpoints...)
					case httpPayloadBuilderFile:
						references = httpCLIPayloadBuilderReferenceAttributes(transportService)
					case httpTransportFile:
						references = httpTransportReferenceAttributes(transportService)
					}
					filePath := path.Join(codegen.Gendir, dir, pathName, side, file.name)
					if err := retainPlannedFileImports(
						plan,
						outputPackage,
						fixedImports,
						httpGeneratedImportPlan(transportService, index == 0, file.kind, servicePackage, viewsPackage),
						definitions,
						references,
						filePath,
					); err != nil {
						return err
					}
				}
				if transport == httpTransport && len(httpWebSocketEndpoints(transportService)) > 0 {
					if err := retainPlannedFileImports(
						plan,
						outputPackage,
						httpFixedFileImports(transportService, index == 0, httpWebSocketFile),
						httpGeneratedImportPlan(transportService, index == 0, httpWebSocketFile, servicePackage, viewsPackage),
						nil,
						serviceReferenceAttributes(httpWebSocketEndpoints(transportService)...),
						path.Join(codegen.Gendir, dir, pathName, side, "websocket.go"),
					); err != nil {
						return err
					}
				}
				sseFile := "sse.go"
				if transport == jsonrpcTransport && index == 0 {
					sseFile = "stream.go"
				}
				if transport == httpTransport && len(httpSSEEndpoints(transportService)) > 0 {
					if err := retainPlannedFileImports(
						plan,
						outputPackage,
						httpFixedFileImports(transportService, index == 0, httpSSEFile),
						httpGeneratedImportPlan(transportService, index == 0, httpSSEFile, servicePackage, viewsPackage),
						nil,
						serviceReferenceAttributes(httpSSEEndpoints(transportService)...),
						path.Join(codegen.Gendir, dir, pathName, side, sseFile),
					); err != nil {
						return err
					}
				}
			}
		}
		if transport == httpTransport {
			var rootOutput *codegen.GeneratedPackage
			for _, transportService := range expressions.Services {
				if !serviceHasMultipartRequest(transportService) {
					continue
				}
				rootPath := path.Dir(generation.GenPkg())
				if rootOutput == nil {
					var err error
					rootOutput, err = generation.ClaimOutputPackage(rootPath, ".")
					if err != nil {
						return err
					}
				}
				servicePackage, _, err := servicePackagePreferences(plan.servicePlan, transportService)
				if err != nil {
					return err
				}
				servicePath := plan.servicePaths[transportService]
				serverPackage := codegen.NewImport(
					servicePackage.Name+"svr",
					path.Join(generation.GenPkg(), "http", servicePath, "server"),
				)
				var multipartEndpoints []*expr.HTTPEndpointExpr
				for _, endpoint := range transportService.HTTPEndpoints {
					if endpoint.MultipartRequest {
						multipartEndpoints = append(multipartEndpoints, endpoint)
					}
				}
				if err := retainPlannedFileImports(
					plan,
					rootOutput,
					[]*codegen.ImportSpec{codegen.SimpleImport("mime/multipart")},
					[]*codegen.ImportSpec{servicePackage, serverPackage},
					nil,
					serviceReferenceAttributes(multipartEndpoints...),
					"multipart.go",
				); err != nil {
					return err
				}
			}
		}
		for _, server := range design.API.Servers {
			serverName := codegen.SnakeCase(codegen.Goify(server.Name, true))
			cliPath := path.Join(generation.GenPkg(), dir, "cli", serverName)
			cliPackage := generation.Package(cliPath)
			var cliGenerated []*codegen.ImportSpec
			var cliServices []*expr.HTTPServiceExpr
			for _, serviceName := range server.Services {
				transportService := expressions.Service(serviceName)
				if transportService == nil {
					continue
				}
				cliServices = append(cliServices, transportService)
				pathName := plan.servicePaths[transportService]
				servicePackage, _, err := servicePackagePreferences(plan.servicePlan, transportService)
				if err != nil {
					return err
				}
				clientImport := codegen.NewImport(
					servicePackage.Name+"c",
					path.Join(generation.GenPkg(), dir, pathName, "client"),
				)
				cliGenerated = append(cliGenerated, clientImport)
				if len(transportService.ServiceExpr.ClientInterceptors) > 0 {
					servicePackage, _, err := plan.servicePlan.ServicePackageImports(transportService.ServiceExpr)
					if err != nil {
						return err
					}
					cliGenerated = append(cliGenerated, servicePackage)
				}
			}
			if err := retainPlannedFileImports(
				plan,
				cliPackage,
				httpCLIParserFixedImports(cliServices...),
				cliGenerated,
				nil,
				nil,
				path.Join(codegen.Gendir, dir, "cli", serverName, "cli.go"),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// planForRoot returns the transport plan that owns one evaluated design.
func planForRoot(plans []*Plan, root *expr.RootExpr) *Plan {
	for _, plan := range plans {
		if plan.root == root {
			return plan
		}
	}
	return nil
}

// servicePackagePreferences returns the service package and optional views
// package recorded for every transport method. All methods in one service must
// agree because their generated files share imports.
func servicePackagePreferences(plan *service.Plan, transportService *expr.HTTPServiceExpr) (*codegen.ImportSpec, *codegen.ImportSpec, error) {
	servicePackage, availableViewsPackage, err := plan.ServicePackageImports(transportService.ServiceExpr)
	if err != nil {
		return nil, nil, err
	}
	var viewsPackage *codegen.ImportSpec
	for _, endpoint := range transportService.HTTPEndpoints {
		methodService, methodViews, err := plan.MethodPackageImports(endpoint.MethodExpr)
		if err != nil {
			return nil, nil, err
		}
		if *servicePackage != *methodService {
			return nil, nil, fmt.Errorf("HTTP service %q methods use different generated service packages", transportService.Name())
		}
		if methodViews == nil {
			continue
		}
		if *availableViewsPackage != *methodViews {
			return nil, nil, fmt.Errorf("HTTP service %q methods use different generated views packages", transportService.Name())
		}
		viewsPackage = availableViewsPackage
	}
	return servicePackage, viewsPackage, nil
}

// retainPlannedFileImports records every package named by one generated file.
// Repeated calls merge imports when several services write the same file.
func retainPlannedFileImports(
	plan *Plan,
	output *codegen.GeneratedPackage,
	fixed, generated []*codegen.ImportSpec,
	definitions, references []*expr.AttributeExpr,
	filePaths ...string,
) error {
	for _, filePath := range filePaths {
		key := filepathKey(filePath)
		imports := plan.fileImports[key]
		if imports == nil {
			imports = codegen.NewGeneratedImportPlan(output)
			plan.fileImports[key] = imports
		}
		if err := imports.Require(fixed...); err != nil {
			return err
		}
		if err := imports.AddGenerated(generated...); err != nil {
			return err
		}
		if err := imports.AddTypeExpressions(definitions...); err != nil {
			return err
		}
		if err := imports.AddRecursiveTypeReferences(references...); err != nil {
			return err
		}
	}
	return nil
}

// examplePackageImportName returns the package name imported by runnable
// examples for the starter service implementations at the module root.
func examplePackageImportName(root *expr.RootExpr) string {
	scope := codegen.NewNameScope()
	for _, service := range root.Services {
		scope.Unique(strings.ToLower(codegen.Goify(service.Name, false)))
	}
	return scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")
}

// serviceHasMultipartRequest reports whether the example package writes a
// multipart callback whose signature uses this service's generated server.
func serviceHasMultipartRequest(service *expr.HTTPServiceExpr) bool {
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.MultipartRequest {
			return true
		}
	}
	return false
}

// newPlans validates the full input set and submits names for every plan.
func newPlans(generation *codegen.Generation, transport transportKind, inputs []PlanInput) ([]*Plan, error) {
	if generation == nil {
		return nil, fmt.Errorf("HTTP plans require a generation")
	}
	if generation.Frozen() {
		return nil, fmt.Errorf("HTTP plans must be collected before generation freeze")
	}
	owned := make(map[*expr.RootExpr]struct{})
	for _, candidate := range generation.Roots() {
		root, ok := candidate.(*expr.RootExpr)
		if ok && len(transportExpressions(root, transport).Services) > 0 {
			owned[root] = struct{}{}
		}
	}
	seen := make(map[*expr.RootExpr]struct{}, len(inputs))
	for _, input := range inputs {
		if input.Root == nil {
			return nil, fmt.Errorf("HTTP plan requires a prepared design root")
		}
		if input.Service == nil {
			return nil, fmt.Errorf("HTTP plan requires a service plan")
		}
		if input.Service.Root() != input.Root {
			return nil, fmt.Errorf("%s root does not match its service plan root", transportLabel(transport))
		}
		if _, ok := owned[input.Root]; !ok {
			return nil, fmt.Errorf("%s root %p is not a transport root owned by generation", transportLabel(transport), input.Root)
		}
		if _, ok := seen[input.Root]; ok {
			return nil, fmt.Errorf("%s root %p is planned more than once", transportLabel(transport), input.Root)
		}
		seen[input.Root] = struct{}{}
	}
	if len(inputs) != len(owned) {
		return nil, fmt.Errorf("%s planning requires all %d transport roots, got %d", transportLabel(transport), len(owned), len(inputs))
	}
	packages := make(map[string]*wireTypeCatalog)
	plans := make([]*Plan, len(inputs))
	for index, input := range inputs {
		plan, err := newPlan(generation, transport, input, packages)
		if err != nil {
			return nil, err
		}
		plans[index] = plan
	}
	if err := planImports(generation, transport, plans); err != nil {
		return nil, err
	}
	for _, catalog := range packages {
		if err := catalog.Declare(); err != nil {
			return nil, err
		}
	}
	for _, plan := range plans {
		for _, serviceTypes := range plan.wireTypes {
			for endpoint, record := range serviceTypes.streamPayloads {
				plan.streams[endpoint] = record.constructor
			}
		}
	}
	return plans, nil
}

// newPlan records one design's HTTP services and submits every function name
// that its generated client and server packages will define.
func newPlan(generation *codegen.Generation, transport transportKind, input PlanInput, packages map[string]*wireTypeCatalog) (*Plan, error) {
	plan := &Plan{
		root:           input.Root,
		servicePlan:    input.Service,
		generation:     generation,
		transport:      transport,
		serverPackages: make(map[*expr.HTTPServiceExpr]*codegen.GeneratedPackage),
		extensions:     make(map[*expr.HTTPServiceExpr]*serverExtensions),
		constructors:   make(map[viewedConstructorKey]*codegen.NameDeclaration),
		payloads:       make(map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration),
		streams:        make(map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration),
		errors:         make(map[*expr.HTTPErrorExpr]*codegen.NameDeclaration),
		wireTypes:      make(map[*expr.HTTPServiceExpr]*plannedWireTypes),
		symbols:        make(map[*expr.HTTPServiceExpr]*httpSymbols),
		servicePaths:   make(map[*expr.HTTPServiceExpr]string),
		cliParsers:     make(map[string]*cli.ParserPlan),
		fileImports:    make(map[string]*codegen.GeneratedImportPlan),
	}
	expressions := transportExpressions(input.Root, transport)
	dir := transportDirectory(transport)
	for _, transportService := range expressions.Services {
		servicePackage, _, err := input.Service.ServicePackageImports(transportService.ServiceExpr)
		if err != nil {
			return nil, err
		}
		servicePath := path.Base(servicePackage.Path)
		plan.servicePaths[transportService] = servicePath
		clientPath := path.Join(generation.GenPkg(), dir, servicePath, "client")
		clientPackage, err := generation.ClaimPackage(clientPath)
		if err != nil {
			return nil, err
		}
		serverPath := path.Join(generation.GenPkg(), dir, servicePath, "server")
		serverPackage, err := generation.ClaimPackage(serverPath)
		if err != nil {
			return nil, err
		}
		plan.serverPackages[transportService] = serverPackage
		plan.extensions[transportService] = &serverExtensions{
			endpointHandlerWrappers: make(map[*expr.HTTPEndpointExpr][]*codegen.NameDeclaration),
		}
		clientCatalog := packages[clientPath]
		if clientCatalog == nil {
			clientCatalog = newWireTypeCatalog(clientPackage)
			packages[clientPath] = clientCatalog
		}
		serverCatalog := packages[serverPath]
		if serverCatalog == nil {
			serverCatalog = newWireTypeCatalog(serverPackage)
			packages[serverPath] = serverCatalog
		}
		planned := &plannedWireTypes{
			server: serverCatalog,
			client: clientCatalog,
			transforms: plannedWireTransforms{
				requests:         make(map[clientBodyConstructorKey]*plannedRequestTransforms),
				responses:        make(map[viewedConstructorKey]*plannedResponseTransforms),
				errors:           make(map[*expr.HTTPErrorExpr]*plannedResponseTransforms),
				streamingResults: make(map[*expr.HTTPEndpointExpr]*plannedResponseTransforms),
			},
			streamPayloads:             make(map[*expr.HTTPEndpointExpr]*wireTypeRecord),
			clientBodyConstructors:     make(map[clientBodyConstructorKey]*codegen.NameDeclaration),
			clientBodyConstructorNames: make(map[clientBodyConstructorKey]string),
		}
		collectPlannedWireTypes(input.Root.API.Name, transportService, planned, input.Service)
		plan.wireTypes[transportService] = planned
		symbols, err := collectHTTPSymbols(plan, transportService, clientPackage, serverPackage)
		if err != nil {
			return nil, err
		}
		plan.symbols[transportService] = symbols
		for _, endpoint := range transportService.HTTPEndpoints {
			order := viewedConstructorOrder{
				transport: dir,
				api:       input.Root.API.Name,
				service:   transportService.Name(),
				method:    endpoint.Name(),
			}
			for _, role := range []wireTypeRole{wireRequestBody, wireStreamPayload} {
				key := clientBodyConstructorKey{endpoint: endpoint, role: role}
				preferred := planned.clientBodyConstructorNames[key]
				if preferred == "" {
					continue
				}
				body := planned.bodies.request(endpoint)
				if role == wireStreamPayload {
					body = planned.bodies.streaming(endpoint)
				}
				preferred = planned.client.releasedCompositeConstructorName(body, jsonBodyPolicy(true, false, false, ""))
				orderRole := "request body"
				if role == wireStreamPayload {
					orderRole = "streaming body"
				}
				declaration, err := declareHTTPConstructor(clientPackage, preferred, order.withRole(orderRole))
				if err != nil {
					return nil, err
				}
				planned.clientBodyConstructors[key] = declaration
			}
			if needInit(endpoint.MethodExpr.Payload.Type) {
				declaration, err := declareHTTPConstructor(serverPackage, endpointPayloadConstructorName(endpoint), order.withRole("payload"))
				if err != nil {
					return nil, err
				}
				plan.payloads[endpoint] = declaration
			}
			if endpoint.UsesWebSocket() && endpoint.MethodExpr.StreamingPayload.Type != expr.Empty && needInit(endpoint.MethodExpr.StreamingPayload.Type) && planned.streamPayloads[endpoint] == nil {
				preferred := "New" + codegen.Goify(endpoint.Name(), true) + codegen.Goify(endpoint.MethodExpr.StreamingPayload.Type.Name(), true)
				declaration, err := declareHTTPConstructor(serverPackage, preferred, order.withRole("streaming payload"))
				if err != nil {
					return nil, err
				}
				plan.streams[endpoint] = declaration
			}
			if needClientResponseInit(endpoint.MethodExpr.Result.Type) {
				resultType, viewed := endpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr)
				var projectedResult *codegen.TypeDeclaration
				if viewed {
					projectedResult, err = input.Service.ProjectedResultDeclaration(endpoint.MethodExpr)
					if err != nil {
						return nil, err
					}
				}
				noTagSeen := false
				for _, response := range endpoint.Responses {
					if response.Tag[0] == "" {
						if noTagSeen {
							continue
						}
						noTagSeen = true
					}
					views := []string{""}
					body := planned.bodies.response(response)
					_, explicitBody := body.Meta["origin:attribute"]
					if viewed && !explicitBody && clientResponseViewNameExpr(endpoint, resultType) == "" && (endpoint.UsesSSE() || endpoint.IsJSONRPC()) {
						views = make([]string, len(resultType.Views))
						for index, view := range resultType.Views {
							views[index] = view.Name
						}
					}
					for _, view := range views {
						key := viewedConstructorKey{endpoint: endpoint, response: response, view: view}
						responseOrder := order.withRole("result")
						responseOrder.status = response.StatusCode
						responseOrder.tagName = response.Tag[0]
						responseOrder.tagValue = response.Tag[1]
						responseOrder.view = view
						declaration, err := declareViewedResultConstructor(clientPackage, endpoint, response, view, projectedResult, responseOrder)
						if err != nil {
							return nil, err
						}
						plan.constructors[key] = declaration
					}
				}
			}
			for _, transportError := range endpoint.HTTPErrors {
				if !needInit(transportError.Type) {
					continue
				}
				errorOrder := order.withRole("error")
				errorOrder.status = transportError.Response.StatusCode
				errorOrder.tagName = transportError.Name
				preferred := "New" + codegen.Goify(endpoint.Name(), true) + codegen.Goify(transportError.ErrorExpr.Name, true)
				declaration, err := declareHTTPConstructor(clientPackage, preferred, errorOrder)
				if err != nil {
					return nil, err
				}
				plan.errors[transportError] = declaration
			}
		}
	}
	for _, server := range input.Root.API.Servers {
		serverPath := path.Join(generation.GenPkg(), dir, "cli", codegen.SnakeCase(codegen.Goify(server.Name, true)))
		serverPackage, err := generation.ClaimPackage(serverPath)
		if err != nil {
			return nil, err
		}
		var commands []cli.CommandDeclarationInput
		for _, serviceName := range server.Services {
			transportService := expressions.Service(serviceName)
			if transportService == nil || len(transportService.HTTPEndpoints) == 0 {
				continue
			}
			command := cli.CommandDeclarationInput{Service: serviceName}
			for _, endpoint := range transportService.HTTPEndpoints {
				command.Methods = append(command.Methods, endpoint.MethodExpr.Name)
				command.NeedsFlagPresence = command.NeedsFlagPresence || httpEndpointNeedsCLIFlagPresence(endpoint)
			}
			commands = append(commands, command)
		}
		parser, err := cli.DeclareParser(serverPackage, dir, input.Root.API.Name, server.Name, commands)
		if err != nil {
			return nil, err
		}
		plan.cliParsers[server.Name] = parser
	}
	return plan, nil
}

// httpEndpointNeedsCLIFlagPresence reports whether the generated command must
// distinguish a missing flag from a flag whose value is empty. Each check uses
// the same authored default that the generated payload builder receives.
func httpEndpointNeedsCLIFlagPresence(endpoint *expr.HTTPEndpointExpr) bool {
	if endpoint.SkipRequestBodyEncodeDecode {
		return true
	}
	if endpoint.MethodExpr.Payload.Type == expr.Empty {
		return false
	}
	if endpoint.Body.Type != expr.Empty {
		origin := endpoint.Body.Meta["origin:attribute"]
		if len(origin) == 0 || endpoint.MethodExpr.Payload.GetDefault(origin[0]) == nil {
			return true
		}
	}
	if !endpoint.PathParams().IsEmpty() {
		return true
	}
	for _, mapped := range []*expr.MappedAttributeExpr{
		endpoint.QueryParams(),
		endpoint.Headers,
		endpoint.Cookies,
	} {
		for _, field := range *expr.AsObject(mapped.Type) {
			if requestElementDefault(endpoint.MethodExpr.Payload, field.Name, field.Attribute) == nil {
				return true
			}
		}
	}
	if policy := jsonRPCRequestIDPolicyFor(endpoint); policy != nil && policy.attribute != nil && policy.defaultValue == nil {
		return true
	}
	for _, requirement := range endpoint.Requirements {
		for _, scheme := range requirement.Schemes {
			if scheme.Kind == expr.BasicAuthKind {
				return true
			}
		}
	}
	return !expr.IsObject(endpoint.MethodExpr.Payload.Type) && endpoint.MethodExpr.Payload.DefaultValue == nil
}

// link reads the generated service names, builds data for every selected service once,
// and stores all files that the public methods on Plan return.
func (p *Plan) link() error {
	serviceData := p.servicePlan.Services()
	if serviceData.Root != p.root {
		return fmt.Errorf("HTTP plan root does not match linked service plan root")
	}
	expressions := transportExpressions(p.root, p.transport)
	services := newServicesData(serviceData, expressions)
	services.jsonrpc = p.transport == jsonrpcTransport
	services.viewedResultConstructors = p.constructors
	services.payloadConstructors = p.payloads
	services.streamConstructors = p.streams
	services.errorConstructors = p.errors
	services.plannedWireTypes = p.wireTypes
	services.plannedSymbols = p.symbols
	services.cliParsers = p.cliParsers
	services.fileImports = make(map[string][]*codegen.ImportSpec, len(p.fileImports))
	services.clientServicePackages = make(map[string]string, len(expressions.Services))
	for filePath, retained := range p.fileImports {
		if err := retained.Link(); err != nil {
			return err
		}
		services.fileImports[filePath] = retained.Imports()
	}
	for _, transportService := range expressions.Services {
		servicePackage, _, err := servicePackagePreferences(p.servicePlan, transportService)
		if err != nil {
			return err
		}
		clientPath := path.Join(p.generation.GenPkg(), transportDirectory(p.transport), p.servicePaths[transportService], "client")
		name := servicePackage.Name
		filePrefix := path.Join(codegen.Gendir, transportDirectory(p.transport), p.servicePaths[transportService], "client") + "/"
		for filePath, retained := range p.fileImports {
			if !strings.HasPrefix(filePath, filePrefix) || !slices.Contains(retained.Paths(), servicePackage.Path) {
				continue
			}
			name = p.generation.Package(clientPath).ImportName(servicePackage.Path)
			break
		}
		services.clientServicePackages[transportService.Name()] = name
	}
	for _, transportService := range services.Expressions.Services {
		if services.ServicesData.Get(transportService.Name()) == nil {
			return fmt.Errorf("HTTP service %q has no linked service model", transportService.Name())
		}
		data := services.analyze(transportService)
		if services.linkErr != nil {
			return services.linkErr
		}
		extensions := p.extensions[transportService]
		data.ServerHandlerWrappers = append([]*codegen.NameDeclaration(nil), extensions.handlerWrappers...)
		for index, endpoint := range data.Endpoints {
			endpoint.ServerHandlerWrappers = combinedHandlerWrappers(extensions, transportService.HTTPEndpoints[index])
		}
		for _, fileServer := range data.FileServers {
			fileServer.ServerHandlerWrappers = append([]*codegen.NameDeclaration(nil), extensions.handlerWrappers...)
		}
		data.ServerMounts = copyServerMounts(extensions.mounts)
		services.HTTPData[transportService.Name()] = data
	}
	for _, planned := range p.wireTypes {
		if err := planned.checkTransformsUsed(); err != nil {
			return err
		}
	}
	p.services = services
	p.viewed = make(map[viewedMethodKey]*viewedResultPlan)
	p.jsonServices = make(map[string]*jsonRPCServicePlan, len(services.HTTPData))
	for serviceName, serviceData := range services.HTTPData {
		transportService := expressions.Service(serviceName)
		if transportService == nil {
			return fmt.Errorf("HTTP service %q has no transport expression", serviceName)
		}
		if len(transportService.HTTPEndpoints) != len(serviceData.Endpoints) {
			return fmt.Errorf("HTTP service %q endpoint analysis does not match its design", serviceName)
		}
		jsonService := &jsonRPCServicePlan{
			data:        serviceData,
			fileImports: make(map[string][]*codegen.ImportSpec),
			clientCodec: clientEncodeDecodeFile(transportService, services),
			serverCodec: serverEncodeDecodeFile(transportService, services),
		}
		if p.transport == jsonrpcTransport {
			jsonService.prepareFileImports(services)
		}
		p.jsonServices[serviceName] = jsonService
		for _, endpoint := range serviceData.Endpoints {
			if endpoint.Method.ViewedResult == nil || endpoint.SSE == nil && !endpoint.IsJSONRPC {
				continue
			}
			var representations []viewedRepresentationPlan
			for _, response := range endpoint.Result.Responses {
				for _, representation := range response.ViewedRepresentations {
					representations = append(representations, viewedRepresentationPlan{
						data:    representation,
						headers: response.Headers,
						cookies: response.Cookies,
					})
				}
			}
			variable := endpoint.Method.ViewedResult.ViewName == ""
			if len(representations) == 0 {
				return fmt.Errorf("HTTP viewed method %q has no response representations", endpoint.Method.Name)
			}
			p.viewed[viewedMethodKey{service: serviceName, method: endpoint.Method.Name}] = &viewedResultPlan{
				variable:        variable,
				fixedView:       endpoint.Method.ViewedResult.ViewName,
				service:         endpoint.Method.ViewedResult,
				representations: representations,
			}
		}
	}
	if p.transport == httpTransport {
		p.server = serverFiles(services)
		p.client = clientFiles(services)
	}
	p.serverTypes = serverTypeFiles(services)
	p.clientTypes = clientTypeFiles(services)
	p.paths = pathFiles(services)
	p.clientCLI = clientCLIFiles(services)
	return nil
}

// checkTransformsUsed verifies that every conversion retained by this service
// plan was written exactly once while its template data was built.
func (p *plannedWireTypes) checkTransformsUsed() error {
	checkRequest := func(transforms *plannedRequestTransforms) error {
		for _, use := range []struct {
			catalog *wireTypeCatalog
			handle  wireTransformHandle
		}{
			{p.client, transforms.clientEncode},
			{p.server, transforms.serverDecode},
			{p.client, transforms.clientDecode},
		} {
			if err := use.catalog.checkTransformUsed(use.handle); err != nil {
				return err
			}
		}
		return nil
	}
	checkResponse := func(transforms *plannedResponseTransforms) error {
		for _, use := range []struct {
			catalog *wireTypeCatalog
			handle  wireTransformHandle
		}{
			{p.server, transforms.serverEncode},
			{p.client, transforms.clientDecode},
		} {
			if err := use.catalog.checkTransformUsed(use.handle); err != nil {
				return err
			}
		}
		return nil
	}
	for _, transforms := range p.transforms.requests {
		if err := checkRequest(transforms); err != nil {
			return err
		}
	}
	for _, transforms := range p.transforms.responses {
		if err := checkResponse(transforms); err != nil {
			return err
		}
	}
	for _, transforms := range p.transforms.errors {
		if err := checkResponse(transforms); err != nil {
			return err
		}
	}
	for _, transforms := range p.transforms.streamingResults {
		if err := checkResponse(transforms); err != nil {
			return err
		}
	}
	return nil
}

// combinedHandlerWrappers copies the service wrappers followed by the wrappers
// declared for one endpoint. Templates nest the first entry outermost.
func combinedHandlerWrappers(extensions *serverExtensions, endpoint *expr.HTTPEndpointExpr) []*codegen.NameDeclaration {
	wrappers := make([]*codegen.NameDeclaration, 0, len(extensions.handlerWrappers)+len(extensions.endpointHandlerWrappers[endpoint]))
	wrappers = append(wrappers, extensions.handlerWrappers...)
	return append(wrappers, extensions.endpointHandlerWrappers[endpoint]...)
}

// copyServerMounts gives render code its own mount functions and route entries.
func copyServerMounts(source []*ServerMount) []*ServerMount {
	result := make([]*ServerMount, len(source))
	for index, mount := range source {
		result[index] = &ServerMount{
			Declaration: mount.Declaration,
			MountPoints: append([]ServerMountPoint(nil), mount.MountPoints...),
		}
	}
	return result
}

// transportExpressions returns the HTTP or JSON-RPC designs requested
// by the caller.
func transportExpressions(root *expr.RootExpr, transport transportKind) *expr.HTTPExpr {
	if transport == jsonrpcTransport {
		return &root.API.JSONRPC.HTTPExpr
	}
	return root.API.HTTP
}

// transportDirectory returns the output directory for HTTP or JSON-RPC files.
func transportDirectory(transport transportKind) string {
	if transport == jsonrpcTransport {
		return "jsonrpc"
	}
	return "http"
}

// transportLabel returns "HTTP" or "JSON-RPC" for error messages.
func transportLabel(transport transportKind) string {
	if transport == jsonrpcTransport {
		return "JSON-RPC"
	}
	return "HTTP"
}

// requireLinked rejects file and service access before Link builds them.
func (p *Plan) requireLinked() {
	if p.services == nil {
		panic("HTTP render model requested before plan linking")
	}
}

// prepareFileImports copies the package paths collected for each JSON-RPC file
// before generation names were frozen.
func (p *jsonRPCServicePlan) prepareFileImports(services *ServicesData) {
	servicePath := p.data.Service.PathName
	clientPath := path.Join(codegen.Gendir, "jsonrpc", servicePath, "client", "client.go")
	serverPath := path.Join(codegen.Gendir, "jsonrpc", servicePath, "server", "server.go")
	p.fileImports[clientPath] = cloneImportSpecs(services.fileImports[filepathKey(clientPath)])
	p.fileImports[serverPath] = cloneImportSpecs(services.fileImports[filepathKey(serverPath)])
	if p.clientCodec != nil {
		p.fileImports[filepathKey(p.clientCodec.Path)] = cloneImportSpecs(services.fileImports[filepathKey(p.clientCodec.Path)])
	}
	if p.serverCodec != nil {
		p.fileImports[filepathKey(p.serverCodec.Path)] = cloneImportSpecs(services.fileImports[filepathKey(p.serverCodec.Path)])
	}
	hasSSE := false
	for _, endpoint := range p.data.Endpoints {
		if endpoint.SSE != nil {
			hasSSE = true
			break
		}
	}
	if hasSSE {
		clientStreamPath := path.Join(codegen.Gendir, "jsonrpc", servicePath, "client", "stream.go")
		serverStreamPath := path.Join(codegen.Gendir, "jsonrpc", servicePath, "server", "sse.go")
		p.fileImports[clientStreamPath] = cloneImportSpecs(services.fileImports[filepathKey(clientStreamPath)])
		p.fileImports[serverStreamPath] = cloneImportSpecs(services.fileImports[filepathKey(serverStreamPath)])
	}
}

// cloneImportSpecs copies an import list and each import value so a caller can
// change both without changing the list stored by the HTTP plan.
func cloneImportSpecs(source []*codegen.ImportSpec) []*codegen.ImportSpec {
	result := make([]*codegen.ImportSpec, len(source))
	for index, spec := range source {
		copy := *spec
		result[index] = &copy
	}
	return result
}

// declareViewedResultConstructor records the function that converts one HTTP
// response into a service result. Released overlapping method and result names
// use the exact projected view type selected by the service plan.
func declareViewedResultConstructor(
	pkg *codegen.GeneratedPackage,
	endpoint *expr.HTTPEndpointExpr,
	response *expr.HTTPResponseExpr,
	view string,
	projectedResult *codegen.TypeDeclaration,
	order viewedConstructorOrder,
) (*codegen.NameDeclaration, error) {
	status := codegen.Goify(http.StatusText(response.StatusCode), true)
	method := codegen.Goify(endpoint.Name(), true)
	result := codegen.Goify(releasedMethodTypeName(endpoint.MethodExpr.Result, "Result"), true)
	if view == "" && projectedResult != nil && strings.HasPrefix(result, method) {
		return pkg.DeclareDependentName(codegen.NameFunction, projectedResult.Declaration(), "New", status, order)
	}
	return declareHTTPConstructor(pkg, viewedResultConstructorName(endpoint, response, view), order)
}

// viewedResultConstructorName returns the preferred constructor spelling for
// results whose released name does not depend on a projected view type.
func viewedResultConstructorName(endpoint *expr.HTTPEndpointExpr, response *expr.HTTPResponseExpr, view string) string {
	status := codegen.Goify(http.StatusText(response.StatusCode), true)
	if view != "" {
		return "New" + codegen.Goify(endpoint.Name(), true) + "Result" + codegen.Goify(view, true) + status
	}
	return releasedMethodTypeConstructorName(endpoint.Name(), releasedMethodTypeName(endpoint.MethodExpr.Result, "Result"), "Result") + status
}

// endpointPayloadConstructorName returns the preferred server function name
// that builds one method payload from its HTTP request values.
func endpointPayloadConstructorName(endpoint *expr.HTTPEndpointExpr) string {
	method := codegen.Goify(endpoint.Name(), true)
	payload := codegen.Goify(releasedMethodTypeName(endpoint.MethodExpr.Payload, "Payload"), true)
	if strings.HasPrefix(payload, method) {
		return "New" + payload
	}
	return "New" + method + payload
}

// releasedMethodTypeName returns the service-side spelling used before HTTP
// constructors were planned separately. It specializes arrays and maps from
// their element types, such as ElemType and MapKeyTypeElemType.
func releasedMethodTypeName(attribute *expr.AttributeExpr, role string) string {
	name := codegen.NewNameScope().GoTypeName(attribute)
	if name == "" {
		return role
	}
	return name
}

// releasedMethodTypeConstructorName joins a method and its service type while
// avoiding a repeated type stem. For example, FetchCustomer and Customer
// produce NewFetchCustomerResult.
func releasedMethodTypeConstructorName(method, typeName, role string) string {
	method = codegen.Goify(method, true)
	typeName = codegen.Goify(typeName, true)
	stem := strings.TrimSuffix(typeName, role)
	if stem != typeName && stem != "" && strings.HasSuffix(method, stem) {
		return "New" + method + role
	}
	return "New" + method + typeName
}

// declareHTTPConstructor submits one constructor name to the generated package
// that will contain both its definition and calls.
func declareHTTPConstructor(pkg *codegen.GeneratedPackage, preferred string, order viewedConstructorOrder) (*codegen.NameDeclaration, error) {
	declaration := codegen.NewPreferredName(codegen.NameFunction, preferred, codegen.ExportedName, order)
	if err := pkg.DeclareName(declaration); err != nil {
		return nil, err
	}
	return declaration, nil
}

// withRole returns an ordering value for one kind of endpoint constructor.
func (o viewedConstructorOrder) withRole(role string) viewedConstructorOrder {
	o.role = role
	return o
}

// ComparePackageName orders view constructors by design service, method,
// response status, and view so input iteration order cannot change names.
func (o viewedConstructorOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(viewedConstructorOrder)
	for _, compared := range []int{
		cmp.Compare(o.transport, right.transport),
		cmp.Compare(o.api, right.api),
		cmp.Compare(o.service, right.service),
		cmp.Compare(o.method, right.method),
		cmp.Compare(o.role, right.role),
		cmp.Compare(o.status, right.status),
		cmp.Compare(o.tagName, right.tagName),
		cmp.Compare(o.tagValue, right.tagValue),
		cmp.Compare(o.view, right.view),
	} {
		if compared != 0 {
			return compared
		}
	}
	return 0
}
