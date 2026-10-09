// This fixture generates ordinary service, HTTP and JSON-RPC contracts with
// the same untagged union. External JSON reaches typed endpoints, including a
// selected result view; invalid values must not reach the service.
package generator

import (
	_ "embed"
	"testing"

	d "goa.design/goa/v3/dsl"
)

//go:embed testdata/untagged_union/wire_test.go
var untaggedUnionRuntime string

func TestGenerateUntaggedUnionTransports(t *testing.T) {
	dir := generateViewedTransportModule(t, untaggedUnionDSL)
	writeGeneratedContractTest(t, dir, "checks", untaggedUnionRuntime)
	runGeneratedPackageTests(t, dir, "./checks")
}

// untaggedUnionDSL uses native aliases and result views so each transport
// consumes the shared declaration and preserves the selected branch's fields.
func untaggedUnionDSL() {
	d.API("untagged-union", func() {})
	file := d.Type("File", func() {
		d.Field(1, "uri", d.String, "Exact file address.", func() { d.MinLength(1) })
		d.Required("uri")
	})
	entry := d.Type("Entry", func() {
		d.OneOf("resources", "Complete files or dynamic content.", func() {
			d.Meta("oneof:json:untagged")
			d.Field(1, "manifest", d.ArrayOfRequired(file), "Complete files, including an empty array.")
			d.Field(2, "dynamic", d.String, "Content without stable digests.", func() { d.Enum("dynamic") })
		})
		d.Required("resources")
	})
	viewed := d.ResultType("application/vnd.untagged-union.entry", func() {
		d.TypeName("ViewedEntry")
		d.Extend(entry)
		d.View("default", func() { d.Attribute("resources") })
		d.View("selected", func() { d.Attribute("resources") })
	})
	selection := d.Type("Selection", func() {
		d.OneOf("value", "Exactly one typed value.", func() {
			d.Meta("oneof:json:untagged")
			d.Attribute("text", d.String, "Text, including an empty string.")
			d.Attribute("number", d.Int64, "An exact signed integer.")
			d.Attribute("enabled", d.Boolean, "An enabled or disabled setting.")
			d.Attribute("record", file, "A nonempty file address.")
			d.Attribute("files", d.ArrayOfRequired(file), "An ordered file collection.")
		})
		d.Required("value")
	})
	settings := d.Type("Settings", func() {
		d.OneOf("setting", "A typed setting supported by protobuf.", func() {
			d.Meta("oneof:json:untagged")
			d.Field(1, "text", d.String, "A text setting.")
			d.Field(2, "enabled", d.Boolean, "An enabled or disabled setting.")
		})
		d.Required("setting")
	})
	d.Service("records", func() {
		d.Method("echo", func() {
			d.Payload(entry)
			d.Result(entry)
			d.HTTP(func() {
				d.POST("/echo")
				d.Response(d.StatusOK)
			})
		})
		d.Method("viewed", func() {
			d.Payload(entry)
			d.Result(viewed)
			d.HTTP(func() {
				d.POST("/viewed")
				d.Response(d.StatusOK)
			})
		})
		d.Method("select", func() {
			d.Payload(selection)
			d.Result(selection)
			d.HTTP(func() {
				d.POST("/select")
				d.Response(d.StatusOK)
			})
		})
		d.Method("configure", func() {
			d.Payload(settings)
			d.Result(settings)
			d.GRPC(func() {})
		})
	})
	d.Service("rpc", func() {
		d.JSONRPC(func() { d.POST("/rpc") })
		d.Method("echo", func() {
			d.Payload(entry)
			d.Result(entry)
			d.JSONRPC(func() {})
		})
	})
}
