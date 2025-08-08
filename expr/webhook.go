package expr

import (
	"goa.design/goa/v3/eval"
)

type (
	// WebhookExpr describes a webhook endpoint that sends HTTP requests to
	// external services when specific events occur. Webhooks are used in
	// OpenAPI 3.1 to document outgoing HTTP requests.
	WebhookExpr struct {
		// Name is the webhook identifier
		Name string
		// Description provides details about when the webhook is triggered
		Description string
		// Payload describes the data sent in the webhook request
		Payload *AttributeExpr
		// HTTP contains HTTP-specific webhook configuration
		HTTP *HTTPWebhookExpr
		// Meta contains metadata
		Meta MetaExpr
	}

	// HTTPWebhookExpr describes HTTP-specific properties of a webhook
	HTTPWebhookExpr struct {
		// Method is the HTTP method (POST, PUT, etc.)
		Method string
		// Path is the URL path for the webhook
		Path string
		// Headers describes HTTP headers sent with the webhook
		Headers *AttributeExpr
		// Body describes the HTTP body
		Body *AttributeExpr
		// Responses describes expected responses indexed by status code
		Responses []*HTTPResponseExpr
		// Parent is the parent WebhookExpr
		Parent *WebhookExpr
	}
)

// EvalName returns the name of the webhook for logging.
func (w *WebhookExpr) EvalName() string {
	if w.Name != "" {
		return "webhook " + w.Name
	}
	return "unnamed webhook"
}

// Validate ensures the webhook expression is valid.
func (w *WebhookExpr) Validate(ctx string, parent eval.Expression) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	
	if w.Name == "" {
		verr.Add(parent, "%s: webhook name is required", ctx)
	}
	
	if w.Payload != nil {
		verr.Merge(w.Payload.Validate(ctx+".Payload", w))
	}
	
	if w.HTTP != nil {
		verr.Merge(w.HTTP.Validate(ctx+".HTTP", w))
	}
	
	return verr
}

// Finalize finalizes the webhook expression.
func (w *WebhookExpr) Finalize() {
	if w.Payload != nil {
		w.Payload.Finalize()
	}
	
	if w.HTTP != nil {
		w.HTTP.Finalize()
	}
}

// Prepare prepares the webhook expression.
func (w *WebhookExpr) Prepare() {
	if w.HTTP != nil {
		w.HTTP.Prepare()
	}
}

// Validate ensures the HTTP webhook expression is valid.
func (h *HTTPWebhookExpr) Validate(ctx string, parent eval.Expression) *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	
	if h.Path == "" {
		verr.Add(parent, "%s: webhook path is required", ctx)
	}
	
	if h.Headers != nil {
		verr.Merge(h.Headers.Validate(ctx+".Headers", h))
	}
	
	if h.Body != nil {
		verr.Merge(h.Body.Validate(ctx+".Body", h))
	}
	
	// Validate responses - simplified validation for webhooks
	for i, r := range h.Responses {
		if r.StatusCode < 100 || r.StatusCode > 599 {
			verr.Add(parent, "%s.Responses[%d]: invalid status code %d", ctx, i, r.StatusCode)
		}
	}
	
	return verr
}

// Finalize finalizes the HTTP webhook expression.
func (h *HTTPWebhookExpr) Finalize() {
	if h.Headers != nil {
		h.Headers.Finalize()
	}
	
	if h.Body != nil {
		h.Body.Finalize()
	}
	
	// Finalize responses
	for _, r := range h.Responses {
		if r.Body != nil {
			r.Body.Finalize()
		}
		if r.Headers != nil {
			r.Headers.Finalize()
		}
	}
}

// Prepare prepares the HTTP webhook expression.
func (h *HTTPWebhookExpr) Prepare() {
	// Set default method if not specified
	if h.Method == "" {
		h.Method = "POST"
	}
	
	// Set default path if not specified
	if h.Path == "" {
		h.Path = "/webhook"
	}
}

// EvalName returns the name for logging.
func (h *HTTPWebhookExpr) EvalName() string {
	if h.Parent != nil {
		return h.Parent.EvalName() + " HTTP"
	}
	return "HTTP webhook"
}