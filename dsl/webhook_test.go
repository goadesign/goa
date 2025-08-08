package dsl

import (
	"errors"
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestWebhook(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Expected *expr.WebhookExpr
	}{
		{
			Name: "basic webhook",
			DSL: func() {
				Service("test", func() {
					Webhook("user.created", func() {
						WebhookDescription("User created webhook")
						WebhookPayload(String)
					})
				})
			},
			Expected: &expr.WebhookExpr{
				Name:        "user.created",
				Description: "User created webhook",
			},
		},
		{
			Name: "webhook with HTTP configuration",
			DSL: func() {
				Service("test", func() {
					Webhook("order.placed", func() {
						WebhookDescription("Order placed webhook")
						WebhookPayload(func() {
							Attribute("order_id", String)
							Attribute("total", Float64)
							Required("order_id", "total")
						})
						WebhookHTTP(func() {
							WebhookPOST("/webhooks/orders")
							WebhookHeaders(func() {
								Attribute("X-Signature", String, "HMAC signature")
								Required("X-Signature")
							})
							WebhookResponse(StatusOK, func() {
								Description("Success")
							})
							WebhookResponse(StatusBadRequest, func() {
								Description("Bad request")
							})
						})
					})
				})
			},
			Expected: &expr.WebhookExpr{
				Name:        "order.placed",
				Description: "Order placed webhook",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Run DSL with proper setup
			root, err := runWebhookDSL(t, c.DSL)
			if err != nil {
				t.Fatalf("unexpected error running DSL: %v", err)
			}

			// Check service was created
			if len(root.Services) == 0 {
				t.Fatal("no services created")
			}

			svc := root.Services[0]
			t.Logf("Service: %s, Webhooks count: %d", svc.Name, len(svc.Webhooks))
			
			// Check webhook was created
			if len(svc.Webhooks) == 0 {
				t.Fatal("no webhooks created")
			}

			webhook := svc.Webhooks[0]
			
			// Verify basic properties
			if webhook.Name != c.Expected.Name {
				t.Errorf("expected webhook name %q, got %q", c.Expected.Name, webhook.Name)
			}
			
			if webhook.Description != c.Expected.Description {
				t.Errorf("expected webhook description %q, got %q", c.Expected.Description, webhook.Description)
			}

			// For HTTP webhook test, verify HTTP configuration
			if c.Name == "webhook with HTTP configuration" {
				if webhook.HTTP == nil {
					t.Fatal("expected HTTP configuration, got nil")
				}
				
				if webhook.HTTP.Method != "POST" {
					t.Errorf("expected method POST, got %q", webhook.HTTP.Method)
				}
				
				if webhook.HTTP.Path != "/webhooks/orders" {
					t.Errorf("expected path /webhooks/orders, got %q", webhook.HTTP.Path)
				}
				
				if webhook.HTTP.Headers == nil {
					t.Fatal("expected headers, got nil")
				}
				
				if len(webhook.HTTP.Responses) != 2 {
					t.Errorf("expected 2 responses, got %d", len(webhook.HTTP.Responses))
				}
			}
		})
	}
}

// runWebhookDSL returns the DSL root resulting from running the given DSL.
func runWebhookDSL(t *testing.T, dsl func()) (*expr.RootExpr, error) {
	t.Helper()
	eval.Reset()
	expr.Root = new(expr.RootExpr)
	expr.GeneratedResultTypes = new(expr.ResultTypesRoot)
	
	if err := eval.Register(expr.Root); err != nil {
		return nil, err
	}
	if err := eval.Register(expr.GeneratedResultTypes); err != nil {
		return nil, err
	}
	
	expr.Root.API = expr.NewAPIExpr("test api", func() {})
	expr.Root.API.Servers = []*expr.ServerExpr{expr.Root.API.DefaultServer()}
	expr.Root.API.HTTP = new(expr.HTTPExpr)
	
	if eval.Execute(dsl, nil) {
		return expr.Root, eval.RunDSL()
	} else {
		return expr.Root, errors.New(eval.Context.Error())
	}
}