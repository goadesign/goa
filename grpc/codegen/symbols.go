// This file chooses every Go name written into generated gRPC client and server
// packages. A definition and every call to it share one stored name.
package codegen

import (
	"cmp"
	"fmt"
	"path"
	"slices"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// grpcSymbols contains the names written for one gRPC service.
	grpcSymbols struct {
		clientStruct *codegen.NameDeclaration
		clientInit   *codegen.NameDeclaration
		serverStruct *codegen.NameDeclaration
		serverInit   *codegen.NameDeclaration
		endpoints    map[*expr.GRPCEndpointExpr]*grpcEndpointSymbols
	}

	// grpcEndpointSymbols contains the names written for one gRPC endpoint.
	grpcEndpointSymbols struct {
		clientStream  *codegen.NameDeclaration
		clientBuild   *codegen.NameDeclaration
		clientEncode  *codegen.NameDeclaration
		clientDecode  *codegen.NameDeclaration
		serverStream  *codegen.NameDeclaration
		serverHandler *codegen.NameDeclaration
		serverDecode  *codegen.NameDeclaration
		serverEncode  *codegen.NameDeclaration
		legacyDecode  *codegen.NameDeclaration
		cliPayload    *grpcConversion
		clientInits   map[grpcInitKey]*grpcConversion
		serverInits   map[grpcInitKey]*grpcConversion
	}

	// grpcConversion contains one top-level conversion function and the extra
	// conversion functions it calls for nested values.
	grpcConversion struct {
		declaration           *codegen.NameDeclaration
		transform             *codegen.TransformPlan
		pkg                   *codegen.GeneratedPackage
		order                 grpcSymbolOrder
		preferredName         string
		fullName              string
		serviceName           string
		messageName           string
		releasedNames         []string
		releasedResponseNames []string
		side                  grpcPackageSide
		bound                 bool
	}

	// grpcTransformRegistry collects every private conversion function written
	// to one generated client or server package.
	grpcTransformRegistry struct {
		pkg     *codegen.GeneratedPackage
		helpers *codegen.TransformHelperRegistry
	}

	// grpcConversionKey contains the package and types that decide one
	// conversion function. endpoint is set when metadata changes its arguments.
	grpcConversionKey struct {
		pkg      *codegen.GeneratedPackage
		message  *protobufMessageRecord
		service  expr.UserType
		endpoint *expr.GRPCEndpointExpr
		view     string
		proto    bool
	}

	// grpcSymbolID records which package, service, method, error, and field
	// produced one Go name. These values decide collision order but do not appear
	// in the name.
	grpcSymbolID struct {
		side       grpcPackageSide
		role       grpcSymbolRole
		api        string
		service    string
		method     string
		subject    string
		view       string
		path       string
		source     string
		target     string
		definition codegen.TransformHelperDefinitionLocation
		operation  int
	}

	// grpcSymbolOrder decides which item keeps an unsuffixed Go name when several
	// items request the same name.
	grpcSymbolOrder grpcSymbolID

	// grpcPackageSide identifies the generated package that contains a name.
	grpcPackageSide uint8

	// grpcSymbolRole identifies what a generated name defines.
	grpcSymbolRole uint8

	// grpcInitRole identifies one conversion constructor used by an endpoint.
	grpcInitRole uint8

	// grpcInitKey says which endpoint value uses a conversion.
	grpcInitKey struct {
		role    grpcInitRole
		subject string
		view    string
	}

	// grpcValidationNameKey contains the generated package and validation-body
	// facts used in the ordinary private function name.
	grpcValidationNameKey struct {
		side        validateKind
		api         string
		service     string
		message     string
		targetName  string
		contextName string
	}
)

const (
	grpcClientPackage grpcPackageSide = iota + 1
	grpcServerPackage
)

const (
	grpcClientStructRole grpcSymbolRole = iota + 1
	grpcClientInitRole
	grpcServerStructRole
	grpcServerInitRole
	grpcClientStreamRole
	grpcClientBuildRole
	grpcClientEncodeRole
	grpcClientDecodeRole
	grpcServerStreamRole
	grpcServerHandlerRole
	grpcServerDecodeRole
	grpcServerEncodeRole
	grpcLegacyDecodeRole
	grpcConversionInitRole
	grpcValidationRole
	grpcTransformHelperRole
)

const (
	grpcRequestInit grpcInitRole = iota + 1
	grpcResponseInit
	grpcStreamingRequestInit
	grpcStreamingResponseInit
	grpcLegacyRequestInit
	grpcErrorInit
)

