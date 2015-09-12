package main

import (
	"os"

	"github.com/raphael/goa/goagen/app"
	"github.com/raphael/goa/goagen/bootstrap"
	"github.com/raphael/goa/goagen/client"
	"github.com/raphael/goa/goagen/docs"
	"github.com/raphael/goa/goagen/js"
	"github.com/raphael/goa/goagen/test"

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

	Context("all commands", func() {
		var a *app.Command
		var d *docs.Command
		var t *test.Command
		var c *client.Command
		var j *js.Command

		BeforeEach(func() {
			a = bootstrap.Commands["app"].(*app.Command)
			d = bootstrap.Commands["docs"].(*docs.Command)
			t = bootstrap.Commands["test"].(*test.Command)
			c = bootstrap.Commands["client"].(*client.Command)
			j = bootstrap.Commands["js"].(*js.Command)
		})

		It("initialiazes", func() {
			Ω(a.Factory).ShouldNot(BeEmpty())
			Ω(d.Factory).ShouldNot(BeEmpty())
			Ω(t.Factory).ShouldNot(BeEmpty())
			Ω(c.Factory).ShouldNot(BeEmpty())
			Ω(j.Factory).ShouldNot(BeEmpty())
		})
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
