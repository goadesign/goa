package openapi

import (
	"encoding/json"
	"maps"
	"strings"

	"goa.design/goa/v3/expr"
)

// ExtensionsFromExpr generates openapi extensions from the given meta
// expression.
func ExtensionsFromExpr(mdata expr.MetaExpr) map[string]any {
	swag := extensionsFromExprWithPrefix(mdata, "swagger:extension:")
	open := extensionsFromExprWithPrefix(mdata, "openapi:extension:")
	if swag == nil {
		return open
	}
	if open == nil {
		return swag
	}
	maps.Copy(swag, open)
	return swag
}

// ExtensionsFromMethod returns the OpenAPI extensions authored as method
// metadata and advertises the method's idempotency contract when present.
func ExtensionsFromMethod(method *expr.MethodExpr) map[string]any {
	extensions := ExtensionsFromExpr(method.Meta)
	if !method.Idempotent {
		return extensions
	}
	if extensions == nil {
		extensions = make(map[string]any)
	}
	extensions["x-goa-idempotent"] = true
	return extensions
}

// extensionsFromExprWithPrefix generates openapi extensions from
// the given meta expression with keys starting the given prefix.
func extensionsFromExprWithPrefix(mdata expr.MetaExpr, prefix string) map[string]any {
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	extensions := make(map[string]any)
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
		ival := any(val)
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
