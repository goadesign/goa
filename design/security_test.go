package design_test

import (
	"fmt"

	. "github.com/goadesign/goa/design"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecuritySchemeDefinition", func() {
	var scheme, host, tokenURL, authorizationURL string

	var def *SecuritySchemeDefinition

	BeforeEach(func() {
		def = nil
		tokenURL = ""
		authorizationURL = ""
		scheme = ""
		host = ""
	})

	JustBeforeEach(func() {
		Design.Schemes = []string{scheme}
		Design.Host = host
		def = &SecuritySchemeDefinition{
			TokenURL:         tokenURL,
			AuthorizationURL: authorizationURL,
		}
	})

	Context("with valid token and authorization URLs", func() {
		BeforeEach(func() {
			tokenURL = "http://valid.com/token"
			authorizationURL = "http://valid.com/auth"
		})

		It("validates", func() {
			Ω(def.Validate()).ShouldNot(HaveOccurred())
		})
	})

	Context("with an invalid token URL", func() {
		BeforeEach(func() {
			tokenURL = ":"
			authorizationURL = "http://valid.com/auth"
		})

		It("does not validate", func() {
			err := def.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(tokenURL))
		})
	})

	Context("with an absolute token URL", func() {
		BeforeEach(func() {
			tokenURL = "http://valid.com/auth"
		})

		It("Finalize does not modify it", func() {
			priorURL := def.TokenURL
			def.Finalize()
			Ω(def.TokenURL).Should(Equal(priorURL))
		})
	})

	Context("with a relative token URL", func() {
		BeforeEach(func() {
			scheme = "http"
			host = "foo.com"
			tokenURL = "/auth"
		})

		It("Finalize makes it absolute", func() {
			priorURL := def.TokenURL
			def.Finalize()
			Ω(def.TokenURL).Should(Equal(fmt.Sprintf("%s://%s%s", scheme, host, priorURL)))
		})
	})

	Context("with an invalid authorization URL", func() {
		BeforeEach(func() {
			tokenURL = "http://valid.com/auth"
			authorizationURL = ":"
		})

		It("does not validate", func() {
			err := def.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(authorizationURL))
		})
	})

	Context("with an absolute authorization URL", func() {
		BeforeEach(func() {
			authorizationURL = "http://valid.com/auth"
		})

		It("Finalize does not modify it", func() {
			priorURL := def.AuthorizationURL
			def.Finalize()
			Ω(def.AuthorizationURL).Should(Equal(priorURL))
		})
	})

	Context("with a relative authorization URL", func() {
		BeforeEach(func() {
			scheme = "http"
			host = "foo.com"
			authorizationURL = "/auth"
		})

		It("Finalize makes it absolute", func() {
			priorURL := def.AuthorizationURL
			def.Finalize()
			Ω(def.AuthorizationURL).Should(Equal(fmt.Sprintf("%s://%s%s", scheme, host, priorURL)))
		})
	})

})
