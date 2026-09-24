// These tests exercise named recursive collections through the generated HTTP
// union codecs, alongside required primitive elements and field defaults.
package generator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestGenerateNamedRecursiveCollections(t *testing.T) {
	registry := testRegistry("gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles))
	codegen.RunDSL(t, namedRecursiveCollectionsDSL)
	directory := t.TempDir()
	generated := filepath.Join(directory, codegen.Gendir)
	writeGeneratedModule(t, generated, "generated.local/gen")
	_, err := generate(directory, "gen", false, registry)
	require.NoError(t, err)
	writeGeneratedContractTest(t, generated, filepath.Join("http", "trees", "server"), namedRecursiveCollectionsHTTPTest)
	runGeneratedTests(t, generated)
}

// namedRecursiveCollectionsDSL closes the map's forward reference while its
// type DSL runs, before Goa finalizes and plans the expression graph.
func namedRecursiveCollectionsDSL() {
	var tree, node expr.UserType
	tree = d.Type("Tree", d.MapOf(d.String, d.String), func() {
		expr.AsMap(tree).ElemType.Type = node
	})
	node = d.Type("Node", func() {
		d.Attribute("children", tree)
		d.Attribute("label", d.String)
	})
	labels := d.Type("Labels", d.ArrayOfRequired(d.String))
	root := d.Type("Root", func() {
		d.Attribute("tree", tree)
		d.Attribute("labels", labels)
		d.Attribute("count", d.Int)
		d.Attribute("enabled", d.Boolean, func() { d.Default(false) })
		d.Attribute("title", d.String, func() { d.Default("untitled") })
		d.OneOf("branch", func() {
			d.Attribute("children", tree)
			d.Attribute("label", d.String)
		})
		d.Required("tree", "labels", "count", "branch")
	})
	d.Service("trees", func() {
		d.Method("exchange", func() {
			d.Payload(root)
			d.Result(root)
			d.HTTP(func() {
				d.POST("/trees")
				d.Response(200)
			})
		})
	})
}

const namedRecursiveCollectionsHTTPTest = `package server
import (
	"encoding/json"
	"testing"
)
func TestNamedRecursiveValues(t *testing.T) {
	const document = ` + "`" + `{"tree":{"first":{"children":{"second":{"label":"leaf"}}},"absent":null},"labels":[""],"count":0,"branch":{"type":"children","value":{"item":{"label":"union leaf"}}}}` + "`" + `
	var body ExchangeRequestBody
	if err := json.Unmarshal([]byte(document), &body); err != nil { t.Fatal(err) }
	if err := ValidateExchangeRequestBody(&body); err != nil { t.Fatal(err) }
	value := NewExchangeRoot(&body)
	if value.Count != 0 || value.Enabled || value.Title != "untitled" { t.Fatal("defaults or zero changed") }
	if value.Tree["absent"] != nil { t.Fatal("nil map element changed") }
	if label := value.Tree["first"].Children["second"].Label; label == nil || *label != "leaf" { t.Fatal("lost nested value") }
	children, ok := value.Branch.AsChildren()
	if !ok { t.Fatal("lost children branch") }
	if label := children["item"].Label; label == nil || *label != "union leaf" { t.Fatal("lost union value") }
	response := NewExchangeResponseBody(value)
	encoded, err := json.Marshal(response)
	if err != nil { t.Fatal(err) }
	var again ExchangeRequestBody
	if err := json.Unmarshal(encoded, &again); err != nil { t.Fatal(err) }
	if err := ValidateExchangeRequestBody(&again); err != nil { t.Fatal(err) }
	if got := NewExchangeRoot(&again); got.Count != 0 || got.Title != "untitled" { t.Fatal("round trip changed values") }
	if err := json.Unmarshal([]byte(` + "`" + `{"tree":{},"labels":[null],"count":0,"branch":{"type":"label","value":""}}` + "`" + `), &body); err != nil { t.Fatal(err) }
	if err := ValidateExchangeRequestBody(&body); err == nil { t.Fatal("accepted null required element") }
}`
