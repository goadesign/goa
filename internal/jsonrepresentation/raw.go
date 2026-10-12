// Package jsonrepresentation recognizes explicit JSON representations selected
// by authored type metadata. Generators use these facts before rendering code.
package jsonrepresentation

import (
	"go/ast"
	"go/parser"

	"goa.design/goa/v3/expr"
)

// IsRawJSON reports whether bytes explicitly use encoding/json.RawMessage.
// Named types inherit this representation; another custom representation stops
// that inheritance. Ordinary byte fields retain their base64 JSON encoding.
func IsRawJSON(attribute *expr.AttributeExpr) bool {
	if args, ok := attribute.Meta["struct:field:type"]; ok {
		if len(args) < 2 || args[1] != "encoding/json" || !primitiveBytes(attribute.Type) {
			return false
		}
		declared, err := parser.ParseExpr(args[0])
		if err != nil {
			return false
		}
		selector, ok := declared.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "RawMessage" {
			return false
		}
		qualifier, ok := selector.X.(*ast.Ident)
		alias := "json"
		if len(args) >= 3 && args[2] != "" {
			alias = args[2]
		}
		return ok && qualifier.Name == alias
	}
	if named, ok := attribute.Type.(expr.UserType); ok {
		return IsRawJSON(named.Attribute())
	}
	return false
}

func primitiveBytes(value expr.DataType) bool {
	if named, ok := value.(expr.UserType); ok {
		return primitiveBytes(named.Attribute().Type)
	}
	return value == expr.Bytes
}
