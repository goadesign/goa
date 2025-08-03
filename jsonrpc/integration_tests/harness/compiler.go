package harness

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GenerateFromDSL generates code from a DSL string using the goa CLI tool.
// This approach allows for better isolation and parallel execution compared
// to in-process generation.
func GenerateFromDSL(ctx context.Context, outputDir string, dslCode string) error {

	// Create design directory
	designDir := filepath.Join(outputDir, "design")
	if err := os.MkdirAll(designDir, 0755); err != nil {
		return fmt.Errorf("failed to create design directory: %w", err)
	}

	// Write DSL to file
	if err := writeDSLFile(designDir, dslCode); err != nil {
		return fmt.Errorf("failed to write DSL file: %w", err)
	}

	// Initialize go module
	if err := initGoModule(ctx, outputDir, "testapp"); err != nil {
		return fmt.Errorf("failed to init module: %w", err)
	}

	// Run goa gen with context
	if err := runGoaCommand(ctx, outputDir, "gen", "testapp/design"); err != nil {
		return fmt.Errorf("goa gen failed: %w", err)
	}

	// Run goa example with context
	if err := runGoaCommand(ctx, outputDir, "example", "testapp/design"); err != nil {
		return fmt.Errorf("goa example failed: %w", err)
	}

	// Run go mod tidy to clean up
	if err := runGoModTidy(ctx, outputDir); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}

// writeDSLFile writes the DSL code to a Go file
func writeDSLFile(designDir string, dslCode string) error {
	content := fmt.Sprintf(`package design

import (
	. "goa.design/goa/v3/dsl"
)

func init() {
%s
}
`, dslCode)

	designFile := filepath.Join(designDir, "design.go")
	return os.WriteFile(designFile, []byte(content), 0644)
}

// runGoaCommand runs a goa command with proper context handling
func runGoaCommand(ctx context.Context, dir, command, designPath string) error {
	// Set a reasonable timeout for code generation
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "goa", command, designPath, "-o", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=on", "GOWORK=off")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("%s canceled: %w", command, ctx.Err())
		}
		return fmt.Errorf("%s failed: %w\nOutput: %s", command, err, output)
	}
	return nil
}

// initGoModule initializes a go module in the directory if needed
func initGoModule(ctx context.Context, dir, name string) error {

	// Check if go.mod already exists
	modPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(modPath); err == nil {
		return nil // Already exists
	}

	// Initialize module with timeout
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(initCtx, "go", "mod", "init", name)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("module init canceled: %w", ctx.Err())
		}
		return fmt.Errorf("go mod init failed: %w\nOutput: %s", err, output)
	}

	// Add replace directive BEFORE running go mod tidy
	if err := addLocalReplace(modPath); err != nil {
		return fmt.Errorf("failed to add local replace: %w", err)
	}

	return nil
}

// runGoModTidy runs go mod tidy with context support
func runGoModTidy(ctx context.Context, dir string) error {
	tidyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(tidyCtx, "go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("go mod tidy canceled: %w", ctx.Err())
		}
		return fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// addLocalReplace adds a replace directive for the local goa module to the
// go.mod file. This ensures tests use the development version of Goa from
// the local filesystem rather than downloading from the module proxy.
func addLocalReplace(modPath string) error {
	// Read current go.mod
	content, err := os.ReadFile(modPath)
	if err != nil {
		return err
	}

	// Get the directory containing the go.mod file
	modDir := filepath.Dir(modPath)

	// Make modDir absolute for consistent path calculations
	absModDir, err := filepath.Abs(modDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Use git to find the repository root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = absModDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find git root: %w", err)
	}
	gitRoot := strings.TrimSpace(string(output))

	// Calculate relative path from modDir to gitRoot
	relPath, err := filepath.Rel(absModDir, gitRoot)
	if err != nil {
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	// Create replace directive
	replaceDir := fmt.Sprintf("replace goa.design/goa/v3 => %s", relPath)
	if !strings.Contains(string(content), "replace goa.design/goa/v3") {
		content = append(content, []byte("\n"+replaceDir+"\n")...)
		if err := os.WriteFile(modPath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

// buildBinary builds a Go binary with context support for quick failure
func buildBinary(ctx context.Context, sourceDir, outputPath string) error {
	debug := os.Getenv("DEBUG_TESTS") == "1"
	if debug {
		fmt.Printf("[BUILD] Starting build of %s at %s\n", sourceDir, time.Now().Format("15:04:05.000"))
		defer func() {
			fmt.Printf("[BUILD] Finished build of %s at %s\n", sourceDir, time.Now().Format("15:04:05.000"))
		}()
	}

	// Find main.go
	mainPath := ""
	patterns := []string{
		filepath.Join(sourceDir, "main.go"),
		filepath.Join(sourceDir, "cmd", "*", "main.go"),
		filepath.Join(sourceDir, "cmd", "*", "-cli", "main.go"),
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			mainPath = filepath.Dir(matches[0])
			break
		}
	}

	if mainPath == "" {
		return fmt.Errorf("main.go not found in %s", sourceDir)
	}

	// Build with context and timeout
	buildCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(buildCtx, "go", "build", "-o", outputPath, ".")
	cmd.Dir = mainPath
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GO111MODULE=on", "GOWORK=off")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("build canceled: %w", ctx.Err())
		}
		return fmt.Errorf("build failed: %w\nOutput: %s", err, output)
	}

	return nil
}
