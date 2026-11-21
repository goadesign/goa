// This file generates HTTP and JSON-RPC transports for multi-level result views
// and verifies that cascading View(...) onto ArrayOf and MapOf attributes
// preserves nested view fields and view-specific validations end to end.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	d "goa.design/goa/v3/dsl"
)

// TestGenerateCascadedResultViews generates a temporary module and runs HTTP
// and JSON-RPC runtime contract checks for cascaded result views.
func TestGenerateCascadedResultViews(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "cascaded_views", "views_test.go"))
	require.NoError(t, err)
	dir := generateViewedTransportModule(t, cascadedResultViewsDSL)
	writeGeneratedContractTest(t, dir, "checks", string(source))
	runGeneratedPackageTests(t, dir, "./checks")
}

// cascadedResultViewsDSL defines a three-level Order -> Item -> Label result
// hierarchy with default and extended views over HTTP and JSON-RPC.
func cascadedResultViewsDSL() {
	d.API("cascaded-views", func() {})

	label := d.ResultType("application/vnd.cascaded.label", func() {
		d.TypeName("Label")
		d.Field(1, "id", d.String, "Label identifier", func() { d.MinLength(1) })
		d.Field(2, "name", d.String, "Extended label name", func() { d.MinLength(2) })
		d.Required("id")
		d.View("default", func() {
			d.Attribute("id")
		})
		d.View("extended", func() {
			d.Attribute("id")
			d.Attribute("name")
		})
	})

	item := d.ResultType("application/vnd.cascaded.item", func() {
		d.TypeName("Item")
		d.Field(1, "id", d.String, "Item identifier", func() { d.MinLength(1) })
		d.Field(2, "labels", d.ArrayOf(label), "Labels attached to the item")
		d.Required("id")
		d.View("default", func() {
			d.Attribute("id")
			d.Attribute("labels")
		})
		d.View("extended", func() {
			d.Attribute("id")
			d.Attribute("labels", func() {
				d.View("extended")
			})
		})
	})

	order := d.ResultType("application/vnd.cascaded.order", func() {
		d.TypeName("Order")
		d.Field(1, "id", d.String, "Order identifier", func() { d.MinLength(1) })
		d.Field(2, "items", d.ArrayOf(item), "Ordered items")
		d.Field(3, "by_key", d.MapOf(d.String, item), "Items indexed by key")
		d.Required("id")
		d.View("default", func() {
			d.Attribute("id")
			d.Attribute("items")
			d.Attribute("by_key")
		})
		d.View("extended", func() {
			d.Attribute("id")
			d.Attribute("items", func() {
				d.View("extended")
			})
			d.Attribute("by_key", func() {
				d.View("extended")
			})
		})
	})

	d.Service("orders", func() {
		d.Method("get", func() {
			d.Payload(func() {
				d.Attribute("view", d.String, func() { d.Default("default") })
			})
			d.Result(order)
			d.HTTP(func() {
				d.GET("/orders")
				d.Param("view")
				d.Response(d.StatusOK)
			})
		})
		d.Method("get_extended", func() {
			d.Result(order, func() {
				d.View("extended")
			})
			d.HTTP(func() {
				d.GET("/orders/extended")
				d.Response(d.StatusOK)
			})
		})
	})

	d.Service("rpcorders", func() {
		d.JSONRPC(func() { d.POST("/rpc") })
		d.Method("get", func() {
			d.Payload(func() {
				d.Attribute("view", d.String, func() { d.Default("default") })
			})
			d.Result(order)
			d.JSONRPC(func() {})
		})
	})
}
