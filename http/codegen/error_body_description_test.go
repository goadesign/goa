// This file verifies generated HTTP error body comments name the service
// errors that use each body type.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestErrorBodyDescriptionNamesSingleError(t *testing.T) {
	root := expr.RunDSL(t, testdata.WithErrorCustomPkgDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	want := "MethodWithErrorCustomPkgErrorNameResponseBody is the type of the\n" +
		"// \"ServiceWithErrorCustomPkg\" service \"MethodWithErrorCustomPkg\" endpoint HTTP\n" +
		"// response body for the \"error_name\" error."

	for _, file := range []struct {
		name     string
		sections string
	}{
		{name: "client", sections: renderHTTPSections(t, plan.ClientTypeFiles()[0])},
		{name: "server", sections: renderHTTPSections(t, plan.ServerTypeFiles()[0])},
	} {
		t.Run(file.name, func(t *testing.T) {
			require.Contains(t, file.sections, want)
		})
	}
}

// renderHTTPSections writes all generated sections after the file header.
func renderHTTPSections(t *testing.T, file *codegen.File) string {
	t.Helper()
	var rendered strings.Builder
	for _, section := range file.SectionTemplates[1:] {
		require.NoError(t, section.Write(&rendered))
	}
	return rendered.String()
}
