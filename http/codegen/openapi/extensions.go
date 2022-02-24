package openapi

import (
	"encoding/json"
	"strings"

	"goa.design/goa/v3/expr"
)

// ExtensionsFromExpr generates openapi extensions from the given meta
// expression.
func ExtensionsFromExpr(mdata expr.MetaExpr) map[string]interface{} {
	swag := extensionsFromExprWithPrefix(mdata, "swagger:extension:")
	open := extensionsFromExprWithPrefix(mdata, "openapi:extension:")
	if swag == nil {
		return open
	}
	if open == nil {
		return swag
	}
	for k, v := range open {
		swag[k] = v
	}
	return swag
}

// extensionsFromExprWithPrefix generates openapi extensions from
// the given meta expression with keys starting the given prefix.
func extensionsFromExprWithPrefix(mdata expr.MetaExpr, prefix string) map[string]interface{} {
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	extensions := make(map[string]interface{})
	for key, value := range mdata {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		name := key[len(prefix):]
		if strings.Contains(name, ":") {
			continue
		}
		if !strings.HasPrefix(name, "x-") {
			continue
		}
		val := value[0]
		ival := interface{}(val)
		if err := json.Unmarshal([]byte(val), &ival); err != nil {
			extensions[name] = val
			continue
		}
		extensions[name] = ival
	}
	if len(extensions) == 0 {
		return nil
	}
	return extensions
}
