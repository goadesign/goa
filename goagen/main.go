package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/goadesign/goa/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	// These are packages required by the generated code but not by goagen.
	// We list them here so that `go get` picks them up.
	_ "gopkg.in/yaml.v2"
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
		designPkg string
		debug     bool
	)

	rootCmd.PersistentFlags().StringP("out", "o", ".", "output directory")
	rootCmd.PersistentFlags().StringVarP(&designPkg, "design", "d", "", "design package import path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode, does not cleanup temporary files.")

	// versionCmd implements the "version" command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of goagen",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("goagen " + version.String() + "\nThe goa generation tool.")
		},
	}
	rootCmd.AddCommand(versionCmd)

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
		force, regen bool
	)
	mainCmd := &cobra.Command{
		Use:   "main",
		Short: "Generate application scaffolding",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genmain", c) },
	}
	mainCmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	mainCmd.Flags().BoolVar(&regen, "regen", false, "regenerate scaffolding, maintaining controller implementations")
	rootCmd.AddCommand(mainCmd)

	// clientCmd implements the "client" command.
	var (
		toolDir, tool string
		notool        bool
	)
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Generate client package and tool",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("genclient", c) },
	}
	clientCmd.Flags().StringVar(&pkg, "pkg", "client", "Name of generated client Go package")
	clientCmd.Flags().StringVar(&toolDir, "tooldir", "tool", "Name of generated tool directory")
	clientCmd.Flags().StringVar(&tool, "tool", "[API-name]-cli", "Name of generated tool")
	clientCmd.Flags().BoolVar(&notool, "notool", false, "Prevent generation of cli tool")
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
		Run:   func(c *cobra.Command, args []string) { files, err = runGen(c, args) },
	}
	genCmd.Flags().StringVar(&pkgPath, "pkg-path", "", "Package import path of generator. The package must implement the Generate global function.")
	// stop parsing arguments after -- to prevent an unknown flag error
	// this also means custom arguments (after --) should be the last arguments
	genCmd.Flags().SetInterspersed(false)
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
	bootCmd.Flags().AddFlagSet(appCmd.Flags())
	bootCmd.Flags().AddFlagSet(mainCmd.Flags())
	bootCmd.Flags().AddFlagSet(clientCmd.Flags())
	bootCmd.Flags().AddFlagSet(swaggerCmd.Flags())
	rootCmd.AddCommand(bootCmd)

	// controllerCmd implements the "controller" command.
	var (
		res, appPkg string
	)
	controllerCmd := &cobra.Command{
		Use:   "controller",
		Short: "Generate controller scaffolding",
		Run:   func(c *cobra.Command, _ []string) { files, err = run("gencontroller", c) },
	}
	controllerCmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	controllerCmd.Flags().BoolVar(&regen, "regen", false, "regenerate scaffolding, maintaining controller implementations")
	controllerCmd.Flags().StringVar(&res, "res", "", "name of the `resource` to generate the controller for, generate all if not specified")
	controllerCmd.Flags().StringVar(&pkg, "pkg", "main", "name of the generated controller `package`")
	controllerCmd.Flags().StringVar(&appPkg, "app-pkg", "app", "`import path` of Go package generated with 'goagen app', may be relative to output")
	rootCmd.AddCommand(controllerCmd)

	// cmdsCmd implements the commands command
	// It lists all the commands and flags in JSON to enable shell integrations.
	cmdsCmd := &cobra.Command{
		Use:   "commands",
		Short: "Lists all commands and flags in JSON",
		Run:   func(c *cobra.Command, _ []string) { runCommands(rootCmd) },
	}
	rootCmd.AddCommand(cmdsCmd)

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
	return generate(pkgName, pkgPath, c, nil)
}

func runGen(c *cobra.Command, args []string) ([]string, error) {
	pkgPath := c.Flag("pkg-path").Value.String()
	pkgSrcPath, err := codegen.PackageSourcePath(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin package import path: %s", err)
	}
	pkgName, err := codegen.PackageName(pkgSrcPath)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin package import path: %s", err)
	}
	return generate(pkgName, pkgPath, c, args)
}

func generate(pkgName, pkgPath string, c *cobra.Command, args []string) ([]string, error) {
	m := make(map[string]string)
	c.Flags().Visit(func(f *pflag.Flag) {
		if f.Name != "pkg-path" {
			m[f.Name] = f.Value.String()
		}
	})
	if _, ok := m["out"]; !ok {
		m["out"] = c.Flag("out").DefValue
	}
	// turn "out" into an absolute path
	var err error
	m["out"], err = filepath.Abs(m["out"])
	if err != nil {
		return nil, err
	}

	gen, err := meta.NewGenerator(
		pkgName+".Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport(pkgPath)},
		m,
		args,
	)
	if err != nil {
		return nil, err
	}
	return gen.Generate()
}

type (
	rootCommand struct {
		Name     string     `json:"name"`
		Commands []*command `json:"commands"`
		Flags    []*flag    `json:"flags"`
	}

	flag struct {
		Long        string `json:"long,omitempty"`
		Short       string `json:"short,omitempty"`
		Description string `json:"description,omitempty"`
		Argument    string `json:"argument,omitempty"`
		Required    bool   `json:"required"`
	}

	command struct {
		Name  string  `json:"name"`
		Flags []*flag `json:"flags,omitempty"`
	}
)

func runCommands(root *cobra.Command) {
	var (
		gblFlags []*flag
		cmds     []*command
	)
	root.Flags().VisitAll(func(fl *pflag.Flag) {
		gblFlags = append(gblFlags, flagJSON(fl))
	})
	cmds = make([]*command, len(root.Commands())-2)
	j := 0
	for _, cm := range root.Commands() {
		if cm.Name() == "help" || cm.Name() == "commands" {
			continue
		}
		cmds[j] = cmdJSON(cm, gblFlags)
		j++
	}
	rc := rootCommand{os.Args[0], cmds, gblFlags}
	b, _ := json.MarshalIndent(rc, "", "    ")
	fmt.Println(string(b))
}

// Lots of assumptions in here, it's OK for what we are doing
// Remember to update as goagen commands and flags evolve
//
// The flag argument values use variable names that cary semantic:
// $DIR for file system directories, $DESIGN_PKG for import path to Go goa design Go packages, $PKG
// for import path to any Go package.
func flagJSON(fl *pflag.Flag) *flag {
	f := &flag{Long: fl.Name, Short: fl.Shorthand, Description: fl.Usage}
	f.Required = fl.Name == "pkg-path" || fl.Name == "design"
	switch fl.Name {
	case "out":
		f.Argument = "$DIR"
	case "design":
		f.Argument = "$DESIGN_PKG"
	case "pkg-path":
		f.Argument = "$PKG"
	}
	return f
}

func cmdJSON(cm *cobra.Command, flags []*flag) *command {
	res := make([]*flag, len(flags))
	for i, fl := range flags {
		f := *fl
		res[i] = &f
	}
	cm.Flags().VisitAll(func(fl *pflag.Flag) {
		res = append(res, flagJSON(fl))
	})
	return &command{cm.Name(), res}
}
