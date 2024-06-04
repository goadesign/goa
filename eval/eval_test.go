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
		"APIKey":           func() { Type("name", func() { APIKey("scheme", "name", 1, 2, 3) }) },
		"APIKeyField":      func() { Type("name", func() { APIKeyField("tag", "scheme", "name", 1, 2, 3) }) },
		"AccessToken":      func() { Type("name", func() { AccessToken("name", 1, 2, 3) }) },
		"AccessTokenField": func() { Type("name", func() { AccessTokenField("tag", "name", 1, 2, 3) }) },
		"ArrayOf":          func() { ArrayOf(String, func() {}, func() {}) },
		"Attribute":        func() { Type("name", func() { Attribute("name", 1, 2, 3, 4) }) },
		"Cookie":           func() { API("name", func() { HTTP(func() { Cookie("name", 1, 2, 3, 4) }) }) },
		"Error":            func() { API("name", func() { Error("name", 1, 2, 3, 4) }) },
		"ErrorName":        func() { Type("name", func() { ErrorName("name", 1, 2, 3) }) },
		"Example":          func() { Example(1, 2, 3) },
		"Field":            func() { Type("name", func() { Field("tag", "name", 1, 2, 3, 4) }) },
		"Files":            func() { Files("path", "filename", func() {}, func() {}) },
		"HTTP":             func() { API("name", func() { HTTP(func() {}, func() {}) }) },
		"Header":           func() { API("name", func() { HTTP(func() { Header("name", 1, 2, 3, 4) }) }) },
		"MapOf":            func() { MapOf(String, String, func() {}, func() {}) },
		"MapParams":        func() { MapParams(1, 2) },
		"OneOf":            func() { OneOf("name", 1, 2, 3) },
		"Param":            func() { API("name", func() { HTTP(func() { Param("name", 1, 2, 3, 4) }) }) },
		"Password":         func() { Type("name", func() { Password("name", 1, 2, 3) }) },
		"PasswordField":    func() { Type("name", func() { PasswordField("tag", "name", 1, 2, 3) }) },
		"Payload":          func() { Payload(String, 1, 2, 3) },
		"Response (int)":   func() { API("name", func() { HTTP(func() { Response(StatusOK, "name", 1, 2) }) }) },
		"Response (func)":  func() { API("name", func() { HTTP(func() { Response("name", func() {}, func() {}) }) }) },
		"Result":           func() { Result(String, 1, 2, 3) },
		"ResultType":       func() { ResultType("identifier", "name", func() {}, func() {}) },
		"Scope":            func() { BasicAuthSecurity("name", func() { Scope("name", "1", "2") }) },
		"Server":           func() { Server("name", func() {}, func() {}) },
		"StreamingPayload": func() { StreamingPayload(String, 1, 2, 3) },
		"StreamingResult":  func() { StreamingResult(String, 1, 2, 3) },
		"Token":            func() { Type("name", func() { Token("name", 1, 2, 3) }) },
		"TokenField":       func() { Type("name", func() { TokenField("tag", "name", 1, 2, 3) }) },
		"Type":             func() { Type("name", 1, 2, 3) },
		"Type (func)":      func() { Type("name", func() {}, func() {}) },
		"Username":         func() { Type("name", func() { Username("name", 1, 2, 3) }) },
		"UsernameField":    func() { Type("name", func() { UsernameField("tag", "name", 1, 2, 3) }) },
		"Variable":         func() { API("a", func() { Server("s", func() { Host("h", func() { Variable("v", 1, 2, 3, 4) }) }) }) },
	}
	for name, dsl := range dsls {
		t.Run(name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, dsl)
			assert.Len(t, strings.Split(err.Error(), "\n"), 1)
			assert.Contains(t, err.Error(), "too many arguments given to "+strings.Split(name, " ")[0])
		})
	}
}
