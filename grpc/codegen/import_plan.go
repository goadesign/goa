// This file records every import in the generated gRPC package that writes
// the reference. Package-local planning keeps an unrelated transport or
// executable from changing a gRPC qualifier.
package codegen

import (
	"fmt"
	"path"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// grpcFileImportInput lists the packages and service types emitted by one
	// generated gRPC file.
	grpcFileImportInput struct {
		required        []*codegen.ImportSpec
		generated       []*codegen.ImportSpec
		design          []*codegen.ImportSpec
		typeDefinitions []*expr.AttributeExpr
		typeReferences  []*expr.AttributeExpr
	}

	// grpcImportFacts records the template branches that add fixed imports to
	// one service's generated files.
	grpcImportFacts struct {
		hasEndpoints            bool
		hasErrors               bool
		hasStreaming            bool
		serverFileUsesErrors    bool
		streamsPayload          bool
		hasViewedResult         bool
		hasViewedClientStream   bool
		hasUnaryViewedResult    bool
		hasCallerViewedStream   bool
		clientCodecUsesMetadata bool
		serverCodecUsesGoa      bool
		responseMetadataUsesGoa bool
		metadataUsesStrconv     bool
		requestMetadataUsesAny  bool
		responseMetadataUsesAny bool
		usesBearerMetadata      bool
		usesNonBasicMetadata    bool
	}

	// grpcMetadataImportFacts records which fixed packages are used by the
	// generated code for one metadata group.
	grpcMetadataImportFacts struct {
		decodeUsesGoa bool
		usesStrconv   bool
		usesAny       bool
	}
)

// collectGRPCImportFacts evaluates the service branches that determine fixed
// imports once, before package names become final.
func collectGRPCImportFacts(servicePlan *grpcServicePlan) grpcImportFacts {
	serviceExpr := servicePlan.expression
	facts := grpcImportFacts{hasEndpoints: len(serviceExpr.GRPCEndpoints) > 0}
	for index, endpoint := range serviceExpr.GRPCEndpoints {
		method := endpoint.MethodExpr
		facts.hasStreaming = facts.hasStreaming || method.IsStreaming()
		if len(endpoint.GRPCErrors) > 0 {
			facts.hasErrors = true
			facts.serverFileUsesErrors = true
		}
		if usesStreamEnvelope(endpoint) {
			facts.serverFileUsesErrors = true
			facts.serverCodecUsesGoa = true
		}
		if method.IsPayloadStreaming() && !isEmpty(endpoint.Request.Type) {
			facts.streamsPayload = true
		}
		if method.Payload.Type != expr.Empty || method.Result.Type != expr.Empty || method.IsStreaming() {
			facts.clientCodecUsesMetadata = true
		}

		viewed := grpcMethodHasViewedResult(method)
		facts.hasViewedResult = facts.hasViewedResult || viewed
		facts.hasViewedClientStream = facts.hasViewedClientStream || viewed && method.IsResultStreaming()
		facts.hasUnaryViewedResult = facts.hasUnaryViewedResult || viewed && !method.IsResultStreaming()
		callerView := grpcMethodUsesCallerSelectedView(method)
		facts.hasCallerViewedStream = facts.hasCallerViewedStream || callerView && method.IsResultStreaming()
		facts.serverCodecUsesGoa = facts.serverCodecUsesGoa || callerView

		plannedMetadata := servicePlan.endpoints[index].metadata
		request := collectGRPCMetadataImportFacts(plannedMetadata[endpoint.Metadata])
		header := collectGRPCMetadataImportFacts(plannedMetadata[endpoint.Response.Headers])
		trailer := collectGRPCMetadataImportFacts(plannedMetadata[endpoint.Response.Trailers])
		facts.serverCodecUsesGoa = facts.serverCodecUsesGoa || request.decodeUsesGoa
		facts.responseMetadataUsesGoa = facts.responseMetadataUsesGoa || header.decodeUsesGoa || trailer.decodeUsesGoa
		facts.metadataUsesStrconv = facts.metadataUsesStrconv || request.usesStrconv || header.usesStrconv || trailer.usesStrconv
		facts.requestMetadataUsesAny = facts.requestMetadataUsesAny || request.usesAny
		facts.responseMetadataUsesAny = facts.responseMetadataUsesAny || header.usesAny || trailer.usesAny

		for _, requirement := range endpoint.Requirements {
			for _, authored := range requirement.Schemes {
				scheme := service.BuildSchemeData(authored, method)
				if scheme == nil {
					continue
				}
				facts.usesBearerMetadata = facts.usesBearerMetadata || scheme.Type == "Bearer" || scheme.Type == "JWT" || scheme.Type == "OAuth2"
				facts.usesNonBasicMetadata = facts.usesNonBasicMetadata || scheme.Type != "Basic"
			}
		}
	}
	return facts
}

