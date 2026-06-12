package expandtool

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arcane-craft/go-macro/internal/expander"
)

// RunExpandCommand generates .gomacro/expand_runner and runs expand with args.
func RunExpandCommand(args []string) error {
	expandArgs := args
	if len(expandArgs) == 0 {
		expandArgs = []string{"./..."}
	}
	modPatterns := moduleLoadPatterns(expandArgs)
	modRoot, err := expander.ModuleRoot(modPatterns)
	if err != nil {
		return err
	}
	links, err := expander.DiscoverProviderLinks(expandArgs)
	if err != nil {
		return err
	}
	runnerDir := filepath.Join(modRoot, ".gomacro", "expand_runner")
	if err := writeExpandRunner(runnerDir, links); err != nil {
		return err
	}
	goArgs := append([]string{"run", "./.gomacro/expand_runner"}, expandArgs...)
	cmd := exec.Command("go", goArgs...)
	cmd.Dir = modRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("expand runner: %w", err)
	}
	return nil
}

func moduleLoadPatterns(args []string) []string {
	if len(args) == 0 {
		return []string{"./..."}
	}
	out := make([]string, len(args))
	for i, a := range args {
		if a == "." {
			out[i] = "./..."
		} else {
			out[i] = a
		}
	}
	return out
}

func writeExpandRunner(dir string, links []expander.ProviderLink) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	seen := make(map[string]bool)
	var b strings.Builder
	b.WriteString("package main\n\nimport (\n")
	b.WriteString("\t\"os\"\n\n")
	b.WriteString("\t\"github.com/arcane-craft/go-macro/macro/expandtool\"\n")
	for _, l := range links {
		alias := importAlias(l.ImportPath, seen)
		fmt.Fprintf(&b, "\t%s %q\n", alias, l.ImportPath)
	}
	b.WriteString(")\n\nfunc main() {\n")
	seen = make(map[string]bool)
	for _, l := range links {
		alias := importAlias(l.ImportPath, seen)
		fmt.Fprintf(&b, "\texpandtool.Register(%q, %s.%s)\n", l.SyntaxID, alias, l.ExpanderName)
	}
	b.WriteString("\tif err := expandtool.Run(os.Args[1:], nil); err != nil {\n")
	b.WriteString("\t\texpandtool.ExitWithError(err)\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return os.WriteFile(filepath.Join(dir, "main.go"), []byte(b.String()), 0o644)
}

func importAlias(importPath string, seen map[string]bool) string {
	base := importPath
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	base = strings.NewReplacer("-", "_", ".", "_").Replace(base)
	if base == "" {
		base = "p"
	}
	alias := base
	for n := 0; seen[alias]; n++ {
		alias = fmt.Sprintf("%s%d", base, n)
	}
	seen[alias] = true
	return alias
}

// ExitWithError prints err to stderr and exits with code 1.
func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "macroexpand: %v\n", err)
	os.Exit(1)
}
