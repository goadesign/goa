// This file verifies that gRPC metadata uses the exact string conversion for
// each primitive type selected by the design.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestMetadataEncodingSpecializesPrimitiveFormatting(t *testing.T) {
	root := expr.RunDSL(t, func() {
		alias := dsl.Type("Count", dsl.Int)
		fields := dsl.Type("Fields", func() {
			dsl.Field(1, "boolean", dsl.Boolean)
			dsl.Field(2, "integer", dsl.Int)
			dsl.Field(3, "small", dsl.Int32)
			dsl.Field(4, "large", dsl.Int64)
			dsl.Field(5, "unsigned", dsl.UInt)
			dsl.Field(6, "unsigned_small", dsl.UInt32)
			dsl.Field(7, "unsigned_large", dsl.UInt64)
			dsl.Field(8, "ratio", dsl.Float32)
			dsl.Field(9, "score", dsl.Float64)
			dsl.Field(10, "text", dsl.String)
			dsl.Field(11, "bytes", dsl.Bytes)
			dsl.Field(12, "count", alias)
			dsl.Field(13, "booleans", dsl.ArrayOf(dsl.Boolean))
			dsl.Field(14, "dynamic", dsl.Any)
			dsl.Field(15, "dynamic_values", dsl.ArrayOf(dsl.Any))
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(fields)
				dsl.Result(fields)
				dsl.GRPC(func() {
					dsl.Metadata(func() {
						metadataFields()
					})
					dsl.Response(func() {
						dsl.Headers(func() {
							metadataFields()
						})
					})
				})
			})
		})
	})

	services := CreateGRPCServices(root)
	generatedClientFiles := clientFiles(services)
	generatedServerFiles := serverFiles(services)
	request := codegen.SectionsCode(t, generatedClientFiles[1].Section("request-encoder"))
	response := codegen.SectionsCode(t, generatedServerFiles[1].Section("response-encoder"))
	for _, generated := range []string{request, response} {
		require.Contains(t, generated, "strconv.FormatBool(booleanWire)")
		require.Contains(t, generated, "strconv.Itoa(integerWire)")
		require.Contains(t, generated, "strconv.FormatInt(int64(smallWire), 10)")
		require.Contains(t, generated, "strconv.FormatInt(largeWire, 10)")
		require.Contains(t, generated, "strconv.FormatUint(uint64(unsignedWire), 10)")
		require.Contains(t, generated, "strconv.FormatUint(uint64(unsignedSmallWire), 10)")
		require.Contains(t, generated, "strconv.FormatUint(unsignedLargeWire, 10)")
		require.Contains(t, generated, "strconv.FormatFloat(float64(ratioWire), 'f', -1, 32)")
		require.Contains(t, generated, "strconv.FormatFloat(scoreWire, 'f', -1, 64)")
		require.Contains(t, generated, `Append("text", textWire)`)
		require.Contains(t, generated, `Append("bytes", string(bytesWire))`)
		require.Contains(t, generated, "strconv.Itoa(countWire)")
		require.Contains(t, generated, "strconv.FormatBool(value)")
		require.Contains(t, generated, `fmt.Sprintf("%v", dynamicWire)`)
		require.Contains(t, generated, `fmt.Sprintf("%v", value)`)
		require.NotContains(t, generated, `fmt.Sprintf("%v", integerWire)`)
		require.NotContains(t, generated, `fmt.Sprintf("%v", booleanWire)`)
	}
	require.Contains(t, sectionCode(t, generatedClientFiles[1].SectionTemplates[0]), `"fmt"`)
	require.Contains(t, sectionCode(t, generatedServerFiles[1].SectionTemplates[0]), `"fmt"`)

	withoutAny := CreateGRPCServices(RunGRPCDSL(t, testdata.MessageWithMetadataDSL))
	require.NotContains(t, sectionCode(t, clientFiles(withoutAny)[1].SectionTemplates[0]), `"fmt"`)
	require.NotContains(t, sectionCode(t, serverFiles(withoutAny)[1].SectionTemplates[0]), `"fmt"`)
}

// metadataFields maps every test field to a metadata key with the same name.
func metadataFields() {
	for _, name := range []string{
		"boolean",
		"integer",
		"small",
		"large",
		"unsigned",
		"unsigned_small",
		"unsigned_large",
		"ratio",
		"score",
		"text",
		"bytes",
		"count",
		"booleans",
		"dynamic",
		"dynamic_values",
	} {
		dsl.Attribute(name)
	}
}
