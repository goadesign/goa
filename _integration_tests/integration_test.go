package tests

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path"
	"testing"
)

func TestBootstrapReadme(t *testing.T) {
	defer os.RemoveAll("./readme/main.go")
	defer os.RemoveAll("./readme/tool")
	if err := goagen("./readme", "bootstrap", "-d", "github.com/goadesign/goa/_integration_tests/readme/design"); err != nil {
		t.Error(err.Error())
	}
	if err := gobuild("./readme"); err != nil {
		t.Error(err.Error())
	}
}

func TestCellar(t *testing.T) {
	if err := os.MkdirAll("./goa-cellar", 0755); err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll("./goa-cellar")
	if err := goagen("./goa-cellar", "bootstrap", "-d", "github.com/goadesign/goa-cellar/design"); err != nil {
		t.Error(err.Error())
	}
	if err := gobuild("./goa-cellar"); err != nil {
		t.Error(err.Error())
	}
	if err := gobuild("./goa-cellar/tool/cellar-cli"); err != nil {
		t.Error(err.Error())
	}
}

func goagen(dir, command string, args ...string) error {
	pkg, err := build.Import("github.com/goadesign/goa/goagen", "", 0)
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "run")
	for _, f := range pkg.GoFiles {
		cmd.Args = append(cmd.Args, path.Join(pkg.Dir, f))
	}
	cmd.Dir = dir
	cmd.Args = append(cmd.Args, command)
	cmd.Args = append(cmd.Args, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", err.Error(), out)
	}
	return nil
}

func gobuild(dir string) error {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", err.Error(), out)
	}
	return nil
}
