package example

import (
	"bytes"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
)

func TestJSONRPCServerGeneration(t *testing.T) {
	// Reset servers data
	Servers = make(ServersData)

	// Create DSL for a service with both HTTP and JSON-RPC
	root := codegen.RunDSL(t, func() {
		dsl.API("testapi", func() {
			dsl.Server("testserver", func() {
				dsl.Host("localhost", func() {
					dsl.URI("http://localhost:8080")
				})
				dsl.Services("testsvc")
			})
		})
		dsl.Service("testsvc", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/jsonrpc")
			})
			dsl.Method("testmethod", func() {
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.Int)
					dsl.Required("value")
				})
				dsl.Result(dsl.Int)
				dsl.JSONRPC(func() {
				})
			})
		})
	})

	// Generate service data
	services := service.NewServicesData(root)

	// Generate server files
	files := ServerFiles("test/package", root, services)

	if len(files) == 0 {
		t.Fatal("No server files generated")
	}

	// Find the main.go file
	var mainFile *codegen.File
	for _, f := range files {
		if strings.HasSuffix(f.Path, "main.go") {
			mainFile = f
			break
		}
	}

	if mainFile == nil {
		t.Fatal("main.go file not found")
	}

	// Render the file to a buffer
	var buf bytes.Buffer
	for _, section := range mainFile.SectionTemplates {
		if err := section.Write(&buf); err != nil {
			t.Fatalf("Failed to render section %s: %v", section.Name, err)
		}
	}

	content := buf.String()

	// Check that httpPortF is declared
	if !strings.Contains(content, "httpPortF") {
		t.Error("Expected httpPortF to be declared")
	}
}
