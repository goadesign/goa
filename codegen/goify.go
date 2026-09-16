package codegen

import (
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/internal/codegenname"
)

// Goify makes a valid Go identifier out of any string. It does that by removing
// any non letter and non digit character and by making sure the first character
// is a letter or "_". Goify produces a "CamelCase" version of the string, if
// firstUpper is true the first character of the identifier is uppercase
// otherwise it's lowercase.
func Goify(str string, firstUpper bool) string {
	key := cacheKey{
		input:      str,
		firstUpper: firstUpper,
		operation:  "goify",
	}
	return globalStringCache.getCached(key, func() string {
		return codegenname.Goify(str, firstUpper)
	})
}

// GoifyAtt honors any struct:field:name meta set on the attribute and calls
// Goify with the tag value if present or the given name otherwise.
func GoifyAtt(att *expr.AttributeExpr, name string, upper bool) string {
	name = codegenname.AttributeName(name, att.Meta["struct:field:name"])
	return Goify(name, upper)
}

// fixReservedGo appends an underscore on to Go reserved keywords.
func fixReservedGo(w string) string {
	return codegenname.FixReserved(w)
}
