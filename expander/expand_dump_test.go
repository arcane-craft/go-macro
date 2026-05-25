package expander

import (
	"bytes"
	"go/ast"
	"go/printer"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/try"
)

func TestExpandFileProducesCorrectBody(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		BuildFlags: []string{"-tags=macro"},
	}
	pkgs, err := packages.Load(cfg, "github.com/arcane-craft/go-macro/examples/readfile")
	if err != nil {
		t.Fatal(err)
	}
	var pkg *packages.Package
	for _, p := range pkgs {
		if p.PkgPath == "github.com/arcane-craft/go-macro/examples/readfile" {
			pkg = p
			break
		}
	}
	if pkg == nil {
		t.Fatal("readfile package not found")
	}
	file := pkg.Syntax[0]
	providers := []Provider{
		{ImportPath: "github.com/arcane-craft/go-macro/try", SyntaxID: "syntax-try", Expand: try.TryExpand},
	}
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
	imports := BuildImportMap(file, pkg.PkgPath)
	if err := engine.ExpandFile(pkg.Fset, file, pkg.TypesInfo, pkg.Types, imports); err != nil {
		t.Fatal(err)
	}
	var fn *ast.FuncDecl
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == "ReadFile" {
			fn = f
		}
	}
	if len(fn.Body.List) != 6 {
		t.Fatalf("body stmts: %d want 6", len(fn.Body.List))
	}
	pr := printer.Config{Mode: printer.RawFormat}
	var printed strings.Builder
	for _, s := range fn.Body.List {
		var buf bytes.Buffer
		_ = pr.Fprint(&buf, pkg.Fset, s)
		printed.Write(buf.Bytes())
	}
	if strings.Count(printed.String(), "file := _v2") != 1 {
		t.Fatalf("unexpected body:\n%s", printed.String())
	}
	if !strings.Contains(printed.String(), "defer file.Close()") {
		t.Fatalf("missing defer:\n%s", printed.String())
	}
}
