// This file prepares JSON-RPC files in two calls. NewPlans receives every
// design and requests all Go names. After Goa assigns those names and builds
// the service and HTTP values, Link creates the generated files.
package codegen

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// PlanInput supplies one design and the prepared values for its generated
	// service, JSON requests, and JSON responses.
	PlanInput struct {
		// Root is the design that declares the JSON-RPC services.
		Root *expr.RootExpr
		// Service provides the generated service types and method definitions.
		Service *service.Plan
		// HTTP provides the JSON request and response types used inside JSON-RPC messages.
		HTTP *httpcodegen.Plan
		// ApplicationHTTP is the ordinary HTTP plan whose runnable server is
		// combined with JSON-RPC. It is nil when Root has no ordinary HTTP services.
		ApplicationHTTP *httpcodegen.Plan
	}

	// Plan stores the JSON-RPC function names chosen by NewPlans and the files
	// created by Link. Goa creates one Plan for each design.
	Plan struct {
		generation      *codegen.Generation
		root            *expr.RootExpr
		service         *service.Plan
		http            *httpcodegen.Plan
		applicationHTTP *httpcodegen.Plan
		services        []*servicePlan
		server          []*codegen.File
		client          []*codegen.File
		example         []*codegen.File
		exampleCLI      []*codegen.File
		linked          bool
	}

	// servicePlan stores one service's generated package path, function names, and
	// HTTP request and response data. It also records whether methods return one
	// response, server-sent events, or WebSocket messages.
	servicePlan struct {
		data          httpcodegen.JSONRPCServiceSnapshot
		name          string
		pathName      string
		endpoints     []*endpointPlan
		helpers       map[string]*viewedHelperDeclarations
		endpointNames map[string]*jsonRPCEndpointNames
		clientNames   jsonRPCClientNames
		serverNames   jsonRPCServerNames
		bodyDecoder   *codegen.NameDeclaration
		hasHTTP       bool
		hasSSE        bool
		hasWebSocket  bool
	}

	// endpointPlan contains the HTTP request, HTTP response, and JSON-RPC result
	// values for one service method.
	endpointPlan struct {
		httpcodegen.JSONRPCEndpointSnapshot
		viewed           *viewedRepresentation
		websocketPending *codegen.NameDeclaration
		websocketResult  *codegen.NameDeclaration
		websocketWrapper *codegen.NameDeclaration
	}

	// jsonRPCClientNames stores the Go names written once for one client.
	jsonRPCClientNames struct {
		bufferPool              *codegen.NameDeclaration
		websocketConnection     *codegen.NameDeclaration
		websocketRequestOwner   *codegen.NameDeclaration
		websocketPendingRequest *codegen.NameDeclaration
		websocketMessage        *codegen.NameDeclaration
		websocketClosedError    *codegen.NameDeclaration
		newWebsocketConnection  *codegen.NameDeclaration
		streamErrorType         *codegen.NameDeclaration
		streamErrorConnection   *codegen.NameDeclaration
		streamErrorProtocol     *codegen.NameDeclaration
		streamErrorParsing      *codegen.NameDeclaration
		streamErrorOrphaned     *codegen.NameDeclaration
		streamErrorTimeout      *codegen.NameDeclaration
		streamErrorHandler      *codegen.NameDeclaration
	}

	// jsonRPCServerNames stores the Go names written once for one server.
	jsonRPCServerNames struct {
		batchWriter     *codegen.NameDeclaration
		encodeError     *codegen.NameDeclaration
		sseStream       *codegen.NameDeclaration
		sseBuffer       *codegen.NameDeclaration
		websocketStream *codegen.NameDeclaration
	}

	// jsonRPCEndpointNames stores the extra Go names written for one WebSocket
	// method.
	jsonRPCEndpointNames struct {
		websocketPending *codegen.NameDeclaration
		websocketResult  *codegen.NameDeclaration
		websocketWrapper *codegen.NameDeclaration
	}

	// viewedRepresentation lists the JSON body type and constructor used for
	// each view that a method may return.
	viewedRepresentation struct {
		variable      bool
		fixedView     string
		branches      []viewBranch
		decode        *codegen.NameDeclaration
		encode        *codegen.NameDeclaration
		streamEncode  *codegen.NameDeclaration
		writeMetadata *codegen.NameDeclaration
		viewedResult  httpcodegen.JSONRPCViewedResultData
		servicePkg    string
		resultRef     string
	}

	// viewBranch stores the mapped service field, JSON body types, and client
	// constructor for one view.
	viewBranch struct {
		view       string
		resultAttr string
		serverBody *httpcodegen.JSONRPCBodyData
		clientBody *httpcodegen.JSONRPCBodyData
		resultInit httpcodegen.InitData
		headers    []httpcodegen.JSONRPCHeaderData
		cookies    []httpcodegen.JSONRPCCookieData
	}

	// viewedHelperDeclarations stores the client decoder and server encoder names
	// written for one method result.
	viewedHelperDeclarations struct {
		decode        *codegen.NameDeclaration
		encode        *codegen.NameDeclaration
		streamEncode  *codegen.NameDeclaration
		writeMetadata *codegen.NameDeclaration
	}

	// jsonRPCNameOrder gives the same Go names the same order on every run.
	jsonRPCNameOrder struct {
		service string
		method  string
		role    uint8
	}
)

