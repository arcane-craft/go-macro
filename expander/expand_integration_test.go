package expander

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/try"
)

func TestExpandReadfileOneRound(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedCompiledGoFiles |
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
	imports := BuildImportMap(file, pkg.PkgPath)
	fset := token.NewFileSet()
	trySrc, _ := parser.ParseFile(fset, "try.go", `package try
//macro: syntax-try
func Try[T any](v T, err error) T { panic("x") }
`, parser.ParseComments)
	reg := macro.NewRegistry()
	_ = reg.RegisterProvider("github.com/arcane-craft/go-macro/try", []*ast.File{trySrc}, "syntax-try", try.TryExpand)
	n := 0
	for {
		calls, err := RecognizeMacroCalls(file, pkg.TypesInfo, imports, reg)
		if err != nil {
			t.Fatal(err)
		}
		if len(calls) == 0 {
			break
		}
		n++
		if n > 5 {
			t.Fatalf("too many rounds; last stub %q", calls[0].StubName)
		}
		mc := calls[0]
		site := classifySiteInFile(file, mc.Call)
		enc := enclosingFuncInFile(file, mc.Call)
		ctx, err := macro.NewContext(pkg.Fset, pkg.TypesInfo, pkg.Types, mc.Call, mc.StubName, mc.SyntaxID, site, enc)
		if err != nil {
			t.Fatal(err)
		}
		result, err := try.TryExpand(ctx, mc.Call)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyExpandResult(file, mc.Call, site, result); err != nil {
			t.Fatal(err)
		}
	}
	if n != 1 {
		t.Fatalf("rounds %d want 1", n)
	}
	var fn *ast.FuncDecl
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("no func decl")
	}
	if len(fn.Body.List) != 6 {
		t.Fatalf("body stmts: %d want 6 (defer+assign+if+assign+defer+return)", len(fn.Body.List))
	}
}
