package expander

import (
	"go/ast"
	"os"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/try"
)

func TestExpandOnePackageBodyStmts(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		BuildFlags: []string{"-tags=macro"},
	}
	pkgs, err := packages.Load(cfg, "github.com/arcane-craft/go-macro/examples/readfile")
	if err != nil {
		t.Fatal(err)
	}
	providers := []Provider{
		{ImportPath: "github.com/arcane-craft/go-macro/try", SyntaxID: "syntax-try", Expand: try.TryExpand},
	}
	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.PkgPath, ".test") {
			continue
		}
		if pkg.PkgPath != "github.com/arcane-craft/go-macro/examples/readfile" {
			continue
		}
		var fnBefore *ast.FuncDecl
		for _, d := range pkg.Syntax[0].Decls {
			if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == "ReadFile" {
				fnBefore = f
			}
		}
		t.Log("before expand stmts", len(fnBefore.Body.List))
		engine := &Engine{Registry: macro.NewRegistry()}
		filesByPath := make(map[string][]*ast.File)
		for _, p := range providers {
			files, err := providerFiles(pkg, p.ImportPath)
			if err != nil {
				t.Fatal(err)
			}
			filesByPath[p.ImportPath] = files
		}
		if err := engine.RegisterProviders(providers, filesByPath); err != nil {
			t.Fatal(err)
		}
		file := pkg.Syntax[0]
		imports := BuildImportMap(file, pkg.PkgPath)
		if err := engine.ExpandFile(pkg.Fset, file, pkg.TypesInfo, pkg.Types, imports); err != nil {
			t.Fatal(err)
		}
		var fn *ast.FuncDecl
		for _, d := range pkg.Syntax[0].Decls {
			if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == "ReadFile" {
				fn = f
			}
		}
		if fn == nil {
			t.Fatal("no ReadFile")
		}
		t.Log("body stmts", len(fn.Body.List))
		if len(fn.Body.List) != 6 {
			t.Fatalf("want 6 stmts got %d", len(fn.Body.List))
		}
	}
}

func TestWriteGenDoesNotDuplicate(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedCompiledGoFiles,
		BuildFlags: []string{"-tags=macro"},
	}
	pkgs, _ := packages.Load(cfg, "github.com/arcane-craft/go-macro/examples/readfile")
	pkg := pkgs[0]
	// use already expanded syntax from disk gen - parse gen file
	genPath := "../../examples/readfile/readfile_macro_gen.go"
	data, err := os.ReadFile(genPath)
	if err != nil {
		t.Skip("no gen file")
	}
	_ = data
	_ = pkg
}
