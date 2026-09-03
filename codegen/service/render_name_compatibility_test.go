// This file verifies that plugins still receive the released Go-name strings
// while Goa templates use the matching planned declarations.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

func TestLinkedRenderNamesMatchDeclarations(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		customError := dsl.Type("CustomError", func() {
			dsl.Attribute("message", dsl.String)
			dsl.Required("message")
		})
		reading := dsl.ResultType("application/vnd.reading", func() {
			dsl.TypeName("Reading")
			dsl.Attribute("value", dsl.String, func() {
				dsl.MinLength(1)
			})
			dsl.Required("value")
			dsl.View("default", func() {
				dsl.Attribute("value")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(reading)
				dsl.Error("failed")
				dsl.Error("custom", customError)
			})
		})
	})

	data := mustServicePlan(t, root).Services().Get("Values")
	endpoints := endpointData(data)
	require.Equal(t, endpoints.EndpointsDeclaration.Name(), endpoints.VarName)
	require.Equal(t, endpoints.ClientDeclaration.Name(), endpoints.ClientVarName)
	require.Equal(t, endpoints.ServiceDeclaration.Name(), endpoints.ServiceVarName)
	for _, method := range endpoints.Methods {
		require.Equal(t, method.ClientDeclaration.Name(), method.ClientVarName)
		require.Equal(t, method.ServiceDeclaration.Name(), method.ServiceVarName)
	}
	require.Len(t, data.errorInits, 1)
	require.Equal(t, "MakeFailed", data.errorInits[0].Name)
	assertErrorInitName(t, data.errorInits[0])

	method := data.Method("Read")
	require.NotEmpty(t, data.ViewsPkg)
	require.Len(t, method.Errors, 2)
	for _, serviceError := range method.Errors {
		switch serviceError.ErrName {
		case "failed":
			assertErrorInitName(t, serviceError)
		case "custom":
			require.Nil(t, serviceError.Declaration)
			require.Empty(t, serviceError.Name)
		default:
			t.Errorf("unexpected service error %q", serviceError.ErrName)
		}
	}
	require.NotNil(t, method.ViewedResult)
	require.Equal(t, "NewViewedReading", method.ViewedResult.Init.Name)
	require.Equal(t, "ValidateReading", method.ViewedResult.Validate.Name)
	assertInitName(t, method.ViewedResult.Init)
	assertInitName(t, method.ViewedResult.ResultInit)
	assertValidateName(t, method.ViewedResult.Validate)

	for _, projected := range data.projectedTypes {
		require.Equal(t, data.ViewsPkg, projected.ViewsPkg)
		for _, init := range projected.Projections {
			assertInitName(t, init)
		}
		for _, init := range projected.TypeInits {
			assertInitName(t, init)
		}
		for _, validation := range projected.Validations {
			assertValidateName(t, validation)
		}
	}
}

// assertErrorInitName checks the compatibility name exposed to plugins.
func assertErrorInitName(t *testing.T, data *ErrorInitData) {
	t.Helper()
	require.NotEmpty(t, data.Name)
	require.Equal(t, data.Declaration.Name(), data.Name)
}

// assertInitName checks the compatibility name exposed to plugins.
func assertInitName(t *testing.T, data *InitData) {
	t.Helper()
	require.NotEmpty(t, data.Name)
	require.Equal(t, data.Declaration.Name(), data.Name)
}

// assertValidateName checks the compatibility name exposed to plugins.
func assertValidateName(t *testing.T, data *ValidateData) {
	t.Helper()
	require.NotEmpty(t, data.Name)
	require.Equal(t, data.Declaration.Name(), data.Name)
}
