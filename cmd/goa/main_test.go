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
		usageCalled  bool
		cmds         []string
		path, output string
		debug        bool
	)

	usage = func() { usageCalled = true }
	gen = func(c []string, p, o string, d bool) { cmds, path, output, debug = c, p, o, d }
	defer func() {
		usage = help
		gen = generate
	}()

	cases := map[string]struct {
		CmdLine          string
		ExpectedUsage    bool
		ExpectedCommands []string
		ExpectedPath     string
		ExpectedOutput   string
		ExpectedDebug    bool
	}{
		"gen": {"gen " + testPkg, false, []string{"client", "openapi", "server"}, testPkg, ".", false},

		"invalid":     {"invalid " + testPkg, true, nil, "", ".", false},
		"empty":       {"", true, nil, "", ".", false},
		"invalid gen": {"invalid gen" + testPkg, true, nil, "", ".", false},

		"output":       {"gen " + testPkg + " -output " + testOutput, false, []string{"client", "openapi", "server"}, testPkg, testOutput, false},
		"output short": {"gen " + testPkg + " -o " + testOutput, false, []string{"client", "openapi", "server"}, testPkg, testOutput, false},

		"debug": {"gen " + testPkg + " -debug", false, []string{"client", "openapi", "server"}, testPkg, ".", true},
	}

	for k, c := range cases {
		{
			args := strings.Split(c.CmdLine, " ")
			os.Args = append([]string{"goa"}, args...)
			usageCalled = false
			cmds = nil
			path = ""
			output = ""
			debug = false
		}

		main()

		if usageCalled != c.ExpectedUsage {
			t.Errorf("%s: Expected usage to be %v but got %v", k, c.ExpectedUsage, usageCalled)
		}
		if len(cmds) != len(c.ExpectedCommands) {
			t.Errorf("%s: Expected %d commands but got %d: %s", k, len(c.ExpectedCommands), len(cmds), strings.Join(cmds, ", "))
		} else {
			for i, cmd := range cmds {
				if cmd != c.ExpectedCommands[i] {
					t.Errorf("%s: Expected command at index %d to be %s but got %s", k, i, c.ExpectedCommands[i], cmds[i])
				}
			}
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
