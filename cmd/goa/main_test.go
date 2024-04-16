package main

import (
	"os"
	"strings"
	"testing"
)

func TestCmdLine(t *testing.T) {
	const (
		testPkg    = "/test"
		testOutput = "testOutput"
	)
	var (
		usageCalled          bool
		cmd                  string
		path, output         string
		debug                bool
		disableVersionString bool
	)

	usage = func() { usageCalled = true }
	gen = func(c string, p, o string, d bool, v bool) error {
		cmd, path, output, debug, disableVersionString = c, p, o, d, v
		return nil
	}
	defer func() {
		usage = help
		gen = generate
	}()

	cases := map[string]struct {
		CmdLine                      string
		ExpectedUsage                bool
		ExpectedCommand              string
		ExpectedPath                 string
		ExpectedOutput               string
		ExpectedDebug                bool
		ExpectedDisableVersionString bool
	}{
		"gen": {"gen " + testPkg, false, "gen", testPkg, ".", false, false},

		"invalid":     {"invalid " + testPkg, true, "", "", ".", false, false},
		"empty":       {"", true, "", "", ".", false, false},
		"invalid gen": {"invalid gen" + testPkg, true, "", "", ".", false, false},

		"output":       {"gen " + testPkg + " -output " + testOutput, false, "gen", testPkg, testOutput, false, false},
		"output short": {"gen " + testPkg + " -o " + testOutput, false, "gen", testPkg, testOutput, false, false},

		"debug": {"gen " + testPkg + " -debug", false, "gen", testPkg, ".", true, false},

		"disable version": {"gen " + testPkg + " -disableVersionString", false, "gen", testPkg, ".", false, true},
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
		if disableVersionString != c.ExpectedDisableVersionString {
			t.Errorf("%s: Expected disableVersionString to be %v but got %v", k, c.ExpectedDisableVersionString, disableVersionString)
		}
	}
}
