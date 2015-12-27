package gengen_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/gen_gen"
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
	var genCmd *gengen.Command

	Context("using fake registry", func() {
		var reg *FakeRegistry

		BeforeEach(func() {
			reg = &FakeRegistry{Flags: make(map[string]string)}
			genCmd = gengen.NewCommand()
		})

		JustBeforeEach(func() {
			genCmd.RegisterFlags(reg)
		})

		It("registers the flags", func() {
			_, ok := reg.Flags["pkg-path"]
			Ω(ok).Should(BeTrue())
			_, ok = reg.Flags["pkg-name"]
			Ω(ok).Should(BeTrue())
		})
	})

	Context("with a dummy generator", func() {
		var tmpPkg string
		var genCmd *gengen.Command
		var oldPkgPath, oldPkgName string
		var oldOutputDir, oldDesignPackagePath string

		BeforeEach(func() {
			var err error
			gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
			tmpPkg, err = ioutil.TempDir(filepath.Join(gopath, "src"), "goagen")
			Ω(err).ShouldNot(HaveOccurred())
			ioutil.WriteFile(filepath.Join(tmpPkg, "dummy.go"), []byte(dummyGenSrc), 0755)
			genCmd = gengen.NewCommand()
			oldPkgPath = gengen.GenPkgPath
			oldPkgName = gengen.GenPkgName
			oldOutputDir = codegen.OutputDir
			oldDesignPackagePath = codegen.DesignPackagePath
			gengen.GenPkgPath, err = filepath.Rel(filepath.Join(gopath, "src"), tmpPkg)
			Ω(err).ShouldNot(HaveOccurred())
			gengen.GenPkgName = "dummy"
			codegen.OutputDir = tmpPkg
			codegen.DesignPackagePath = "github.com/raphael/goa/examples/cellar/design"
		})

		It("invokes the generator", func() {
			files, err := genCmd.Run()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(files).Should(Equal([]string{"worked"}))
		})

		AfterEach(func() {
			gengen.GenPkgPath = gengen.GenPkgPath
			gengen.GenPkgName = gengen.GenPkgName
			codegen.OutputDir = oldOutputDir
			codegen.DesignPackagePath = oldDesignPackagePath
			os.RemoveAll(tmpPkg)
		})
	})
})

const dummyGenSrc = `package dummy

import "github.com/raphael/goa/design"

func Generate(api *design.APIDefinition) ([]string, error) {
	return []string{"worked"}, nil
}
`
