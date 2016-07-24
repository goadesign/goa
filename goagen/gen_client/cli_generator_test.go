package genclient_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_client"
	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_client/test_"

	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + outDir, "--design=foo", "--version=" + version.String()}
	})

	JustBeforeEach(func() {
		files, genErr = genclient.Generate()
	})

	AfterEach(func() {
		os.RemoveAll(outDir)
		delete(codegen.Reserved, "client")
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
			}
		})

		It("generates a dummy app", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(8))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "tool", "testapi-cli", "main.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 16))
			_, err = gexec.Build(filepath.Join(testgenPackagePath, "tool", "testapi-cli"))
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("with an action with an integer parameter with no default value", func() {
		BeforeEach(func() {
			codegen.TempCount = 0
			design.Design = &design.APIDefinition{
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
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
									},
								},
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
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

		It("generates the correct command flag initialization code", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(9))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "tool", "cli", "commands.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 3))
			Ω(content).Should(ContainSubstring("var param int"))
			Ω(content).Should(ContainSubstring("var time_ string"))
			Ω(content).Should(ContainSubstring(".Flags()"))
			_, err = gexec.Build(filepath.Join(testgenPackagePath, "tool", "testapi-cli"))
			Ω(err).ShouldNot(HaveOccurred())

		})

		Context("with an action with a multiline description", func() {
			const multiline = "multi\nline"

			BeforeEach(func() {
				design.Design.Resources["foo"].Actions["show"].Description = multiline
			})

			It("properly escapes the multi-line string used in the short description", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(9))
				content, err := ioutil.ReadFile(filepath.Join(outDir, "tool", "cli", "commands.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(content)).Should(ContainSubstring(multiline))
			})
		})

		Context("with an action with a description containing backticks", func() {
			const pre = "pre"
			const post = "post"

			BeforeEach(func() {
				design.Design.Resources["foo"].Actions["show"].Description = pre + "`" + post
			})

			It("properly escapes the multi-line string used in the short description", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(9))
				content, err := ioutil.ReadFile(filepath.Join(outDir, "tool", "cli", "commands.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(content)).Should(ContainSubstring(pre + "` + \"`\" + `" + post))
			})
		})
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

		It("generates registers the signer flags from main", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(9))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "tool", "testapi-cli", "main.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("jwt1Signer := newJWT1Signer()"))
			Ω(content).Should(ContainSubstring("c.SetJWT1Signer(jwt1Signer)"))
		})
	})
})
