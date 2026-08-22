// This file verifies the typed package-level declaration identities used by
// retained service plans before any generated source is rendered.
package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// TestServiceNamesAreIndependentOfDiscoveryOrder verifies that stable semantic
// identities, rather than traversal order, decide collision suffixes.
func TestServiceNamesAreIndependentOfDiscoveryOrder(t *testing.T) {
	ids := []serviceSymbolID{
		{role: serviceValidatorNameRole, service: "calc", subject: "Result"},
		{role: serviceErrorConstructorNameRole, service: "calc", subject: "Result"},
		{role: serviceMethodEndpointNameRole, service: "calc", method: "add"},
	}

	generate := func(order []int) map[serviceSymbolID]string {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg := mustClaimTestPackage(t, generation, "generated.local/gen/calc")
		names := make(serviceNames)
		for _, index := range order {
			_, err := names.declare(pkg, ids[index], "Build")
			require.NoError(t, err)
		}
		require.NoError(t, generation.Freeze())
		result := make(map[serviceSymbolID]string, len(ids))
		for _, id := range ids {
			result[id] = names.declaration(id).Name()
		}
		return result
	}

	require.Equal(t, generate([]int{0, 1, 2}), generate([]int{2, 0, 1}))
}

// TestServiceNamesShareTheAuthoredPackageNamespace verifies that generated
// functions collide with exact authored types and with one another in the one
// namespace enforced by the Go compiler.
func TestServiceNamesShareTheAuthoredPackageNamespace(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/calc")
	authored := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "ValidateResult",
		UID:           "authored-validate-result",
	}
	authoredDeclaration, err := pkg.DeclareUserType(authored)
	require.NoError(t, err)

	names := make(serviceNames)
	validator, err := names.declare(pkg, serviceSymbolID{
		role:    serviceValidatorNameRole,
		service: "calc",
		subject: "Result",
	}, "ValidateResult")
	require.NoError(t, err)
	constructor, err := names.declare(pkg, serviceSymbolID{
		role:    serviceErrorConstructorNameRole,
		service: "calc",
		subject: "Result",
	}, "ValidateResult")
	require.NoError(t, err)

	require.NoError(t, generation.Freeze())
	require.Equal(t, "ValidateResult", authoredDeclaration.Name())
	require.Equal(t, "ValidateResult2", constructor.Name())
	require.Equal(t, "ValidateResult3", validator.Name())
}

// TestServiceNamesOwnOneCanonicalDeclaration verifies that rebuilding an exact
// semantic identity returns the original record and rejects changed spelling
// or ownership.
func TestServiceNamesOwnOneCanonicalDeclaration(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	servicePackage := mustClaimTestPackage(t, generation, "generated.local/gen/calc")
	viewsPackage := mustClaimTestPackage(t, generation, "generated.local/gen/calc/views")
	names := make(serviceNames)
	id := serviceSymbolID{role: serviceViewConstructorNameRole, service: "calc", subject: "Result"}

	first, err := names.declare(servicePackage, id, "NewViewedResult")
	require.NoError(t, err)
	second, err := names.declare(servicePackage, id, "NewViewedResult")
	require.NoError(t, err)
	require.Same(t, first, second)

	_, err = names.declare(servicePackage, id, "NewResultView")
	require.ErrorContains(t, err, "cannot declare both")
	_, err = names.declare(viewsPackage, id, "NewViewedResult")
	require.ErrorContains(t, err, "already belongs")
}

// TestServiceNamesDeriveCompanionsFromFrozenTypes verifies validators follow
// the exact projected type declaration when that type receives a suffix.
func TestServiceNamesDeriveCompanionsFromFrozenTypes(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/calc/views")
	require.NoError(t, pkg.DeclareName(codegen.NewExactName(codegen.NameType, "Result")))
	base := codegen.NewPreferredName(codegen.NameType, "Result", codegen.ExportedName, serviceNameOrder{
		role:    serviceInterfaceNameRole,
		service: "calc",
		subject: "result",
	})
	require.NoError(t, pkg.DeclareName(base))
	names := make(serviceNames)
	validator, err := names.declareDependent(pkg, serviceSymbolID{
		role:    serviceValidatorNameRole,
		service: "calc",
		subject: "result",
	}, base, "Validate", "")
	require.NoError(t, err)

	require.NoError(t, generation.Freeze())
	require.Equal(t, "Result2", base.Name())
	require.Equal(t, "ValidateResult2", validator.Name())
}