// collectGRPCMetadataImportFacts evaluates the template branches used to
// encode and decode one metadata group.
func collectGRPCMetadataImportFacts(metadata []*grpcMetadataPlan) grpcMetadataImportFacts {
	var facts grpcMetadataImportFacts
	for _, field := range metadata {
		kind := grpcMetadataElementKind(field.wire)
		facts.decodeUsesGoa = facts.decodeUsesGoa || field.required || kind != expr.StringKind && kind != expr.AnyKind
		facts.usesStrconv = facts.usesStrconv || grpcMetadataKindUsesStrconv(kind)
		facts.usesAny = facts.usesAny || kind == expr.AnyKind
	}
	return facts
}

// planGRPCImports records imports for each generated client, server, and
// command-parser package before Generation.Freeze chooses their names.
func planGRPCImports(generation *codegen.Generation, plan *Plan) error {
	for _, servicePlan := range plan.servicesPlan {
		service := servicePlan.expression
		facts := collectGRPCImportFacts(servicePlan)
		pathName := servicePlan.packages.pathName
		clientPath := path.Join(generation.GenPkg(), "grpc", pathName, "client")
		serverPath := path.Join(generation.GenPkg(), "grpc", pathName, "server")
		protobufPath := path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName)
		client := generation.Package(clientPath)
		server := generation.Package(serverPath)
		protobufImport := codegen.NewImport(pathName+"pb", protobufPath)
		allReferences := grpcEndpointAttributes(service.GRPCEndpoints...)
		codecReferences := grpcCodecAttributes(service.GRPCEndpoints...)
		payloadReferences := grpcPayloadAttributes(service.GRPCEndpoints...)
		streamReferences := grpcStreamAttributes(service.GRPCEndpoints...)
		hasReferences := grpcAttributesHaveValues(allReferences)
		hasCodecReferences := grpcAttributesHaveValues(codecReferences)
		hasPayloadReferences := grpcAttributesHaveValues(payloadReferences)

		clientFileRequired := []*codegen.ImportSpec{codegen.SimpleImport("google.golang.org/grpc")}
		if facts.hasEndpoints {
			clientFileRequired = append(clientFileRequired,
				codegen.SimpleImport("context"),
				codegen.GoaImport(""),
				codegen.GoaNamedImport("grpc", "goagrpc"),
				codegen.GoaNamedImport("grpc/pb", "goapb"),
			)
		}
		clientFileGenerated := []*codegen.ImportSpec{protobufImport}
		if facts.hasStreaming {
			clientFileGenerated = append(clientFileGenerated, servicePlan.packages.service)
		}
		if facts.hasViewedClientStream {
			clientFileGenerated = append(clientFileGenerated, servicePlan.packages.views)
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join(codegen.Gendir, "grpc", pathName, "client", "client.go"),
			client,
			grpcFileImportInput{
				required:        clientFileRequired,
				generated:       clientFileGenerated,
				typeDefinitions: streamReferences,
			},
		); err != nil {
			return err
		}

		var clientCodecRequired []*codegen.ImportSpec
		if facts.hasEndpoints {
			clientCodecRequired = append(clientCodecRequired,
				codegen.SimpleImport("context"),
				codegen.GoaNamedImport("grpc", "goagrpc"),
				codegen.SimpleImport("google.golang.org/grpc"),
			)
		}
		if facts.clientCodecUsesMetadata {
			clientCodecRequired = append(clientCodecRequired, codegen.SimpleImport("google.golang.org/grpc/metadata"))
		}
		if facts.responseMetadataUsesGoa {
			clientCodecRequired = append(clientCodecRequired, codegen.GoaImport(""))
		}
		if facts.metadataUsesStrconv {
			clientCodecRequired = append(clientCodecRequired, codegen.SimpleImport("strconv"))
		}
		if facts.requestMetadataUsesAny {
			clientCodecRequired = append(clientCodecRequired, codegen.SimpleImport("fmt"))
		}
		if facts.usesBearerMetadata {
			clientCodecRequired = append(clientCodecRequired, codegen.SimpleImport("strings"))
		}
		var clientCodecGenerated []*codegen.ImportSpec
		if facts.hasEndpoints {
			clientCodecGenerated = append(clientCodecGenerated, protobufImport)
		}
		if hasCodecReferences {
			clientCodecGenerated = append(clientCodecGenerated, servicePlan.packages.service)
		}
		if facts.hasUnaryViewedResult {
			clientCodecGenerated = append(clientCodecGenerated, servicePlan.packages.views)
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join(codegen.Gendir, "grpc", pathName, "client", "encode_decode.go"),
			client,
			grpcFileImportInput{
				required:       clientCodecRequired,
				generated:      clientCodecGenerated,
				typeReferences: codecReferences,
			},
		); err != nil {
			return err
		}

		clientTypesRequired := grpcValidationRuntimeImportSpecs(plan.protobuf[service], validateClient)
		if servicePlan.usesAnyInErrors {
			clientTypesRequired = append(clientTypesRequired,
				codegen.SimpleImport("fmt"),
				codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"),
			)
		}
		var clientTypesGenerated []*codegen.ImportSpec
		if hasReferences {
			clientTypesGenerated = append(clientTypesGenerated, servicePlan.packages.service)
		}
		if facts.hasEndpoints {
			clientTypesGenerated = append(clientTypesGenerated, protobufImport)
		}
		if facts.hasViewedResult {
			clientTypesGenerated = append(clientTypesGenerated, servicePlan.packages.views)
		}
		var clientTypesDesign []*codegen.ImportSpec
		if hasReferences {
			clientTypesDesign = servicePlan.protoGoImports
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join(codegen.Gendir, "grpc", pathName, "client", "types.go"),
			client,
			grpcFileImportInput{
				required:       clientTypesRequired,
				generated:      clientTypesGenerated,
				design:         clientTypesDesign,
				typeReferences: allReferences,
			},
		); err != nil {
			return err
		}

		if facts.hasEndpoints {
			clientCLIRequired := grpcPayloadBuilderImportPreferences(servicePlan, plan.protobuf[service])
			var clientCLIGenerated []*codegen.ImportSpec
			if hasPayloadReferences {
				clientCLIGenerated = append(clientCLIGenerated, servicePlan.packages.service)
			}
			if grpcServiceHasCLIMessage(plan.protobuf[service]) {
				clientCLIGenerated = append(clientCLIGenerated, protobufImport)
			}
			if err := recordGRPCFileImports(
				plan,
				path.Join(codegen.Gendir, "grpc", pathName, "client", "cli.go"),
				client,
				grpcFileImportInput{
					required:       clientCLIRequired,
					generated:      clientCLIGenerated,
					typeReferences: payloadReferences,
				},
			); err != nil {
				return err
			}
		}

		var serverFileRequired []*codegen.ImportSpec
		if facts.hasEndpoints {
			serverFileRequired = append(serverFileRequired,
				codegen.SimpleImport("context"),
				codegen.GoaImport(""),
				codegen.GoaNamedImport("grpc", "goagrpc"),
			)
		}
		if facts.hasErrors {
			serverFileRequired = append(serverFileRequired, codegen.SimpleImport("google.golang.org/grpc/codes"))
		}
		if facts.serverFileUsesErrors {
			serverFileRequired = append(serverFileRequired, codegen.SimpleImport("errors"))
		}
		if facts.streamsPayload {
			serverFileRequired = append(serverFileRequired, codegen.SimpleImport("io"))
		}
		if facts.hasCallerViewedStream {
			serverFileRequired = append(serverFileRequired, codegen.SimpleImport("google.golang.org/grpc/metadata"))
		}
		serverFileGenerated := []*codegen.ImportSpec{servicePlan.packages.service, protobufImport}
		if err := recordGRPCFileImports(
			plan,
			path.Join(codegen.Gendir, "grpc", pathName, "server", "server.go"),
			server,
			grpcFileImportInput{
				required:        serverFileRequired,
				generated:       serverFileGenerated,
				typeDefinitions: streamReferences,
			},
		); err != nil {
			return err
		}

		var serverCodecRequired []*codegen.ImportSpec
		if facts.hasEndpoints {
			serverCodecRequired = append(serverCodecRequired,
				codegen.SimpleImport("context"),
				codegen.GoaNamedImport("grpc", "goagrpc"),
				codegen.SimpleImport("google.golang.org/grpc/metadata"),
			)
		}
		if facts.serverCodecUsesGoa {
			serverCodecRequired = append(serverCodecRequired, codegen.GoaImport(""))
		}
		if facts.metadataUsesStrconv {
			serverCodecRequired = append(serverCodecRequired, codegen.SimpleImport("strconv"))
		}
		if facts.usesNonBasicMetadata {
			serverCodecRequired = append(serverCodecRequired, codegen.SimpleImport("strings"))
		}
		if facts.responseMetadataUsesAny {
			serverCodecRequired = append(serverCodecRequired, codegen.SimpleImport("fmt"))
		}
		var serverCodecGenerated []*codegen.ImportSpec
		if facts.hasEndpoints {
			serverCodecGenerated = append(serverCodecGenerated, protobufImport)
		}
		if hasCodecReferences {
			serverCodecGenerated = append(serverCodecGenerated, servicePlan.packages.service)
		}
		if facts.hasViewedResult {
			serverCodecGenerated = append(serverCodecGenerated, servicePlan.packages.views)
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join(codegen.Gendir, "grpc", pathName, "server", "encode_decode.go"),
			server,
			grpcFileImportInput{
				required:       serverCodecRequired,
				generated:      serverCodecGenerated,
				typeReferences: codecReferences,
			},
		); err != nil {
			return err
		}

		serverTypesRequired := grpcValidationRuntimeImportSpecs(plan.protobuf[service], validateServer)
		if servicePlan.usesAnyInErrors {
			serverTypesRequired = append(serverTypesRequired,
				codegen.SimpleImport("fmt"),
				codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"),
			)
		}
		var serverTypesGenerated []*codegen.ImportSpec
		if hasReferences {
			serverTypesGenerated = append(serverTypesGenerated, servicePlan.packages.service)
		}
		if facts.hasEndpoints {
			serverTypesGenerated = append(serverTypesGenerated, protobufImport)
		}
		if facts.hasViewedResult {
			serverTypesGenerated = append(serverTypesGenerated, servicePlan.packages.views)
		}
		var serverTypesDesign []*codegen.ImportSpec
		if hasReferences {
			serverTypesDesign = servicePlan.protoGoImports
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join(codegen.Gendir, "grpc", pathName, "server", "types.go"),
			server,
			grpcFileImportInput{
				required:       serverTypesRequired,
				generated:      serverTypesGenerated,
				design:         serverTypesDesign,
				typeReferences: allReferences,
			},
		); err != nil {
			return err
		}
	}

	for _, serverPlan := range plan.cli.servers {
		serverName := codegen.SnakeCase(codegen.Goify(serverPlan.name, true))
		outputPath := path.Join(generation.GenPkg(), "grpc", "cli", serverName)
		output := generation.Package(outputPath)
		required := []*codegen.ImportSpec{
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("os"),
			codegen.GoaImport(""),
			codegen.NewImport("grpc", "google.golang.org/grpc"),
		}
		var generated []*codegen.ImportSpec
		for _, serviceName := range serverPlan.expression.Services {
			servicePlan := grpcServicePlanByName(plan.servicesPlan, serviceName)
			if servicePlan == nil {
				continue
			}
			pathName := servicePlan.packages.pathName
			generated = append(generated,
				codegen.NewImport(servicePlan.packages.service.Name+"c", path.Join(generation.GenPkg(), "grpc", pathName, "client")),
			)
			if len(servicePlan.source.ServiceExpr.ClientInterceptors) > 0 {
				generated = append(generated, servicePlan.packages.service)
			}
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join(codegen.Gendir, "grpc", "cli", serverName, "cli.go"),
			output,
			grpcFileImportInput{required: required, generated: generated},
		); err != nil {
			return err
		}
	}
	return nil
}