// collectGRPCSymbols requests the client and server names that can be chosen
// directly from the service and endpoint designs.
func collectGRPCSymbols(generation *codegen.Generation, input PlanInput, service *expr.GRPCServiceExpr, pathName string) (*grpcSymbols, error) {
	clientPackage, err := generation.ClaimPackage(path.Join(generation.GenPkg(), "grpc", pathName, "client"))
	if err != nil {
		return nil, err
	}
	serverPackage, err := generation.ClaimPackage(path.Join(generation.GenPkg(), "grpc", pathName, "server"))
	if err != nil {
		return nil, err
	}
	declare := func(pkg *codegen.GeneratedPackage, kind codegen.PackageNameKind, preferred string, visibility codegen.PackageNameVisibility, id grpcSymbolID) (*codegen.NameDeclaration, error) {
		declaration := codegen.NewPreferredName(kind, preferred, visibility, grpcSymbolOrder(id))
		if err := pkg.DeclareName(declaration); err != nil {
			return nil, err
		}
		return declaration, nil
	}
	serviceID := grpcSymbolID{api: input.Root.API.Name, service: service.Name()}
	symbols := &grpcSymbols{endpoints: make(map[*expr.GRPCEndpointExpr]*grpcEndpointSymbols)}
	if symbols.clientStruct, err = declare(clientPackage, codegen.NameType, "Client", codegen.ExportedName, serviceID.client(grpcClientStructRole)); err != nil {
		return nil, err
	}
	if symbols.clientInit, err = declare(clientPackage, codegen.NameFunction, "NewClient", codegen.ExportedName, serviceID.client(grpcClientInitRole)); err != nil {
		return nil, err
	}
	if symbols.serverStruct, err = declare(serverPackage, codegen.NameType, "Server", codegen.ExportedName, serviceID.server(grpcServerStructRole)); err != nil {
		return nil, err
	}
	if symbols.serverInit, err = declare(serverPackage, codegen.NameFunction, "New", codegen.ExportedName, serviceID.server(grpcServerInitRole)); err != nil {
		return nil, err
	}
	for _, endpoint := range service.GRPCEndpoints {
		names, err := input.Service.HTTPMethodNames(endpoint.MethodExpr)
		if err != nil {
			return nil, err
		}
		id := serviceID.withMethod(endpoint.Name())
		endpointSymbols := &grpcEndpointSymbols{
			clientInits: make(map[grpcInitKey]*grpcConversion),
			serverInits: make(map[grpcInitKey]*grpcConversion),
		}
		endpointSymbols.clientBuild, err = declare(clientPackage, codegen.NameFunction, "Build"+names.Method+"Func", codegen.ExportedName, id.client(grpcClientBuildRole))
		if err != nil {
			return nil, err
		}
		if endpoint.MethodExpr.Payload.Type != expr.Empty {
			endpointSymbols.clientEncode, err = declare(clientPackage, codegen.NameFunction, "Encode"+names.Method+"Request", codegen.ExportedName, id.client(grpcClientEncodeRole))
			if err != nil {
				return nil, err
			}
			endpointSymbols.serverDecode, err = declare(serverPackage, codegen.NameFunction, "Decode"+names.Method+"Request", codegen.ExportedName, id.server(grpcServerDecodeRole))
			if err != nil {
				return nil, err
			}
		}
		if endpoint.MethodExpr.Result.Type != expr.Empty || endpoint.MethodExpr.IsStreaming() {
			endpointSymbols.clientDecode, err = declare(clientPackage, codegen.NameFunction, "Decode"+names.Method+"Response", codegen.ExportedName, id.client(grpcClientDecodeRole))
			if err != nil {
				return nil, err
			}
		}
		endpointSymbols.serverEncode, err = declare(serverPackage, codegen.NameFunction, "Encode"+names.Method+"Response", codegen.ExportedName, id.server(grpcServerEncodeRole))
		if err != nil {
			return nil, err
		}
		endpointSymbols.serverHandler, err = declare(serverPackage, codegen.NameFunction, "New"+names.Method+"Handler", codegen.ExportedName, id.server(grpcServerHandlerRole))
		if err != nil {
			return nil, err
		}
		if endpoint.MethodExpr.IsStreaming() {
			endpointSymbols.clientStream, err = declare(clientPackage, codegen.NameType, names.ClientStream, codegen.ExportedName, id.client(grpcClientStreamRole))
			if err != nil {
				return nil, err
			}
			endpointSymbols.serverStream, err = declare(serverPackage, codegen.NameType, names.ServerStream, codegen.ExportedName, id.server(grpcServerStreamRole))
			if err != nil {
				return nil, err
			}
		}
		if endpoint.LegacyStreamCompat() {
			preferred := "decode" + names.Method + "LegacyRequest"
			endpointSymbols.legacyDecode, err = declare(serverPackage, codegen.NameFunction, preferred, codegen.UnexportedName, id.server(grpcLegacyDecodeRole))
			if err != nil {
				return nil, err
			}
		}
		symbols.endpoints[endpoint] = endpointSymbols
	}
	return symbols, nil
}

