package testdata

import . "goa.design/goa/v3/dsl"

var ValidErrorsDSL = func() {
	var Result = ResultType("application/vnd.goa.error", func() {
		TypeName("Result")
		Attributes(func() {
			Attribute("a", String)
			ErrorName("b")
			Required("b")
		})
	})
	var AType = Type("AType", func() {
		Attribute("a", String)
	})
	Service("ValidErrors", func() {
		Error("default_service_level")
		Error("custom_errors", String, "String error")
		Method("Method", func() {
			Error("error1", Result)
			Error("error2", Result)
			Error("custom_errors", AType) // override service error
		})
	})
}

var InvalidStructErrorNameDSL = func() {
	var Common = Type("Common", func() {
		ErrorName("a", Int) // invalid type for error name
		Required("a")
	})
	var Result = ResultType("application/vnd.goa.error", func() {
		TypeName("Error")
		Attributes(func() {
			Extend(Common)
			ErrorName("b") // invalid duplication of error name
			Required("b")
		})
	})
	var ServiceError = Type("ServiceError", func() {
		ErrorName("a")
		// invalid: error name "a" must be required
	})
	var ErrorType = Type("ErrorType", func() {
		Attribute("a", String)
	})
	Service("InvalidStructErrorName", func() {
		Error("service_error", ServiceError)
		Method("Method", func() {
			Error("struct_error_name", Result)
			Error("error1", ErrorType)
			Error("error2", ErrorType)
		})
	})
}

var ServiceErrorDSL = func() {
	var ServiceError = Type("ServiceError", func() {
		ErrorName("a")
		// invalid: error name "a" must be required
	})
	Service("InvalidStructErrorName", func() {
		Error("service_error", ServiceError)
		Method("Method", func() {})
	})
}
