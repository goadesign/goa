package genapp_test

import (
	"os"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("RegisterFlags", func() {
	const testCmd = "testCmd"
	var appCmd *genapp.Command
	var root *cobra.Command

	Context("registering flags", func() {
		BeforeEach(func() {
			appCmd = genapp.NewCommand()
			root = &cobra.Command{}
		})

		JustBeforeEach(func() {
			appCmd.RegisterFlags(root)
		})

		It("registers the required flags", func() {
			f := root.Flags().Lookup("pkg")
			Ω(f).ShouldNot(BeNil())
		})
	})

	Context("with command line flags", func() {
		var root *cobra.Command
		const flagVal = "testme"
		var args []string

		BeforeEach(func() {
			root = &cobra.Command{Use: "testCmd"}
			args = []string{os.Args[0], testCmd, "-o" + flagVal, "-d=design", "--pkg=dummy"}
		})

		JustBeforeEach(func() {
			codegen.RegisterFlags(root)
			appCmd.RegisterFlags(root)
			os.Args = args
		})

		It("parses the default flags", func() {
			err := root.Execute()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(codegen.OutputDir).Should(Equal(flagVal))
		})
	})
})
