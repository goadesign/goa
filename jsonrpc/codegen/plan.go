// This file declares the fixed import qualifiers used by JSON-RPC-generated
// files before service package aliases are frozen for the generation.
package codegen

import (
	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// Plan reserves every literal import qualifier used by JSON-RPC render
// templates, including WebSocket and server-sent event support.
func Plan(generation *codegen.Generation) error {
	if err := httpcodegen.Plan(generation); err != nil {
		return err
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bufio"),
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("mime/multipart"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("path"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("sync"),
		codegen.SimpleImport("sync/atomic"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("github.com/gorilla/websocket"),
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		codegen.GoaImport("jsonrpc"),
	}
	for _, spec := range imports {
		if err := generation.RequireImport(spec); err != nil {
			return err
		}
	}
	return nil
}
