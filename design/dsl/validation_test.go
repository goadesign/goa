package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("Validation", func() {
	Context("with a type attribute", func() {
		const attName = "attName"
		var dsl func()

		var att *AttributeDefinition

		JustBeforeEach(func() {
			Design = nil
			Errors = nil
			Type("bar", func() {
				dsl()
			})
			RunDSL()
			if Errors == nil {
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
					Attribute(attName, String, func() {
						Enum("red", "blue")
					})
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&EnumValidationDefinition{}))
				expected := &EnumValidationDefinition{Values: []interface{}{"red", "blue"}}
				Ω(v.(*EnumValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an incompatible enum validation type", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Enum(1, "blue")
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid format validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Format("email")
					})
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&FormatValidationDefinition{}))
				expected := &FormatValidationDefinition{Format: "email"}
				Ω(v.(*FormatValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid format validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Format("emailz")
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid pattern validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Pattern("^foo$")
					})
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&PatternValidationDefinition{}))
				expected := &PatternValidationDefinition{Pattern: "^foo$"}
				Ω(v.(*PatternValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid pattern validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Pattern("[invalid")
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with an invalid format validation type", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Format("email")
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid min value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Minimum(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&MinimumValidationDefinition{}))
				expected := &MinimumValidationDefinition{Min: 2}
				Ω(v.(*MinimumValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid min value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Minimum(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid max value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Maximum(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&MaximumValidationDefinition{}))
				expected := &MaximumValidationDefinition{Max: 2}
				Ω(v.(*MaximumValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid max value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Maximum(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid min length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, ArrayOf(Integer), func() {
						MinLength(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&MinLengthValidationDefinition{}))
				expected := &MinLengthValidationDefinition{MinLength: 2}
				Ω(v.(*MinLengthValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid min length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						MinLength(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid max length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						MaxLength(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&MaxLengthValidationDefinition{}))
				expected := &MaxLengthValidationDefinition{MaxLength: 2}
				Ω(v.(*MaxLengthValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid max length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						MaxLength(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(Errors).Should(HaveOccurred())
			})
		})

		Context("with a required field validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String)
					Required(attName)
				}
			})

			It("records the validation", func() {
				Ω(Errors).ShouldNot(HaveOccurred())
				Ω(Design.Types["bar"].Validations).Should(HaveLen(1))
				v := Design.Types["bar"].Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&RequiredValidationDefinition{}))
				expected := &RequiredValidationDefinition{Names: []string{attName}}
				Ω(v.(*RequiredValidationDefinition)).Should(Equal(expected))
			})
		})
	})
})
