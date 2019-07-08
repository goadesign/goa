package main

import (
	"reflect"
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
			Expected: "goa.design/goa/v3@v3.0.3-0.20190704022140-85024ebc66dc",
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
			Expected: "goa.design/goa/v3@v3.0.3-0.20190704022140-85024ebc66dc",
		},
		{
			Name: "with comment",
			Mod: `module calc
go 1.12
replace goa.design/goa/v3 => ../../../goa.design/goa
require (
        github.com/ikawaha/kagome v1.0.0 // indirect
        goa.design/goa/v3 v3.0.3-0.20190704022140-85024ebc66dc // indirect // comment
        goa.design/plugins/v3 v3.0.1
)
`,
			Package:  "goa.design/goa/v3",
			Expected: "goa.design/goa/v3@v3.0.3-0.20190704022140-85024ebc66dc",
		},
		{
			Name:     "require with comment",
			Mod:      " require goa.design/goa/v3 v3.0.3-0.20190704022140-85024ebc66dc// indirect // comment",
			Package:  "goa.design/goa/v3",
			Expected: "goa.design/goa/v3@v3.0.3-0.20190704022140-85024ebc66dc",
		},
		{
			Name: "comment out",
			Mod: `module calc
go 1.12
replace goa.design/goa/v3 => ../../../goa.design/goa
require (
        github.com/ikawaha/kagome v1.0.0 // indirect
        // goa.design/goa/v3 v3.0.3-0.20190704022140-85024ebc66dc
        goa.design/plugins/v3 v3.0.1
)
`,
			Package:  "goa.design/goa/v3",
			Expected: "goa.design/goa/v3",
		},
		{
			Name:     "without version",
			Mod:      "     goa.design/goa/v3//v3.0.3-0.20190704022140-85024ebc66dc",
			Package:  "goa.design/goa/v3",
			Expected: "goa.design/goa/v3",
		},
		{
			Name:     "different version",
			Mod:      "goa.design/goa/v2 v2.0.3-0.20190704022140-85024ebc66dc // comment",
			Package:  "goa.design/goa/v3",
			Expected: "goa.design/goa/v3",
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

func TestModPatternRegexp(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected []string
	}{
		{
			Name:     "goa v3 release version",
			Input:    "goa.design/goa/v3 v3.0.2",
			Expected: []string{"goa.design/goa/v3 v3.0.2", "goa.design/goa/v3", "v3.0.2"},
		},
		{
			Name:     "goa v3 commit number",
			Input:    "goa.design/goa/v3 02c84c9b39d845611bfe6e8a96465d698ea374b1",
			Expected: []string{"goa.design/goa/v3 02c84c9b39d845611bfe6e8a96465d698ea374b1", "goa.design/goa/v3", "02c84c9b39d845611bfe6e8a96465d698ea374b1"},
		},
		{
			Name:     "goa v3 with comment",
			Input:    "goa.design/goa/v3   v3.0.2 //comment1 //comment2 //comment3",
			Expected: []string{"goa.design/goa/v3   v3.0.2 //comment1 //comment2 //comment3", "goa.design/goa/v3", "v3.0.2"},
		},
		{
			Name:     "goa v3 with require and comment",
			Input:    "require goa.design/goa/v3 v3.0.2 //comment",
			Expected: []string{"require goa.design/goa/v3 v3.0.2 //comment", "goa.design/goa/v3", "v3.0.2"},
		},
		{
			Name:     "goa v2",
			Input:    "goa.design/goa/v2 v2.0.0",
			Expected: []string{"goa.design/goa/v2 v2.0.0", "goa.design/goa/v2", "v2.0.0"},
		},
		{
			Name:     "goa v123",
			Input:    "   goa.design/goa/v123    v123.4.5   ",
			Expected: []string{"   goa.design/goa/v123    v123.4.5   ", "goa.design/goa/v123", "v123.4.5"},
		},
		{
			Name:     "without version",
			Input:    "goa.design/goa/v3",
			Expected: nil,
		},
		{
			Name:     "comment out",
			Input:    "// goa.design/goa/v3 v3.0.0",
			Expected: nil,
		},
		{
			Name:     "irrelevant package",
			Input:    "github.com/ikawaha/kagome v1.10.0",
			Expected: nil,
		},
	}
	for _, c := range cases {
		if got := reMod.FindStringSubmatch(c.Input); !reflect.DeepEqual(c.Expected, got) {
			t.Errorf("expected %+v, got %+v", c.Expected, got)
		}
	}
}