// recordGRPCFileImports registers the imports used by one generated file and
// saves their paths so the header can use the chosen names after freeze.
func recordGRPCFileImports(
	plan *Plan,
	filePath string,
	output *codegen.GeneratedPackage,
	input grpcFileImportInput,
) error {
	key := grpcFilePathKey(filePath)
	if plan.fileImports[key] != nil {
		return fmt.Errorf("gRPC file %q already has an import plan", filePath)
	}
	imports := codegen.NewGeneratedImportPlan(output)
	if err := imports.Require(input.required...); err != nil {
		return err
	}
	if err := imports.AddGenerated(input.generated...); err != nil {
		return err
	}
	if err := imports.AddDesign(input.design...); err != nil {
		return err
	}
	if err := imports.AddTypeExpressions(input.typeDefinitions...); err != nil {
		return err
	}
	if err := imports.AddRecursiveTypeReferences(input.typeReferences...); err != nil {
		return err
	}
	plan.fileImports[key] = imports
	return nil
}

// planGRPCExampleImports adds the gRPC files' imports to the executable
// packages already claimed by the shared example planner.
func planGRPCExampleImports(generation *codegen.Generation, plan *Plan, root *example.Root) error {
	rootPath := path.Dir(generation.GenPkg())
	for _, server := range root.Servers {
		serverPath := path.Join(rootPath, "cmd", server.Dir)
		serverPackage, err := generation.ClaimOutputPackage(serverPath, path.Join("cmd", server.Dir))
		if err != nil {
			return err
		}
		serverImports := grpcFileImportInput{required: []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("net"),
			codegen.SimpleImport("net/url"),
			codegen.SimpleImport("sync"),
			codegen.SimpleImport("goa.design/clue/debug"),
			codegen.SimpleImport("goa.design/clue/log"),
			codegen.SimpleImport("google.golang.org/grpc"),
			codegen.SimpleImport("google.golang.org/grpc/reflection"),
		}}
		for _, serviceName := range server.Services {
			servicePlan := grpcServicePlanByName(plan.servicesPlan, serviceName)
			if servicePlan == nil {
				continue
			}
			pathName := servicePlan.packages.pathName
			serverImports.generated = append(serverImports.generated,
				codegen.NewImport(servicePlan.packages.service.Name+"svr", path.Join(generation.GenPkg(), "grpc", pathName, "server")),
				codegen.NewImport(pathName+"pb", path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName)),
			)
			if len(servicePlan.expression.GRPCEndpoints) > 0 {
				serverImports.generated = append(serverImports.generated, servicePlan.packages.service)
			}
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join("cmd", server.Dir, "grpc.go"),
			serverPackage,
			serverImports,
		); err != nil {
			return err
		}

		if server.DefaultTransport() == nil {
			continue
		}
		clientPath := path.Join(rootPath, "cmd", server.Dir+"-cli")
		clientPackage, err := generation.ClaimOutputPackage(clientPath, path.Join("cmd", server.Dir+"-cli"))
		if err != nil {
			return err
		}
		clientImports := grpcFileImportInput{required: []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("errors"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("io"),
			codegen.SimpleImport("google.golang.org/grpc"),
			codegen.SimpleImport("google.golang.org/grpc/credentials/insecure"),
		}, generated: []*codegen.ImportSpec{codegen.NewImport(
			"cli",
			path.Join(generation.GenPkg(), "grpc", "cli", server.Dir),
		)}}
		hasEndpoints := false
		for _, serviceName := range server.Services {
			service := plan.root.API.GRPC.Service(serviceName)
			if service == nil {
				continue
			}
			hasEndpoints = hasEndpoints || len(service.GRPCEndpoints) > 0
			if grpcServiceStreamsResult(service) {
				clientImports.generated = append(clientImports.generated, plan.packages[service].service)
			}
			servicePlan := grpcServicePlanByName(plan.servicesPlan, serviceName)
			if servicePlan != nil && len(servicePlan.source.ServiceExpr.ClientInterceptors) > 0 {
				clientImports.generated = append(clientImports.generated, codegen.NewImport("interceptors", rootPath+"/interceptors"))
			}
		}
		if hasEndpoints {
			clientImports.required = append(clientImports.required, codegen.SimpleImport("flag"))
		}
		if err := recordGRPCFileImports(
			plan,
			path.Join("cmd", server.Dir+"-cli", "grpc.go"),
			clientPackage,
			clientImports,
		); err != nil {
			return err
		}
	}
	return nil
}

