package design

import (
	"fmt"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

type (
	// HTTPResponseExpr defines a HTTP response including its status, headers and media type.
	HTTPResponseExpr struct {
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
		// Response view name if MediaType is the id of a MediaTypeExpr
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
)

// EvalName returns the generic definition name used in error messages.
func (r *HTTPResponseExpr) EvalName() string {
	var prefix, suffix string
	if r.Name != "" {
		prefix = fmt.Sprintf("HTTP response %#v", r.Name)
	} else {
		prefix = "unnamed HTTP response"
	}
	if r.Parent != nil {
		suffix = fmt.Sprintf(" of %s", r.Parent.EvalName())
	}
	return prefix + suffix
}

// Validate checks that the response definition is consistent: its status is set and the media
// type definition if any is valid.
func (r *HTTPResponseExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if r.Headers != nil {
		verr.Merge(r.Headers.Validate("HTTP response headers", r))
	}
	if r.Status == 0 {
		verr.Add(r, "HTTP response status not defined")
	}
	return verr
}

// Finalize sets the response media type from its type if the type is a media type and no media
// type is already specified.
func (r *HTTPResponseExpr) Finalize() {
	if r.Type == nil {
		return
	}
	if r.MediaType != "" && r.MediaType != "text/plain" {
		return
	}
	mt, ok := r.Type.(*design.MediaTypeExpr)
	if !ok {
		return
	}
	r.MediaType = mt.Identifier
}

// Dup returns a copy of the response definition.
func (r *HTTPResponseExpr) Dup() *HTTPResponseExpr {
	res := HTTPResponseExpr{
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
func (r *HTTPResponseExpr) Merge(other *HTTPResponseExpr) {
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
