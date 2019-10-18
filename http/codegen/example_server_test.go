package codegen

import (
	"bytes"
	"reflect"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestExampleServerFiles(t *testing.T) {
	t.Run("package name check", func(t *testing.T) {
		cases := []struct {
			Name     string
			DSL      func()
			Expected string
		}{
			{
				Name:     "conflict with API name and service names including multipart",
				DSL:      ctestdata.ConflictWithAPINameAndServiceNamesIncludingMultipartDSL,
				Expected: "package alohaapi2",
			},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				// reset global variable
				HTTPServices = make(ServicesData)
				service.Services = make(service.ServicesData)
				example.Servers = make(example.ServersData)
				codegen.RunDSL(t, c.DSL)
				if len(expr.Root.Services) != 3 {
					t.Fatalf("got %d services, expected 3", len(expr.Root.Services))
				}
				fs := ExampleServerFiles("", expr.Root)
				if len(fs) != 2 {
					t.Fatalf("got %d example files, expected 2", len(fs))
				}
				for i, f := range fs {
					if i < len(fs)-1 {
						// Skip example http server.
						continue
					}
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

	t.Run("code check", func(t *testing.T) {
		cases := []struct {
			Name string
			DSL  func()
			Code string
		}{
			{"no-server", ctestdata.NoServerDSL, testdata.NoServerServerHandleCode},
			{"server-hosting-service-with-file-server", ctestdata.ServerHostingServiceWithFileServerDSL, testdata.ServerHostingServiceWithFileServerHandlerCode},
			{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL, testdata.ServerHostingServiceSubsetServerHandleCode},
			{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL, testdata.ServerHostingMultipleServicesServerHandleCode},
			{"streaming", testdata.StreamingMultipleServicesDSL, testdata.StreamingServerHandleCode},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				// reset global variable
				HTTPServices = make(ServicesData)
				service.Services = make(service.ServicesData)
				example.Servers = make(example.ServersData)
				codegen.RunDSL(t, c.DSL)
				fs := ExampleServerFiles("", expr.Root)
				if len(fs) == 0 {
					t.Fatalf("got 0 files, expected 1")
				}
				if len(fs[0].SectionTemplates) == 0 {
					t.Fatalf("got 0 sections, expected at least 1")
				}
				var buf bytes.Buffer
				for _, s := range fs[0].SectionTemplates[1:] {
					if err := s.Write(&buf); err != nil {
						t.Fatal(err)
					}
				}
				code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
				if code != c.Code {
					t.Errorf("invalid code for %s: got\n%s\ngot vs. expected:\n%s", fs[0].Path, code, codegen.Diff(t, code, c.Code))
				}
			})
		}
	})
}
