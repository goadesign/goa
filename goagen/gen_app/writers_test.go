package genapp_test

import (
	"io/ioutil"
	"os"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/design/apidsl"
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
			dslengine.Reset()
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

			Context("with a media type setting a ContentType", func() {
				var contentType = "application/json"

				BeforeEach(func() {
					mediaType := &design.MediaTypeDefinition{
						UserTypeDefinition: &design.UserTypeDefinition{
							AttributeDefinition: &design.AttributeDefinition{
								Type: design.Object{"foo": {Type: design.String}},
							},
						},
						Identifier:  "application/vnd.goa.test",
						ContentType: contentType,
					}
					defView := &design.ViewDefinition{
						AttributeDefinition: mediaType.AttributeDefinition,
						Name:                "default",
						Parent:              mediaType,
					}
					mediaType.Views = map[string]*design.ViewDefinition{"default": defView}
					design.Design = new(design.APIDefinition)
					design.Design.MediaTypes = map[string]*design.MediaTypeDefinition{
						design.CanonicalIdentifier(mediaType.Identifier): mediaType,
					}
					design.ProjectedMediaTypes = make(map[string]*design.MediaTypeDefinition)
					responses = map[string]*design.ResponseDefinition{"OK": {
						Name:      "OK",
						Status:    200,
						MediaType: mediaType.Identifier,
					}}
				})

				It("the generated code sets the Content-Type header", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(`ctx.ResponseData.Header().Set("Content-Type", "` + contentType + `")`))
				})
			})

			Context("with a collection media type", func() {
				BeforeEach(func() {
					elemType := &design.MediaTypeDefinition{
						UserTypeDefinition: &design.UserTypeDefinition{
							AttributeDefinition: &design.AttributeDefinition{
								Type: design.Object{"foo": {Type: design.String}},
							},
						},
						Identifier: "application/vnd.goa.test",
					}
					defView := &design.ViewDefinition{
						AttributeDefinition: elemType.AttributeDefinition,
						Name:                "default",
						Parent:              elemType,
					}
					elemType.Views = map[string]*design.ViewDefinition{"default": defView}
					design.Design = new(design.APIDefinition)
					design.Design.MediaTypes = map[string]*design.MediaTypeDefinition{
						design.CanonicalIdentifier(elemType.Identifier): elemType,
					}
					design.ProjectedMediaTypes = make(map[string]*design.MediaTypeDefinition)
					mediaType := apidsl.CollectionOf(elemType)
					dslengine.Execute(mediaType.DSL(), mediaType)
					responses = map[string]*design.ResponseDefinition{"OK": {
						Name:   "OK",
						Status: 200,
						Type:   mediaType,
					}}
				})

				It("the generated code sets the response to an empty collection if value is nil", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(`	if r == nil {
		r = Collection{}
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)`))
				})
			})

			Context("with an integer param", func() {
				var (
					intParam   *design.AttributeDefinition
					dataType   design.Object
					validation *dslengine.ValidationDefinition
				)

				BeforeEach(func() {
					intParam = &design.AttributeDefinition{Type: design.Integer}
					dataType = design.Object{
						"param": intParam,
					}
					validation = &dslengine.ValidationDefinition{}
					params = &design.AttributeDefinition{
						Type:       dataType,
						Validation: validation,
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

				Context("with a default value", func() {
					BeforeEach(func() {
						intParam.SetDefault(2)
					})

					It("writes the integer contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(intDefaultContext))
						Ω(written).Should(ContainSubstring(intDefaultContextFactory))
					})
				})

				Context("with required attribute", func() {
					BeforeEach(func() {
						validation.Required = []string{"param"}
					})

					It("writes the integer contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(intRequiredContext))
						Ω(written).Should(ContainSubstring(intRequiredContextFactory))
					})

					Context("with a default value", func() {
						BeforeEach(func() {
							intParam.SetDefault(2)
						})

						It("writes the integer contexts code", func() {
							err := writer.Execute(data)
							Ω(err).ShouldNot(HaveOccurred())
							b, err := ioutil.ReadFile(filename)
							Ω(err).ShouldNot(HaveOccurred())
							written := string(b)
							Ω(written).ShouldNot(BeEmpty())
							Ω(written).Should(ContainSubstring(intRequiredDefaultContext))
							Ω(written).Should(ContainSubstring(intRequiredDefaultContextFactory))
						})
					})
				})
			})

			Context("with an string param", func() {
				var (
					strParam   *design.AttributeDefinition
					dataType   design.Object
					validation *dslengine.ValidationDefinition
				)

				BeforeEach(func() {
					strParam = &design.AttributeDefinition{Type: design.String}
					dataType = design.Object{
						"param": strParam,
					}
					validation = &dslengine.ValidationDefinition{}
					params = &design.AttributeDefinition{
						Type:       dataType,
						Validation: validation,
					}
				})

				It("writes the string contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(strContext))
					Ω(written).Should(ContainSubstring(strContextFactory))
				})

				Context("with a default value", func() {
					BeforeEach(func() {
						strParam.SetDefault("foo")
					})

					It("writes the string contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(strNonOptionalContext))
						Ω(written).Should(ContainSubstring(strDefaultContextFactory))
					})
				})

				Context("with required attribute", func() {
					BeforeEach(func() {
						validation.Required = []string{"param"}
					})

					It("writes the String contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(strNonOptionalContext))
						Ω(written).Should(ContainSubstring(strRequiredContextFactory))
					})

					Context("with a default value", func() {
						BeforeEach(func() {
							strParam.SetDefault("foo")
						})

						It("writes the integer contexts code", func() {
							err := writer.Execute(data)
							Ω(err).ShouldNot(HaveOccurred())
							b, err := ioutil.ReadFile(filename)
							Ω(err).ShouldNot(HaveOccurred())
							written := string(b)
							Ω(written).ShouldNot(BeEmpty())
							Ω(written).Should(ContainSubstring(strNonOptionalContext))
							Ω(written).Should(ContainSubstring(strDefaultContextFactory))
						})
					})
				})
			})

			Context("with a number param", func() {
				var (
					numParam   *design.AttributeDefinition
					dataType   design.Object
					validation *dslengine.ValidationDefinition
				)

				BeforeEach(func() {
					numParam = &design.AttributeDefinition{Type: design.Number}
					dataType = design.Object{
						"param": numParam,
					}
					validation = &dslengine.ValidationDefinition{}
					params = &design.AttributeDefinition{
						Type:       dataType,
						Validation: validation,
					}
				})

				It("writes the number contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(numContext))
					Ω(written).Should(ContainSubstring(numContextFactory))
				})

				Context("with a default value", func() {
					BeforeEach(func() {
						numParam.SetDefault(2.3)
					})

					It("writes the number contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(numNonOptionalContext))
						Ω(written).Should(ContainSubstring(numDefaultContextFactory))
					})
				})

				Context("with required attribute", func() {
					BeforeEach(func() {
						validation.Required = []string{"param"}
					})

					It("writes the number contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(numNonOptionalContext))
						Ω(written).Should(ContainSubstring(numRequiredContextFactory))
					})

					Context("with a default value", func() {
						BeforeEach(func() {
							numParam.SetDefault(2.3)
						})

						It("writes the number contexts code", func() {
							err := writer.Execute(data)
							Ω(err).ShouldNot(HaveOccurred())
							b, err := ioutil.ReadFile(filename)
							Ω(err).ShouldNot(HaveOccurred())
							written := string(b)
							Ω(written).ShouldNot(BeEmpty())
							Ω(written).Should(ContainSubstring(numNonOptionalContext))
							Ω(written).Should(ContainSubstring(numDefaultContextFactory))
						})
					})
				})
			})

			Context("with a boolean param", func() {
				var (
					boolParam  *design.AttributeDefinition
					dataType   design.Object
					validation *dslengine.ValidationDefinition
				)

				BeforeEach(func() {
					boolParam = &design.AttributeDefinition{Type: design.Boolean}
					dataType = design.Object{
						"param": boolParam,
					}
					validation = &dslengine.ValidationDefinition{}
					params = &design.AttributeDefinition{
						Type:       dataType,
						Validation: validation,
					}
				})

				It("writes the boolean contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(boolContext))
					Ω(written).Should(ContainSubstring(boolContextFactory))
				})

				Context("with a default value", func() {
					BeforeEach(func() {
						boolParam.SetDefault(true)
					})

					It("writes the boolean contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(boolNonOptionalContext))
						Ω(written).Should(ContainSubstring(boolDefaultContextFactory))
					})
				})

				Context("with required attribute", func() {
					BeforeEach(func() {
						validation.Required = []string{"param"}
					})

					It("writes the boolean contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(boolNonOptionalContext))
						Ω(written).Should(ContainSubstring(boolRequiredContextFactory))
					})

					Context("with a default value", func() {
						BeforeEach(func() {
							boolParam.SetDefault(true)
						})

						It("writes the boolean contexts code", func() {
							err := writer.Execute(data)
							Ω(err).ShouldNot(HaveOccurred())
							b, err := ioutil.ReadFile(filename)
							Ω(err).ShouldNot(HaveOccurred())
							written := string(b)
							Ω(written).ShouldNot(BeEmpty())
							Ω(written).Should(ContainSubstring(boolNonOptionalContext))
							Ω(written).Should(ContainSubstring(boolDefaultContextFactory))
						})
					})
				})
			})

			Context("with a array param", func() {
				var (
					arrayParam *design.AttributeDefinition
					dataType   design.Object
					validation *dslengine.ValidationDefinition
				)

				BeforeEach(func() {
					arrayParam = &design.AttributeDefinition{Type: &design.Array{ElemType: &design.AttributeDefinition{Type: design.String}}}
					dataType = design.Object{
						"param": arrayParam,
					}
					validation = &dslengine.ValidationDefinition{}
					params = &design.AttributeDefinition{
						Type:       dataType,
						Validation: validation,
					}
				})

				It("writes the array contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(arrayContext))
					Ω(written).Should(ContainSubstring(arrayContextFactory))
				})

				Context("with a default value", func() {
					BeforeEach(func() {
						arrayParam.SetDefault([]interface{}{"foo", "bar", "baz"})
					})

					It("writes the array contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(arrayContext))
						Ω(written).Should(ContainSubstring(arrayDefaultContextFactory))
					})
				})

				Context("with required attribute", func() {
					BeforeEach(func() {
						validation.Required = []string{"param"}
					})

					It("writes the array contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(arrayContext))
						Ω(written).Should(ContainSubstring(arrayRequiredContextFactory))
					})

					Context("with a default value", func() {
						BeforeEach(func() {
							arrayParam.SetDefault([]interface{}{"foo", "bar", "baz"})
						})

						It("writes the array contexts code", func() {
							err := writer.Execute(data)
							Ω(err).ShouldNot(HaveOccurred())
							b, err := ioutil.ReadFile(filename)
							Ω(err).ShouldNot(HaveOccurred())
							written := string(b)
							Ω(written).ShouldNot(BeEmpty())
							Ω(written).Should(ContainSubstring(arrayContext))
							Ω(written).Should(ContainSubstring(arrayDefaultContextFactory))
						})
					})
				})
			})

			Context("with an int array param", func() {
				var (
					arrayParam *design.AttributeDefinition
					dataType   design.Object
					validation *dslengine.ValidationDefinition
				)

				BeforeEach(func() {
					arrayParam = &design.AttributeDefinition{Type: &design.Array{ElemType: &design.AttributeDefinition{Type: design.Integer}}}
					dataType = design.Object{
						"param": arrayParam,
					}
					validation = &dslengine.ValidationDefinition{}
					params = &design.AttributeDefinition{
						Type:       dataType,
						Validation: validation,
					}
				})

				It("writes the array contexts code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(intArrayContext))
					Ω(written).Should(ContainSubstring(intArrayContextFactory))
				})

				Context("with a default value", func() {
					BeforeEach(func() {
						arrayParam.SetDefault([]interface{}{1, 1, 2, 3, 5, 8})
					})

					It("writes the array contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(intArrayContext))
						Ω(written).Should(ContainSubstring(intArrayDefaultContextFactory))
					})
				})

				Context("with required attribute", func() {
					BeforeEach(func() {
						validation.Required = []string{"param"}
					})

					It("writes the array contexts code", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(intArrayContext))
						Ω(written).Should(ContainSubstring(intArrayRequiredContextFactory))
					})

					Context("with a default value", func() {
						BeforeEach(func() {
							arrayParam.SetDefault([]interface{}{1, 1, 2, 3, 5, 8})
						})

						It("writes the array contexts code", func() {
							err := writer.Execute(data)
							Ω(err).ShouldNot(HaveOccurred())
							b, err := ioutil.ReadFile(filename)
							Ω(err).ShouldNot(HaveOccurred())
							written := string(b)
							Ω(written).ShouldNot(BeEmpty())
							Ω(written).Should(ContainSubstring(intArrayContext))
							Ω(written).Should(ContainSubstring(intArrayDefaultContextFactory))
						})
					})
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

			Context("with a custom name param", func() {
				BeforeEach(func() {
					intParam := &design.AttributeDefinition{
						Type: design.Integer,
						Metadata: dslengine.MetadataDefinition{
							"struct:field:name": []string{"custom"},
						},
					}
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
					Ω(written).Should(ContainSubstring(customContext))
					Ω(written).Should(ContainSubstring(customContextFactory))
				})
			})

			Context("with a string header", func() {
				BeforeEach(func() {
					strHeader := &design.AttributeDefinition{Type: design.String}
					dataType := design.Object{
						"Header": strHeader,
					}
					headers = &design.AttributeDefinition{
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
					Ω(written).Should(ContainSubstring(strHeaderContext))
					Ω(written).Should(ContainSubstring(strHeaderContextFactory))
				})
			})

			Context("with a string header and param with the same name", func() {
				BeforeEach(func() {
					str := &design.AttributeDefinition{Type: design.String}
					dataType := design.Object{
						"param": str,
					}
					params = &design.AttributeDefinition{
						Type: dataType,
					}
					headers = &design.AttributeDefinition{
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
					Ω(written).Should(ContainSubstring(strHeaderParamContextFactory))
				})
			})

			Context("with a simple payload", func() {
				BeforeEach(func() {
					design.Design = new(design.APIDefinition)
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
					design.Design = new(design.APIDefinition)
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

		Context("with file servers", func() {
			requestPath := "/swagger.json"
			filePath := "swagger/swagger.json"
			var origins []*design.CORSDefinition
			var preflightPaths []string

			var data []*genapp.ControllerTemplateData

			BeforeEach(func() {
				origins = nil
				preflightPaths = nil
			})

			JustBeforeEach(func() {
				codegen.TempCount = 0
				fileServer := &design.FileServerDefinition{
					FilePath:    filePath,
					RequestPath: requestPath,
				}
				d := &genapp.ControllerTemplateData{
					API:            &design.APIDefinition{},
					Origins:        origins,
					PreflightPaths: preflightPaths,
					Resource:       "Public",
					FileServers:    []*design.FileServerDefinition{fileServer},
				}
				data = []*genapp.ControllerTemplateData{d}
			})

			It("writes the file server code", func() {
				err := writer.Execute(data)
				Ω(err).ShouldNot(HaveOccurred())
				b, err := ioutil.ReadFile(filename)
				Ω(err).ShouldNot(HaveOccurred())
				written := string(b)
				Ω(written).ShouldNot(BeEmpty())
				Ω(written).Should(ContainSubstring(simpleFileServer))
			})

			Context("with CORS", func() {
				BeforeEach(func() {
					origins = []*design.CORSDefinition{
						{
							// NB: including backslash to ensure proper escaping
							Origin:      "here.example.com",
							Headers:     []string{"X-One", "X-Two"},
							Methods:     []string{"GET", "POST"},
							Exposed:     []string{"X-Three"},
							Credentials: true,
						},
					}
					preflightPaths = []string{"/public/star\\*star/*filepath"}
				})

				It("writes the OPTIONS handler code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(fileServerOptionsHandler))
				})
			})
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

			Context("with regexp origins", func() {
				BeforeEach(func() {
					actions = []string{"List"}
					verbs = []string{"GET"}
					paths = []string{"/accounts"}
					contexts = []string{"ListBottleContext"}
					origins = []*design.CORSDefinition{
						{
							Origin:      "[here|there].example.com",
							Headers:     []string{"X-One", "X-Two"},
							Methods:     []string{"GET", "POST"},
							Exposed:     []string{"X-Three"},
							Credentials: true,
							Regexp:      true,
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
					Ω(written).Should(ContainSubstring(regexpOriginsHandler))
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

				Context("and a canonical action with no param", func() {
					BeforeEach(func() {
						canoTemplate = "/bottles"
					})

					It("writes the href method", func() {
						err := writer.Execute(data)
						Ω(err).ShouldNot(HaveOccurred())
						b, err := ioutil.ReadFile(filename)
						Ω(err).ShouldNot(HaveOccurred())
						written := string(b)
						Ω(written).ShouldNot(BeEmpty())
						Ω(written).Should(ContainSubstring(noParamHref))
					})
				})
			})
		})
	})
})

var _ = Describe("UserTypesWriter", func() {
	var writer *genapp.UserTypesWriter
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
		writer, err = genapp.NewUserTypesWriter(filename)
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("correctly configured", func() {
		Context("with data", func() {
			var data *design.UserTypeDefinition
			var attDef *design.AttributeDefinition
			var typeName string

			BeforeEach(func() {
				data = nil
				attDef = nil
				typeName = ""
			})

			JustBeforeEach(func() {
				data = &design.UserTypeDefinition{
					AttributeDefinition: attDef,
					TypeName:            typeName,
				}
			})

			Context("with a simple user type", func() {
				BeforeEach(func() {
					attDef = &design.AttributeDefinition{
						Type: design.Object{
							"name": &design.AttributeDefinition{
								Type: design.String,
							},
						},
					}
					typeName = "SimplePayload"
				})
				It("writes the simple user type code", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(simpleUserType))
				})
			})

			Context("with a user type including hash", func() {
				BeforeEach(func() {
					attDef = &design.AttributeDefinition{
						Type: design.Object{
							"name": &design.AttributeDefinition{
								Type: design.String,
							},
							"misc": &design.AttributeDefinition{
								Type: &design.Hash{
									KeyType: &design.AttributeDefinition{
										Type: design.Integer,
									},
									ElemType: &design.AttributeDefinition{
										Type: &design.UserTypeDefinition{
											AttributeDefinition: &design.AttributeDefinition{
												Type: &design.UserTypeDefinition{
													AttributeDefinition: &design.AttributeDefinition{
														Type: design.Object{},
													},
													TypeName: "Misc",
												},
											},
											TypeName: "MiscPayload",
										},
									},
								},
							},
						},
					}
					typeName = "ComplexPayload"
				})
				It("writes the user type including hash", func() {
					err := writer.Execute(data)
					Ω(err).ShouldNot(HaveOccurred())
					b, err := ioutil.ReadFile(filename)
					Ω(err).ShouldNot(HaveOccurred())
					written := string(b)
					Ω(written).ShouldNot(BeEmpty())
					Ω(written).Should(ContainSubstring(userTypeIncludingHash))
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
}
`

	emptyContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	return &rctx, err
}
`

	intContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param *int
}
`

	intContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
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

	intDefaultContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param int
}
`

	intDefaultContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		rctx.Param = 2
	} else {
		rawParam := paramParam[0]
		if param, err2 := strconv.Atoi(rawParam); err2 == nil {
			rctx.Param = param
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "integer"))
		}
	}
	return &rctx, err
}
`

	intRequiredDefaultContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param int
}
`

	intRequiredDefaultContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		rctx.Param = 2
	} else {
		rawParam := paramParam[0]
		if param, err2 := strconv.Atoi(rawParam); err2 == nil {
			rctx.Param = param
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "integer"))
		}
	}
	return &rctx, err
}
`

	intRequiredContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param int
}
`
	intRequiredContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("param"))
	} else {
		rawParam := paramParam[0]
		if param, err2 := strconv.Atoi(rawParam); err2 == nil {
			rctx.Param = param
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
	Param *string
}
`

	strContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		rawParam := paramParam[0]
		rctx.Param = &rawParam
	}
	return &rctx, err
}
`

	strNonOptionalContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param string
}
`

	strDefaultContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		rctx.Param = "foo"
	} else {
		rawParam := paramParam[0]
		rctx.Param = rawParam
	}
	return &rctx, err
}
`

	strRequiredContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("param"))
	} else {
		rawParam := paramParam[0]
		rctx.Param = rawParam
	}
	return &rctx, err
}
`

	strHeaderContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Header *string
}
`

	strHeaderContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	headerHeader := req.Header["Header"]
	if len(headerHeader) > 0 {
		rawHeader := headerHeader[0]
		req.Params["Header"] = []string{rawHeader}
		rctx.Header = &rawHeader
	}
	return &rctx, err
}
`

	strHeaderParamContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	headerParam := req.Header["Param"]
	if len(headerParam) > 0 {
		rawParam := headerParam[0]
		req.Params["param"] = []string{rawParam}
		rctx.Param = &rawParam
	}
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
	Param *float64
}
`

	numContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
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

	numNonOptionalContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param float64
}
`

	numDefaultContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		rctx.Param = 2.300000
	} else {
		rawParam := paramParam[0]
		if param, err2 := strconv.ParseFloat(rawParam, 64); err2 == nil {
			rctx.Param = param
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "number"))
		}
	}
	return &rctx, err
}
`

	numRequiredContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("param"))
	} else {
		rawParam := paramParam[0]
		if param, err2 := strconv.ParseFloat(rawParam, 64); err2 == nil {
			rctx.Param = param
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
	Param *bool
}
`

	boolContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
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

	boolNonOptionalContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param bool
}
`

	boolDefaultContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		rctx.Param = true
	} else {
		rawParam := paramParam[0]
		if param, err2 := strconv.ParseBool(rawParam); err2 == nil {
			rctx.Param = param
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "boolean"))
		}
	}
	return &rctx, err
}
`

	boolRequiredContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("param"))
	} else {
		rawParam := paramParam[0]
		if param, err2 := strconv.ParseBool(rawParam); err2 == nil {
			rctx.Param = param
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
	Param []string
}
`

	arrayContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		params := paramParam
		rctx.Param = params
	}
	return &rctx, err
}
`

	arrayDefaultContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		rctx.Param = []string{"foo", "bar", "baz"}
	} else {
		params := paramParam
		rctx.Param = params
	}
	return &rctx, err
}
`

	arrayRequiredContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("param"))
	} else {
		params := paramParam
		rctx.Param = params
	}
	return &rctx, err
}
`

	intArrayContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Param []int
}
`

	intArrayContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) > 0 {
		params := make([]int, len(paramParam))
		for i, rawParam := range paramParam {
			if param, err2 := strconv.Atoi(rawParam); err2 == nil {
				params[i] = param
			} else {
				err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "integer"))
			}
		}
		rctx.Param = params
	}
	return &rctx, err
}
`

	intArrayDefaultContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		rctx.Param = []int{1, 1, 2, 3, 5, 8}
	} else {
		params := make([]int, len(paramParam))
		for i, rawParam := range paramParam {
			if param, err2 := strconv.Atoi(rawParam); err2 == nil {
				params[i] = param
			} else {
				err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "integer"))
			}
		}
		rctx.Param = params
	}
	return &rctx, err
}
`

	intArrayRequiredContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramParam := req.Params["param"]
	if len(paramParam) == 0 {
		err = goa.MergeErrors(err, goa.MissingParamError("param"))
	} else {
		params := make([]int, len(paramParam))
		for i, rawParam := range paramParam {
			if param, err2 := strconv.Atoi(rawParam); err2 == nil {
				params[i] = param
			} else {
				err = goa.MergeErrors(err, goa.InvalidParamTypeError("param", rawParam, "integer"))
			}
		}
		rctx.Param = params
	}
	return &rctx, err
}
`

	resContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Int *int
}
`

	resContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
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
	Int int
}
`

	requiredContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
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

	customContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Custom *int
}
`

	customContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramInt := req.Params["int"]
	if len(paramInt) > 0 {
		rawInt := paramInt[0]
		if int_, err2 := strconv.Atoi(rawInt); err2 == nil {
			tmp2 := int_
			tmp1 := &tmp2
			rctx.Custom = tmp1
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
	Payload ListBottlePayload
}
`

	payloadContextFactory = `
func NewListBottleContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListBottleContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListBottleContext{Context: ctx, ResponseData: resp, RequestData: req}
	return &rctx, err
}
`
	payloadObjContext = `
type ListBottleContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
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
		// Initialize payload with private data structure so it can be logged
		goa.ContextRequest(ctx).Payload = payload
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

	simpleFileServer = `// PublicController is the controller interface for the Public actions.
type PublicController interface {
	goa.Muxer
	goa.FileServer
}
`

	fileServerOptionsHandler = `service.Mux.Handle("OPTIONS", "/public/star\\*star/*filepath", ctrl.MuxHandler("preflight", handlePublicOrigin(cors.HandlePreflight()), nil))`

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
			rw.Header().Set("Access-Control-Allow-Origin", origin)
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
			rw.Header().Set("Access-Control-Allow-Origin", origin)
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

	regexpOriginsHandler = `// handleBottlesOrigin applies the CORS response headers corresponding to the origin.
func handleBottlesOrigin(h goa.Handler) goa.Handler {
	spec0 := regexp.MustCompile("[here|there].example.com")

	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		origin := req.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			return h(ctx, rw, req)
		}
		if cors.MatchOriginRegexp(origin, spec0) {
			ctx = goa.WithLogContext(ctx, "origin", origin)
			rw.Header().Set("Access-Control-Allow-Origin", origin)
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
			rw.Header().Set("Access-Control-Allow-Origin", origin)
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
		// Check if there was an error loading the request
		if err := goa.ContextError(ctx); err != nil {
			return err
		}
		// Build the context
		rctx, err := NewListBottleContext(ctx, req, service)
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
		// Check if there was an error loading the request
		if err := goa.ContextError(ctx); err != nil {
			return err
		}
		// Build the context
		rctx, err := NewListBottleContext(ctx, req, service)
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
		// Check if there was an error loading the request
		if err := goa.ContextError(ctx); err != nil {
			return err
		}
		// Build the context
		rctx, err := NewListBottleContext(ctx, req, service)
		if err != nil {
			return err
		}
		return ctrl.List(rctx)
	}
	service.Mux.Handle("GET", "/accounts/:accountID/bottles", ctrl.MuxHandler("List", h, nil))
	service.LogInfo("mount", "ctrl", "Bottles", "action", "List", "route", "GET /accounts/:accountID/bottles")

	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		// Check if there was an error loading the request
		if err := goa.ContextError(ctx); err != nil {
			return err
		}
		// Build the context
		rctx, err := NewShowBottleContext(ctx, req, service)
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
	paramid := strings.TrimLeftFunc(fmt.Sprintf("%v", id), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/bottles/%v", paramid)
}
`
	noParamHref = `func BottleHref() string {
	return "/bottles"
}
`

	simpleUserType = `// simplePayload user type.
type simplePayload struct {
	Name *string ` + "`" + `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"` + "`" + `
}



// Publicize creates SimplePayload from simplePayload
func (ut *simplePayload) Publicize() *SimplePayload {
	var pub SimplePayload
		if ut.Name != nil {
		pub.Name = ut.Name
	}
	return &pub
}

// SimplePayload user type.
type SimplePayload struct {
	Name *string ` + "`" + `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"` + "`" + `
}
`

	userTypeIncludingHash = `// complexPayload user type.
type complexPayload struct {
	Misc map[int]*miscPayload ` + "`" + `form:"misc,omitempty" json:"misc,omitempty" xml:"misc,omitempty"` + "`" + `
	Name *string ` + "`" + `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"` + "`" + `
}



// Publicize creates ComplexPayload from complexPayload
func (ut *complexPayload) Publicize() *ComplexPayload {
	var pub ComplexPayload
		if ut.Misc != nil {
		pub.Misc = make(map[int]*MiscPayload, len(ut.Misc))
		for k2, v2 := range ut.Misc {
						pubk2 := k2
			var pubv2 *MiscPayload
			if v2 != nil {
						pubv2 = v2.Publicize()
			}
			pub.Misc[pubk2] = pubv2
		}
	}
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	return &pub
}

// ComplexPayload user type.
type ComplexPayload struct {
	Misc map[int]*MiscPayload ` + "`" + `form:"misc,omitempty" json:"misc,omitempty" xml:"misc,omitempty"` + "`" + `
	Name *string ` + "`" + `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"` + "`" + `
}
`
)
