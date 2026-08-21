// This file declares the fixed import qualifiers used by HTTP-generated files
// before service package aliases are frozen for the generation.
package codegen

import (
	"goa.design/goa/v3/codegen"
)

// Plan reserves every literal import qualifier used by HTTP render templates.
// Generated service packages are planned separately and receive a suffix when
// their preferred qualifier conflicts with one of these required names.
func Plan(generation *codegen.Generation) error {
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bufio"),
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("flag"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("mime/multipart"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("path"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("sync"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("github.com/google/uuid"),
		codegen.SimpleImport("github.com/gorilla/websocket"),
		codegen.SimpleImport("goa.design/clue/debug"),
		codegen.SimpleImport("goa.design/clue/log"),
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		codegen.GoaImport("middleware"),
	}
	for _, spec := range imports {
		if err := generation.RequireImport(spec); err != nil {
			return err
		}
	}
	return nil
}
