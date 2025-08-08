package openapiv3

import (
	"testing"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestWebhookGeneration(t *testing.T) {
	// Define a simple service with a webhook
	root := runWebhookDSL(t, func() {
		dsl.API("test", func() {
			dsl.Title("Test API")
			dsl.Version("1.0")
		})
		
		dsl.Service("TestService", func() {
			dsl.Webhook("test.event", func() {
				dsl.WebhookDescription("Test webhook")
				dsl.WebhookPayload(func() {
					dsl.Attribute("id", expr.String)
					dsl.Attribute("name", expr.String)
					dsl.Required("id")
				})
				dsl.WebhookHTTP(func() {
					dsl.WebhookPOST("/webhook/test")
					dsl.WebhookHeaders(func() {
						dsl.Attribute("X-Signature", expr.String, "Signature")
						dsl.Required("X-Signature")
					})
					dsl.WebhookResponse(200, func() {
						dsl.Description("Success")
					})
				})
			})
			
			// Add a simple method to make the service valid
			dsl.Method("test", func() {
				dsl.HTTP(func() {
					dsl.GET("/test")
				})
			})
		})
	})
	
	// Generate OpenAPI specification
	spec := New(root)
	if spec == nil {
		t.Fatal("failed to generate OpenAPI spec: got nil")
	}
	
	// Check that webhooks were generated
	if spec.Webhooks == nil {
		t.Fatal("expected webhooks in OpenAPI spec, got nil")
	}
	
	if len(spec.Webhooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(spec.Webhooks))
	}
	
	// Check webhook exists with correct name
	webhook, ok := spec.Webhooks["test.event"]
	if !ok {
		t.Fatal("expected webhook 'test.event' not found")
	}
	
	// Check webhook has POST operation
	if webhook.Post == nil {
		t.Fatal("expected POST operation on webhook")
	}
	
	// Verify operation details
	if webhook.Post.Summary != "test.event" {
		t.Errorf("expected summary 'test.event', got %q", webhook.Post.Summary)
	}
	
	if webhook.Post.Description != "Test webhook" {
		t.Errorf("expected description 'Test webhook', got %q", webhook.Post.Description)
	}
	
	// Check request body exists
	if webhook.Post.RequestBody == nil {
		t.Fatal("expected request body on webhook POST operation")
	}
	
	// Check parameters (headers)
	if len(webhook.Post.Parameters) != 1 {
		t.Fatalf("expected 1 parameter (header), got %d", len(webhook.Post.Parameters))
	}
	
	param := webhook.Post.Parameters[0].Value
	if param.Name != "X-Signature" {
		t.Errorf("expected header parameter 'X-Signature', got %q", param.Name)
	}
	
	if param.In != "header" {
		t.Errorf("expected parameter in 'header', got %q", param.In)
	}
	
	// Check responses
	if len(webhook.Post.Responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(webhook.Post.Responses))
	}
	
	_, ok = webhook.Post.Responses["200"]
	if !ok {
		t.Error("expected response with status code 200")
	}
}

// runWebhookDSL helper for testing
func runWebhookDSL(t *testing.T, dslFunc func()) *expr.RootExpr {
	t.Helper()
	eval.Reset()
	expr.Root = new(expr.RootExpr)
	expr.GeneratedResultTypes = new(expr.ResultTypesRoot)
	
	if err := eval.Register(expr.Root); err != nil {
		t.Fatalf("failed to register root: %v", err)
	}
	if err := eval.Register(expr.GeneratedResultTypes); err != nil {
		t.Fatalf("failed to register result types: %v", err)
	}
	
	expr.Root.API = expr.NewAPIExpr("test", func() {})
	expr.Root.API.HTTP = new(expr.HTTPExpr)
	expr.Root.API.Servers = []*expr.ServerExpr{expr.Root.API.DefaultServer()}
	
	if !eval.Execute(dslFunc, nil) {
		t.Fatalf("DSL execution failed: %s", eval.Context.Error())
	}
	
	if err := eval.RunDSL(); err != nil {
		t.Fatalf("DSL run failed: %v", err)
	}
	
	return expr.Root
}