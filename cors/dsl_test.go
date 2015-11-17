package cors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
	"github.com/raphael/goa/cors"
)

var _ = Describe("valid CORS DSL", func() {
	var dsl func()
	var spec cors.Specification
	var dslErrors error

	JustBeforeEach(func() {
		spec, dslErrors = cors.New(dsl)
		Ω(dslErrors).ShouldNot(HaveOccurred())
	})

	Context("with an empty DSL", func() {
		BeforeEach(func() {
			dsl = nil
		})

		It("returns an empty spec", func() {
			Ω(spec).ShouldNot(BeNil())
			Ω(spec).Should(HaveLen(0))
		})
	})

	Context("Origin", func() {
		const origin = "ORIGIN"
		const path = "PATH"

		BeforeEach(func() {
			dsl = func() {
				cors.Origin(origin, func() {
					cors.Resource(path, func() {
						cors.Methods("GET")
					})
				})
			}
		})

		It("sets the resource origin", func() {
			Ω(spec).Should(HaveLen(1))
			Ω(spec[0]).ShouldNot(BeNil())
			Ω(spec[0].Origin).Should(Equal(origin))
			Ω(spec[0].Path).Should(Equal(path))
			Ω(spec[0].IsPathPrefix).Should(BeFalse())
		})

		Context("Headers", func() {
			headers := []string{"X-Foo", "X-Bar"}

			BeforeEach(func() {
				dsl = func() {
					cors.Origin(origin, func() {
						cors.Resource(path, func() {
							cors.Headers(headers[0], headers[1])
						})
					})
				}
			})

			It("sets the resource headers", func() {
				Ω(spec).Should(HaveLen(1))
				Ω(spec[0]).ShouldNot(BeNil())
				Ω(spec[0].Headers).Should(HaveLen(len(headers)))
				for i, h := range spec[0].Headers {
					Ω(h).Should(Equal(headers[i]))
				}
			})
		})

		Context("Methods", func() {
			methods := []string{"GET", "HEAD"}

			BeforeEach(func() {
				dsl = func() {
					cors.Origin(origin, func() {
						cors.Resource(path, func() {
							cors.Methods(methods[0], methods[1])
						})
					})
				}
			})

			It("sets the resource methods", func() {
				Ω(spec).Should(HaveLen(1))
				Ω(spec[0]).ShouldNot(BeNil())
				Ω(spec[0].Methods).Should(HaveLen(len(methods)))
				for i, m := range spec[0].Methods {
					Ω(m).Should(Equal(methods[i]))
				}
			})
		})

		Context("Expose", func() {
			expose := []string{"X-Foo", "X-Bar"}

			BeforeEach(func() {
				dsl = func() {
					cors.Origin(origin, func() {
						cors.Resource(path, func() {
							cors.Expose(expose[0], expose[1])
						})
					})
				}
			})

			It("sets the resource expose", func() {
				Ω(spec).Should(HaveLen(1))
				Ω(spec[0]).ShouldNot(BeNil())
				Ω(spec[0].Expose).Should(HaveLen(len(expose)))
				for i, h := range spec[0].Expose {
					Ω(h).Should(Equal(expose[i]))
				}
			})
		})

		Context("MaxAge", func() {
			const maxAge = 200

			BeforeEach(func() {
				dsl = func() {
					cors.Origin(origin, func() {
						cors.Resource(path, func() {
							cors.MaxAge(maxAge)
						})
					})
				}
			})

			It("sets the resource max age", func() {
				Ω(spec).Should(HaveLen(1))
				Ω(spec[0]).ShouldNot(BeNil())
				Ω(spec[0].MaxAge).Should(Equal(maxAge))
			})
		})

		Context("Credentials", func() {
			const credentials = true

			BeforeEach(func() {
				dsl = func() {
					cors.Origin(origin, func() {
						cors.Resource(path, func() {
							cors.Credentials(credentials)
						})
					})
				}
			})

			It("sets the resource credentials flag", func() {
				Ω(spec).Should(HaveLen(1))
				Ω(spec[0]).ShouldNot(BeNil())
				Ω(spec[0].Credentials).Should(Equal(credentials))
			})
		})

		Context("Vary", func() {
			vary := []string{"X-Origin"}

			BeforeEach(func() {
				dsl = func() {
					cors.Origin(origin, func() {
						cors.Resource(path, func() {
							cors.Vary(vary[0])
						})
					})
				}
			})

			It("sets the resource 'Vary' response header", func() {
				Ω(spec).Should(HaveLen(1))
				Ω(spec[0]).ShouldNot(BeNil())
				Ω(spec[0].Vary).Should(HaveLen(len(vary)))
				for i, h := range spec[0].Vary {
					Ω(h).Should(Equal(vary[i]))
				}
			})
		})

		Context("Check", func() {
			check := cors.CheckFunc(func(*goa.Context) bool { return false })

			BeforeEach(func() {
				dsl = func() {
					cors.Origin(origin, func() {
						cors.Resource(path, func() {
							cors.Check(check)
						})
					})
				}
			})

			It("sets the resource check", func() {
				Ω(spec).Should(HaveLen(1))
				Ω(spec[0]).ShouldNot(BeNil())
				Ω(spec[0].Check).ShouldNot(BeNil())
			})
		})
	})
})

var _ = Describe("invalid CORS DSL", func() {
	var dsl func()
	var spec cors.Specification
	var dslErrors error

	JustBeforeEach(func() {
		spec, dslErrors = cors.New(dsl)
	})

	Context("invalid top level", func() {
		BeforeEach(func() {
			dsl = func() {
				cors.Origin("foo", func() {
					cors.Headers("Bar")
				})
			}
		})

		It("returns a nil spec and an error", func() {
			Ω(spec).Should(BeNil())
			Ω(dslErrors).ShouldNot(BeNil())
			Ω(dslErrors.Error()).Should(ContainSubstring("invalid CORS specification"))
		})

	})

})
