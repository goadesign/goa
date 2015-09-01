package main

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
)

var _ = Describe("ContextWriter", func() {
	var writer *ContextsWriter
	var filename string
	var newErr error

	JustBeforeEach(func() {
		writer, newErr = NewContextsWriter(filename)
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

		It("NewContextsWriter creates a writer", func() {
			Ω(newErr).ShouldNot(HaveOccurred())
		})

		Context("with data", func() {
			var params, headers *design.AttributeDefinition
			var payload *design.UserTypeDefinition
			var responses map[string]*design.ResponseDefinition

			var data *ContextData

			BeforeEach(func() {
				params = nil
				headers = nil
				payload = nil
				responses = nil
				data = nil
			})

			JustBeforeEach(func() {
				data = &ContextData{
					Name:         "ListBottleContext",
					ResourceName: "bottles",
					ActionName:   "list",
					Params:       params,
					Payload:      payload,
					Headers:      headers,
					Responses:    responses,
				}
			})

			Context("with simple data", func() {
				It("writes the contexts code", func() {
					err := writer.Write(data)
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
					err := writer.Write(data)
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
					err := writer.Write(data)
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
					err := writer.Write(data)
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
					err := writer.Write(data)
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
					err := writer.Write(data)
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
					err := writer.Write(data)
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
					err := writer.Write(data)
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
					err := writer.Write(data)
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
						Name:                "ListBottlePayload",
					}
				})

				It("writes the contexts code", func() {
					err := writer.Write(data)
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
						Name: "ListBottlePayload",
					}
				})

				It("writes the contexts code", func() {
					Ω(newErr).ShouldNot(HaveOccurred())
					err := writer.Write(data)
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

const emptyContext = `
type ListBottleContext struct {
	*goa.Context
}
`

const emptyContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	return &ctx, err
}
`

const intContext = `
type ListBottleContext struct {
	*goa.Context
	Param int
}
`

const intContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	if param, err := strconv.Atoi(rawParam); err == nil {
		ctx.Param = int(param)
	} else {
		err = goa.InvalidParamValue("param", rawParam, "integer", err)
	}
	return &ctx, err
}
`

const strContext = `
type ListBottleContext struct {
	*goa.Context
	Param string
}
`

const strContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	ctx.Param = rawParam
	return &ctx, err
}
`

const numContext = `
type ListBottleContext struct {
	*goa.Context
	Param float64
}
`

const numContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	if param, err := strconv.ParseFloat(rawParam, 64); err == nil {
		ctx.Param = param
	} else {
		err = goa.InvalidParamValue("param", rawParam, "number", err)
	}
	return &ctx, err
}
`
const boolContext = `
type ListBottleContext struct {
	*goa.Context
	Param bool
}
`

const boolContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	if param, err := strconv.ParseBool(rawParam); err == nil {
		ctx.Param = param
	} else {
		err = goa.InvalidParamValue("param", rawParam, "boolean", err)
	}
	return &ctx, err
}
`

const arrayContext = `
type ListBottleContext struct {
	*goa.Context
	Param []string
}
`

const arrayContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	elemsParam := strings.Split(rawParam, ",")
	ctx.Param = elemsParam
	return &ctx, err
}
`

const intArrayContext = `
type ListBottleContext struct {
	*goa.Context
	Param []int
}
`

const intArrayContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawParam, _ := c.Get("param")
	elemsParam := strings.Split(rawParam, ",")
	elemsParam2 := make([]int, len(elemsParam))
	for i, rawElem := range elemsParam {
		if elem, err := strconv.Atoi(rawElem); err == nil {
			elemsParam2[i] = int(elem)
		} else {
			err = goa.InvalidParamValue("elem", rawElem, "integer", err)
		}
	}
	ctx.Param = elemsParam2
	return &ctx, err
}
`

const resContext = `
type ListBottleContext struct {
	*goa.Context
	Int int
}
`

const resContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawInt, _ := c.Get("int")
	if int_, err := strconv.Atoi(rawInt); err == nil {
		ctx.Int = int(int_)
	} else {
		err = goa.InvalidParamValue("int", rawInt, "integer", err)
	}
	return &ctx, err
}
`

const requiredContext = `
type ListBottleContext struct {
	*goa.Context
	Int int
}
`

const requiredContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawInt, ok := c.Get("int")
	if !ok {
		err = goa.MissingParam("int", err)
	} else {
		if int_, err := strconv.Atoi(rawInt); err == nil {
			ctx.Int = int(int_)
		} else {
			err = goa.InvalidParamValue("int", rawInt, "integer", err)
		}
	}
	return &ctx, err
}
`

const payloadContext = `
type ListBottleContext struct {
	*goa.Context
	payload *ListBottlePayload
}
`

const payloadContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	var p *ListBottlePayload
	if err := c.Bind(&p); err != nil {
		return nil, err
	}
	ctx.Payload = &p
	return &ctx, err
}
`
const payloadObjContext = `
type ListBottleContext struct {
	*goa.Context
	payload *ListBottlePayload
}
`

const payloadObjContextFactory = `
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	var p *ListBottlePayload
	if err := c.Bind(&p); err != nil {
		return nil, err
	}
	ctx.Payload = &p
	return &ctx, err
}
`
