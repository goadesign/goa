package genapp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/goagen"
	"github.com/raphael/goa/goagen/gen_app"
	"gopkg.in/alecthomas/kingpin.v2"
)

// FakeRegistry captures flags defined by RegisterFlags.
type FakeRegistry struct {
	// Flags keeps track of all registered flags. It indexes their
	// descriptions by name.
	Flags map[string]string
}

// Flag implement FlagRegistry
func (f *FakeRegistry) Flag(n, h string) *kingpin.FlagClause {
	f.Flags[n] = h
	return new(kingpin.FlagClause)
}

var _ = Describe("RegisterFlags", func() {
	const testCmd = "testCmd"
	var appCmd *genapp.Command

	Context("using fake registry", func() {
		var reg *FakeRegistry

		BeforeEach(func() {
			reg = &FakeRegistry{Flags: make(map[string]string)}
			appCmd = genapp.NewCommand()
		})

		JustBeforeEach(func() {
			appCmd.RegisterFlags(reg)
		})

		It("registers the required flags", func() {
			_, ok := reg.Flags["target"]
			Ω(ok).Should(BeTrue())
		})
	})

	Context("with command line flags", func() {
		var kapp *kingpin.Application
		var cmd *kingpin.CmdClause
		const flagVal = "testme"
		var args []string
		var parsedCmd string

		BeforeEach(func() {
			kapp = kingpin.New("test", "test")
			cmd = kapp.Command("testCmd", "testCmd")
			args = []string{testCmd, "-o" + flagVal, "-d=design", "--target=dummy"}
		})

		JustBeforeEach(func() {
			goagen.RegisterFlags(cmd)
			appCmd.RegisterFlags(cmd)
			var err error
			parsedCmd, err = kapp.Parse(args)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("parses the default flags", func() {
			Ω(parsedCmd).Should(Equal(testCmd))
			Ω(goagen.OutputDir).Should(Equal(flagVal))
		})
	})
})
