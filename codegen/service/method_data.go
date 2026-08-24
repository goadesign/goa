// This file builds the template data for service methods and their streams from
// the values and Go declarations recorded during planning.
package service

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// buildMethodData formats one method using the names and types chosen during
// service planning.
func buildMethodData(facts *methodFacts, resolver *declarationResolver, serviceFacts *serviceFacts) *MethodData {
	var (
		vname       string
		desc        string
		payloadName string
		payloadLoc  *codegen.Location
		payloadDef  string
		payloadRef  string
		payloadDesc string
		payloadEx   any
		rname       string
		resultLoc   *codegen.Location
		resultDef   string
		resultRef   string
		resultDesc  string
		resultEx    any
		errors      []*ErrorInitData
		errorLocs   map[string]*codegen.Location
		reqs        = facts.requirements
		schemes     = facts.schemes
	)
	vname = facts.varName
	desc = facts.description
	if desc == "" {
		desc = codegen.Goify(facts.name, true) + " implements " + facts.name + "."
	}
	if facts.payload != nil && facts.payload.present {
		payloadLoc = facts.payload.location
		payloadName, payloadDef, payloadRef = retainedMethodTypeData(facts.payload, resolver)
		payloadDesc = facts.payload.description
		if payloadDesc == "" {
			payloadDesc = fmt.Sprintf("%s is the payload type of the %s service %s method.",
				payloadName, serviceFacts.name, facts.name)
		}
		payloadEx = facts.payload.example
	}
	if facts.result != nil && facts.result.present {
		resultLoc = facts.result.location
		rname, resultDef, resultRef = retainedMethodTypeData(facts.result, resolver)
		resultDesc = facts.result.description
		if resultDesc == "" {
			resultDesc = fmt.Sprintf("%s is the result type of the %s service %s method.",
				rname, serviceFacts.name, facts.name)
		}
		resultEx = facts.result.example
	}
	if len(facts.errors) > 0 {
		errors = make([]*ErrorInitData, len(facts.errors))
		errorLocs = make(map[string]*codegen.Location, len(facts.errors))
		for i, errorFacts := range facts.errors {
			errors[i] = buildRetainedErrorInitData(errorFacts, resolver, serviceFacts.errorConstructors[errorFacts.name])
			errorLocs[errorFacts.name] = errorFacts.location
		}
	}
	data := &MethodData{
		EndpointDeclaration: serviceFacts.names.declaration(serviceSymbolID{
			role: serviceMethodEndpointNameRole, service: serviceFacts.name, method: facts.varName,
		}),
		EndpointInputDeclaration: serviceFacts.names[serviceSymbolID{
			role: serviceEndpointInputNameRole, service: serviceFacts.name, method: facts.varName,
		}].declaration,
		ServerStreamDeclaration: serviceFacts.names[serviceSymbolID{
			role: serviceServerStreamNameRole, service: serviceFacts.name, method: facts.varName,
		}].declaration,
		ClientStreamDeclaration: serviceFacts.names[serviceSymbolID{
			role: serviceClientStreamNameRole, service: serviceFacts.name, method: facts.varName,
		}].declaration,
		RequestDeclaration: serviceFacts.names[serviceSymbolID{
			role: serviceRequestNameRole, service: serviceFacts.name, method: facts.varName,
		}].declaration,
		ResponseDeclaration: serviceFacts.names[serviceSymbolID{
			role: serviceResponseNameRole, service: serviceFacts.name, method: facts.varName,
		}].declaration,
		ServerEndpointWrapperDeclaration: serviceFacts.names[serviceSymbolID{
			role: serviceServerEndpointWrapperNameRole, service: serviceFacts.name, method: facts.varName,
		}].declaration,
		ClientEndpointWrapperDeclaration: serviceFacts.names[serviceSymbolID{
			role: serviceClientEndpointWrapperNameRole, service: serviceFacts.name, method: facts.varName,
		}].declaration,
		Name:                         facts.name,
		VarName:                      vname,
		Description:                  desc,
		Idempotent:                   facts.idempotent,
		Payload:                      payloadName,
		PayloadLoc:                   payloadLoc,
		PayloadDef:                   payloadDef,
		PayloadRef:                   payloadRef,
		PayloadDeclaration:           facts.payload.layout.TypeDeclaration(),
		PayloadDesc:                  payloadDesc,
		PayloadEx:                    payloadEx,
		PayloadDefault:               facts.payload.defaultValue,
		Result:                       rname,
		ResultLoc:                    resultLoc,
		ResultDef:                    resultDef,
		ResultRef:                    resultRef,
		ResultDeclaration:            facts.result.layout.TypeDeclaration(),
		ResultDesc:                   resultDesc,
		ResultEx:                     resultEx,
		Errors:                       errors,
		ErrorLocs:                    errorLocs,
		Requirements:                 reqs,
		Schemes:                      schemes,
		StreamKind:                   facts.streamKind,
		HasMixedResults:              facts.hasMixedResults,
		SkipRequestBodyEncodeDecode:  facts.skipRequestBodyEncodeDecode,
		SkipResponseBodyEncodeDecode: facts.skipResponseBodyEncodeDecode,
		RequestStruct:                vname + "RequestData",
		ResponseStruct:               vname + "ResponseData",
		EndpointField:                facts.endpointField,
		StreamEndpointField:          facts.streamEndpointField,
	}

	initStreamData(data, facts, resolver)
	return data
}

