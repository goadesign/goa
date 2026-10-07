// This file generates both JSON union mappings and verifies that an authored
// branch view excludes private fields through native HTTP and JSON-RPC peers.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	d "goa.design/goa/v3/dsl"
)

func TestGenerateUnionBranchViews(t *testing.T) {
	for _, test := range []struct {
		name    string
		flatten bool
	}{
		{"tagged", false},
		{"flat", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := generateViewedTransportModule(t, func() { unionBranchViewsDSL(test.flatten) })
			source, err := os.ReadFile(filepath.Join("testdata", "union_branch_views", "views_test.go"))
			require.NoError(t, err)
			writeGeneratedContractTest(t, dir, "checks", string(source))
			runGeneratedPackageTests(t, dir, "./checks")
		})
	}
}

// unionBranchViewsDSL selects different views of one result type in sibling
// branches. A plain object branch remains unaffected by result view selection.
func unionBranchViewsDSL(flatten bool) {
	d.API("union-branch-views", func() {})
	receipt := d.ResultType("application/vnd.branch.receipt", func() {
		d.TypeName("Receipt")
		d.Field(1, "reference", d.String, "Public receipt reference", func() { d.MinLength(1) })
		d.Field(2, "internal", d.String, "Value included only in the default view", func() { d.MinLength(1) })
		d.Required("reference", "internal")
		d.View("public", func() { d.Attribute("reference") })
		d.Field(3, "note", d.String, "Optional public note")
		d.View("note", func() { d.Attribute("note") })
		d.View("default", func() { d.Attribute("reference"); d.Attribute("internal") })
	})
	pending := d.Type("MoreInput", func() {
		d.Field(1, "state", d.String, "Unfinished operation state")
		d.Required("state")
	})
	group := d.Type("ReceiptGroup", func() {
		d.Field(1, "single", receipt, "One public receipt", func() { d.View("public") })
		d.Field(2, "many", d.ArrayOf(receipt, func() { d.View("public") }), "Public receipts in order")
		d.Field(3, "named", d.MapOf(d.String, receipt, func() { d.Elem(func() { d.View("public") }) }), "Public receipts by name")
		d.Field(4, "collected", d.CollectionOf(receipt), "Public result collection", func() { d.View("public") })
		d.Required("single", "many", "named", "collected")
	})
	outcome := d.ResultType("application/vnd.branch.outcome", func() {
		d.TypeName("OutcomeResult")
		d.OneOf("outcome", func() {
			d.TypeName("Outcome")
			if flatten {
				d.Meta("oneof:json:flatten")
			}
			d.Attribute("complete", receipt, "Public receipt", func() { d.View("public") })
			d.Attribute("full", receipt, "Full receipt")
			d.Attribute("note", receipt, "Optional public note only", func() { d.View("note") })
			d.Attribute("group", group, "Public receipts in nested containers")
			d.Attribute("input_required", pending, "Input still required")
		})
		d.Required("outcome")
		d.View("default", func() { d.Attribute("outcome") })
	})
	d.Service("receipts", func() {
		d.Method("book", func() {
			d.Payload(func() {
				d.Field(1, "destination", d.String, "Requested destination")
				d.Required("destination")
			})
			d.Result(outcome)
			d.HTTP(func() { d.POST("/book"); d.Response(d.StatusOK) })
		})
	})
	d.Service("rpcreceipts", func() {
		d.JSONRPC(func() { d.POST("/rpc") })
		d.Method("book", func() {
			d.Payload(func() {
				d.Field(1, "destination", d.String, "Requested destination")
				d.Required("destination")
			})
			d.Result(outcome)
			d.JSONRPC(func() {})
		})
	})
}
