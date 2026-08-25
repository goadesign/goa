// This file copies each gRPC service and endpoint while NewPlans can still
// read the evaluated design. Link and the file builders read these copies.
package codegen

import (
	"fmt"
	"path"
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// grpcServicePlan stores one copied gRPC service and the values selected for
	// its generated files.
	grpcServicePlan struct {
		source          *expr.GRPCServiceExpr
		expression      *expr.GRPCServiceExpr
		packages        *grpcServicePackage
		endpoints       []*grpcEndpointPlan
		endpointByExpr  map[*expr.GRPCEndpointExpr]*grpcEndpointPlan
		imports         []string
		protoImports    []string
		protoGoImports  []*codegen.ImportSpec
		scope           *codegen.NameScope
		usesAny         bool
		usesAnyInErrors bool
	}

	// grpcEndpointPlan stores one copied endpoint and the metadata conversions
	// that must be prepared before generated names are available.
	grpcEndpointPlan struct {
		expression     *expr.GRPCEndpointExpr
		legacyStream   bool
		legacyMetadata *expr.MappedAttributeExpr
		metadata       map[*expr.MappedAttributeExpr][]*grpcMetadataPlan
	}

	// grpcMetadataPlan stores one metadata field, its Go field, and the code that
	// converts the value in both directions.
	grpcMetadataPlan struct {
		name         string
		element      string
		required     bool
		fieldName    string
		fieldType    expr.DataType
		pointer      bool
		serviceField *expr.AttributeExpr
		wire         *expr.AttributeExpr
		scope        *codegen.NameScope
		validation   string
		encode       *codegen.TransformPlan
		decode       *codegen.TransformPlan
	}
)

// collectGRPCServicePlans copies the services selected for one gRPC plan and
// records the imports and metadata conversions used by their generated files.
func collectGRPCServicePlans(generation *codegen.Generation, plan *Plan) ([]*grpcServicePlan, error) {
	services := make([]*grpcServicePlan, len(plan.expressions))
	for index, source := range plan.expressions {
		service, err := copyGRPCService(source)
		if err != nil {
			return nil, err
		}
		service.scope = codegen.NewNameScope()
		service.usesAny = usesAnyType(service.expression.GRPCEndpoints, false)
		service.usesAnyInErrors = usesAnyType(service.expression.GRPCEndpoints, true)
		service.imports = grpcServiceImportPaths(generation, service.expression)
		service.packages = plan.packages[source]

		plannedProtobuf := plan.protobuf[source]
		plannedTools := plan.tools[source]
		plannedSymbols := plan.symbols[source]
		service.protoImports, service.protoGoImports = collectGRPCProtobufImports(plannedProtobuf)
		plan.protobuf[service.expression] = plannedProtobuf
		plan.tools[service.expression] = plannedTools
		plan.symbols[service.expression] = plannedSymbols
		for endpointIndex, endpoint := range service.endpoints {
			sourceEndpoint := source.GRPCEndpoints[endpointIndex]
			if plannedSymbols != nil {
				plannedSymbols.endpoints[endpoint.expression] = plannedSymbols.endpoints[sourceEndpoint]
			}
			if plannedProtobuf != nil {
				plannedProtobuf.methods[endpoint.expression] = plannedProtobuf.methods[sourceEndpoint]
			}
			if declaration := plan.cli.builders[sourceEndpoint]; declaration != nil {
				plan.cli.builders[endpoint.expression] = declaration
			}
			if err := planEndpointMetadata(endpoint); err != nil {
				return nil, fmt.Errorf("plan gRPC metadata for %q.%q: %w", service.expression.Name(), endpoint.expression.Name(), err)
			}
		}
		if err := replaceGRPCTransforms(plan.service, source, service, plannedProtobuf, plannedSymbols); err != nil {
			return nil, fmt.Errorf("copy gRPC conversions for service %q: %w", service.expression.Name(), err)
		}
		services[index] = service
	}
	return services, nil
}

