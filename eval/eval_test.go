package eval_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestTooManyArgError(t *testing.T) {
	cases := map[string]struct {
		DSL   func()
		Error string
	}{
		"ArrayOf":          {func() { ArrayOf(String, func() {}, func() {}) }, "too many arguments given to ArrayOf"},
		"Attribute":        {func() { Type("name", func() { Attribute("name", 1, 2, 3, 4) }) }, "too many arguments given to Attribute"},
		"Example":          {func() { Example(1, 2, 3) }, "too many arguments given to Example"},
		"Files":            {func() { Files("path", "filename", func() {}, func() {}) }, "too many arguments given to Files"},
		"MapOf":            {func() { MapOf(String, String, func() {}, func() {}) }, "too many arguments given to MapOf"},
		"MapParams":        {func() { MapParams(1, 2) }, "too many arguments given to MapParams"},
		"Payload":          {func() { Payload(String, 1, 2, 3) }, "too many arguments given to Payload"},
		"Response":         {func() { API("name", func() { HTTP(func() { Response(StatusOK, "name", 1, 2) }) }) }, "too many arguments given to Response"},
		"Result":           {func() { Result(String, 1, 2, 3) }, "too many arguments given to Result"},
		"ResultType":       {func() { ResultType("identifier", "name", func() {}, func() {}) }, "too many arguments given to ResultType"},
		"Scope":            {func() { BasicAuthSecurity("name", func() { Scope("name", "1", "2") }) }, "too many arguments given to Scope"},
		"Server":           {func() { Server("name", func() {}, func() {}) }, "too many arguments given to Server"},
		"StreamingPayload": {func() { StreamingPayload(String, 1, 2, 3) }, "too many arguments given to StreamingPayload"},
		"StreamingResult":  {func() { StreamingResult(String, 1, 2, 3) }, "too many arguments given to StreamingResult"},
		"Type":             {func() { Type("name", 1, 2, 3) }, "too many arguments given to Type"},
	}
	for _, tc := range cases {
		err := expr.RunInvalidDSL(t, tc.DSL)
		assert.Len(t, strings.Split(err.Error(), "\n"), 1)
		assert.Contains(t, err.Error(), tc.Error)
	}
}
