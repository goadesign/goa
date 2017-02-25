package codegen

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/rest/design"
)

// EncodingWriter returns the HTTP transport encoding writer.
func EncodingWriter(r *design.RootExpr) codegen.FileWriter {
	return nil
}
