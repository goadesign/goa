package codegen_test

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"runtime"

	"github.com/goadesign/goa/goagen/codegen"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helpers", func() {
	Describe("KebabCase", func() {
		It("should change uppercase letters to lowercase letters", func() {
			Expect(codegen.KebabCase("test-B")).To(Equal("test-b"))
			Expect(codegen.KebabCase("teste")).To(Equal("teste"))
		})

		It("should not add a dash before an abbreviation or acronym", func() {
			Expect(codegen.KebabCase("testABC")).To(Equal("testabc"))
		})

		It("should add a dash before a title", func() {
			Expect(codegen.KebabCase("testAa")).To(Equal("test-aa"))
			Expect(codegen.KebabCase("testAbc")).To(Equal("test-abc"))
		})

		It("should replace underscores to dashes", func() {
			Expect(codegen.KebabCase("test_cA")).To(Equal("test-ca"))
			Expect(codegen.KebabCase("test_D")).To(Equal("test-d"))
		})
	})

	Describe("CommandLine", func() {
		Context("with exported GOPATH", func() {
			oldGOPATH, oldArgs := build.Default.GOPATH, os.Args
			BeforeEach(func() {
				os.Setenv("GOPATH", "/xx")
			})
			AfterEach(func() {
				os.Setenv("GOPATH", oldGOPATH)
				os.Args = oldArgs
			})

			It("should not touch free arguments", func() {
				os.Args = []string{"foo", "/xx/bar/xx/42"}

				Expect(codegen.CommandLine()).To(Equal("$ foo /xx/bar/xx/42"))
			})

			It("should replace GOPATH one match only in a long option", func() {
				os.Args = []string{"foo", "--opt=/xx/bar/xx/42"}

				Expect(codegen.CommandLine()).To(Equal("$ foo\n\t--opt=$(GOPATH)/bar/xx/42"))
			})

			It("should not replace GOPATH if a match is not at the beginning of a long option", func() {
				os.Args = []string{"foo", "--opt=/bar/xx/42"}

				Expect(codegen.CommandLine()).To(Equal("$ foo\n\t--opt=/bar/xx/42"))
			})
		})

		Context("with default GOPATH", func() {
			oldGOPATH, oldArgs := build.Default.GOPATH, os.Args
			BeforeEach(func() {
				os.Setenv("GOPATH", defaultGOPATH()) // Simulate a situation with no GOPATH exported.
			})
			AfterEach(func() {
				os.Setenv("GOPATH", oldGOPATH)
				os.Args = oldArgs
			})

			It("should not touch free arguments", func() {
				os.Args = []string{"foo", "/xx/bar/xx/42"}

				Expect(codegen.CommandLine()).To(Equal("$ foo /xx/bar/xx/42"))
			})

			It("should replace GOPATH one match only in a long option", func() {
				os.Args = []string{"foo", fmt.Sprintf("--opt=%s/bar/xx/42", defaultGOPATH())}

				Expect(codegen.CommandLine()).To(Equal("$ foo\n\t--opt=$(GOPATH)/bar/xx/42"))
			})
		})
	})
})

// Copied from go/build/build.go
func defaultGOPATH() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	if home := os.Getenv(env); home != "" {
		def := filepath.Join(home, "go")
		if filepath.Clean(def) == filepath.Clean(runtime.GOROOT()) {
			// Don't set the default GOPATH to GOROOT,
			// as that will trigger warnings from the go tool.
			return ""
		}
		return def
	}
	return ""
}
