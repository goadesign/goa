package genapp_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/gen_app"
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
				var version *design.APIVersionDefinition
				if design.Design != nil {
					version = design.Design.APIVersionDefinition
				}
				data = &genapp.ContextTemplateData{
					Name:         "ListBottleContext",
					ResourceName: "bottles",
					ActionName:   "list",
					Params:       params,
					Payload:      payload,
					Headers:      headers,
					Responses:    responses,
					API:          design.Design,
					Version:      version,
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

var _ = Describe("ControllersWriter", func() {
	var writer *genapp.ControllersWriter
	var filename string

	JustBeforeEach(func() {
		var err error
		writer, err = genapp.NewControllersWriter(filename)
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
			var actions, verbs, paths, contexts []string

			var data []*genapp.ControllerTemplateData

			BeforeEach(func() {
				actions = nil
				verbs = nil
				paths = nil
				contexts = nil
			})

			JustBeforeEach(func() {
				d := &genapp.ControllerTemplateData{
					Resource: "Bottles",
				}
				as := make([]map[string]interface{}, len(actions))
				for i, a := range actions {
					as[i] = map[string]interface{}{
						"Name": a,
						"Routes": []*design.RouteDefinition{
							&design.RouteDefinition{
								Verb: verbs[i],
								Path: paths[i],
							}},
						"Context": contexts[i],
					}
				}
				if len(as) > 0 {
					d.Actions = as
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
					actions = []string{"list"}
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

			Context("with multiple controllers", func() {
				BeforeEach(func() {
					actions = []string{"list", "show"}
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
		})
	})
})

var _ = Describe("HrefWriter", func() {
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
	*goa.Context
}
`

	emptyContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	return &ctx, err
}
`

	intContext = `
type ListBottleContext struct {
	*goa.Context
	Param int

	HasParam bool
}
`

	intContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, ok := c.Get("param")
	if ok {
		if param, err2 := strconv.Atoi(rawParam); err2 == nil {
			ctx.Param = int(param)
		} else {
			err = goa.InvalidParamTypeError("param", rawParam, "integer", err)
		}
		ctx.HasParam = true
	}
	return &ctx, err
}
`

	strContext = `
type ListBottleContext struct {
	*goa.Context
	Param string

	HasParam bool
}
`

	strContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, ok := c.Get("param")
	if ok {
		ctx.Param = rawParam
		ctx.HasParam = true
	}
	return &ctx, err
}
`

	numContext = `
type ListBottleContext struct {
	*goa.Context
	Param float64

	HasParam bool
}
`

	numContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, ok := c.Get("param")
	if ok {
		if param, err2 := strconv.ParseFloat(rawParam, 64); err2 == nil {
			ctx.Param = param
		} else {
			err = goa.InvalidParamTypeError("param", rawParam, "number", err)
		}
		ctx.HasParam = true
	}
	return &ctx, err
}
`
	boolContext = `
type ListBottleContext struct {
	*goa.Context
	Param bool

	HasParam bool
}
`

	boolContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, ok := c.Get("param")
	if ok {
		if param, err2 := strconv.ParseBool(rawParam); err2 == nil {
			ctx.Param = param
		} else {
			err = goa.InvalidParamTypeError("param", rawParam, "boolean", err)
		}
		ctx.HasParam = true
	}
	return &ctx, err
}
`

	arrayContext = `
type ListBottleContext struct {
	*goa.Context
	Param []string

	HasParam bool
}
`

	arrayContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, ok := c.Get("param")
	if ok {
		elemsParam := strings.Split(rawParam, ",")
		ctx.Param = elemsParam
		ctx.HasParam = true
	}
	return &ctx, err
}
`

	intArrayContext = `
type ListBottleContext struct {
	*goa.Context
	Param []int

	HasParam bool
}
`

	intArrayContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, ok := c.Get("param")
	if ok {
		elemsParam := strings.Split(rawParam, ",")
		elemsParam2 := make([]int, len(elemsParam))
		for i, rawElem := range elemsParam {
			if elem, err2 := strconv.Atoi(rawElem); err2 == nil {
				elemsParam2[i] = int(elem)
			} else {
				err = goa.InvalidParamTypeError("elem", rawElem, "integer", err)
			}
		}
		ctx.Param = elemsParam2
		ctx.HasParam = true
	}
	return &ctx, err
}
`

	resContext = `
type ListBottleContext struct {
	*goa.Context
	Int int

	HasInt bool
}
`

	resContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawInt, ok := c.Get("int")
	if ok {
		if int_, err2 := strconv.Atoi(rawInt); err2 == nil {
			ctx.Int = int(int_)
		} else {
			err = goa.InvalidParamTypeError("int", rawInt, "integer", err)
		}
		ctx.HasInt = true
	}
	return &ctx, err
}
`

	requiredContext = `
type ListBottleContext struct {
	*goa.Context
	Int int
}
`

	requiredContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawInt, ok := c.Get("int")
	if !ok {
		err = goa.MissingParamError("int", err)
	} else {
		if int_, err2 := strconv.Atoi(rawInt); err2 == nil {
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
	*goa.Context
	Payload ListBottlePayload
}
`

	payloadContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	p, err := NewListBottlePayload(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}
`
	payloadObjContext = `
type ListBottleContext struct {
	*goa.Context
	Payload *ListBottlePayload
}
`

	payloadObjContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	p, err := NewListBottlePayload(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}
`

	simpleController = `// BottlesController is the controller interface for the Bottles actions.
type BottlesController interface {
	goa.Controller
	list(*ListBottleContext) error
}
`

	simpleMount = `func MountBottlesController(service goa.Service, ctrl BottlesController) {
	router := service.HTTPHandler().(*httprouter.Router)
	var h goa.Handler
	h = func(c *goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.list(ctx)
	}
	router.Handle("GET", "/accounts/:accountID/bottles", ctrl.NewHTTPRouterHandle("list", h))
	service.Info("mount", "ctrl", "Bottles", "action", "list", "route", "GET /accounts/:accountID/bottles")
}
`

	multiController = `// BottlesController is the controller interface for the Bottles actions.
type BottlesController interface {
	goa.Controller
	list(*ListBottleContext) error
	show(*ShowBottleContext) error
}
`

	multiMount = `func MountBottlesController(service goa.Service, ctrl BottlesController) {
	router := service.HTTPHandler().(*httprouter.Router)
	var h goa.Handler
	h = func(c *goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.list(ctx)
	}
	router.Handle("GET", "/accounts/:accountID/bottles", ctrl.NewHTTPRouterHandle("list", h))
	service.Info("mount", "ctrl", "Bottles", "action", "list", "route", "GET /accounts/:accountID/bottles")
	h = func(c *goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		return ctrl.show(ctx)
	}
	router.Handle("GET", "/accounts/:accountID/bottles/:id", ctrl.NewHTTPRouterHandle("show", h))
	service.Info("mount", "ctrl", "Bottles", "action", "show", "route", "GET /accounts/:accountID/bottles/:id")
}
`

	simpleResourceHref = `func BottleHref(id interface{}) string {
	return fmt.Sprintf("/bottles/%v", id)
}
`
)