const (
	viewedBodyDecoderRole uint8 = iota + 1
	viewedResultDecoderRole
	viewedResultEncoderRole
	viewedStreamEncoderRole
	viewedMetadataWriterRole
	jsonRPCBufferPoolRole
	jsonRPCBatchWriterRole
	jsonRPCEncodeErrorRole
	jsonRPCSSEStreamRole
	jsonRPCSSEBufferRole
	jsonRPCWebSocketConnectionRole
	jsonRPCWebSocketRequestOwnerRole
	jsonRPCWebSocketPendingRequestRole
	jsonRPCWebSocketMessageRole
	jsonRPCWebSocketClosedErrorRole
	jsonRPCNewWebSocketConnectionRole
	jsonRPCStreamErrorTypeRole
	jsonRPCStreamErrorConnectionRole
	jsonRPCStreamErrorProtocolRole
	jsonRPCStreamErrorParsingRole
	jsonRPCStreamErrorOrphanedRole
	jsonRPCStreamErrorTimeoutRole
	jsonRPCStreamErrorHandlerRole
	jsonRPCWebSocketServerStreamRole
	jsonRPCWebSocketMethodPendingRole
	jsonRPCWebSocketMethodResultRole
	jsonRPCWebSocketServerWrapperRole
)

// NewPlans checks that inputs contain every design with JSON-RPC services once,
// then creates one Plan for each input. It requests every helper name before
// Goa chooses unique Go names, so generated definitions and calls agree.
func NewPlans(generation *codegen.Generation, inputs ...PlanInput) ([]*Plan, error) {
	if generation == nil {
		return nil, fmt.Errorf("JSON-RPC plans require a generation")
	}
	if generation.Frozen() {
		return nil, fmt.Errorf("JSON-RPC plans must be collected before generation freeze")
	}
	if err := validatePlanInputs(generation, inputs); err != nil {
		return nil, err
	}
	if err := example.Plan(generation); err != nil {
		return nil, err
	}
	if err := planImports(generation, inputs); err != nil {
		return nil, err
	}
	plans := make([]*Plan, len(inputs))
	for index, input := range inputs {
		plan := &Plan{
			generation:      generation,
			root:            input.Root,
			service:         input.Service,
			http:            input.HTTP,
			applicationHTTP: input.ApplicationHTTP,
		}
		for _, transport := range input.Root.API.JSONRPC.Services {
			planned, err := collectServicePlan(generation, transport)
			if err != nil {
				return nil, err
			}
			plan.services = append(plan.services, planned)
		}
		sort.Slice(plan.services, func(i, j int) bool {
			return plan.services[i].name < plan.services[j].name
		})
		plans[index] = plan
	}
	return plans, nil
}

// Root returns the design used to create p.
func (p *Plan) Root() *expr.RootExpr {
	return p.root
}

