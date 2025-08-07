package expr_test

import (
	"errors"
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestJSONRPCTransportConsistency(t *testing.T) {
	cases := []struct {
		Name     string
		Setup    func() *expr.HTTPServiceExpr
		WantErr  bool
		ErrorMsg string
	}{
		{
			Name: "valid HTTP and SSE mix",
			Setup: func() *expr.HTTPServiceExpr {
				service := &expr.ServiceExpr{
					Name: "TestService",
					Meta: expr.MetaExpr{"jsonrpc:service": []string{}},
				}
				
				httpService := &expr.HTTPServiceExpr{
					ServiceExpr: service,
					Root:        &expr.HTTPExpr{},
				}
				
				// Regular HTTP method
				m1 := &expr.MethodExpr{
					Name:    "GetUser",
					Service: service,
					Payload: &expr.AttributeExpr{Type: expr.String},
					Result:  &expr.AttributeExpr{Type: expr.String},
				}
				e1 := &expr.HTTPEndpointExpr{
					MethodExpr: m1,
					Service:    httpService,
					Meta:       expr.MetaExpr{"jsonrpc": []string{}},
					Headers:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Cookies:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Params:     &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
				}
				
				// SSE streaming method
				m2 := &expr.MethodExpr{
					Name:    "WatchUsers",
					Service: service,
					Payload: &expr.AttributeExpr{Type: expr.String},
					Result:  &expr.AttributeExpr{Type: expr.String},
					Stream:  expr.ServerStreamKind,
				}
				e2 := &expr.HTTPEndpointExpr{
					MethodExpr: m2,
					Service:    httpService,
					Meta:       expr.MetaExpr{"jsonrpc": []string{}},
					Headers:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Cookies:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Params:     &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					SSE:        &expr.HTTPSSEExpr{},
				}
				
				httpService.HTTPEndpoints = []*expr.HTTPEndpointExpr{e1, e2}
				return httpService
			},
			WantErr: false,
		},
		{
			Name: "invalid WebSocket and HTTP mix",
			Setup: func() *expr.HTTPServiceExpr {
				service := &expr.ServiceExpr{
					Name: "TestService",
					Meta: expr.MetaExpr{"jsonrpc:service": []string{}},
				}
				
				httpService := &expr.HTTPServiceExpr{
					ServiceExpr: service,
					Root:        &expr.HTTPExpr{},
				}
				
				// WebSocket streaming method
				m1 := &expr.MethodExpr{
					Name:             "Stream",
					Service:          service,
					StreamingPayload: &expr.AttributeExpr{Type: expr.String},
					Result:           &expr.AttributeExpr{Type: expr.String},
					Stream:           expr.BidirectionalStreamKind,
				}
				e1 := &expr.HTTPEndpointExpr{
					MethodExpr: m1,
					Service:    httpService,
					Meta:       expr.MetaExpr{"jsonrpc": []string{}},
					Headers:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Cookies:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Params:     &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
				}
				
				// Regular HTTP method
				m2 := &expr.MethodExpr{
					Name:    "Get",
					Service: service,
					Payload: &expr.AttributeExpr{Type: expr.String},
					Result:  &expr.AttributeExpr{Type: expr.String},
				}
				e2 := &expr.HTTPEndpointExpr{
					MethodExpr: m2,
					Service:    httpService,
					Meta:       expr.MetaExpr{"jsonrpc": []string{}},
					Headers:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Cookies:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Params:     &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
				}
				
				httpService.HTTPEndpoints = []*expr.HTTPEndpointExpr{e1, e2}
				return httpService
			},
			WantErr:  true,
			ErrorMsg: "cannot mix WebSocket with other transports",
		},
		{
			Name: "invalid WebSocket and SSE mix",
			Setup: func() *expr.HTTPServiceExpr {
				service := &expr.ServiceExpr{
					Name: "TestService",
					Meta: expr.MetaExpr{"jsonrpc:service": []string{}},
				}
				
				httpService := &expr.HTTPServiceExpr{
					ServiceExpr: service,
					Root:        &expr.HTTPExpr{},
				}
				
				// WebSocket streaming method
				m1 := &expr.MethodExpr{
					Name:             "Stream",
					Service:          service,
					StreamingPayload: &expr.AttributeExpr{Type: expr.String},
					Result:           &expr.AttributeExpr{Type: expr.String},
					Stream:           expr.BidirectionalStreamKind,
				}
				e1 := &expr.HTTPEndpointExpr{
					MethodExpr: m1,
					Service:    httpService,
					Meta:       expr.MetaExpr{"jsonrpc": []string{}},
					Headers:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Cookies:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Params:     &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
				}
				
				// SSE streaming method
				m2 := &expr.MethodExpr{
					Name:    "Watch",
					Service: service,
					Payload: &expr.AttributeExpr{Type: expr.String},
					Result:  &expr.AttributeExpr{Type: expr.String},
					Stream:  expr.ServerStreamKind,
				}
				e2 := &expr.HTTPEndpointExpr{
					MethodExpr: m2,
					Service:    httpService,
					Meta:       expr.MetaExpr{"jsonrpc": []string{}},
					Headers:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Cookies:    &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					Params:     &expr.MappedAttributeExpr{AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}},
					SSE:        &expr.HTTPSSEExpr{},
				}
				
				httpService.HTTPEndpoints = []*expr.HTTPEndpointExpr{e1, e2}
				return httpService
			},
			WantErr:  true,
			ErrorMsg: "cannot mix WebSocket with other transports",
		},
	}
	
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			svc := c.Setup()
			err := svc.Validate()
			
			if c.WantErr {
				if err == nil {
					t.Errorf("expected error containing %q but got none", c.ErrorMsg)
				} else if !containsStr(err.Error(), c.ErrorMsg) {
					t.Errorf("expected error containing %q but got %q", c.ErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					// Check if it's a ValidationErrors with no actual errors
					var verr *eval.ValidationErrors
					if errors.As(err, &verr) && len(verr.Errors) == 0 {
						// Empty validation errors, ignore
					} else {
						t.Logf("Error type: %T", err)
						t.Errorf("unexpected error: %v", err)
					}
				}
			}
		})
	}
}

func containsStr(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}