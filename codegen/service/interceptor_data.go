// This file formats retained interceptor applicability and access facts for interceptor templates.
package service

import (
	"goa.design/goa/v3/codegen"
)

// buildInterceptorData creates the data needed to generate interceptor code.
func buildInterceptorData(service *serviceFacts, facts *interceptorFacts, methods map[*methodFacts]*MethodData, resolver *declarationResolver, server bool) *InterceptorData {
	lookup := func(role serviceNameRole, method, subject string) *codegen.NameDeclaration {
		return service.names[serviceSymbolID{
			role: role, service: service.name, method: method, subject: subject,
		}].declaration
	}
	data := &InterceptorData{
		InfoDeclaration:             lookup(serviceInterceptorInfoNameRole, "", facts.name),
		PayloadDeclaration:          lookup(serviceInterceptorPayloadNameRole, "", facts.name),
		ResultDeclaration:           lookup(serviceInterceptorResultNameRole, "", facts.name),
		StreamingPayloadDeclaration: lookup(serviceInterceptorStreamingPayloadNameRole, "", facts.name),
		StreamingResultDeclaration:  lookup(serviceInterceptorStreamingResultNameRole, "", facts.name),
		Name:                        codegen.Goify(facts.name, true),
		DesignName:                  facts.name,
		Description:                 facts.description,
	}
	if len(facts.methods) == 0 {
		return data
	}
	data.ReadPayload = formatInterceptorAccess(facts.readPayloadFields, resolver)
	data.WritePayload = formatInterceptorAccess(facts.writePayloadFields, resolver)
	data.ReadResult = formatInterceptorAccess(facts.readResultFields, resolver)
	data.WriteResult = formatInterceptorAccess(facts.writeResultFields, resolver)
	data.ReadStreamingPayload = formatInterceptorAccess(facts.readStreamingPayloadFields, resolver)
	data.WriteStreamingPayload = formatInterceptorAccess(facts.writeStreamingPayloadFields, resolver)
	data.ReadStreamingResult = formatInterceptorAccess(facts.readStreamingResultFields, resolver)
	data.WriteStreamingResult = formatInterceptorAccess(facts.writeStreamingResultFields, resolver)
	data.HasPayloadAccess = len(data.ReadPayload) > 0 || len(data.WritePayload) > 0
	data.HasResultAccess = len(data.ReadResult) > 0 || len(data.WriteResult) > 0
	data.HasStreamingPayloadAccess = len(data.ReadStreamingPayload) > 0 || len(data.WriteStreamingPayload) > 0
	data.HasStreamingResultAccess = len(data.ReadStreamingResult) > 0 || len(data.WriteStreamingResult) > 0
	for _, method := range facts.methods {
		md := methods[method]
		data.Methods = append(data.Methods, buildInterceptorMethodData(service, facts.name, md))
		if server {
			md.ServerInterceptors = append(md.ServerInterceptors, facts.name)
		} else {
			md.ClientInterceptors = append(md.ClientInterceptors, facts.name)
		}
	}
	return data
}

// formatInterceptorAccess resolves the frozen type spelling for fields chosen
// during interceptor planning.
func formatInterceptorAccess(facts []*interceptorAccessFacts, resolver *declarationResolver) []*AttributeData {
	if len(facts) == 0 {
		return nil
	}
	data := make([]*AttributeData, len(facts))
	for index, field := range facts {
		data[index] = &AttributeData{
			Name:    field.name,
			TypeRef: field.layout.Link(resolver.outputPath, retainedTypeQualifier(resolver.aliases)).Ref(),
			Pointer: field.pointer,
		}
	}
	return data
}

