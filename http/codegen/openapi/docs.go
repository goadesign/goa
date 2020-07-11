package openapi

import "goa.design/goa/v3/expr"

// ExternalDocs represents an OpenAPI External Documentation object as defined in
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#externalDocumentationObject
type ExternalDocs struct {
	Description string                 `json:"description,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Extensions  map[string]interface{} `json:"-" yaml:"-"`
}

// DocsFromExpr builds a ExternalDocs from the Goa docs expression.
func DocsFromExpr(docs *expr.DocsExpr, meta expr.MetaExpr) *ExternalDocs {
	if docs == nil {
		return nil
	}
	return &ExternalDocs{
		Description: docs.Description,
		URL:         docs.URL,
		Extensions:  ExtensionsFromExpr(meta),
	}
}
