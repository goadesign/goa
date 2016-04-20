package genapp_test

import (
	"io/ioutil"
	"os"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContextsWriter", func() {
	var writer *genapp.ContextsWriter
	var filename string
	var workspace *codegen.Workspace

	JustBeforeEach(func() {
		var err error
		workspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		pkg, err := workspace.NewPackage("contexts")
		Ω(err).ShouldNot(HaveOccurred())
		src := pkg.CreateSourceFile("test.go")
		filename = src.Abs()
		writer, err = genapp.NewContextsWriter(filename)
		Ω(err).ShouldNot(HaveOccurred())
		codegen.TempCount = 0
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("correctly configured", func() {
		var f *os.File
		BeforeEach(func() {
			f, _ = ioutil.TempFile("", "")
			filename = f.Name()
		})

		AfterEach(func() {
			os.Remove(filename)
		})

		Context("with data", func() {
			var params, headers *design.AttributeDefinition
			var payload *design.UserTypeDefinition
			var responses map[string]*design.ResponseDefinition

			var data *genapp.ContextTemplateData

			BeforeEach(func() {
				params = nil
				headers = nil
				payload = nil
				responses = nil
				data = nil
			})

			JustBeforeEach(func() {
				data = &genapp.ContextTemplateData{
					Name:         "ListBottleContext",
					ResourceName: "bottles",
					ActionName:   "list",
					Params:       params,
					Payload:      payload,
					Headers:      headers,
					Responses:    responses,
					API:          design.Design,
					DefaultPkg:   "",
				}
			})

			Context("with simple data", func() {
				It("writes the simple contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(emptyContext))
					Ω(written).Should(ContainSubstring(emptyContextFactory))
				})
			})

			Context("with an integer param", func() {
				BeforeEach(func() {
					intParam := &design.AttributeDefinition{Type: design.Integer}
					dataType := design.Object{
						"param": intParam,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
				})

				It("writes the integer contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(intContext))
					Ω(written).Should(ContainSubstring(intContextFactory))
				})
			})

			Context("with a string param", func() {
				BeforeEach(func() {
					strParam := &design.AttributeDefinition{Type: design.String}
					dataType := design.Object{
						"param": strParam,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(strContext))
					Ω(written).Should(ContainSubstring(strContextFactory))
				})
			})

			Context("with a number param", func() {
				BeforeEach(func() {
					numParam := &design.AttributeDefinition{Type: design.Number}
					dataType := design.Object{
						"param": numParam,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(numContext))
					Ω(written).Should(ContainSubstring(numContextFactory))
				})
			})

			Context("with a boolean param", func() {
				BeforeEach(func() {
					boolParam := &design.AttributeDefinition{Type: design.Boolean}
					dataType := design.Object{
						"param": boolParam,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(boolContext))
					Ω(written).Should(ContainSubstring(boolContextFactory))
				})
			})

			Context("with an array param", func() {
				BeforeEach(func() {
					str := &design.AttributeDefinition{Type: design.String}
					arrayParam := &design.AttributeDefinition{
						Type: &design.Array{ElemType: str},
					}
					dataType := design.Object{
						"param": arrayParam,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(arrayContext))
					Ω(written).Should(ContainSubstring(arrayContextFactory))
				})
			})

			Context("with an integer array param", func() {
				BeforeEach(func() {
					i := &design.AttributeDefinition{Type: design.Integer}
					intArrayParam := &design.AttributeDefinition{
						Type: &design.Array{ElemType: i},
					}
					dataType := design.Object{
						"param": intArrayParam,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(intArrayContext))
					Ω(written).Should(ContainSubstring(intArrayContextFactory))
				})
			})

			Context("with an param using a reserved keyword as name", func() {
				BeforeEach(func() {
					intParam := &design.AttributeDefinition{Type: design.Integer}
					dataType := design.Object{
						"int": intParam,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(resContext))
					Ω(written).Should(ContainSubstring(resContextFactory))
				})
			})

			Context("with a required param", func() {
				BeforeEach(func() {
					intParam := &design.AttributeDefinition{Type: design.Integer}
					dataType := design.Object{
						"int": intParam,
					}
					required := &dslengine.ValidationDefinition{
						Required: []string{"int"},
					}
					params = &design.AttributeDefinition{
						Type:       dataType,
						Validation: required,
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(requiredContext))
					Ω(written).Should(ContainSubstring(requiredContextFactory))
				})
			})

			Context("with a simple payload", func() {
				BeforeEach(func() {
					payload = &design.UserTypeDefinition{
						AttributeDefinition: &design.AttributeDefinition{Type: design.String},
						TypeName:            "ListBottlePayload",
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(payloadContext))
					Ω(written).Should(ContainSubstring(payloadContextFactory))
				})
			})

			Context("with a object payload", func() {
				BeforeEach(func() {
					intParam := &design.AttributeDefinition{Type: design.Integer}
					strParam := &design.AttributeDefinition{Type: design.String}
					dataType := design.Object{
						"int": intParam,
						"str": strParam,
					}
					required := &dslengine.ValidationDefinition{
						Required: []string{"int"},
					}
					payload = &design.UserTypeDefinition{
						AttributeDefinition: &design.AttributeDefinition{
							Type:       dataType,
							Validation: required,
						},
						TypeName: "ListBottlePayload",
					}
				})

				It("writes the contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(payloadObjContext))
				})

				var _ = Describe("IterateResponses", func() {
					var resps []*design.ResponseDefinition
					var testIt = func(r *design.ResponseDefinition) error {
						resps = append(resps, r)
						return nil
					}
					Context("with responses", func() {
						BeforeEach(func() {
							responses = map[string]*design.ResponseDefinition{
								"OK":      {Status: 200},
								"Created": {Status: 201},
							}
						})
						It("iterates responses in order", func() {
							data.IterateResponses(testIt)
							Ω(resps).Should(Equal([]*design.ResponseDefinition{
								responses["OK"],
								responses["Created"],
							}))
						})
					})
				})
			})
		})
	})
})

