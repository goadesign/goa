package codegen

import (
	"bytes"
	"testing"
)

func TestHeader(t *testing.T) {
	const (
		pack          = "testpackage"
		title         = "test title"
		noTitleHeader = `package testpackage

`
		titleHeader = `// Code generated by goagen v2.0.0-wip, command line:
// $ codegen.test
//
// test title
//
// The content of this file is auto-generated, DO NOT MODIFY

package testpackage

`
		singleImportHeader = `package testpackage

import 	"test"

`
		manyImportsHeader = `package testpackage

import (
	"test"
	"other"
)

`
	)
	var (
		imprt   = []*ImportSpec{&ImportSpec{Path: "test"}}
		imports = append(imprt, &ImportSpec{Path: "other"})
	)
	cases := map[string]struct {
		Title    string
		Imports  []*ImportSpec
		Expected string
	}{
		"no-title":      {Expected: noTitleHeader},
		"title":         {Title: title, Expected: titleHeader},
		"single-import": {Imports: imprt, Expected: singleImportHeader},
		"many-imports":  {Imports: imports, Expected: manyImportsHeader},
	}
	for k, tc := range cases {
		buf := new(bytes.Buffer)
		s := Header(tc.Title, "testpackage", tc.Imports)
		s.Write(buf)
		actual := buf.String()
		if actual != tc.Expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.Expected)
		}
	}
}
