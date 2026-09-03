// This file verifies every core service package symbol participates in the
// single Go package namespace with authored types and other generated symbols.
package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type serviceCollisionResult struct {
	authored   string
	subject    string
	competitor string
}

// TestEveryServiceNameRoleSharesOnePackageNamespace catches a service symbol
// family that bypasses exact authored types or uses discovery order to choose
// collision suffixes.
func TestEveryServiceNameRoleSharesOnePackageNamespace(t *testing.T) {
	roles := []serviceNameRole{
		serviceInterfaceNameRole,
		serviceAutherNameRole,
		serviceAPINameRole,
		serviceAPIVersionNameRole,
		serviceNameConstantRole,
		serviceMethodNamesRole,
		serviceServerStreamNameRole,
		serviceClientStreamNameRole,
		serviceErrorConstructorNameRole,
		serviceViewConstructorNameRole,
		servicePrivateProjectionConstructorNameRole,
		serviceValidatorNameRole,
		serviceViewMapNameRole,
		serviceEndpointsNameRole,
		serviceNewEndpointsNameRole,
		serviceClientNameRole,
		serviceNewClientNameRole,
		serviceMethodEndpointNameRole,
		serviceEndpointInputNameRole,
		serviceRequestNameRole,
		serviceResponseNameRole,
		serviceServerInterceptorsNameRole,
		serviceClientInterceptorsNameRole,
		serviceInterceptorInfoNameRole,
		serviceInterceptorPayloadNameRole,
		serviceInterceptorResultNameRole,
		serviceInterceptorStreamingPayloadNameRole,
		serviceInterceptorStreamingResultNameRole,
		serviceInterceptorPayloadAccessNameRole,
		serviceInterceptorResultAccessNameRole,
		serviceInterceptorStreamingPayloadAccessNameRole,
		serviceInterceptorStreamingResultAccessNameRole,
		serviceInterceptorMethodInfoNameRole,
		serviceInterceptorServerUnaryInfoNameRole,
		serviceInterceptorClientUnaryInfoNameRole,
		serviceInterceptorStreamingSendInfoNameRole,
		serviceInterceptorStreamingRecvInfoNameRole,
		serviceServerEndpointWrapperNameRole,
		serviceClientEndpointWrapperNameRole,
		serviceServerInterceptorWrapperNameRole,
		serviceClientInterceptorWrapperNameRole,
		serviceServerStreamWrapperNameRole,
		serviceClientStreamWrapperNameRole,
		serviceTransformHelperNameRole,
		serviceExampleStructNameRole,
		serviceExampleConstructorNameRole,
		serviceExampleServerInterceptorsStructNameRole,
		serviceExampleServerInterceptorsConstructorNameRole,
		serviceExampleClientInterceptorsStructNameRole,
		serviceExampleClientInterceptorsConstructorNameRole,
	}

	for _, role := range roles {
		t.Run(fmt.Sprintf("role-%d", role), func(t *testing.T) {
			forward := serviceCollisionNames(t, role, false)
			reverse := serviceCollisionNames(t, role, true)
			wantGenerated := []string{"Symbol2", "Symbol3"}
			if role.visibility() == codegen.UnexportedName {
				wantGenerated = []string{"symbol", "symbol2"}
			}

			require.Equal(t, forward, reverse)
			require.Equal(t, "Symbol", forward.authored)
			require.ElementsMatch(t, wantGenerated, []string{
				forward.subject,
				forward.competitor,
			})
		})
	}
}

// serviceCollisionNames declares one role and an unrelated generated function
// in the requested order, then returns their frozen names.
func serviceCollisionNames(t *testing.T, role serviceNameRole, reverse bool) serviceCollisionResult {
	t.Helper()
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	generatedPackage := mustClaimTestPackage(t, generation, "generated.local/gen/calc")
	authored := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "Symbol",
		UID:           fmt.Sprintf("authored-symbol-%d", role),
	}
	authoredDeclaration, err := generatedPackage.DeclareUserType(authored)
	require.NoError(t, err)

	competitorRole := serviceErrorConstructorNameRole
	if role.visibility() == codegen.UnexportedName {
		competitorRole = servicePrivateProjectionConstructorNameRole
		if role == competitorRole {
			competitorRole = serviceTransformHelperNameRole
		}
	} else if role == competitorRole {
		competitorRole = serviceValidatorNameRole
	}
	subjectID := serviceSymbolID{
		role:    role,
		service: "calc",
		subject: "subject",
	}
	competitorID := serviceSymbolID{
		role:    competitorRole,
		service: "calc",
		subject: "competitor",
	}
	names := make(serviceNames)
	declare := func(id serviceSymbolID) *codegen.NameDeclaration {
		declaration, err := names.declare(generatedPackage, id, "Symbol")
		require.NoError(t, err)
		repeated, err := names.declare(generatedPackage, id, "Symbol")
		require.NoError(t, err)
		require.Same(t, declaration, repeated)
		return declaration
	}
	var subject, competitor *codegen.NameDeclaration
	if reverse {
		competitor = declare(competitorID)
		subject = declare(subjectID)
	} else {
		subject = declare(subjectID)
		competitor = declare(competitorID)
	}
	require.NoError(t, generation.Freeze())

	return serviceCollisionResult{
		authored:   authoredDeclaration.Name(),
		subject:    subject.Name(),
		competitor: competitor.Name(),
	}
}
