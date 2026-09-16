// This file verifies that copied HTTP request and response fields still point
// to the service fields that supplied their descriptions and examples.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPBodyAttributesKeepAuthoredAttribute(t *testing.T) {
	payloadChild := &AttributeExpr{Type: String}
	payload := &AttributeExpr{Type: &Object{{Name: "message", Attribute: payloadChild}}}
	resultChild := &AttributeExpr{Type: String}
	result := &AttributeExpr{Type: &Object{{Name: "message", Attribute: resultChild}}}
	serviceMethod := &ServiceExpr{Name: "messages"}
	serviceMethod.design = &RootExpr{API: NewAPIExpr("messages", nil)}
	method := &MethodExpr{
		Name:    "show",
		Service: serviceMethod,
		Payload: payload,
		Result:  result,
	}
	service := &HTTPServiceExpr{ServiceExpr: serviceMethod}
	endpoint := &HTTPEndpointExpr{
		MethodExpr: method,
		Service:    service,
		Params:     NewEmptyMappedAttributeExpr(),
		Headers:    NewEmptyMappedAttributeExpr(),
		Cookies:    NewEmptyMappedAttributeExpr(),
	}
	response := &HTTPResponseExpr{
		Headers: NewEmptyMappedAttributeExpr(),
		Cookies: NewEmptyMappedAttributeExpr(),
	}

	requestBody := httpRequestBody(endpoint)
	responseBody := buildHTTPResponseBody("show", result, response, MethodResultExampleIdentity(method))

	require.Same(t, payload, requestBody.AuthoredAttribute())
	require.Same(t, result, responseBody.AuthoredAttribute())
	require.Same(t, payloadChild, AsObject(requestBody.Type).Attribute("message").AuthoredAttribute())
	require.Same(t, resultChild, AsObject(responseBody.Type).Attribute("message").AuthoredAttribute())
}

func TestHTTPStreamingBodyKeepsAuthoredAttribute(t *testing.T) {
	streaming := &AttributeExpr{Type: &Object{{Name: "message", Attribute: &AttributeExpr{Type: String}}}}
	method := &MethodExpr{
		Name:             "watch",
		Service:          &ServiceExpr{Name: "events"},
		Payload:          &AttributeExpr{Type: Empty},
		Result:           &AttributeExpr{Type: Empty},
		StreamingPayload: streaming,
		Stream:           ClientStreamKind,
	}
	endpoint := &HTTPEndpointExpr{MethodExpr: method}

	body := httpStreamingBody(endpoint)

	require.Same(t, streaming, body.AuthoredAttribute())
}

func TestHTTPPlaceholderKeepsAuthoredAttribute(t *testing.T) {
	authored := &AttributeExpr{Type: String, Description: "description"}
	placeholder := &AttributeExpr{Type: String}

	initAttrFromDesign(placeholder, authored)

	require.Same(t, authored, placeholder.AuthoredAttribute())
	require.Equal(t, "description", placeholder.Description)
}
