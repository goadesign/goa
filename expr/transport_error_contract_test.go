// This file verifies that reusable transport error mappings never replace the
// service error contract selected by an endpoint method.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestHTTPInheritedErrorMappingUsesMethodError(t *testing.T) {
	root := expr.RunDSL(t, equivalentHTTPErrorMappingDSL)
	endpoint := root.API.HTTP.Services[0].HTTPEndpoints[0]

	require.Same(t, endpoint.MethodExpr.Error("bad_request"), endpoint.HTTPErrors[0].ErrorExpr)
}

func TestHTTPInheritedErrorMappingRejectsIncompatibleError(t *testing.T) {
	for _, test := range []struct {
		name string
		dsl  func()
	}{
		{"different type", incompatibleHTTPErrorMappingDSL},
		{"different validation", incompatibleHTTPErrorValidationDSL},
		{"different named type", incompatibleHTTPNamedErrorMappingDSL},
		{"different object", incompatibleHTTPObjectErrorMappingDSL},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, test.dsl)
			require.ErrorContains(t, err, `HTTP error mapping "bad_request"`)
			require.ErrorContains(t, err, `method "Show" of service "Errors"`)
			require.ErrorContains(t, err, "must define the same error attribute")
		})
	}
}

func TestGRPCInheritedErrorMappingUsesMethodError(t *testing.T) {
	root := expr.RunDSL(t, equivalentGRPCErrorMappingDSL)
	endpoint := root.API.GRPC.Services[0].GRPCEndpoints[0]

	require.Same(t, endpoint.MethodExpr.Error("bad_request"), endpoint.GRPCErrors[0].ErrorExpr)
}

func TestServiceErrorMappingsUseMethodError(t *testing.T) {
	root := expr.RunDSL(t, equivalentServiceErrorMappingDSL)
	httpEndpoint := root.API.HTTP.Services[0].HTTPEndpoints[0]
	grpcEndpoint := root.API.GRPC.Services[0].GRPCEndpoints[0]
	methodError := httpEndpoint.MethodExpr.Error("bad_request")

	require.Same(t, methodError, httpEndpoint.HTTPErrors[0].ErrorExpr)
	require.Same(t, methodError, grpcEndpoint.GRPCErrors[0].ErrorExpr)
}

func TestHTTPInheritedErrorMappingAcceptsEquivalentValidationOrder(t *testing.T) {
	root := expr.RunDSL(t, equivalentHTTPErrorValidationOrderDSL)
	endpoint := root.API.HTTP.Services[0].HTTPEndpoints[0]

	require.Same(t, endpoint.MethodExpr.Error("bad_request"), endpoint.HTTPErrors[0].ErrorExpr)
}

func TestHTTPInheritedErrorMappingUsesEffectiveInheritedContract(t *testing.T) {
	root := expr.RunDSL(t, equivalentHTTPInheritedErrorMappingDSL)
	endpoint := root.API.HTTP.Services[0].HTTPEndpoints[0]

	require.Same(t, endpoint.MethodExpr.Error("bad_request"), endpoint.HTTPErrors[0].ErrorExpr)
}

func TestGRPCInheritedErrorMappingUsesEffectiveInheritedContract(t *testing.T) {
	root := expr.RunDSL(t, equivalentGRPCInheritedErrorMappingDSL)
	endpoint := root.API.GRPC.Services[0].GRPCEndpoints[0]

	require.Same(t, endpoint.MethodExpr.Error("bad_request"), endpoint.GRPCErrors[0].ErrorExpr)
}

func TestInheritedErrorMappingRejectsDifferentEffectiveBases(t *testing.T) {
	for _, test := range []struct {
		name string
		dsl  func()
	}{
		{"HTTP", incompatibleHTTPInheritedErrorMappingDSL},
		{"gRPC", incompatibleGRPCInheritedErrorMappingDSL},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, test.dsl)
			require.ErrorContains(t, err, `error mapping "bad_request"`)
			require.ErrorContains(t, err, "must define the same error attribute")
		})
	}
}

func TestGRPCInheritedErrorMappingRejectsIncompatibleError(t *testing.T) {
	err := expr.RunInvalidDSL(t, incompatibleGRPCErrorMappingDSL)
	require.ErrorContains(t, err, `gRPC error mapping "bad_request"`)
	require.ErrorContains(t, err, `method "Show" of service "Errors"`)
	require.ErrorContains(t, err, "must define the same error attribute")
}

func TestInheritedErrorMappingReportsDifferentQualifiers(t *testing.T) {
	qualifiers := []struct {
		name  string
		apply func()
	}{
		{name: "temporary", apply: Temporary},
		{name: "timeout", apply: func() { Timeout() }},
		{name: "fault", apply: Fault},
	}
	for _, transport := range []string{"HTTP", "gRPC"} {
		for _, qualifier := range qualifiers {
			t.Run(transport+" "+qualifier.name, func(t *testing.T) {
				err := expr.RunInvalidDSL(t, qualifierErrorMappingDSL(transport, qualifier.apply))
				require.ErrorContains(t, err, transport+` error mapping "busy"`)
				require.ErrorContains(t, err, qualifier.name+" setting differs")
			})
		}
	}
}