// TestServiceNameRolesOwnDeclarationKinds verifies that the closed service
// symbol family, not an individual caller, selects each Go declaration kind.
func TestServiceNameRolesOwnDeclarationKinds(t *testing.T) {
	tests := []struct {
		role serviceNameRole
		kind codegen.PackageNameKind
	}{
		{serviceInterfaceNameRole, codegen.NameType},
		{serviceAutherNameRole, codegen.NameType},
		{serviceAPINameRole, codegen.NameConstant},
		{serviceAPIVersionNameRole, codegen.NameConstant},
		{serviceNameConstantRole, codegen.NameConstant},
		{serviceMethodNamesRole, codegen.NameVariable},
		{serviceMethodEventNameRole, codegen.NameType},
		{serviceServerStreamNameRole, codegen.NameType},
		{serviceClientStreamNameRole, codegen.NameType},
		{serviceStreamNameRole, codegen.NameType},
		{serviceEventNameRole, codegen.NameType},
		{serviceErrorConstructorNameRole, codegen.NameFunction},
		{serviceViewConstructorNameRole, codegen.NameFunction},
		{servicePrivateProjectionConstructorNameRole, codegen.NameFunction},
		{serviceValidatorNameRole, codegen.NameFunction},
		{serviceViewMapNameRole, codegen.NameVariable},
		{serviceEndpointsNameRole, codegen.NameType},
		{serviceNewEndpointsNameRole, codegen.NameFunction},
		{serviceClientNameRole, codegen.NameType},
		{serviceNewClientNameRole, codegen.NameFunction},
		{serviceMethodEndpointNameRole, codegen.NameFunction},
		{serviceEndpointInputNameRole, codegen.NameType},
		{serviceRequestNameRole, codegen.NameType},
		{serviceResponseNameRole, codegen.NameType},
		{serviceServerInterceptorsNameRole, codegen.NameType},
		{serviceClientInterceptorsNameRole, codegen.NameType},
		{serviceInterceptorInfoNameRole, codegen.NameType},
		{serviceInterceptorPayloadNameRole, codegen.NameType},
		{serviceInterceptorResultNameRole, codegen.NameType},
		{serviceInterceptorStreamingPayloadNameRole, codegen.NameType},
		{serviceInterceptorStreamingResultNameRole, codegen.NameType},
		{serviceInterceptorPayloadAccessNameRole, codegen.NameType},
		{serviceInterceptorResultAccessNameRole, codegen.NameType},
		{serviceInterceptorStreamingPayloadAccessNameRole, codegen.NameType},
		{serviceInterceptorStreamingResultAccessNameRole, codegen.NameType},
		{serviceServerEndpointWrapperNameRole, codegen.NameFunction},
		{serviceClientEndpointWrapperNameRole, codegen.NameFunction},
		{serviceServerInterceptorWrapperNameRole, codegen.NameFunction},
		{serviceClientInterceptorWrapperNameRole, codegen.NameFunction},
		{serviceServerStreamWrapperNameRole, codegen.NameType},
		{serviceClientStreamWrapperNameRole, codegen.NameType},
		{serviceTransformHelperNameRole, codegen.NameFunction},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("role-%d", test.role), func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			pkg := mustClaimTestPackage(t, generation, "generated.local/gen/calc")
			names := make(serviceNames)
			declaration, err := names.declare(pkg, serviceSymbolID{
				role:    test.role,
				service: "calc",
				subject: "result",
			}, "Symbol")
			require.NoError(t, err)
			require.Equal(t, test.kind, declaration.Kind())
		})
	}
}