var _ = Describe("ControllersWriter", func() {
	var writer *genapp.ControllersWriter
	var workspace *codegen.Workspace
	var filename string

	BeforeEach(func() {
		var err error
		workspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		pkg, err := workspace.NewPackage("controllers")
		Ω(err).ShouldNot(HaveOccurred())
		src := pkg.CreateSourceFile("test.go")
		filename = src.Abs()
	})

	JustBeforeEach(func() {
		var err error
		writer, err = genapp.NewControllersWriter(filename)
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("correctly configured", func() {
		BeforeEach(func() {
			os.Create(filename)
		})

		Context("with data", func() {
			var actions, verbs, paths, contexts, unmarshals []string
			var payloads []*design.UserTypeDefinition
			var encoders, decoders []*genapp.EncoderTemplateData
			var origins []*design.CORSDefinition

			var data []*genapp.ControllerTemplateData

			BeforeEach(func() {
				actions = nil
				verbs = nil
				paths = nil
				contexts = nil
				unmarshals = nil
				payloads = nil
				encoders = nil
				decoders = nil
				origins = nil
			})

			JustBeforeEach(func() {
				codegen.TempCount = 0
				api := &design.APIDefinition{}
				d := &genapp.ControllerTemplateData{
					Resource: "Bottles",
					Origins:  origins,
				}
				as := make([]map[string]interface{}, len(actions))
				for i, a := range actions {
					var unmarshal string
					var payload *design.UserTypeDefinition
					if i < len(unmarshals) {
						unmarshal = unmarshals[i]
					}
					if i < len(payloads) {
						payload = payloads[i]
					}
					as[i] = map[string]interface{}{
						"Name": a,
						"Routes": []*design.RouteDefinition{
							{
								Verb: verbs[i],
								Path: paths[i],
							}},
						"Context":   contexts[i],
						"Unmarshal": unmarshal,
						"Payload":   payload,
					}
				}
				if len(as) > 0 {
					d.API = api
					d.Actions = as
					d.Encoders = encoders
					d.Decoders = decoders
					data = []*genapp.ControllerTemplateData{d}
				} else {
					data = nil
				}
			})

			Context("with missing data", func() {
				It("returns an empty string", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).Should(BeEmpty())
				})
			})

			Context("with a simple controller", func() {
				BeforeEach(func() {
					actions = []string{"List"}
					verbs = []string{"GET"}
					paths = []string{"/accounts/:accountID/bottles"}
					contexts = []string{"ListBottleContext"}
				})

				It("writes the controller code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(simpleController))
					Ω(written).Should(ContainSubstring(simpleMount))
				})
			})

			Context("with actions that take a payload", func() {
				BeforeEach(func() {
					actions = []string{"List"}
					verbs = []string{"GET"}
					paths = []string{"/accounts/:accountID/bottles"}
					contexts = []string{"ListBottleContext"}
					unmarshals = []string{"unmarshalListBottlePayload"}
					payloads = []*design.UserTypeDefinition{
						{
							TypeName: "ListBottlePayload",
							AttributeDefinition: &design.AttributeDefinition{
								Type: design.Object{
									"id": &design.AttributeDefinition{
										Type: design.String,
									},
								},
							},
						},
					}
				})

				It("writes the payload unmarshal function", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).Should(ContainSubstring(payloadNoValidationsObjUnmarshal))
				})
			})
			Context("with actions that take a payload with a required validation", func() {
				BeforeEach(func() {
					actions = []string{"List"}
					required := &dslengine.ValidationDefinition{
						Required: []string{"id"},
					}
					verbs = []string{"GET"}
					paths = []string{"/accounts/:accountID/bottles"}
					contexts = []string{"ListBottleContext"}
					unmarshals = []string{"unmarshalListBottlePayload"}
					payloads = []*design.UserTypeDefinition{
						{
							TypeName: "ListBottlePayload",
							AttributeDefinition: &design.AttributeDefinition{
								Type: design.Object{
									"id": &design.AttributeDefinition{
										Type: design.String,
									},
								},
								Validation: required,
							},
						},
					}
				})

				It("writes the payload unmarshal function", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).Should(ContainSubstring(payloadObjUnmarshal))
				})
			})

			Context("with multiple controllers", func() {
				BeforeEach(func() {
					actions = []string{"List", "Show"}
					verbs = []string{"GET", "GET"}
					paths = []string{"/accounts/:accountID/bottles", "/accounts/:accountID/bottles/:id"}
					contexts = []string{"ListBottleContext", "ShowBottleContext"}
				})

				It("writes the controllers code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(multiController))
					Ω(written).Should(ContainSubstring(multiMount))
				})
			})

			Context("with encoder and decoder maps", func() {
				BeforeEach(func() {
					actions = []string{"List"}
					verbs = []string{"GET"}
					paths = []string{"/accounts/:accountID/bottles"}
					contexts = []string{"ListBottleContext"}
					encoders = []*genapp.EncoderTemplateData{
						{
							PackageName: "goa",
							Function:    "NewEncoder",
							MIMETypes:   []string{"application/json"},
						},
					}
					decoders = []*genapp.EncoderTemplateData{
						{
							PackageName: "goa",
							Function:    "NewDecoder",
							MIMETypes:   []string{"application/json"},
						},
					}
				})

				It("writes the controllers code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(encoderController))
				})
			})

			Context("with multiple origins", func() {
				BeforeEach(func() {
					actions = []string{"List"}
					verbs = []string{"GET"}
					paths = []string{"/accounts"}
					contexts = []string{"ListBottleContext"}
					origins = []*design.CORSDefinition{
						{
							Origin:      "here.example.com",
							Headers:     []string{"X-One", "X-Two"},
							Methods:     []string{"GET", "POST"},
							Exposed:     []string{"X-Three"},
							Credentials: true,
						},
						{
							Origin:  "there.example.com",
							Headers: []string{"*"},
							Methods: []string{"*"},
						},
					}

				})

				It("writes the controller code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(originsIntegration))
					Ω(written).Should(ContainSubstring(originsHandler))
				})
			})

		})
	})
})