func qualifierErrorMappingDSL(transport string, qualifier func()) func() {
	return func() {
		API("errors", func() {
			Error("busy", qualifier)
			if transport == "HTTP" {
				HTTP(func() { Response(StatusServiceUnavailable, "busy") })
			} else {
				GRPC(func() { Response("busy", CodeUnavailable) })
			}
		})
		Service("Jobs", func() {
			Method("Run", func() {
				Error("busy")
				if transport == "HTTP" {
					HTTP(func() { POST("/run") })
				} else {
					GRPC(func() {})
				}
			})
		})
	}
}

var equivalentHTTPErrorMappingDSL = func() {
	API("errors", func() {
		Error("bad_request", String)
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", String)
			HTTP(func() { GET("/") })
		})
	})
}

var incompatibleHTTPErrorMappingDSL = func() {
	API("errors", func() {
		Error("bad_request", String)
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", Int)
			HTTP(func() { GET("/") })
		})
	})
}

var incompatibleHTTPErrorValidationDSL = func() {
	API("errors", func() {
		Error("bad_request", String, func() { MinLength(2) })
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", String, func() { MinLength(3) })
			HTTP(func() { GET("/") })
		})
	})
}

var equivalentHTTPErrorValidationOrderDSL = func() {
	API("errors", func() {
		Error("bad_request", String, func() { Enum("first", "second") })
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", String, func() { Enum("second", "first") })
			HTTP(func() { GET("/") })
		})
	})
}

var incompatibleHTTPNamedErrorMappingDSL = func() {
	firstError := Type("FirstError", func() { Attribute("message", String) })
	secondError := Type("SecondError", func() { Attribute("message", String) })
	API("errors", func() {
		Error("bad_request", firstError)
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", secondError)
			HTTP(func() { GET("/") })
		})
	})
}

var incompatibleHTTPObjectErrorMappingDSL = func() {
	stringError := Type("StringError", func() { Attribute("header") })
	API("errors", func() {
		Error("bad_request", stringError)
		HTTP(func() {
			Response("bad_request", StatusBadRequest, func() { Header("header") })
		})
	})
	Service("Errors", func() {
		Error("bad_request")
		Method("Show", func() { HTTP(func() { GET("/") }) })
	})
}

var equivalentGRPCErrorMappingDSL = func() {
	API("errors", func() {
		Error("bad_request", String)
		GRPC(func() { Response("bad_request", CodeInvalidArgument) })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", String)
			GRPC(func() {})
		})
	})
}

var incompatibleGRPCErrorMappingDSL = func() {
	API("errors", func() {
		Error("bad_request", String)
		GRPC(func() { Response("bad_request", CodeInvalidArgument) })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", Int)
			GRPC(func() {})
		})
	})
}

var equivalentServiceErrorMappingDSL = func() {
	Service("Errors", func() {
		Error("bad_request", String)
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
		GRPC(func() { Response("bad_request", CodeInvalidArgument) })
		Method("Show", func() {
			Error("bad_request", String)
			HTTP(func() { GET("/") })
			GRPC(func() {})
		})
	})
}

var equivalentHTTPInheritedErrorMappingDSL = func() {
	base := Type("HTTPErrorBase", func() {
		Attribute("message", String, func() {
			Default("invalid")
			Meta("struct:field:name", "Message")
		})
		Required("message")
	})
	API("errors", func() {
		Error("bad_request", &expr.Object{}, func() { Extend(base) })
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", &expr.Object{}, func() {
				Attribute("message", String, func() {
					Default("invalid")
					Meta("struct:field:name", "Message")
				})
				Required("message")
			})
			HTTP(func() { GET("/") })
		})
	})
}

var equivalentGRPCInheritedErrorMappingDSL = func() {
	base := Type("GRPCErrorBase", func() {
		Attribute("message", String, func() { Meta("rpc:tag", "1") })
		Required("message")
	})
	API("errors", func() {
		Error("bad_request", &expr.Object{}, func() { Extend(base) })
		GRPC(func() { Response("bad_request", CodeInvalidArgument) })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", &expr.Object{}, func() {
				Attribute("message", String, func() { Meta("rpc:tag", "1") })
				Required("message")
			})
			GRPC(func() {})
		})
	})
}

var incompatibleHTTPInheritedErrorMappingDSL = func() {
	stringBase := Type("HTTPStringErrorBase", func() { Attribute("value", String) })
	integerBase := Type("HTTPIntegerErrorBase", func() { Attribute("value", Int) })
	API("errors", func() {
		Error("bad_request", &expr.Object{}, func() { Extend(stringBase) })
		HTTP(func() { Response(StatusBadRequest, "bad_request") })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", &expr.Object{}, func() { Extend(integerBase) })
			HTTP(func() { GET("/") })
		})
	})
}

var incompatibleGRPCInheritedErrorMappingDSL = func() {
	stringBase := Type("GRPCStringErrorBase", func() {
		Attribute("value", String, func() { Meta("rpc:tag", "1") })
	})
	integerBase := Type("GRPCIntegerErrorBase", func() {
		Attribute("value", Int, func() { Meta("rpc:tag", "1") })
	})
	API("errors", func() {
		Error("bad_request", &expr.Object{}, func() { Extend(stringBase) })
		GRPC(func() { Response("bad_request", CodeInvalidArgument) })
	})
	Service("Errors", func() {
		Method("Show", func() {
			Error("bad_request", &expr.Object{}, func() { Extend(integerBase) })
			GRPC(func() {})
		})
	})
}
