package eval_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestTooManyArgError(t *testing.T) {
	dsls := map[string]func(){
		"ArrayOf":          func() { ArrayOf(String, func() {}, func() {}) },
		"Attribute":        func() { Type("name", func() { Attribute("name", 1, 2, 3, 4) }) },
		"Example":          func() { Example(1, 2, 3) },
		"Files":            func() { Files("path", "filename", func() {}, func() {}) },
		"MapOf":            func() { MapOf(String, String, func() {}, func() {}) },
		"MapParams":        func() { MapParams(1, 2) },
		"Payload":          func() { Payload(String, 1, 2, 3) },
		"Response":         func() { API("name", func() { HTTP(func() { Response(StatusOK, "name", 1, 2) }) }) },
		"Result":           func() { Result(String, 1, 2, 3) },
		"ResultType":       func() { ResultType("identifier", "name", func() {}, func() {}) },
		"Scope":            func() { BasicAuthSecurity("name", func() { Scope("name", "1", "2") }) },
		"Server":           func() { Server("name", func() {}, func() {}) },
		"StreamingPayload": func() { StreamingPayload(String, 1, 2, 3) },
		"StreamingResult":  func() { StreamingResult(String, 1, 2, 3) },
		"Type":             func() { Type("name", 1, 2, 3) },
	}
	for name, dsl := range dsls {
		t.Run(name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, dsl)
			assert.Len(t, strings.Split(err.Error(), "\n"), 1)
			assert.Contains(t, err.Error(), "too many arguments given to "+name)
		})
	}
}
