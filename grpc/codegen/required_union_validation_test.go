// This file checks the validation functions generated for gRPC server requests
// and client responses which contain a required OneOf.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

func TestRequiredUnionValidationUsesCompleteProtobufBranches(t *testing.T) {
	root := RunGRPCDSL(t, requiredUnionValidationDSL)
	services := CreateGRPCServices(root)

	for _, test := range []struct {
		Name  string
		Files []*codegen.File
	}{
		{Name: "server", Files: ServerTypeFiles(services)},
		{Name: "client", Files: ClientTypeFiles(services)},
	} {
		t.Run(test.Name, func(t *testing.T) {
			require.Len(t, test.Files, 1)
			generated := sectionCode(t, test.Files[0].SectionTemplates[1:]...)

			require.Contains(t, generated, `goa.MissingFieldError("choice", "message")`)
			require.Contains(t, generated, `goa.MissingFieldError("detail", "message.choice")`)
			require.Contains(t, generated, `goa.MissingFieldError("inactive", "message.choice")`)
			require.Contains(t, generated, `goa.MissingFieldError("blob", "message.choice")`)
			require.Contains(t, generated, `goa.MissingFieldError("metadata", "message.choice")`)
			require.Contains(t, generated, "if v == nil {")
			require.Contains(t, generated, "if v.Detail == nil {")
			require.Contains(t, generated, "if v.Inactive == nil {")
			require.Contains(t, generated, "if v.Metadata == nil {")
			require.Contains(t, generated, "if v.Blob == nil {")
			require.NotContains(t, generated, "if v.Token == nil {")
		})
	}
}

// requiredUnionValidationDSL creates branches whose values include scalars,
// messages, an empty message, a byte slice, a named string, and Any.
func requiredUnionValidationDSL() {
	token := d.Type("Token", d.String)
	detail := d.Type("Detail", func() {
		d.Field(1, "label", d.String)
		d.Required("label")
	})
	inactive := d.Type("Inactive", func() {})
	request := d.Type("RequestChoice", func() {
		d.OneOf("choice", func() {
			d.Field(1, "number", d.Int, func() { d.Minimum(1) })
			d.Field(2, "detail", detail)
			d.Field(3, "inactive", inactive)
			d.Field(4, "blob", d.Bytes)
			d.Field(5, "token", token)
			d.Field(6, "metadata", d.Any)
		})
		d.Required("choice")
	})
	response := d.Type("ResponseChoice", func() {
		d.OneOf("choice", func() {
			d.Field(1, "number", d.Int, func() { d.Minimum(1) })
			d.Field(2, "detail", detail)
			d.Field(3, "inactive", inactive)
			d.Field(4, "blob", d.Bytes)
			d.Field(5, "token", token)
			d.Field(6, "metadata", d.Any)
		})
		d.Required("choice")
	})
	d.Service("UnionValidation", func() {
		d.Method("Exchange", func() {
			d.Payload(request)
			d.Result(response)
			d.GRPC(func() {})
		})
	})
}
