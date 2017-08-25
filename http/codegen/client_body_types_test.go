package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	. "goa.design/goa/http/codegen/testing"
	httpdesign "goa.design/goa/http/design"
)

func TestBodyTypeDecl(t *testing.T) {
	const genpkg = "gen"

	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"body-user-inner", PayloadBodyUserInnerDSL, BodyUserInnerDeclCode},
		{"body-path-user-validate", PayloadBodyPathUserValidateDSL, BodyPathUserValidateDeclCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			httpdesign.RunHTTPDSL(t, c.DSL)
			fs := clientType(genpkg, httpdesign.Root.HTTPServices[0], make(map[string]struct{}))
			section := fs.Sections[1]
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
		{"body-user-inner", PayloadBodyUserInnerDSL, 3, BodyUserInnerInitCode},
		{"body-path-user-validate", PayloadBodyPathUserValidateDSL, 2, BodyPathUserValidateInitCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := clientType(genpkg, httpdesign.Root.HTTPServices[0], make(map[string]struct{}))
			section := fs.Sections[c.SectionIndex]
			code := codegen.SectionCode(t, section)
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

const BodyUserInnerDeclCode = `// MethodBodyUserInnerRequestBody is the type of the ServiceBodyUserInner
// MethodBodyUserInner HTTP endpoint request body.
type MethodBodyUserInnerRequestBody struct {
	Inner *InnerTypeRequestBody ` + "`" + `form:"inner,omitempty" json:"inner,omitempty" xml:"inner,omitempty"` + "`" + `
}
`

const BodyPathUserValidateDeclCode = `// MethodUserBodyPathValidateRequestBody is the type of the
// ServiceBodyPathUserValidate MethodUserBodyPathValidate HTTP endpoint request
// body.
type MethodUserBodyPathValidateRequestBody struct {
	A string ` + "`" + `form:"a" json:"a" xml:"a"` + "`" + `
}
`

const BodyUserInnerInitCode = `// NewMethodBodyUserInnerRequestBody builds the ServiceBodyUserInner service
// MethodBodyUserInner endpoint request body from a payload.
func NewMethodBodyUserInnerRequestBody(p *servicebodyuserinner.PayloadType) *MethodBodyUserInnerRequestBody {
	body := &MethodBodyUserInnerRequestBody{}
	if p.Inner != nil {
		body.Inner = innerTypeToInnerTypeRequestBodyNoDefault(p.Inner)
	}

	return body
}
`

const BodyPathUserValidateInitCode = `// NewMethodUserBodyPathValidateRequestBody builds the
// ServiceBodyPathUserValidate service MethodUserBodyPathValidate endpoint
// request body from a payload.
func NewMethodUserBodyPathValidateRequestBody(p *servicebodypathuservalidate.PayloadType) *MethodUserBodyPathValidateRequestBody {
	body := &MethodUserBodyPathValidateRequestBody{
		A: p.A,
	}

	return body
}
`
