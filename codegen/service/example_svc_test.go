// This file verifies the starter service implementations generated from
// normalized service methods and their frozen package references.
package service

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/codegen/testutil"
)

func TestExampleServiceFiles(t *testing.T) {
	t.Run("package name check", func(t *testing.T) {
		cases := []struct {
			Name     string
			DSL      func()
			Expected string
		}{
			{
				Name:     "conflict with API name and service names",
				DSL:      testdata.ConflictWithAPINameAndServiceNameDSL,
				Expected: "package alohaapi2",
			},
			{
				Name:     "conflict with goified API name and goified service names",
				DSL:      testdata.ConflictWithGoifiedAPINameAndServiceNamesDSL,
				Expected: "package goodbyapi2",
			},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				root := codegen.RunDSL(t, c.DSL)
				plan := mustServicePlan(t, root)
				require.Len(t, root.Services, 3)
				fs := ExampleServiceFiles(plan)
				require.Len(t, fs, 3)
				for _, f := range fs {
					require.Greater(t, len(f.SectionTemplates), 0)
					var b bytes.Buffer
					require.NoError(t, f.SectionTemplates[0].Write(&b))
					line, err := b.ReadBytes('\n')
					require.NoError(t, err)
					got := string(bytes.TrimRight(line, "\r\n"))
					assert.Equal(t, c.Expected, got)
				}
			})
		}
	})

	t.Run("mixed result methods", func(t *testing.T) {
		cases := []struct {
			Name   string
			DSL    func()
			Golden string
		}{
			{"result and stream", testdata.MixedResultsEndpointDSL, "testdata/golden/example_service-mixed-results.go.golden"},
			{"result view and stream", testdata.MixedResultsWithViewsEndpointDSL, "testdata/golden/example_service-mixed-results-with-views.go.golden"},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				root := codegen.RunDSL(t, c.DSL)
				plan := mustServicePlan(t, root)
				files := ExampleServiceFiles(plan)
				require.Len(t, files, 1)
				testutil.AssertGo(t, c.Golden, renderSections(t, files[0].SectionTemplates))
			})
		}
	})
}
