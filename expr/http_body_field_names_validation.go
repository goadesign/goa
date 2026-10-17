// This file checks mapped HTTP unions before finalization. It builds temporary
// bodies using the same request and response selection as generation, then
// applies the existing union rules without modifying service declarations.
package expr

import "goa.design/goa/v3/eval"

func (svc *HTTPServiceExpr) validateMappedBodyUnions(errors *eval.ValidationErrors) {
	for _, original := range svc.HTTPEndpoints {
		endpoint := *original
		method := *original.MethodExpr
		endpoint.MethodExpr = &method
		method.Payload = DupAtt(method.Payload)
		method.Payload.Finalize()
		if endpoint.Body != nil {
			endpoint.Body = DupAtt(endpoint.Body)
			endpoint.Body.Finalize()
		}
		source := bodyFieldNameSource(endpoint.Body, method.Payload)
		svc.validateMappedBodyUnion(httpRequestBody(&endpoint), source, errors)
		if method.StreamingPayload != nil {
			method.StreamingPayload = DupAtt(method.StreamingPayload)
			method.StreamingPayload.Finalize()
			if body := httpStreamingBody(&endpoint); body != nil {
				svc.validateMappedBodyUnion(body, bodyFieldNameSource(nil, method.StreamingPayload), errors)
			}
		}
		for _, response := range endpoint.Responses {
			svc.validateMappedResponseUnion(&endpoint, method.Result, response, errors)
		}
		for _, failure := range endpoint.HTTPErrors {
			if declared := method.Error(failure.Name); declared != nil {
				svc.validateMappedResponseUnion(&endpoint, declared.AttributeExpr, failure.Response, errors)
			}
		}
	}
}

func (svc *HTTPServiceExpr) validateMappedResponseUnion(endpoint *HTTPEndpointExpr, attribute *AttributeExpr, response *HTTPResponseExpr, errors *eval.ValidationErrors) {
	if attribute == nil {
		return
	}
	attribute = DupAtt(attribute)
	attribute.Finalize()
	copy := *response
	if copy.Body != nil {
		copy.Body = DupAtt(copy.Body)
		copy.Body.Finalize()
	}
	source := bodyFieldNameSource(copy.Body, attribute)
	body := buildHTTPResponseBody(endpoint.Name(), attribute, &copy, ResponseBodyExampleIdentity(endpoint, &copy))
	svc.validateMappedBodyUnion(body, source, errors)
}

func (svc *HTTPServiceExpr) validateMappedBodyUnion(body *AttributeExpr, source UserType, errors *eval.ValidationErrors) {
	body.Finalize()
	mapped := svc.mapBodyFieldNames(body, source)
	validateMappedUnions(mapped, svc, errors, make(map[UserType]bool))
}

// validateMappedUnions checks each completed union once. The existing validator
// rejects a branch property that would overwrite its discriminator in JSON.
func validateMappedUnions(attribute *AttributeExpr, service *HTTPServiceExpr, errors *eval.ValidationErrors, seen map[UserType]bool) {
	switch actual := attribute.Type.(type) {
	case UserType:
		if seen[actual] {
			return
		}
		seen[actual] = true
		validateMappedUnions(actual.Attribute(), service, errors, seen)
	case *Object:
		for _, field := range *actual {
			validateMappedUnions(field.Attribute, service, errors, seen)
		}
	case *Array:
		validateMappedUnions(actual.ElemType, service, errors, seen)
	case *Map:
		validateMappedUnions(actual.ElemType, service, errors, seen)
	case *Union:
		errors.Merge(validateUnionJSON(attribute, service))
		for _, branch := range actual.Values {
			validateMappedUnions(branch.Attribute, service, errors, seen)
		}
	}
}
