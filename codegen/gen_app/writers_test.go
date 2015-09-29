package genapp_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/codegen/gen_app"
	"github.com/raphael/goa/design"
)

var _ = Describe("ContextsWriter", func() {
	var writer *genapp.ContextsWriter
	var filename string

	JustBeforeEach(func() {
		var err error
		writer, err = genapp.NewContextsWriter(filename)
		Ω(err).ShouldNot(HaveOccurred())
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
			var mediaTypes map[string]*design.MediaTypeDefinition

			var data *genapp.ContextTemplateData

			BeforeEach(func() {
				params = nil
				headers = nil
				payload = nil
				responses = nil
				mediaTypes = nil
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
					MediaTypes:   mediaTypes,
				}
			})

			Context("with simple data", func() {
				It("writes the contexts code", func() {
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

				It("writes the contexts code", func() {
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
					required := design.RequiredValidationDefinition{
						Names: []string{"int"},
					}
					params = &design.AttributeDefinition{
						Type:        dataType,
						Validations: []design.ValidationDefinition{&required},
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
					required := design.RequiredValidationDefinition{
						Names: []string{"int"},
					}
					payload = &design.UserTypeDefinition{
						AttributeDefinition: &design.AttributeDefinition{
							Type:        dataType,
							Validations: []design.ValidationDefinition{&required},
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
					Ω(written).Should(ContainSubstring(payloadObjContextFactory))
				})
			})

		})
	})
})

var _ = Describe("HandlersWriter", func() {
	var writer *genapp.HandlersWriter
	var filename string

	JustBeforeEach(func() {
		var err error
		writer, err = genapp.NewHandlersWriter(filename)
		Ω(err).ShouldNot(HaveOccurred())
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
			var actions, verbs, paths, names, contexts []string

			var data []*genapp.HandlerTemplateData

			BeforeEach(func() {
				actions = nil
				verbs = nil
				paths = nil
				names = nil
				contexts = nil
			})

			JustBeforeEach(func() {
				data = make([]*genapp.HandlerTemplateData, len(actions))
				for i := 0; i < len(actions); i++ {
					e := &genapp.HandlerTemplateData{
						Resource: "bottles",
						Action:   actions[i],
						Verb:     verbs[i],
						Path:     paths[i],
						Name:     names[i],
						Context:  contexts[i],
					}
					data[i] = e
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

			Context("with a simple handler", func() {
				BeforeEach(func() {
					actions = []string{"list"}
					verbs = []string{"GET"}
					paths = []string{"/accounts/:accountID/bottles"}
					names = []string{"listBottlesHandler"}
					contexts = []string{"ListBottleContext"}
				})

				It("writes the handlers code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(simpleHandler))
					Ω(written).Should(ContainSubstring(simpleInit))
				})
			})

			Context("with multiple handlers", func() {
				BeforeEach(func() {
					actions = []string{"list", "show"}
					verbs = []string{"GET", "GET"}
					paths = []string{"/accounts/:accountID/bottles", "/accounts/:accountID/bottles/:id"}
					names = []string{"listBottlesHandler", "showBottlesHandler"}
					contexts = []string{"ListBottleContext", "ShowBottleContext"}
				})

				It("writes the handlers code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(multiHandler1))
					Ω(written).Should(ContainSubstring(multiHandler2))
					Ω(written).Should(ContainSubstring(multiInit))
				})
			})
		})
	})
})

var _ = Describe("ResourceWriter", func() {
	var writer *genapp.ResourcesWriter
	var filename string

	JustBeforeEach(func() {
		var err error
		writer, err = genapp.NewResourcesWriter(filename)
		Ω(err).ShouldNot(HaveOccurred())
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
			var canoTemplate string
			var canoParams []string
			var userType *design.UserTypeDefinition

			var data *genapp.ResourceTemplateData

			BeforeEach(func() {
				userType = nil
				canoTemplate = ""
				canoParams = nil
				data = nil
			})

			JustBeforeEach(func() {
				data = &genapp.ResourceTemplateData{
					Name:              "Bottle",
					Identifier:        "vnd.acme.com/resources",
					Description:       "A bottle resource",
					Type:              userType,
					CanonicalTemplate: canoTemplate,
					CanonicalParams:   canoParams,
				}
			})

			Context("with missing resource type definition", func() {
				It("returns an error", func() {
					err := writer.Execute(data)
					Ω(err).Should(HaveOccurred())
				})
			})

			Context("with a string resource", func() {
				BeforeEach(func() {
					attDef := &design.AttributeDefinition{
						Type: design.String,
					}
					userType = &design.UserTypeDefinition{
						AttributeDefinition: attDef,
						TypeName:            "Bottle",
					}
				})
				It("writes the resources code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(stringResource))
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
					userType = &design.UserTypeDefinition{
						AttributeDefinition: attDef,
						TypeName:            "Bottle",
					}
				})

				It("writes the resources code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(simpleResource))
				})

				Context("and a canonical action", func() {
					BeforeEach(func() {
						canoTemplate = "/bottles/%s"
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
	goa.Context
}
`

	emptyContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	return &ctx, err
}
`

	intContext = `
type ListBottleContext struct {
	goa.Context
	Param int
}
`

	intContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	if param, err := strconv.Atoi(rawParam); err == nil {
		ctx.Param = int(param)
	} else {
		err = goa.InvalidParamTypeError("param", rawParam, "integer", err)
	}
	return &ctx, err
}
`

	strContext = `
type ListBottleContext struct {
	goa.Context
	Param string
}
`

	strContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	ctx.Param = rawParam
	return &ctx, err
}
`

	numContext = `
type ListBottleContext struct {
	goa.Context
	Param float64
}
`

	numContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	if param, err := strconv.ParseFloat(rawParam, 64); err == nil {
		ctx.Param = param
	} else {
		err = goa.InvalidParamTypeError("param", rawParam, "number", err)
	}
	return &ctx, err
}
`
	boolContext = `
type ListBottleContext struct {
	goa.Context
	Param bool
}
`

	boolContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	if param, err := strconv.ParseBool(rawParam); err == nil {
		ctx.Param = param
	} else {
		err = goa.InvalidParamTypeError("param", rawParam, "boolean", err)
	}
	return &ctx, err
}
`

	arrayContext = `
type ListBottleContext struct {
	goa.Context
	Param []string
}
`

	arrayContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	elemsParam := strings.Split(rawParam, ",")
	ctx.Param = elemsParam
	return &ctx, err
}
`

	intArrayContext = `
type ListBottleContext struct {
	goa.Context
	Param []int
}
`

	intArrayContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	elemsParam := strings.Split(rawParam, ",")
	elemsParam2 := make([]int, len(elemsParam))
	for i, rawElem := range elemsParam {
		if elem, err := strconv.Atoi(rawElem); err == nil {
			elemsParam2[i] = int(elem)
		} else {
			err = goa.InvalidParamTypeError("elem", rawElem, "integer", err)
		}
	}
	ctx.Param = elemsParam
	return &ctx, err
}
`

	resContext = `
type ListBottleContext struct {
	goa.Context
	Int int
}
`

	resContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawInt, _ := c.Get("int")
	if int_, err := strconv.Atoi(rawInt); err == nil {
		ctx.Int = int(int_)
	} else {
		err = goa.InvalidParamTypeError("int", rawInt, "integer", err)
	}
	return &ctx, err
}
`

	requiredContext = `
type ListBottleContext struct {
	goa.Context
	Int int
}
`

	requiredContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawInt, ok := c.Get("int")
	if !ok {
		err = goa.MissingParamError("int", err)
	} else {
		if int_, err := strconv.Atoi(rawInt); err == nil {
			ctx.Int = int(int_)
		} else {
			err = goa.InvalidParamTypeError("int", rawInt, "integer", err)
		}
	}
	return &ctx, err
}
`

	payloadContext = `
type ListBottleContext struct {
	goa.Context
	payload *ListBottlePayload
}
`

	payloadContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	if payload := c.Payload(); payload != nil {
		p, err := NewListBottlePayload(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
	return &ctx, err
}
`
	payloadObjContext = `
type ListBottleContext struct {
	goa.Context
	payload *ListBottlePayload
}
`

	payloadObjContextFactory = `
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	if payload := c.Payload(); payload != nil {
		p, err := NewListBottlePayload(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
	return &ctx, err
}
`

	simpleHandler = `func listBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ListBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action list bottles, expected 'func(c *ListBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`
	simpleInit = `func init() {
	goa.RegisterHandlers(
		&goa.HandlerFactory{"bottles", "list", "GET", "/accounts/:accountID/bottles", listBottlesHandler},
	)
}
`
	multiHandler1 = `func listBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ListBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action list bottles, expected 'func(c *ListBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`
	multiHandler2 = `func showBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ShowBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action show bottles, expected 'func(c *ShowBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`
	multiInit = `func init() {
	goa.RegisterHandlers(
		&goa.HandlerFactory{"bottles", "list", "GET", "/accounts/:accountID/bottles", listBottlesHandler},
		&goa.HandlerFactory{"bottles", "show", "GET", "/accounts/:accountID/bottles/:id", showBottlesHandler},
	)
}
`
	stringResource = `type Bottle string`

	simpleResource = `type Bottle struct {
	Int int ` + "`" + `json:"int,omitempty"` + "`" + `
	Str string ` + "`" + `json:"str,omitempty"` + "`" + `
}
`
	simpleResourceHref = `func BottleHref(id string) string {
	return fmt.Sprintf("/bottles/%s", id)
}`
)
