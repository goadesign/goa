package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	var (
		files            []string
		err              error
		terminatedByUser bool
	)

	// rootCmd is the base command used when goagen is called with no argument.
	rootCmd := &cobra.Command{
		Use:   "goagen",
		Short: "goa code generation tool",
		Long: `The goagen tool generates artifacts from a goa service design package.

Each command supported by the tool produces a specific type of artifacts. For example
the "app" command generates the code that supports the service controllers.

The "bootstrap" command runs the "app", "main", "client" and "swagger" commands generating the
controllers supporting code and main skeleton code (if not already present) as well as a client
package and tool and the Swagger specification for the API.
`}
	var (
		cwd, designPkg string
		debug          bool
	)
	cwd, err = os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	rootCmd.PersistentFlags().StringVarP(&cwd, "out", "o", cwd, "output directory")
	rootCmd.PersistentFlags().StringVarP(&designPkg, "design", "d", "", "design package import path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode, does not cleanup temporary files.")

	// appCmd implements the "app" command.
	var (
		pkg    string
		notest bool
	)
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Generate application code",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genapp", c) },
	}
	appCmd.Flags().StringVar(&pkg, "pkg", "app", "Name of generated Go package containing controllers supporting code (contexts, media types, user types etc.)")
	appCmd.Flags().BoolVar(&notest, "notest", false, "Prevent generation of test helpers")
	rootCmd.AddCommand(appCmd)

	// mainCmd implements the "main" command.
	var (
		force bool
	)
	mainCmd := &cobra.Command{
		Use:   "main",
		Short: "Generate application scaffolding",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genmain", c) },
	}
	mainCmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	rootCmd.AddCommand(mainCmd)

	// clientCmd implements the "client" command.
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Generate client package and tool",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genclient", c) },
	}
	clientCmd.Flags().StringVar(&pkg, "pkg", "client", "Name of generated client Go package")
	rootCmd.AddCommand(clientCmd)

	// swaggerCmd implements the "swagger" command.
	swaggerCmd := &cobra.Command{
		Use:   "swagger",
		Short: "Generate Swagger",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genswagger", c) },
	}
	rootCmd.AddCommand(swaggerCmd)

	// jsCmd implements the "js" command.
	var (
		timeout      = time.Duration(20) * time.Second
		scheme, host string
		noexample    bool
	)
	jsCmd := &cobra.Command{
		Use:   "js",
		Short: "Generate JavaScript client",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genjs", c) },
	}
	jsCmd.Flags().DurationVar(&timeout, "timeout", timeout, `the duration before the request times out.`)
	jsCmd.Flags().StringVar(&scheme, "scheme", "", `the URL scheme used to make requests to the API, defaults to the scheme defined in the API design if any.`)
	jsCmd.Flags().StringVar(&host, "host", "", `the API hostname, defaults to the hostname defined in the API design if any`)
	jsCmd.Flags().BoolVar(&noexample, "noexample", false, `Skip generation of example HTML and controller`)
	rootCmd.AddCommand(jsCmd)

	// schemaCmd implements the "schema" command.
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Generate JSON Schema",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genschema", c) },
	}
	rootCmd.AddCommand(schemaCmd)

	// genCmd implements the "gen" command.
	var (
		pkgPath string
	)
	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Run third-party generator",
		Run:   func(c *cobra.Command, _ []string) { files, err = runGen(c) },
	}
	genCmd.Flags().StringVar(&pkgPath, "pkg-path", "", "Package import path of generator. The package must implement the Generate global function.")
	rootCmd.AddCommand(genCmd)

	// boostrapCmd implements the "bootstrap" command.
	bootCmd := &cobra.Command{
		Use:   "bootstrap",
		Short: `Equivalent to running the "app", "main", "client" and "swagger" commands.`,
		Run: func(c *cobra.Command, a []string) {
			appCmd.Run(c, a)
			if err != nil {
				return
			}
			prev := files

			mainCmd.Run(c, a)
			if err != nil {
				return
			}
			prev = append(prev, files...)

			clientCmd.Run(c, a)
			if err != nil {
				return
			}
			prev = append(prev, files...)

			swaggerCmd.Run(c, a)
			files = append(prev, files...)
		},
	}
	rootCmd.AddCommand(bootCmd)

	// Now proceed with code generation
	cleanup := func() {
		for _, f := range files {
			os.RemoveAll(f)
		}
	}

	go utils.Catch(nil, func() {
		terminatedByUser = true
	})

	rootCmd.Execute()

	if terminatedByUser {
		cleanup()
		return
	}

	if err != nil {
		cleanup()
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	rels := make([]string, len(files))
	cd, _ := os.Getwd()
	for i, f := range files {
		r, err := filepath.Rel(cd, f)
		if err == nil {
			rels[i] = r
		} else {
			rels[i] = f
		}
	}
	fmt.Println(strings.Join(rels, "\n"))
}

func run(pkg string, c *cobra.Command) ([]string, error) {
	pkgPath := fmt.Sprintf("github.com/goadesign/goa/goagen/gen_%s", pkg[3:])
	pkgSrcPath, err := codegen.PackageSourcePath(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin package import path: %s", err)
	}
	pkgName, err := codegen.PackageName(pkgSrcPath)
	if err != nil {
		return nil, fmt.Errorf("invalid package import path: %s", err)
	}
	return generate(pkgName, pkgPath, c)
}

func runGen(c *cobra.Command) ([]string, error) {
	pkgPath := c.Flag("pkg-path").Value.String()
	pkgSrcPath, err := codegen.PackageSourcePath(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin package import path: %s", err)
	}
	pkgName, err := codegen.PackageName(pkgSrcPath)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin package import path: %s", err)
	}
	return generate(pkgName, pkgPath, c)
}

func generate(pkgName, pkgPath string, c *cobra.Command) ([]string, error) {
	m := make(map[string]string)
	c.Flags().Visit(func(f *pflag.Flag) {
		if f.Name != "pkg-path" {
			m[f.Name] = f.Value.String()
		}
	})
	if _, ok := m["out"]; !ok {
		m["out"] = c.Flag("out").DefValue
	}
	gen, err := meta.NewGenerator(
		pkgName+".Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport(pkgPath)},
		m,
	)
	if err != nil {
		return nil, err
	}
	return gen.Generate()
}
