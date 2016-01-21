package genclient_test

import (
	"github.com/goadesign/goa/goagen/gen_main"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	var appCmd *genmain.Command

	Context("using fake registry", func() {
		var reg *FakeRegistry

		BeforeEach(func() {
			reg = &FakeRegistry{Flags: make(map[string]string)}
			appCmd = genmain.NewCommand()
		})

		JustBeforeEach(func() {
			appCmd.RegisterFlags(reg)
		})

		It("registers the flags", func() {
			_, ok := reg.Flags["name"]
			Ω(ok).Should(BeTrue())
		})
	})
})
