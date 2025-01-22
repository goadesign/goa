package eval_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestInvalidArgError(t *testing.T) {
	dsls := map[string]struct {
		dsl  func()
		want string
	}{
		"Attribute":            {func() { Type("name", func() { Attribute("name", String, "description", 1) }) }, "cannot use 1 (type int) as type func()"},
		"Body":                 {func() { Service("s", func() { Method("m", func() { HTTP(func() { Body(1) }) }) }) }, "cannot use 1 (type int) as type attribute name, user type or DSL"},
		"ErrorName (bool)":     {func() { Type("name", func() { ErrorName(true) }) }, "cannot use true (type bool) as type name or position"},
		"ErrorName (int)":      {func() { Type("name", func() { ErrorName(1, 2) }) }, "cannot use 2 (type int) as type name"},
		"Example":              {func() { Example(1, 2) }, "cannot use 1 (type int) as type summary (string)"},
		"Headers":              {func() { Headers(1) }, "cannot use 1 (type int) as type function"},
		"MapParams":            {func() { Service("s", func() { Method("m", func() { HTTP(func() { MapParams(1) }) }) }) }, "cannot use 1 (type int) as type string"},
		"OneOf (function)":     {func() { Type("name", func() { OneOf("name", "description", 1) }) }, "cannot use 1 (type int) as type function"},
		"OneOf (string)":       {func() { Type("name", func() { OneOf("name", 1, func() {}) }) }, "cannot use 1 (type int) as type string"},
		"Param":                {func() { API("name", func() { HTTP(func() { Params(1) }) }) }, "cannot use 1 (type int) as type function"},
		"Response":             {func() { Service("s", func() { HTTP(func() { Response(1) }) }) }, "cannot use 1 (type int) as type name of error"},
		"ResultType":           {func() { ResultType("identifier", 1) }, "cannot use 1 (type int) as type function or string"},
		"Security":             {func() { Security(1) }, "cannot use 1 (type int) as type security scheme or security scheme name"},
		"Security (typed nil)": {func() { Security((*expr.SchemeExpr)(nil)) }, "cannot use (*expr.SchemeExpr)(nil) (type *expr.SchemeExpr) as type security scheme"},
		"Type":                 {func() { Type("name", 1) }, "cannot use 1 (type int) as type type or function"},
		"Type (DataType)":      {func() { Type("name", String, 1) }, "cannot use 1 (type int) as type function"},
	}
	for name, tc := range dsls {
		t.Run(name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, tc.dsl)
			assert.Len(t, strings.Split(err.Error(), "\n"), 1)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestTooFewArgError(t *testing.T) {
	dsls := map[string]func(){
		"Body":            func() { Body() },
		"ErrorName":       func() { ErrorName() },
		"ErrorName (int)": func() { Type("name", func() { ErrorName(1) }) },
		"Example":         func() { Example() },
		"OneOf":           func() { OneOf("name") },
		"Response (grpc)": func() { Service("s", func() { GRPC(func() { Response("name") }) }) },
		"Response (http)": func() { Service("s", func() { HTTP(func() { Response("name") }) }) },
	}
	for name, dsl := range dsls {
		t.Run(name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, dsl)
			assert.Len(t, strings.Split(err.Error(), "\n"), 1)
			assert.Contains(t, err.Error(), "too few arguments given to "+strings.Split(name, " ")[0])
		})
	}
}

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
