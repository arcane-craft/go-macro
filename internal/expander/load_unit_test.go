package expander

import (
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestFindImportedPackageWalk(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, "github.com/arcane-craft/go-macro/internal/expander")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages")
	}
	dep := findImportedPackage(pkgs[0], "github.com/arcane-craft/go-macro/macro")
	if dep == nil || dep.PkgPath != "github.com/arcane-craft/go-macro/macro" {
		t.Fatalf("expected macro dependency, got %v", dep)
	}
	if findImportedPackage(pkgs[0], "example.com/missing") != nil {
		t.Fatal("unexpected package")
	}
}

func TestImportedProviderPaths(t *testing.T) {
	cfg := &packages.Config{
		Mode:       packages.NeedImports,
		BuildFlags: []string{"-tags=macro"},
	}
	pkgs, err := packages.Load(cfg, "./examples/readfile")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) == 0 || pkgs[0].PkgPath == "" {
		t.Skip("examples module not available")
	}
	m := importedProviderPaths(pkgs[0])
	if !m["github.com/arcane-craft/go-macro-contrib/try"] {
		t.Fatalf("imports: %v", m)
	}
}
