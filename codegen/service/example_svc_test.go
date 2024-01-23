package service

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/expr"
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
				codegen.RunDSL(t, c.DSL)
				expr.Root.GeneratedTypes = &expr.GeneratedRoot{}
				require.Len(t, expr.Root.Services, 3)
				fs := ExampleServiceFiles("", expr.Root)
				require.Len(t, fs, 3)
				for _, f := range fs {
					require.Greater(t, len(f.SectionTemplates), 0)
					var b bytes.Buffer
					require.NoError(t, f.SectionTemplates[0].Write(&b))
					line, err := b.ReadBytes('\n')
					require.NoError(t, err)
					got := string(bytes.TrimRight(line, "\n"))
					assert.Equal(t, c.Expected, got)
				}
			})
		}
	})
}
