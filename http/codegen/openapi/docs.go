// This file converts Goa documentation links into OpenAPI values while
// allowing one specification build to replace their displayed description.
package openapi

import "goa.design/goa/v3/expr"

// ExternalDocs represents an OpenAPI External Documentation object as defined in
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#externalDocumentationObject
type ExternalDocs struct {
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url,omitempty"`
	Extensions  map[string]any `json:"-" yaml:"-"`
}

// DocsFromExpr builds a ExternalDocs from the Goa docs expression.
func DocsFromExpr(docs *expr.DocsExpr, meta expr.MetaExpr) *ExternalDocs {
	return docsFromExpr(docs, meta, Values{})
}

// DocsFromExprWithValues builds ExternalDocs and uses values for its
// description when one is present for docs.
func DocsFromExprWithValues(docs *expr.DocsExpr, meta expr.MetaExpr, values Values) *ExternalDocs {
	return docsFromExpr(docs, meta, values)
}

// docsFromExpr is the one implementation used by ordinary and customized
// OpenAPI builds.
func docsFromExpr(docs *expr.DocsExpr, meta expr.MetaExpr, values Values) *ExternalDocs {
	if docs == nil {
		return nil
	}
	return &ExternalDocs{
		Description: values.Description(docs, docs.Description),
		URL:         docs.URL,
		Extensions:  ExtensionsFromExpr(meta),
	}
}
