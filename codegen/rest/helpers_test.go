package rest

import (
	"bytes"
	"testing"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/service"
	"goa.design/goa.v2/design/rest"
)

// RunRestDSL returns the rest DSL root resulting from running the given DSL.
func RunRestDSL(t *testing.T, dsl func()) *rest.RootExpr {
	// reset all roots and codegen data structures
	service.Services = make(service.ServicesData)
	Resources = make(ResourcesData)
	return rest.RunRestDSL(t, dsl)
}

// SectionCode generates and formats the code for the given section.
func SectionCode(t *testing.T, section *codegen.Section) string {
	var code bytes.Buffer
	if err := section.Write(&code); err != nil {
		t.Fatal(err)
	}
	return codegen.FormatTestCode(t, "package foo\n"+code.String())
}
