package main

import (
	"os"

	"github.com/raphael/goa/goagen/bootstrap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("goagen", func() {
	It("registers all commands", func() {
		Ω(bootstrap.Commands).Should(HaveKey("app"))
		Ω(bootstrap.Commands).Should(HaveKey("docs"))
		Ω(bootstrap.Commands).Should(HaveKey("test"))
		Ω(bootstrap.Commands).Should(HaveKey("client"))
		Ω(bootstrap.Commands).Should(HaveKey("js"))
	})

	Context("with a valid command line", func() {
		const cmdName = "app"

		BeforeEach(func() {
			os.Args = []string{"goagen", cmdName, "--package", "design", "--out", "."}
		})

		It("command returns the correct command", func() {
			cmd := command()
			Ω(cmd).ShouldNot(BeNil())
			Ω(cmd.Name()).Should(Equal(cmdName))
		})
	})
})