// buildInterceptorMethodData creates the data needed to generate interceptor
// method code.
func buildInterceptorMethodData(service *serviceFacts, interceptorName string, md *MethodData) *MethodInterceptorData {
	declaration := func(role serviceNameRole) *codegen.NameDeclaration {
		return service.names[serviceSymbolID{
			role: role, service: service.name, method: md.VarName, subject: interceptorName,
		}].declaration
	}
	var serverStream, clientStream *StreamInterceptorData
	if md.ServerStream != nil {
		serverStream = &StreamInterceptorData{
			InterfaceDeclaration: md.ServerStreamDeclaration,
			WrapperDeclaration: service.names[serviceSymbolID{
				role: serviceServerStreamWrapperNameRole, service: service.name, method: md.VarName,
			}].declaration,
			Interface:           md.ServerStream.Interface,
			SendName:            md.ServerStream.SendName,
			SendWithContextName: md.ServerStream.SendWithContextName,
			SendTypeRef:         md.ServerStream.SendTypeRef,
			RecvName:            md.ServerStream.RecvName,
			RecvWithContextName: md.ServerStream.RecvWithContextName,
			RecvTypeRef:         md.ServerStream.RecvTypeRef,
			MustClose:           md.ServerStream.MustClose,
			EndpointStruct:      md.ServerStream.EndpointStruct,
		}
	}
	if md.ClientStream != nil {
		clientStream = &StreamInterceptorData{
			InterfaceDeclaration: md.ClientStreamDeclaration,
			WrapperDeclaration: service.names[serviceSymbolID{
				role: serviceClientStreamWrapperNameRole, service: service.name, method: md.VarName,
			}].declaration,
			Interface:           md.ClientStream.Interface,
			SendName:            md.ClientStream.SendName,
			SendWithContextName: md.ClientStream.SendWithContextName,
			SendTypeRef:         md.ClientStream.SendTypeRef,
			RecvName:            md.ClientStream.RecvName,
			RecvWithContextName: md.ClientStream.RecvWithContextName,
			RecvTypeRef:         md.ClientStream.RecvTypeRef,
			MustClose:           md.ClientStream.MustClose,
		}
	}
	payloadAccessDeclaration := declaration(serviceInterceptorPayloadAccessNameRole)
	resultAccessDeclaration := declaration(serviceInterceptorResultAccessNameRole)
	streamingPayloadAccessDeclaration := declaration(serviceInterceptorStreamingPayloadAccessNameRole)
	streamingResultAccessDeclaration := declaration(serviceInterceptorStreamingResultAccessNameRole)
	var payloadAccess, resultAccess, streamingPayloadAccess, streamingResultAccess string
	if payloadAccessDeclaration != nil {
		payloadAccess = payloadAccessDeclaration.Name()
	}
	if resultAccessDeclaration != nil {
		resultAccess = resultAccessDeclaration.Name()
	}
	if streamingPayloadAccessDeclaration != nil {
		streamingPayloadAccess = streamingPayloadAccessDeclaration.Name()
	}
	if streamingResultAccessDeclaration != nil {
		streamingResultAccess = streamingResultAccessDeclaration.Name()
	}
	return &MethodInterceptorData{
		PayloadAccessDeclaration:          payloadAccessDeclaration,
		ResultAccessDeclaration:           resultAccessDeclaration,
		StreamingPayloadAccessDeclaration: streamingPayloadAccessDeclaration,
		StreamingResultAccessDeclaration:  streamingResultAccessDeclaration,
		ServerWrapperDeclaration:          declaration(serviceServerInterceptorWrapperNameRole),
		ClientWrapperDeclaration:          declaration(serviceClientInterceptorWrapperNameRole),
		MethodName:                        md.VarName,
		PayloadAccess:                     payloadAccess,
		ResultAccess:                      resultAccess,
		PayloadRef:                        md.PayloadRef,
		ResultRef:                         md.ResultRef,
		StreamingPayloadAccess:            streamingPayloadAccess,
		StreamingPayloadRef:               md.StreamingPayloadRef,
		StreamingResultAccess:             streamingResultAccess,
		StreamingResultRef:                md.ResultRef,
		ClientStream:                      clientStream,
		ServerStream:                      serverStream,
	}
}
