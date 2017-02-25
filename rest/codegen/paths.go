package codegen

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/rest/design"
)

// PathWriter returns the HTTP endpoint path generators writer.
func PathWriter(r *design.RootExpr) codegen.FileWriter {
	return nil
}
