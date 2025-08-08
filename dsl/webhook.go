package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Webhook defines a webhook endpoint that sends HTTP requests to external
// services when specific events occur. This is used to generate OpenAPI 3.1
// webhook documentation.
//
// Webhook must appear in a Service expression.
//
// Webhook accepts two arguments: the name of the webhook and a DSL function
// that defines the webhook's payload and other properties.
//
// Example:
//
//	var _ = Service("notifications", func() {
//	    Webhook("user.created", func() {
//	        Description("Triggered when a new user is created")
//	        Payload(User)
//	        HTTP(func() {
//	            POST("/webhook")
//	            Response(StatusOK)
//	        })
//	    })
//	})
func Webhook(name string, fn ...func()) {
	if len(fn) > 1 {
		eval.TooManyArgError()
		return
	}

	// Create webhook expression
	webhook := &expr.WebhookExpr{
		Name: name,
	}

	// Execute DSL if provided
	if len(fn) > 0 {
		if !eval.Execute(fn[0], webhook) {
			return
		}
	}

	// Add to service
	switch e := eval.Current().(type) {
	case *expr.ServiceExpr:
		e.Webhooks = append(e.Webhooks, webhook)
	default:
		// Try to get from eval.Current() context
		if svc := serviceFor(eval.Current()); svc != nil {
			svc.Webhooks = append(svc.Webhooks, webhook)
		} else {
			eval.IncompatibleDSL()
		}
	}
}

// serviceFor returns the service expression for the given eval context
func serviceFor(e eval.Expression) *expr.ServiceExpr {
	switch actual := e.(type) {
	case *expr.ServiceExpr:
		return actual
	case *expr.MethodExpr:
		return actual.Service
	case *expr.HTTPServiceExpr:
		return actual.ServiceExpr
	default:
		return nil
	}
}

// WebhookPayload defines the payload that will be sent to the webhook endpoint.
// It behaves similarly to the Payload DSL but is specifically for webhook data.
//
// WebhookPayload must appear in a Webhook expression.
//
// Example:
//
//	Webhook("user.created", func() {
//	    WebhookPayload(func() {
//	        Attribute("id", String, "User ID")
//	        Attribute("email", String, "User email", func() {
//	            Format(FormatEmail)
//	        })
//	        Attribute("created_at", String, "Creation timestamp", func() {
//	            Format(FormatDateTime)
//	        })
//	        Required("id", "email", "created_at")
//	    })
//	})
func WebhookPayload(p any) {
	if w, ok := eval.Current().(*expr.WebhookExpr); ok {
		switch actual := p.(type) {
		case func():
			attr := &expr.AttributeExpr{Type: &expr.Object{}}
			if !eval.Execute(actual, attr) {
				return
			}
			w.Payload = attr
		case expr.UserType:
			// UserType implements DataType
			w.Payload = &expr.AttributeExpr{Type: actual}
		case expr.DataType:
			// Any other DataType
			w.Payload = &expr.AttributeExpr{Type: actual}
		default:
			eval.InvalidArgError("function, type or DataType", p)
		}
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookDescription sets the description for a webhook.
//
// WebhookDescription must appear in a Webhook expression.
//
// Example:
//
//	Webhook("user.created", func() {
//	    WebhookDescription("This webhook is triggered whenever a new user account is created in the system")
//	})
func WebhookDescription(d string) {
	if w, ok := eval.Current().(*expr.WebhookExpr); ok {
		w.Description = d
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookHTTP defines HTTP-specific properties for the webhook.
//
// WebhookHTTP must appear in a Webhook expression.
//
// Example:
//
//	Webhook("user.created", func() {
//	    WebhookPayload(User)
//	    WebhookHTTP(func() {
//	        POST("/webhooks/user-created")
//	        Headers(func() {
//	            Header("X-Webhook-Event", String, "Event type")
//	            Header("X-Webhook-Signature", String, "HMAC signature")
//	        })
//	        Response(StatusOK)
//	        Response(StatusBadRequest)
//	    })
//	})
func WebhookHTTP(fn func()) {
	if w, ok := eval.Current().(*expr.WebhookExpr); ok {
		if w.HTTP == nil {
			w.HTTP = &expr.HTTPWebhookExpr{
				Parent: w,
			}
		}
		if !eval.Execute(fn, w.HTTP) {
			return
		}
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookMethod sets the HTTP method for the webhook.
// This is a helper function that can be used inside WebhookHTTP.
//
// Example:
//
//	WebhookHTTP(func() {
//	    WebhookMethod("POST")
//	    WebhookPath("/webhook")
//	})
func WebhookMethod(method string) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		h.Method = method
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookPath sets the URL path for the webhook.
// This is a helper function that can be used inside WebhookHTTP.
//
// Example:
//
//	WebhookHTTP(func() {
//	    WebhookPath("/webhooks/events")
//	})
func WebhookPath(path string) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		h.Path = path
	} else {
		eval.IncompatibleDSL()
	}
}

// Convenience methods for setting webhook HTTP method and path.
// These can be used inside WebhookHTTP as shortcuts.

// WebhookPOST sets the webhook to use POST method with the given path.
func WebhookPOST(path string) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		h.Method = "POST"
		h.Path = path
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookGET sets the webhook to use GET method with the given path.
func WebhookGET(path string) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		h.Method = "GET"
		h.Path = path
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookPUT sets the webhook to use PUT method with the given path.
func WebhookPUT(path string) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		h.Method = "PUT"
		h.Path = path
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookDELETE sets the webhook to use DELETE method with the given path.
func WebhookDELETE(path string) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		h.Method = "DELETE"
		h.Path = path
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookPATCH sets the webhook to use PATCH method with the given path.
func WebhookPATCH(path string) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		h.Method = "PATCH"
		h.Path = path
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookHeaders defines the HTTP headers sent with the webhook.
//
// WebhookHeaders must appear in a WebhookHTTP expression.
//
// Example:
//
//	WebhookHTTP(func() {
//	    WebhookHeaders(func() {
//	        Attribute("X-Event-Type", String, "Type of event")
//	        Attribute("X-Signature", String, "HMAC signature")
//	        Required("X-Event-Type")
//	    })
//	})
func WebhookHeaders(fn func()) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		attr := &expr.AttributeExpr{Type: &expr.Object{}}
		if !eval.Execute(fn, attr) {
			return
		}
		h.Headers = attr
	} else {
		eval.IncompatibleDSL()
	}
}

// WebhookResponse defines an expected response from the webhook consumer.
//
// WebhookResponse must appear in a WebhookHTTP expression.
//
// Example:
//
//	WebhookHTTP(func() {
//	    WebhookResponse(StatusOK, func() {
//	        Description("Webhook processed successfully")
//	    })
//	    WebhookResponse(StatusBadRequest, func() {
//	        Description("Invalid webhook payload")
//	    })
//	})
func WebhookResponse(status int, fn ...func()) {
	if h, ok := eval.Current().(*expr.HTTPWebhookExpr); ok {
		response := &expr.HTTPResponseExpr{
			StatusCode: status,
		}
		if len(fn) > 0 {
			if !eval.Execute(fn[0], response) {
				return
			}
		}
		if h.Responses == nil {
			h.Responses = make([]*expr.HTTPResponseExpr, 0)
		}
		h.Responses = append(h.Responses, response)
	} else {
		eval.IncompatibleDSL()
	}
}