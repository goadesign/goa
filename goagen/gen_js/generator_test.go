package genjs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"time"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/gen_js"
	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_js/test_"

	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + outDir, "--design=foo", "--host=baz", "--version=" + version.String()}
	})

	JustBeforeEach(func() {
		files, genErr = genjs.Generate()
	})

	AfterEach(func() {
		os.RemoveAll(outDir)
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
			}
		})

		It("generates a dummy js", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(3))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "js", "client.js"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 13))
		})
	})

	Context("with an example action with query parameters", func() {
		BeforeEach(func() {
			action := &design.ActionDefinition{
				Name: "show",
				Routes: []*design.RouteDefinition{{
					Verb: "GET",
					Path: "/",
				}},
				Params: &design.AttributeDefinition{
					Type: design.Object{
						"query": {Type: design.String},
					},
				},
				QueryParams: &design.AttributeDefinition{
					Type: design.Object{
						"query": {Type: design.String},
					},
				},
			}
			design.Design = &design.APIDefinition{
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
				Resources: map[string]*design.ResourceDefinition{
					"bottle": {
						Name: "bottle",
						Actions: map[string]*design.ActionDefinition{
							"show": action,
						},
					},
				},
			}
			action.Parent = design.Design.Resources["bottle"]
		})

		It("generates an example HTML", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(5))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "js", "index.html"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 13))
		})
	})
})

var _ = Describe("NewGenerator", func() {
	var generator *genjs.Generator

	var args = struct {
		api       *design.APIDefinition
		outDir    string
		timeout   time.Duration
		scheme    string
		host      string
		noExample bool
	}{
		api: &design.APIDefinition{
			Name: "test api",
		},
		outDir:    "out_dir",
		timeout:   time.Millisecond * 500,
		scheme:    "http",
		host:      "localhost",
		noExample: true,
	}

	Context("with options all options set", func() {
		BeforeEach(func() {

			generator = genjs.NewGenerator(
				genjs.API(args.api),
				genjs.OutDir(args.outDir),
				genjs.Timeout(args.timeout),
				genjs.Scheme(args.scheme),
				genjs.Host(args.host),
				genjs.NoExample(args.noExample),
			)
		})

		It("has all public properties set with expected value", func() {
			Ω(generator).ShouldNot(BeNil())
			Ω(generator.API.Name).Should(Equal(args.api.Name))
			Ω(generator.OutDir).Should(Equal(args.outDir))
			Ω(generator.Timeout).Should(Equal(args.timeout))
			Ω(generator.Scheme).Should(Equal(args.scheme))
			Ω(generator.Host).Should(Equal(args.host))
			Ω(generator.NoExample).Should(Equal(args.noExample))
		})

	})
})
