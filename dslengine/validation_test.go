package dslengine_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/design/apidsl"
	"github.com/kyokomi/goa-v1/dslengine"
)

var _ = Describe("Validation", func() {
	Context("with a type attribute", func() {
		const attName = "attName"
		var dsl func()

		var att *AttributeDefinition

		JustBeforeEach(func() {
			dslengine.Reset()
			apidsl.Type("bar", func() {
				dsl()
			})
			dslengine.Run()
			if dslengine.Errors == nil {
				Ω(Design.Types).ShouldNot(BeNil())
				Ω(Design.Types).Should(HaveKey("bar"))
				Ω(Design.Types["bar"]).ShouldNot(BeNil())
				Ω(Design.Types["bar"].Type).Should(BeAssignableToTypeOf(Object{}))
				o := Design.Types["bar"].Type.(Object)
				Ω(o).Should(HaveKey(attName))
				att = o[attName]
			}
		})

		Context("with a valid enum validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.Enum("red", "blue")
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validation).ShouldNot(BeNil())
				Ω(att.Validation.Values).Should(Equal([]interface{}{"red", "blue"}))
			})
		})

		Context("with an incompatible enum validation type", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, Integer, func() {
						apidsl.Enum(1, "blue")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid format validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.Format("email")
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validation).ShouldNot(BeNil())
				Ω(att.Validation.Format).Should(Equal("email"))
			})
		})

		Context("with an invalid format validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.Format("emailz")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid pattern validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.Pattern("^foo$")
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validation).ShouldNot(BeNil())
				Ω(att.Validation.Pattern).Should(Equal("^foo$"))
			})
		})

		Context("with an invalid pattern validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.Pattern("[invalid")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with an invalid format validation type", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, Integer, func() {
						apidsl.Format("email")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid min value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, Integer, func() {
						apidsl.Minimum(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validation).ShouldNot(BeNil())
				Ω(att.Validation.Minimum).ShouldNot(BeNil())
				Ω(*att.Validation.Minimum).Should(Equal(float64(2)))
			})
		})

		Context("with an invalid min value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.Minimum(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid max value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, Integer, func() {
						apidsl.Maximum(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validation).ShouldNot(BeNil())
				Ω(att.Validation.Maximum).ShouldNot(BeNil())
				Ω(*att.Validation.Maximum).Should(Equal(float64(2)))
			})
		})

		Context("with an invalid max value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.Maximum(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid min length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, apidsl.ArrayOf(Integer), func() {
						apidsl.MinLength(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validation).ShouldNot(BeNil())
				Ω(att.Validation.MinLength).ShouldNot(BeNil())
				Ω(*att.Validation.MinLength).Should(Equal(2))
			})
		})

		Context("with an invalid min length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, Integer, func() {
						apidsl.MinLength(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid max length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String, func() {
						apidsl.MaxLength(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validation.MaxLength).ShouldNot(BeNil())
				Ω(*att.Validation.MaxLength).Should(Equal(2))
			})
		})

		Context("with an invalid max length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, Integer, func() {
						apidsl.MaxLength(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a required field validation", func() {
			BeforeEach(func() {
				dsl = func() {
					apidsl.Attribute(attName, String)
					apidsl.Required(attName)
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(Design.Types["bar"].Validation).ShouldNot(BeNil())
				Ω(Design.Types["bar"].Validation.Required).Should(Equal([]string{attName}))
			})
		})
	})
})
