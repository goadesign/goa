package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/testdata"
)

func TestBodyTypeDecl(t *testing.T) {
	const genpkg = "gen"

	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"body-user-inner", testdata.PayloadBodyUserInnerDSL, BodyUserInnerDeclCode},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL, BodyPathUserValidateDeclCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			expr.RunHTTPDSL(t, c.DSL)
			fs := clientType(genpkg, expr.Root.API.HTTP.Services[0], make(map[string]struct{}))
			section := fs.SectionTemplates[1]
			code := codegen.SectionCode(t, section)
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

func TestBodyTypeInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name         string
		DSL          func()
		SectionIndex int
		Code         string
	}{
		{"body-user-inner", testdata.PayloadBodyUserInnerDSL, 3, BodyUserInnerInitCode},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL, 2, BodyPathUserValidateInitCode},
		{"body-primitive-array-user-validate", testdata.PayloadBodyPrimitiveArrayUserValidateDSL, 2, BodyPrimitiveArrayUserValidateInitCode},
		{"result-body-user", testdata.ResultBodyObjectHeaderDSL, 2, ResultBodyObjectHeaderInitCode},
		{"result-explicit-body-primitive", testdata.ExplicitBodyPrimitiveResultMultipleViewsDSL, 1, ExplicitBodyPrimitiveResultMultipleViewsInitCode},
		{"result-explicit-body-user-type", testdata.ExplicitBodyUserResultMultipleViewsDSL, 2, ExplicitBodyUserResultMultipleViewsInitCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := clientType(genpkg, expr.Root.API.HTTP.Services[0], make(map[string]struct{}))
			section := fs.SectionTemplates[c.SectionIndex]
			code := codegen.SectionCode(t, section)
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

const BodyUserInnerDeclCode = `// MethodBodyUserInnerRequestBody is the type of the "ServiceBodyUserInner"
// service "MethodBodyUserInner" endpoint HTTP request body.
type MethodBodyUserInnerRequestBody struct {
	Inner *InnerTypeRequestBody ` + "`" + `form:"inner,omitempty" json:"inner,omitempty" xml:"inner,omitempty"` + "`" + `
}
`

const BodyPathUserValidateDeclCode = `// MethodUserBodyPathValidateRequestBody is the type of the
// "ServiceBodyPathUserValidate" service "MethodUserBodyPathValidate" endpoint
// HTTP request body.
type MethodUserBodyPathValidateRequestBody struct {
	A string ` + "`" + `form:"a" json:"a" xml:"a"` + "`" + `
}
`

const BodyPrimitiveArrayUserValidateInitCode = `// NewPayloadTypeRequestBody builds the HTTP request body from the payload of
// the "MethodBodyPrimitiveArrayUserValidate" endpoint of the
// "ServiceBodyPrimitiveArrayUserValidate" service.
func NewPayloadTypeRequestBody(p []*servicebodyprimitivearrayuservalidate.PayloadType) []*PayloadTypeRequestBody {
	body := make([]*PayloadTypeRequestBody, len(p))
	for i, val := range p {
		body[i] = &PayloadTypeRequestBody{
			A: val.A,
		}
	}
	return body
}
`

const BodyUserInnerInitCode = `// NewMethodBodyUserInnerRequestBody builds the HTTP request body from the
// payload of the "MethodBodyUserInner" endpoint of the "ServiceBodyUserInner"
// service.
func NewMethodBodyUserInnerRequestBody(p *servicebodyuserinner.PayloadType) *MethodBodyUserInnerRequestBody {
	body := &MethodBodyUserInnerRequestBody{}
	if p.Inner != nil {
		body.Inner = marshalServicebodyuserinnerInnerTypeToInnerTypeRequestBody(p.Inner)
	}
	return body
}
`

const BodyPathUserValidateInitCode = `// NewMethodUserBodyPathValidateRequestBody builds the HTTP request body from
// the payload of the "MethodUserBodyPathValidate" endpoint of the
// "ServiceBodyPathUserValidate" service.
func NewMethodUserBodyPathValidateRequestBody(p *servicebodypathuservalidate.PayloadType) *MethodUserBodyPathValidateRequestBody {
	body := &MethodUserBodyPathValidateRequestBody{
		A: p.A,
	}
	return body
}
`

const ResultBodyObjectHeaderInitCode = `// NewMethodBodyObjectHeaderResultOK builds a "ServiceBodyObjectHeader" service
// "MethodBodyObjectHeader" endpoint result from a HTTP "OK" response.
func NewMethodBodyObjectHeaderResultOK(body *MethodBodyObjectHeaderResponseBody, b *string) *servicebodyobjectheader.MethodBodyObjectHeaderResult {
	v := &servicebodyobjectheader.MethodBodyObjectHeaderResult{
		A: body.A,
	}
	v.B = b
	return v
}
`

const ExplicitBodyPrimitiveResultMultipleViewsInitCode = `// NewMethodExplicitBodyPrimitiveResultMultipleViewResulttypemultipleviewsOK
// builds a "ServiceExplicitBodyPrimitiveResultMultipleView" service
// "MethodExplicitBodyPrimitiveResultMultipleView" endpoint result from a HTTP
// "OK" response.
func NewMethodExplicitBodyPrimitiveResultMultipleViewResulttypemultipleviewsOK(body string, c *string) *serviceexplicitbodyprimitiveresultmultipleviewviews.ResulttypemultipleviewsView {
	v := body
	res := &serviceexplicitbodyprimitiveresultmultipleviewviews.ResulttypemultipleviewsView{
		A: &v,
	}
	res.C = c
	return res
}
`

const ExplicitBodyUserResultMultipleViewsInitCode = `// NewMethodExplicitBodyUserResultMultipleViewResulttypemultipleviewsOK builds
// a "ServiceExplicitBodyUserResultMultipleView" service
// "MethodExplicitBodyUserResultMultipleView" endpoint result from a HTTP "OK"
// response.
func NewMethodExplicitBodyUserResultMultipleViewResulttypemultipleviewsOK(body *UserType, c *string) *serviceexplicitbodyuserresultmultipleviewviews.ResulttypemultipleviewsView {
	v := &serviceexplicitbodyuserresultmultipleviewviews.UserTypeView{
		X: body.X,
		Y: body.Y,
	}
	res := &serviceexplicitbodyuserresultmultipleviewviews.ResulttypemultipleviewsView{
		A: v,
	}
	res.C = c
	return res
}
`