var _ = Describe("HrefWriter", func() {
	var writer *genapp.ResourcesWriter
	var workspace *codegen.Workspace
	var filename string

	BeforeEach(func() {
		var err error
		workspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		pkg, err := workspace.NewPackage("controllers")
		Ω(err).ShouldNot(HaveOccurred())
		src := pkg.CreateSourceFile("test.go")
		filename = src.Abs()
	})

	JustBeforeEach(func() {
		var err error
		writer, err = genapp.NewResourcesWriter(filename)
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("correctly configured", func() {
		Context("with data", func() {
			var canoTemplate string
			var canoParams []string
			var mediaType *design.MediaTypeDefinition

			var data *genapp.ResourceData

			BeforeEach(func() {
				mediaType = nil
				canoTemplate = ""
				canoParams = nil
				data = nil
			})

			JustBeforeEach(func() {
				data = &genapp.ResourceData{
					Name:              "Bottle",
					Identifier:        "vnd.acme.com/resources",
					Description:       "A bottle resource",
					Type:              mediaType,
					CanonicalTemplate: canoTemplate,
					CanonicalParams:   canoParams,
				}
			})

			Context("with missing resource type definition", func() {
				It("does not return an error", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			Context("with a string resource", func() {
				BeforeEach(func() {
					attDef := &design.AttributeDefinition{
						Type: design.String,
					}
					mediaType = &design.MediaTypeDefinition{
						UserTypeDefinition: &design.UserTypeDefinition{
							AttributeDefinition: attDef,
							TypeName:            "Bottle",
						},
					}
				})
				It("does not return an error", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			Context("with a user type resource", func() {
				BeforeEach(func() {
					intParam := &design.AttributeDefinition{Type: design.Integer}
					strParam := &design.AttributeDefinition{Type: design.String}
					dataType := design.Object{
						"int": intParam,
						"str": strParam,
					}
					attDef := &design.AttributeDefinition{
						Type: dataType,
					}
					mediaType = &design.MediaTypeDefinition{
						UserTypeDefinition: &design.UserTypeDefinition{
							AttributeDefinition: attDef,
							TypeName:            "Bottle",
						},
					}
				})

				Context("and a canonical action", func() {
					BeforeEach(func() {
						canoTemplate = "/bottles/%v"
						canoParams = []string{"id"}
					})

					It("writes the href method", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(simpleResourceHref))
					})
				})
			})
		})
	})
})

const (
	emptyContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
}
`

	emptyContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	return &rctx, err
}
`

	intContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Param *int
}
`

	intContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		rawParam := paramParam[0]
		if param, err2 := strconv.Atoi(rawParam); err2 == nil {
			tmp2 := param
			tmp1 := &tmp2
			rctx.Param = tmp1
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "integer"))
		}
	}
	return &rctx, err
}
`

	strContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Param *string
}
`

	strContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		rawParam := paramParam[0]
		rctx.Param = &rawParam
	}
	return &rctx, err
}
`

	numContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Param *float64
}
`

	numContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		rawParam := paramParam[0]
		if param, err2 := strconv.ParseFloat(rawParam, 64); err2 == nil {
			tmp1 := &param
			rctx.Param = tmp1
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "number"))
		}
	}
	return &rctx, err
}
`
	boolContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Param *bool
}
`

	boolContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		rawParam := paramParam[0]
		if param, err2 := strconv.ParseBool(rawParam); err2 == nil {
			tmp1 := &param
			rctx.Param = tmp1
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "boolean"))
		}
	}
	return &rctx, err
}
`

	arrayContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Param []string
}
`

	arrayContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		var params []string
		for _, rawParam := range paramParam {
			elemsParam := strings.Split(rawParam, ",")
			params = elemsParam
			rctx.Param = append(rctx.Param, params...)
		}
	}
	return &rctx, err
}
`

	intArrayContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Param []int
}
`

	intArrayContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		var params []int
		for _, rawParam := range paramParam {
			elemsParam := strings.Split(rawParam, ",")
			elemsParam2 := make([]int, len(elemsParam))
			for i, rawElem := range elemsParam {
				if elem, err2 := strconv.Atoi(rawElem); err2 == nil {
					elemsParam2[i] = elem
				} else {
					err = goa.MergeErrors(err, goa.InvalidParamTypeError("elem", rawElem, "integer"))
				}
			}
			params = elemsParam2
			rctx.Param = append(rctx.Param, params...)
		}
	}
	return &rctx, err
}
`

	resContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Int *int
}
`

	resContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramInt := req.Params["int"]
	if len(paramInt) > 0 {
		rawInt := paramInt[0]
		if int_, err2 := strconv.Atoi(rawInt); err2 == nil {
			tmp2 := int_
			tmp1 := &tmp2
			rctx.Int = tmp1
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("int", rawInt, "integer"))
		}
	}
	return &rctx, err
}
`

	requiredContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Int int
}
`

	requiredContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	paramInt := req.Params["int"]
	if len(paramInt) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("int"))
	} else {
		rawInt := paramInt[0]
		if int_, err2 := strconv.Atoi(rawInt); err2 == nil {
			rctx.Int = int_
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("int", rawInt, "integer"))
		}
	}
	return &rctx, err
}
`

	payloadContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Payload ListBottlePayload
}
`

	payloadContextFactory = `
