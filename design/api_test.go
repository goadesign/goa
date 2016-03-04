package design_test

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CanonicalIdentifier", func() {
	var id string
	var canonical string

	JustBeforeEach(func() {
		canonical = design.CanonicalIdentifier(id)
	})

	Context("with a canonical identifier", func() {
		BeforeEach(func() {
			id = "application/json"
		})

		It("returns it", func() {
			Ω(canonical).Should(Equal(id))
		})
	})

	Context("with a non canonical identifier", func() {
		BeforeEach(func() {
			id = "application/json+xml; foo=bar"
		})

		It("canonicalizes it", func() {
			Ω(canonical).Should(Equal("application/json; foo=bar"))
		})
	})
})

var _ = Describe("ExtractWildcards", func() {
	var path string
	var wcs []string

	JustBeforeEach(func() {
		wcs = design.ExtractWildcards(path)
	})

	Context("with a path with no wildcard", func() {
		BeforeEach(func() {
			path = "/foo"
		})

		It("returns the empty slice", func() {
			Ω(wcs).Should(HaveLen(0))
		})
	})

	Context("with a path with wildcards", func() {
		BeforeEach(func() {
			path = "/a/:foo/:bar/b/:baz/c"
		})

		It("extracts them", func() {
			Ω(wcs).Should(Equal([]string{"foo", "bar", "baz"}))
		})
	})
})

var _ = Describe("MediaTypeRoot", func() {
	var root design.MediaTypeRoot

	BeforeEach(func() {
		design.Design.MediaTypes = make(map[string]*design.MediaTypeDefinition)
		root = design.MediaTypeRoot{}
	})

	It("has a non empty DSL name", func() {
		Ω(root.DSLName()).ShouldNot(BeEmpty())
	})

	It("depends on the goa API design root", func() {
		Ω(root.DependsOn()).Should(Equal([]dslengine.Root{design.Design}))
	})

	It("iterates over the generated media types when it's empty", func() {
		var sets []dslengine.DefinitionSet
		it := func(s dslengine.DefinitionSet) error {
			sets = append(sets, s)
			return nil
		}
		root.IterateSets(it)
		Ω(sets).Should(HaveLen(1))
		Ω(sets[0]).Should(BeEmpty())
	})

	It("iterates over the generated media types", func() {
		var sets []dslengine.DefinitionSet
		it := func(s dslengine.DefinitionSet) error {
			sets = append(sets, s)
			return nil
		}
		root["foo"] = &design.MediaTypeDefinition{Identifier: "application/json"}
		root.IterateSets(it)
		Ω(sets).Should(HaveLen(1))
		Ω(sets[0]).Should(HaveLen(1))
		Ω(sets[0][0]).Should(Equal(root["foo"]))
	})

	It("iterates over the generated media types in order", func() {
		var sets []dslengine.DefinitionSet
		it := func(s dslengine.DefinitionSet) error {
			sets = append(sets, s)
			return nil
		}
		root["foo"] = &design.MediaTypeDefinition{Identifier: "application/json"}
		root["bar"] = &design.MediaTypeDefinition{Identifier: "application/xml"}
		root.IterateSets(it)
		Ω(sets).Should(HaveLen(1))
		Ω(sets[0]).Should(HaveLen(2))
		Ω(sets[0][0]).Should(Equal(root["foo"]))
		Ω(sets[0][1]).Should(Equal(root["bar"]))
	})
})