// initStreamData initializes the streaming payload data structures and methods.
func initStreamData(data *MethodData, facts *methodFacts, resolver *declarationResolver) {
	if !facts.isStreaming && !facts.hasMixedResults {
		return
	}
	var (
		spayloadName string
		spayloadRef  string
		spayloadDef  string
		spayloadDesc string
		spayloadEx   any
		srname       string
		srref        string
		srdef        string
	)
	if facts.streamingResult != nil && facts.streamingResult.present {
		srname, srdef, srref = retainedMethodTypeData(facts.streamingResult, resolver)
	}
	data.StreamingResultRef = srref

	// Mixed-result methods return StreamingResult from their streaming endpoint
	// and Result from their ordinary endpoint.
	if facts.hasMixedResults && facts.streamingResult != nil && facts.streamingResult.present {
		data.StreamingResult = srname
		data.StreamingResultDef = srdef
		data.StreamingResultDeclaration = facts.streamingResult.layout.TypeDeclaration()
		data.StreamingResultDesc = facts.streamingResult.description
		if data.StreamingResultDesc == "" {
			data.StreamingResultDesc = fmt.Sprintf("%s is the streaming result type of the %s service %s method.",
				srname, facts.serviceName, facts.name)
		}
		data.StreamingResultEx = facts.streamingResult.example
	}

	if facts.streamingPayload != nil && facts.streamingPayload.present {
		spayloadName, spayloadDef, spayloadRef = retainedMethodTypeData(facts.streamingPayload, resolver)
		data.StreamingPayloadDeclaration = facts.streamingPayload.layout.TypeDeclaration()
		spayloadDesc = facts.streamingPayload.description
		if spayloadDesc == "" {
			spayloadDesc = fmt.Sprintf("%s is the streaming payload type of the %s service %s method.",
				spayloadName, facts.serviceName, facts.name)
		}
		spayloadEx = facts.streamingPayload.example
	}
	// Streaming endpoint calls carry the request value and stream together.
	var endpointStruct string
	if data.EndpointInputDeclaration != nil {
		endpointStruct = data.EndpointInputDeclaration.Name()
	}
	// A mixed-result SSE method sends results from the server even though its
	// service method is not otherwise marked as streaming.
	streamKind := facts.streamKind
	if facts.hasMixedResults && !facts.isStreaming {
		streamKind = expr.ServerStreamKind
	}
	svrStream := &StreamData{
		Interface:           data.ServerStreamDeclaration.Name(),
		VarName:             facts.serverStreamVarName,
		EndpointStruct:      endpointStruct,
		Kind:                streamKind,
		SendName:            "Send",
		SendDesc:            fmt.Sprintf("Send streams instances of %q.", srname),
		SendWithContextName: "SendWithContext",
		SendWithContextDesc: fmt.Sprintf("SendWithContext streams instances of %q with context.", srname),
		SendTypeName:        srname,
		SendTypeRef:         srref,
		MustClose:           true,
	}
	cliStream := &StreamData{
		Interface:           data.ClientStreamDeclaration.Name(),
		VarName:             facts.clientStreamVarName,
		Kind:                streamKind,
		RecvName:            "Recv",
		RecvDesc:            fmt.Sprintf("Recv reads instances of %q from the stream.", srname),
		RecvWithContextName: "RecvWithContext",
		RecvWithContextDesc: fmt.Sprintf("RecvWithContext reads instances of %q from the stream with context.", srname),
		RecvTypeName:        srname,
		RecvTypeRef:         srref,
	}
	if streamKind == expr.ClientStreamKind || streamKind == expr.BidirectionalStreamKind {
		switch streamKind {
		case expr.ClientStreamKind:
			if srref != "" {
				svrStream.SendName = "SendAndClose"
				svrStream.SendDesc = fmt.Sprintf("SendAndClose streams instances of %q and closes the stream.", srname)
				svrStream.SendWithContextName = "SendAndCloseWithContext"
				svrStream.SendWithContextDesc = fmt.Sprintf("SendAndCloseWithContext streams instances of %q and closes the stream with context.", srname)
				svrStream.MustClose = false
				cliStream.RecvName = "CloseAndRecv"
				cliStream.RecvDesc = fmt.Sprintf("CloseAndRecv stops sending messages to the stream and reads instances of %q from the stream.", srname)
				cliStream.RecvWithContextName = "CloseAndRecvWithContext"
				cliStream.RecvWithContextDesc = fmt.Sprintf("CloseAndRecvWithContext stops sending messages to the stream and reads instances of %q from the stream with context.", srname)
			} else {
				cliStream.MustClose = true
			}
		case expr.BidirectionalStreamKind:
			cliStream.MustClose = true
		}
		svrStream.RecvName = "Recv"
		svrStream.RecvDesc = fmt.Sprintf("Recv reads instances of %q from the stream.", spayloadName)
		svrStream.RecvWithContextName = "RecvWithContext"
		svrStream.RecvWithContextDesc = fmt.Sprintf("RecvWithContext reads instances of %q from the stream with context.", spayloadName)
		svrStream.RecvTypeName = spayloadName
		svrStream.RecvTypeRef = spayloadRef
		cliStream.SendName = "Send"
		cliStream.SendDesc = fmt.Sprintf("Send streams instances of %q.", spayloadName)
		cliStream.SendWithContextName = "SendWithContext"
		cliStream.SendWithContextDesc = fmt.Sprintf("SendWithContext streams instances of %q with context.", spayloadName)
		cliStream.SendTypeName = spayloadName
		cliStream.SendTypeRef = spayloadRef
	}
	data.ClientStream = cliStream
	data.ServerStream = svrStream
	data.StreamingPayload = spayloadName
	data.StreamingPayloadDef = spayloadDef
	data.StreamingPayloadRef = spayloadRef
	data.StreamingPayloadDesc = spayloadDesc
	data.StreamingPayloadEx = spayloadEx
}

// This helper returns the Go name, definition, and reference for one payload or
// result relative to the service output package. It reads the type layout
// copied during planning instead of rereading the design expression.
func retainedMethodTypeData(facts *methodAttributeFacts, resolver *declarationResolver) (string, string, string) {
	linked := facts.layout.Link(resolver.outputPath, retainedTypeQualifier(resolver.aliases, resolver.outputPath))
	definition := ""
	if facts.definition != nil {
		definition = facts.definition.Link(facts.layout.Owner(), retainedTypeQualifier(resolver.aliases, facts.layout.Owner())).Def()
	}
	return linked.Name(), definition, linked.Ref()
}
