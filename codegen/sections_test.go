package codegen

import (
	"bytes"
	"fmt"
	"testing"

	"goa.design/goa/v3/pkg"
)

func TestHeader(t *testing.T) {
	const (
		title         = "test title"
		noTitleHeader = `package testpackage

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

		pathImportHeader = `package testpackage

import 	"import/with/slashes"

`

		pathImportsHeader = `package testpackage

import (
	"import/with/slashes"
	"other/import/with/slashes"
)

`

		pathNamedImportHeader = `package testpackage

import 	myname "import/with/slashes"

`
		pathNamedImportsHeader = `package testpackage

import (
	myname "import/with/slashes"
	myothername "other/import/with/slashes"
)

`
	)
	var (
		titleHeader = fmt.Sprintf(`// Code generated by goa %s, DO NOT EDIT.
//
// test title
//
// Command:
// $ goa

package testpackage

`, pkg.Version())
		imprt            = []*ImportSpec{{Path: "test"}}
		imports          = append(imprt, &ImportSpec{Path: "other"})
		pathImport       = []*ImportSpec{{Path: "import/with/slashes"}}
		pathImports      = append(pathImport, &ImportSpec{Path: "other/import/with/slashes"})
		pathNamedImport  = []*ImportSpec{{Name: "myname", Path: "import/with/slashes"}}
		pathNamedImports = append(pathNamedImport, &ImportSpec{Name: "myothername", Path: "other/import/with/slashes"})
	)
	cases := map[string]struct {
		Title    string
		Imports  []*ImportSpec
		Expected string
	}{
		"no-title":           {Expected: noTitleHeader},
		"title":              {Title: title, Expected: titleHeader},
		"single-import":      {Imports: imprt, Expected: singleImportHeader},
		"many-imports":       {Imports: imports, Expected: manyImportsHeader},
		"path-import":        {Imports: pathImport, Expected: pathImportHeader},
		"path-imports":       {Imports: pathImports, Expected: pathImportsHeader},
		"path-named-import":  {Imports: pathNamedImport, Expected: pathNamedImportHeader},
		"path-named-imports": {Imports: pathNamedImports, Expected: pathNamedImportsHeader},
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
