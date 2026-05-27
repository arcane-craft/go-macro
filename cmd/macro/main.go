package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arcane-craft/go-macro/macro/expandtool"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "init":
		if len(os.Args) < 3 || os.Args[2] != "provider" {
			fmt.Fprintf(os.Stderr, "usage: go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>\n")
			os.Exit(2)
		}
		name := "mymac"
		if len(os.Args) >= 4 {
			name = os.Args[3]
		}
		if err := runInitProvider(name); err != nil {
			fmt.Fprintf(os.Stderr, "macro init: %v\n", err)
			os.Exit(1)
		}
	case "expand":
		args := os.Args[2:]
		if err := expandtool.RunExpandCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "macro expand: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `macro — Go procedural macro toolchain

Usage:
  go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>
  go run github.com/arcane-craft/go-macro/cmd/macro@latest expand [patterns...]

Examples:
  go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac
  go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
`)
}

func moduleImportPath() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func runInitProvider(name string) error {
	dir := name
	mod := moduleImportPath()
	importPath := name
	if mod != "" {
		importPath = mod + "/" + name
	}
	syntaxID := "syntax-" + name
	files := map[string]string{
		"stubs.go": fmt.Sprintf(`package %s

//macro: %s
// Macro is a syntax stub. Do not call at runtime.
func Macro[T any](v T) T {
	panic("Macro is a macro stub and must not be called at runtime")
}
`, name, syntaxID),
		"expand.go": fmt.Sprintf(`package %s

import (
	"go/ast"

	"github.com/arcane-craft/go-macro/macro"
)

//macro: %s

// MacroExpand expands Macro calls (placeholder: returns argument).
func MacroExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error) {
	if ctx.Site() != macro.SiteExpr {
		return macro.ExpandResult{}, macro.ErrorAt(ctx.FileSet(), ctx.MacroPos(), "Macro only allowed in expression position")
	}
	if len(call.Args) != 1 {
		return macro.ExpandResult{}, macro.ErrorAt(ctx.FileSet(), ctx.MacroPos(), "Macro expects one argument")
	}
	return macro.ExpandResult{Expr: call.Args[0]}, nil
}
`, name, syntaxID),
		"expand_test.go": fmt.Sprintf(`package %s_test

import (
	"testing"

	"%s"
	"github.com/arcane-craft/go-macro/macro/mactest"
)

func TestMacroExpand(t *testing.T) {
	_, err := mactest.Expand(%s.MacroExpand, "Macro", "%s", `+"`"+`
func f() int { return Macro(42) }
`+"`"+`)
	if err != nil {
		t.Fatal(err)
	}
}
`, name, importPath, name, syntaxID),
		"README.md": fmt.Sprintf("# %s macro provider\n\nSee [author guide](https://github.com/arcane-craft/go-macro/blob/main/docs/author-guide.md).\n\nConsumers run:\n\n```go\n//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .\n```\n", name),
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("created provider skeleton in %s/\n", dir)
	return nil
}
