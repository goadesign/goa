package genclient_test

import (
	"github.com/goadesign/goa/goagen/gen_main"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("RegisterFlags", func() {
	const testCmd = "testCmd"
	var appCmd *genmain.Command
	var root *cobra.Command

	Context("using fake registry", func() {
		BeforeEach(func() {
			root = &cobra.Command{}
			appCmd = genmain.NewCommand()
		})

		JustBeforeEach(func() {
			appCmd.RegisterFlags(root)
		})

		It("registers the flags", func() {
			f := root.Flags().Lookup("name")
			Ω(f).ShouldNot(BeNil())
		})
	})
})
