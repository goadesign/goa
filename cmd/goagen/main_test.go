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
		gens, debug  bool
	)

	usage = func() { usageCalled = true }
	gen = func(c []string, p, o string, g, d bool) { cmds, path, output, gens, debug = c, p, o, g, d }
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
		ExpectedGens     bool
		ExpectedDebug    bool
	}{
		"client":                {"client " + testPkg, false, []string{"client"}, testPkg, ".", false, false},
		"client server":         {"client server " + testPkg, false, []string{"client", "server"}, testPkg, ".", false, false},
		"client openapi":        {"client openapi " + testPkg, false, []string{"client", "openapi"}, testPkg, ".", false, false},
		"client server openapi": {"client server openapi " + testPkg, false, []string{"client", "openapi", "server"}, testPkg, ".", false, false},
		"client openapi server": {"client openapi server " + testPkg, false, []string{"client", "openapi", "server"}, testPkg, ".", false, false},
		"client client":         {"client client " + testPkg, false, []string{"client"}, testPkg, ".", false, false},

		"server":                {"server " + testPkg, false, []string{"server"}, testPkg, ".", false, false},
		"server client":         {"server client " + testPkg, false, []string{"client", "server"}, testPkg, ".", false, false},
		"server openapi":        {"server openapi " + testPkg, false, []string{"openapi", "server"}, testPkg, ".", false, false},
		"server client openapi": {"server client openapi " + testPkg, false, []string{"client", "openapi", "server"}, testPkg, ".", false, false},
		"server openapi client": {"server openapi client " + testPkg, false, []string{"client", "openapi", "server"}, testPkg, ".", false, false},
		"server server":         {"server server " + testPkg, false, []string{"server"}, testPkg, ".", false, false},

		"openapi":               {"openapi " + testPkg, false, []string{"openapi"}, testPkg, ".", false, false},
		"openapi client":        {"openapi client " + testPkg, false, []string{"client", "openapi"}, testPkg, ".", false, false},
		"openapi server":        {"openapi server " + testPkg, false, []string{"openapi", "server"}, testPkg, ".", false, false},
		"openapi client server": {"openapi client server " + testPkg, false, []string{"client", "openapi", "server"}, testPkg, ".", false, false},
		"openapi server client": {"openapi server client " + testPkg, false, []string{"client", "openapi", "server"}, testPkg, ".", false, false},
		"openapi openapi":       {"openapi openapi " + testPkg, false, []string{"openapi"}, testPkg, ".", false, false},

		"invalid":        {"invalid " + testPkg, true, nil, "", ".", false, false},
		"empty":          {"", true, nil, "", ".", false, false},
		"invalid client": {"invalid client " + testPkg, true, nil, "", ".", false, false},

		"output":       {"client " + testPkg + " -output " + testOutput, false, []string{"client"}, testPkg, testOutput, false, false},
		"output short": {"client " + testPkg + " -o " + testOutput, false, []string{"client"}, testPkg, testOutput, false, false},

		"gens":       {"client " + testPkg + " -scaffold", false, []string{"client"}, testPkg, ".", true, false},
		"gens short": {"client " + testPkg + " -s", false, []string{"client"}, testPkg, ".", true, false},

		"debug": {"client " + testPkg + " -debug", false, []string{"client"}, testPkg, ".", false, true},
	}

	for k, c := range cases {
		{
			args := strings.Split(c.CmdLine, " ")
			os.Args = append([]string{"goagen"}, args...)
			usageCalled = false
			cmds = nil
			path = ""
			output = ""
			gens = false
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
		if gens != c.ExpectedGens {
			t.Errorf("%s: Expected gens to be %v but got %v", k, c.ExpectedGens, gens)
		}
		if debug != c.ExpectedDebug {
			t.Errorf("%s: Expected debug to be %v but got %v", k, c.ExpectedDebug, debug)
		}
	}
}