// collectGRPCProtobufImports records the protobuf schema files and Go packages
// selected by every saved protobuf message.
func collectGRPCProtobufImports(protobuf *protobufServicePlan) ([]string, []*codegen.ImportSpec) {
	if protobuf == nil {
		return nil, nil
	}
	var attributes []*expr.AttributeExpr
	for _, messages := range protobuf.messages {
		attributes = append(attributes, messages.request, messages.streamingRequest, messages.requestEnvelope, messages.response)
		errorNames := make([]string, 0, len(messages.errors))
		for name := range messages.errors {
			errorNames = append(errorNames, name)
		}
		sort.Strings(errorNames)
		for _, name := range errorNames {
			attributes = append(attributes, messages.errors[name])
		}
	}
	var protoImports []string
	var goImports []*codegen.ImportSpec
	seenProto := make(map[string]struct{})
	seenGo := make(map[string]struct{})
	seenTypes := make(map[expr.UserType]struct{})
	var walk func(*expr.AttributeExpr)
	walk = func(attribute *expr.AttributeExpr) {
		if attribute == nil {
			return
		}
		if field := attribute.Meta["struct:field:proto"]; len(field) > 1 {
			if _, ok := seenProto[field[1]]; !ok {
				seenProto[field[1]] = struct{}{}
				protoImports = append(protoImports, field[1])
			}
			if len(field) > 3 {
				if _, ok := seenGo[field[3]]; !ok {
					seenGo[field[3]] = struct{}{}
					goImports = append(goImports, codegen.NewImport(path.Base(field[3]), field[3]))
				}
			}
		}
		if attribute.Type.Kind() == expr.AnyKind {
			const structProto = "google/protobuf/struct.proto"
			if _, ok := seenProto[structProto]; !ok {
				seenProto[structProto] = struct{}{}
				protoImports = append(protoImports, structProto)
			}
			return
		}
		switch actual := attribute.Type.(type) {
		case expr.UserType:
			if _, ok := seenTypes[actual]; ok {
				return
			}
			seenTypes[actual] = struct{}{}
			walk(actual.Attribute())
		case *expr.Object:
			for _, named := range *actual {
				walk(named.Attribute)
			}
		case *expr.Array:
			walk(actual.ElemType)
		case *expr.Map:
			walk(actual.KeyType)
			walk(actual.ElemType)
		case *expr.Union:
			for _, named := range actual.Values {
				walk(named.Attribute)
			}
		}
	}
	for _, attribute := range attributes {
		walk(attribute)
	}
	return protoImports, goImports
}

