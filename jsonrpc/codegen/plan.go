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
	// PlanInput supplies one design and the copied values used to generate its
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
		servicesByExpr  map[*expr.HTTPServiceExpr]*servicePlan
		server          []*codegen.File
		client          []*codegen.File
		linked          bool
	}

	// ExamplePlan builds runnable JSON-RPC programs from server data and
	// generated services that came from the same design.
	ExamplePlan struct {
		transport *Plan
		http      *httpcodegen.ExamplePlan
	}

	// servicePlan stores one service's generated package path, function names,
	// and HTTP request and response data.
	servicePlan struct {
		data        httpcodegen.JSONRPCServiceSnapshot
		api         string
		name        string
		pathName    string
		endpoints   []*endpointPlan
		helpers     map[string]*viewedHelperDeclarations
		clientNames jsonRPCClientNames
		serverNames jsonRPCServerNames
		bodyDecoder *codegen.NameDeclaration
		hasHTTP     bool
		hasSSE      bool
	}

	// endpointPlan contains the HTTP request, HTTP response, and JSON-RPC result
	// values for one service method.
	endpointPlan struct {
		httpcodegen.JSONRPCEndpointSnapshot
		viewed *viewedRepresentation
	}

	// jsonRPCClientNames stores the Go names written once for one client.
	jsonRPCClientNames struct {
		bufferPool *codegen.NameDeclaration
	}

	// jsonRPCServerNames stores the Go names written once for one server.
	jsonRPCServerNames struct {
		batchWriter    *codegen.NameDeclaration
		encodeError    *codegen.NameDeclaration
		sseStream      *codegen.NameDeclaration
		sseBuffer      *codegen.NameDeclaration
		noOutputWriter *codegen.NameDeclaration
	}

	// viewedRepresentation lists the JSON body type and constructor used for
	// each view that a method may return.
	viewedRepresentation struct {
		variable     bool
		fixedView    string
		branches     []viewBranch
		decode       *codegen.NameDeclaration
		encode       *codegen.NameDeclaration
		streamEncode *codegen.NameDeclaration
		viewedResult httpcodegen.JSONRPCViewedResultData
		servicePkg   string
		resultRef    string
	}

	// viewBranch stores the mapped service field, JSON body types, and client
	// constructor for one view.
	viewBranch struct {
		view       string
		resultAttr string
		serverBody *httpcodegen.JSONRPCBodyData
		clientBody *httpcodegen.JSONRPCBodyData
		resultInit httpcodegen.InitData
	}

	// viewedHelperDeclarations stores the client decoder and server encoder names
	// written for one method result.
	viewedHelperDeclarations struct {
		decode       *codegen.NameDeclaration
		encode       *codegen.NameDeclaration
		streamEncode *codegen.NameDeclaration
	}

	// jsonRPCNameOrder gives the same Go names the same order on every run.
	jsonRPCNameOrder struct {
		api     string
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
	jsonRPCBufferPoolRole
	jsonRPCBatchWriterRole
	jsonRPCEncodeErrorRole
	jsonRPCSSEStreamRole
	jsonRPCSSEBufferRole
	jsonRPCNoOutputWriterRole
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
	plans := make([]*Plan, len(inputs))
	for index, input := range inputs {
		plan := &Plan{
			generation:      generation,
			root:            input.Root,
			service:         input.Service,
			http:            input.HTTP,
			applicationHTTP: input.ApplicationHTTP,
			servicesByExpr:  make(map[*expr.HTTPServiceExpr]*servicePlan),
		}
		for _, transport := range input.Root.API.JSONRPC.Services {
			planned, err := collectServicePlan(generation, input, transport)
			if err != nil {
				return nil, err
			}
			plan.services = append(plan.services, planned)
			plan.servicesByExpr[transport] = planned
		}
		sort.Slice(plan.services, func(i, j int) bool {
			return plan.services[i].name < plan.services[j].name
		})
		plans[index] = plan
	}
	if err := planImports(generation, inputs); err != nil {
		return nil, err
	}
	return plans, nil
}

// NewExamplePlan returns an example renderer only when examples contains the
// server data copied from transport's service design.
func NewExamplePlan(transport *Plan, examples *example.Plan) (*ExamplePlan, error) {
	if _, ok := examples.Root(transport.service); !ok {
		return nil, fmt.Errorf("JSON-RPC examples require server data created from the same service design")
	}
	httpPlan, err := httpcodegen.NewExamplePlan(transport.http, examples)
	if err != nil {
		return nil, err
	}
	return &ExamplePlan{transport: transport, http: httpPlan}, nil
}

// Root returns the design used to create p.
func (p *Plan) Root() *expr.RootExpr {
	return p.root
}