// Link reads the completed service and HTTP plans and builds every JSON-RPC
// file. The caller must first ask Goa to choose unique Go names and then link
// both input plans so all JSON body types and constructors are available.
func (p *Plan) Link() error {
	if !p.generation.Frozen() {
		return fmt.Errorf("JSON-RPC plan cannot link before generation freeze")
	}
	if p.linked {
		return fmt.Errorf("JSON-RPC plan is already linked")
	}
	for _, planned := range p.services {
		data, ok := p.http.JSONRPCService(planned.name)
		if !ok {
			return fmt.Errorf("HTTP plan has no data for JSON-RPC service %q", planned.name)
		}
		planned.data = data
		planned.pathName = data.Service.PathName
		for _, endpoint := range data.Endpoints {
			helper := planned.helpers[endpoint.Method.Name]
			names := planned.endpointNames[endpoint.Method.Name]
			viewed, hasViewedResult := p.http.ViewedResult(planned.name, endpoint.Method.Name)
			plannedEndpoint := &endpointPlan{
				JSONRPCEndpointSnapshot: endpoint,
				viewed:                  planViewedRepresentation(&endpoint, viewed, hasViewedResult, helper),
			}
			if names != nil {
				plannedEndpoint.websocketPending = names.websocketPending
				plannedEndpoint.websocketResult = names.websocketResult
				plannedEndpoint.websocketWrapper = names.websocketWrapper
			}
			planned.endpoints = append(planned.endpoints, plannedEndpoint)
			switch {
			case endpoint.SSE != nil:
				planned.hasSSE = true
			case isJSONRPCWebSocketEndpoint(endpoint):
				planned.hasWebSocket = true
			default:
				planned.hasHTTP = true
			}
		}
	}
	p.server = serverFiles(p.services)
	p.client = clientFiles(p.services)
	p.example = p.http.CombinedExampleServerFiles(p.applicationHTTP)
	p.exampleCLI = p.http.ExampleCLIFiles()
	p.linked = true
	return nil
}

// ServerFiles returns the JSON-RPC server files built by Link.
func (p *Plan) ServerFiles() []*codegen.File {
	p.requireLinked()
	return p.server
}

// ClientFiles returns the JSON-RPC client files built by Link.
func (p *Plan) ClientFiles() []*codegen.File {
	p.requireLinked()
	return p.client
}

// ServerTypeFiles returns the server JSON body files supplied by the HTTP plan.
func (p *Plan) ServerTypeFiles() []*codegen.File {
	p.requireLinked()
	return p.http.ServerTypeFiles()
}

// ClientTypeFiles returns the client JSON body files supplied by the HTTP plan.
func (p *Plan) ClientTypeFiles() []*codegen.File {
	p.requireLinked()
	return p.http.ClientTypeFiles()
}

// PathFiles returns the URL path helper files supplied by the HTTP plan.
func (p *Plan) PathFiles() []*codegen.File {
	p.requireLinked()
	return p.http.PathFiles()
}

// ClientCLIFiles returns the command-line client files supplied by the HTTP plan.
func (p *Plan) ClientCLIFiles() []*codegen.File {
	p.requireLinked()
	return p.http.ClientCLIFiles()
}

// ExampleServerFiles returns the runnable servers built by Link. Each file
// mounts both ordinary HTTP and JSON-RPC services declared on that server.
func (p *Plan) ExampleServerFiles() []*codegen.File {
	p.requireLinked()
	return p.example
}

// ExampleCLIFiles returns runnable command-line clients for p's JSON-RPC services.
func (p *Plan) ExampleCLIFiles() []*codegen.File {
	p.requireLinked()
	return p.exampleCLI
}

