package genclient_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_client"
	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	Context("with a required UUID header", func() {
		BeforeEach(func() {
			codegen.TempCount = 0
			o := design.Object{
				"header_name": &design.AttributeDefinition{Type: design.UUID},
			}
			design.Design = &design.APIDefinition{
				Name: "testapi",
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
								},
								Headers: &design.AttributeDefinition{
									Type: o,
									Validation: &dslengine.ValidationDefinition{
										Required: []string{"header_name"},
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

		It("generates header initialization code that compiles", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(9))
			c, err := ioutil.ReadFile(filepath.Join(outDir, "client", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())
			content := string(c)
			Ω(content).Should(ContainSubstring("header.Set(\"header_name\", tmp3)\n"))
		})
	})

	Context("with jsonapi like querystring params", func() {
		BeforeEach(func() {
			codegen.TempCount = 0
			o := design.Object{
				"fields[foo]": &design.AttributeDefinition{Type: design.String},
				"fields[bar]": &design.AttributeDefinition{Type: &design.Array{ElemType: &design.AttributeDefinition{Type: design.String}}},
				"fields[baz]": &design.AttributeDefinition{Type: &design.Array{ElemType: &design.AttributeDefinition{Type: design.Integer}}},
				"fields[bat]": &design.AttributeDefinition{Type: design.DateTime},
			}
			design.Design = &design.APIDefinition{
				Name: "testapi",
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
								},
								QueryParams: &design.AttributeDefinition{Type: o},
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

		It("generates param initialization code that uses the param name given in the design", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(9))
			c, err := ioutil.ReadFile(filepath.Join(outDir, "client", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())
			content := string(c)
			Ω(content).Should(ContainSubstring("func ShowFooPath("))
			Ω(content).Should(ContainSubstring(`values.Set("fields[foo]", *fieldsFoo)`))
			Ω(content).Should(ContainSubstring(`	for _, p := range fieldsBar {
		tmp3 := p
		values.Add("fields[bar]", tmp3)
	}
`))
			Ω(content).Should(ContainSubstring(`	for _, p := range fieldsBaz {
		tmp5 := strconv.Itoa(p)
		values.Add("fields[baz]", tmp5)
	}
`))
			Ω(content).Should(ContainSubstring(`	tmp4 := fieldsBat.Format(time.RFC3339)
		values.Set("fields[bat]", tmp4)`))
		})

		Context("with --notool", func() {
			BeforeEach(func() {
				os.Args = append(os.Args, "--notool")
			})

			It("should not return an error", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(5)) // 9, minus 4 entries for tool paths
			})
		})
	})

	Context("with an action with multiple routes", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				Name: "testapi",
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
									{
										Verb: "GET",
										Path: "/foo",
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
			showAct.Routes[1].Parent = showAct
		})

		It("generates Path function with unique names", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(9))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("func ShowFooPath("))
			Ω(strings.Count(string(content), "func ShowFooPath(")).Should(Equal(1))
			Ω(content).Should(ContainSubstring("func ShowFooPath2("))
			Ω(strings.Count(string(content), "func ShowFooPath2(")).Should(Equal(1))
		})

		Context("with a file server", func() {
			BeforeEach(func() {
				res := design.Design.Resources["foo"]
				res.FileServers = []*design.FileServerDefinition{
					{
						Parent:      res,
						FilePath:    "/swagger/swagger.json",
						RequestPath: "/swagger.json",
					},
				}
			})

			It("generates a Download function", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(9))
				content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "foo.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).Should(ContainSubstring("func (c *Client) DownloadSwaggerJSON("))
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
			Ω(files).Should(HaveLen(9))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "client.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("JWT1Signer goaclient.Signer"))
			Ω(content).Should(ContainSubstring("func (c *Client) SetJWT1Signer(signer goaclient.Signer) {\n	c.JWT1Signer = signer\n}"))
		})

		It("generates the Signer.Sign call from Action", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(9))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("c.JWT1Signer.Sign(req)"))
		})
	})

	Context("with an action with a user type payload", func() {
		BeforeEach(func() {
			codegen.TempCount = 0
			testType := &design.UserTypeDefinition{
				AttributeDefinition: &design.AttributeDefinition{
					Type: design.Object{
						"param": &design.AttributeDefinition{Type: design.Integer},
						"time":  &design.AttributeDefinition{Type: design.DateTime},
						"uuid":  &design.AttributeDefinition{Type: design.UUID},
					},
				},
				TypeName: "TestType",
			}
			design.Design = &design.APIDefinition{
				Types: map[string]*design.UserTypeDefinition{
					"TestType": testType,
				},
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
								},
								Payload: testType,
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

		It("generates the user type imports", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(9))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "user_types.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("uuid \"github.com/goadesign/goa/uuid\""))
		})
	})
})