// Service returns the finalized JSON-RPC data for the exact service used to
// build this plan. Callers must call Link before reading the service data.
func (p *Plan) Service(service *expr.HTTPServiceExpr) (httpcodegen.JSONRPCServiceSnapshot, bool) {
	p.requireLinked()
	planned, ok := p.servicesByExpr[service]
	if !ok {
		return httpcodegen.JSONRPCServiceSnapshot{}, false
	}
	return p.http.JSONRPCService(planned.name)
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
			viewed, hasViewedResult := p.http.ViewedResult(planned.name, endpoint.Method.Name)
			plannedEndpoint := &endpointPlan{
				JSONRPCEndpointSnapshot: endpoint,
				viewed:                  planViewedRepresentation(&endpoint, viewed, hasViewedResult, helper),
			}
			planned.endpoints = append(planned.endpoints, plannedEndpoint)
			if endpoint.SSE != nil {
				planned.hasSSE = true
			} else {
				planned.hasHTTP = true
			}
		}
	}
	p.server = serverFiles(p.services)
	p.client = clientFiles(p.services)
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

// ServerFiles builds runnable servers that mount the saved ordinary HTTP and
// JSON-RPC services for each copied server.
func (p *ExamplePlan) ServerFiles() []*codegen.File {
	p.transport.requireLinked()
	return p.http.CombinedServerFiles(p.transport.applicationHTTP)
}

// CLIFiles builds runnable JSON-RPC clients for each copied server.
func (p *ExamplePlan) CLIFiles() []*codegen.File {
	p.transport.requireLinked()
	return p.http.CLIFiles()
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
		variable:     viewed.Variable,
		fixedView:    viewed.FixedView,
		decode:       helpers.decode,
		encode:       helpers.encode,
		streamEncode: helpers.streamEncode,
		viewedResult: viewed.Service,
		servicePkg:   endpoint.ServicePkgName,
		resultRef:    endpoint.Result.Ref,
	}
	for _, branch := range viewed.Representations {
		representation.branches = append(representation.branches, viewBranch{
			view:       branch.View,
			resultAttr: branch.ResultAttr,
			serverBody: branch.ServerBody,
			clientBody: branch.ClientBody,
			resultInit: branch.ResultInit,
		})
	}
	return representation
}

// servicePlanForOutput copies the service package qualifiers written by one
// JSON-RPC client or server package. The two packages reserve imports
// independently, so a standard-library name used only by the server may suffix
// the generated service import only on that side.
func servicePlanForOutput(planned *servicePlan, client bool) *servicePlan {
	copy := *planned
	copy.data = planned.data
	serviceImport := planned.data.ServerServiceImport()
	if client {
		serviceImport = planned.data.ClientServiceImport()
	}
	copy.data.Service.PkgName = serviceImport.Name
	copy.data.Endpoints = make([]httpcodegen.JSONRPCEndpointSnapshot, len(planned.data.Endpoints))
	copy.endpoints = make([]*endpointPlan, len(planned.endpoints))
	for index, endpoint := range planned.endpoints {
		endpointCopy := *endpoint
		endpointCopy.ServicePkgName = serviceImport.Name
		if endpoint.viewed != nil {
			viewedCopy := *endpoint.viewed
			viewedCopy.servicePkg = serviceImport.Name
			endpointCopy.viewed = &viewedCopy
		}
		copy.endpoints[index] = &endpointCopy
		copy.data.Endpoints[index] = endpointCopy.JSONRPCEndpointSnapshot
	}
	return &copy
}

// requireLinked stops callers from reading files before Link has built them.
func (p *Plan) requireLinked() {
	if !p.linked {
		panic("JSON-RPC files requested before Plan.Link")
	}
}