// grpcMethodHasViewedResult reports whether a method returns a result type
// projected through a Goa view.
func grpcMethodHasViewedResult(method *expr.MethodExpr) bool {
	_, ok := method.Result.Type.(*expr.ResultTypeExpr)
	return ok
}

// grpcMethodUsesCallerSelectedView reports whether the caller supplies the
// view name for a result type with more than one available view.
func grpcMethodUsesCallerSelectedView(method *expr.MethodExpr) bool {
	result, ok := method.Result.Type.(*expr.ResultTypeExpr)
	if !ok || !result.HasMultipleViews() {
		return false
	}
	_, fixed := method.Result.Meta.Last(expr.ViewMetaKey)
	return !fixed
}

// grpcValidationRuntimeImportSpecs returns the standard packages used by the
// validation functions written on one side of a gRPC service.
func grpcValidationRuntimeImportSpecs(protobuf *protobufServicePlan, side validateKind) []*codegen.ImportSpec {
	var imports []*codegen.ImportSpec
	for _, validator := range protobuf.catalog.validators {
		if validator.side != side {
			continue
		}
		for _, runtimeImport := range codegen.ValidationRuntimeImports(
			validator.attribute,
			codegen.GoLayoutPolicy{Pointer: true},
		) {
			imports = appendGRPCImportSpec(imports, codegen.NewImport(runtimeImport.Name, runtimeImport.Path))
		}
	}
	return imports
}

