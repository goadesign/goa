package design

import (
	"fmt"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

type (
	// ResponseExpr defines a HTTP response status and optional validation rules.
	ResponseExpr struct {
		// Response name
		Name string
		// HTTP status
		Status int
		// Response description
		Description string
		// Response body type if any
		Type design.DataType
		// Response body media type if any
		MediaType string
		// Response view name if MediaType is MediaTypeExpr
		ViewName string
		// Response header expressions
		Headers *design.AttributeExpr
		// Parent action or resource
		Parent eval.Expression
		// Metadata is a list of key/value pairs
		Metadata design.MetadataExpr
		// Standard is true if the response is one of the default responses.
		Standard bool
	}

	// ResponseTemplateExpr defines a response template.
	// A response template is a function that takes an arbitrary number
	// of strings and returns a response expression.
	ResponseTemplateExpr struct {
		// Response template name
		Name string
		// Response template function
		Template func(...string) *ResponseExpr
	}

	// ResponseIterator is the type of functions given to IterateResponses.
	ResponseIterator func(*ResponseExpr) error
)

// EvalName returns the generic definition name used in error messages.
func (r *ResponseExpr) EvalName() string {
	var prefix, suffix string
	if r.Name != "" {
		prefix = fmt.Sprintf("response %#v", r.Name)
	} else {
		prefix = "unnamed response"
	}
	if r.Parent != nil {
		suffix = fmt.Sprintf(" of %s", r.Parent.EvalName())
	}
	return prefix + suffix
}

// Validate checks that the response definition is consistent: its status is set and the media
// type definition if any is valid.
func (r *ResponseExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if r.Headers != nil {
		verr.Merge(r.Headers.Validate("response headers", r))
	}
	if r.Status == 0 {
		verr.Add(r, "response status not defined")
	}
	return verr
}

// Finalize sets the response media type from its type if the type is a media type and no media
// type is already specified.
func (r *ResponseExpr) Finalize() {
	if r.Type == nil {
		return
	}
	if r.MediaType != "" && r.MediaType != "text/plain" {
		return
	}
	mt, ok := r.Type.(*MediaTypeExpr)
	if !ok {
		return
	}
	r.MediaType = mt.Identifier
}

// Dup returns a copy of the response definition.
func (r *ResponseExpr) Dup() *ResponseExpr {
	res := ResponseExpr{
		Name:        r.Name,
		Status:      r.Status,
		Description: r.Description,
		MediaType:   r.MediaType,
		ViewName:    r.ViewName,
	}
	if r.Headers != nil {
		res.Headers = design.DupAtt(r.Headers)
	}
	return &res
}

// Merge merges other into target. Only the fields of target that are not already set are merged.
func (r *ResponseExpr) Merge(other *ResponseExpr) {
	if other == nil {
		return
	}
	if r.Name == "" {
		r.Name = other.Name
	}
	if r.Status == 0 {
		r.Status = other.Status
	}
	if r.Description == "" {
		r.Description = other.Description
	}
	if r.MediaType == "" {
		r.MediaType = other.MediaType
		r.ViewName = other.ViewName
	}
	if other.Headers != nil {
		otherHeaders := other.Headers.Type.(design.Object)
		if len(otherHeaders) > 0 {
			if r.Headers == nil {
				r.Headers = &design.AttributeExpr{Type: design.Object{}}
			}
			headers := r.Headers.Type.(design.Object)
			for n, h := range otherHeaders {
				if _, ok := headers[n]; !ok {
					headers[n] = h
				}
			}
		}
	}
}

// EvalName returns the generic definition name used in error messages.
func (r *ResponseTemplateExpr) EvalName() string {
	if r.Name != "" {
		return fmt.Sprintf("response template %#v", r.Name)
	}
	return "unnamed response template"
}