// planGRPCTransforms records each conversion and requests the names of any
// nested conversion functions it will call. It does not read those names yet.
func planGRPCTransforms(
	generation *codegen.Generation,
	input PlanInput,
	service *expr.GRPCServiceExpr,
	protobuf *protobufServicePlan,
	symbols *grpcSymbols,
	conversions map[grpcConversionKey]*grpcConversion,
	registries map[*codegen.GeneratedPackage]*grpcTransformRegistry,
	pathName string,
) error {
	clientPackage := generation.Package(path.Join(generation.GenPkg(), "grpc", pathName, "client"))
	serverPackage := generation.Package(path.Join(generation.GenPkg(), "grpc", pathName, "server"))
	for index, endpoint := range service.GRPCEndpoints {
		messages := protobuf.messages[index]
		endpointSymbols := symbols.endpoints[endpoint]
		result := endpoint.MethodExpr.Result
		_, viewed := result.Type.(*expr.ResultTypeExpr)
		if viewed {
			projected, err := input.Service.ProjectedResult(endpoint.MethodExpr)
			if err != nil {
				return err
			}
			result = projected
		}
		conversionFor := func(side grpcPackageSide, source, target *expr.AttributeExpr, proto, endpointSpecific bool, view string) (*grpcConversion, error) {
			pkg := clientPackage
			if side == grpcServerPackage {
				pkg = serverPackage
			}
			protobufAttribute := source
			serviceAttribute := target
			if proto {
				protobufAttribute = target
				serviceAttribute = source
			}
			message := protobuf.catalog.messageRecord(protobufAttribute)
			var serviceType expr.UserType
			if userType, ok := serviceAttribute.Type.(expr.UserType); ok {
				serviceType = userType.Origin()
			}
			conversionKey := grpcConversionKey{
				pkg:     pkg,
				message: message,
				service: serviceType,
				view:    view,
				proto:   proto,
			}
			if endpointSpecific {
				conversionKey.endpoint = endpoint
			}
			conversion := conversions[conversionKey]
			preferred, fullName, serviceName, messageName := grpcConversionNames(serviceAttribute, message, proto)
			viewKey := grpcInitKey{view: view}
			preferred = grpcViewedConversionName(endpoint.MethodExpr, viewKey, preferred)
			fullName = grpcViewedConversionName(endpoint.MethodExpr, viewKey, fullName)
			id := grpcSymbolID{
				side:      side,
				role:      grpcConversionInitRole,
				api:       input.Root.API.Name,
				service:   service.Name(),
				subject:   serviceName,
				view:      conversionKey.view,
				path:      messageName,
				operation: grpcConversionDirection(proto),
			}
			if endpointSpecific {
				id.method = endpoint.Name()
			}
			if conversion == nil {
				transform, err := newGRPCTransformPlan(source, target, proto, protobuf)
				if err != nil {
					return nil, err
				}
				sourceLayout, targetLayout, err := planGRPCTransformLayouts(
					generation, input, endpoint.MethodExpr, source, target, proto, protobuf, pathName,
				)
				if err != nil {
					return nil, err
				}
				registry := registries[pkg]
				if registry == nil {
					registry = &grpcTransformRegistry{pkg: pkg, helpers: codegen.NewTransformHelperRegistry()}
					registries[pkg] = registry
				}
				helperID := id
				helperID.role = grpcTransformHelperRole
				if err := registry.helpers.Collect(transform, sourceLayout, targetLayout,
					func(location codegen.TransformHelperDefinitionLocation) codegen.PackageNameOrder {
						helperID.definition = location
						return grpcSymbolOrder(helperID)
					}); err != nil {
					return nil, err
				}
				conversion = &grpcConversion{
					transform:     transform,
					pkg:           pkg,
					order:         grpcSymbolOrder(id),
					preferredName: preferred,
					fullName:      fullName,
					serviceName:   serviceName,
					messageName:   messageName,
					side:          side,
				}
				conversions[conversionKey] = conversion
			}
			return conversion, nil
		}
		plan := func(side grpcPackageSide, key grpcInitKey, source, target *expr.AttributeExpr, proto, endpointSpecific bool) error {
			conversion, err := conversionFor(side, source, target, proto, endpointSpecific, key.view)
			if err != nil {
				return err
			}
			releasedName := releasedGRPCConversionName(endpoint, key, source, target, proto, conversion)
			conversion.releasedNames = append(conversion.releasedNames, releasedName)
			if key.role == grpcResponseInit && !slices.Contains(conversion.releasedResponseNames, releasedName) {
				conversion.releasedResponseNames = append(conversion.releasedResponseNames, releasedName)
			}
			inits := endpointSymbols.clientInits
			if side == grpcServerPackage {
				inits = endpointSymbols.serverInits
			}
			inits[key] = conversion
			return nil
		}
		if endpoint.MethodExpr.Payload.Type != expr.Empty {
			if err := plan(grpcServerPackage, grpcInitKey{role: grpcRequestInit}, messages.request, endpoint.MethodExpr.Payload, false, !endpoint.Metadata.IsEmpty()); err != nil {
				return err
			}
			cliConversion, err := conversionFor(grpcClientPackage, messages.request, endpoint.MethodExpr.Payload, false, false, "")
			if err != nil {
				return err
			}
			endpointSymbols.cliPayload = cliConversion
		}
		if !(endpoint.MethodExpr.IsPayloadStreaming() && isEmpty(endpoint.Request.Type)) {
			if err := plan(grpcClientPackage, grpcInitKey{role: grpcRequestInit}, endpoint.MethodExpr.Payload, messages.request, true, false); err != nil {
				return err
			}
		}
		if viewed {
			for _, view := range grpcResultViews(endpoint.MethodExpr) {
				viewResult, err := grpcResultForView(result, view)
				if err != nil {
					return err
				}
				if err := plan(grpcServerPackage, grpcInitKey{role: grpcResponseInit, view: view}, viewResult, messages.response, true, false); err != nil {
					return err
				}
			}
		} else if err := plan(grpcServerPackage, grpcInitKey{role: grpcResponseInit}, result, messages.response, true, false); err != nil {
			return err
		}
		if endpoint.MethodExpr.Result.Type != expr.Empty && !endpoint.MethodExpr.IsStreaming() {
			responseMetadata := !endpoint.Response.Headers.IsEmpty() || !endpoint.Response.Trailers.IsEmpty()
			if viewed {
				for _, view := range grpcResultViews(endpoint.MethodExpr) {
					viewResult, err := grpcResultForView(result, view)
					if err != nil {
						return err
					}
					if err := plan(grpcClientPackage, grpcInitKey{role: grpcResponseInit, view: view}, messages.response, viewResult, false, responseMetadata); err != nil {
						return err
					}
				}
			} else if err := plan(grpcClientPackage, grpcInitKey{role: grpcResponseInit}, messages.response, result, false, responseMetadata); err != nil {
				return err
			}
		}
		if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty {
			key := grpcInitKey{role: grpcStreamingRequestInit}
			if err := plan(grpcServerPackage, key, messages.streamingRequest, endpoint.MethodExpr.StreamingPayload, false, false); err != nil {
				return err
			}
			if err := plan(grpcClientPackage, key, endpoint.MethodExpr.StreamingPayload, messages.streamingRequest, true, false); err != nil {
				return err
			}
		}
		if endpoint.MethodExpr.Result.Type != expr.Empty && endpoint.MethodExpr.IsStreaming() {
			if viewed {
				for _, view := range grpcResultViews(endpoint.MethodExpr) {
					viewResult, err := grpcResultForView(result, view)
					if err != nil {
						return err
					}
					key := grpcInitKey{role: grpcStreamingResponseInit, view: view}
					if err := plan(grpcServerPackage, key, viewResult, messages.response, true, false); err != nil {
						return err
					}
				}
			} else {
				key := grpcInitKey{role: grpcStreamingResponseInit}
				if err := plan(grpcServerPackage, key, result, messages.response, true, false); err != nil {
					return err
				}
			}
			if viewed {
				for _, view := range grpcResultViews(endpoint.MethodExpr) {
					viewResult, err := grpcResultForView(result, view)
					if err != nil {
						return err
					}
					key := grpcInitKey{role: grpcStreamingResponseInit, view: view}
					if err := plan(grpcClientPackage, key, messages.response, viewResult, false, false); err != nil {
						return err
					}
				}
			} else {
				key := grpcInitKey{role: grpcStreamingResponseInit}
				if err := plan(grpcClientPackage, key, messages.response, result, false, false); err != nil {
					return err
				}
			}
		}
		if endpoint.LegacyStreamCompat() && expr.IsObject(endpoint.MethodExpr.Payload.Type) {
			key := grpcInitKey{role: grpcLegacyRequestInit}
			if err := plan(grpcServerPackage, key, &expr.AttributeExpr{Type: expr.Empty}, endpoint.MethodExpr.Payload, false, true); err != nil {
				return err
			}
		}
		for _, grpcError := range endpoint.GRPCErrors {
			message := messages.errors[grpcError.Name]
			if message == nil {
				continue
			}
			key := grpcInitKey{role: grpcErrorInit, subject: grpcError.Name}
			if err := plan(grpcServerPackage, key, grpcError.AttributeExpr, message, true, false); err != nil {
				return err
			}
			if err := plan(grpcClientPackage, key, message, grpcError.AttributeExpr, false, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// grpcResultViews returns the views the server may send for method. A view set
// in the design is the only possible value. Otherwise callers may select any
// view declared by the result type.
func grpcResultViews(method *expr.MethodExpr) []string {
	if method.Result.Meta != nil {
		if view, ok := method.Result.Meta.Last(expr.ViewMetaKey); ok {
			return []string{view}
		}
	}
	resultType := method.Result.Type.(*expr.ResultTypeExpr)
	views := make([]string, len(resultType.Views))
	for index, view := range resultType.Views {
		views[index] = view.Name
	}
	return views
}

// grpcResultForView keeps the generated projected Go type but limits the
// conversion plan to fields included in view.
func grpcResultForView(result *expr.AttributeExpr, view string) (*expr.AttributeExpr, error) {
	resultType := result.Type.(*expr.ResultTypeExpr)
	selected, err := expr.Project(resultType, view)
	if err != nil {
		return nil, err
	}
	selectedResult := expr.DupAtt(result)
	selectedResult.Type = selected
	return grpcViewAttribute(result, selectedResult, make(map[expr.UserType]expr.UserType)), nil
}

// grpcViewAttribute copies only the selected fields while reusing the Go types
// already generated for the complete result.
func grpcViewAttribute(full, selected *expr.AttributeExpr, seen map[expr.UserType]expr.UserType) *expr.AttributeExpr {
	filtered := expr.DupAtt(selected)
	switch selectedType := selected.Type.(type) {
	case expr.UserType:
		if existing, ok := seen[selectedType]; ok {
			filtered.Type = existing
			return filtered
		}
		fullType := full.Type.(expr.UserType)
		copy := fullType.Dup(expr.DupAtt(selectedType.Attribute()))
		seen[selectedType] = copy
		copy.SetAttribute(grpcViewAttribute(fullType.Attribute(), selectedType.Attribute(), seen))
		filtered.Type = copy
	case *expr.Array:
		fullType := full.Type.(*expr.Array)
		filtered.Type = &expr.Array{
			ElemType:         grpcViewAttribute(fullType.ElemType, selectedType.ElemType, seen),
			NonNullableElems: selectedType.NonNullableElems,
		}
	case *expr.Map:
		fullType := full.Type.(*expr.Map)
		filtered.Type = &expr.Map{
			KeyType:  grpcViewAttribute(fullType.KeyType, selectedType.KeyType, seen),
			ElemType: grpcViewAttribute(fullType.ElemType, selectedType.ElemType, seen),
		}
	case *expr.Object:
		fullType := full.Type.(*expr.Object)
		object := make(expr.Object, 0, len(*selectedType))
		for _, field := range *selectedType {
			object = append(object, &expr.NamedAttributeExpr{
				Name:      field.Name,
				Attribute: grpcViewAttribute(fullType.Attribute(field.Name), field.Attribute, seen),
			})
		}
		filtered.Type = &object
	case *expr.Union:
		fullType := full.Type.(*expr.Union)
		union := &expr.Union{
			TypeName: selectedType.TypeName,
			TypeKey:  selectedType.TypeKey,
			ValueKey: selectedType.ValueKey,
			Values:   make([]*expr.NamedAttributeExpr, 0, len(selectedType.Values)),
		}
		for index, branch := range selectedType.Values {
			union.Values = append(union.Values, &expr.NamedAttributeExpr{
				Name:      branch.Name,
				Attribute: grpcViewAttribute(fullType.Values[index].Attribute, branch.Attribute, seen),
			})
		}
		filtered.Type = union
	}
	return filtered
}

// grpcProtobufValidationForView keeps the full protobuf message layout while
// retaining checks only for fields selected by one result view.
func grpcProtobufValidationForView(response, selected *expr.AttributeExpr) *expr.AttributeExpr {
	validation := expr.DupAtt(response)
	limitGRPCProtobufValidation(validation, selected, make(map[protobufAttributePair]struct{}))
	return validation
}

// limitGRPCProtobufValidation removes checks for response fields that the
// selected result view does not carry.
func limitGRPCProtobufValidation(response, selected *expr.AttributeExpr, seen map[protobufAttributePair]struct{}) {
	pair := protobufAttributePair{left: response, right: selected}
	if _, ok := seen[pair]; ok {
		return
	}
	seen[pair] = struct{}{}
	if responseType, ok := response.Type.(expr.UserType); ok {
		if isWrappedAttr(response) {
			selected = unwrapSelectedGRPCView(selected, unwrapAttr(response).Type.Kind())
			limitGRPCProtobufValidation(unwrapAttr(response), selected, seen)
			return
		}
		if selectedType, ok := selected.Type.(expr.UserType); ok {
			limitGRPCProtobufValidation(responseType.Attribute(), selectedType.Attribute(), seen)
			return
		}
		limitGRPCProtobufValidation(responseType.Attribute(), selected, seen)
		return
	}
	if selectedType, ok := selected.Type.(expr.UserType); ok {
		limitGRPCProtobufValidation(response, selectedType.Attribute(), seen)
		return
	}

	switch responseType := response.Type.(type) {
	case *expr.Object:
		selectedObject := expr.AsObject(selected.Type)
		if selectedObject == nil {
			panic("selected gRPC result view does not match its protobuf response")
		}
		if response.Validation != nil {
			required := response.Validation.Required[:0]
			for _, name := range response.Validation.Required {
				if selectedObject.Attribute(name) != nil && selected.IsRequired(name) {
					required = append(required, name)
				}
			}
			response.Validation.Required = required
		}
		for _, field := range *responseType {
			selectedField := selectedObject.Attribute(field.Name)
			if selectedField == nil {
				// A selected sibling may use the same nested type. Copy this
				// field before removing checks so the sibling keeps its rules.
				field.Attribute = expr.DupAtt(field.Attribute)
				clearGRPCProtobufValidation(field.Attribute, make(map[*expr.AttributeExpr]struct{}))
				continue
			}
			limitGRPCProtobufValidation(field.Attribute, selectedField, seen)
		}
	case *expr.Array:
		selectedArray := expr.AsArray(selected.Type)
		if selectedArray == nil {
			panic("selected gRPC result view does not match its protobuf response array")
		}
		limitGRPCProtobufValidation(responseType.ElemType, selectedArray.ElemType, seen)
	case *expr.Map:
		selectedMap := expr.AsMap(selected.Type)
		if selectedMap == nil {
			panic("selected gRPC result view does not match its protobuf response map")
		}
		limitGRPCProtobufValidation(responseType.KeyType, selectedMap.KeyType, seen)
		limitGRPCProtobufValidation(responseType.ElemType, selectedMap.ElemType, seen)
	case *expr.Union:
		selectedUnion := expr.AsUnion(selected.Type)
		if selectedUnion == nil || len(responseType.Values) != len(selectedUnion.Values) {
			panic("selected gRPC result view does not match its protobuf response union")
		}
		for index, branch := range responseType.Values {
			limitGRPCProtobufValidation(branch.Attribute, selectedUnion.Values[index].Attribute, seen)
		}
	}
}

// unwrapSelectedGRPCView follows generated service type names until the value
// has the same collection shape as a protobuf wrapper field.
func unwrapSelectedGRPCView(selected *expr.AttributeExpr, kind expr.Kind) *expr.AttributeExpr {
	for selected.Type.Kind() != kind {
		userType, ok := selected.Type.(expr.UserType)
		if !ok {
			panic("selected gRPC result view does not match its protobuf wrapper")
		}
		selected = userType.Attribute()
	}
	return selected
}

// clearGRPCProtobufValidation removes checks from an unselected protobuf field
// without changing the message fields or their generated Go types.
func clearGRPCProtobufValidation(attribute *expr.AttributeExpr, seen map[*expr.AttributeExpr]struct{}) {
	if _, ok := seen[attribute]; ok {
		return
	}
	seen[attribute] = struct{}{}
	attribute.Validation = nil
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		clearGRPCProtobufValidation(actual.Attribute(), seen)
	case *expr.Object:
		for _, field := range *actual {
			clearGRPCProtobufValidation(field.Attribute, seen)
		}
	case *expr.Array:
		clearGRPCProtobufValidation(actual.ElemType, seen)
	case *expr.Map:
		clearGRPCProtobufValidation(actual.KeyType, seen)
		clearGRPCProtobufValidation(actual.ElemType, seen)
	case *expr.Union:
		for _, branch := range actual.Values {
			clearGRPCProtobufValidation(branch.Attribute, seen)
		}
	}
}

// declareGRPCTransforms keeps a released response name when one method owns
// it. Conversions shared by several methods use names based on their types.
func declareGRPCTransforms(conversions map[grpcConversionKey]*grpcConversion, registries map[*codegen.GeneratedPackage]*grpcTransformRegistry) error {
	type nameKey struct {
		pkg  *codegen.GeneratedPackage
		name string
	}
	counts := make(map[nameKey]int)
	for _, conversion := range conversions {
		if len(conversion.releasedNames) > 1 {
			counts[nameKey{pkg: conversion.pkg, name: conversion.preferredName}]++
		}
	}
	for _, conversion := range conversions {
		if len(conversion.releasedNames) == 0 {
			continue
		}
		name := conversion.releasedNames[0]
		useResponseName := len(conversion.releasedResponseNames) == 1
		useTypeName := !useResponseName && len(conversion.releasedNames) > 1
		if useResponseName {
			name = conversion.releasedResponseNames[0]
		} else if useTypeName {
			name = conversion.preferredName
		}
		if useTypeName && counts[nameKey{pkg: conversion.pkg, name: name}] > 1 {
			name = conversion.fullName
		}
		declaration := codegen.NewPreferredName(codegen.NameFunction, name, codegen.ExportedName, conversion.order)
		if err := conversion.pkg.DeclareName(declaration); err != nil {
			return err
		}
		conversion.declaration = declaration
	}
	for _, registry := range registries {
		groups, err := registry.helpers.Finalize()
		if err != nil {
			return err
		}
		for _, group := range groups {
			order := group.Order().(grpcSymbolOrder)
			name := grpcTransformHelperName(group.Definition(), order.operation == grpcConversionDirection(true))
			declaration := codegen.NewPreferredName(codegen.NameFunction, name, codegen.UnexportedName, order)
			if err := registry.pkg.DeclareName(declaration); err != nil {
				return err
			}
			if err := group.Bind(declaration); err != nil {
				return err
			}
		}
	}
	return nil
}

// releasedGRPCConversionName returns the constructor name generated before
// conversions shared by several methods were combined.
func releasedGRPCConversionName(endpoint *expr.GRPCEndpointExpr, key grpcInitKey, source, target *expr.AttributeExpr, proto bool, conversion *grpcConversion) string {
	method := codegen.Goify(endpoint.Name(), true)
	switch key.role {
	case grpcRequestInit:
		if !proto {
			return "New" + method + "Payload"
		}
		return "NewProto" + conversion.messageName
	case grpcResponseInit:
		if !proto {
			return grpcViewedConversionName(endpoint.MethodExpr, key, "New"+method+"Result")
		}
		name := conversion.messageName
		bodyIsStruct := expr.IsUnion(target.Type)
		if object := expr.AsObject(target.Type); object != nil {
			bodyIsStruct = len(*object) > 0
		}
		if !bodyIsStruct && key.view == "" {
			name = conversion.serviceName
		}
		return grpcViewedConversionName(endpoint.MethodExpr, key, "NewProto"+name)
	case grpcStreamingRequestInit, grpcStreamingResponseInit:
		name := releasedGRPCStreamConversionName(source, target, proto, conversion)
		return grpcViewedConversionName(endpoint.MethodExpr, key, name)
	case grpcLegacyRequestInit:
		return "New" + method + "PayloadFromMetadata"
	case grpcErrorInit:
		return "New" + method + codegen.Goify(key.subject, true) + "Error"
	default:
		panic("unknown gRPC conversion role")
	}
}

// grpcViewedConversionName keeps the existing constructor name for the only
// view selected by a design. When callers choose a view, additional
// constructors include the view name.
func grpcViewedConversionName(method *expr.MethodExpr, key grpcInitKey, name string) string {
	if key.view == "" || key.view == expr.DefaultView {
		return name
	}
	if method.Result.Meta != nil {
		if _, fixed := method.Result.Meta.Last(expr.ViewMetaKey); fixed {
			return name
		}
	}
	return name + codegen.Goify(key.view, true)
}

// releasedGRPCStreamConversionName returns the name used by released Goa
// versions for a conversion of one streamed value.
func releasedGRPCStreamConversionName(source, target *expr.AttributeExpr, proto bool, conversion *grpcConversion) string {
	name := "New"
	if proto {
		name += "Proto"
	}
	if _, ok := source.Type.(expr.UserType); ok {
		if proto {
			name += conversion.serviceName
		} else {
			name += conversion.messageName
		}
	}
	targetName := conversion.serviceName
	if proto {
		targetName = conversion.messageName
	}
	if !expr.IsObject(target.Type) && !expr.IsUnion(target.Type) {
		targetName = conversion.messageName
		if proto {
			targetName = conversion.serviceName
		}
	}
	return name + targetName
}

// grpcConversionNames returns the short and complete constructor names and
// the type names used to order colliding requests.
func grpcConversionNames(serviceAttribute *expr.AttributeExpr, message *protobufMessageRecord, proto bool) (string, string, string, string) {
	var messageName string
	if message != nil {
		messageName = message.plannedName
	}
	serviceName := messageName
	if userType, ok := serviceAttribute.Type.(expr.UserType); ok && serviceAttribute.Type != expr.Empty {
		serviceName = codegen.Goify(userType.Name(), true)
	}
	if serviceName == "" {
		serviceName = codegen.Goify(serviceAttribute.Type.Name(), true)
	}
	name := "New" + serviceName
	fullName := "New" + serviceName + "FromProto" + messageName
	if message == nil {
		fullName = "New" + serviceName + "FromMetadata"
	}
	if proto {
		name = "NewProto" + serviceName
		fullName = "NewProto" + messageName + "From" + serviceName
	}
	return name, fullName, serviceName, messageName
}

// grpcTransformHelperName returns the private name derived from the two nested
// types converted by one helper.
func grpcTransformHelperName(definition codegen.TransformHelperDefinition, proto bool) string {
	source := grpcTransformTypeName(definition.Source)
	target := grpcTransformTypeName(definition.Target)
	if proto {
		return codegen.Goify("transform"+source+"ToProto"+target, false)
	}
	return codegen.Goify("transformProto"+source+"To"+target, false)
}

// grpcTransformTypeName returns the declared type name used in a private
// conversion function signature.
func grpcTransformTypeName(attribute *expr.AttributeExpr) string {
	if userType, ok := attribute.Type.(expr.UserType); ok {
		return codegen.Goify(userType.Name(), true)
	}
	return codegen.Goify(attribute.Type.Name(), true)
}

// grpcConversionDirection returns the fixed number used to order conversions
// to and from protobuf messages.
func grpcConversionDirection(proto bool) int {
	if proto {
		return 1
	}
	return 2
}

// planGRPCValidations records each validation function in the client or server
// package that writes it.
func planGRPCValidations(generation *codegen.Generation, input PlanInput, service *expr.GRPCServiceExpr, protobuf *protobufServicePlan, pathName string) error {
	clientPackage := generation.Package(path.Join(generation.GenPkg(), "grpc", pathName, "client"))
	serverPackage := generation.Package(path.Join(generation.GenPkg(), "grpc", pathName, "server"))
	for index, endpoint := range service.GRPCEndpoints {
		messages := protobuf.messages[index]
		source := protobufValidationSource{
			api:     input.Root.API.Name,
			service: service.Name(),
			method:  endpoint.Name(),
		}
		if protobuf.catalog.messageRecord(messages.request) != nil {
			request := source
			request.role = protobufRequestValidation
			protobuf.catalog.collectValidation(messages.request, validateServer, request, "message", "message")
			if protobufCLIRequestNeedsMessage(messages.request) {
				// Command-line clients parse this request without a server, so
				// they need the same check before converting it.
				protobuf.catalog.collectValidation(messages.request, validateClient, request, "message", "message")
			}
		}
		if protobuf.catalog.messageRecord(messages.response) != nil {
			response := source
			response.role = protobufResponseValidation
			if messages.responseValidations == nil {
				protobuf.catalog.collectValidation(messages.response, validateClient, response, "message", "message")
			} else {
				for _, view := range grpcResultViews(endpoint.MethodExpr) {
					viewSource := response
					viewSource.view = view
					protobuf.catalog.collectValidation(messages.responseValidations[view], validateClient, viewSource, "message", "message")
				}
			}
		}
		for _, grpcError := range endpoint.GRPCErrors {
			message := messages.errors[grpcError.Name]
			if message == nil {
				continue
			}
			errorSource := source
			errorSource.error = grpcError.Name
			errorSource.role = protobufErrorValidation
			protobuf.catalog.collectValidation(message, validateClient, errorSource, "errmsg", "errmsg")
		}
		if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty {
			stream := source
			stream.role = protobufStreamingRequestValidation
			protobuf.catalog.collectValidation(messages.streamingRequest, validateServer, stream, "stream", "stream")
		}
	}
	privateNameCounts := make(map[grpcValidationNameKey]int)
	for _, validator := range protobuf.catalog.validators {
		if validator.source.path != "" {
			privateNameCounts[grpcValidationKey(validator)]++
		}
	}
	for _, validator := range protobuf.catalog.validators {
		pkg := clientPackage
		side := grpcClientPackage
		if validator.side == validateServer {
			pkg = serverPackage
			side = grpcServerPackage
		}
		id := grpcSymbolID{
			side:      side,
			role:      grpcValidationRole,
			api:       validator.source.api,
			service:   validator.source.service,
			method:    validator.source.method,
			subject:   validator.source.error,
			view:      validator.source.view,
			path:      validator.source.path,
			operation: int(validator.source.role),
		}
		validator.declaration = grpcValidationDeclaration(
			validator,
			id,
			privateNameCounts[grpcValidationKey(validator)] > 1,
		)
		if err := pkg.DeclareName(validator.declaration); err != nil {
			return err
		}
	}
	return nil
}

// grpcValidationDeclaration keeps private semantic names exact. A collision is
// a planning error instead of silently turning the name into a numbered name.
func grpcValidationDeclaration(validator *protobufValidationRecord, id grpcSymbolID, includeSource bool) *codegen.NameDeclaration {
	name, visibility := grpcValidationName(validator, includeSource)
	if visibility == codegen.UnexportedName {
		return codegen.NewExactName(codegen.NameFunction, name)
	}
	return codegen.NewPreferredName(
		codegen.NameFunction,
		name,
		visibility,
		grpcSymbolOrder(id),
	)
}

// grpcValidationName keeps the released name for a validator called directly
// by generated transport code. A nested validator uses the API, service,
// message, argument, and error context that determine its body. The complete
// method and field source is added only when distinct bodies would otherwise
// receive the same name.
func grpcValidationName(validator *protobufValidationRecord, includeSource bool) (string, codegen.PackageNameVisibility) {
	if validator.source.path == "" {
		name := "Validate" + validator.message.plannedName
		if validator.source.view != "" && validator.source.view != expr.DefaultView {
			name += codegen.Goify(validator.source.view, true)
		}
		return name, codegen.ExportedName
	}

	name := "validate" + grpcValidationNamePart(validator.source.api) +
		"_" + grpcValidationNamePart(validator.source.service) +
		"_" + grpcValidationNamePart(validator.message.plannedName)
	if validator.targetName == validator.contextName {
		name += "_At_" + grpcValidationNamePart(validator.contextName)
	} else {
		name += "_Target_" + grpcValidationNamePart(validator.targetName) +
			"_Context_" + grpcValidationNamePart(validator.contextName)
	}
	if !includeSource {
		return name, codegen.UnexportedName
	}
	name += "_From_" + grpcValidationNamePart(validator.source.method)
	switch validator.source.role {
	case protobufRequestValidation:
		name += "_Request"
	case protobufResponseValidation:
		name += "_Response"
	case protobufErrorValidation:
		name += "_Error_" + grpcValidationNamePart(validator.source.error)
	case protobufStreamingRequestValidation:
		name += "_StreamingRequest"
	default:
		panic("unknown protobuf validation role")
	}
	if validator.source.view != "" && validator.source.view != expr.DefaultView {
		name += "_View_" + grpcValidationNamePart(validator.source.view)
	}
	for _, step := range validator.source.pathSteps {
		name += "_" + grpcValidationPathStepName(step)
	}
	return name, codegen.UnexportedName
}

// grpcValidationKey returns the stable facts used to detect private names that
// need their endpoint source added.
func grpcValidationKey(validator *protobufValidationRecord) grpcValidationNameKey {
	return grpcValidationNameKey{
		side:        validator.side,
		api:         validator.source.api,
		service:     validator.source.service,
		message:     validator.message.plannedName,
		targetName:  validator.targetName,
		contextName: validator.contextName,
	}
}

// grpcValidationPathStepName writes one unambiguous part of a private
// validation function name.
func grpcValidationPathStepName(step protobufValidationPathStep) string {
	switch step.kind {
	case protobufValidationField:
		return "Field_" + grpcValidationNamePart(step.name)
	case protobufValidationArrayElement:
		return "ArrayElement"
	case protobufValidationMapKey:
		return "MapKey"
	case protobufValidationMapValue:
		return "MapValue"
	case protobufValidationUnionBranch:
		return "Union_" + grpcValidationNamePart(step.name)
	default:
		panic("unknown protobuf validation path kind")
	}
}

// grpcValidationNamePart writes one authored label into a Go identifier. ASCII
// letters and digits stay readable. Every other byte uses its hexadecimal value
// between underscores, so two different labels can never produce one spelling.
func grpcValidationNamePart(value string) string {
	if value == "" {
		panic("empty protobuf validation name part")
	}
	const hexadecimal = "0123456789ABCDEF"
	var name strings.Builder
	for index := range len(value) {
		character := value[index]
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' {
			name.WriteByte(character)
			continue
		}
		name.WriteByte('_')
		name.WriteByte(hexadecimal[character>>4])
		name.WriteByte(hexadecimal[character&0x0f])
		name.WriteByte('_')
	}
	return name.String()
}

// newGRPCTransformPlan copies the protobuf input or output type, then records
// every nested conversion function that the generated code will call.
func newGRPCTransformPlan(source, target *expr.AttributeExpr, proto bool, protobuf *protobufServicePlan) (*codegen.TransformPlan, error) {
	prefix := "protobuf"
	if proto {
		original := target
		target = expr.DupAtt(target)
		protobuf.bindAttributeCopy(original, target)
		removeMeta(target)
		prefix = "svc"
	} else {
		original := source
		source = expr.DupAtt(source)
		protobuf.bindAttributeCopy(original, source)
		removeMeta(source)
	}
	return grpcTransformProgram(proto).Plan(source, target, prefix)
}

// planGRPCTransformLayouts records the exact service and protobuf Go types
// supplied to one conversion. The helper registry compares these retained
// declarations and field pointer rules before sharing a private function.
func planGRPCTransformLayouts(
	generation *codegen.Generation,
	input PlanInput,
	method *expr.MethodExpr,
	source, target *expr.AttributeExpr,
	proto bool,
	protobuf *protobufServicePlan,
	pathName string,
) (*codegen.GoTypePlan, *codegen.GoTypePlan, error) {
	serviceAttribute, protobufAttribute := target, source
	if proto {
		serviceAttribute, protobufAttribute = source, target
	}
	serviceLayout, err := input.Service.MethodTypeLayout(method, serviceAttribute)
	if err != nil {
		return nil, nil, err
	}
	protobufLayout, err := planGRPCProtobufLayout(generation, protobuf, protobufAttribute, pathName)
	if err != nil {
		return nil, nil, err
	}
	if proto {
		return serviceLayout, protobufLayout, nil
	}
	return protobufLayout, serviceLayout, nil
}

// planGRPCProtobufLayout records the message, oneof, field, and pointer form
// produced by protoc for attribute before those declarations receive names.
func planGRPCProtobufLayout(generation *codegen.Generation, protobuf *protobufServicePlan, attribute *expr.AttributeExpr, pathName string) (*codegen.GoTypePlan, error) {
	owner := path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName)
	planned := expr.DupAtt(attribute)
	protobuf.bindAttributeCopy(attribute, planned)
	removeMeta(planned)
	if err := codegen.Walk(planned, func(candidate *expr.AttributeExpr) error {
		for {
			userType, ok := candidate.Type.(expr.UserType)
			if !ok || protobuf.catalog.messageRecord(candidate) != nil {
				break
			}
			candidate.Type = userType.Attribute().Type
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("plan protobuf aliases: %w", err)
	}
	return codegen.PlanGoType(planned, codegen.GoTypePlanOptions{
		Owner:            owner,
		RetainNamedValue: true,
		Policy:           codegen.GoLayoutPolicy{Pointer: true},
		Bind: func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
			switch request.Kind {
			case codegen.GoNamed:
				record := protobuf.catalog.messageRecord(request.Attribute)
				if record == nil || record.declaration == nil {
					return codegen.GoTypeBinding{}, fmt.Errorf("protobuf type %q has no generated declaration", request.Attribute.Type.Name())
				}
				return codegen.GoTypeBinding{Owner: owner, Declaration: record.declaration}, nil
			case codegen.GoUnion:
				key, ok := protobuf.oneofs[request.Attribute]
				if !ok || protobuf.names[key] == nil {
					return codegen.GoTypeBinding{}, fmt.Errorf("protobuf oneof %q has no generated declaration", request.Attribute.Type.Name())
				}
				return codegen.GoTypeBinding{Owner: owner, Declaration: protobuf.names[key]}, nil
			default:
				return codegen.GoTypeBinding{}, fmt.Errorf("bind unsupported protobuf Go type kind %s", request.Kind)
			}
		},
	})
}

// ComparePackageName orders generated declarations by package, purpose, API,
// service, method, error, and field.
func (left grpcSymbolOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(grpcSymbolOrder)
	return cmp.Or(
		cmp.Compare(left.side, right.side),
		cmp.Compare(left.role, right.role),
		strings.Compare(left.api, right.api),
		strings.Compare(left.service, right.service),
		strings.Compare(left.method, right.method),
		strings.Compare(left.subject, right.subject),
		strings.Compare(left.view, right.view),
		strings.Compare(left.path, right.path),
		strings.Compare(left.source, right.source),
		strings.Compare(left.target, right.target),
		left.definition.Compare(right.definition),
		cmp.Compare(left.operation, right.operation),
	)
}

// withMethod returns the declaration details for one method in the same
// service.
func (id grpcSymbolID) withMethod(method string) grpcSymbolID {
	id.method = method
	return id
}

// client selects the generated client package and the kind of declaration.
func (id grpcSymbolID) client(role grpcSymbolRole) grpcSymbolID {
	id.side = grpcClientPackage
	id.role = role
	return id
}

// server selects the generated server package and the kind of declaration.
func (id grpcSymbolID) server(role grpcSymbolRole) grpcSymbolID {
	id.side = grpcServerPackage
	id.role = role
	return id
}
