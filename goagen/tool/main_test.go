package main

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("goagen", func() {
	Context("with a valid command line", func() {
		const cmdName = "app"

		BeforeEach(func() {
			os.Args = []string{"goagen", cmdName, "--target", "target", "-o", ".", "--design", "d"}
		})

		It("command returns the correct command", func() {
			cmd := command()
			Ω(cmd).ShouldNot(BeNil())
			Ω(cmd.Name()).Should(Equal(cmdName))
		})
	})
})
