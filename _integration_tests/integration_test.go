package tests

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
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

func TestDefaultMedia(t *testing.T) {
	defer os.RemoveAll("./media/main.go")
	defer os.RemoveAll("./media/tool")
	if err := goagen("./media", "bootstrap", "-d", "github.com/goadesign/goa/_integration_tests/media/design"); err != nil {
		t.Error(err.Error())
	}
	if err := gobuild("./media"); err != nil {
		t.Error(err.Error())
	}
	b, err := ioutil.ReadFile("./media/app/contexts.go")
	if err != nil {
		t.Fatal("failed to load contexts.go")
	}
	expected := `// CreateGreetingPayload is the Greeting create action payload.
type CreateGreetingPayload struct {
	// A required string field in the parent type.
	Message string ` + "`" + `form:"message" json:"message" xml:"message"` + "`" + `
	// An optional boolean field in the parent type.
	ParentOptional *bool ` + "`" + `form:"parent_optional,omitempty" json:"parent_optional,omitempty" xml:"parent_optional,omitempty"` + "`" + `
}
`
	if !strings.Contains(string(b), expected) {
		t.Errorf("DefaultMedia attribute definitions reference failed. Generated context:\n%s", string(b))
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