// planViewedRepresentation copies each allowed result view and its JSON body
// type into the values used to write JSON-RPC files. A method fixed to one
// view stores one branch. A method that chooses a view for each result stores
// every branch and includes the selected view name in each response.
func planViewedRepresentation(endpoint *httpcodegen.JSONRPCEndpointSnapshot, viewed httpcodegen.ViewedResultSnapshot, hasViewedResult bool, helpers *viewedHelperDeclarations) *viewedRepresentation {
	if !hasViewedResult {
		return nil
	}
	if helpers == nil {
		panic(fmt.Sprintf("JSON-RPC viewed endpoint %q has no helper names declared by NewPlans", endpoint.Method.Name))
	}
	representation := &viewedRepresentation{
		variable:      viewed.Variable,
		fixedView:     viewed.FixedView,
		decode:        helpers.decode,
		encode:        helpers.encode,
		streamEncode:  helpers.streamEncode,
		writeMetadata: helpers.writeMetadata,
		viewedResult:  viewed.Service,
		servicePkg:    endpoint.ServicePkgName,
		resultRef:     endpoint.Result.Ref,
	}
	for _, branch := range viewed.Representations {
		representation.branches = append(representation.branches, viewBranch{
			view:       branch.View,
			resultAttr: branch.ResultAttr,
			serverBody: branch.ServerBody,
			clientBody: branch.ClientBody,
			resultInit: branch.ResultInit,
			headers:    branch.Headers,
			cookies:    branch.Cookies,
		})
	}
	return representation
}

// requireLinked stops callers from reading files before Link has built them.
func (p *Plan) requireLinked() {
	if !p.linked {
		panic("JSON-RPC files requested before Plan.Link")
	}
}

