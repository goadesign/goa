package codegen

import (
	"bytes"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/testdata"
)

func TestServerTypes(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"multiple-methods", testdata.MultipleMethodsDSL, MultipleMethodsServerTypesFile},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := serverType(genpkg, expr.Root.API.HTTP.Services[0], make(map[string]struct{}))
			var buf bytes.Buffer
			for _, s := range fs.SectionTemplates[1:] {
				if err := s.Write(&buf); err != nil {
					t.Fatal(err)
				}
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

const MultipleMethodsServerTypesFile = `// MethodARequestBody is the type of the "ServiceMultipleMethods" service
// "MethodA" endpoint HTTP request body.
type MethodARequestBody struct {
	A *string ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
}

// MethodBRequestBody is the type of the "ServiceMultipleMethods" service
// "MethodB" endpoint HTTP request body.
type MethodBRequestBody struct {
	A *string              ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
	B *string              ` + "`" + `form:"b,omitempty" json:"b,omitempty" xml:"b,omitempty"` + "`" + `
	C *APayloadRequestBody ` + "`" + `form:"c,omitempty" json:"c,omitempty" xml:"c,omitempty"` + "`" + `
}

// APayloadRequestBody is used to define fields on request body types.
type APayloadRequestBody struct {
	A *string ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
}

// NewMethodAAPayload builds a ServiceMultipleMethods service MethodA endpoint
// payload.
func NewMethodAAPayload(body *MethodARequestBody) *servicemultiplemethods.APayload {
	v := &servicemultiplemethods.APayload{
		A: body.A,
	}
	return v
}

// NewMethodBPayloadType builds a ServiceMultipleMethods service MethodB
// endpoint payload.
func NewMethodBPayloadType(body *MethodBRequestBody) *servicemultiplemethods.PayloadType {
	v := &servicemultiplemethods.PayloadType{
		A: *body.A,
		B: body.B,
	}
	v.C = unmarshalAPayloadRequestBodyToServicemultiplemethodsAPayload(body.C)
	return v
}

// ValidateMethodARequestBody runs the validations defined on MethodARequestBody
func ValidateMethodARequestBody(body *MethodARequestBody) (err error) {
	if body.A != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.a", *body.A, "patterna"))
	}
	return
}

// ValidateMethodBRequestBody runs the validations defined on MethodBRequestBody
func ValidateMethodBRequestBody(body *MethodBRequestBody) (err error) {
	if body.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "body"))
	}
	if body.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "body"))
	}
	if body.A != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.a", *body.A, "patterna"))
	}
	if body.B != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.b", *body.B, "patternb"))
	}
	if body.C != nil {
		if err2 := ValidateAPayloadRequestBody(body.C); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateAPayloadRequestBody runs the validations defined on
// APayloadRequestBody
func ValidateAPayloadRequestBody(body *APayloadRequestBody) (err error) {
	if body.A != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.a", *body.A, "patterna"))
	}
	return
}
`
