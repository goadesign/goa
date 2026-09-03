// This file checks every HTTP security scheme with named string credentials.
// Both directions must preserve service types while HTTP carries plain strings.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestNamedSecurityTypes(t *testing.T) {
	root := expr.RunDSL(t, testdata.NamedSecurityTypesDSL)
	plan := linkedHTTPPlanForRoot(t, root)

	serverFiles := plan.ServerFiles()
	require.Len(t, serverFiles, 2)
	serverCode := codegen.SectionsCode(t, serverFiles[1].Section("request-decoder"))
	serverCode += codegen.SectionsCode(t, plan.ServerTypeFiles()[0].Section("server-payload-init"))
	testutil.AssertGo(t, "testdata/golden/security_alias_server.go.golden", serverCode)

	clientFiles := plan.ClientFiles()
	require.Len(t, clientFiles, 2)
	clientCode := codegen.SectionsCode(t, clientFiles[1].Section("request-encoder"))
	testutil.AssertGo(t, "testdata/golden/security_alias_client.go.golden", clientCode)
}
