package genclient_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_client/test_"

	var outDir string
	var files []string
	var genErr error

	var oldCommand string

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"codegen", "--out=" + outDir, "--design=foo"}
		oldCommand = codegen.CommandName
		codegen.CommandName = "app"
	})

	JustBeforeEach(func() {
		files, genErr = genclient.Generate()
	})

	AfterEach(func() {
		codegen.CommandName = oldCommand
		os.RemoveAll(outDir)
	})

	Context("with an action with security configured", func() {
		BeforeEach(func() {
			codegen.TempCount = 0
			securitySchemeDef := &design.SecuritySchemeDefinition{
				SchemeName: "jwt-1",
				Kind:       design.JWTSecurityKind,
			}
			design.Design = &design.APIDefinition{
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
				SecuritySchemes: []*design.SecuritySchemeDefinition{
					securitySchemeDef,
				},
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								QueryParams: &design.AttributeDefinition{
									Type: design.Object{
										"param": &design.AttributeDefinition{Type: design.Integer},
										"time":  &design.AttributeDefinition{Type: design.DateTime},
										"uuid":  &design.AttributeDefinition{Type: design.UUID},
									},
								},
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
								},
								Security: &design.SecurityDefinition{
									Scheme: securitySchemeDef,
								},
							},
						},
					},
				},
			}
			fooRes := design.Design.Resources["foo"]
			showAct := fooRes.Actions["show"]
			showAct.Parent = fooRes
			showAct.Routes[0].Parent = showAct
		})

		It("generates the correct client Fields", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(7))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "client.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("SignerJWT1 goaclient.Signer"))
			Ω(content).Should(ContainSubstring("SignerJWT1: &goaclient.JWTSigner{},"))
		})

		It("generates the Signer.Sign call from Action", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(7))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("c.SignerJWT1.Sign(ctx, req)"))
		})
	})
})
