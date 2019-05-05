package expr

import (
	"goa.design/goa/v3/eval"
)

// findKey finds the given key in the endpoint expression and returns the
// transport element name and the position (header, query, or body for HTTP or
// message, metadata for gRPC endpoint).
func findKey(exp eval.Expression, keyAtt string) (string, string) {
	switch e := exp.(type) {
	case *HTTPEndpointExpr:
		if n, exists := e.Params.FindKey(keyAtt); exists {
			return n, "query"
		} else if n, exists := e.Headers.FindKey(keyAtt); exists {
			return n, "header"
		} else if e.Body == nil {
			return "", "header"
		}
		if _, ok := e.Body.Meta["http:body"]; ok {
			if e.Body.Find(keyAtt) != nil {
				return keyAtt, "body"
			}
			if m, ok := e.Body.Meta["origin:attribute"]; ok && m[0] == keyAtt {
				return keyAtt, "body"
			}
		}
		return "", "header"
	case *GRPCEndpointExpr:
		if e.Request.Find(keyAtt) != nil {
			return keyAtt, "message"
		} else if n, exists := e.Metadata.FindKey(keyAtt); exists {
			return n, "metadata"
		}
		return "", "metadata"
	}
	return "", ""
}
