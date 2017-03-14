package writers

import (
	"goa.design/goa.v2/codegen"
	rest "goa.design/goa.v2/rest/design"
)

// OpenAPIWriter returns the codegen.FileWriter for the OpenAPI spec of the given
// HTTP API.
func OpenAPIWriter(root *rest.RootExpr) codegen.FileWriter {
	return nil
}
