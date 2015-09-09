package app_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/codegen/code/app"
	"github.com/raphael/goa/design"
)

var _ = Describe("ResourceWriter", func() {
	var writer *app.ResourcesWriter
	var filename string
	var newErr error

	JustBeforeEach(func() {
		writer, newErr = app.NewResourcesWriter(filename)
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

		It("NewResourcesWriter creates a writer", func() {
			Ω(newErr).ShouldNot(HaveOccurred())
		})

		Context("with data", func() {
			var canoTemplate string
			var canoParams []string
			var userType *design.UserTypeDefinition

			var data *app.ResourceTemplateData

			BeforeEach(func() {
				userType = nil
				canoTemplate = ""
				canoParams = nil
				data = nil
			})

			JustBeforeEach(func() {
				data = &app.ResourceTemplateData{
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
					err := writer.Write(data)
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
						Name:                "Bottle",
						Description:         "Bottle type",
					}
				})
				It("writes the resources code", func() {
					err := writer.Write(data)
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
						Name:                "Bottle",
						Description:         "Bottle type",
					}
				})

				It("writes the resources code", func() {
					err := writer.Write(data)
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
						err := writer.Write(data)
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
