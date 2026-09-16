// This file verifies service package declarations are collected only for the
// conditions that emit them and never depend on declarations in other packages.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestRelocatedResultViewConstructorsCompile catches constructor declarations
// that incorrectly depend on a result type declaration owned by another Go
// package.
func TestRelocatedResultViewConstructorsCompile(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		reading := dsl.ResultType("application/vnd.reading", func() {
			dsl.TypeName("Reading")
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("name", dsl.String)
			dsl.Attribute("value", dsl.Int)
			dsl.Required("name", "value")
			dsl.View("default", func() {
				dsl.Attribute("name")
				dsl.Attribute("value")
			})
			dsl.View("summary", func() {
				dsl.Attribute("name")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(reading)
			})
		})
	})

	plan := retainedServicePlanForPackage(t, root)
	files, err := Files(plan)
	require.NoError(t, err)
	files = append(files, ExampleServiceFiles(plan)...)
	compileGeneratedServiceFiles(t, files)
}
