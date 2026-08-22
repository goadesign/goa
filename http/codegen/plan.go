// This file builds HTTP output in two steps. NewPlans requests every Go package
// name that the output files need. Plan.Link then builds the HTTP and JSON-RPC
// data after the service names are known.
package codegen

import (
	"cmp"
	"fmt"
	"net/http"
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
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

	// Plan records package names for one design and later builds its HTTP files.
	Plan struct {
		root         *expr.RootExpr
		servicePlan  *service.Plan
		generation   *codegen.Generation
		transport    transportKind
		constructors map[viewedConstructorKey]*codegen.NameDeclaration
		payloads     map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration
		streams      map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration
		errors       map[*expr.HTTPErrorExpr]*codegen.NameDeclaration
		wireTypes    map[*expr.HTTPServiceExpr]*plannedWireTypes
		symbols      map[*expr.HTTPServiceExpr]*httpSymbols
		cliParsers   map[*expr.ServerExpr]*cli.ParserPlan
		services     *ServicesData
		viewed       map[viewedMethodKey]*viewedResultPlan
		jsonServices map[string]*jsonRPCServicePlan
		server       []*codegen.File
		client       []*codegen.File
		serverTypes  []*codegen.File
		clientTypes  []*codegen.File
		paths        []*codegen.File
		clientCLI    []*codegen.File
		example      []*codegen.File
		exampleCLI   []*codegen.File
	}

	// jsonRPCServicePlan stores the HTTP data copied for the JSON-RPC file writer.
	jsonRPCServicePlan struct {
		data        *ServiceData
		services    *ServicesData
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

	// JSONRPCServiceSnapshot holds a separate copy of the HTTP service data used to write
	// JSON-RPC client and server files. Callers may change it without changing
	// the HTTP plan or a later copy.
	JSONRPCServiceSnapshot struct {
		// Service is a copy of the generated Goa service description.
		Service JSONRPCServiceData
		// Endpoints contains the JSON-RPC method data in design order.
		Endpoints []JSONRPCEndpointSnapshot
		// ClientStructDeclaration supplies the client type name written in HTTP files.
		ClientStructDeclaration *codegen.NameDeclaration
		// ClientInitDeclaration supplies the client constructor name.
		ClientInitDeclaration *codegen.NameDeclaration
		// ServerStructDeclaration supplies the server type name written in HTTP files.
		ServerStructDeclaration *codegen.NameDeclaration
		// ServerInitDeclaration supplies the server constructor name written in HTTP files.
		ServerInitDeclaration *codegen.NameDeclaration
		// MountServerDeclaration supplies the route mounting function name written in HTTP files.
		MountServerDeclaration *codegen.NameDeclaration
		// ServerService is the generated function that returns the service implementation.
		ServerService string
		serviceImport *codegen.ImportSpec
		viewImport    *codegen.ImportSpec
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
		// StreamDeclaration supplies the shared service stream name.
		StreamDeclaration *codegen.NameDeclaration
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
		// HandlerInitDeclaration supplies the server handler constructor name.
		HandlerInitDeclaration *codegen.NameDeclaration
		// ClientStructDeclaration supplies the client type name used by request builders.
		ClientStructDeclaration *codegen.NameDeclaration
		// RequestEncoderDeclaration supplies the request encoder name written in HTTP files.
		RequestEncoderDeclaration *codegen.NameDeclaration
		// RequestDecoderDeclaration supplies the request decoder name written in HTTP files.
		RequestDecoderDeclaration *codegen.NameDeclaration
		// ResponseDecoderDeclaration supplies the response decoder name written in HTTP files.
		ResponseDecoderDeclaration *codegen.NameDeclaration
		// SSE contains event-stream values when the method uses server-sent events.
		SSE *JSONRPCSSEData
		// ClientWebSocket contains client stream names when the method uses WebSocket.
		ClientWebSocket *JSONRPCWebSocketData
		// ServerWebSocket contains server stream names when the method uses WebSocket.
		ServerWebSocket *JSONRPCWebSocketData
	}

	// JSONRPCMethodData contains the service method values written in JSON-RPC files.
	JSONRPCMethodData struct {
		// Name is the design method name.
		Name string
		// VarName is the exported Go method name.
		VarName string
		// EventDeclaration supplies the service event interface name used by
		// server-sent-event streams.
		EventDeclaration *codegen.NameDeclaration
		// Result is the generated service result type name.
		Result string
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
		// IDAttributeRequired reports whether IDAttribute is a value instead of a pointer.
		IDAttributeRequired bool
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
		// PayloadTypeName is the Goa name for the payload type.
		PayloadTypeName string
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
		// MustValidate reports whether decoded request values require validation.
		MustValidate bool
	}

	// JSONRPCResultData contains the response values read by JSON-RPC files.
	JSONRPCResultData struct {
		// Ref is the service result type reference.
		Ref string
		// Responses contains the successful HTTP responses in design order.
		Responses []JSONRPCResponseData
		// IDAttribute is the result field that supplies the JSON-RPC response ID.
		IDAttribute string
		// IDAttributeRequired reports whether IDAttribute is a value instead of a pointer.
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
		// RequestIDField is the payload field that receives Last-Event-ID.
		RequestIDField string
	}

	// JSONRPCWebSocketData contains the stream names read by JSON-RPC WebSocket files.
	JSONRPCWebSocketData struct {
		// VarName is the generated stream implementation type name.
		VarName string
		// VarDeclaration supplies the stream implementation type name.
		VarDeclaration *codegen.NameDeclaration
		// SendName is the method that sends a stream value.
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
		// RecvName is the method that receives a stream value.
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
	}

	// JSONRPCBodyData contains only the JSON body fields read by JSON-RPC files.
	JSONRPCBodyData struct {
		// VarName is the generated body type name.
		VarName string
		// Ref is the generated body type reference.
		Ref string
		// ValidateRef is the validation statement run after decoding.
		ValidateRef string
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
		// ElemTypeName is the Goa name of an array element. It is empty for non-arrays.
		ElemTypeName string
		// ElemTypeRef is the generated Go reference for an array element. It is empty for non-arrays.
		ElemTypeRef string
		// TypeRef is the generated Go type reference.
		TypeRef string
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
	// after Goa assigns every generated package name.
	plannedWireTypes struct {
		bodies         shapedBodies
		server         *wireTypeCatalog
		client         *wireTypeCatalog
		streamPayloads map[*expr.HTTPEndpointExpr]*wireTypeRecord
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

// MatchesHTTP reports whether NewPlans created p for root and servicePlan.
func (p *Plan) MatchesHTTP(root *expr.RootExpr, servicePlan *service.Plan) bool {
	return p.transport == httpTransport && p.root == root && p.servicePlan == servicePlan
}

// MatchesJSONRPC reports whether NewJSONRPCPlans created p for root and
// servicePlan.
func (p *Plan) MatchesJSONRPC(root *expr.RootExpr, servicePlan *service.Plan) bool {
	return p.transport == jsonrpcTransport && p.root == root && p.servicePlan == servicePlan
}

// Link reads the assigned package names, builds data for each HTTP service once, and
// builds every file returned by this plan.
func (p *Plan) Link() error {
	if !p.generation.Frozen() {
		return fmt.Errorf("HTTP plan cannot link before generation freeze")
	}
	if p.services != nil {
		return fmt.Errorf("HTTP plan is already linked")
	}
	return p.link()
}

// ServerFiles returns the HTTP server files built by Link.
func (p *Plan) ServerFiles() []*codegen.File {
	p.requireLinked()
	return p.server
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

// ExampleServerFiles returns the runnable HTTP server files built by Link.
func (p *Plan) ExampleServerFiles() []*codegen.File {
	p.requireLinked()
	return p.example
}

// ExampleCLIFiles returns the runnable HTTP client files built by Link.
func (p *Plan) ExampleCLIFiles() []*codegen.File {
	p.requireLinked()
	return p.exampleCLI
}

// CombinedExampleServerFiles returns new runnable server files containing this
// plan's JSON-RPC services and application's ordinary HTTP services. Pass nil
// when the design has no ordinary HTTP services.
func (p *Plan) CombinedExampleServerFiles(application *Plan) []*codegen.File {
	p.requireLinked()
	if p.transport != jsonrpcTransport {
		panic("combined example servers require a JSON-RPC HTTP plan")
	}
	var applicationServices *ServicesData
	if application != nil {
		application.requireLinked()
		if application.transport != httpTransport || application.root != p.root || application.servicePlan != p.servicePlan {
			panic("ordinary HTTP and JSON-RPC plans must use the same design root and service plan")
		}
		applicationServices = application.services
	}
	return combinedExampleServerFiles(p.services, applicationServices)
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
	var viewImport *codegen.ImportSpec
	if serviceHasViewedResult(planned.data, nil) {
		viewImport = planned.services.ViewImport(planned.data.Service.Name)
	}
	return JSONRPCServiceSnapshot{
		Service: JSONRPCServiceData{
			Name:                   planned.data.Service.Name,
			StructName:             planned.data.Service.StructName,
			EndpointsDeclaration:   planned.data.Service.EndpointsDeclaration,
			StreamDeclaration:      planned.data.Service.StreamDeclaration,
			MethodNamesDeclaration: planned.data.Service.MethodNamesDeclaration,
			PkgName:                planned.data.Service.PkgName,
			PathName:               planned.data.Service.PathName,
		},
		Endpoints:               endpoints,
		ClientStructDeclaration: planned.data.ClientStructDeclaration,
		ClientInitDeclaration:   planned.data.ClientInitDeclaration,
		ServerStructDeclaration: planned.data.ServerStructDeclaration,
		ServerInitDeclaration:   planned.data.ServerInitDeclaration,
		MountServerDeclaration:  planned.data.MountServerDeclaration,
		ServerService:           planned.data.ServerService,
		serviceImport:           cloneImportSpec(planned.services.ServiceImport(planned.data.Service.Name)),
		viewImport:              cloneImportSpec(viewImport),
		fileImports:             fileImports,
		clientCodec:             planned.clientCodec,
		serverCodec:             planned.serverCodec,
	}, true
}

// ServiceImport returns the import for the generated Goa service package.
func (p JSONRPCServiceSnapshot) ServiceImport() *codegen.ImportSpec {
	return cloneImportSpec(p.serviceImport)
}

// ViewImport returns the import for the generated result-view package.
func (p JSONRPCServiceSnapshot) ViewImport() *codegen.ImportSpec {
	if p.viewImport == nil {
		panic("JSON-RPC service does not use result views")
	}
	return cloneImportSpec(p.viewImport)
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

// planImports requests every import name written directly in an HTTP file.
// This happens before generated service packages receive their import names.
func planImports(generation *codegen.Generation, transport transportKind) error {
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bufio"),
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("flag"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("mime/multipart"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("path"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("sync"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("github.com/google/uuid"),
		codegen.SimpleImport("github.com/gorilla/websocket"),
		codegen.SimpleImport("goa.design/clue/debug"),
		codegen.SimpleImport("goa.design/clue/log"),
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		codegen.GoaImport("middleware"),
	}
	for _, spec := range imports {
		if err := generation.RequireImport(spec); err != nil {
			return err
		}
	}
	for _, root := range generation.Roots() {
		design, ok := root.(*expr.RootExpr)
		if !ok {
			continue
		}
		expressions := transportExpressions(design, transport)
		dir := transportDirectory(transport)
		for _, service := range expressions.Services {
			pathName := codegen.SnakeCase(codegen.Goify(service.Name(), false))
			packageName := strings.ToLower(codegen.Goify(service.Name(), false))
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"c", path.Join(generation.GenPkg(), dir, pathName, "client"))); err != nil {
				return err
			}
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"svr", path.Join(generation.GenPkg(), dir, pathName, "server"))); err != nil {
				return err
			}
		}
		if len(expressions.Services) > 0 {
			for _, server := range design.API.Servers {
				serverName := codegen.SnakeCase(codegen.Goify(server.Name, true))
				if err := generation.ReserveGeneratedImport(codegen.NewImport("cli", path.Join(generation.GenPkg(), dir, "cli", serverName))); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// newPlans validates the full input set and submits names for every plan.
func newPlans(generation *codegen.Generation, transport transportKind, inputs []PlanInput) ([]*Plan, error) {
	if generation == nil {
		return nil, fmt.Errorf("HTTP plans require a generation")
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
	if err := planImports(generation, transport); err != nil {
		return nil, err
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
		root:         input.Root,
		servicePlan:  input.Service,
		generation:   generation,
		transport:    transport,
		constructors: make(map[viewedConstructorKey]*codegen.NameDeclaration),
		payloads:     make(map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration),
		streams:      make(map[*expr.HTTPEndpointExpr]*codegen.NameDeclaration),
		errors:       make(map[*expr.HTTPErrorExpr]*codegen.NameDeclaration),
		wireTypes:    make(map[*expr.HTTPServiceExpr]*plannedWireTypes),
		symbols:      make(map[*expr.HTTPServiceExpr]*httpSymbols),
		cliParsers:   make(map[*expr.ServerExpr]*cli.ParserPlan),
	}
	expressions := transportExpressions(input.Root, transport)
	dir := transportDirectory(transport)
	for _, transportService := range expressions.Services {
		clientPath := path.Join(generation.GenPkg(), dir, codegen.SnakeCase(transportService.Name()), "client")
		clientPackage, err := generation.ClaimPackage(clientPath)
		if err != nil {
			return nil, err
		}
		serverPath := path.Join(generation.GenPkg(), dir, codegen.SnakeCase(transportService.Name()), "server")
		serverPackage, err := generation.ClaimPackage(serverPath)
		if err != nil {
			return nil, err
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
			server:         serverCatalog,
			client:         clientCatalog,
			streamPayloads: make(map[*expr.HTTPEndpointExpr]*wireTypeRecord),
		}
		collectPlannedWireTypes(transportService, planned, input.Service)
		plan.wireTypes[transportService] = planned
		symbols, err := collectHTTPSymbols(plan, transportService, clientPackage, serverPackage)
		if err != nil {
			return nil, err
		}
		plan.symbols[transportService] = symbols
		for _, endpoint := range transportService.HTTPEndpoints {
			order := viewedConstructorOrder{
				transport: dir,
				service:   transportService.Name(),
				method:    endpoint.Name(),
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
			if needInit(endpoint.MethodExpr.Result.Type) {
				resultType, viewed := endpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr)
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
						declaration, err := declareHTTPConstructor(clientPackage, viewedResultConstructorName(endpoint, response, view), responseOrder)
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
			}
			commands = append(commands, command)
		}
		parser, err := cli.DeclareParser(serverPackage, dir, input.Root.API.Name, server.Name, commands)
		if err != nil {
			return nil, err
		}
		plan.cliParsers[server] = parser
	}
	return plan, nil
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
	for _, transportService := range services.Expressions.Services {
		if services.ServicesData.Get(transportService.Name()) == nil {
			return fmt.Errorf("HTTP service %q has no linked service model", transportService.Name())
		}
		services.HTTPData[transportService.Name()] = services.analyze(transportService)
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
			services:    services,
			fileImports: make(map[string][]*codegen.ImportSpec),
			clientCodec: clientEncodeDecodeFile(transportService, services),
			serverCodec: serverEncodeDecodeFile(transportService, services),
		}
		if p.transport == jsonrpcTransport {
			jsonService.prepareFileImports(transportService, services)
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
		p.example = exampleServerFiles(services)
	}
	p.exampleCLI = exampleCLIFiles(services)
	p.serverTypes = serverTypeFiles(services)
	p.clientTypes = clientTypeFiles(services)
	p.paths = pathFiles(services)
	p.clientCLI = clientCLIFiles(services)
	return nil
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

// prepareFileImports computes the service-type imports for every JSON-RPC file
// that this service can generate. JSON-RPC later reads these lists without
// walking the HTTP endpoint types again.
func (p *jsonRPCServicePlan) prepareFileImports(transportService *expr.HTTPServiceExpr, services *ServicesData) {
	var all, sse, websocket []*expr.AttributeExpr
	for index, endpoint := range transportService.HTTPEndpoints {
		references := serviceReferenceAttributes(endpoint)
		all = append(all, references...)
		switch {
		case p.data.Endpoints[index].SSE != nil:
			sse = append(sse, references...)
		case IsWebSocketEndpoint(p.data.Endpoints[index]):
			websocket = append(websocket, references...)
		}
	}
	servicePath := p.data.Service.PathName
	clientPackage := path.Join(services.GenPkg(), "jsonrpc", servicePath, "client")
	serverPackage := path.Join(services.GenPkg(), "jsonrpc", servicePath, "server")
	clientAll := services.AttributeImports(clientPackage, all...)
	serverAll := services.AttributeImports(serverPackage, all...)
	p.fileImports[path.Join(codegen.Gendir, "jsonrpc", servicePath, "client", "client.go")] = clientAll
	p.fileImports[path.Join(codegen.Gendir, "jsonrpc", servicePath, "server", "server.go")] = serverAll
	if p.clientCodec != nil {
		p.fileImports[p.clientCodec.Path] = clientAll
	}
	if p.serverCodec != nil {
		p.fileImports[p.serverCodec.Path] = serverAll
	}
	if len(sse) > 0 {
		p.fileImports[path.Join(codegen.Gendir, "jsonrpc", servicePath, "client", "stream.go")] = services.AttributeImports(clientPackage, sse...)
		p.fileImports[path.Join(codegen.Gendir, "jsonrpc", servicePath, "server", "sse.go")] = services.AttributeImports(serverPackage, sse...)
	}
	if len(websocket) > 0 {
		p.fileImports[path.Join(codegen.Gendir, "jsonrpc", servicePath, "client", "websocket.go")] = services.AttributeImports(clientPackage, websocket...)
		if len(sse) == 0 {
			p.fileImports[path.Join(codegen.Gendir, "jsonrpc", servicePath, "server", "websocket.go")] = services.AttributeImports(serverPackage, websocket...)
		}
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

// cloneImportSpec copies one import so callers can change its path or name
// without changing the import stored by the HTTP plan.
func cloneImportSpec(source *codegen.ImportSpec) *codegen.ImportSpec {
	if source == nil {
		return nil
	}
	copy := *source
	return &copy
}

// viewedResultConstructorName returns the preferred constructor spelling for
// one client response body selected by a result view.
func viewedResultConstructorName(endpoint *expr.HTTPEndpointExpr, response *expr.HTTPResponseExpr, view string) string {
	return "New" + codegen.Goify(endpoint.Name(), true) + "Result" + codegen.Goify(view, true) + codegen.Goify(http.StatusText(response.StatusCode), true)
}

// endpointPayloadConstructorName returns the preferred server function name
// that builds one method payload from its HTTP request values.
func endpointPayloadConstructorName(endpoint *expr.HTTPEndpointExpr) string {
	return "New" + codegen.Goify(endpoint.Name(), true) + "Payload"
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