// planImports records every import name written directly into JSON-RPC files.
func planImports(generation *codegen.Generation, inputs []PlanInput) error {
	clientImports := []*codegen.ImportSpec{
		codegen.SimpleImport("bufio"),
		codegen.SimpleImport("sync"),
		codegen.GoaImport("jsonrpc"),
	}
	serverImports := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("mime"),
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
	}
	for _, input := range inputs {
		for _, transport := range input.Root.API.JSONRPC.Services {
			serviceImport, _, err := input.Service.ServicePackageImports(transport.ServiceExpr)
			if err != nil {
				return err
			}
			pathName := path.Base(serviceImport.Path)
			for index, outputPackage := range []*codegen.GeneratedPackage{
				generation.Package(path.Join(generation.GenPkg(), "jsonrpc", pathName, "client")),
				generation.Package(path.Join(generation.GenPkg(), "jsonrpc", pathName, "server")),
			} {
				imports := clientImports
				if index == 1 {
					imports = serverImports
				}
				for _, spec := range imports {
					if err := outputPackage.RequireImport(spec); err != nil {
						return err
					}
				}
				if err := outputPackage.ReserveGeneratedImport(serviceImport); err != nil {
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
func collectServicePlan(generation *codegen.Generation, input PlanInput, transport *expr.HTTPServiceExpr) (*servicePlan, error) {
	serviceImport, _, err := input.Service.ServicePackageImports(transport.ServiceExpr)
	if err != nil {
		return nil, err
	}
	pathName := path.Base(serviceImport.Path)
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
		api:      input.Root.API.Name,
		name:     transport.Name(),
		pathName: pathName,
		helpers:  make(map[string]*viewedHelperDeclarations),
	}
	declare := func(pkg *codegen.GeneratedPackage, kind codegen.PackageNameKind, preferred string, role uint8) (*codegen.NameDeclaration, error) {
		declaration := codegen.NewPreferredName(kind, preferred, codegen.UnexportedName, jsonRPCNameOrder{
			api:     planned.api,
			service: planned.name,
			role:    role,
		})
		if err := pkg.DeclareName(declaration); err != nil {
			return nil, err
		}
		return declaration, nil
	}
	hasHTTP, hasSSE := false, false
	for _, endpoint := range transport.HTTPEndpoints {
		if endpoint.UsesSSE() {
			hasSSE = true
		} else {
			hasHTTP = true
		}
	}
	planned.clientNames.bufferPool, err = declare(client, codegen.NameVariable, "bufferPool", jsonRPCBufferPoolRole)
	if err != nil {
		return nil, err
	}
	planned.serverNames.encodeError, err = declare(server, codegen.NameFunction, "encodeJSONRPCError", jsonRPCEncodeErrorRole)
	if err != nil {
		return nil, err
	}
	planned.serverNames.noOutputWriter, err = declare(server, codegen.NameType, "noOutputResponseWriter", jsonRPCNoOutputWriterRole)
	if err != nil {
		return nil, err
	}
	if hasHTTP {
		planned.serverNames.batchWriter, err = declare(server, codegen.NameType, "batchWriter", jsonRPCBatchWriterRole)
		if err != nil {
			return nil, err
		}
	}
	if hasSSE {
		planned.serverNames.sseStream, err = declare(server, codegen.NameType, "sseServerStream", jsonRPCSSEStreamRole)
		if err != nil {
			return nil, err
		}
		planned.serverNames.sseBuffer, err = declare(server, codegen.NameType, "sseEventBuffer", jsonRPCSSEBufferRole)
		if err != nil {
			return nil, err
		}
	}
	for _, endpoint := range transport.HTTPEndpoints {
		method := endpoint.MethodExpr
		if _, ok := method.Result.Type.(*expr.ResultTypeExpr); !ok {
			continue
		}
		if planned.bodyDecoder == nil {
			planned.bodyDecoder = codegen.NewPreferredName(
				codegen.NameFunction,
				"decodeJSONRPCResult",
				codegen.UnexportedName,
				jsonRPCNameOrder{api: planned.api, service: planned.name, role: viewedBodyDecoderRole},
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
				jsonRPCNameOrder{api: planned.api, service: planned.name, method: method.Name, role: viewedResultDecoderRole},
			),
			encode: codegen.NewPreferredName(
				codegen.NameFunction,
				"encode"+methodName+"ViewedResult",
				codegen.UnexportedName,
				jsonRPCNameOrder{api: planned.api, service: planned.name, method: method.Name, role: viewedResultEncoderRole},
			),
		}
		if method.IsResultStreaming() {
			helpers.streamEncode = codegen.NewPreferredName(
				codegen.NameFunction,
				"encode"+methodName+"Result",
				codegen.UnexportedName,
				jsonRPCNameOrder{api: planned.api, service: planned.name, method: method.Name, role: viewedStreamEncoderRole},
			)
		}
		if err := client.DeclareName(helpers.decode); err != nil {
			return nil, err
		}
		if err := server.DeclareName(helpers.encode); err != nil {
			return nil, err
		}
		if helpers.streamEncode != nil {
			if err := server.DeclareName(helpers.streamEncode); err != nil {
				return nil, err
			}
		}
		planned.helpers[method.Name] = helpers
	}
	return planned, nil
}

// ComparePackageName orders Go declarations by service, method, and use.
func (o jsonRPCNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(jsonRPCNameOrder)
	if compared := strings.Compare(o.api, right.api); compared != 0 {
		return compared
	}
	if compared := strings.Compare(o.service, right.service); compared != 0 {
		return compared
	}
	if compared := strings.Compare(o.method, right.method); compared != 0 {
		return compared
	}
	return int(o.role) - int(right.role)
}
