// These tests compile generated collection conversions and verify that union
// values retain their selected branch across HTTP request and response bodies.
package generator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestGenerateCollectionUnionValues(t *testing.T) {
	registry := testRegistry("gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles))
	codegen.RunDSL(t, collectionUnionValuesDSL)
	directory := t.TempDir()
	generated := filepath.Join(directory, codegen.Gendir)
	writeGeneratedModule(t, generated, "generated.local/gen")
	_, err := generate(directory, "gen", false, registry)
	require.NoError(t, err)
	writeGeneratedContractTest(t, generated, filepath.Join("http", "collections", "server"), collectionUnionValuesHTTPTest)
	runGeneratedTests(t, generated)
}

func collectionUnionValuesDSL() {
	entry := d.Type("Entry", func() {
		d.Attribute("label", d.String)
		d.Required("label")
	})
	choices := d.Type("Choices", d.ArrayOf(&expr.Union{TypeName: "Choice"}, func() {
		d.Attribute("text", d.String)
		d.Attribute("number", d.Int, func() { d.Minimum(1) })
		d.Attribute("entry", entry)
	}))
	lookup := d.Type("Lookup", d.MapOf(d.String, &expr.Union{TypeName: "MapChoice"}, func() {
		d.Elem(func() {
			d.Attribute("text", d.String)
			d.Attribute("number", d.Int, func() { d.Minimum(1) })
			d.Attribute("entry", entry)
		})
	}))
	root := d.Type("Root", func() {
		d.Attribute("choices", choices)
		d.Attribute("lookup", lookup)
		d.Attribute("rows", d.ArrayOf(d.ArrayOf(d.String)))
		d.Attribute("rowLookup", d.MapOf(d.String, d.ArrayOf(d.String)))
		d.Attribute("requiredRows", d.ArrayOfRequired(d.ArrayOf(d.String)))
		d.Required("choices", "lookup", "rows", "rowLookup", "requiredRows")
	})
	d.Service("collections", func() {
		d.Method("exchange", func() {
			d.Payload(root)
			d.Result(root)
			d.HTTP(func() {
				d.POST("/collections")
				d.Response(200)
			})
		})
	})
}

const collectionUnionValuesHTTPTest = `package server
import (
	"encoding/json"
	"reflect"
	"testing"
)
func TestCollectionUnionRoundTrip(t *testing.T) {
	const document = ` + "`" + `{"choices":[{"type":"text","value":"hello"},{"type":"number","value":1},{"type":"entry","value":{"label":""}}],"lookup":{"first":{"type":"text","value":"world"},"second":{"type":"entry","value":{"label":""}}},"rows":[null,[]],"rowLookup":{"nil":null,"empty":[]},"requiredRows":[[]]}` + "`" + `
	var body ExchangeRequestBody
	if err := json.Unmarshal([]byte(document), &body); err != nil { t.Fatal(err) }
	if err := ValidateExchangeRequestBody(&body); err != nil { t.Fatal(err) }
	value := NewExchangeRoot(&body)
	if text, ok := value.Choices[0].AsText(); !ok || text != "hello" { t.Fatal("lost array text branch") }
	if number, ok := value.Choices[1].AsNumber(); !ok || number != 1 { t.Fatal("lost array number branch") }
	entry := value.Lookup["first"]
	if text, ok := entry.AsText(); !ok || text != "world" { t.Fatal("lost map text branch") }
	if value.Rows[0] != nil || value.Rows[1] == nil { t.Fatal("changed nil and empty array rows") }
	if value.RowLookup["nil"] != nil || value.RowLookup["empty"] == nil { t.Fatal("changed nil and empty map values") }
	response := NewExchangeResponseBody(value)
	data, err := json.Marshal(response)
	if err != nil { t.Fatal(err) }
	var again ExchangeRequestBody
	if err := json.Unmarshal(data, &again); err != nil { t.Fatal(err) }
	if err := ValidateExchangeRequestBody(&again); err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(value, NewExchangeRoot(&again)) { t.Fatal("round trip changed collection values") }
}
func TestCollectionUnionInvalidBranch(t *testing.T) {
	for _, document := range []string{
		` + "`" + `{"choices":[{"type":"number","value":0}],"lookup":{},"rows":[],"rowLookup":{},"requiredRows":[]}` + "`" + `,
		` + "`" + `{"choices":[],"lookup":{"first":{"type":"number","value":0}},"rows":[],"rowLookup":{},"requiredRows":[]}` + "`" + `,
		` + "`" + `{"choices":[],"lookup":{},"rows":[],"rowLookup":{},"requiredRows":[null]}` + "`" + `,
		` + "`" + `{"choices":[{"type":"entry","value":{}}],"lookup":{},"rows":[],"rowLookup":{},"requiredRows":[]}` + "`" + `,
		` + "`" + `{"choices":[],"lookup":{"first":{"type":"entry","value":{}}},"rows":[],"rowLookup":{},"requiredRows":[]}` + "`" + `,
	} {
		var body ExchangeRequestBody
		if err := json.Unmarshal([]byte(document), &body); err != nil { continue }
		if err := ValidateExchangeRequestBody(&body); err == nil { t.Fatalf("accepted invalid collection element: %s", document) }
	}
}
`