// replaceGRPCTransforms rebuilds every conversion from the copied method. This
// prevents later changes to the original design from changing generated code.
func replaceGRPCTransforms(
	servicePlan *service.Plan,
	source *expr.GRPCServiceExpr,
	grpcService *grpcServicePlan,
	protobuf *protobufServicePlan,
	symbols *grpcSymbols,
) error {
	if protobuf == nil || symbols == nil {
		return fmt.Errorf("protobuf messages or Go names are missing")
	}
	replaced := make(map[*grpcConversion]struct{})
	replace := func(conversion *grpcConversion, source, target *expr.AttributeExpr, proto bool) error {
		if conversion == nil {
			return nil
		}
		if _, ok := replaced[conversion]; ok {
			return nil
		}
		transform, err := newGRPCTransformPlan(source, target, proto, protobuf)
		if err != nil {
			return err
		}
		oldDefinitions := conversion.transform.HelperDefinitions()
		newDefinitions := transform.HelperDefinitions()
		if len(oldDefinitions) != len(newDefinitions) {
			return fmt.Errorf("saved conversion helper definition count changed from %d to %d", len(oldDefinitions), len(newDefinitions))
		}
		for index, definition := range newDefinitions {
			if err := transform.BindHelperDefinition(definition.ID, oldDefinitions[index].Declaration); err != nil {
				return err
			}
		}
		conversion.transform = transform
		conversion.bound = false
		replaced[conversion] = struct{}{}
		return nil
	}
	for index, endpointPlan := range grpcService.endpoints {
		endpoint := endpointPlan.expression
		sourceEndpoint := source.GRPCEndpoints[index]
		messages := protobuf.messages[index]
		endpointSymbols := symbols.endpoints[endpoint]
		result := endpoint.MethodExpr.Result
		if _, viewed := sourceEndpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr); viewed {
			projected, err := servicePlan.ProjectedResult(sourceEndpoint.MethodExpr)
			if err != nil {
				return err
			}
			result = expr.DupAtt(projected)
		}
		if endpoint.MethodExpr.Payload.Type != expr.Empty {
			if err := replace(endpointSymbols.serverInits[grpcInitKey{role: grpcRequestInit}], messages.request, endpoint.MethodExpr.Payload, false); err != nil {
				return err
			}
			if err := replace(endpointSymbols.cliPayload, messages.request, endpoint.MethodExpr.Payload, false); err != nil {
				return err
			}
		}
		if !(endpoint.MethodExpr.IsPayloadStreaming() && isEmpty(endpoint.Request.Type)) {
			if err := replace(endpointSymbols.clientInits[grpcInitKey{role: grpcRequestInit}], endpoint.MethodExpr.Payload, messages.request, true); err != nil {
				return err
			}
		}
		if err := replace(endpointSymbols.serverInits[grpcInitKey{role: grpcResponseInit}], result, messages.response, true); err != nil {
			return err
		}
		if endpoint.MethodExpr.Result.Type != expr.Empty && !endpoint.MethodExpr.IsStreaming() {
			if err := replace(endpointSymbols.clientInits[grpcInitKey{role: grpcResponseInit}], messages.response, result, false); err != nil {
				return err
			}
		}
		if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty {
			key := grpcInitKey{role: grpcStreamingRequestInit}
			if err := replace(endpointSymbols.serverInits[key], messages.streamingRequest, endpoint.MethodExpr.StreamingPayload, false); err != nil {
				return err
			}
			if err := replace(endpointSymbols.clientInits[key], endpoint.MethodExpr.StreamingPayload, messages.streamingRequest, true); err != nil {
				return err
			}
		}
		if endpoint.MethodExpr.Result.Type != expr.Empty && endpoint.MethodExpr.IsStreaming() {
			key := grpcInitKey{role: grpcStreamingResponseInit}
			if err := replace(endpointSymbols.serverInits[key], result, messages.response, true); err != nil {
				return err
			}
			if err := replace(endpointSymbols.clientInits[key], messages.response, result, false); err != nil {
				return err
			}
		}
		if endpointPlan.legacyStream && expr.IsObject(endpoint.MethodExpr.Payload.Type) {
			key := grpcInitKey{role: grpcLegacyRequestInit}
			if err := replace(endpointSymbols.serverInits[key], &expr.AttributeExpr{Type: expr.Empty}, endpoint.MethodExpr.Payload, false); err != nil {
				return err
			}
		}
		for _, grpcError := range endpoint.GRPCErrors {
			message := messages.errors[grpcError.Name]
			if message == nil {
				continue
			}
			key := grpcInitKey{role: grpcErrorInit, subject: grpcError.Name}
			if err := replace(endpointSymbols.serverInits[key], grpcError.AttributeExpr, message, true); err != nil {
				return err
			}
			if err := replace(endpointSymbols.clientInits[key], message, grpcError.AttributeExpr, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyGRPCService makes a private copy of the service values read by gRPC
// planning and rendering.
func copyGRPCService(source *expr.GRPCServiceExpr) (*grpcServicePlan, error) {
	serviceExpr := &expr.ServiceExpr{
		Name:        source.ServiceExpr.Name,
		Description: source.ServiceExpr.Description,
		Meta:        copyGRPCMeta(source.ServiceExpr.Meta),
	}
	serviceExpr.Errors = make([]*expr.ErrorExpr, len(source.ServiceExpr.Errors))
	for index, sourceError := range source.ServiceExpr.Errors {
		serviceExpr.Errors[index] = &expr.ErrorExpr{
			Name:          sourceError.Name,
			AttributeExpr: copyGRPCErrorAttribute(sourceError.AttributeExpr),
		}
	}
	service := &expr.GRPCServiceExpr{
		ServiceExpr: serviceExpr,
		ParentName:  source.ParentName,
		ProtoPkg:    source.ProtoPkg,
		Meta:        copyGRPCMeta(source.Meta),
	}
	result := &grpcServicePlan{
		source:         source,
		expression:     service,
		endpoints:      make([]*grpcEndpointPlan, len(source.GRPCEndpoints)),
		endpointByExpr: make(map[*expr.GRPCEndpointExpr]*grpcEndpointPlan, len(source.GRPCEndpoints)),
	}
	for index, sourceEndpoint := range source.GRPCEndpoints {
		method := copyGRPCMethod(sourceEndpoint.MethodExpr, serviceExpr)
		serviceExpr.Methods = append(serviceExpr.Methods, method)
		endpoint := &expr.GRPCEndpointExpr{
			MethodExpr:       method,
			Service:          service,
			Request:          expr.DupAtt(sourceEndpoint.Request),
			StreamingRequest: expr.DupAtt(sourceEndpoint.StreamingRequest),
			Response:         copyGRPCResponse(sourceEndpoint.Response),
			Metadata:         expr.DupMappedAtt(sourceEndpoint.Metadata),
			Requirements:     copyGRPCRequirements(sourceEndpoint.Requirements),
			Meta:             copyGRPCMeta(sourceEndpoint.Meta),
		}
		endpoint.Response.Parent = endpoint
		endpoint.GRPCErrors = make([]*expr.GRPCErrorExpr, len(sourceEndpoint.GRPCErrors))
		for errorIndex, sourceError := range sourceEndpoint.GRPCErrors {
			methodError := method.Error(sourceError.Name)
			if methodError == nil {
				return nil, fmt.Errorf("gRPC error %q is not defined by method %q", sourceError.Name, method.Name)
			}
			response := copyGRPCResponse(sourceError.Response)
			response.Parent = endpoint
			endpoint.GRPCErrors[errorIndex] = &expr.GRPCErrorExpr{
				ErrorExpr: methodError,
				Name:      sourceError.Name,
				Response:  response,
			}
		}
		endpointPlan := &grpcEndpointPlan{
			expression:   endpoint,
			legacyStream: sourceEndpoint.LegacyStreamCompat(),
			metadata:     make(map[*expr.MappedAttributeExpr][]*grpcMetadataPlan),
		}
		service.GRPCEndpoints = append(service.GRPCEndpoints, endpoint)
		result.endpoints[index] = endpointPlan
		result.endpointByExpr[endpoint] = endpointPlan
	}
	return result, nil
}

// copyGRPCMethod copies the method fields read by the gRPC transport.
func copyGRPCMethod(source *expr.MethodExpr, service *expr.ServiceExpr) *expr.MethodExpr {
	method := &expr.MethodExpr{
		Name:             source.Name,
		Description:      source.Description,
		Payload:          expr.DupAtt(source.Payload),
		Result:           expr.DupAtt(source.Result),
		Requirements:     copyGRPCRequirements(source.Requirements),
		Service:          service,
		Meta:             copyGRPCMeta(source.Meta),
		Idempotent:       source.Idempotent,
		Stream:           source.Stream,
		StreamingPayload: expr.DupAtt(source.StreamingPayload),
	}
	switch {
	case source.StreamingResult == nil:
		method.StreamingResult = nil
	case source.StreamingResult == source.Result:
		method.StreamingResult = method.Result
	default:
		method.StreamingResult = expr.DupAtt(source.StreamingResult)
	}
	method.Errors = make([]*expr.ErrorExpr, len(source.Errors))
	for index, sourceError := range source.Errors {
		method.Errors[index] = &expr.ErrorExpr{
			Name:          sourceError.Name,
			AttributeExpr: copyGRPCErrorAttribute(sourceError.AttributeExpr),
		}
	}
	return method
}

// copyGRPCErrorAttribute preserves Goa's built-in error type while copying
// fields from a custom error type.
func copyGRPCErrorAttribute(source *expr.AttributeExpr) *expr.AttributeExpr {
	result := expr.DupAtt(source)
	if expr.IsErrorResult(source.Type) {
		result.Type = expr.ErrorResult
	}
	return result
}

// copyGRPCResponse copies one success or error response.
func copyGRPCResponse(source *expr.GRPCResponseExpr) *expr.GRPCResponseExpr {
	return &expr.GRPCResponseExpr{
		StatusCode:  source.StatusCode,
		Description: source.Description,
		Message:     expr.DupAtt(source.Message),
		Headers:     expr.DupMappedAtt(source.Headers),
		Trailers:    expr.DupMappedAtt(source.Trailers),
		Meta:        copyGRPCMeta(source.Meta),
	}
}

// copyGRPCRequirements copies security lists so later list edits cannot change
// generated metadata handling.
func copyGRPCRequirements(source []*expr.SecurityExpr) []*expr.SecurityExpr {
	result := make([]*expr.SecurityExpr, len(source))
	for index, requirement := range source {
		copy := expr.DupRequirement(requirement)
		copy.Scopes = append([]string(nil), requirement.Scopes...)
		for schemeIndex, scheme := range copy.Schemes {
			scheme.Scopes = append([]*expr.ScopeExpr(nil), scheme.Scopes...)
			scheme.Flows = append([]*expr.FlowExpr(nil), scheme.Flows...)
			scheme.Meta = copyGRPCMeta(scheme.Meta)
			copy.Schemes[schemeIndex] = scheme
		}
		result[index] = copy
	}
	return result
}

// copyGRPCMeta copies every value list in one Meta map.
func copyGRPCMeta(source expr.MetaExpr) expr.MetaExpr {
	if source == nil {
		return nil
	}
	result := make(expr.MetaExpr, len(source))
	for name, values := range source {
		result[name] = append([]string(nil), values...)
	}
	return result
}

// planEndpointMetadata prepares every request, response, and legacy request
// metadata field used by one endpoint.
func planEndpointMetadata(endpoint *grpcEndpointPlan) error {
	expression := endpoint.expression
	groups := []struct {
		mapped  *expr.MappedAttributeExpr
		service *expr.AttributeExpr
	}{
		{expression.Metadata, expression.MethodExpr.Payload},
		{expression.Response.Headers, expression.MethodExpr.Result},
		{expression.Response.Trailers, expression.MethodExpr.Result},
	}
	if endpoint.legacyStream {
		endpoint.legacyMetadata = legacyRequestMetadata(expression)
		groups = append(groups, struct {
			mapped  *expr.MappedAttributeExpr
			service *expr.AttributeExpr
		}{endpoint.legacyMetadata, expression.MethodExpr.Payload})
	}
	for _, group := range groups {
		plans, err := planMetadataFields(group.mapped, group.service)
		if err != nil {
			return err
		}
		endpoint.metadata[group.mapped] = plans
	}
	return nil
}

// planMetadataFields prepares the Go value and both conversions for every field
// in one metadata group.
func planMetadataFields(mapped *expr.MappedAttributeExpr, service *expr.AttributeExpr) ([]*grpcMetadataPlan, error) {
	var result []*grpcMetadataPlan
	err := codegen.WalkMappedAttr(mapped, func(name, element string, required bool, attribute *expr.AttributeExpr) error {
		wire := nativeMetadataAttribute(attribute)
		scope := codegen.NewNameScope()
		wireContext := codegen.NewAttributeContext(false, false, true, "", scope).Enter(wire)
		serviceField := service
		fieldType := service.Type
		fieldName := codegen.Goify(name, true)
		var pointer bool
		if !expr.IsObject(service.Type) {
			fieldName = ""
		} else {
			pointer = service.IsPrimitivePointer(name, true)
			serviceField = service.Find(name)
			fieldType = serviceField.Type
		}
		encode, err := codegen.NewTransformPlan(serviceField, wire, "", nil)
		if err != nil {
			return err
		}
		decode, err := codegen.NewTransformPlan(wire, serviceField, "", nil)
		if err != nil {
			return err
		}
		if len(encode.Helpers()) > 0 || len(decode.Helpers()) > 0 {
			return fmt.Errorf("metadata field %q needs a separate conversion function", name)
		}
		result = append(result, &grpcMetadataPlan{
			name:         name,
			element:      element,
			required:     required,
			fieldName:    fieldName,
			fieldType:    fieldType,
			pointer:      pointer,
			serviceField: serviceField,
			wire:         wire,
			scope:        scope,
			validation:   codegen.AttributeValidationCode(wire, nil, wireContext, required, false, codegen.Goify(name, false), name),
			encode:       encode,
			decode:       decode,
		})
		return nil
	})
	return result, err
}

// legacyRequestMetadata builds the metadata fields used by clients that send
// the first streamed payload through metadata.
func legacyRequestMetadata(endpoint *expr.GRPCEndpointExpr) *expr.MappedAttributeExpr {
	payload := endpoint.MethodExpr.Payload
	legacy := expr.DupMappedAtt(endpoint.Metadata)
	metadataObject := expr.AsObject(legacy.Type)
	if payloadObject := expr.AsObject(payload.Type); payloadObject != nil {
		for _, named := range *payloadObject {
			if metadataObject.Attribute(named.Name) == nil {
				metadataObject.Set(named.Name, expr.DupAtt(named.Attribute))
			}
			if payload.IsRequired(named.Name) {
				legacy.Validation.AddRequired(named.Name)
			}
		}
	} else {
		metadataObject.Set("goa_payload", expr.DupAtt(payload))
		legacy.Validation.AddRequired("goa_payload")
	}
	return legacy
}

// grpcServiceImportPaths records every package used by service values in gRPC
// client, server, type, and command-line files.
func grpcServiceImportPaths(generation *codegen.Generation, service *expr.GRPCServiceExpr) []string {
	paths := make(map[string]struct{})
	seen := make(map[expr.UserType]struct{})
	for _, attribute := range grpcEndpointAttributes(service.GRPCEndpoints...) {
		collectGRPCAttributeImportPaths(generation, attribute, paths, seen)
	}
	result := make([]string, 0, len(paths))
	for importPath := range paths {
		result = append(result, importPath)
	}
	sort.Strings(result)
	return result
}

// collectGRPCAttributeImportPaths walks one service value and records named
// generated packages and explicit field packages.
func collectGRPCAttributeImportPaths(generation *codegen.Generation, attribute *expr.AttributeExpr, paths map[string]struct{}, seen map[expr.UserType]struct{}) {
	if attribute == nil || attribute.Type == expr.Empty {
		return
	}
	if _, spec := codegen.GetMetaType(attribute); spec != nil {
		paths[spec.Path] = struct{}{}
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if location := codegen.UserTypeLocation(actual); location != nil {
			paths[path.Join(generation.GenPkg(), location.RelImportPath)] = struct{}{}
		}
		if _, ok := seen[actual]; ok {
			return
		}
		seen[actual] = struct{}{}
		collectGRPCAttributeImportPaths(generation, actual.Attribute(), paths, seen)
	case *expr.Object:
		for _, named := range *actual {
			collectGRPCAttributeImportPaths(generation, named.Attribute, paths, seen)
		}
	case *expr.Array:
		collectGRPCAttributeImportPaths(generation, actual.ElemType, paths, seen)
	case *expr.Map:
		collectGRPCAttributeImportPaths(generation, actual.KeyType, paths, seen)
		collectGRPCAttributeImportPaths(generation, actual.ElemType, paths, seen)
	case *expr.Union:
		for _, named := range actual.Values {
			collectGRPCAttributeImportPaths(generation, named.Attribute, paths, seen)
		}
	}
}
