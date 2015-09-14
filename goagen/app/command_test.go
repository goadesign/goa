package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/goagen/app"
	"gopkg.in/alecthomas/kingpin.v2"
)

var _ = Describe("RegisterFlags", func() {
	const testCmd = "testCmd"
	var appCmd *app.Command
	var kapp *kingpin.Application
	var cmd *kingpin.CmdClause

	BeforeEach(func() {
		appCmd = app.NewCommand()
		kapp = kingpin.New("test", "fake")
		cmd = kapp.Command(testCmd, "fake too")
	})

	JustBeforeEach(func() {
		appCmd.RegisterFlags(cmd)
	})

	It("registers the default flags", func() {
		Ω(appCmd.Generator.Flags).Should(HaveKey("OutDir"))
	})

	Context("with command line flags", func() {
		const flagVal = "testme"
		var args []string
		var parsedCmd string

		BeforeEach(func() {
			args = []string{testCmd, "--out=" + flagVal, "--package=dummy"}
		})

		JustBeforeEach(func() {
			var err error
			parsedCmd, err = kapp.Parse(args)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("parses the default flags", func() {
			Ω(parsedCmd).Should(Equal(testCmd))
			Ω(appCmd.Generator.Flags).Should(HaveKey("OutDir"))
			Ω(appCmd.Generator.Flags["OutDir"]).ShouldNot(BeNil())
			Ω(*appCmd.Generator.Flags["OutDir"]).Should(Equal(flagVal))
		})
	})
})
