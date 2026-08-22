// This file copies service method, error, streaming, and interceptor membership before generated package names freeze.
package service

import (
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// collectServiceFacts copies service membership and the transport decisions
// that renderers need so linking never consults mutable root collections.
func collectServiceFacts(root *expr.RootExpr, service *expr.ServiceExpr, examples *expr.ExampleGenerator) *serviceFacts {
	facts := &serviceFacts{
		service:        service,
		name:           service.Name,
		description:    service.Description,
		methods:        append([]*expr.MethodExpr(nil), service.Methods...),
		methodByExpr:   make(map[*expr.MethodExpr]*methodFacts, len(service.Methods)),
		errors:         append([]*expr.ErrorExpr(nil), service.Errors...),
		reachableTypes: make(map[expr.UserType]struct{}),
		projections:    make(map[*expr.MethodExpr]*projectionFacts),
	}
	for _, serviceError := range facts.errors {
		facts.errorFacts = append(facts.errorFacts, retainErrorRenderFacts(serviceError))
		facts.referenceAttributes = append(facts.referenceAttributes, serviceError.AttributeExpr)
		retainServiceValueTypes(facts, serviceError.AttributeExpr)
	}
	methodScope := codegen.NewNameScope()
	methodScope.Unique("Use")
	methodScope.Unique("websocket")
	for _, method := range service.Methods {
		methodFacts := &methodFacts{
			method:          method,
			serviceName:     service.Name,
			name:            method.Name,
			description:     method.Description,
			idempotent:      method.Idempotent,
			payload:         retainMethodAttribute(method.Payload, examples.At(expr.MethodPayloadExampleIdentity(method))),
			result:          retainMethodAttribute(method.Result, examples.At(expr.MethodResultExampleIdentity(method))),
			streamKind:      method.Stream,
			isStreaming:     method.IsStreaming(),
			hasMixedResults: method.HasMixedResults(),
			varName:         methodScope.Unique(codegen.Goify(method.Name, true), "Endpoint"),
		}
		methodFacts.streamingPayload = retainMethodAttribute(
			method.StreamingPayload,
			examples.At(expr.MethodStreamingPayloadExampleIdentity(method)),
		)
		methodFacts.streamingResult = retainMethodAttribute(
			method.StreamingResult,
			examples.At(expr.MethodStreamingResultExampleIdentity(method)),
		)
		_, methodFacts.isJSONRPC = method.Meta["jsonrpc"]
		methodFacts.requirements, methodFacts.schemes = retainMethodSecurity(method)
		for _, methodError := range method.Errors {
			methodFacts.errors = append(methodFacts.errors, retainErrorRenderFacts(methodError))
		}
		if method.IsStreaming() || method.HasMixedResults() {
			methodFacts.serverStreamVarName = methodScope.Unique(codegen.Goify(method.Name, true), "ServerStream")
			methodFacts.clientStreamVarName = methodScope.Unique(codegen.Goify(method.Name, true), "ClientStream")
		}
		if _, jsonRPC := method.Meta["jsonrpc"]; jsonRPC && method.IsStreaming() {
			if jsonRPCService := root.API.JSONRPC.HTTPExpr.Service(service.Name); jsonRPCService != nil {
				for _, endpoint := range jsonRPCService.HTTPEndpoints {
					if endpoint.MethodExpr == method {
						methodFacts.isJSONRPCSSE = endpoint.SSE != nil
						methodFacts.isJSONRPCWebSocket = endpoint.SSE == nil
						break
					}
				}
			}
		}
		for _, httpService := range root.API.HTTP.Services {
			if httpService.Name() != service.Name {
				continue
			}
			if endpoint := httpService.Endpoint(method.Name); endpoint != nil {
				methodFacts.skipRequestBodyEncodeDecode = endpoint.SkipRequestBodyEncodeDecode
				methodFacts.skipResponseBodyEncodeDecode = endpoint.SkipResponseBodyEncodeDecode
			}
			break
		}
		facts.methodByExpr[method] = methodFacts
		facts.orderedMethods = append(facts.orderedMethods, methodFacts)
		facts.referenceAttributes = append(
			facts.referenceAttributes,
			method.Payload,
			method.StreamingPayload,
			method.Result,
		)
		retainServiceValueTypes(facts, method.Payload)
		retainServiceValueTypes(facts, method.StreamingPayload)
		retainServiceValueTypes(facts, method.Result)
		if method.HasMixedResults() {
			facts.referenceAttributes = append(facts.referenceAttributes, method.StreamingResult)
			retainServiceValueTypes(facts, method.StreamingResult)
		}
		for _, methodError := range method.Errors {
			facts.referenceAttributes = append(facts.referenceAttributes, methodError.AttributeExpr)
			retainServiceValueTypes(facts, methodError.AttributeExpr)
		}
	}
	for _, method := range facts.methods {
		methodFacts := facts.methodByExpr[method]
		methodFacts.endpointField = methodScope.Unique(methodFacts.varName+"Endpoint", "")
		if method.HasMixedResults() {
			methodFacts.streamEndpointField = methodScope.Unique(methodFacts.varName+"StreamEndpoint", "")
		}
	}
	facts.serverInterceptors = retainedInterceptors(root.API.ServerInterceptors, service.ServerInterceptors, facts.methods, true)
	facts.clientInterceptors = retainedInterceptors(root.API.ClientInterceptors, service.ClientInterceptors, facts.methods, false)
	facts.serverInterceptorFacts = collectInterceptorFacts(facts.serverInterceptors, facts.methods, facts.methodByExpr, true)
	facts.clientInterceptorFacts = collectInterceptorFacts(facts.clientInterceptors, facts.methods, facts.methodByExpr, false)
	return facts
}

// retainServiceValueTypes records every named type reachable from one service
// value contract. External mappings use this set so stream and error values
// receive the same generated conversion ownership as payloads and results.
func retainServiceValueTypes(facts *serviceFacts, attribute *expr.AttributeExpr) {
	if attribute == nil || attribute.Type == expr.Empty {
		return
	}
	err := codegen.Walk(attribute, func(attribute *expr.AttributeExpr) error {
		if userType, ok := attribute.Type.(expr.UserType); ok {
			facts.reachableTypes[userType.Origin()] = struct{}{}
		}
		return nil
	})
	if err != nil {
		panic(err) // the collector callback cannot return an error
	}
}

// collectInterceptorFacts fixes method applicability during planning so
// linking never walks service methods or interceptor expression lists again.
func collectInterceptorFacts(interceptors []*expr.InterceptorExpr, methods []*expr.MethodExpr, methodFacts map[*expr.MethodExpr]*methodFacts, server bool) []*interceptorFacts {
	result := make([]*interceptorFacts, len(interceptors))
	for index, interceptor := range interceptors {
		facts := &interceptorFacts{
			name:                  interceptor.Name,
			description:           interceptor.Description,
			readPayload:           interceptor.ReadPayload,
			writePayload:          interceptor.WritePayload,
			readResult:            interceptor.ReadResult,
			writeResult:           interceptor.WriteResult,
			readStreamingPayload:  interceptor.ReadStreamingPayload,
			writeStreamingPayload: interceptor.WriteStreamingPayload,
			readStreamingResult:   interceptor.ReadStreamingResult,
			writeStreamingResult:  interceptor.WriteStreamingResult,
		}
		for _, method := range methods {
			applied := method.ClientInterceptors
			if server {
				applied = method.ServerInterceptors
			}
			if interceptorNamed(applied, interceptor.Name) {
				facts.methods = append(facts.methods, methodFacts[method])
			}
		}
		result[index] = facts
	}
	return result
}

// retainMethodAttribute copies the top-level method contract and evaluates its
// example during collection. Nested type layout is retained separately by the
// generated Go type plan.
func retainMethodAttribute(attribute *expr.AttributeExpr, examples *expr.ExampleGenerator) *methodAttributeFacts {
	if attribute == nil {
		return nil
	}
	retained := *attribute
	if attribute.Meta != nil {
		retained.Meta = attribute.Meta.Dup()
	}
	return &methodAttributeFacts{
		attribute:    &retained,
		present:      attribute.Type != expr.Empty,
		isObject:     expr.IsObject(attribute.Type),
		location:     codegen.UserTypeLocation(attribute.Type),
		description:  attribute.Description,
		defaultValue: cloneRetainedValue(attribute.DefaultValue),
		example:      cloneRetainedValue(attribute.Example(examples)),
	}
}

// retainErrorRenderFacts copies the error text, type wrapper, location, and
// marker flags that generated constructors and client comments consume.
func retainErrorRenderFacts(errorExpression *expr.ErrorExpr) *errorRenderFacts {
	_, temporary := errorExpression.Meta["goa:error:temporary"]
	_, timeout := errorExpression.Meta["goa:error:timeout"]
	_, fault := errorExpression.Meta["goa:error:fault"]
	attribute := *errorExpression.AttributeExpr
	if errorExpression.Meta != nil {
		attribute.Meta = errorExpression.Meta.Dup()
	}
	return &errorRenderFacts{
		attribute:   &attribute,
		name:        errorExpression.Name,
		description: errorExpression.Description,
		location:    codegen.UserTypeLocation(errorExpression.Type),
		temporary:   temporary,
		timeout:     timeout,
		fault:       fault,
		serviceType: errorExpression.Type == expr.ErrorResult,
	}
}

// retainMethodSecurity evaluates scheme credential fields and scopes while
// the finalized method payload and requirements are still collection inputs.
func retainMethodSecurity(method *expr.MethodExpr) (RequirementsData, SchemesData) {
	requirements := make(RequirementsData, 0, len(method.Requirements))
	var schemes SchemesData
	for _, requirement := range expr.EffectiveSecurityRequirements(method.Requirements) {
		var requirementSchemes SchemesData
		for _, scheme := range requirement.Schemes {
			data := cloneSchemeData(BuildSchemeData(scheme, method))
			requirementSchemes = requirementSchemes.Append(data)
			schemes = schemes.Append(data)
		}
		requirements = append(requirements, &RequirementData{
			Schemes: requirementSchemes,
			Scopes:  append([]string(nil), requirement.Scopes...),
		})
	}
	return requirements, schemes
}

// cloneSchemeData detaches the collection values retained in one scheme.
func cloneSchemeData(source *SchemeData) *SchemeData {
	if source == nil {
		return nil
	}
	cloned := *source
	cloned.Scopes = append([]string(nil), source.Scopes...)
	cloned.Flows = make([]*expr.FlowExpr, len(source.Flows))
	for index, flow := range source.Flows {
		copy := *flow
		cloned.Flows[index] = &copy
	}
	return &cloned
}

// cloneRetainedValue copies the collection shapes accepted by Goa examples
// and defaults. Primitive values are immutable and may be shared.
func cloneRetainedValue(source any) any {
	switch actual := source.(type) {
	case expr.Val:
		cloned := make(expr.Val, len(actual))
		for name, value := range actual {
			cloned[name] = cloneRetainedValue(value)
		}
		return cloned
	case expr.ArrayVal:
		cloned := make(expr.ArrayVal, len(actual))
		for index, value := range actual {
			cloned[index] = cloneRetainedValue(value)
		}
		return cloned
	case expr.MapVal:
		cloned := make(expr.MapVal, len(actual))
		for key, value := range actual {
			cloned[cloneRetainedValue(key)] = cloneRetainedValue(value)
		}
		return cloned
	case []any:
		cloned := make([]any, len(actual))
		for index, value := range actual {
			cloned[index] = cloneRetainedValue(value)
		}
		return cloned
	case []byte:
		return append([]byte(nil), actual...)
	case map[string]any:
		cloned := make(map[string]any, len(actual))
		for name, value := range actual {
			cloned[name] = cloneRetainedValue(value)
		}
		return cloned
	case map[any]any:
		cloned := make(map[any]any, len(actual))
		for key, value := range actual {
			cloned[cloneRetainedValue(key)] = cloneRetainedValue(value)
		}
		return cloned
	default:
		return actual
	}
}

// retainedInterceptors returns one stable, name-ordered interceptor set without
// sorting or appending into any expression-owned slice.
func retainedInterceptors(api, service []*expr.InterceptorExpr, methods []*expr.MethodExpr, server bool) []*expr.InterceptorExpr {
	interceptors := append([]*expr.InterceptorExpr(nil), api...)
	interceptors = append(interceptors, service...)
	for _, method := range methods {
		if server {
			interceptors = append(interceptors, method.ServerInterceptors...)
		} else {
			interceptors = append(interceptors, method.ClientInterceptors...)
		}
	}
	sort.Slice(interceptors, func(i, j int) bool {
		return interceptors[i].Name < interceptors[j].Name
	})
	result := interceptors[:0]
	for _, interceptor := range interceptors {
		if len(result) == 0 || result[len(result)-1].Name != interceptor.Name {
			result = append(result, interceptor)
		}
	}
	return result
}
