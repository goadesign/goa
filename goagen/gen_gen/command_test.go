package gengen_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_gen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("RegisterFlags", func() {
	const testCmd = "testCmd"
	var genCmd *gengen.Command
	var root *cobra.Command

	Context("using fake registry", func() {
		BeforeEach(func() {
			root = &cobra.Command{}
			genCmd = gengen.NewCommand()
		})

		JustBeforeEach(func() {
			genCmd.RegisterFlags(root)
		})

		It("registers the flags", func() {
			f := root.Flags().Lookup("pkg-path")
			Ω(f).ShouldNot(BeNil())
			f = root.Flags().Lookup("pkg-name")
			Ω(f).ShouldNot(BeNil())
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
			ioutil.WriteFile(filepath.Join(tmpPkg, "dummy.go"), []byte(dummyGenSrc), 0644)
			genCmd = gengen.NewCommand()
			oldPkgPath = gengen.GenPkgPath
			oldPkgName = gengen.GenPkgName
			oldOutputDir = codegen.OutputDir
			oldDesignPackagePath = codegen.DesignPackagePath
			gengen.GenPkgPath, err = filepath.Rel(filepath.Join(gopath, "src"), tmpPkg)
			Ω(err).ShouldNot(HaveOccurred())
			gengen.GenPkgName = "dummy"
			codegen.OutputDir = tmpPkg
			codegen.DesignPackagePath = "github.com/goadesign/goa-cellar/design"
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

func Generate(roots []interface{}) ([]string, error) {
	return []string{"worked"}, nil
}
`
