package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdLine(t *testing.T) {
	const (
		testPkg    = "/test"
		testOutput = "testOutput"
	)
	var (
		usageCalled  bool
		cmd          string
		path, output string
		debug        bool
	)

	usage = func() { usageCalled = true }
	gen = func(c string, p, o string, d bool) error { cmd, path, output, debug = c, p, o, d; return nil }
	defer func() {
		usage = help
		gen = generate
	}()

	cases := map[string]struct {
		CmdLine         string
		ExpectedUsage   bool
		ExpectedCommand string
		ExpectedPath    string
		ExpectedOutput  string
		ExpectedDebug   bool
	}{
		"gen": {"gen " + testPkg, false, "gen", testPkg, ".", false},

		"invalid":     {"invalid " + testPkg, true, "", "", "", false},
		"empty":       {"", true, "", "", "", false},
		"invalid gen": {"invalid gen" + testPkg, true, "", "", "", false},

		"output":       {"gen " + testPkg + " -output " + testOutput, false, "gen", testPkg, testOutput, false},
		"output short": {"gen " + testPkg + " -o " + testOutput, false, "gen", testPkg, testOutput, false},

		"debug": {"gen " + testPkg + " -debug", false, "gen", testPkg, ".", true},
	}

	for k, c := range cases {
		{
			args := strings.Split(c.CmdLine, " ")
			os.Args = append([]string{"goa"}, args...)
			usageCalled = false
			cmd = ""
			path = ""
			output = ""
			debug = false
		}

		main()

		if usageCalled != c.ExpectedUsage {
			t.Errorf("%s: Expected usage to be %v but got %v", k, c.ExpectedUsage, usageCalled)
		}
		if cmd != c.ExpectedCommand {
			t.Errorf("%s: Expected command to be %s but got %s", k, c.ExpectedCommand, cmd)
		}
		if path != c.ExpectedPath {
			t.Errorf("%s: Expected path to be %s but got %s", k, c.ExpectedPath, path)
		}
		if output != c.ExpectedOutput {
			t.Errorf("%s: Expected output to be %s but got %s", k, c.ExpectedOutput, output)
		}
		if debug != c.ExpectedDebug {
			t.Errorf("%s: Expected debug to be %v but got %v", k, c.ExpectedDebug, debug)
		}
	}
}

func TestGeneratorWriteCreatesAbsoluteTemporaryDirectory(t *testing.T) {
	generator := &Generator{
		Command:       "gen",
		DesignPath:    "example.com/design",
		Output:        t.TempDir(),
		DesignVersion: 3,
		bin:           "goa",
	}
	t.Cleanup(generator.Remove)

	if err := generator.Write(false); err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if !filepath.IsAbs(generator.tmpDir) {
		t.Errorf("Write() temporary directory = %q, want absolute path", generator.tmpDir)
	}
}