func NewListBottleContext(ctx context.Context, service *goa.Service) (*ListBottleContext, error) {
	var err error
	req := goa.ContextRequest(ctx)
	rctx := ListBottleContext{Context: ctx, ResponseData: goa.ContextResponse(ctx), RequestData: req, Service: service}
	return &rctx, err
}
`
	payloadObjContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Service *goa.Service
	Payload *ListBottlePayload
}
`

	payloadObjUnmarshal = `
func unmarshalListBottlePayload(ctx context.Context, service *goa.Service, req *http.Request) error {
	payload := &listBottlePayload{}
	if err := service.DecodeRequest(req, payload); err != nil {
		return err
	}
	if err := payload.Validate(); err != nil {
		return err
	}
	goa.ContextRequest(ctx).Payload = payload.Publicize()
	return nil
}
`
	payloadNoValidationsObjUnmarshal = `
func unmarshalListBottlePayload(ctx context.Context, service *goa.Service, req *http.Request) error {
	payload := &listBottlePayload{}
	if err := service.DecodeRequest(req, payload); err != nil {
		return err
	}
	goa.ContextRequest(ctx).Payload = payload.Publicize()
	return nil
}
`

	simpleController = `// BottlesController is the controller interface for the Bottles actions.
type BottlesController interface {
	goa.Muxer
	List(*ListBottleContext) error
}
`

	originsIntegration = `}
	h = handleBottlesOrigin(h)
	service.Mux.Handle`

	originsHandler = `// handleBottlesOrigin applies the CORS response headers corresponding to the origin.
func handleBottlesOrigin(h goa.Handler) goa.Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		origin := req.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			return h(ctx, rw, req)
		}
		if cors.MatchOrigin(origin, "here.example.com") {
			ctx = goa.WithLogContext(ctx, "origin", origin)
			rw.Header().Set("Access-Control-Allow-Origin", "here.example.com")
			rw.Header().Set("Vary", "Origin")
			rw.Header().Set("Access-Control-Expose-Headers", "X-Three")
			rw.Header().Set("Access-Control-Allow-Credentials", "true")
			if acrm := req.Header.Get("Access-Control-Request-Method"); acrm != "" {
				// We are handling a preflight request
				rw.Header().Set("Access-Control-Allow-Methods", "GET, POST")
				rw.Header().Set("Access-Control-Allow-Headers", "X-One, X-Two")
			}
			return h(ctx, rw, req)
		}
		if cors.MatchOrigin(origin, "there.example.com") {
			ctx = goa.WithLogContext(ctx, "origin", origin)
			rw.Header().Set("Access-Control-Allow-Origin", "there.example.com")
			rw.Header().Set("Vary", "Origin")
			rw.Header().Set("Access-Control-Allow-Credentials", "false")
			if acrm := req.Header.Get("Access-Control-Request-Method"); acrm != "" {
				// We are handling a preflight request
				rw.Header().Set("Access-Control-Allow-Methods", "*")
				rw.Header().Set("Access-Control-Allow-Headers", "*")
			}
			return h(ctx, rw, req)
		}

		return h(ctx, rw, req)
	}
}
`

	encoderController = `
// MountBottlesController "mounts" a Bottles resource controller on the given service.
func MountBottlesController(service *goa.Service, ctrl BottlesController) {
	initService(service)
	var h goa.Handler

	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rctx, err := NewListBottleContext(ctx, service)
		if err != nil {
			return err
		}
		return ctrl.List(rctx)
	}
	service.Mux.Handle("GET", "/accounts/:accountID/bottles", ctrl.MuxHandler("List", h, nil))
	service.LogInfo("mount", "ctrl", "Bottles", "action", "List", "route", "GET /accounts/:accountID/bottles")
}
`

	simpleMount = `func MountBottlesController(service *goa.Service, ctrl BottlesController) {
	initService(service)
	var h goa.Handler

	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rctx, err := NewListBottleContext(ctx, service)
		if err != nil {
			return err
		}
		return ctrl.List(rctx)
	}
	service.Mux.Handle("GET", "/accounts/:accountID/bottles", ctrl.MuxHandler("List", h, nil))
	service.LogInfo("mount", "ctrl", "Bottles", "action", "List", "route", "GET /accounts/:accountID/bottles")
}
`

	multiController = `// BottlesController is the controller interface for the Bottles actions.
type BottlesController interface {
	goa.Muxer
	List(*ListBottleContext) error
	Show(*ShowBottleContext) error
}
`

	multiMount = `func MountBottlesController(service *goa.Service, ctrl BottlesController) {
	initService(service)
	var h goa.Handler

	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rctx, err := NewListBottleContext(ctx, service)
		if err != nil {
			return err
		}
		return ctrl.List(rctx)
	}
	service.Mux.Handle("GET", "/accounts/:accountID/bottles", ctrl.MuxHandler("List", h, nil))
	service.LogInfo("mount", "ctrl", "Bottles", "action", "List", "route", "GET /accounts/:accountID/bottles")

	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rctx, err := NewShowBottleContext(ctx, service)
		if err != nil {
			return err
		}
		return ctrl.Show(rctx)
	}
	service.Mux.Handle("GET", "/accounts/:accountID/bottles/:id", ctrl.MuxHandler("Show", h, nil))
	service.LogInfo("mount", "ctrl", "Bottles", "action", "Show", "route", "GET /accounts/:accountID/bottles/:id")
}
`

	simpleResourceHref = `func BottleHref(id interface{}) string {
	return fmt.Sprintf("/bottles/%v", id)
}
`
)
