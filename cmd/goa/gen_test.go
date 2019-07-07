package main

import (
	"strings"
	"testing"
)

func TestGenerator_goaPackage(t *testing.T) {
	cases := []struct {
		Name     string
		Version  int
		Expected string
	}{
		{
			Name:     "specify v2",
			Version:  2,
			Expected: "goa.design/goa",
		},
		{
			Name:     "specify v3, but go.mod file does not exist",
			Version:  3,
			Expected: "goa.design/goa/v3",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pkg, err := (&Generator{
				DesignVersion: c.Version,
			}).goaPackage()
			if err != nil {
				t.Fatalf("unexpected error, %v", err)
			}
			if pkg != c.Expected {
				t.Errorf("expected %v, got %v", c.Expected, pkg)
			}
		})
	}
}

func TestParseGoModGoaPackage(t *testing.T) {
	cases := []struct {
		Name     string
		Mod      string
		Package  string
		Expected string
	}{
		{
			Name:     "simple require",
			Mod:      "require    goa.design/goa/v3   v3.0.3-0.20190704022140-85024ebc66dc",
			Package:  "goa.design/goa/v3",
			Expected: "goa.design/goa/v3@v3.0.3-0.20190704022140-85024ebc66dc",
		},
		{
			Name: "in require block",
			Mod: `module calc
go 1.12
require (
        github.com/ikawaha/kagome v1.0.0 // indirect
        goa.design/goa/v3 v3.0.3-0.20190704022140-85024ebc66dc
        goa.design/plugins/v3 v3.0.1
)
`,
			Package:  "goa.design/goa/v3",
			Expected: `goa.design/goa/v3@v3.0.3-0.20190704022140-85024ebc66dc`,
		},
		{
			Name: "not found",
			Mod: `module calc
go 1.12
require (
        github.com/ikawaha/kagome v1.0.0 // indirect
        goa.design/plugins/v3 v3.0.1
)
`,
			Package:  "goa.design/goa/v3",
			Expected: "goa.design/goa/v3",
		},
		{
			Name: "with replace",
			Mod: `module calc
go 1.12
replace goa.design/goa/v3 => ../../../goa.design/goa
require (
        github.com/ikawaha/kagome v1.0.0 // indirect
        goa.design/goa/v3 v3.0.3-0.20190704022140-85024ebc66dc
        goa.design/plugins/v3 v3.0.1
)
`,
			Package:  "goa.design/goa/v3",
			Expected: `goa.design/goa/v3@v3.0.3-0.20190704022140-85024ebc66dc`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pkg, err := parseGoModGoaPackage("goa.design/goa/v3", strings.NewReader(c.Mod))
			if err != nil {
				t.Fatalf("unexpected error, %v", err)
			}
			if pkg != c.Expected {
				t.Errorf("expected %v, got %v", c.Expected, pkg)
			}
		})
	}
}