// planImports records every import name written directly into JSON-RPC files.
func planImports(generation *codegen.Generation, inputs []PlanInput) error {
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bufio"),
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("mime/multipart"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("path"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("sync"),
		codegen.SimpleImport("sync/atomic"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("github.com/gorilla/websocket"),
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		codegen.GoaImport("jsonrpc"),
	}
	for _, spec := range imports {
		if err := generation.RequireImport(spec); err != nil {
			return err
		}
	}
	for _, input := range inputs {
		design := input.Root
		for _, service := range design.API.JSONRPC.Services {
			pathName := codegen.SnakeCase(codegen.Goify(service.Name(), false))
			packageName := strings.ToLower(codegen.Goify(service.Name(), false))
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"c", path.Join(generation.GenPkg(), "jsonrpc", pathName, "client"))); err != nil {
				return err
			}
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"jssvr", path.Join(generation.GenPkg(), "jsonrpc", pathName, "server"))); err != nil {
				return err
			}
		}
		if len(design.API.JSONRPC.Services) > 0 {
			for _, server := range design.API.Servers {
				serverName := codegen.SnakeCase(codegen.Goify(server.Name, true))
				if err := generation.ReserveGeneratedImport(codegen.NewImport("cli", path.Join(generation.GenPkg(), "jsonrpc", "cli", serverName))); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validatePlanInputs checks every root and plan before NewPlans submits an
// import or generated helper name. This keeps a rejected input from changing
// names chosen for later generators in the same run.
func validatePlanInputs(generation *codegen.Generation, inputs []PlanInput) error {
	roots := make(map[*expr.RootExpr]struct{})
	for _, candidate := range generation.Roots() {
		root, ok := candidate.(*expr.RootExpr)
		if ok && len(root.API.JSONRPC.Services) > 0 {
			roots[root] = struct{}{}
		}
	}
	seen := make(map[*expr.RootExpr]struct{}, len(inputs))
	for _, input := range inputs {
		if input.Root == nil || !generation.HasRoot(input.Root) {
			return fmt.Errorf("JSON-RPC plan requires a root in this generation")
		}
		if _, ok := roots[input.Root]; !ok {
			return fmt.Errorf("root does not declare JSON-RPC services")
		}
		if _, ok := seen[input.Root]; ok {
			return fmt.Errorf("JSON-RPC root is planned more than once: %s", rootServiceName(input.Root))
		}
		seen[input.Root] = struct{}{}
		if input.Service == nil {
			return fmt.Errorf("JSON-RPC root %s requires a service plan", rootServiceName(input.Root))
		}
		if input.Service.Root() != input.Root {
			return fmt.Errorf("JSON-RPC service plan does not belong to root %s", rootServiceName(input.Root))
		}
		if input.HTTP == nil {
			return fmt.Errorf("JSON-RPC root %s requires an HTTP plan for its JSON request and response types", rootServiceName(input.Root))
		}
		if !input.HTTP.MatchesJSONRPC(input.Root, input.Service) {
			return fmt.Errorf("JSON-RPC HTTP plan does not belong to root %s and its service plan", rootServiceName(input.Root))
		}
		hasHTTP := len(input.Root.API.HTTP.Services) > 0
		if hasHTTP && input.ApplicationHTTP == nil {
			return fmt.Errorf("JSON-RPC root %s requires its application HTTP plan", rootServiceName(input.Root))
		}
		if !hasHTTP && input.ApplicationHTTP != nil {
			return fmt.Errorf("JSON-RPC root %s has no ordinary HTTP services", rootServiceName(input.Root))
		}
		if input.ApplicationHTTP != nil && !input.ApplicationHTTP.MatchesHTTP(input.Root, input.Service) {
			return fmt.Errorf("application HTTP plan does not belong to root %s and its service plan", rootServiceName(input.Root))
		}
	}
	if len(inputs) != len(roots) {
		return fmt.Errorf("JSON-RPC planning requires all %d JSON-RPC roots, got %d", len(roots), len(inputs))
	}
	return nil
}

// rootServiceName returns the name shown in errors after validation has proved
// that root declares a JSON-RPC service.
func rootServiceName(root *expr.RootExpr) string {
	return root.Services[0].Name
}

// isJSONRPCSSEEndpoint reports whether the supplied method writes server-sent
// events.
func isJSONRPCSSEEndpoint(data any) bool {
	return jsonRPCEndpoint(data).SSE != nil
}

// isJSONRPCWebSocketEndpoint reports whether the supplied method sends or
// receives JSON-RPC messages through a WebSocket.
func isJSONRPCWebSocketEndpoint(data any) bool {
	endpoint := jsonRPCEndpoint(data)
	return endpoint.ClientWebSocket != nil || endpoint.ServerWebSocket != nil
}

// jsonRPCEndpoint returns the method values used to write a generated file.
func jsonRPCEndpoint(data any) *httpcodegen.JSONRPCEndpointSnapshot {
	switch endpoint := data.(type) {
	case httpcodegen.JSONRPCEndpointSnapshot:
		return &endpoint
	case *httpcodegen.JSONRPCEndpointSnapshot:
		return endpoint
	case *endpointPlan:
		return &endpoint.JSONRPCEndpointSnapshot
	default:
		panic(fmt.Sprintf("JSON-RPC template received endpoint data of type %T", data))
	}
}

// collectServicePlan stores the designed service name and generated package
// path, then requests client decoder and server encoder names for every method
// that returns a result view. Link later adds the HTTP endpoint data used to
// build that service's files.
func collectServicePlan(generation *codegen.Generation, transport *expr.HTTPServiceExpr) (*servicePlan, error) {
	pathName := codegen.SnakeCase(codegen.Goify(transport.Name(), false))
	clientPath := path.Join(generation.GenPkg(), "jsonrpc", pathName, "client")
	serverPath := path.Join(generation.GenPkg(), "jsonrpc", pathName, "server")
	client, err := generation.ClaimPackage(clientPath)
	if err != nil {
		return nil, err
	}
	server, err := generation.ClaimPackage(serverPath)
	if err != nil {
		return nil, err
	}
	planned := &servicePlan{
		name:          transport.Name(),
		pathName:      pathName,
		helpers:       make(map[string]*viewedHelperDeclarations),
		endpointNames: make(map[string]*jsonRPCEndpointNames),
	}
	declare := func(pkg *codegen.GeneratedPackage, kind codegen.PackageNameKind, preferred string, visibility codegen.PackageNameVisibility, method string, role uint8) (*codegen.NameDeclaration, error) {
		declaration := codegen.NewPreferredName(kind, preferred, visibility, jsonRPCNameOrder{
			service: planned.name,
			method:  method,
			role:    role,
		})
		if err := pkg.DeclareName(declaration); err != nil {
			return nil, err
		}
		return declaration, nil
	}
	hasHTTP, hasSSE, hasWebSocket := false, false, false
	for _, endpoint := range transport.HTTPEndpoints {
		switch {
		case endpoint.UsesSSE():
			hasSSE = true
		case endpoint.UsesWebSocket():
			hasWebSocket = true
		default:
			hasHTTP = true
		}
	}
	if !hasWebSocket {
		planned.clientNames.bufferPool, err = declare(client, codegen.NameVariable, "bufferPool", codegen.UnexportedName, "", jsonRPCBufferPoolRole)
		if err != nil {
			return nil, err
		}
		planned.serverNames.encodeError, err = declare(server, codegen.NameFunction, "encodeJSONRPCError", codegen.UnexportedName, "", jsonRPCEncodeErrorRole)
		if err != nil {
			return nil, err
		}
	}
	if hasHTTP {
		planned.serverNames.batchWriter, err = declare(server, codegen.NameType, "batchWriter", codegen.UnexportedName, "", jsonRPCBatchWriterRole)
		if err != nil {
			return nil, err
		}
	}
	if hasSSE {
		planned.serverNames.sseStream, err = declare(server, codegen.NameType, "sseServerStream", codegen.UnexportedName, "", jsonRPCSSEStreamRole)
		if err != nil {
			return nil, err
		}
		planned.serverNames.sseBuffer, err = declare(server, codegen.NameType, "sseEventBuffer", codegen.UnexportedName, "", jsonRPCSSEBufferRole)
		if err != nil {
			return nil, err
		}
	}
	if hasWebSocket {
		clientDeclarations := []struct {
			target     **codegen.NameDeclaration
			kind       codegen.PackageNameKind
			preferred  string
			visibility codegen.PackageNameVisibility
			role       uint8
		}{
			{&planned.clientNames.websocketConnection, codegen.NameType, "websocketClientConn", codegen.UnexportedName, jsonRPCWebSocketConnectionRole},
			{&planned.clientNames.websocketRequestOwner, codegen.NameType, "websocketRequestOwner", codegen.UnexportedName, jsonRPCWebSocketRequestOwnerRole},
			{&planned.clientNames.websocketPendingRequest, codegen.NameType, "websocketPendingRequest", codegen.UnexportedName, jsonRPCWebSocketPendingRequestRole},
			{&planned.clientNames.websocketMessage, codegen.NameType, "websocketMessage", codegen.UnexportedName, jsonRPCWebSocketMessageRole},
			{&planned.clientNames.websocketClosedError, codegen.NameVariable, "errWebsocketMethodStreamClosed", codegen.UnexportedName, jsonRPCWebSocketClosedErrorRole},
			{&planned.clientNames.newWebsocketConnection, codegen.NameFunction, "newWebsocketClientConn", codegen.UnexportedName, jsonRPCNewWebSocketConnectionRole},
			{&planned.clientNames.streamErrorType, codegen.NameType, "StreamErrorType", codegen.ExportedName, jsonRPCStreamErrorTypeRole},
			{&planned.clientNames.streamErrorConnection, codegen.NameConstant, "StreamErrorConnection", codegen.ExportedName, jsonRPCStreamErrorConnectionRole},
			{&planned.clientNames.streamErrorProtocol, codegen.NameConstant, "StreamErrorProtocol", codegen.ExportedName, jsonRPCStreamErrorProtocolRole},
			{&planned.clientNames.streamErrorParsing, codegen.NameConstant, "StreamErrorParsing", codegen.ExportedName, jsonRPCStreamErrorParsingRole},
			{&planned.clientNames.streamErrorOrphaned, codegen.NameConstant, "StreamErrorOrphaned", codegen.ExportedName, jsonRPCStreamErrorOrphanedRole},
			{&planned.clientNames.streamErrorTimeout, codegen.NameConstant, "StreamErrorTimeout", codegen.ExportedName, jsonRPCStreamErrorTimeoutRole},
			{&planned.clientNames.streamErrorHandler, codegen.NameType, "StreamErrorHandler", codegen.ExportedName, jsonRPCStreamErrorHandlerRole},
		}
		for _, item := range clientDeclarations {
			*item.target, err = declare(client, item.kind, item.preferred, item.visibility, "", item.role)
			if err != nil {
				return nil, err
			}
		}
		preferredStream := codegen.Goify(transport.Name(), false) + "Stream"
		planned.serverNames.websocketStream, err = declare(server, codegen.NameType, preferredStream, codegen.UnexportedName, "", jsonRPCWebSocketServerStreamRole)
		if err != nil {
			return nil, err
		}
	}
	for _, endpoint := range transport.HTTPEndpoints {
		method := endpoint.MethodExpr
		if endpoint.UsesWebSocket() {
			names := &jsonRPCEndpointNames{}
			if method.StreamingResult != nil {
				names.websocketPending, err = declare(client, codegen.NameType, codegen.Goify(method.Name, false)+"ClientStreamPendingRequest", codegen.UnexportedName, method.Name, jsonRPCWebSocketMethodPendingRole)
				if err != nil {
					return nil, err
				}
				names.websocketResult, err = declare(client, codegen.NameType, codegen.Goify(method.Name, false)+"ClientStreamStreamResult", codegen.UnexportedName, method.Name, jsonRPCWebSocketMethodResultRole)
				if err != nil {
					return nil, err
				}
			}
			if method.Stream == expr.ServerStreamKind || method.Stream == expr.BidirectionalStreamKind {
				names.websocketWrapper, err = declare(server, codegen.NameType, codegen.Goify(method.Name, false)+"StreamWrapper", codegen.UnexportedName, method.Name, jsonRPCWebSocketServerWrapperRole)
				if err != nil {
					return nil, err
				}
			}
			planned.endpointNames[method.Name] = names
		}
		if _, ok := method.Result.Type.(*expr.ResultTypeExpr); !ok {
			continue
		}
		if planned.bodyDecoder == nil {
			planned.bodyDecoder = codegen.NewPreferredName(
				codegen.NameFunction,
				"decodeJSONRPCResult",
				codegen.UnexportedName,
				jsonRPCNameOrder{service: planned.name, role: viewedBodyDecoderRole},
			)
			if err := client.DeclareName(planned.bodyDecoder); err != nil {
				return nil, err
			}
		}
		methodName := codegen.Goify(method.Name, true)
		helpers := &viewedHelperDeclarations{
			decode: codegen.NewPreferredName(
				codegen.NameFunction,
				"decode"+methodName+"ViewedResult",
				codegen.UnexportedName,
				jsonRPCNameOrder{service: planned.name, method: method.Name, role: viewedResultDecoderRole},
			),
			encode: codegen.NewPreferredName(
				codegen.NameFunction,
				"encode"+methodName+"ViewedResult",
				codegen.UnexportedName,
				jsonRPCNameOrder{service: planned.name, method: method.Name, role: viewedResultEncoderRole},
			),
			streamEncode: codegen.NewPreferredName(
				codegen.NameFunction,
				"encode"+methodName+"Result",
				codegen.UnexportedName,
				jsonRPCNameOrder{service: planned.name, method: method.Name, role: viewedStreamEncoderRole},
			),
			writeMetadata: codegen.NewPreferredName(
				codegen.NameFunction,
				"write"+methodName+"ViewedResponseMetadata",
				codegen.UnexportedName,
				jsonRPCNameOrder{service: planned.name, method: method.Name, role: viewedMetadataWriterRole},
			),
		}
		if err := client.DeclareName(helpers.decode); err != nil {
			return nil, err
		}
		if err := server.DeclareName(helpers.encode); err != nil {
			return nil, err
		}
		if err := server.DeclareName(helpers.streamEncode); err != nil {
			return nil, err
		}
		if err := server.DeclareName(helpers.writeMetadata); err != nil {
			return nil, err
		}
		planned.helpers[method.Name] = helpers
	}
	return planned, nil
}

// ComparePackageName orders Go declarations by service, method, and use.
func (o jsonRPCNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(jsonRPCNameOrder)
	if compared := strings.Compare(o.service, right.service); compared != 0 {
		return compared
	}
	if compared := strings.Compare(o.method, right.method); compared != 0 {
		return compared
	}
	return int(o.role) - int(right.role)
}