// grpcPayloadBuilderImportPreferences returns the fixed packages used to turn
// command-line flag text into one service's gRPC payloads.
func grpcPayloadBuilderImportPreferences(servicePlan *grpcServicePlan, protobuf *protobufServicePlan) []*codegen.ImportSpec {
	var imports []*codegen.ImportSpec
	for index, endpoint := range servicePlan.endpoints {
		request := protobuf.messages[index].request
		if protobufCLIRequestNeedsMessage(request) {
			imports = appendGRPCImportSpec(imports, codegen.SimpleImport("fmt"))
			imports = appendGRPCImportSpec(imports, codegen.SimpleImport("google.golang.org/protobuf/encoding/protojson"))
		}
		for _, metadata := range endpoint.metadata[endpoint.expression.Metadata] {
			preferences := cli.FlagImportPreferences(metadata.wire, metadata.validation != "")
			for _, preference := range preferences {
				imports = appendGRPCImportSpec(imports, preference)
				switch preference.Path {
				case "encoding/json", "strconv":
					imports = appendGRPCImportSpec(imports, codegen.SimpleImport("fmt"))
				}
			}
			if metadata.required && metadata.wire.DefaultValue == nil {
				imports = appendGRPCImportSpec(imports, codegen.SimpleImport("fmt"))
			}
		}
	}
	if grpcServicePayloadUsesAny(servicePlan.expression) {
		imports = appendGRPCImportSpec(imports, codegen.SimpleImport("fmt"))
		imports = appendGRPCImportSpec(imports, codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"))
	}
	return imports
}

// grpcServiceHasCLIMessage reports whether a payload builder parses a
// protobuf request before constructing the service payload.
func grpcServiceHasCLIMessage(protobuf *protobufServicePlan) bool {
	for _, messages := range protobuf.messages {
		if protobufCLIRequestNeedsMessage(messages.request) {
			return true
		}
	}
	return false
}

// grpcServicePayloadUsesAny reports whether a payload builder converts an Any
// value into a protobuf request field.
func grpcServicePayloadUsesAny(service *expr.GRPCServiceExpr) bool {
	for _, endpoint := range service.GRPCEndpoints {
		if hasAnyType(endpoint.MethodExpr.Payload) {
			return true
		}
	}
	return false
}

// grpcMetadataElementKind returns the kind formatted for one metadata value.
func grpcMetadataElementKind(attribute *expr.AttributeExpr) expr.Kind {
	if array := expr.AsArray(attribute.Type); array != nil {
		return array.ElemType.Type.Kind()
	}
	return attribute.Type.Kind()
}

// grpcMetadataKindUsesStrconv reports whether metadata templates call the
// standard strconv package for one primitive kind.
func grpcMetadataKindUsesStrconv(kind expr.Kind) bool {
	switch kind {
	case expr.BooleanKind,
		expr.IntKind,
		expr.Int32Kind,
		expr.Int64Kind,
		expr.UIntKind,
		expr.UInt32Kind,
		expr.UInt64Kind,
		expr.Float32Kind,
		expr.Float64Kind:
		return true
	default:
		return false
	}
}

// appendGRPCImportSpec adds one complete import path if it is not already in
// the list.
func appendGRPCImportSpec(imports []*codegen.ImportSpec, spec *codegen.ImportSpec) []*codegen.ImportSpec {
	for _, existing := range imports {
		if existing.Path == spec.Path {
			return imports
		}
	}
	return append(imports, spec)
}

func grpcServiceStreamsResult(service *expr.GRPCServiceExpr) bool {
	for _, endpoint := range service.GRPCEndpoints {
		if endpoint.MethodExpr.IsResultStreaming() && !endpoint.MethodExpr.IsPayloadStreaming() {
			return true
		}
	}
	return false
}

func grpcServicePlanByName(services []*grpcServicePlan, name string) *grpcServicePlan {
	for _, service := range services {
		if service.expression.Name() == name {
			return service
		}
	}
	return nil
}
