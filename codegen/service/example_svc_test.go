package service

import (
	"bytes"
	"reflect"
	"testing"

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
				if len(expr.Root.Services) != 3 {
					t.Fatalf("got %d services, expected 3", len(expr.Root.Services))
				}
				fs := ExampleServiceFiles("", expr.Root)
				if len(fs) != 3 {
					t.Fatalf("got %d example file services, expected 3", len(fs))
				}
				for _, f := range fs {
					if len(f.SectionTemplates) == 0 {
						t.Fatalf("got empty templates, expected not empty")
					}
					var b bytes.Buffer
					if err := f.SectionTemplates[0].Write(&b); err != nil {
						t.Fatal(err)
					}
					if line, err := b.ReadBytes('\n'); err != nil {
						t.Fatal(err)
					} else if got := string(bytes.TrimRight(line, "\n")); !reflect.DeepEqual(got, c.Expected) {
						t.Fatalf("got %s, expected %s", got, c.Expected)
					}
				}
			})
		}
	})
}
